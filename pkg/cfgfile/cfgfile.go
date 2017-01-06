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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/xeipuuv/gojsonschema"
)

// we're wrapping a function from the underlying io/ioutil package here so
// that it's easier to mock when we define our *small* tests for this package...
// first, define an interface for reading a file (by path)
type reader interface {
	ReadFile(s string) ([]byte, error)
}

// then define a type (struct) that will implement the ReadFile method
type cfgReaderType struct{}

// then define the implementation for that type; it simply calls the
// underlying ioutil.Readfile() method and returns the result
func (r *cfgReaderType) ReadFile(s string) ([]byte, error) {
	return ioutil.ReadFile(s)
}

// then define a private variable that is of type reader (interface)
var cfgReader reader

// now, to make things even more testable, do the same for the process of
// validating the schema; first, create an interface to wrap the underlying
// method that will be used to validate the schema
type schemaValidator interface {
	validateSchema(schema, cfg string) []serror.SnapError
}

// then, define a type (struct) that will validate the underlying schema
type schemaValidatorType struct{}

// and define an implementation for that type that performs the schema validation
func (r *schemaValidatorType) validateSchema(schema, cfg string) []serror.SnapError {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	testDoc := gojsonschema.NewStringLoader(cfg)
	result, err := gojsonschema.Validate(schemaLoader, testDoc)
	var serrors []serror.SnapError
	// Check for invalid json
	if err != nil {
		serrors = append(serrors, serror.New(err))
		return serrors
	}
	// check if result passes validation
	if result.Valid() {
		return nil
	}
	for _, err := range result.Errors() {
		serr := serror.New(errors.New("Validate schema error"))
		serr.SetFields(map[string]interface{}{
			"value":       err.Value(),
			"context":     err.Context().String("::"),
			"description": err.Description(),
		})
		serrors = append(serrors, serr)
	}
	return serrors
}

// then define a private variable that is of type reader (interface)
var cfgValidator schemaValidator

// and finally, initialize to set the instance of that variable
func init() {
	cfgReader = &cfgReaderType{}
	cfgValidator = &schemaValidatorType{}
}

// Read an input configuration file, parsing it (as YAML or JSON)
// into the input 'interface{}', v
func Read(path string, v interface{}, schema string) []serror.SnapError {
	// read bytes from file
	b, err := cfgReader.ReadFile(path)
	if err != nil {
		return []serror.SnapError{serror.New(err)}
	}
	b = []byte(os.ExpandEnv(string(b)))
	// convert from YAML to JSON (remember, JSON is actually valid YAML)
	jb, err := yaml.YAMLToJSON(b)
	if err != nil {
		return []serror.SnapError{serror.New(fmt.Errorf("error converting YAML to JSON: %v", err))}
	}
	// validate the resulting JSON against the input the schema
	if errors := cfgValidator.validateSchema(schema, string(jb)); errors != nil {
		// if invalid, construct (and return?) a SnapError from the errors identified
		// during schema validation
		return errors
	}
	// if valid, parse the JSON byte-stream (above)
	if parseErr := json.Unmarshal(jb, v); parseErr != nil {
		// remove any YAML-specific prefix that might have been added by then
		// yaml.Unmarshal() method or JSON-specific prefix that might have been
		// added if the resulting JSON string could not be marshalled into our
		// input interface correctly (note, if there is no match to either of
		// these prefixes then the error message will be passed through unchanged)
		tmpErr := strings.TrimPrefix(parseErr.Error(), "error converting YAML to JSON: yaml: ")
		errRet := strings.TrimPrefix(tmpErr, "error unmarshaling JSON: json: ")
		return []serror.SnapError{serror.New(fmt.Errorf("Error while parsing configuration file: %v", errRet))}
	}
	return nil
}

// Validate an input JSON string against the input schema (and return a set of errors
// or nil if the input JSON string matches the constraints defined in that schema)
func ValidateSchema(schema, cfg string) []serror.SnapError {
	return cfgValidator.validateSchema(schema, cfg)
}
