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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/vrischmann/jsonutil"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest"
	"github.com/intelsdi-x/snap/mgmt/tribe"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/scheduler"
	"google.golang.org/grpc/grpclog"
)

const (
	commandLineErrorPrefix = "Command Line Error:"
	configFileErrorPrefix  = "ConfigFile Error:"
)

var (
	flMaxProcs = cli.StringFlag{
		Name:   "max-procs, c",
		Usage:  fmt.Sprintf("Set max cores to use for Snap Agent (default: %v)", defaultGoMaxProcs),
		EnvVar: "GOMAXPROCS",
	}
	// plugin
	flLogPath = cli.StringFlag{
		Name:   "log-path, o",
		Usage:  "Path for logs. Empty path logs to stdout.",
		EnvVar: "SNAP_LOG_PATH",
	}
	flLogTruncate = cli.BoolFlag{
		Name:  "log-truncate",
		Usage: "Log file truncating mode. Default is false => append (true => truncate).",
	}
	flLogColors = cli.BoolTFlag{
		Name:  "log-colors",
		Usage: "Log file coloring mode. Default is true => colored (--log-colors=false => no colors).",
	}
	flLogLevel = cli.StringFlag{
		Name:   "log-level, l",
		Usage:  fmt.Sprintf("1-5 (Debug, Info, Warning, Error, Fatal; default: %v)", defaultLogLevel),
		EnvVar: "SNAP_LOG_LEVEL",
	}
	flConfig = cli.StringFlag{
		Name:   "config",
		Usage:  "A path to a config file",
		EnvVar: "SNAP_CONFIG_PATH",
	}

	gitversion  string
	coreModules []coreModule

	// used to save a reference to the CLi App
	cliApp *cli.App

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
	defaultLogLevel    int    = 3
	defaultGoMaxProcs  int    = 1
	defaultLogPath     string = ""
	defaultLogTruncate bool   = false
	defaultLogColors   bool   = true
	defaultConfigPath  string = "/etc/snap/snapteld.conf"
)

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	LogLevel    int               `json:"log_level,omitempty"yaml:"log_level,omitempty"`
	GoMaxProcs  int               `json:"gomaxprocs,omitempty"yaml:"gomaxprocs,omitempty"`
	LogPath     string            `json:"log_path,omitempty"yaml:"log_path,omitempty"`
	LogTruncate bool              `json:"log_truncate,omitempty"yaml:"log_truncate,omitempty"`
	LogColors   bool              `json:"log_colors,omitempty"yaml:"log_colors,omitempty"`
	Control     *control.Config   `json:"control,omitempty"yaml:"control,omitempty"`
	Scheduler   *scheduler.Config `json:"scheduler,omitempty"yaml:"scheduler,omitempty"`
	RestAPI     *rest.Config      `json:"restapi,omitempty"yaml:"restapi,omitempty"`
	Tribe       *tribe.Config     `json:"tribe,omitempty"yaml:"tribe,omitempty"`
}

const (
	CONFIG_CONSTRAINTS = `{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "snapteld global config schema",
		"type": ["object", "null"],
		"properties": {
			"log_level": {
				"description": "log verbosity level for snapteld. Range between 1: debug, 2: info, 3: warning, 4: error, 5: fatal",
				"type": "integer",
				"minimum": 1,
				"maximum": 5
			},
			"log_path": {
				"description": "path to log file for snapteld to use",
				"type": "string"
			},
			"log_truncate": {
				"description": "truncate log file default is false",
				"type": "boolean"
			},
			"log_colors": {
				"description": "log file colored output default is true",
				"type": "boolean"
			},
			"gomaxprocs": {
				"description": "value to be used for gomaxprocs",
				"type": "integer",
				"minimum": 1
			},
			"control": { "$ref": "#/definitions/control" },
			"scheduler": { "$ref": "#/definitions/scheduler"},
			"restapi" : { "$ref": "#/definitions/restapi"},
			"tribe": { "$ref": "#/definitions/tribe"}
		},
		"additionalProperties": false,
		"definitions": { ` +
		control.CONFIG_CONSTRAINTS + `,` +
		scheduler.CONFIG_CONSTRAINTS + `,` +
		rest.CONFIG_CONSTRAINTS + `,` +
		tribe.CONFIG_CONSTRAINTS +
		`}` +
		`}`
	logModule = "snapteld"
)

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

