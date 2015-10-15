package control

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/intelsdi-x/pulse/core"
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
			Convey("with an entry for ALL collector plugins", func() {
				cfg.Plugins.Collector.All.AddItem("user", ctypes.ConfigValueStr{"jane"})
				cfg.Plugins.Collector.All.AddItem("password", ctypes.ConfigValueStr{"P@ssw0rd"})
				So(len(cfg.Plugins.Collector.All.Table()), ShouldEqual, 2)
				Convey("an entry for a specific plugin of any version", func() {
					cfg.Plugins.Collector.Plugins["test"] = newPluginConfigItem(optAddPluginConfigItem("user", ctypes.ConfigValueStr{"john"}))
					So(len(cfg.Plugins.Collector.Plugins["test"].Table()), ShouldEqual, 1)
					Convey("and an entry for a specific plugin of a specific version", func() {
						cfg.Plugins.Collector.Plugins["test"].Versions[1] = cdata.NewNode()
						cfg.Plugins.Collector.Plugins["test"].Versions[1].AddItem("vvar", ctypes.ConfigValueBool{Value: true})
						So(len(cfg.Plugins.Collector.Plugins["test"].Versions[1].Table()), ShouldEqual, 1)
						Convey("we can get the merged conf for the given plugin", func() {
							cd := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "test", 1)
							So(len(cd.Table()), ShouldEqual, 4)
							So(cd.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{"john"})
						})
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
			So(cfg.Plugins.All.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			So(cfg.Plugins.Collector.Plugins["pcm"], ShouldNotBeNil)
			So(cfg.Plugins.Collector.Plugins["pcm"].Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
			So(cfg.Plugins, ShouldNotBeNil)
			So(cfg.Plugins.All, ShouldNotBeNil)
			So(cfg.Plugins.Collector.Plugins["pcm"].Versions[1].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
			So(cfg.Plugins.Processor.Plugins["movingaverage"].Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})

			Convey("We can access the config for plugins", func() {
				Convey("Getting the values of a specific version of a plugin", func() {
					c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "pcm", 1)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
					So(c.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
					So(c.Table()["float"], ShouldResemble, ctypes.ConfigValueFloat{Value: 3.14})
					So(c.Table()["int"], ShouldResemble, ctypes.ConfigValueInt{Value: 1234})
					So(c.Table()["flag"], ShouldResemble, ctypes.ConfigValueBool{Value: true})
				})
				Convey("Getting the common config for collectors", func() {
					c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "", -2)
					So(c, ShouldNotBeNil)
					So(len(c.Table()), ShouldEqual, 2)
					So(c.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				})
				Convey("Overwritting the value of a variable defined for all plugins", func() {
					c := cfg.Plugins.getPluginConfigDataNode(core.ProcessorPluginType, "movingaverage", 1)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
				})
				Convey("Retrieving the value of a variable defined for all versions of a plugin", func() {
					c := cfg.Plugins.getPluginConfigDataNode(core.CollectorPluginType, "pcm", 0)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				})
				Convey("Overwritting the value of a variable defined for all versions of the plugin", func() {
					c := cfg.Plugins.getPluginConfigDataNode(core.ProcessorPluginType, "movingaverage", 1)
					So(c, ShouldNotBeNil)
					So(c.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "new password"})
				})

			})
		})
	})
}
