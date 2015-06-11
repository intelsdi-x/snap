package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest"
	// "github.com/intelsdi-x/pulse/pkg/logger"
	"github.com/intelsdi-x/pulse/scheduler"
)

var (
	// Pulse Flags for command line
	version  = flag.Bool("version", false, "Print Pulse version")
	restMgmt = flag.Bool("rest", false, "start rest interface for Pulse")
	maxProcs = flag.Int("max_procs", 0, "Set max cores to use for Pulse Agent. Default is 1 core.")
	logPath  = flag.String("log_path", "", "Path for logs. Empty path logs to stdout.")
	logLevel = flag.Int("log_level", 2, "1-5 (Debug, Info, Warning, Error, Fatal")

	gitversion string
)

const (
	defaultQueueSize uint = 25
	defaultPoolSize  uint = 4
)

type coreModule interface {
	Start() error
	Stop()
	Name() string
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("Pulse version:", gitversion)
		os.Exit(0)
	}

	if *logLevel < 1 || *logLevel > 5 {
		log.WithFields(
			log.Fields{
				"block": "main",
				"level": *logLevel,
			}).Fatal("log level was invalid (needs: 1-5)")
		os.Exit(1)
	}

	log.SetLevel(getLevel(*logLevel))

	if *logPath != "" {

		f, err := os.Stat(*logPath)
		if err != nil {

			log.WithFields(
				log.Fields{
					"block":   "main",
					"error":   err.Error(),
					"logpath": *logPath,
				}).Fatal("bad log path (must be a dir)")
			os.Exit(0)
		}
		if !f.IsDir() {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"logpath": *logPath,
				}).Fatal("bad log path this is not a directory")
			os.Exit(0)
		}

		file, err2 := os.OpenFile(fmt.Sprintf("%s/pulse.log", *logPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err2 != nil {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"error":   err2.Error(),
					"logpath": *logPath,
				}).Fatal("bad log path")
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// Set Max Processors for the Pulse agent.
	setMaxProcs()

	log.WithFields(
		log.Fields{
			"block": "main",
		}).Info("pulse agent starting")
	c := control.New()
	s := scheduler.New(
		scheduler.CollectQSizeOption(defaultQueueSize),
		scheduler.CollectWkrSizeOption(defaultPoolSize),
		scheduler.PublishQSizeOption(defaultQueueSize),
		scheduler.PublishWkrSizeOption(defaultPoolSize),
		scheduler.ProcessQSizeOption(defaultQueueSize),
		scheduler.ProcessWkrSizeOption(defaultPoolSize))
	s.SetMetricManager(c)

	// Set interrupt handling so we can die gracefully.
	startInterruptHandling(c, s)

	//  Start our modules
	if err := startModule(c); err != nil {
		printErrorAndExit(c.Name(), err)
	}
	if err := startModule(s); err != nil {
		if c.Started {
			c.Stop()
		}
		printErrorAndExit(s.Name(), err)
	}

	// init rest mgmt interface
	if *restMgmt {
		r := rest.New()
		r.BindMetricManager(c)
		r.BindTaskManager(s)
		r.Start(":8181")
	}

	select {} //run forever and ever
}

func setMaxProcs() {
	var _maxProcs int
	envGoMaxProcs := runtime.GOMAXPROCS(-1)
	numProcs := runtime.NumCPU()
	if *maxProcs == 0 && envGoMaxProcs <= numProcs {
		// By default if max_procs is not set, we set _maxProcs to the ENV variable GOMAXPROCS on the system. If this variable is not set by the user or is a negative number, runtime.GOMAXPROCS(-1) returns 1
		_maxProcs = envGoMaxProcs
	} else if envGoMaxProcs > numProcs {
		// We do not allow the user to exceed the number of cores in the system.
		log.WithFields(
			log.Fields{
				"block":    "main",
				"maxprocs": _maxProcs,
			}).Warning("ENV variable GOMAXPROCS greater than number of cores in the system")
		log.WithFields(
			log.Fields{
				"block":    "main",
				"maxprocs": _maxProcs,
			}).Error("setting pulse to use the number of cores in the system")
		_maxProcs = numProcs
	} else if *maxProcs > 0 && *maxProcs <= numProcs {
		// Our flag override is set. Use this value
		_maxProcs = *maxProcs
	} else if *maxProcs > numProcs {
		// Do not let the user set a value larger than number of cores in the system
		log.WithFields(
			log.Fields{
				"block":    "main",
				"maxprocs": _maxProcs,
			}).Warning("flag max_procs exceeds number of cores in the system. Setting Pulse to use the number of cores in the system")
		_maxProcs = numProcs
	} else if *maxProcs < 0 {
		// Do not let the user set a negative value to get around number of cores limit
		log.WithFields(
			log.Fields{
				"block":    "main",
				"maxprocs": _maxProcs,
			}).Warning("flag max_procs set to negative number. Setting Pulse to use 1 core")
		_maxProcs = 1
	}

	log.WithFields(
		log.Fields{
			"block":    "main",
			"maxprocs": _maxProcs,
		}).Debug("maxprocs")
	runtime.GOMAXPROCS(_maxProcs)

	//Verify setting worked
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != _maxProcs {
		log.WithFields(
			log.Fields{
				"block":          "main",
				"given maxprocs": _maxProcs,
				"real maxprocs":  actualNumProcs,
			}).Warning("not using given maxprocs")
	}
}

func startModule(m coreModule) error {
	err := m.Start()
	if err == nil {
		log.WithFields(
			log.Fields{
				"block":  "startModule",
				"module": m.Name(),
			}).Info("module started")
	}
	return err
}

func printErrorAndExit(name string, err error) {
	log.WithFields(
		log.Fields{
			"block":  "main",
			"error":  err.Error(),
			"module": name,
		}).Fatal("error starting module")
	os.Exit(1)
}

func startInterruptHandling(modules ...coreModule) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	//Let's block until someone tells us to quit
	go func() {
		sig := <-c
		log.WithFields(
			log.Fields{
				"block": "main",
			}).Info("shutting down modules")

		for _, m := range modules {
			log.WithFields(
				log.Fields{
					"block":  "main",
					"module": m.Name(),
				}).Info("stopping module")
			m.Stop()
		}
		log.WithFields(
			log.Fields{
				"block":  "main",
				"signal": sig.String(),
			}).Info("exiting on signal")
		os.Exit(0)
	}()
}

func getLevel(i int) log.Level {
	switch i {
	case 1:
		return log.DebugLevel
	case 2:
		return log.InfoLevel
	case 3:
		return log.WarnLevel
	case 4:
		return log.ErrorLevel
	case 5:
		return log.FatalLevel
	default:
		panic("bad level")
	}
}
