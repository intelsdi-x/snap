package main

import (
	"flag"
	"fmt"
	// "log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/pkg/logger"
	"github.com/intelsdilabs/pulse/schedule"
)

var (
	// Pulse Flags for command line
	version  = flag.Bool("version", false, "Print Pulse version")
	maxProcs = flag.Int("max_procs", 0, "Set max cores to use for Pulse Agent. Default is 1 core.")
	logPath  = flag.String("log_path", "", "Path for logs. Empty path logs to stdout.")
	logLevel = flag.Int("log_level", 2, "1-5 (Debug, Info, Warning, Error, Fatal")

	gitversion string
)

const (
	defaultQueueSize int = 25
	defaultPoolSize  int = 4
)

type coreModule interface {
	Start() error
	Stop()
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("Pulse version:", gitversion)
		os.Exit(0)
	}

	if *logLevel < 1 || *logLevel > 5 {
		logger.Fatal("main", fmt.Sprintf("log level is invalid: was %d, valid (1-5)", *logLevel))
		os.Exit(1)
	}

	logger.SetLevel(logger.LogLevel(*logLevel))

	if *logPath != "" {

		f, err := os.Stat(*logPath)
		if err != nil {

			logger.Error("main", fmt.Sprintf("bad log path(%s) - %s\n", *logPath, err.Error()))
			os.Exit(0)
		}
		if !f.IsDir() {
			logger.Error("main", fmt.Sprintf("bad log path(%s) - not a directory\n", *logPath))
			os.Exit(0)
		}

		file, err2 := os.OpenFile(fmt.Sprintf("%s/pulse.log", *logPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err2 != nil {
			logger.Error("main", fmt.Sprintf("bad log path(%s) - %s\n", *logPath, err.Error()))
		}
		defer file.Close()
		logger.SetOutput(file)
	}

	// Set Max Processors for the Pulse agent.
	setMaxProcs()

	logger.Info("main", "pulse agent starting")
	c := control.New()
	s := schedule.New(defaultPoolSize, defaultQueueSize)
	s.SetMetricManager(c)

	// Set interrupt handling so we can die gracefully.
	startInterruptHandling(c, s)

	//  Start our modules
	if err := startModule("plugin controller", c); err != nil {
		printErrorAndExit("plugin controller", err)
	}
	if err := startModule("scheduler", s); err != nil {
		if c.Started {
			c.Stop()
		}
		printErrorAndExit("scheduler", err)
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
		logger.Warning("main", "ENV variable GOMAXPROCS greater than number of cores in the system.")
		logger.Info("main", "setting pulse to use the number of cores in the system.")
		_maxProcs = numProcs
	} else if *maxProcs > 0 && *maxProcs <= numProcs {
		// Our flag override is set. Use this value
		_maxProcs = *maxProcs
	} else if *maxProcs > numProcs {
		// Do not let the user set a value larger than number of cores in the system
		logger.Warning("main", "flag max_procs exceeds number of cores in the system. Setting Pulse to use the number of cores in the system")
		_maxProcs = numProcs
	} else if *maxProcs < 0 {
		// Do not let the user set a negative value to get around number of cores limit
		logger.Warning("main", "flag max_procs set to negative number. Setting Pulse to use 1 core.")
		_maxProcs = 1
	}

	logger.Info("main", fmt.Sprintf("GOMAXPROCS=%v\n", _maxProcs))
	runtime.GOMAXPROCS(_maxProcs)

	//Verify setting worked
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != _maxProcs {
		logger.Warning("main", fmt.Sprintf("specified max procs of %v but using %v", _maxProcs, actualNumProcs))
	}
}

func startModule(name string, m coreModule) error {
	err := m.Start()
	if err == nil {
		logger.Info("main", fmt.Sprintf("%s module started\n", name))
	}
	return err
}

func printErrorAndExit(name string, err error) {
	logger.Error("main", err.Error())
	logger.Error("main", fmt.Sprintf("error starting pulse agent %s module. Exiting now.", name))
	os.Exit(1)
}

func startInterruptHandling(modules ...coreModule) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	//Let's block until someone tells us to quit
	go func() {
		sig := <-c
		logger.Info("main", "stopping pulse agent modules")
		for _, m := range modules {
			m.Stop()
		}
		logger.Info("main", fmt.Sprintf("exiting given signal: %v", sig))
		os.Exit(0)
	}()
}
