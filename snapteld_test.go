// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package main

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vrischmann/jsonutil"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/mgmt/rest"
	"github.com/intelsdi-x/snap/mgmt/tribe"
	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/scheduler"
)

var validCmdlineFlags_input = mockFlags{
	"max-procs":               "11",
	"log-level":               "1",
	"log-path":                "/no/logs/allowed",
	"log-truncate":            "true",
	"log-colors":              "true",
	"max-running-plugins":     "12",
	"plugin-load-timeout":     "20",
	"plugin-trust":            "1",
	"auto-discover":           "/no/plugins/here",
	"keyring-paths":           "/no/keyrings/here",
	"cache-expiration":        "30ms",
	"control-listen-addr":     "100.101.102.103",
	"control-listen-port":     "10400",
	"pprof":                   "true",
	"temp_dir_path":           "/no/temp/files",
	"tls-cert":                "/no/cert/here",
	"tls-key":                 "/no/key/here",
	"ca-cert-paths":           "/no/root/certs",
	"disable-api":             "false",
	"api-port":                "12400",
	"api-addr":                "120.121.122.123",
	"rest-https":              "true",
	"rest-cert":               "/no/rest/cert",
	"rest-key":                "/no/rest/key",
	"rest-auth":               "true",
	"rest-auth-pwd":           "noway",
	"allowed_origins":         "140.141.142.143",
	"work-manager-queue-size": "70",
	"work-manager-pool-size":  "71",
	"tribe-node-name":         "bonk",
	"tribe":                   "true",
	"tribe-addr":              "160.161.162.163",
	"tribe-port":              "16400",
	"tribe-seed":              "180.181.182.183",
}

var validCmdlineFlags_expected = &Config{
	Control: &control.Config{
		MaxRunningPlugins: 12,
		PluginLoadTimeout: 20,
		PluginTrust:       1,
		AutoDiscoverPath:  "/no/plugins/here",
		KeyringPaths:      "/no/keyrings/here",
		CacheExpiration:   jsonutil.Duration{30 * time.Millisecond},
		ListenAddr:        "100.101.102.103",
		ListenPort:        10400,
		Pprof:             true,
		TempDirPath:       "/no/temp/files",
		TLSCertPath:       "/no/cert/here",
		TLSKeyPath:        "/no/key/here",
		CACertPaths:       "/no/root/certs",
	},
	RestAPI: &rest.Config{
		Enable:           true,
		Port:             12400,
		Address:          "120.121.122.123:12400",
		HTTPS:            true,
		RestCertificate:  "/no/rest/cert",
		RestKey:          "/no/rest/key",
		RestAuth:         true,
		RestAuthPassword: "noway",
		Pprof:            true,
		Corsd:            "140.141.142.143",
	},
	Tribe: &tribe.Config{
		Name:     "bonk",
		Enable:   true,
		BindAddr: "160.161.162.163",
		BindPort: 16400,
		Seed:     "180.181.182.183",
	},
	Scheduler: &scheduler.Config{
		WorkManagerQueueSize: 70,
		WorkManagerPoolSize:  71,
	},
	GoMaxProcs:  11,
	LogLevel:    1,
	LogPath:     "/no/logs/allowed",
	LogTruncate: true,
	LogColors:   true,
}

func TestSnapConfig(t *testing.T) {
	Convey("Test Config", t, func() {
		Convey("with defaults", func() {
			cfg := getDefaultConfig()
			jb, _ := json.Marshal(cfg)
			serrs := cfgfile.ValidateSchema(CONFIG_CONSTRAINTS, string(jb))
			So(len(serrs), ShouldEqual, 0)
		})
	})
}

func TestSnapConfigEmpty(t *testing.T) {
	Convey("Test Config", t, func() {
		Convey("with empty plugin config sections", func() {
			cfg := getDefaultConfig()
			readConfig(cfg, "./examples/configs/snap-config-empty.yaml")
			jb, _ := json.Marshal(cfg)
			serrs := cfgfile.ValidateSchema(CONFIG_CONSTRAINTS, string(jb))
			So(len(serrs), ShouldEqual, 0)
		})
	})
}

type mockFlags map[string]string

