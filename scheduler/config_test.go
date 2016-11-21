// +build legacy

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
	"testing"

	"github.com/intelsdi-x/snap/pkg/cfgfile"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	MOCK_CONSTRAINTS = `{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "snapteld global config schema",
		"type": ["object", "null"],
		"properties": {
			"control": { "$ref": "#/definitions/control" },
			"scheduler": { "$ref": "#/definitions/scheduler"},
			"restapi" : { "$ref": "#/definitions/restapi"},
			"tribe": { "$ref": "#/definitions/tribe"}
		},
		"additionalProperties": true,
		"definitions": { ` +
		`"control": {}, ` + CONFIG_CONSTRAINTS + `, "restapi": {}, "tribe":{}` +
		`}` +
		`}`
)

type mockConfig struct {
	Scheduler *Config
}

func TestSchedulerConfigJSON(t *testing.T) {
	config := &mockConfig{
		Scheduler: GetDefaultConfig(),
	}
	path := "../examples/configs/snap-config-sample.json"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Scheduler
	}
	Convey("Provided a valid config in JSON", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("WorkManagerQueueSize should equal 10", func() {
			So(cfg.WorkManagerQueueSize, ShouldEqual, 10)
		})
		Convey("WorkManagerPoolSize should equal 2", func() {
			So(cfg.WorkManagerPoolSize, ShouldEqual, 2)
		})
	})

}

func TestSchedulerConfigYaml(t *testing.T) {
	config := &mockConfig{
		Scheduler: GetDefaultConfig(),
	}
	path := "../examples/configs/snap-config-sample.yaml"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Scheduler
	}
	Convey("Provided a valid config in YAML", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("WorkManagerQueueSize should equal 10", func() {
			So(cfg.WorkManagerQueueSize, ShouldEqual, 10)
		})
		Convey("WorkManagerPoolSize should equal 2", func() {
			So(cfg.WorkManagerPoolSize, ShouldEqual, 2)
		})
	})

}

func TestSchedulerDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	Convey("Provided a default config", t, func() {
		Convey("WorkManagerQueueSize should equal 25", func() {
			So(cfg.WorkManagerQueueSize, ShouldEqual, 25)
		})
		Convey("WorkManagerPoolSize should equal 4", func() {
			So(cfg.WorkManagerPoolSize, ShouldEqual, 4)
		})
	})
}