type runtimeFlagsContext interface {
	String(key string) string
	Int(key string) int
	Bool(key string) bool
	IsSet(key string) bool
}

func main() {
	// Add a check to see if gitversion is blank from the build process

	if gitversion == "" {
		gitversion = "unknown"
	}

	cliApp = cli.NewApp()
	cliApp.Name = "snapteld"
	cliApp.Version = gitversion
	cliApp.Usage = "The open telemetry framework"
	cliApp.Flags = []cli.Flag{
		flLogLevel,
		flLogPath,
		flLogTruncate,
		flLogColors,
		flMaxProcs,
		flConfig,
	}
	cliApp.Flags = append(cliApp.Flags, control.Flags...)
	cliApp.Flags = append(cliApp.Flags, scheduler.Flags...)
	cliApp.Flags = append(cliApp.Flags, rest.Flags...)
	cliApp.Flags = append(cliApp.Flags, tribe.Flags...)

	cliApp.Action = action

	if cliApp.Run(os.Args) != nil {
		os.Exit(1)
	}
}

func action(ctx *cli.Context) error {
	// get default configuration
	cfg := getDefaultConfig()

	// read config file
	readConfig(cfg, ctx.String("config"))

	// apply values that may have been passed from the command line
	// to the configuration that we have built so far, overriding the
	// values that may have already been set (if any) for the
	// same variables in that configuration
	applyCmdLineFlags(cfg, ctx)

	// test the resulting configuration to ensure the values it contains still pass the
	// constraints after applying the environment variables and command-line parameters;
	// if errors are found, report them and exit with a fatal error
	jb, _ := json.Marshal(cfg)
	serrs := cfgfile.ValidateSchema(CONFIG_CONSTRAINTS, string(jb))
	if serrs != nil {
		for _, serr := range serrs {
			log.WithFields(serr.Fields()).Error(serr.Error())
		}
		log.Fatal("Errors found after applying command-line flags")
	}

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
		aMode := os.O_APPEND
		if cfg.LogTruncate {
			aMode = os.O_TRUNC
		}
		file, err := os.OpenFile(fmt.Sprintf("%s/snapteld.log", logPath), os.O_RDWR|os.O_CREATE|aMode, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// verify the temDirPath points to existing directory
	tempDirPath := cfg.Control.TempDirPath
	f, err := os.Stat(tempDirPath)
	if err != nil {
		log.Fatal(err)
	}
	if !f.IsDir() {
		log.Fatal("temp dir path provided must be a directory")
	}

	// Because even though github.com/sirupsen/logrus states that
	// 'Logs the event in colors if stdout is a tty, otherwise without colors'
	// Seems like this does not work
	// Please note however that the default output format without colors is somewhat different (timestamps, ...)
	//
	// We could also restrict this command line parameter to only apply when no logpath is given
	// and forcing the coloring to off when using a file but this might not please users who like to use
	// redirect mechanisms like # snapteld -t 0 -l 1 2>&1 | tee my.log
	if !cfg.LogColors {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	}

	// Validate log level and trust level settings for snapteld
	validateLevelSettings(cfg.LogLevel, cfg.Control.PluginTrust)

	// Switch log level to user defined
	log.SetLevel(getLevel(cfg.LogLevel))

	//Set standard logger as logger for grpc
	grpclog.SetLogger(log.StandardLogger())

	log.Info("setting log level to: ", l[cfg.LogLevel])

	log.Info("setting temp dir path to: ", tempDirPath)

	log.Info("Starting snapteld (version: ", gitversion, ")")

	// Set Max Processors for snapteld.
	setMaxProcs(cfg.GoMaxProcs)

	c := control.New(cfg.Control)
	if c.Config.AutoDiscoverPath != "" && c.Config.IsTLSEnabled() {
		log.Fatal("TLS security is not supported in autodiscovery mode")
	}

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
	if cfg.Tribe.Enable && c.Config.IsTLSEnabled() {
		log.Fatal("TLS security is not supported in tribe mode")
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
		coreModules = append(coreModules, r)
		log.Info("REST API is enabled")
	} else {
		log.Info("REST API is disabled")
	}

	// Set interrupt handling so we can either restart the app on a SIGHUP or
	// die gracefully when an interrupt, kill, etc. are received
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
					"_module": logModule,
				}).Fatal("need keyring file when trust is on (--keyring-file or -k)")
		}
		for _, k := range keyrings {
			keyringPath, err := filepath.Abs(k)
			if err != nil {
				log.WithFields(
					log.Fields{
						"block":       "main",
						"_module":     logModule,
						"error":       err.Error(),
						"keyringPath": keyringPath,
					}).Fatal("Unable to determine absolute path to keyring file")
			}
			f, err := os.Stat(keyringPath)
			if err != nil {
				log.WithFields(
					log.Fields{
						"block":       "main",
						"_module":     logModule,
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
							"_module":     logModule,
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
									"_module":     logModule,
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
							"_module":     logModule,
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

	log.WithFields(
		log.Fields{
			"block":   "main",
			"_module": logModule,
		}).Info("snapteld started", `
                                        ss                  ss
	                             odyssyhhyo         oyhysshhoyo
                                 ddddyssyyysssssyyyyyyyssssyyysssyhy+-
                           ssssssooosyhhysssyyyyyyyyyyyyyyyyyyyyyyyyyyyyssyhh+.
                          ssss lhyssssssyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyssydo
                sssssssssshhhhs lsyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyshh.
           ssyyyysssssssssssyhdo syyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyd.
       syyyyyyyyhhyyyyyyyyyyyyhdo syyyyyyyyyyyyydddhyyyyyyyyyyyyyhhhyyyyyyyyyyyyhh
     ssyyyyyyyh  hhyyyyyyyyyyyyhdo syyyyyyyyyyyddyddddhhhhhhhhdddhhddyyyyyyyyyyyydo
     ddyyyyyyh |  hyyyyyyyyyyyyydds syyyyyyyhhdhyyyhhhhhhhhhhyyyyyyhhdhyyyyyyyyyyym+
     dhyyyyyyyhhhhyyyyyyyyyyyyyyhmds shyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyydyyl
     dhddhyyyyyyhdhyyyyyyyyyyyyyydhmo yhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhdhsh
     dhyyhysshhdmdhyyyyyyyyyyyyyyhhdh  hhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhhdyoohmh ylo
      yy       dmyyyyyyyyyyyyyyyyhhhm  odhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhdhsoshdyyhdddy ylo
            odhhyyyyyyyyyyyyyyyyyyhhdy  oyhhhyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyhhmhyhdhhyyyyyhddoyddy ylo
           dddhhyyyyyyyyyyyyyyyhhhyhhdhs ooosydhhhhyyyyyyyyyyyyyyyyyyyyhhhhhyso+oymyyyyyyyyyyyhhddydmmyys
             ohdyyyyyyyyyyyyyyyhdhyyhhhddyoooohhhhhhhhhhhhhhhhhhhhhdhhyysooosyhhhhhyhhhhhhyyyhyyyhhhhddyy
                dyhyyyyyyyyyyyyydhyyyyyhhdddoooooooooooooooooooooooyysyyhddddhhyyyyydmdddddddmmddddhyyy
               dmmmmoNddddddmddhhhhyyyyyhhhdddddddhhhhhhhhhdddddddddhhhyyyyyyyyyyhNmmmooooooooyyy
                     Nhhhhhhhdmmddmhyyyyyyyyyhhhhhhhhhhhhhhhhhhhhyyyyyyyyyyyyyyyhhm
                     NhhhhhhhhhdmyhdyyyyyyyyyyyyhyyyyyyyyyyyyyyyhhhdhyyyyyyyyyyyhhN
                     NhyyyyyyyyyN dmyyyyyyyyyyhdmdhhhhhhhhhhdhhmmmmN NyyyyyyyyyyhhN
                     NhyyyyyyyyyN  Nyyyyyyyyyhhm               NmddmH Nyyyyyyyyyhdm
       .d8888b.      dmomomommmmh  dhhhhhhyyhhmh               NddddmH Nyyyyyyyyhdh
      d88P  Y88b                   dmomomommmmmh                dmomoH dmomomommmmh
      Y88b.
      "Y888b.   88888b.   8888b.  88888b.
         "Y88b. 888 "88b     "88b 888 "88b
           "888 888  888 .d888888 888  888
     Y88b  d88P 888  888 888  888 888 d88P
      "Y8888P"  888  888 "Y888888 88888P"
                                  888
                                  888
                                  888      `)

	select {} //run forever and ever
}

// get the default snapteld configuration
func getDefaultConfig() *Config {
	return &Config{
		LogLevel:    defaultLogLevel,
		GoMaxProcs:  defaultGoMaxProcs,
		LogPath:     defaultLogPath,
		LogTruncate: defaultLogTruncate,
		LogColors:   defaultLogColors,
		Control:     control.GetDefaultConfig(),
		Scheduler:   scheduler.GetDefaultConfig(),
		RestAPI:     rest.GetDefaultConfig(),
		Tribe:       tribe.GetDefaultConfig(),
	}
}

// Read the snapteld configuration from a configuration file
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

	serrs := cfgfile.Read(path, &cfg, CONFIG_CONSTRAINTS)
	if serrs != nil {
		for _, serr := range serrs {
			log.WithFields(serr.Fields()).Error(serr.Error())
		}
		log.Fatal("Errors found while parsing global configuration file")
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
func setBoolVal(field bool, ctx runtimeFlagsContext, flagName string, inverse ...bool) bool {
	// check to see if a value was set (either on the command-line or via the associated
	// environment variable, if any); if so, use that as value for the input field
	val := ctx.Bool(flagName)
	if ctx.IsSet(flagName) || val {
		field = val
		if len(inverse) > 0 {
			field = !field
		}
	}
	return field
}

func setStringVal(field string, ctx runtimeFlagsContext, flagName string) string {
	// check to see if a value was set (either on the command-line or via the associated
	// environment variable, if any); if so, use that as value for the input field
	val := ctx.String(flagName)
	if ctx.IsSet(flagName) || val != "" {
		field = val
	}
	return field
}

func setIntVal(field int, ctx runtimeFlagsContext, flagName string) int {
	// check to see if a value was set (either on the command-line or via the associated
	// environment variable, if any); if so, use that as value for the input field
	val := ctx.String(flagName)
	if ctx.IsSet(flagName) || val != "" {
		parsedField, err := strconv.Atoi(val)
		if err != nil {
			splitErr := strings.Split(err.Error(), ": ")
			errStr := splitErr[len(splitErr)-1]
			log.Fatal(fmt.Sprintf("Error Parsing %v; value '%v' cannot be parsed as an integer (%v)", flagName, val, errStr))
		}
		field = int(parsedField)
	}
	return field
}

func setUIntVal(field uint, ctx runtimeFlagsContext, flagName string) uint {
	// check to see if a value was set (either on the command-line or via the associated
	// environment variable, if any); if so, use that as value for the input field
	val := ctx.String(flagName)
	if ctx.IsSet(flagName) || val != "" {
		parsedField, err := strconv.Atoi(val)
		if err != nil {
			splitErr := strings.Split(err.Error(), ": ")
			errStr := splitErr[len(splitErr)-1]
			log.Fatal(fmt.Sprintf("Error Parsing %v; value '%v' cannot be parsed as an unsigned integer (%v)", flagName, val, errStr))
		}
		field = uint(parsedField)
	}
	return field
}

func setDurationVal(field time.Duration, ctx runtimeFlagsContext, flagName string) time.Duration {
	// check to see if a value was set (either on the command-line or via the associated
	// environment variable, if any); if so, use that as value for the input field
	val := ctx.String(flagName)
	if ctx.IsSet(flagName) || val != "" {
		parsedField, err := time.ParseDuration(val)
		if err != nil {
			splitErr := strings.Split(err.Error(), ": ")
			errStr := splitErr[len(splitErr)-1]
			log.Fatal(fmt.Sprintf("Error Parsing %v; value '%v' cannot be parsed as a duration (%v)", flagName, val, errStr))
		}
		field = parsedField
	}
	return field
}

// checks the input addr to see if it can be parsed as an IP address or used
// as a hostname. If both of those tests fail, then it returns an error, otherwise
// it returns a nil (no error) indicating that the addr string is a valid address
func isValidAddress(addr string, errPrefix string) error {
	parsedIP := net.ParseIP(addr)
	if parsedIP == nil {
		_, err := net.LookupHost(addr)
		if err != nil {
			errString := fmt.Sprintf("%s Address '%s' is not a valid address/hostname", errPrefix, addr)
			return errors.New(errString)
		}
	}
	return nil
}

// sanity checks the input address and port values to ensure they are set
// appropriately, specifically:
//
//	  - ensure that if the port value is set, the addr value does not also
//		include as part of the addr value (i.e. that the addr value is not a
//		string of the form IP_ADDR:PORT or HOSTNAME:PORT)
//	  - ensures that if a port is specified as part of the addr value, that port
//		string is not an empty string (i.e. that the ':' character is not the
//		last character in the addr value)
//	  - ensures that the address portion of the addr value can be either parsed
//		as an IP address or used as a hostname
//	  - ensures that the port detected as part of the addr value (if there is one)
//		can be parsed as an integer
//
// this function returns a boolean indicating whether or not a port number was
// found in the address and either nil or an error (depending on whether or not
// an error was detected while parsing the addr string)
func checkHostPortVals(addr string, port *int, errPrefix string) (bool, error) {
	portInAddrFlag := false
	// if the address field is empty, then we don't need to worry
	if len(addr) == 0 {
		return portInAddrFlag, nil
	}
	// If the input address ontains a comma, return an error
	if strings.Index(addr, ",") != -1 {
		errString := fmt.Sprintf("%s Invalid address; comma-separated IP address values are not supported", errPrefix)
		return false, errors.New(errString)
	}
	// check to see if the input address contains a colon or not
	idx := strings.Index(addr, ":")
	if idx == -1 {
		// if we don't find a colon in the address, then just try to parse it as
		// an IP address; return an error if we can't parse it successfully
		err := isValidAddress(addr, errPrefix)
		if err != nil {
			return false, err
		}
	} else if idx == (len(addr) - 1) {
		// if the last character is a colon character, then return an error because
		// while there is a colon, there's no value after that colon to represent
		// the port the RESTful APi should listen on
		errString := fmt.Sprintf("%s Empty port specified as part of API IP address", errPrefix)
		return false, errors.New(errString)
	} else {
		// otherwise attempt to split the input address into a host and port,
		// then try to parse resulting host string as an IP address; return
		// an error if the address cannot be split into a host and port or
		// the resulting host cannot be successfully parsed as an IP address
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return false, err
		}
		addrErr := isValidAddress(host, errPrefix)
		if addrErr != nil {
			return false, addrErr
		}
		portFromAddr, convErr := strconv.Atoi(portStr)
		if convErr != nil {
			errString := fmt.Sprintf("%s Port detected in address ('%s') cannot be parsed as an integer value", errPrefix, portStr)
			return false, errors.New(errString)
		}
		// if we get this far, and the port is also specified via the 'port' input
		// to this function, then return an error
		if *port > 0 {
			errString := fmt.Sprintf("%s Port can not be specified both as a port value and as part of address", errPrefix)
			return false, errors.New(errString)
		}
		// otherwise save the port that was parsed from the 'addr' as the 'port'
		// value before we return
		portInAddrFlag = true
		*port = portFromAddr
	}
	return portInAddrFlag, nil
}

// santiy check of the command-line flags to ensure that values are set
// appropriately; returns the port read from the command-line arguments, a flag
// indicating whether or not a port was detected in the address read from the
// command-line arguments, and an error if one is detected
func checkCmdLineFlags(ctx runtimeFlagsContext) (int, bool, error) {
	tlsCert := ctx.String("tls-cert")
	tlsKey := ctx.String("tls-key")
	if _, err := checkTLSEnabled(tlsCert, tlsKey, commandLineErrorPrefix); err != nil {
		return -1, false, err
	}
	// Check to see if the API address is specified (either via the CLI or through
	// the associated environment variable); if so, grab the port and check that the
	// address and port against the constraints (above)
	addr := ctx.String("api-addr")
	port := ctx.Int("api-port")
	if ctx.IsSet("api-addr") || addr != "" {
		portInAddr, err := checkHostPortVals(addr, &port, commandLineErrorPrefix)
		if err != nil {
			return -1, portInAddr, err
		}
		return port, portInAddr, nil
	}

	return port, false, nil
}

// santiy check of the configuration file parameters to ensure that values are set
// appropriately; returns the port read from the global configuration file, a flag
// indicating whether or not a port was detected in the address read from the
// global configuration file, and an error if one is detected
func checkCfgSettings(cfg *Config) (int, bool, error) {
	tlsCert := cfg.Control.TLSCertPath
	tlsKey := cfg.Control.TLSKeyPath
	if _, err := checkTLSEnabled(tlsCert, tlsKey, configFileErrorPrefix); err != nil {
		return -1, false, err
	}
	addr := cfg.RestAPI.Address
	var port int
	if cfg.RestAPI.PortSetByConfigFile() {
		port = cfg.RestAPI.Port
	} else {
		port = -1
	}
	portInAddr, err := checkHostPortVals(addr, &port, configFileErrorPrefix)
	if err != nil {
		return -1, portInAddr, err
	}
	return port, portInAddr, nil
}

func checkTLSEnabled(certPath, keyPath, errPrefix string) (tlsEnabled bool, err error) {
	if certPath != "" && keyPath != "" {
		return true, nil
	}
	if certPath != "" || keyPath != "" {
		return false, fmt.Errorf("%s certificate and key path must be given both or none", errPrefix)
	}
	return false, nil
}

// Apply the command line flags set (if any) to override the values
// in the input configuration
func applyCmdLineFlags(cfg *Config, ctx runtimeFlagsContext) {
	// check the settings for the command-line arguments included in the cli.Context
	cmdLinePort, cmdLinePortInAddr, cmdLineErr := checkCmdLineFlags(ctx)
	if cmdLineErr != nil {
		log.Fatal(cmdLineErr)
	}
	// check the settings in the input configuration (and return an error if any issues are found)
	cfgFilePort, cfgFilePortInAddr, cfgFileErr := checkCfgSettings(cfg)
	if cfgFileErr != nil {
		log.Fatal(cfgFileErr)
	}
	invertBoolean := true
	// apply any command line flags that might have been set, first for the
	// snapteld-related flags
	cfg.GoMaxProcs = setIntVal(cfg.GoMaxProcs, ctx, "max-procs")
	cfg.LogLevel = setIntVal(cfg.LogLevel, ctx, "log-level")
	cfg.LogPath = setStringVal(cfg.LogPath, ctx, "log-path")
	cfg.LogTruncate = setBoolVal(cfg.LogTruncate, ctx, "log-truncate")
	cfg.LogColors = setBoolVal(cfg.LogColors, ctx, "log-colors")
	// next for the flags related to the control package
	cfg.Control.MaxRunningPlugins = setIntVal(cfg.Control.MaxRunningPlugins, ctx, "max-running-plugins")
	cfg.Control.PluginLoadTimeout = setIntVal(cfg.Control.PluginLoadTimeout, ctx, "plugin-load-timeout")
	cfg.Control.PluginTrust = setIntVal(cfg.Control.PluginTrust, ctx, "plugin-trust")
	cfg.Control.AutoDiscoverPath = setStringVal(cfg.Control.AutoDiscoverPath, ctx, "auto-discover")
	cfg.Control.KeyringPaths = setStringVal(cfg.Control.KeyringPaths, ctx, "keyring-paths")
	cfg.Control.CacheExpiration = jsonutil.Duration{setDurationVal(cfg.Control.CacheExpiration.Duration, ctx, "cache-expiration")}
	cfg.Control.ListenAddr = setStringVal(cfg.Control.ListenAddr, ctx, "control-listen-addr")
	cfg.Control.ListenPort = setIntVal(cfg.Control.ListenPort, ctx, "control-listen-port")
	cfg.Control.Pprof = setBoolVal(cfg.Control.Pprof, ctx, "pprof")
	cfg.Control.TempDirPath = setStringVal(cfg.Control.TempDirPath, ctx, "temp_dir_path")
	cfg.Control.TLSCertPath = setStringVal(cfg.Control.TLSCertPath, ctx, "tls-cert")
	cfg.Control.TLSKeyPath = setStringVal(cfg.Control.TLSKeyPath, ctx, "tls-key")
	cfg.Control.CACertPaths = setStringVal(cfg.Control.CACertPaths, ctx, "ca-cert-paths")
	// next for the RESTful server related flags
	cfg.RestAPI.Enable = setBoolVal(cfg.RestAPI.Enable, ctx, "disable-api", invertBoolean)
	cfg.RestAPI.Port = setIntVal(cfg.RestAPI.Port, ctx, "api-port")
	cfg.RestAPI.Address = setStringVal(cfg.RestAPI.Address, ctx, "api-addr")
	cfg.RestAPI.HTTPS = setBoolVal(cfg.RestAPI.HTTPS, ctx, "rest-https")
	cfg.RestAPI.RestCertificate = setStringVal(cfg.RestAPI.RestCertificate, ctx, "rest-cert")
	cfg.RestAPI.RestKey = setStringVal(cfg.RestAPI.RestKey, ctx, "rest-key")
	cfg.RestAPI.RestAuth = setBoolVal(cfg.RestAPI.RestAuth, ctx, "rest-auth")
	cfg.RestAPI.RestAuthPassword = setStringVal(cfg.RestAPI.RestAuthPassword, ctx, "rest-auth-pwd")
	cfg.RestAPI.Pprof = setBoolVal(cfg.RestAPI.Pprof, ctx, "pprof")
	cfg.RestAPI.Corsd = setStringVal(cfg.RestAPI.Corsd, ctx, "allowed_origins")

	// next for the scheduler related flags
	cfg.Scheduler.WorkManagerQueueSize = setUIntVal(cfg.Scheduler.WorkManagerQueueSize, ctx, "work-manager-queue-size")
	cfg.Scheduler.WorkManagerPoolSize = setUIntVal(cfg.Scheduler.WorkManagerPoolSize, ctx, "work-manager-pool-size")
	// and finally for the tribe-related flags
	cfg.Tribe.Name = setStringVal(cfg.Tribe.Name, ctx, "tribe-node-name")
	cfg.Tribe.Enable = setBoolVal(cfg.Tribe.Enable, ctx, "tribe")
	cfg.Tribe.BindAddr = setStringVal(cfg.Tribe.BindAddr, ctx, "tribe-addr")
	cfg.Tribe.BindPort = setIntVal(cfg.Tribe.BindPort, ctx, "tribe-port")
	cfg.Tribe.Seed = setStringVal(cfg.Tribe.Seed, ctx, "tribe-seed")
	// check to see if we have duplicate port definitions (check the various
	// combinations of the config file and command-line parameter values that
	// could be used to define the port and make sure we only have one)
	if cmdLinePort > 0 && cfgFilePortInAddr && !cmdLinePortInAddr {
		log.Fatal("Usage Error: Port can not be specified both as a port value on the CLI and as part of address in the global config file")
	} else if cfgFilePort > 0 && cmdLinePortInAddr && !cfgFilePortInAddr {
		log.Fatal("Usage Error: Port can not be specified both as a port value in the global config file and as part of address on the CLI")
	}
	// if we retrieved the port from an address, then use that value as the
	// cfg.RestAPI.Port value (so that we can validate it against the constraints
	// placed on the value for this parameter and ensure that the port in the
	// address complies with those constraints)
	if cmdLinePortInAddr {
		cfg.RestAPI.Port = cmdLinePort
	} else if cfgFilePortInAddr {
		cfg.RestAPI.Port = cfgFilePort
	} else {
		// if get to here, then there is no port number in the input address
		// (regardless of whether it came from the default configuration, configuration
		// file, an environment variable, or a command-line flag); in that case we should
		// set the address in the RestAPI configuration to be the current address and port
		// (separated by a ':')
		cfg.RestAPI.Address = fmt.Sprintf("%v:%v", cfg.RestAPI.Address, cfg.RestAPI.Port)
	}
}

func monitorErrors(ch <-chan error) {
	// Block on the error channel. Will return exit status 1 for an error or just return if the channel closes.
	err, ok := <-ch
	if !ok {
		return
	}
	log.Fatal(err)
}

// setMaxProcs configures runtime.GOMAXPROCS for snapteld. GOMAXPROCS can be set by using
// the env variable GOMAXPROCS and snapteld will honor this setting. A user can override the env
// variable by setting max-procs flag on the command line. snapteld will be limited to the max CPUs
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
				"_module":  logModule,
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS to an invalid value")
		_maxProcs = 1
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  logModule,
				"maxprocs": _maxProcs,
			}).Warning("Setting GOMAXPROCS to 1")
		_maxProcs = 1
	} else if maxProcs <= numProcs {
		_maxProcs = maxProcs
	} else {
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  logModule,
				"maxprocs": maxProcs,
			}).Error("Trying to set GOMAXPROCS larger than number of CPUs available on system")
		_maxProcs = numProcs
		log.WithFields(
			log.Fields{
				"_block":   "main",
				"_module":  logModule,
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
				"_module":        logModule,
				"given maxprocs": _maxProcs,
				"real maxprocs":  actualNumProcs,
			}).Warning("not using given maxprocs")
	}
}

