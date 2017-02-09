// +build medium

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

package rest

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
		`"control": {}, "scheduler": {}, ` + CONFIG_CONSTRAINTS + `, "tribe":{}` +
		`}` +
		`}`
)

type mockRestAPIConfig struct {
	RestAPI *Config
}

func TestRestAPIConfigJSON(t *testing.T) {
	config := &mockRestAPIConfig{
		RestAPI: GetDefaultConfig(),
	}
	path := "../../examples/configs/snap-config-sample.json"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.RestAPI
	}
	Convey("Provided a valid config in JSON", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("Enable should be true", func() {
			So(cfg.Enable, ShouldEqual, true)
		})
		Convey("HTTPS should be true", func() {
			So(cfg.HTTPS, ShouldEqual, true)
		})
		Convey("Port should be 8282", func() {
			So(cfg.Port, ShouldEqual, 8282)
		})
		Convey("Address should equal 127.0.0.1:12345", func() {
			So(cfg.Address, ShouldEqual, "127.0.0.1:12345")
		})
		Convey("RestAuth should be true", func() {
			So(cfg.RestAuth, ShouldEqual, true)
		})
		Convey("RestAuthPassword should equal changeme", func() {
			So(cfg.RestAuthPassword, ShouldEqual, "changeme")
		})
		Convey("RestCertificate should equal /etc/snap/cert.pem", func() {
			So(cfg.RestCertificate, ShouldEqual, "/etc/snap/cert.pem")
		})
		Convey("RestKey should equal /etc/snap/cert.key", func() {
			So(cfg.RestKey, ShouldEqual, "/etc/snap/cert.key")
		})
	})

}

func TestRestAPIConfigYaml(t *testing.T) {
	config := &mockRestAPIConfig{
		RestAPI: GetDefaultConfig(),
	}
	path := "../../examples/configs/snap-config-sample.yaml"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.RestAPI
	}
	Convey("Provided a valid config in YAML", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("Enable should be true", func() {
			So(cfg.Enable, ShouldEqual, true)
		})
		Convey("HTTPS should be true", func() {
			So(cfg.HTTPS, ShouldEqual, true)
		})
		Convey("Port should be 8282", func() {
			So(cfg.Port, ShouldEqual, 8282)
		})
		Convey("Address should equal 127.0.0.1:12345", func() {
			So(cfg.Address, ShouldEqual, "127.0.0.1:12345")
		})
		Convey("RestAuth should be true", func() {
			So(cfg.RestAuth, ShouldEqual, true)
		})
		Convey("RestAuthPassword should equal changeme", func() {
			So(cfg.RestAuthPassword, ShouldEqual, "changeme")
		})
		Convey("RestCertificate should equal /etc/snap/cert.pem", func() {
			So(cfg.RestCertificate, ShouldEqual, "/etc/snap/cert.pem")
		})
		Convey("RestKey should equal /etc/snap/cert.key", func() {
			So(cfg.RestKey, ShouldEqual, "/etc/snap/cert.key")
		})
	})

}

func TestRestAPIDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	Convey("Provided a default RestAPI config", t, func() {
		Convey("Enable should be true", func() {
			So(cfg.Enable, ShouldEqual, true)
		})
		Convey("HTTPS should be false", func() {
			So(cfg.HTTPS, ShouldEqual, false)
		})
		Convey("Port should be 8181", func() {
			So(cfg.Port, ShouldEqual, 8181)
		})
		Convey("Address should equal empty string", func() {
			So(cfg.Address, ShouldEqual, "")
		})
		Convey("RestAuth should be false", func() {
			So(cfg.RestAuth, ShouldEqual, false)
		})
		Convey("RestAuthPassword should be empty", func() {
			So(cfg.RestAuthPassword, ShouldEqual, "")
		})
		Convey("RestCertificate should be empty", func() {
			So(cfg.RestCertificate, ShouldEqual, "")
		})
		Convey("RestKey should be empty", func() {
			So(cfg.RestKey, ShouldEqual, "")
		})
	})
}
