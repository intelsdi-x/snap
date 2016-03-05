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

package globalconfig

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

type config struct {
	Flags flagConfig `json:"flags"`
}

// FlagConfig struct has all of the snapd flags
type flagConfig struct {
	LogPath          *string `json:"log-path"`
	LogLevel         *int    `json:"log-level"`
	MaxProcs         *int    `json:"max-procs"`
	DisableAPI       *bool   `json:"disable-api"`
	APIPort          *int    `json:"api-port"`
	AutodiscoverPath *string `json:"auto-discover"`
	MaxRunning       *int    `json:"max-running-plugins"`
	PluginTrust      *int    `json:"plugin-trust"`
	KeyringPaths     *string `json:"keyring-paths"`
	Cachestr         *string `json:"cache-expiration"`
	IsTribeEnabled   *bool   `json:"tribe"`
	TribeSeed        *string `json:"tribe-seed"`
	TribeNodeName    *string `json:"tribe-node-name"`
	TribeAddr        *string `json:"tribe-addr"`
	TribePort        *int    `json:"tribe-port"`
	RestHTTPS        *bool   `json:"rest-https"`
	RestKey          *string `json:"rest-key"`
	RestCert         *string `json:"rest-cert"`
	RestAuth         *bool   `json:"rest-auth"`
	RestAuthPwd      *string `json:"rest-auth-pwd"`
}

// NewConfig returns a reference to a global config type for the snap daemon
func NewConfig() *config {
	return &config{}
}

func (f *config) LoadConfig(b []byte, cfg map[string]interface{}) {
	if _, ok := cfg["flags"]; ok {
		err := json.Unmarshal(b, &f)
		if err != nil {
			log.WithFields(log.Fields{
				"block":   "LoadConfig",
				"_module": "globalconfig",
				"error":   err.Error(),
			}).Fatal("invalid config")
		}
	}
}

// GetFlagInt eturns the integer value for the flag to be used by snapd
func GetFlagInt(ctx *cli.Context, cfgVal *int, flag string) int {
	// Checks if the flag is in the config and if the command line flag is not set
	if cfgVal != nil && !ctx.IsSet(flag) {
		return *cfgVal
	}
	return ctx.Int(flag)
}

// GetFlagBool returns the boolean value for the flag to be used by snapd
func GetFlagBool(ctx *cli.Context, cfgVal *bool, flag string) bool {
	// Checks if the flag is in the config and if the command line flag is not set
	if cfgVal != nil && !ctx.IsSet(flag) {
		return *cfgVal
	}
	return ctx.Bool(flag)
}

// GetFlagString returns the string value for the flag to be used by snapd
func GetFlagString(ctx *cli.Context, cfgVal *string, flag string) string {
	// Checks if the flag is in the config and if the command line flag is not set
	if cfgVal != nil && !ctx.IsSet(flag) {
		return *cfgVal
	}
	return ctx.String(flag)
}
