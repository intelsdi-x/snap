/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/vrischmann/jsonutil"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest"
	"github.com/intelsdi-x/snap/mgmt/tribe"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/scheduler"
)

var (
	flAPIDisabled = cli.BoolFlag{
		Name:  "disable-api, d",
		Usage: "Disable the agent REST API",
	}
	flAPIPort = cli.IntFlag{
		Name:  "api-port,  p",
		Usage: "API port (Default: 8181)",
	}
	flMaxProcs = cli.IntFlag{
		Name:   "max-procs, c",
		Usage:  "Set max cores to use for snap Agent. Default is 1 core.",
		EnvVar: "GOMAXPROCS",
	}
	flNumberOfPLs = cli.IntFlag{
		Name:   "max-running-plugins, m",
		Usage:  "The maximum number of instances of a loaded plugin to run",
		EnvVar: "SNAP_MAX_PLUGINS",
	}
	// plugin
	flLogPath = cli.StringFlag{
		Name:   "log-path, o",
		Usage:  "Path for logs. Empty path logs to stdout.",
		EnvVar: "SNAP_LOG_PATH",
	}
	flLogLevel = cli.IntFlag{
		Name:   "log-level, l",
		Usage:  "1-5 (Debug, Info, Warning, Error, Fatal)",
		EnvVar: "SNAP_LOG_LEVEL",
	}
	flAutoDiscover = cli.StringFlag{
		Name:   "auto-discover, a",
		Usage:  "Auto discover paths separated by colons.",
		EnvVar: "SNAP_AUTODISCOVER_PATH",
	}
	flPluginTrust = cli.IntFlag{
		Name:   "plugin-trust, t",
		Usage:  "0-2 (Disabled, Enabled, Warning)",
		EnvVar: "SNAP_TRUST_LEVEL",
	}
	flKeyringPaths = cli.StringFlag{
		Name:   "keyring-paths, k",
		Usage:  "Keyring paths for signing verification separated by colons",
		EnvVar: "SNAP_KEYRING_PATHS",
	}
	flCache = cli.DurationFlag{
		Name:   "cache-expiration",
		Usage:  "The time limit for which a metric cache entry is valid",
		EnvVar: "SNAP_CACHE_EXPIRATION",
	}
	flConfig = cli.StringFlag{
		Name:   "config",
		Usage:  "A path to a config file",
		EnvVar: "SNAP_CONFIG_PATH",
	}
	flRestHTTPS = cli.BoolFlag{
		Name:  "rest-https",
		Usage: "start snap's API as https",
	}
	flRestCert = cli.StringFlag{
		Name:  "rest-cert",
		Usage: "A path to a certificate to use for HTTPS deployment of snap's REST API",
	}
	flRestKey = cli.StringFlag{
		Name:  "rest-key",
		Usage: "A path to a key file to use for HTTPS deployment of snap's REST API",
	}
	flRestAuth = cli.BoolFlag{
		Name:  "rest-auth",
		Usage: "Enables snap's REST API authentication",
	}

	gitversion  string
	coreModules []coreModule

	// log levels
	l = map[int]string{
		1: "debug",
		2: "info",
		3: "warning",
		4: "error",
		5: "fatal",
	}
	// plugin trust levels
	t = map[int]string{
		0: "disabled",
		1: "enabled",
		2: "warning",
	}
)

// default configuration values
const (
	defaultLogLevel   int    = 3
	defaultGoMaxProcs int    = 1
	defaultLogPath    string = ""
	defaultConfigPath string = "/etc/snap/snapd.conf"
)

// holds the configuration passed in through the SNAP config file
type Config struct {
	LogLevel   int               `json:"log_level,omitempty"yaml:"log_level,omitempty"`
	GoMaxProcs int               `json:"gomaxprocs,omitempty"yaml:"gomaxprocs,omitempty"`
	LogPath    string            `json:"log_path,omitempty"yaml:"log_path,omitempty"`
	Control    *control.Config   `json:"control,omitempty"yaml:"control,omitempty"`
	Scheduler  *scheduler.Config `json:"scheduler,omitempty"yaml:"scheduler,omitempty"`
	RestAPI    *rest.Config      `json:"restapi,omitempty"yaml:"restapi,omitempty"`
	Tribe      *tribe.Config     `json:"tribe,omitempty"yaml:"tribe,omitempty"`
}

type coreModule interface {
	Start() error
	Stop()
	Name() string
}

