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

package control

import (
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
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
		CONFIG_CONSTRAINTS + `,` + `"scheduler": {}, "restapi": {}, "tribe":{}` +
		`}` +
		`}`
)

type mockConfig struct {
	Control *Config
}

func TestPluginConfig(t *testing.T) {
	Convey("Given a plugin config", t, func() {
		cfg := GetDefaultConfig()
		So(cfg, ShouldNotBeNil)
		Convey("with an entry for ALL plugins", func() {
			cfg.Plugins.All.AddItem("gvar", ctypes.ConfigValueBool{Value: true})
			So(len(cfg.Plugins.All.Table()), ShouldEqual, 1)
			Convey("with an entry for ALL collector plugins", func() {
				cfg.Plugins.Collector.All.AddItem("user", ctypes.ConfigValueStr{Value: "jane"})
				cfg.Plugins.Collector.All.AddItem("password", ctypes.ConfigValueStr{Value: "P@ssw0rd"})
				So(len(cfg.Plugins.Collector.All.Table()), ShouldEqual, 2)
				Convey("an entry for a specific plugin of any version", func() {
					cfg.Plugins.Collector.Plugins["test"] = newPluginConfigItem(optAddPluginConfigItem("user", ctypes.ConfigValueStr{Value: "john"}))
					So(len(cfg.Plugins.Collector.Plugins["test"].Table()), ShouldEqual, 1)
					Convey("and an entry for a specific plugin of a specific version", func() {
						cfg.Plugins.Collector.Plugins["test"].Versions[1] = cdata.NewNode()
						cfg.Plugins.Collector.Plugins["test"].Versions[1].AddItem("vvar", ctypes.ConfigValueBool{Value: true})
						So(len(cfg.Plugins.Collector.Plugins["test"].Versions[1].Table()), ShouldEqual, 1)
						Convey("we can get the merged conf for the given plugin", func() {
							cd := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "test", 1)
							So(len(cd.Table()), ShouldEqual, 4)
							So(cd.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
						})
					})
				})
			})
		})
	})

	Convey("Provided a config in JSON we are able to unmarshal it into a valid config", t, func() {
		config := &mockConfig{
			Control: GetDefaultConfig(),
		}
		path := "../examples/configs/snap-config-sample.json"
		err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
		So(err, ShouldBeNil)
		cfg := config.Control
		So(cfg.Plugins, ShouldNotBeNil)
		So(cfg.Plugins.All, ShouldNotBeNil)
		So(cfg.Plugins.All.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
		So(cfg.Plugins.Collector, ShouldNotBeNil)
		So(cfg.Plugins.All, ShouldNotBeNil)
		So(cfg.Plugins.Collector.Plugins["pcm"], ShouldNotBeNil)
		So(cfg.Plugins.Collector.Plugins["pcm"].Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
		So(cfg.Plugins.Collector.Plugins["pcm"].Versions[1].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
		So(cfg.Plugins.Processor, ShouldNotBeNil)
		So(cfg.Plugins.Processor.Plugins["movingaverage"].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
		So(cfg.Tags["/intel/psutil"], ShouldNotBeNil)
		So(cfg.Tags["/intel/psutil"]["context"], ShouldNotBeNil)
		So(cfg.Tags["/intel/psutil"]["context"], ShouldEqual, "config_example")
		So(cfg.Tags["/"], ShouldNotBeNil)
		So(cfg.Tags["/"]["color"], ShouldNotBeNil)
		So(cfg.Tags["/"]["color"], ShouldEqual, "green")

		Convey("We can access the config for plugins", func() {
			Convey("Getting the values of a specific version of a plugin", func() {
				c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "pcm", 1)
				So(c, ShouldNotBeNil)
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				So(c.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(c.Table()["somefloat"], ShouldResemble, ctypes.ConfigValueFloat{Value: 3.14})
				So(c.Table()["someint"], ShouldResemble, ctypes.ConfigValueInt{Value: 1234})
				So(c.Table()["somebool"], ShouldResemble, ctypes.ConfigValueBool{Value: true})
			})
			Convey("Getting the common config for collectors", func() {
				c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "", -2)
				So(c, ShouldNotBeNil)
				So(len(c.Table()), ShouldEqual, 2)
				So(c.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Overwriting the value of a variable defined for all plugins", func() {
				c := cfg.Plugins.getPluginConfigDataNode(core.ProcessorPluginType, "movingaverage", 1)
				So(c, ShouldNotBeNil)
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
			})
			Convey("Retrieving the value of a variable defined for all versions of a plugin", func() {
				c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "pcm", 0)
				So(c, ShouldNotBeNil)
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Overwriting the value of a variable defined for all versions of the plugin", func() {
				c := cfg.Plugins.getPluginConfigDataNode(core.ProcessorPluginType, "movingaverage", 1)
				So(c, ShouldNotBeNil)
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
			})

		})
	})
}

func TestControlConfigJSON(t *testing.T) {
	config := &mockConfig{
		Control: GetDefaultConfig(),
	}
	path := "../examples/configs/snap-config-sample.json"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Control
	}
	Convey("Provided a valid config in JSON", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("AutoDiscoverPath should be set to /opt/snap/plugins:/opt/snap/tasks", func() {
			So(cfg.AutoDiscoverPath, ShouldEqual, "/opt/snap/plugins:/opt/snap/tasks")
		})
		Convey("CacheExpiration should be set to 750ms", func() {
			So(cfg.CacheExpiration.Duration, ShouldResemble, 750*time.Millisecond)
		})
		Convey("MaxRunningPlugins should be set to 1", func() {
			So(cfg.MaxRunningPlugins, ShouldEqual, 1)
		})
		Convey("max_plugin_restarts should be set to 10", func() {
			So(cfg.MaxPluginRestarts, ShouldEqual, 10)
		})
		Convey("ListenAddr should be set to 0.0.0.0", func() {
			So(cfg.ListenAddr, ShouldEqual, "0.0.0.0")
		})
		Convey("ListenPort should be set to 10082", func() {
			So(cfg.ListenPort, ShouldEqual, 10082)
		})
		Convey("KeyringPaths should be set to /etc/snap/keyrings", func() {
			So(cfg.KeyringPaths, ShouldEqual, "/etc/snap/keyrings")
		})
		Convey("PluginTrust should be set to 0", func() {
			So(cfg.PluginTrust, ShouldEqual, 0)
		})
		Convey("Plugins section of control configuration should not be nil", func() {
			So(cfg.Plugins, ShouldNotBeNil)
		})
		Convey("Plugins.All section should not be nil", func() {
			So(cfg.Plugins.All, ShouldNotBeNil)
		})
		Convey("A password should be configured for all plugins", func() {
			So(cfg.Plugins.All.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
		})
		Convey("Plugins.Collector section should not be nil", func() {
			So(cfg.Plugins.Collector, ShouldNotBeNil)
		})
		Convey("Plugins.Collector should have config for pcm collector plugin", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"], ShouldNotBeNil)
		})
		Convey("Config for pcm should set path to pcm binary to /usr/local/pcm/bin", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"].Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
		})
		Convey("Config for pcm plugin at version 1 should set user to john", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"].Versions[1].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
		})
		Convey("Plugins.Processor section should not be nil", func() {
			So(cfg.Plugins.Processor, ShouldNotBeNil)
		})
		Convey("Movingaverage processor plugin should have user set to jane", func() {
			So(cfg.Plugins.Processor.Plugins["movingaverage"].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
		})
		Convey("Plugins.Publisher should not be nil", func() {
			So(cfg.Plugins.Publisher, ShouldNotBeNil)
		})
	})

}

