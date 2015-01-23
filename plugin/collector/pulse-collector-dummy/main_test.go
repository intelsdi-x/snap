package main

import (
	"fmt"
	"github.com/intelsdilabs/pulse/control"
	"os"
	"path"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	PluginName = "pulse-collector-dummy"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", "collector", "pulse-collector-dummy")
)

func TestDummyPluginLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		Convey("ensure plugin loads and responds", t, func() {
			c := control.Control()
			c.Start()
			err := c.Load(PluginPath)

			So(err, ShouldBeNil)
		})
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}
