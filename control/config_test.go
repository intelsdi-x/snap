package control

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginConfig(t *testing.T) {
	Convey("Given a plugin config", t, func() {
		cfg := NewConfig()
		So(cfg, ShouldNotBeNil)
		Convey("with an entry for ALL plugins", func() {
			cfg.Plugins.All.AddItem("gvar", ctypes.ConfigValueBool{Value: true})
			So(len(cfg.Plugins.All.Table()), ShouldEqual, 1)
			Convey("an entry for a specific plugin of any version", func() {
				cfg.Plugins.Collector["test"] = newPluginConfigItem(optAddPluginConfigItem("pvar", ctypes.ConfigValueBool{Value: true}))
				So(len(cfg.Plugins.Collector["test"].Table()), ShouldEqual, 1)
				Convey("and an entry for a specific plugin of a specific version", func() {
					cfg.Plugins.Collector["test"].Versions[1] = cdata.NewNode()
					cfg.Plugins.Collector["test"].Versions[1].AddItem("vvar", ctypes.ConfigValueBool{Value: true})
					So(len(cfg.Plugins.Collector["test"].Versions[1].Table()), ShouldEqual, 1)
					Convey("we can get the merged conf for the given plugin", func() {
						cd := cfg.Plugins.get(plugin.CollectorPluginType, "test", 1)
						So(len(cd.Table()), ShouldEqual, 3)
					})
				})
			})
		})
	})

	Convey("Provided a config in json", t, func() {
		cfg := NewConfig()
		b, err := ioutil.ReadFile("../examples/configs/pulse-config-sample.json")
		So(b, ShouldNotBeEmpty)
		So(b, ShouldNotBeNil)
		So(err, ShouldBeNil)
		Convey("We are able to unmarshal it into a valid config", func() {
			err = json.Unmarshal(b, &cfg)
			So(err, ShouldBeNil)
			So(cfg.Control, ShouldNotBeNil)
			So(cfg.Control.Table()["cache_ttl"], ShouldResemble, ctypes.ConfigValueStr{Value: "5s"})
			So(cfg.Plugins.All.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			So(cfg.Plugins.Collector["pcm"], ShouldNotBeNil)
			So(cfg.Plugins.Collector["pcm"].Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
			So(cfg.Plugins, ShouldNotBeNil)
			So(cfg.Plugins.All, ShouldNotBeNil)
			So(cfg.Plugins.Collector["pcm"].Versions[1].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
			So(cfg.Plugins.Processor["movingaverage"].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})

			Convey("Through the config object we can access the stored configurations for plugins", func() {
				c := cfg.Plugins.get(plugin.ProcessorPluginType, "movingaverage", 0)
				So(c, ShouldNotBeNil)
				So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				Convey("Overwritting the value of a variable defined for all plugins", func() {
					c := cfg.Plugins.get(plugin.ProcessorPluginType, "movingaverage", 1)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
				})
				Convey("Retrieving the value of a variable defined for all plugins", func() {
					c := cfg.Plugins.get(plugin.CollectorPluginType, "pcm", 0)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				})
				Convey("Overwritting the value of a variable defined for all versions of the plugin", func() {
					c := cfg.Plugins.get(plugin.ProcessorPluginType, "movingaverage", 1)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
				})
			})
		})
	})
}
