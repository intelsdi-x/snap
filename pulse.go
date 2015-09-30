/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

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
		Name:   "max-procs, c",
		Usage:  "Set max cores to use for Pulse Agent. Default is 1 core.",
		Value:  1,
		EnvVar: "GOMAXPROCS",
	}
	flNumberOfPLs = cli.IntFlag{
		Name:   "max-running-plugins, m",
		Usage:  "The maximum number of instances of a loaded plugin to run",
		Value:  3,
		EnvVar: "PULSE_MAX_PLUGINS",
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
	flPluginTrust = cli.IntFlag{
		Name:   "plugin-trust, t",
		Usage:  "0-2 (Disabled, Enabled, Warning)",
		EnvVar: "PULSE_TRUST_LEVEL",
		Value:  1,
	}
	flKeyringFile = cli.StringFlag{
		Name:   "keyring-file, k",
		Usage:  "Keyring file for signing verification",
		EnvVar: "PULSE_KEYRING_FILE",
	}
	flCache = cli.StringFlag{
		Name:   "cache-expiration",
		Usage:  "The time limit for which a metric cache entry is valid",
		EnvVar: "PULSE_CACHE_EXPIRATION",
		Value:  "500ms",
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
	app.Flags = []cli.Flag{flAPIDisabled, flAPIPort, flLogLevel, flLogPath, flMaxProcs, flPluginVersion, flNumberOfPLs, flCache, flPluginTrust, flKeyringFile}

	app.Action = action
	app.Run(os.Args)
}

func action(ctx *cli.Context) {
	var l = map[int]string{
		1: "debug",
		2: "info",
		3: "warning",
		4: "error",
		5: "fatal",
	}

	var t = map[int]string{
		0: "disabled",
		1: "enabled",
		2: "warning",
	}

	logLevel := ctx.Int("log-level")
	logPath := ctx.String("log-path")
	maxProcs := ctx.Int("max-procs")
	disableApi := ctx.Bool("disable-api")
	apiPort := ctx.Int("api-port")
	autodiscoverPath := ctx.String("auto-discover")
	maxRunning := ctx.Int("max-running-plugins")
	pluginTrust := ctx.Int("plugin-trust")
	keyringFile := ctx.String("keyring-file")
	cachestr := ctx.String("cache-expiration")
	cache, err := time.ParseDuration(cachestr)
	if err != nil {
		log.Fatal(fmt.Sprintf("invalid cache-expiration format: %s", cachestr))
	}

	log.Info("Starting pulsed (version: ", gitversion, ")")

	// Set Max Processors for pulsed.
	setMaxProcs(maxProcs)

	if logLevel < 1 || logLevel > 5 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "pulsed",
				"level":   logLevel,
			}).Fatal("log level was invalid (needs: 1-5)")
		os.Exit(1)
	}

	if logPath != "" {
		f, err := os.Stat(logPath)
		if err != nil {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulsed",
					"error":   err.Error(),
					"logpath": logPath,
				}).Fatal("bad log path (must be a dir)")
			os.Exit(0)
		}
		if !f.IsDir() {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulsed",
					"logpath": logPath,
				}).Fatal("bad log path (this is not a directory)")
			os.Exit(0)
		}

		file, err2 := os.OpenFile(fmt.Sprintf("%s/pulse.log", logPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err2 != nil {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulsed",
					"error":   err2.Error(),
					"logpath": logPath,
				}).Fatal("bad log path")
		}
		defer file.Close()
		log.Info("setting log path to: ", logPath)
		log.SetOutput(file)
	}
	log.Info("setting log path to: stdout")

	c := control.New(
		control.MaxRunningPlugins(maxRunning),
		control.CacheExpiration(cache),
	)
	s := scheduler.New(
		scheduler.CollectQSizeOption(defaultQueueSize),
		scheduler.CollectWkrSizeOption(defaultPoolSize),
		scheduler.PublishQSizeOption(defaultQueueSize),
		scheduler.PublishWkrSizeOption(defaultPoolSize),
		scheduler.ProcessQSizeOption(defaultQueueSize),
		scheduler.ProcessWkrSizeOption(defaultPoolSize),
	)
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

	//Plugin Trust
	if pluginTrust < 0 || pluginTrust > 3 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "pulsed",
				"level":   pluginTrust,
			}).Fatal("Plugin trust was invalid (needs: 0-2)")
		os.Exit(1)
	}
	c.SetPluginTrustLevel(pluginTrust)
	log.Info("setting plugin trust level to: ", t[pluginTrust])
	//Keyring checking for trust levels 1 and 2
	if pluginTrust > 0 {
		if keyringFile == "" {
			log.WithFields(
				log.Fields{
					"block":       "main",
					"_module":     "pulsed",
					"keyringFile": keyringFile,
				}).Fatal("need keyring file when trust is on (--keyring-file or -k)")
			os.Exit(0)
		}
		f, err := os.Stat(keyringFile)
		if err != nil {
			log.WithFields(
				log.Fields{
					"block":       "main",
					"_module":     "pulsed",
					"error":       err.Error(),
					"keyringFile": keyringFile,
				}).Fatal("bad keyring file")
			os.Exit(0)
		}
		if f.IsDir() {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "pulsed",
					"logpath": keyringFile,
				}).Fatal("bad keyring file (this is not a file)")
			os.Exit(0)
		}
		file, err := os.Open(keyringFile)
		if err != nil {
			log.WithFields(
				log.Fields{
					"block":       "main",
					"_module":     "pulsed",
					"error":       err.Error(),
					"keyringFile": keyringFile,
				}).Fatal("can't open keyring path")
			os.Exit(0)
			defer file.Close()
		}
		log.Info("setting keyring file to: ", keyringFile)
		c.SetKeyringFile(keyringFile)
	}

	//Autodiscover
	if autodiscoverPath != "" {
		log.Info("auto discover path is enabled")
		log.Info("autoloading plugins from: ", autodiscoverPath)
		paths := filepath.SplitList(autodiscoverPath)
		c.SetAutodiscoverPaths(paths)
		for _, path := range paths {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "main",
						"_module":          "pulsed",
						"autodiscoverpath": path,
					}).Fatal(err)
				os.Exit(0)
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if !strings.Contains(file.Name(), ".") {
					pl, err := c.Load(fmt.Sprintf("%s/%s", path, file.Name()))
					if err != nil {
						log.WithFields(log.Fields{
							"_block":           "main",
							"_module":          "pulsed",
							"autodiscoverpath": path,
							"plugin":           file,
						}).Error(err)
					} else {
						log.WithFields(log.Fields{
							"_block":           "main",
							"_module":          "pulsed",
							"autodiscoverpath": path,
							"plugin":           file,
							"plugin-name":      pl.Name,
							"plugin-version":   pl.Version,
							"plugin-type":      pl.TypeName,
						}).Info("Loading plugin")
					}
				}
			}
		}
	} else {
		log.Info("auto discover path is disabled")
	}

	//API
	if !disableApi {
		log.Info("Rest API enabled on port ", apiPort)
		r := rest.New()
		r.BindMetricManager(c)
		r.BindTaskManager(s)
		r.Start(fmt.Sprintf(":%d", apiPort))
	} else {
		log.Info("Rest API is disabled")
	}

	log.WithFields(
		log.Fields{
			"block":   "main",
			"_module": "pulsed",
		}).Info("pulsed started")

	// Switch log level to user defined
	log.Info("setting log level to: ", l[logLevel])
	log.SetLevel(getLevel(logLevel))

	select {} //run forever and ever
}

