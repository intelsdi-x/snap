package control

import (
	"errors"
	"net"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
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
			ap, err := newAvailablePlugin(resp, -1, nil)
			So(ap, ShouldHaveSameTypeAs, new(availablePlugin))
			So(err, ShouldBeNil)
		})
	})

	Convey("Stop()", t, func() {
		Convey("returns nil if plugin successfully stopped", func() {
			r := newRunner()
			a := plugin.Arg{
				PluginLogPath: "/tmp/pulse-test-plugin-stop.log",
			}

			exPlugin, _ := plugin.NewExecutablePlugin(a, PluginPath)
			ap, _ := r.startPlugin(exPlugin)

			err := ap.Stop("testing")
			So(err, ShouldBeNil)
		})
	})

	Convey("makeKey()", t, func() {
		Convey("creates ap.Key from ap.Name and ap.Version", func() {
			ap := &availablePlugin{
				name:    "testPlugin",
				version: 1,
			}
			ap.makeKey()
			So(ap.Key, ShouldEqual, "testPlugin:1")
		})
	})
}

func TestAPCollection(t *testing.T) {
	Convey("newAPCollection()", t, func() {
		apc := newAPCollection()
		So(apc, ShouldHaveSameTypeAs, new(apCollection))
	})
	Convey("Add()", t, func() {
		Convey("it returns an error if an availablePlugin already exists in the table", func() {
			apc := newAPCollection()
			ap := &availablePlugin{}
			apc.Add(ap)
			err := apc.Add(ap)
			So(err, ShouldResemble, errors.New("plugin instance already available at index 0"))
		})
	})
}

func TestAvailablePlugins(t *testing.T) {
	Convey("newAvailablePlugins()", t, func() {
		Convey("returns a pointer to an availablePlugins struct", func() {
			aps := newAvailablePlugins()
			So(aps, ShouldHaveSameTypeAs, new(availablePlugins))
		})
	})
	Convey("Insert()", t, func() {
		Convey("adds a collector into the collectors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.CollectorPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldBeNil)

			tabe := aps.Collectors.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("adds a collector into the publishers collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.PublisherPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldBeNil)

			tabe := aps.Publishers.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("adds a collector into the processors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldBeNil)

			tabe := aps.Processors.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    99,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)

			So(err, ShouldResemble, errors.New("cannot insert into available plugins, unknown plugin type"))
		})
	})
	Convey("Remove()", t, func() {
		Convey("removes a collector from the collectors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.CollectorPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			err := aps.Remove(ap)
			So(err, ShouldBeNil)

			tabe := aps.Publishers.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("removes a publisher from the publishers collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.PublisherPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			err := aps.Remove(ap)
			So(err, ShouldBeNil)

			tabe := aps.Publishers.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("removes a processor from the processors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			err := aps.Remove(ap)
			So(err, ShouldBeNil)

			tabe := aps.Publishers.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    99,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()

			err := aps.Remove(ap)
			So(err, ShouldResemble, errors.New("cannot remove from available plugins, unknown plugin type"))
		})
		Convey("returns an error if does not exist", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				name:    "test",
				version: 1,
				Index:   0,
			}
			ap.makeKey()
			ap1 := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				name:    "test",
				version: 1,
				Index:   1,
			}
			ap1.makeKey()
			aps.Insert(ap)

			err := aps.Remove(ap1)
			So(err, ShouldResemble, errors.New("Warning: plugin does not exist in table"))
		})
		Convey("it returns an error if client cannot be created", func() {
			resp := &plugin.Response{
				Meta: plugin.PluginMeta{
					Name:    "test",
					Version: 1,
				},
				Type:          plugin.CollectorPluginType,
				ListenAddress: "localhost:",
			}
			ap, err := newAvailablePlugin(resp, -1, nil)
			So(ap, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
