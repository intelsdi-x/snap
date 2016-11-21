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

package scheduler

import (
	"encoding/json"
	"fmt"
)

// default configuration values
const (
	defaultWorkManagerQueueSize uint = 25
	defaultWorkManagerPoolSize  uint = 4
)

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	WorkManagerQueueSize uint `json:"work_manager_queue_size"yaml:"work_manager_queue_size"`
	WorkManagerPoolSize  uint `json:"work_manager_pool_size"yaml:"work_manager_pool_size"`
}

const (
	CONFIG_CONSTRAINTS = `
			"scheduler": {
				"type": ["object", "null"],
				"properties" : {
					"work_manager_queue_size" : {
						"type": "integer",
						"minimum": 1
					},
					"work_manager_pool_size" : {
						"type": "integer",
						"minimum": 1
					}
				},
				"additionalProperties": false
			}
	`
)

// get the default snapteld configuration
func GetDefaultConfig() *Config {
	return &Config{
		WorkManagerQueueSize: defaultWorkManagerQueueSize,
		WorkManagerPoolSize:  defaultWorkManagerPoolSize,
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
		case "work_manager_queue_size":
			if err := json.Unmarshal(v, &(c.WorkManagerQueueSize)); err != nil {
				return fmt.Errorf("%v (while parsing 'scheduler::work_manager_queue_size')", err)
			}
		case "work_manager_pool_size":
			if err := json.Unmarshal(v, &(c.WorkManagerPoolSize)); err != nil {
				return fmt.Errorf("%v (while parsing 'scheduler::work_manager_pool_size')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file while parsing 'scheduler'", k)
		}
	}
	return nil
}
