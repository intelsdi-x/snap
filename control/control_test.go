package control

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

// Mock Executor used to test
type MockPluginExecutor struct {
	Killed          bool
	Response        string
	WaitTime        time.Duration
	WaitError       error
	WaitForResponse func(time.Duration) (*plugin.Response, error)
}

// Mock plugin manager that will fail swap on the last rollback for testing rollback failure is caught
type MockPluginManagerBadSwap struct {
	Mode           int
	ExistingPlugin CatalogedPlugin
}

func (m *MockPluginManagerBadSwap) LoadPlugin(string) (*loadedPlugin, error) {
	return new(loadedPlugin), nil
}

func (m *MockPluginManagerBadSwap) UnloadPlugin(c CatalogedPlugin) error {
	return errors.New("fake")
}

func (m *MockPluginManagerBadSwap) LoadedPlugins() *loadedPlugins {
	return nil
}

func TestControlNew(t *testing.T) {

}

func TestPluginControlGenerateArgs(t *testing.T) {
	Convey("pluginControl.Start", t, func() {
		Convey("starts successfully", func() {
			c := New()
			c.Start()
			So(c.Started, ShouldBeTrue)
		})
	})
}

func TestPluginControlStart(t *testing.T) {
	Convey("pluginControl.generateArgs", t, func() {
		Convey("returns arg", func() {
			c := New()
			c.Start()
			a := c.generateArgs()
			So(a, ShouldNotBeNil)
		})
	})
}

