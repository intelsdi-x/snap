package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "pulse-publisher-riemann"
	PluginType = "publisher"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", PluginType, PluginName)
)

func TestRiemannPublisherLoad(t *testing.T) {
	if PulsePath != "" {
		helper.BuildPlugin(PluginType, PluginName)
		SkipConvey("ensure plugin loads and responds", t, func() {
			c := control.New()
			c.Start()
			err := c.Load(PluginPath)
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