func (m mockFlags) String(key string) string {
	return m[key]
}

func (m mockFlags) Int(key string) int {
	if v, err := strconv.Atoi(m[key]); err == nil {
		return v
	}
	return 0
}

func (m mockFlags) Bool(key string) bool {
	if v, err := strconv.ParseBool(m[key]); err == nil {
		return v
	}
	return false
}

func (m mockFlags) IsSet(key string) bool {
	_, gotIt := m[key]
	return gotIt
}

func (m mockFlags) getCopy() mockFlags {
	r := mockFlags{}
	for k, v := range m {
		r[k] = v
	}
	return r
}

func (m mockFlags) copyWithout(keys ...string) mockFlags {
	r := m.getCopy()
	for _, k := range keys {
		delete(r, k)
	}
	return r
}

func (m mockFlags) update(key, value string) mockFlags {
	m[key] = value
	return m
}

type mockCfg Config

func (c *mockCfg) setTLSCert(tlsCertPath string) *mockCfg {
	c.Control.TLSCertPath = tlsCertPath
	return c
}

func (c *mockCfg) setTLSKey(tlsKeyPath string) *mockCfg {
	c.Control.TLSKeyPath = tlsKeyPath
	return c
}

func (c *mockCfg) setCACertPaths(caCertPaths string) *mockCfg {
	c.Control.CACertPaths = caCertPaths
	return c
}

func (c *mockCfg) setApiAddr(apiAddr string) *mockCfg {
	c.RestAPI.Address = apiAddr
	return c
}

func (c *mockCfg) getCopy() (r *mockCfg) {
	r = &mockCfg{}
	b, err := json.Marshal(*c)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, r)
	if err != nil {
		panic(err)
	}
	return r
}

func (c *mockCfg) export() *Config {
	return (*Config)(c)
}

func Test_checkCmdLineFlags(t *testing.T) {
	testCtx := mockFlags{
		"tls-cert":      "mock-cli.crt",
		"tls-key":       "mock-cli.key",
		"ca-cert-paths": "mock-ca.crt",
		"api-addr":      "localhost",
		"api-port":      "9000"}
	tests := []struct {
		name           string
		msg            func(func(string))
		ctx            runtimeFlagsContext
		wantErr        bool
		wantPort       int
		wantPortInAddr bool
	}{
		{name: "CmdlineArgsParseWell",
			msg: func(f func(string)) {
				f("Having valid command line flags, parsing succeeds")
			},
			ctx:            testCtx.getCopy(),
			wantErr:        false,
			wantPort:       9000,
			wantPortInAddr: false},
		{name: "CmdlineArgsWithoutTLSConfigParseWell",
			msg: func(f func(string)) {
				f("Having valid command line flags without any TLS parameters, parsing succeeds")
			},
			ctx: testCtx.
				copyWithout("tls-cert", "tls-key", "ca-cert-paths", "api-port").
				update("api-addr", "127.0.0.1:9002"),
			wantErr:        false,
			wantPort:       9002,
			wantPortInAddr: true},
		{name: "ArgsWithTLSCertWithoutKey_Fail",
			msg: func(f func(string)) {
				f("Having command line flags with TLS cert without key, parsing fails")
			},
			ctx:     testCtx.copyWithout("tls-key"),
			wantErr: true,
		},
		{name: "ArgsWithTLSKeyWithoutCert_Fail",
			msg: func(f func(string)) {
				f("Having command line flags with TLS key without cert, parsing fails")
			},
			ctx:     testCtx.copyWithout("tls-cert"),
			wantErr: true,
		},
	}
	for _, tc := range tests {
		runThisCase := func(f func(msg string)) {
			t.Run(tc.name, func(_ *testing.T) {
				tc.msg(f)
			})
		}
		runThisCase(func(msg string) {
			Convey(msg, t, func() {
				gotPort, gotPortInAddr, err := checkCmdLineFlags(tc.ctx)
				if tc.wantErr {
					So(err, ShouldNotBeNil)
					return
				}
				So(err, ShouldBeNil)
				So(gotPort, ShouldEqual, tc.wantPort)
				So(gotPortInAddr, ShouldEqual, tc.wantPortInAddr)
			})
		})
	}
}

