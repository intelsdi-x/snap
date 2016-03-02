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

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return uuid.New()
	}
	return hostname
}
