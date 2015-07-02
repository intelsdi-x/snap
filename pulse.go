package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest"
	"github.com/intelsdi-x/pulse/scheduler"
)

var (
	flAPIDisabled = cli.BoolFlag{
		Name:  "disable-api, d",
		Usage: "Disable the agent REST API",
	}
	flAPIPort = cli.IntFlag{
		Name:  "api-port,  p",
		Usage: "API port (Default: 8181)",
		Value: 8181,
	}
	flMaxProcs = cli.IntFlag{
		Name:  "max-procs, c",
		Usage: "Set max cores to use for Pulse Agent. Default is 1 core.",
		Value: 1,
	}
	// plugin
	flLogPath = cli.StringFlag{
		Name:   "log-path, o",
		Usage:  "Path for logs. Empty path logs to stdout.",
		EnvVar: "PULSE_LOG_PATH",
	}
	flLogLevel = cli.IntFlag{
		Name:   "log-level, l",
		Usage:  "1-5 (Debug, Info, Warning, Error, Fatal)",
		EnvVar: "PULSE_LOG_LEVEL",
		Value:  3,
	}
	flPluginVersion = cli.StringFlag{
		Name:   "auto-discover, a",
		Usage:  "Auto discover paths separated by colons.",
		EnvVar: "PULSE_AUTOLOAD_PATH",
	}
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
	// Add a check to see if gitversion is blank from the build process
	if gitversion == "" {
		gitversion = "unknown"
	}

	app := cli.NewApp()
	app.Name = "pulsed"
	app.Version = gitversion
	app.Usage = "A powerful telemetry agent framework"
	app.Flags = []cli.Flag{flAPIDisabled, flAPIPort, flLogLevel, flLogPath, flMaxProcs, flPluginVersion}

	app.Action = action
	app.Run(os.Args)
}

func action(ctx *cli.Context) {
	log.Info("Starting pulsed")
	logLevel := ctx.Int("log-level")
	logPath := ctx.String("log-path")
	maxProcs := ctx.Int("max-procs")
	disableApi := ctx.Bool("disable-api")
	apiPort := ctx.Int("api-port")
	autodiscoverPath := ctx.String("auto-discover")

	if logLevel < 1 || logLevel > 5 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "pulse-agent",
				"level":   logLevel,
			}).Fatal("log level was invalid (needs: 1-5)")
		os.Exit(1)
	}

	log.SetLevel(getLevel(logLevel))

	if logPath != "" {

		f, err := os.Stat(logPath)
		if err != nil {

			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulse-agent",
					"error":   err.Error(),
					"logpath": logPath,
				}).Fatal("bad log path (must be a dir)")
			os.Exit(0)
		}
		if !f.IsDir() {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulse-agent",
					"logpath": logPath,
				}).Fatal("bad log path this is not a directory")
			os.Exit(0)
		}

		file, err2 := os.OpenFile(fmt.Sprintf("%s/pulse.log", logPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err2 != nil {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulse-agent",
					"error":   err2.Error(),
					"logpath": logPath,
				}).Fatal("bad log path")
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// Set Max Processors for the Pulse agent.
	setMaxProcs(maxProcs)

	log.WithFields(
		log.Fields{
			"block":   "main",
			"_module": "pulse-agent",
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
	startInterruptHandling(s, c)

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

	if autodiscoverPath != "" {
		paths := filepath.SplitList(autodiscoverPath)
		c.SetAutodiscoverPaths(paths)
		for _, path := range paths {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":  "main",
						"_module": "pulse-agent",
						"logpath": path,
					}).Fatal(err)
				os.Exit(0)
			}
			for _, file := range files {
				pl, err := c.Load(fmt.Sprintf("%s/%s", path, file.Name()))
				if err != nil {
					log.WithFields(log.Fields{
						"_block":  "main",
						"_module": "pulse-agent",
						"logpath": path,
						"plugin":  file,
					}).Fatal(err)
				} else {
					log.WithFields(log.Fields{
						"_block":         "main",
						"_module":        "pulse-agent",
						"logpath":        path,
						"plugin":         file,
						"plugin-name":    pl.Name,
						"plugin-version": pl.Version,
						"plugin-type":    pl.TypeName,
					}).Info("Loading plugin")
				}
			}
		}
	}
	if !disableApi {
		r := rest.New()
		r.BindMetricManager(c)
		r.BindTaskManager(s)
		r.Start(fmt.Sprintf(":%d", apiPort))
	}

	select {} //run forever and ever
}

func setMaxProcs(maxProcs int) {
	var _maxProcs int
	envGoMaxProcs := runtime.GOMAXPROCS(-1)
	numProcs := runtime.NumCPU()
	if maxProcs == 0 && envGoMaxProcs <= numProcs {
		// By default if max_procs is not set, we set _maxProcs to the ENV variable GOMAXPROCS on the system. If this variable is not set by the user or is a negative number, runtime.GOMAXPROCS(-1) returns 1
		_maxProcs = envGoMaxProcs
	} else if envGoMaxProcs > numProcs {
		// We do not allow the user to exceed the number of cores in the system.
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulse-agent",
				"maxprocs": _maxProcs,
			}).Warning("ENV variable GOMAXPROCS greater than number of cores in the system")
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulse-agent",
				"maxprocs": _maxProcs,
			}).Error("setting pulse to use the number of cores in the system")
		_maxProcs = numProcs
	} else if maxProcs > 0 && maxProcs <= numProcs {
		// Our flag override is set. Use this value
		_maxProcs = maxProcs
	} else if maxProcs > numProcs {
		// Do not let the user set a value larger than number of cores in the system
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulse-agent",
				"maxprocs": _maxProcs,
			}).Warning("flag max_procs exceeds number of cores in the system. Setting Pulse to use the number of cores in the system")
		_maxProcs = numProcs
	} else if maxProcs < 0 {
		// Do not let the user set a negative value to get around number of cores limit
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulse-agent",
				"maxprocs": _maxProcs,
			}).Warning("flag max_procs set to negative number. Setting Pulse to use 1 core")
		_maxProcs = 1
	}

	log.WithFields(
		log.Fields{
			"_block":   "main",
			"_module":  "pulse-agent",
			"maxprocs": _maxProcs,
		}).Debug("maxprocs")
	runtime.GOMAXPROCS(_maxProcs)

	//Verify setting worked
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != _maxProcs {
		log.WithFields(
			log.Fields{
				"block":          "main",
				"_module":        "pulse-agent",
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
				"block":        "main",
				"_module":      "pulse-agent",
				"pulse-module": m.Name(),
			}).Info("module started")
	}
	return err
}

func printErrorAndExit(name string, err error) {
	log.WithFields(
		log.Fields{
			"block":        "main",
			"_module":      "pulse-agent",
			"error":        err.Error(),
			"pulse-module": name,
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
				"block":   "main",
				"_module": "pulse-agent",
			}).Info("shutting down modules")

		for _, m := range modules {
			log.WithFields(
				log.Fields{
					"block":        "main",
					"_module":      "pulse-agent",
					"pulse-module": m.Name(),
				}).Info("stopping module")
			m.Stop()
		}
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "pulse-agent",
				"signal":  sig.String(),
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