type managesTribe interface {
	GetAgreement(name string) (*agreement.Agreement, serror.SnapError)
	GetAgreements() map[string]*agreement.Agreement
	AddAgreement(name string) serror.SnapError
	RemoveAgreement(name string) serror.SnapError
	JoinAgreement(agreementName, memberName string) serror.SnapError
	LeaveAgreement(agreementName, memberName string) serror.SnapError
	GetMembers() []string
	GetMember(name string) *agreement.Member
}

func main() {
	// Add a check to see if gitversion is blank from the build process
	if gitversion == "" {
		gitversion = "unknown"
	}

	app := cli.NewApp()
	app.Name = "snapd"
	app.Version = gitversion
	app.Usage = "A powerful telemetry framework"
	app.Flags = []cli.Flag{
		flAPIDisabled,
		flAPIPort,
		flLogLevel,
		flLogPath,
		flMaxProcs,
		flAutoDiscover,
		flNumberOfPLs,
		flCache,
		flPluginTrust,
		flKeyringPaths,
		flRestCert,
		flConfig,
		flRestHTTPS,
		flRestKey,
		flRestAuth,
	}
	app.Flags = append(app.Flags, scheduler.Flags...)
	app.Flags = append(app.Flags, tribe.Flags...)

	app.Action = action
	app.Run(os.Args)
}

