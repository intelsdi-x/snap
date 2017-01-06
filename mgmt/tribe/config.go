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

package tribe

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/intelsdi-x/snap/pkg/netutil"
	"github.com/pborman/uuid"
)

// default configuration values
const (
	defaultEnable                    bool          = false
	defaultBindPort                  int           = 6000
	defaultSeed                      string        = ""
	defaultPushPullInterval          time.Duration = 300 * time.Second
	defaultRestAPIProto              string        = "http"
	defaultRestAPIPassword           string        = ""
	defaultRestAPIPort               int           = 8181
	defaultRestAPIInsecureSkipVerify string        = "true"
)

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	Name                      string             `json:"name"yaml:"name"`
	Enable                    bool               `json:"enable"yaml:"enable"`
	BindAddr                  string             `json:"bind_addr"yaml:"bind_addr"`
	BindPort                  int                `json:"bind_port"yaml:"bind_port"`
	Seed                      string             `json:"seed"yaml:"seed"`
	MemberlistConfig          *memberlist.Config `json:"-"yaml:"-"`
	RestAPIProto              string             `json:"-"yaml:"-"`
	RestAPIPassword           string             `json:"-"yaml:"-"`
	RestAPIPort               int                `json:"-"yaml:"-"`
	RestAPIInsecureSkipVerify string             `json:"-"yaml:"-"`
}

const (
	CONFIG_CONSTRAINTS = `
			"tribe": {
				"type": ["object", "null"],
				"properties": {
					"enable" : {
						"type": "boolean"
					},
					"bind_addr": {
						"type": "string"
					},
					"bind_port": {
						"type": "integer",
						"minimum": 0,
						"maximum": 65535
					},
					"name": {
						"type": "string"
					},
					"seed": {
						"type" : "string"
					}
				},
				"additionalProperties": false
			}
	`
)

// get the default snapteld configuration
func GetDefaultConfig() *Config {
	mlCfg := memberlist.DefaultLANConfig()
	mlCfg.PushPullInterval = defaultPushPullInterval
	mlCfg.GossipNodes = mlCfg.GossipNodes * 2
	return &Config{
		Name:                      getHostname(),
		Enable:                    defaultEnable,
		BindAddr:                  netutil.GetIP(),
		BindPort:                  defaultBindPort,
		Seed:                      defaultSeed,
		MemberlistConfig:          mlCfg,
		RestAPIProto:              defaultRestAPIProto,
		RestAPIPassword:           defaultRestAPIPassword,
		RestAPIPort:               defaultRestAPIPort,
		RestAPIInsecureSkipVerify: defaultRestAPIInsecureSkipVerify,
	}
}

// UnmarshalJSON unmarshals valid json into a Config.  An example Config can be found
// at github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (c *Config) UnmarshalJSON(data []byte) error {
	// construct a map of strings to json.RawMessages (to defer the parsing of individual
	// fields from the unmarshalled interface until later) and unmarshal the input
	// byte array into that map
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// loop through the individual map elements, parse each in turn, and set
	// the appropriate field in this configuration
	for k, v := range t {
		switch k {
		case "name":
			if err := json.Unmarshal(v, &(c.Name)); err != nil {
				return fmt.Errorf("%v (while parsing 'tribe::name')", err)
			}
		case "enable":
			if err := json.Unmarshal(v, &(c.Enable)); err != nil {
				return fmt.Errorf("%v (while parsing 'tribe::enable')", err)
			}
		case "bind_addr":
			if err := json.Unmarshal(v, &(c.BindAddr)); err != nil {
				return fmt.Errorf("%v (while parsing 'tribe::bind_addr')", err)
			}
		case "bind_port":
			if err := json.Unmarshal(v, &(c.BindPort)); err != nil {
				return fmt.Errorf("%v (while parsing 'tribe::bind_port')", err)
			}
		case "seed":
			if err := json.Unmarshal(v, &(c.Seed)); err != nil {
				return fmt.Errorf("%v (while parsing 'tribe::seed')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file while parsing 'tribe'", k)
		}
	}
	return nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return uuid.New()
	}
	return hostname
}
