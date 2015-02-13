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

func TestLoadedPlugins(t *testing.T) {
	Convey("Append", t, func() {
		Convey("returns an error when loading duplicate plugins", func() {
			lp := newLoadedPlugins()
			lp.Append(new(loadedPlugin))

			p, _ := lp.Get(0)
			err := lp.Append(p)
			So(err, ShouldResemble, errors.New("plugin already loaded at index 0"))

		})
	})
	Convey("Get", t, func() {
		Convey("returns an error when index is out of range", func() {
			lp := newLoadedPlugins()
			lp.Append(new(loadedPlugin))

			_, err := lp.Get(1)
			So(err, ShouldResemble, errors.New("index out of range"))

		})
	})
	Convey("Splice", t, func() {
		Convey("splices an item out of the table", func() {
			lp := newLoadedPlugins()
			lp.Append(new(loadedPlugin))
			lp.Append(new(loadedPlugin))
			lp.Append(new(loadedPlugin))
			lp.Splice(1)
			So(len(lp.Table()), ShouldResemble, 2)

		})
	})
}

// Uses the dummy collector plugin to simulate loading
func TestLoadPlugin(t *testing.T) {
	// These tests only work if PULSE_PATH is known
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir

	if PulsePath != "" {
		Convey("PluginManager.LoadPlugin", t, func() {

			Convey("loads plugin successfully", func() {
				p := newPluginManager()
				lp, err := p.LoadPlugin(PluginPath)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.LoadedPlugins(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.LoadedPlugins().Table()), ShouldBeGreaterThan, 0)
			})

		})

	}
}

func TestUnloadPlugin(t *testing.T) {
	if PulsePath != "" {
		Convey("pluginManager.UnloadPlugin", t, func() {

			Convey("when a loaded plugin is unloaded", func() {
				Convey("then it is removed from the loadedPlugins", func() {
					p := newPluginManager()
					_, err := p.LoadPlugin(PluginPath)

					num_plugins_loaded := len(p.LoadedPlugins().Table())
					lp, _ := p.LoadedPlugins().Get(0)
					err = p.UnloadPlugin(lp)

					So(err, ShouldBeNil)
					So(len(p.LoadedPlugins().Table()), ShouldEqual, num_plugins_loaded-1)
				})
			})

			Convey("when a loaded plugin is not in a PluginLoaded state", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					_, err := p.LoadPlugin(PluginPath)
					lp, _ := p.LoadedPlugins().Get(0)
					lp.State = DetectedState
					err = p.UnloadPlugin(lp)
					So(err, ShouldResemble, errors.New("Plugin must be in a LoadedState"))
				})
			})

			Convey("when a plugin is already unloaded", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					_, err := p.LoadPlugin(PluginPath)

					plugin, _ := p.LoadedPlugins().Get(0)
					err = p.UnloadPlugin(plugin)

					err = p.UnloadPlugin(plugin)
					So(err, ShouldResemble, errors.New("plugin [/dummy/dumb] -- [1] not found (has it already been unloaded?)"))

				})
			})
		})
	}
}

func TestLoadedPlugin(t *testing.T) {
	lp := new(loadedPlugin)
	lp.Meta = plugin.PluginMeta{Name: "test", Version: 1}
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
