// +build legacy

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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
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
)

type testConfig struct {
	Foo string
	Bar string
}

func writeYaml(tc testConfig) string {
	b, err := yaml.Marshal(tc)
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile("", "yaml.conf")
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(b); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fp, err := filepath.Abs(f.Name())
	if err != nil {
		panic(err)
	}
	return fp
}

func writeJson(tc testConfig) string {
	b, err := json.Marshal(tc)
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile("", "json.conf")
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(b); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fp, err := filepath.Abs(f.Name())
	if err != nil {
		panic(err)
	}
	return fp
}

func TestReadConfig(t *testing.T) {
	tc := testConfig{"Tom", "Justin"}
	yamlFile := writeYaml(tc)
	jsonFile := writeJson(tc)

	Convey("Unmarshal yaml file", t, func() {
		config := testConfig{}
		err := Read(yamlFile, &config, MOCK_CONSTRAINTS)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})

	Convey("Unmarshal json file", t, func() {
		config := testConfig{}
		err := Read(jsonFile, &config, MOCK_CONSTRAINTS)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})
}
