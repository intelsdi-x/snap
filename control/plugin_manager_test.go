package control

import (
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
				p := PluginManager()
				p.Start()
				err := p.LoadPlugin(PluginPath)

				So(p.LoadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)
			})

			Convey("returns error if PluginManager is not started", func() {
				p := PluginManager()
				err := p.LoadPlugin(PluginPath)

				So(p.LoadedPlugins, ShouldBeEmpty)
				So(err, ShouldNotBeNil)
			})
		})

	}
}

func TestPluginManagerStop(t *testing.T) {
	Convey("PluginManager.Stop", t, func() {
		p := PluginManager()
		p.Start()
		Convey("stops successfully", func() {
			p.Stop()
			So(p.Started, ShouldBeFalse)
		})
	})
}
