package control

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "pulse-collector-dummy"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", "collector", PluginName)
)

// Uses the dummy collector plugin to simulate loading
func TestLoadPlugin(t *testing.T) {
	// These tests only work if PULSE_PATH is known
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir

	if PulsePath != "" {
		Convey("PluginManager.LoadPlugin", t, func() {

			Convey("loads plugin successfully", func() {
				p := newPluginManager()
				p.Start()
				err := p.LoadPlugin(PluginPath)

				So(p.LoadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldBeGreaterThan, 0)
			})

			Convey("returns error if PluginManager is not started", func() {
				p := newPluginManager()
				err := p.LoadPlugin(PluginPath)

				So(p.LoadedPlugins, ShouldBeEmpty)
				So(err, ShouldNotBeNil)
			})
		})

	}
}

func TestPluginManagerStop(t *testing.T) {
	Convey("PluginManager.Stop", t, func() {
		p := newPluginManager()
		p.Start()
		Convey("stops successfully", func() {
			p.Stop()
			So(p.Started, ShouldBeFalse)
		})
	})
}

func TestUnloadPlugin(t *testing.T) {
	if PulsePath != "" {
		Convey("pluginManager.UnloadPlugin", t, func() {
			Convey("when pluginManager is not started", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.Start()
					p.LoadPlugin(PluginPath)
					p.Stop()
					err := p.UnloadPlugin(p.LoadedPlugins[0])
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Must start pluginManager before calling UnloadPlugin()")

				})
			})

			Convey("when a loaded plugin is unloaded", func() {
				Convey("then it is removed from the loadedPlugins", func() {
					p := newPluginManager()
					p.Start()
					err := p.LoadPlugin(PluginPath)

					num_plugins_loaded := len(p.LoadedPlugins)
					err = p.UnloadPlugin(p.LoadedPlugins[0])

					So(err, ShouldBeNil)
					So(len(p.LoadedPlugins), ShouldEqual, num_plugins_loaded-1)
				})
			})

			Convey("when a loaded plugin is not in a PluginLoaded state", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.Start()
					err := p.LoadPlugin(PluginPath)
					p.LoadedPlugins[0].State = DetectedState
					err = p.UnloadPlugin(p.LoadedPlugins[0])
					So(err, ShouldResemble, errors.New("Plugin must be in a LoadedState"))
				})
			})

			Convey("when a plugin is already unloaded", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.Start()
					err := p.LoadPlugin(PluginPath)

					plugin := p.LoadedPlugins[0]
					err = p.UnloadPlugin(plugin)

					err = p.UnloadPlugin(plugin)
					So(err, ShouldResemble, errors.New("Must load plugin before calling UnloadPlugin()"))

				})
			})
		})
	}
}

func TestLoadedPlugin(t *testing.T) {
	lp := new(loadedPlugin)
	lp.Meta = plugin.PluginMeta{"test", 1}
	Convey(".Name()", t, func() {
		Convey("it returns the name from the plugin metadata", func() {
			So(lp.Name(), ShouldEqual, "test")
		})
	})
	Convey(".Version()", t, func() {
		Convey("it returns the version from the plugin metadata", func() {
			So(lp.Version(), ShouldEqual, 1)
		})
	})
	Convey(".TypeName()", t, func() {
		lp.Type = 0
		Convey("it returns the string representation of the plugin type", func() {
			So(lp.TypeName(), ShouldEqual, "collector")
		})
	})
	Convey(".Status()", t, func() {
		lp.State = LoadedState
		Convey("it returns a string of the current plugin state", func() {
			So(lp.Status(), ShouldEqual, "loaded")
		})
	})
	Convey(".LoadedTimestamp()", t, func() {
		ts := time.Now()
		lp.LoadedTime = ts
		Convey("it returns the Unix timestamp of the LoadedTime", func() {
			So(lp.LoadedTimestamp(), ShouldEqual, ts.Unix())
		})
	})
}