// UnmarshalJSON unmarshals valid json into a Config.  An example Config can be found
// at github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (c *Config) UnmarshalJSON(data []byte) error {
	// construct a map of strings to json.RawMessages (to defer the parsing of individual
	// fields from the unmarshalled interface until later), then unmarshal the input
	// byte array into that map
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// loop through the individual map elements, parse each in turn, and set
	// the appropriate field in this configuration
	for k, v := range t {
		switch k {
		case "log_level":
			if err := json.Unmarshal(v, &(c.LogLevel)); err != nil {
				return fmt.Errorf("%v (while parsing 'log_level')", err)
			}
		case "gomaxprocs":
			if err := json.Unmarshal(v, &(c.GoMaxProcs)); err != nil {
				return fmt.Errorf("%v (while parsing 'gomaxprocs')", err)
			}
		case "log_path":
			if err := json.Unmarshal(v, &(c.LogPath)); err != nil {
				return fmt.Errorf("%v (while parsing 'log_path')", err)
			}
		case "log_truncate":
			if err := json.Unmarshal(v, &(c.LogTruncate)); err != nil {
				return fmt.Errorf("%v (while parsing 'log_truncate')", err)
			}
		case "log_colors":
			if err := json.Unmarshal(v, &(c.LogColors)); err != nil {
				return fmt.Errorf("%v (while parsing 'log_colors')", err)
			}
		case "control":
			if err := json.Unmarshal(v, c.Control); err != nil {
				return err
			}
		case "restapi":
			if err := json.Unmarshal(v, c.RestAPI); err != nil {
				return err
			}
		case "scheduler":
			if err := json.Unmarshal(v, c.Scheduler); err != nil {
				return err
			}
		case "tribe":
			if err := json.Unmarshal(v, c.Tribe); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file", k)
		}
	}
	return nil
}

