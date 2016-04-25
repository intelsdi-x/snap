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
type Config struct {
	WorkManagerQueueSize uint `json:"work_manager_queue_size,omitempty"yaml:"work_manager_queue_size,omitempty"`
	WorkManagerPoolSize  uint `json:"work_manager_pool_size,omitempty"yaml:"work_manager_pool_size,omitempty"`
}

// get the default snapd configuration
func GetDefaultConfig() *Config {
	return &Config{
		WorkManagerQueueSize: defaultWorkManagerQueueSize,
		WorkManagerPoolSize:  defaultWorkManagerPoolSize,
	}
}

// construct a new scheduler Config from a hash map
func NewConfig(configMap map[string]interface{}) (*Config, error) {
	c := GetDefaultConfig()
	// set the WorkManagerQueueSize value (if it was included in the input hash map)
	if v, ok := configMap["work_manager_queue_size"]; ok && v != nil {
		if val, ok := v.(json.Number); ok {
			tmpVal, err := val.Int64()
			if err != nil {
				return nil, err
			}
			c.WorkManagerQueueSize = uint(tmpVal)
		} else {
			return nil, fmt.Errorf("Error parsing 'work_manager_queue_size' from config; expected 'json.Number' but found '%T'", v)
		}
	}
	// set the WorkManagerPoolSize value (if it was included in the input hash map)
	if v, ok := configMap["work_manager_pool_size"]; ok && v != nil {
		if val, ok := v.(json.Number); ok {
			tmpVal, err := val.Int64()
			if err != nil {
				return nil, err
			}
			c.WorkManagerPoolSize = uint(tmpVal)
		} else {
			return nil, fmt.Errorf("Error parsing 'work_manager_pool_size' from config; expected 'json.Number' but found '%T'", v)
		}
	}
	return c, nil
}