func TestSwapPlugin(t *testing.T) {
	if PulsePath != "" {
		Convey("SwapPlugin", t, func() {
			c := New()
			c.Start()
			e := c.Load(PluginPath)

			So(e, ShouldBeNil)

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

			Convey("rollback failure returns error", func() {
				dummy := pc[0]
				pm := new(MockPluginManagerBadSwap)
				pm.ExistingPlugin = dummy
				c.pluginManager = pm

				err := c.SwapPlugins(facterPath, dummy)
				So(err, ShouldNotBeNil)
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
				c := New()
				c.Start()
				err := c.Load(PluginPath)

				So(c.pluginManager.LoadedPlugins(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)
			})

			Convey("returns error if not started", func() {
				c := New()
				err := c.Load(PluginPath)

				So(len(c.pluginManager.LoadedPlugins().Table()), ShouldEqual, 0)
				So(err, ShouldNotBeNil)
			})

			Convey("adds to pluginControl.pluginManager.LoadedPlugins on successful load", func() {
				c := New()
				c.Start()
				err := c.Load(PluginPath)

				So(err, ShouldBeNil)
				So(len(c.pluginManager.LoadedPlugins().Table()), ShouldBeGreaterThan, 0)
			})

			Convey("returns error from pluginManager.LoadPlugin()", func() {
				c := New()
				c.Start()
				err := c.Load(PluginPath + "foo")

				So(err, ShouldNotBeNil)
				// So(len(c.pluginManager.LoadedPlugins.Table()), ShouldBeGreaterThan, 0)
			})

		})
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestUnload(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		Convey("pluginControl.Unload", t, func() {
			Convey("unloads successfully", func() {
				c := New()
				c.Start()
				err := c.Load(PluginPath)

				So(c.pluginManager.LoadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)

				pc := c.PluginCatalog()

				So(len(pc), ShouldEqual, 1)
				err = c.Unload(pc[0])
				So(err, ShouldBeNil)
			})

			Convey("returns error on unload for unknown plugin(or already unloaded)", func() {
				c := New()
				c.Start()
				err := c.Load(PluginPath)

				So(c.pluginManager.LoadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)

				pc := c.PluginCatalog()

				So(len(pc), ShouldEqual, 1)
				err = c.Unload(pc[0])
				So(err, ShouldBeNil)
				err = c.Unload(pc[0])
				So(err, ShouldResemble, errors.New("plugin [dummy] -- [1] not found (has it already been unloaded?)"))
			})
		})

	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestStop(t *testing.T) {
	Convey("pluginControl.Stop", t, func() {
		c := New()
		c.Start()
		c.Stop()

		Convey("stops", func() {
			So(c.Started, ShouldBeFalse)
		})

	})

}

func TestSubscribeMetric(t *testing.T) {
	Convey("adds a subscription", t, func() {
		c := New()
		c.SubscribeMetric([]string{"test", "foo"})
		So(c.subscriptions.Count("test.foo"), ShouldEqual, 1)
	})
}

func TestUnsubscribeMetric(t *testing.T) {
	c := New()
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

	c := New()

	lp1 := new(loadedPlugin)
	lp1.Meta = plugin.PluginMeta{Name: "test1", Version: 1}
	lp1.Type = 0
	lp1.State = "loaded"
	lp1.LoadedTime = ts
	c.pluginManager.LoadedPlugins().Append(lp1)

	lp2 := new(loadedPlugin)
	lp2.Meta = plugin.PluginMeta{Name: "test2", Version: 1}
	lp2.Type = 0
	lp2.State = "loaded"
	lp2.LoadedTime = ts
	c.pluginManager.LoadedPlugins().Append(lp2)

	lp3 := new(loadedPlugin)
	lp3.Meta = plugin.PluginMeta{Name: "test3", Version: 1}
	lp3.Type = 0
	lp3.State = "loaded"
	lp3.LoadedTime = ts
	c.pluginManager.LoadedPlugins().Append(lp3)

	pc := c.PluginCatalog()

	Convey("it returns a list of CatalogedPlugins (PluginCatalog)", t, func() {
		So(pc, ShouldHaveSameTypeAs, PluginCatalog{})
	})

	Convey("the loadedPlugins implement the interface CatalogedPlugin interface", t, func() {
		So(lp1.Name(), ShouldEqual, "test1")
	})

}

func TestResolvePlugin(t *testing.T) {
	Convey(".resolvePlugin()", t, func() {
		c := Control()
		lp := &loadedPlugin{}
		mt := newMetricType([]string{"foo", "bar"}, time.Now().Unix(), lp)
		c.metricCatalog.Add(mt)
		Convey("it resolves the plugin", func() {
			p, err := c.resolvePlugin([]string{"foo", "bar"})
			So(err, ShouldBeNil)
			So(p, ShouldEqual, lp)
		})
		Convey("it returns an error if the metricType cannot be found", func() {
			p, err := c.resolvePlugin([]string{"baz", "qux"})
			So(p, ShouldBeNil)
			So(err, ShouldResemble, errors.New("metric not found"))
		})
	})
}

func TestExportedMetricCatalog(t *testing.T) {
	Convey(".MetricCatalog()", t, func() {
		c := Control()
		lp := &loadedPlugin{}
		mt := newMetricType([]string{"foo", "bar"}, time.Now().Unix(), lp)
		c.metricCatalog.Add(mt)
		Convey("it returns a collection of core.MetricTypes", func() {
			t := c.MetricCatalog()
			So(len(t), ShouldEqual, 1)
			So(t[0].Namespace(), ShouldResemble, []string{"foo", "bar"})
		})
	})
}

type mc struct {
	e int
}

func (m *mc) Get([]string) (*metricType, error) {
	if m.e == 1 {
		return &metricType{}, nil
	}
	return nil, metricNotFound
}

func (m *mc) Add(*metricType)               {}
func (m *mc) Table() map[string]*metricType { return map[string]*metricType{} }
func (m *mc) Item() (string, *metricType)   { return "", &metricType{} }

func (m *mc) Next() bool {
	m.e = 1
	return false
}

func TestMetricExists(t *testing.T) {
	Convey("MetricExists()", t, func() {
		c := Control()
		c.metricCatalog = &mc{}
		So(c.MetricExists([]string{"hi"}), ShouldEqual, false)
		c.metricCatalog.Next()
		So(c.MetricExists([]string{"hi"}), ShouldEqual, true)
	})
}