func action(ctx *cli.Context) {
	// get default configuration
	cfg := getDefaultConfig()

	// read config file
	readConfig(cfg, ctx.String("config"))

	// apply values that may have been passed from the command line
	// to the configuration that we have built so far, overriding the
	// values that may have already been set (if any) for the
	// same variables in that configuration
	applyCmdLineFlags(cfg, ctx)

	// If logPath is set, we verify the logPath and set it so that all logging
	// goes to the log file instead of stdout.
	logPath := cfg.LogPath
	if logPath != "" {
		f, err := os.Stat(logPath)
		if err != nil {
			log.Fatal(err)
		}
		if !f.IsDir() {
			log.Fatal("log path provided must be a directory")
		}

		file, err := os.OpenFile(fmt.Sprintf("%s/snapd.log", logPath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	log.Info("Starting snapd (version: ", gitversion, ")")

	// Set Max Processors for snapd.
	setMaxProcs(cfg.GoMaxProcs)

	// Validate log level and trust level settings for snapd
	validateLevelSettings(cfg.LogLevel, cfg.Control.PluginTrust)

	c := control.New(cfg.Control)

	coreModules = []coreModule{}

	coreModules = append(coreModules, c)
	s := scheduler.New(cfg.Scheduler)
	s.SetMetricManager(c)
	coreModules = append(coreModules, s)

	// Auth requested and not provided as part of config
	if cfg.RestAPI.Enable && cfg.RestAPI.RestAuth && cfg.RestAPI.RestAuthPassword == "" {
		fmt.Println("What password do you want to use for authentication?")
		fmt.Print("Password:")
		password, err := terminal.ReadPassword(0)
		fmt.Println()
		if err != nil {
			log.Fatal("Failed to get credentials")
		}
		cfg.RestAPI.RestAuthPassword = string(password)
	}

	var tr managesTribe
	if cfg.Tribe.Enable {
		cfg.Tribe.RestAPIPort = cfg.RestAPI.Port
		if cfg.RestAPI.RestAuth {
			cfg.Tribe.RestAPIPassword = cfg.RestAPI.RestAuthPassword
		}
		log.Info("Tribe is enabled")
		t, err := tribe.New(cfg.Tribe)
		if err != nil {
			printErrorAndExit(t.Name(), err)
		}
		c.RegisterEventHandler("tribe", t)
		t.SetPluginCatalog(c)
		s.RegisterEventHandler("tribe", t)
		t.SetTaskManager(s)
		coreModules = append(coreModules, t)
		tr = t
	}

	//Setup RESTful API if it was enabled in the configuration
	if cfg.RestAPI.Enable {
		r, err := rest.New(cfg.RestAPI)
		if err != nil {
			log.Fatal(err)
		}
		r.BindMetricManager(c)
		r.BindConfigManager(c.Config)
		r.BindTaskManager(s)

		//Rest Authentication
		if cfg.RestAPI.RestAuth {
			log.Info("REST API authentication is enabled")
			r.SetAPIAuth(cfg.RestAPI.RestAuth)
			log.Info("REST API authentication password is set")
			r.SetAPIAuthPwd(cfg.RestAPI.RestAuthPassword)
			if !cfg.RestAPI.HTTPS {
				log.Warning("Using REST API authentication without HTTPS enabled.")
			}
		}

		if tr != nil {
			r.BindTribeManager(tr)
		}
		go monitorErrors(r.Err())
		r.SetAddress(fmt.Sprintf(":%d", cfg.RestAPI.Port))
		coreModules = append(coreModules, r)
		log.Info("REST API is enabled")
	} else {
		log.Info("REST API is disabled")
	}

	// Set interrupt handling so we can die gracefully.
	startInterruptHandling(coreModules...)

	// Start our modules
	var started []coreModule
	for _, m := range coreModules {
		if err := startModule(m); err != nil {
			for _, m := range started {
				m.Stop()
			}
			printErrorAndExit(m.Name(), err)
		}
		started = append(started, m)
	}

	// Plugin Trust
	c.SetPluginTrustLevel(cfg.Control.PluginTrust)
	log.Info("setting plugin trust level to: ", t[cfg.Control.PluginTrust])
	// Keyring checking for trust levels 1 and 2
	if cfg.Control.PluginTrust > 0 {
		keyrings := filepath.SplitList(cfg.Control.KeyringPaths)
		if len(keyrings) == 0 {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": "snapd",
				}).Fatal("need keyring file when trust is on (--keyring-file or -k)")
		}
		for _, k := range keyrings {
			keyringPath, err := filepath.Abs(k)
			if err != nil {
				log.WithFields(
					log.Fields{
						"block":       "main",
						"_module":     "snapd",
						"error":       err.Error(),
						"keyringPath": keyringPath,
					}).Fatal("Unable to determine absolute path to keyring file")
			}
			f, err := os.Stat(keyringPath)
			if err != nil {
				log.WithFields(
					log.Fields{
						"block":       "main",
						"_module":     "snapd",
						"error":       err.Error(),
						"keyringPath": keyringPath,
					}).Fatal("bad keyring file")
			}
			if f.IsDir() {
				log.Info("Adding keyrings from: ", keyringPath)
				files, err := ioutil.ReadDir(keyringPath)
				if err != nil {
					log.WithFields(
						log.Fields{
							"_block":      "main",
							"_module":     "snapd",
							"error":       err.Error(),
							"keyringPath": keyringPath,
						}).Fatal(err)
				}
				if len(files) == 0 {
					log.Fatal(fmt.Sprintf("given keyring path [%s] is an empty directory!", keyringPath))
				}
				for _, keyringFile := range files {
					if keyringFile.IsDir() {
						continue
					}
					if strings.HasSuffix(keyringFile.Name(), ".gpg") || (strings.HasSuffix(keyringFile.Name(), ".pub")) || (strings.HasSuffix(keyringFile.Name(), ".pubring")) {
						f, err := os.Open(keyringPath)
						if err != nil {
							log.WithFields(
								log.Fields{
									"block":       "main",
									"_module":     "snapd",
									"error":       err.Error(),
									"keyringPath": keyringPath,
								}).Warning("unable to open keyring file. not adding to keyring path")
							continue
						}
						f.Close()
						log.Info("adding keyring file: ", keyringPath+"/"+keyringFile.Name())
						c.SetKeyringFile(keyringPath + "/" + keyringFile.Name())
					}
				}
			} else {
				f, err := os.Open(keyringPath)
				if err != nil {
					log.WithFields(
						log.Fields{
							"block":       "main",
							"_module":     "snapd",
							"error":       err.Error(),
							"keyringPath": keyringPath,
						}).Fatal("unable to open keyring file.")
				}
				f.Close()
				log.Info("adding keyring file ", keyringPath)
				c.SetKeyringFile(keyringPath)
			}
		}
	}

	//Autodiscover
	if cfg.Control.AutoDiscoverPath != "" {
		log.Info("auto discover path is enabled")
		paths := filepath.SplitList(cfg.Control.AutoDiscoverPath)
		c.SetAutodiscoverPaths(paths)
		for _, p := range paths {
			fullPath, err := filepath.Abs(p)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "main",
						"_module":          "snapd",
						"autodiscoverpath": p,
					}).Fatal(err)
			}
			log.Info("autoloading plugins from: ", fullPath)
			files, err := ioutil.ReadDir(fullPath)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "main",
						"_module":          "snapd",
						"autodiscoverpath": fullPath,
					}).Fatal(err)
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if strings.HasSuffix(file.Name(), ".aci") || !(strings.HasSuffix(file.Name(), ".asc")) {
					rp, err := core.NewRequestedPlugin(path.Join(fullPath, file.Name()))
					if err != nil {
						log.WithFields(log.Fields{
							"_block":           "main",
							"_module":          "snapd",
							"autodiscoverpath": fullPath,
							"plugin":           file,
						}).Error(err)
					}
					signatureFile := file.Name() + ".asc"
					if _, err := os.Stat(path.Join(fullPath, signatureFile)); err == nil {
						err = rp.ReadSignatureFile(path.Join(fullPath, signatureFile))
						if err != nil {
							log.WithFields(log.Fields{
								"_block":           "main",
								"_module":          "snapd",
								"autodiscoverpath": fullPath,
								"plugin":           file.Name() + ".asc",
							}).Error(err)
						}
					}
					pl, err := c.Load(rp)
					if err != nil {
						log.WithFields(log.Fields{
							"_block":           "main",
							"_module":          "snapd",
							"autodiscoverpath": fullPath,
							"plugin":           file,
						}).Error(err)
					} else {
						log.WithFields(log.Fields{
							"_block":           "main",
							"_module":          "snapd",
							"autodiscoverpath": fullPath,
							"plugin-file-name": file.Name(),
							"plugin-name":      pl.Name(),
							"plugin-version":   pl.Version(),
							"plugin-type":      pl.TypeName(),
						}).Info("Loading plugin")
					}
				}
			}
		}
	} else {
		log.Info("auto discover path is disabled")
	}

	log.WithFields(
		log.Fields{
			"block":   "main",
			"_module": "snapd",
		}).Info("snapd started")

	// Switch log level to user defined
	log.Info("setting log level to: ", l[cfg.LogLevel])
	log.SetLevel(getLevel(cfg.LogLevel))

	select {} //run forever and ever
}

