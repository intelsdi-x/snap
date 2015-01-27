package control

import (
	"fmt"
	"strings"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Mock Executor used to test
type MockPluginExecutor struct {
	Killed          bool
	Response        string
	WaitTime        time.Duration
	WaitError       error
	WaitForResponse func(time.Duration) (*plugin.Response, error)
}

func TestPluginControlStart(t *testing.T) {
	Convey("pluginControl.Start", t, func() {
		Convey("starts successfully", func() {
			c := Control()
			c.Start()
			So(c.Started, ShouldBeTrue)
		})
	})
}

func TestSwapPlugin(t *testing.T) {
	if PulsePath != "" {
		Convey("SwapPlugin", t, func() {
			c := Control()
			c.Start()
			c.Load(PluginPath)

			facterPath := strings.Replace(PluginPath, "pulse-collector-dummy", "pulse-collector-facter", 1)
			pc := c.PluginCatalog()
			dummy := pc[0]

			Convey("successfully swaps plugins", func() {
				err := c.SwapPlugins(facterPath, dummy)
				pc = c.PluginCatalog()
				So(err, ShouldBeNil)
				So(pc[0].Name(), ShouldEqual, "facter")
			})
			Convey("does not unload & returns an error if it cannot load a plugin", func() {
				err := c.SwapPlugins("/fake/plugin/path", pc[0])
				So(err, ShouldNotBeNil)
				So(pc[0].Name(), ShouldEqual, "dummy")
			})
			Convey("rollsback loaded plugin & returns an error if it cannot unload a plugin", func() {
				dummy := pc[0]
				c.SwapPlugins(facterPath, dummy)
				pc = c.PluginCatalog()

				err := c.SwapPlugins(PluginPath, dummy)
				So(err, ShouldNotBeNil)
				So(pc[0].Name(), ShouldEqual, "facter")
			})
		})
	}
}

// Uses the dummy collector plugin to simulate Loading
func TestLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		Convey("pluginControl.Load", t, func() {

			Convey("loads successfully", func() {
				c := Control()
				c.Start()
				err := c.Load(PluginPath)

				So(c.pluginManager.LoadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)
			})

			Convey("returns error if not started", func() {
				c := Control()
				err := c.Load(PluginPath)

				So(len(c.pluginManager.LoadedPlugins.Table()), ShouldEqual, 0)
				So(err, ShouldNotBeNil)
			})

			Convey("adds to pluginControl.pluginManager.LoadedPlugins on successful load", func() {
				c := Control()
				c.Start()
				err := c.Load(PluginPath)

				So(err, ShouldBeNil)
				So(len(c.pluginManager.LoadedPlugins.Table()), ShouldBeGreaterThan, 0)
			})

		})

	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestStop(t *testing.T) {
	Convey("pluginControl.Stop", t, func() {
		c := Control()
		c.Start()
		c.Stop()

		Convey("stops", func() {
			So(c.Started, ShouldBeFalse)
		})

	})

}

func TestSubscribeMetric(t *testing.T) {
	Convey("adds a subscription", t, func() {
		c := Control()
		c.SubscribeMetric([]string{"test", "foo"})
		So(c.subscriptions.Count("test.foo"), ShouldEqual, 1)
	})
}

func TestUnsubscribeMetric(t *testing.T) {
	c := Control()
	Convey("When no error is returned", t, func() {
		c.SubscribeMetric([]string{"test", "foo"})
		Convey("it decrements a metric's subscriptions", func() {
			c.UnsubscribeMetric([]string{"test", "foo"})
			So(c.subscriptions.Count("test.foo"), ShouldEqual, 0)
		})
	})
	Convey("When an error is returned", t, func() {
		Convey("it panics", func() {
			So(func() { c.UnsubscribeMetric([]string{"test", "bar"}) }, ShouldPanic)
		})
	})
}

func TestPluginCatalog(t *testing.T) {
	ts := time.Now()

	c := Control()

	lp1 := new(loadedPlugin)
	lp1.Meta = plugin.PluginMeta{"test1", 1}
	lp1.Type = 0
	lp1.State = "loaded"
	lp1.LoadedTime = ts
	c.pluginManager.LoadedPlugins.Append(lp1)

	lp2 := new(loadedPlugin)
	lp2.Meta = plugin.PluginMeta{"test2", 1}
	lp2.Type = 0
	lp2.State = "loaded"
	lp2.LoadedTime = ts
	c.pluginManager.LoadedPlugins.Append(lp2)

	lp3 := new(loadedPlugin)
	lp3.Meta = plugin.PluginMeta{"test3", 1}
	lp3.Type = 0
	lp3.State = "loaded"
	lp3.LoadedTime = ts
	c.pluginManager.LoadedPlugins.Append(lp3)

	pc := c.PluginCatalog()

	Convey("it returns a list of CatalogedPlugins (PluginCatalog)", t, func() {
		So(pc, ShouldHaveSameTypeAs, PluginCatalog{})
	})

	Convey("the loadedPlugins implement the interface CatalogedPlugin interface", t, func() {
		So(lp1.Name(), ShouldEqual, "test1")
	})

}
