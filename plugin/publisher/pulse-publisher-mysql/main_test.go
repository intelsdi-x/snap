package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "pulse-publisher-mysql"
	PluginType = "publisher"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", PluginName)
)

func TestMySQLPublisherLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		// Helper plugin trigger build if possible for this plugin
		helper.BuildPlugin(PluginType, PluginName)
		//
		//TODO cannot test this locally. We need AMQP and integration tests.
		SkipConvey("ensure plugin loads and responds", t, func() {
			c := control.New()
			c.Start()
			_, err := c.Load(PluginPath)

			So(err, ShouldBeNil)
		})
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestMain(t *testing.T) {
	Convey("ensure plugin loads and responds", t, func() {
		os.Args = []string{"", "{\"NoDaemon\": true}"}
		So(func() { main() }, ShouldNotPanic)
	})
}
