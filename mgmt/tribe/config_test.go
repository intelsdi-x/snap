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

package tribe

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/pkg/netutil"
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
		`"control": {}, "scheduler": {}, "restapi":{}, ` + CONFIG_CONSTRAINTS +
		`}` +
		`}`
)

type mockConfig struct {
	Tribe *Config
}

func TestTribeConfigJSON(t *testing.T) {
	config := &mockConfig{
		Tribe: GetDefaultConfig(),
	}
	path := "../../examples/configs/snap-config-sample.json"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Tribe
	}
	Convey("Provided a valid config in JSON", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("Enable should be true", func() {
			So(cfg.Enable, ShouldEqual, true)
		})
		Convey("BindAddr should equal 127.0.0.1", func() {
			So(cfg.BindAddr, ShouldEqual, "127.0.0.1")
		})
		Convey("BindPort should be 16000", func() {
			So(cfg.BindPort, ShouldEqual, 16000)
		})
		Convey("Name should equal localhost", func() {
			So(cfg.Name, ShouldEqual, "localhost")
		})
		Convey("Seed should be 1.1.1.1:16000", func() {
			So(cfg.Seed, ShouldEqual, "1.1.1.1:16000")
		})
	})

}

func TestTribeConfigYaml(t *testing.T) {
	config := &mockConfig{
		Tribe: GetDefaultConfig(),
	}
	path := "../../examples/configs/snap-config-sample.yaml"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Tribe
	}
	Convey("Provided a valid config in YAML", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("Enable should be true", func() {
			So(cfg.Enable, ShouldEqual, true)
		})
		Convey("BindAddr should equal 127.0.0.1", func() {
			So(cfg.BindAddr, ShouldEqual, "127.0.0.1")
		})
		Convey("BindPort should be 16000", func() {
			So(cfg.BindPort, ShouldEqual, 16000)
		})
		Convey("Name should equal localhost", func() {
			So(cfg.Name, ShouldEqual, "localhost")
		})
		Convey("Seed should be 1.1.1.1:16000", func() {
			So(cfg.Seed, ShouldEqual, "1.1.1.1:16000")
		})
	})

}

func TestTribeDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	Convey("Provided a default tribe config", t, func() {
		Convey("Name should be the hostname of the system", func() {
			So(cfg.Name, ShouldEqual, getHostname())
		})
		Convey("Enable should be false", func() {
			So(cfg.Enable, ShouldEqual, false)
		})
		Convey("BindAddr should be not empty", func() {
			So(cfg.BindAddr, ShouldEqual, netutil.GetIP())
		})
		Convey("BindPort should be 6000", func() {
			So(cfg.BindPort, ShouldEqual, 6000)
		})
		Convey("Seed should be empty", func() {
			So(cfg.Seed, ShouldEqual, "")
		})
		Convey("MemberlistConfig.PushPullInterval should be 300s", func() {
			So(cfg.MemberlistConfig.PushPullInterval, ShouldEqual, 300*time.Second)
		})
		Convey("MemberlistConfig.GossipNodes should be 6", func() {
			So(cfg.MemberlistConfig.GossipNodes, ShouldEqual, 6)
		})
		Convey("RestAPIProto should be http", func() {
			So(cfg.RestAPIProto, ShouldEqual, "http")
		})
		Convey("RestAPIPassword should be empty", func() {
			So(cfg.RestAPIPassword, ShouldEqual, "")
		})
		Convey("RestAPIPort should be 8181", func() {
			So(cfg.RestAPIPort, ShouldEqual, 8181)
		})
		Convey("RestAPIInsecureSkipVerify should be true", func() {
			So(cfg.RestAPIInsecureSkipVerify, ShouldEqual, "true")
		})
	})
}
