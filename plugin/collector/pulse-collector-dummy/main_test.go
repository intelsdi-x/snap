package main

import (
	"fmt"
	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/plugin/helper"
	"os"
	"path"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	PluginName = "pulse-collector-dummy"
	PluginType = "collector"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", PluginType, PluginName)
)

func TestDummyPluginLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		// Helper plugin trigger build if possible for this plugin
		helper.BuildPlugin(PluginType, PluginName)
		//
		Convey("ensure plugin loads and responds", t, func() {
			c := control.Control()
			c.Start()
			loadedPlugin, err := c.Load(PluginPath)

			So(err, ShouldBeNil)
			So(loadedPlugin, ShouldNotBeNil)
		})
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}