func Test_checkCfgSettings(t *testing.T) {
	const DontCheckInt = -99
	testCfg := &mockCfg{
		Control: &control.Config{},
		RestAPI: &rest.Config{},
	}
	tests := []struct {
		name           string
		msg            func(func(string))
		cfg            *Config
		wantErr        bool
		wantPort       int
		wantPortInAddr bool
	}{
		{name: "DefaultConfigSettingsValidateWell",
			msg: func(f func(string)) {
				f("Having all default (empty) values for config, validation succeeds")
			},
			cfg:            (&mockCfg{Control: control.GetDefaultConfig(), RestAPI: rest.GetDefaultConfig()}).export(),
			wantErr:        false,
			wantPort:       DontCheckInt,
			wantPortInAddr: false},
		{name: "ConfigSettingsValidateWell",
			msg: func(f func(string)) {
				f("Having correct values, config validation succeeds")
			},
			cfg: testCfg.getCopy().
				setApiAddr("localhost:9000").
				setTLSCert("mock-cli.crt").
				setTLSKey("mock-cli.key").
				setCACertPaths("mock-ca.crt").
				export(),
			wantErr:        false,
			wantPort:       9000,
			wantPortInAddr: true},
		{name: "ConfigSettingsWithoutTLSConfigValidateWell",
			msg: func(f func(string)) {
				f("Having correct values without any TLS parameters, config validation succeeds")
			},
			cfg: testCfg.getCopy().
				setApiAddr("localhost:9000").
				export(),
			wantErr:        false,
			wantPort:       9000,
			wantPortInAddr: true},
		{name: "ConfigSettingsWithTLSCertWithoutKey_Fail",
			msg: func(f func(string)) {
				f("Having config with TLS cert without key, config fails to validate")
			},
			cfg: testCfg.getCopy().
				setApiAddr("localhost:9000").
				setTLSCert("mock-cli.crt").
				export(),
			wantErr:        true,
			wantPort:       9000,
			wantPortInAddr: true},
		{name: "ConfigSettingsWithTLSKeyWithoutCert_Fail",
			msg: func(f func(string)) {
				f("Having config with TLS key without cert, config fails to validate")
			},
			cfg: testCfg.getCopy().
				setApiAddr("localhost:9000").
				setTLSKey("mock-cli.crt").
				export(),
			wantErr:        true,
			wantPort:       9000,
			wantPortInAddr: true},
	}

	for _, tc := range tests {
		runThisCase := func(f func(msg string)) {
			t.Run(tc.name, func(_ *testing.T) {
				tc.msg(f)
			})
		}
		runThisCase(func(msg string) {
			Convey(msg, t, func() {
				gotPort, gotPortInAddr, err := checkCfgSettings(tc.cfg)
				if tc.wantErr {
					So(err, ShouldNotBeNil)
					return
				}
				So(err, ShouldBeNil)
				if tc.wantPort != DontCheckInt {
					So(gotPort, ShouldEqual, tc.wantPort)
				}
				So(gotPortInAddr, ShouldEqual, tc.wantPortInAddr)
			})
		})
	}
}

func Test_applyCmdLineFlags(t *testing.T) {
	Convey("Having arguments given on command line", t, func() {
		gotConfig := Config{
			Control:   &control.Config{},
			RestAPI:   &rest.Config{},
			Tribe:     &tribe.Config{},
			Scheduler: &scheduler.Config{},
		}
		applyCmdLineFlags(&gotConfig, validCmdlineFlags_input)
		Convey("config should be filled with correct values", func() {
			So(*gotConfig.Control, ShouldResemble, *validCmdlineFlags_expected.Control)
			So(*gotConfig.RestAPI, ShouldResemble, *validCmdlineFlags_expected.RestAPI)
			So(*gotConfig.Tribe, ShouldResemble, *validCmdlineFlags_expected.Tribe)
			So(*gotConfig.Scheduler, ShouldResemble, *validCmdlineFlags_expected.Scheduler)
			So(gotConfig, ShouldResemble, *validCmdlineFlags_expected)
		})
	})
}