// get the default snapd configuration
func getDefaultConfig() *Config {
	return &Config{
		LogLevel:   defaultLogLevel,
		GoMaxProcs: defaultGoMaxProcs,
		LogPath:    defaultLogPath,
		Control:    control.GetDefaultConfig(),
		Scheduler:  scheduler.GetDefaultConfig(),
		RestAPI:    rest.GetDefaultConfig(),
		Tribe:      tribe.GetDefaultConfig(),
	}
}

// Read the snapd configuration from a configuration file
func readConfig(cfg *Config, fpath string) {
	var path string
	if !defaultConfigFile() && fpath == "" {
		return
	}
	if defaultConfigFile() && fpath == "" {
		path = defaultConfigPath
	}
	if fpath != "" {
		f, err := os.Stat(fpath)
		if err != nil {
			log.Fatal(err)
		}
		if f.IsDir() {
			log.Fatal("configuration path provided must be a file")
		}
		path = fpath
	}

	err := cfgfile.Read(path, &cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func defaultConfigFile() bool {
	_, err := os.Stat(defaultConfigPath)
	if err != nil {
		return false
	}
	return true
}

// used to set fields in the configuration to values from the
// command line context if the corresponding flagName is set
// in that context
func setBoolVal(field bool, ctx *cli.Context, flagName string, inverse ...bool) bool {
	if ctx.IsSet(flagName) {
		field = ctx.Bool(flagName)
		if len(inverse) > 0 {
			field = !field
		}
	}
	return field
}

func setStringVal(field string, ctx *cli.Context, flagName string) string {
	if ctx.IsSet(flagName) {
		field = ctx.String(flagName)
	}
	return field
}

func setIntVal(field int, ctx *cli.Context, flagName string) int {
	if ctx.IsSet(flagName) {
		field = ctx.Int(flagName)
	}
	return field
}

func setUIntVal(field uint, ctx *cli.Context, flagName string) uint {
	if ctx.IsSet(flagName) {
		field = uint(ctx.Int(flagName))
	}
	return field
}

func setDurationVal(field time.Duration, ctx *cli.Context, flagName string) time.Duration {
	if ctx.IsSet(flagName) {
		field = ctx.Duration(flagName)
	}
	return field
}

// Apply the command line flags set (if any) to override the values
// in the input configuration
func applyCmdLineFlags(cfg *Config, ctx *cli.Context) {
	invertBoolean := true
	// apply any command line flags that might have been set, first for the
	// snapd-related flags
	cfg.GoMaxProcs = setIntVal(cfg.GoMaxProcs, ctx, "max-procs")
	cfg.LogLevel = setIntVal(cfg.LogLevel, ctx, "log-level")
	cfg.LogPath = setStringVal(cfg.LogPath, ctx, "log-path")
	// next for the flags related to the control package
	cfg.Control.MaxRunningPlugins = setIntVal(cfg.Control.MaxRunningPlugins, ctx, "max-running-plugins")
	cfg.Control.PluginTrust = setIntVal(cfg.Control.PluginTrust, ctx, "plugin-trust")
	cfg.Control.AutoDiscoverPath = setStringVal(cfg.Control.AutoDiscoverPath, ctx, "auto-discover")
	cfg.Control.KeyringPaths = setStringVal(cfg.Control.KeyringPaths, ctx, "keyring-paths")
	cfg.Control.CacheExpiration = jsonutil.Duration{setDurationVal(cfg.Control.CacheExpiration.Duration, ctx, "cache-expiration")}
	// next for the RESTful server related flags
	cfg.RestAPI.Enable = setBoolVal(cfg.RestAPI.Enable, ctx, "disable-api", invertBoolean)
	cfg.RestAPI.Port = setIntVal(cfg.RestAPI.Port, ctx, "api-port")
	cfg.RestAPI.HTTPS = setBoolVal(cfg.RestAPI.HTTPS, ctx, "rest-https")
	cfg.RestAPI.RestCertificate = setStringVal(cfg.RestAPI.RestCertificate, ctx, "rest-cert")
	cfg.RestAPI.RestKey = setStringVal(cfg.RestAPI.RestKey, ctx, "rest-key")
	cfg.RestAPI.RestAuth = setBoolVal(cfg.RestAPI.RestAuth, ctx, "rest-auth")
	cfg.RestAPI.RestAuthPassword = setStringVal(cfg.RestAPI.RestAuthPassword, ctx, "rest-auth-pwd")
	// next for the scheduler related flags
	cfg.Scheduler.WorkManagerQueueSize = setUIntVal(cfg.Scheduler.WorkManagerQueueSize, ctx, "work-manager-queue-size")
	cfg.Scheduler.WorkManagerPoolSize = setUIntVal(cfg.Scheduler.WorkManagerPoolSize, ctx, "work-manager-pool-size")
	// and finally for the tribe-related flags
	cfg.Tribe.Name = setStringVal(cfg.Tribe.Name, ctx, "tribe-node-name")
	cfg.Tribe.Enable = setBoolVal(cfg.Tribe.Enable, ctx, "tribe")
	cfg.Tribe.BindAddr = setStringVal(cfg.Tribe.BindAddr, ctx, "tribe-addr")
	cfg.Tribe.BindPort = setIntVal(cfg.Tribe.BindPort, ctx, "tribe-port")
	cfg.Tribe.Seed = setStringVal(cfg.Tribe.Seed, ctx, "tribe-seed")
}

func monitorErrors(ch <-chan error) {
	// Block on the error channel. Will return exit status 1 for an error or just return if the channel closes.
	err, ok := <-ch
	if !ok {
		return
	}
	log.Fatal(err)
}

// setMaxProcs configures runtime.GOMAXPROCS for snapd. GOMAXPROCS can be set by using
// the env variable GOMAXPROCS and snapd will honor this setting. A user can override the env
// variable by setting max-procs flag on the command line. Snapd will be limited to the max CPUs
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
				"_module":  "snapd",
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS to an invalid value")
		_maxProcs = 1
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "snapd",
				"maxprocs": _maxProcs,
			}).Warning("Setting GOMAXPROCS to 1")
		_maxProcs = 1
	} else if maxProcs <= numProcs {
		_maxProcs = maxProcs
	} else {
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "snapd",
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS larger than number of CPUs available on system")
		_maxProcs = numProcs
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  "snapd",
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
				"_module":        "snapd",
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
				"block":       "main",
				"_module":     "snapd",
				"snap-module": m.Name(),
			}).Info("module started")
	}
	return err
}

func printErrorAndExit(name string, err error) {
	log.WithFields(
		log.Fields{
			"block":       "main",
			"_module":     "snapd",
			"error":       err.Error(),
			"snap-module": name,
		}).Fatal("error starting module")
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
				"_module": "snapd",
			}).Info("shutting down modules")

		for _, m := range modules {
			log.WithFields(
				log.Fields{
					"block":       "main",
					"_module":     "snapd",
					"snap-module": m.Name(),
				}).Info("stopping module")
			m.Stop()
		}
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "snapd",
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

func validateLevelSettings(logLevel, pluginTrust int) {
	if logLevel < 1 || logLevel > 5 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "snapd",
				"level":   logLevel,
			}).Fatal("log level was invalid (needs: 1-5)")
	}
	if pluginTrust < 0 || pluginTrust > 2 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": "snapd",
				"level":   pluginTrust,
			}).Fatal("Plugin trust was invalid (needs: 0-2)")
	}
}