func startModule(m coreModule) error {
	err := m.Start()
	if err == nil {
		log.WithFields(
			log.Fields{
				"block":       "main",
				"_module":     logModule,
				"snap-module": m.Name(),
			}).Info("module started")
	}
	return err
}

func printErrorAndExit(name string, err error) {
	log.WithFields(
		log.Fields{
			"block":       "main",
			"_module":     logModule,
			"error":       err.Error(),
			"snap-module": name,
		}).Fatal("error starting module")
}

func startInterruptHandling(modules ...coreModule) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)

	//Let's block until someone tells us to quit
	go func() {
		sig := <-c
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": logModule,
			}).Info("shutting down modules")

		for _, m := range modules {
			log.WithFields(
				log.Fields{
					"block":       "main",
					"_module":     logModule,
					"snap-module": m.Name(),
				}).Info("stopping module")
			m.Stop()
		}
		if sig == syscall.SIGHUP {
			// log the action we're taking (restarting the app)
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": logModule,
					"signal":  sig.String(),
				}).Info("restarting app")
			// and restart the app (with the current configuration)
			err := cliApp.Run(os.Args)
			if err != nil {
				os.Exit(1)
			}
		} else {
			log.WithFields(
				log.Fields{
					"block":   "main",
					"_module": logModule,
					"signal":  sig.String(),
				}).Info("exiting on signal")
			os.Exit(0)
		}
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
				"_module": logModule,
				"level":   logLevel,
			}).Fatal("log level was invalid (needs: 1-5)")
	}
	if pluginTrust < 0 || pluginTrust > 2 {
		log.WithFields(
			log.Fields{
				"block":   "main",
				"_module": logModule,
				"level":   pluginTrust,
			}).Fatal("Plugin trust was invalid (needs: 0-2)")
	}
}
