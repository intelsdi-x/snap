// Unit testing for load, unload, and swap of plugins
package control

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

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
