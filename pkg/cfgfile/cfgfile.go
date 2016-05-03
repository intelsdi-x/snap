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

package cfgfile

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
)

// Read an input configuration file, parsing it (as YAML or JSON)
// into the input 'interface{}'', v
func Read(path string, v interface{}) error {
	// read bytes from file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// parse the byte-stream read (above); Note that we can parse both JSON
	// and YAML using the same yaml.Unmarshal() method since a JSON string is
	// a valid YAML string (but the converse is not true)
	if parseErr := yaml.Unmarshal(b, v); parseErr != nil {
		// remove any YAML-specific prefix that might have been added by then
		// yaml.Unmarshal() method or JSON-specific prefix that might have been
		// added if the resulting JSON string could not be marshalled into our
		// input interface correctly (note, if there is no match to either of
		// these prefixes then the error message will be passed through unchanged)
		tmpErr := strings.TrimPrefix(parseErr.Error(), "error converting YAML to JSON: yaml: ")
		errRet := strings.TrimPrefix(tmpErr, "error unmarshaling JSON: json: ")
		return fmt.Errorf("Error while parsing configuration file: %v", errRet)
	}
	return nil
}
