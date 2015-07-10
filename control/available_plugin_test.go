package control

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/routing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAvailablePlugin(t *testing.T) {
	Convey("newAvailablePlugin()", t, func() {
		Convey("returns an availablePlugin", func() {
			ln, _ := net.Listen("tcp", ":4000")
			defer ln.Close()
			resp := &plugin.Response{
				Meta: plugin.PluginMeta{
					Name:    "testPlugin",
					Version: 1,
				},
				Type:          plugin.CollectorPluginType,
				ListenAddress: "127.0.0.1:4000",
			}
			ap, err := newAvailablePlugin(resp, nil, nil)
			So(ap, ShouldHaveSameTypeAs, new(availablePlugin))
			So(err, ShouldBeNil)
		})
	})

	Convey("Stop()", t, func() {
		Convey("returns nil if plugin successfully stopped", func() {
			r := newRunner(&routing.RoundRobinStrategy{})
			a := plugin.Arg{
				PluginLogPath: "/tmp/pulse-test-plugin-stop.log",
			}

			exPlugin, _ := plugin.NewExecutablePlugin(a, PluginPath)
			ap, err := r.startPlugin(exPlugin)
			So(err, ShouldBeNil)

			err = ap.Stop("testing")
			So(err, ShouldBeNil)
		})
	})
}

func TestAvailablePlugins(t *testing.T) {
	Convey("newAvailablePlugins()", t, func() {
		Convey("returns a pointer to an availablePlugins struct", func() {
			aps := newAvailablePlugins(&routing.RoundRobinStrategy{})
			So(aps, ShouldHaveSameTypeAs, new(availablePlugins))
		})
	})
	Convey("insert()", t, func() {
		Convey("adds a collector into the collectors collection", func() {
			aps := newAvailablePlugins(&routing.RoundRobinStrategy{})
			ap := &availablePlugin{
				pluginType: plugin.CollectorPluginType,
				name:       "test",
				version:    1,
			}
			err := aps.insert(ap)
			fmt.Println("AP ID:", ap.id)
			So(err, ShouldBeNil)

			pool, err := aps.getPool("collector:test:1")
			fmt.Println("poool:", pool.plugins)
			So(err, ShouldBeNil)
			nap, ok := pool.plugins[ap.id]
			So(ok, ShouldBeTrue)
			So(nap, ShouldEqual, ap)
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins(&routing.RoundRobinStrategy{})
			ap := &availablePlugin{
				pluginType: 99,
				name:       "test",
				version:    1,
			}
			err := aps.insert(ap)

			So(err, ShouldResemble, errors.New("bad plugin type"))
		})
	})
	Convey("it returns an error if client cannot be created", t, func() {
		resp := &plugin.Response{
			Meta: plugin.PluginMeta{
				Name:    "test",
				Version: 1,
			},
			Type:          plugin.CollectorPluginType,
			ListenAddress: "localhost:",
		}
		ap, err := newAvailablePlugin(resp, nil, nil)
		So(ap, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}
