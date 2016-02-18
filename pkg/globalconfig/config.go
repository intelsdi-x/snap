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
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

// Read reads the JSON file and unmarshals it into a map[string]interface{}
func Read(path string) ([]byte, map[string]interface{}) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithFields(log.Fields{
			"block":   "Read",
			"_module": "globalconfig",
			"error":   err.Error(),
			"path":    path,
		}).Fatal("unable to read config")
	}
	var cfg map[string]interface{}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"block":   "Read",
			"_module": "globalconfig",
			"error":   err.Error(),
		}).Fatal("invalid config")
	}
	return b, cfg
}
