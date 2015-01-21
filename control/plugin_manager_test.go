package control

import (
	"errors"
	"os"
	"path"
	"testing"

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

func TestUnload(t *testing.T) {
	if PulsePath != "" {
		Convey("pluginControl.Unload", t, func() {
			Convey("when plugin control is not started", func() {
				Convey("then an error is thrown", func() {
					c := Control()
					loadedPlugin, _ := c.Load(PluginPath)
					err := c.UnloadPlugin(loadedPlugin)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Must start plugin control before calling Load()")

				})
			})

			Convey("when a loaded plugin is unloaded", func() {
				Convey("then it is removed from the loadedPlugins", func() {
					c := Control()
					c.Start()
					loadedPlugin, err := c.Load(PluginPath)

					So(loadedPlugin, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(len(c.LoadedPlugins), ShouldBeGreaterThan, 0)

					num_plugins_loaded := len(c.LoadedPlugins)
					err = c.UnloadPlugin(loadedPlugin)

					So(err, ShouldBeNil)
					So(len(c.LoadedPlugins), ShouldEqual, num_plugins_loaded-1)
				})
			})

			Convey("when a loaded plugin is not in a PluginLoaded state", func() {
				Convey("then an error is thrown", func() {
					c := Control()
					c.Start()
					loadedPlugin, err := c.Load(PluginPath)
					So(err, ShouldBeNil)
					loadedPlugin.State = DetectedState
					err = c.UnloadPlugin(loadedPlugin)
					So(err, ShouldResemble, errors.New("Plugin must be in a LoadedState"))
				})
			})

			Convey("when a plugin is already unloaded", func() {
				Convey("then an error is thrown", func() {
					c := Control()
					c.Start()
					loadedPlugin, err := c.Load(PluginPath)
					So(loadedPlugin, ShouldNotBeNil)
					So(err, ShouldBeNil)

					num_plugins_loaded := len(c.LoadedPlugins)
					err = c.UnloadPlugin(loadedPlugin)

					So(err, ShouldBeNil)
					So(len(c.LoadedPlugins), ShouldEqual, num_plugins_loaded-1)

					err = c.UnloadPlugin(loadedPlugin)
					So(err, ShouldResemble, errors.New("Must load plugin before calling Unload()"))

				})
			})
		})
	}
}