func TestControlConfigYaml(t *testing.T) {
	config := &mockConfig{
		Control: GetDefaultConfig(),
	}
	os.Setenv("password", "$password")
	path := "../examples/configs/snap-config-sample.yaml"
	err := cfgfile.Read(path, &config, MOCK_CONSTRAINTS)
	var cfg *Config
	if err == nil {
		cfg = config.Control
	}
	Convey("Provided a valid config in YAML", t, func() {
		Convey("An error should not be returned when unmarshalling the config", func() {
			So(err, ShouldBeNil)
		})
		Convey("AutoDiscoverPath should be set to /opt/snap/plugins:/opt/snap/tasks", func() {
			So(cfg.AutoDiscoverPath, ShouldEqual, "/opt/snap/plugins:/opt/snap/tasks")
		})
		Convey("CacheExpiration should be set to 750ms", func() {
			So(cfg.CacheExpiration.Duration, ShouldResemble, 750*time.Millisecond)
		})
		Convey("MaxRunningPlugins should be set to 1", func() {
			So(cfg.MaxRunningPlugins, ShouldEqual, 1)
		})
		Convey("max_plugin_restarts should be set to 10", func() {
			So(cfg.MaxPluginRestarts, ShouldEqual, 10)
		})
		Convey("ListenAddr should be set to 0.0.0.0", func() {
			So(cfg.ListenAddr, ShouldEqual, "0.0.0.0")
		})
		Convey("ListenPort should be set to 10082", func() {
			So(cfg.ListenPort, ShouldEqual, 10082)
		})
		Convey("KeyringPaths should be set to /etc/snap/keyrings", func() {
			So(cfg.KeyringPaths, ShouldEqual, "/etc/snap/keyrings")
		})
		Convey("PluginTrust should be set to 0", func() {
			So(cfg.PluginTrust, ShouldEqual, 0)
		})
		Convey("Plugins section of control configuration should not be nil", func() {
			So(cfg.Plugins, ShouldNotBeNil)
		})
		Convey("Plugins.All section should not be nil", func() {
			So(cfg.Plugins.All, ShouldNotBeNil)
		})
		Convey("A password should be configured for all plugins", func() {
			So(cfg.Plugins.All.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
		})
		Convey("Plugins.Collector section should not be nil", func() {
			So(cfg.Plugins.Collector, ShouldNotBeNil)
		})
		Convey("Plugins.Collector should have config for pcm collector plugin", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"], ShouldNotBeNil)
		})
		Convey("Config for pcm should set path to pcm binary to /usr/local/pcm/bin", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"].Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
		})
		Convey("Config for pcm plugin at version 1 should set user to john", func() {
			So(cfg.Plugins.Collector.Plugins["pcm"].Versions[1].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
		})
		Convey("Plugins.Processor section should not be nil", func() {
			So(cfg.Plugins.Processor, ShouldNotBeNil)
		})
		Convey("Movingaverage processor plugin should have user set to jane", func() {
			So(cfg.Plugins.Processor.Plugins["movingaverage"].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
		})
		Convey("Plugins.Publisher should not be nil", func() {
			So(cfg.Plugins.Publisher, ShouldNotBeNil)
		})
	})

}

func TestControlDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	Convey("Provided a default config", t, func() {
		Convey("AutoDiscoverPath should be empty", func() {
			So(cfg.AutoDiscoverPath, ShouldEqual, "")
		})
		Convey("CacheExpiration should equal 500ms", func() {
			So(cfg.CacheExpiration.Duration, ShouldEqual, 500*time.Millisecond)
		})
		Convey("MaxRunningPlugins should equal 3", func() {
			So(cfg.MaxRunningPlugins, ShouldEqual, 3)
		})
		Convey("ListenAddr should be set to 127.0.0.1", func() {
			So(cfg.ListenAddr, ShouldEqual, "127.0.0.1")
		})
		Convey("ListenPort should be set to 8082", func() {
			So(cfg.ListenPort, ShouldEqual, 8082)
		})
		Convey("PluginTrust should equal 1", func() {
			So(cfg.PluginTrust, ShouldEqual, 1)
		})
		Convey("max_plugin_restarts should be set to 3", func() {
			So(cfg.MaxPluginRestarts, ShouldEqual, 3)
		})
	})
}
