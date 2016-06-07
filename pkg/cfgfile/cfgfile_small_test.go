// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	"testing"

	"github.com/ghodss/yaml"
	"github.com/intelsdi-x/snap/core/serror"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	MOCK_CONSTRAINTS = `{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "test schema",
		"type": ["object", "null"],
		"properties": {
			"Foo": {
				"type": "string"
			},
			"Bar": {
				"type": "string"
			}
		}
	}
	`
	INVALID_MOCK_CONSTRAINTS = `{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "test schema",
		"type": ["object", "null"],
		"properties": {
			"Foo": {
				"type": "boolean"
			},
			"Bar": {
				"type": "string"
			}
		}
	}
	`
)

type testConfig struct {
	Foo string
	Bar string
}

// create an enumeration of the response types we expect back from
// our mockReader
const (
	JSON = iota
	YAML
	FILE_NOT_FOUND
	INVALID_YAML
	UNMATCHED_SCHEMA
)

// create a type for the entries in our test-table
type entry struct {
	out []byte
	err error
}

// next, define a function we can use to generate new test-table entries
func newEntry(b []byte, e error) entry {
	return entry{b, e}
}

// then, define our test-table as a map of integers to entry values
var testTable = map[int]entry{
	JSON:             newEntry(json.Marshal(testConfig{"Tom", "Justin"})),
	YAML:             newEntry(yaml.Marshal(testConfig{"Tom", "Justin"})),
	FILE_NOT_FOUND:   newEntry(nil, errors.New("File not found")),
	INVALID_YAML:     newEntry([]byte("not:\tvalid: YAML"), nil),
	UNMATCHED_SCHEMA: newEntry(json.Marshal(map[string]int{"Foo": 1, "Bar": 2})),
}

// create a mock reader that implements the cfgfile.reader interface
type mockReader struct {
	vals entry
}

// and redefine the ReadFile method for our mock type
func (r *mockReader) ReadFile(s string) ([]byte, error) {
	return r.vals.out, r.vals.err
}

// create a mock schema validator that implements the cfgfile.schemaValidator interface
type mockSchemaValidator struct {
	returnError bool
}

// and redefine the validateSchema method to always return nil (no errors found)
func (r *mockSchemaValidator) validateSchema(schema, cfg string) []serror.SnapError {
	if r.returnError {
		return []serror.SnapError{serror.New(errors.New("Invalid schema"))}
	}
	return nil
}

func TestReadConfig(t *testing.T) {

	Convey("Unmarshal yaml file", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[YAML]}
		cfgValidator = &mockSchemaValidator{false}
		err := Read("/tmp/dummy.yaml", &config, MOCK_CONSTRAINTS)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})

	Convey("Unmarshal json file", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[JSON]}
		cfgValidator = &mockSchemaValidator{false}
		err := Read("/tmp/dummy.json", &config, MOCK_CONSTRAINTS)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})

	Convey("Throw file not found error", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[FILE_NOT_FOUND]}
		cfgValidator = &mockSchemaValidator{false}
		err := Read("/tmp/dummy.yaml", &config, MOCK_CONSTRAINTS)
		So(err, ShouldNotBeNil)
		So(err[0].Error(), ShouldResemble, "File not found")
	})

	Convey("Throw invalid YAML error", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[INVALID_YAML]}
		cfgValidator = &mockSchemaValidator{false}
		err := Read("/tmp/dummy.yaml", &config, MOCK_CONSTRAINTS)
		So(err, ShouldNotBeNil)
		So(err[0].Error(), ShouldStartWith, "error converting YAML to JSON")
	})

	Convey("Throw schema validation error", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[JSON]}
		cfgValidator = &mockSchemaValidator{true}
		err := Read("/tmp/dummy.yaml", &config, MOCK_CONSTRAINTS)
		So(err, ShouldNotBeNil)
		So(err[0].Error(), ShouldResemble, "Invalid schema")
	})

	Convey("Throw unmatched schema error", t, func() {
		config := testConfig{}
		cfgReader = &mockReader{testTable[UNMATCHED_SCHEMA]}
		cfgValidator = &mockSchemaValidator{false}
		err := Read("/tmp/dummy.yaml", &config, MOCK_CONSTRAINTS)
		So(err, ShouldNotBeNil)
		So(err[0].Error(), ShouldStartWith, "Error while parsing configuration file")
	})
}

func TestValidateSchema(t *testing.T) {

	Convey("Test valid schema", t, func() {
		config := testConfig{"Tom", "Justin"}
		cfgValidator = &schemaValidatorType{}
		jb, _ := json.Marshal(config)
		errs := ValidateSchema(MOCK_CONSTRAINTS, string(jb))
		So(errs, ShouldBeNil)
	})

	Convey("Test invalid schema", t, func() {
		config := testConfig{"Tom", "Justin"}
		cfgValidator = &schemaValidatorType{}
		jb, _ := json.Marshal(config)
		errs := ValidateSchema(INVALID_MOCK_CONSTRAINTS, string(jb))
		So(errs, ShouldNotBeNil)
	})

	Convey("Test invalid json", t, func() {
		config := testConfig{"Tom", "Justin"}
		cfgValidator = &schemaValidatorType{}
		yb, _ := yaml.Marshal(config)
		errs := ValidateSchema(MOCK_CONSTRAINTS, string(yb))
		So(errs, ShouldNotBeNil)
	})
}
