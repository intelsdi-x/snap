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
	"testing"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginConfig(t *testing.T) {
	Convey("Given a plugin config", t, func() {
		cfg := NewConfig()
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
		cfg := NewConfig()
		cfg.LoadConfig("../examples/configs/snap-config-sample.json")
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
