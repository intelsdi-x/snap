package control

import (
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginConfig(t *testing.T) {
	Convey("Given a plugin config", t, func() {
		cfg := newConfig()
		So(cfg, ShouldNotBeNil)
		Convey("with an entry for ALL plugins", func() {
			cfg.plugins.all.AddItem("gvar", ctypes.ConfigValueBool{Value: true})
			So(len(cfg.plugins.all.Table()), ShouldEqual, 1)
			Convey("an entry for a specific plugin of any version", func() {
				cfg.plugins.collector["test"] = newPluginConfigItem(optAddPluginConfigItem("pvar", ctypes.ConfigValueBool{Value: true}))
				So(len(cfg.plugins.collector["test"].Table()), ShouldEqual, 1)
				Convey("and an entry for a specific plugin of a specific version", func() {
					cfg.plugins.collector["test"].versions[1] = cdata.NewNode()
					cfg.plugins.collector["test"].versions[1].AddItem("vvar", ctypes.ConfigValueBool{Value: true})
					So(len(cfg.plugins.collector["test"].versions[1].Table()), ShouldEqual, 1)
					Convey("we can get the merged conf for the given plugin", func() {
						cd := cfg.plugins.get(plugin.CollectorPluginType, "test", 1)
						So(len(cd.Table()), ShouldEqual, 3)
					})
				})
			})
		})

	})
}