// setMaxProcs configures runtime.GOMAXPROCS for pulsed. GOMAXPROCS can be set by using
// the env variable GOMAXPROCS and pulsed will honor this setting. A user can override the env
// variable by setting max-procs flag on the command line. Pulsed will be limited to the max CPUs
// on the system even if the env variable or the command line setting is set above the max CPUs.
// The default value if the env variable or the command line option is not set is 1.
func setMaxProcs(maxProcs int) {
	var _maxProcs int
	numProcs := runtime.NumCPU()
	if maxProcs <= 0 {
		// We prefer sane values for GOMAXPROCS
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulsed",
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS to an invalid value")
		_maxProcs = 1
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulsed",
				"maxprocs": _maxProcs,
			}).Warning("Setting GOMAXPROCS to 1")
		_maxProcs = 1
	} else if maxProcs <= numProcs {
		_maxProcs = maxProcs
	} else {
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulsed",
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS larger than number of CPUs available on system")
		_maxProcs = numProcs
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "pulsed",
				"maxprocs": _maxProcs,
			}).Warning("Setting GOMAXPROCS to number of CPUs on host")
	}

	log.Info("setting GOMAXPROCS to: ", _maxProcs, " core(s)")
	runtime.GOMAXPROCS(_maxProcs)
	//Verify setting worked
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != _maxProcs {
		log.WithFields(
			log.Fields{
				"block":          "main",
				"_module":        "pulsed",
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
				"_module":      "pulsed",
				"pulse-module": m.Name(),
			}).Info("module started")
	}
	return err
}

func printErrorAndExit(name string, err error) {
	log.WithFields(
		log.Fields{
			"block":        "main",
			"_module":      "pulsed",
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
				"_module": "pulsed",
			}).Info("shutting down modules")

		for _, m := range modules {
			log.WithFields(
				log.Fields{
					"block":        "main",
					"_module":      "pulsed",
					"pulse-module": m.Name(),
				}).Info("stopping module")
			m.Stop()
		}
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "pulsed",
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
