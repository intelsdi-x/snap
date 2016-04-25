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
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
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
type Config struct {
	Name                      string             `json:"name,omitempty"yaml:"name,omitempty"`
	Enable                    bool               `json:"enable,omitempty"yaml:"enable,omitempty"`
	BindAddr                  string             `json:"bind_addr,omitempty"yaml:"bind_addr,omitempty"`
	BindPort                  int                `json:"bind_port,omitempty"yaml:"bind_port,omitempty"`
	Seed                      string             `json:"seed,omitempty"yaml:"seed,omitempty"`
	MemberlistConfig          *memberlist.Config `json:"-"yaml:"-"`
	RestAPIProto              string             `json:"-"yaml:"-"`
	RestAPIPassword           string             `json:"-"yaml:"-"`
	RestAPIPort               int                `json:"-"yaml:"-"`
	RestAPIInsecureSkipVerify string             `json:"-"yaml:"-"`
}

// get the default snapd configuration
func GetDefaultConfig() *Config {
	mlCfg := memberlist.DefaultLANConfig()
	mlCfg.PushPullInterval = defaultPushPullInterval
	mlCfg.GossipNodes = mlCfg.GossipNodes * 2
	return &Config{
		Name:                      getHostname(),
		Enable:                    defaultEnable,
		BindAddr:                  getIP(),
		BindPort:                  defaultBindPort,
		Seed:                      defaultSeed,
		MemberlistConfig:          mlCfg,
		RestAPIProto:              defaultRestAPIProto,
		RestAPIPassword:           defaultRestAPIPassword,
		RestAPIPort:               defaultRestAPIPort,
		RestAPIInsecureSkipVerify: defaultRestAPIInsecureSkipVerify,
	}
}

// construct a new tribe Config from a hash map
func NewConfig(configMap map[string]interface{}) (*Config, error) {
	c := GetDefaultConfig()
	// set the Name value (if it was included in the input hash map)
	if v, ok := configMap["name"]; ok && v != nil {
		if str, ok := v.(string); ok {
			c.Name = str
		} else {
			return nil, fmt.Errorf("Error parsing 'name' from config; expected 'string' but found '%T'", v)
		}
	}
	// set the Enable value (if it was included in the input hash map)
	if v, ok := configMap["enable"]; ok && v != nil {
		if str, ok := v.(string); ok {
			boolVal, err := strconv.ParseBool(str)
			if err != nil {
				return nil, err
			}
			c.Enable = boolVal
		} else {
			return nil, fmt.Errorf("Error parsing 'enable' from config; expected 'string' but found '%T'", v)
		}
	}
	// set the BindAddr value (if it was included in the input hash map)
	if v, ok := configMap["bind_addr"]; ok && v != nil {
		if str, ok := v.(string); ok {
			c.BindAddr = str
		} else {
			return nil, fmt.Errorf("Error parsing 'bind_addr' from config; expected 'string' but found '%T'", v)
		}
	}
	// set the BindPort value (if it was included in the input hash map)
	if v, ok := configMap["bind_port"]; ok && v != nil {
		if val, ok := v.(json.Number); ok {
			tmpVal, err := val.Int64()
			if err != nil {
				return nil, err
			}
			c.BindPort = int(tmpVal)
		} else {
			return nil, fmt.Errorf("Error parsing 'bind_port' from config; expected 'json.Number' but found '%T'", v)
		}
	}
	// set the Seed value (if it was included in the input hash map)
	if v, ok := configMap["seed"]; ok && v != nil {
		if str, ok := v.(string); ok {
			c.Seed = str
		} else {
			return nil, fmt.Errorf("Error parsing 'seed' from config; expected 'string' but found '%T'", v)
		}
	}
	return c, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return uuid.New()
	}
	return hostname
}

func getIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.WithField("_block", "getIP").Error(err)
		return "127.0.0.1"
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.WithField("_block", "getIP").Error(err)
			return "127.0.0.1"
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPAddr:
				ip = v.IP
			case *net.IPNet:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}
