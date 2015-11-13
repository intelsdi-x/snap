/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/control_event"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/core/perror"
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
	ExistingPlugin core.CatalogedPlugin
	loadedPlugins  *loadedPlugins
}

func (m *MockPluginManagerBadSwap) LoadPlugin(string, gomit.Emitter) (*loadedPlugin, perror.PulseError) {
	return new(loadedPlugin), nil
}
func (m *MockPluginManagerBadSwap) UnloadPlugin(c core.Plugin) (*loadedPlugin, perror.PulseError) {
	return nil, perror.New(errors.New("fake"))
}
func (m *MockPluginManagerBadSwap) get(string) (*loadedPlugin, error) { return nil, nil }
func (m *MockPluginManagerBadSwap) teardown()                         {}
func (m *MockPluginManagerBadSwap) SetPluginConfig(*pluginConfig)     {}
func (m *MockPluginManagerBadSwap) SetMetricCatalog(catalogsMetrics)  {}
func (m *MockPluginManagerBadSwap) SetEmitter(gomit.Emitter)          {}
func (m *MockPluginManagerBadSwap) GenerateArgs(string) plugin.Arg    { return plugin.Arg{} }

func (m *MockPluginManagerBadSwap) all() map[string]*loadedPlugin {
	return m.loadedPlugins.table
}

func load(c *pluginControl, paths ...string) (core.CatalogedPlugin, perror.PulseError) {
	// This is a Travis optimized loading of plugins. From time to time, tests will error in Travis
	// due to a timeout when waiting for a response from a plugin. We are going to attempt loading a plugin
	// 3 times before letting the error through. Hopefully this cuts down on the number of Travis failures
	var e perror.PulseError
	var p core.CatalogedPlugin
	for i := 0; i < 3; i++ {
		p, e = c.Load(paths...)
		if e == nil {
			break
		}
		if e != nil && i == 2 {
			return nil, e

		}
	}
	return p, nil
}

func TestPluginControlGenerateArgs(t *testing.T) {
	Convey("pluginControl.Start", t, func() {
		c := New()
		Convey("starts successfully", func() {
			err := c.Start()
			So(c.Started, ShouldBeTrue)
			So(err, ShouldBeNil)
			n := c.Name()
			So(n, ShouldResemble, "control")
		})
		Convey("sets monitor duration", func() {
			c.SetMonitorOptions(MonitorDurationOption(time.Millisecond * 100))
			So(c.pluginRunner.Monitor().duration, ShouldResemble, 100*time.Millisecond)
		})
	})
}

func TestSwapPlugin(t *testing.T) {
	if PulsePath != "" {
		c := New()
		c.Start()
		time.Sleep(100 * time.Millisecond)
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginsSwapped", lpe)

		_, e := load(c, PluginPath)
		<-lpe.done
		// Load first plugin as a plugin is needed to be loaded in order to swap plugins
		Convey("Loading first plugin", t, func() {
			Convey("Should not error", func() {
				So(e, ShouldBeNil)
			})
			Convey("First plugin in catalog should have name mock2", func() {
				So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock2")
			})
		})
		mock1Path := strings.Replace(PluginPath, "pulse-collector-mock2", "pulse-collector-mock1", 1)
		err := c.SwapPlugins(mock1Path, c.PluginCatalog()[0])
		<-lpe.done

		// Swap plugin that was loaded with a different plugin
		Convey("Swapping plugins", t, func() {
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should generate a swapped plugins event", func() {
				Convey("So first plugin in catalog after swap should have name mock1", func() {
					So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock1")
				})
				Convey("So swapped plugins event should show loaded plugin name as mock1", func() {
					So(lpe.plugin.LoadedPluginName, ShouldEqual, "mock1")
				})
				Convey("So swapped plugins event should show loaded plugin version as 1", func() {
					So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 1)
				})
				Convey("So swapped plugins event should show unloaded plugin name as mock2", func() {
					So(lpe.plugin.UnloadedPluginName, ShouldEqual, "mock2")
				})
				Convey("So swapped plugins event should show unloaded plugin version as 2", func() {
					So(lpe.plugin.UnloadedPluginVersion, ShouldEqual, 2)
				})
				Convey("So swapped plugins event should show plugin type as collector", func() {
					So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
				})
			})
		})

		// Let's try swapping in a plugin that does not exist
		err = c.SwapPlugins("/fake/plugin/path", c.PluginCatalog()[0])
		Convey("Swapping in a plugin that does not exist", t, func() {
			Convey("Should throw an error", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Plugin in catalog should still be mock1", func() {
				So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock1")
			})
		})

		//
		// TODO: Write a proper rollback test as previous test was not testing rollback
		//

		// Rollback will throw an error if a plugin can not unload

		Convey("Rollback failure returns error", t, func() {
			lp := c.PluginCatalog()[0]
			pm := new(MockPluginManagerBadSwap)
			pm.ExistingPlugin = lp
			c.pluginManager = pm

			err := c.SwapPlugins(mock1Path, lp)
			Convey("So err should be received if rollback fails", func() {
				So(err, ShouldNotBeNil)
			})
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	}
}

type mockPluginEvent struct {
	LoadedPluginName      string
	LoadedPluginVersion   int
	UnloadedPluginName    string
	UnloadedPluginVersion int
	PluginType            int
}

type listenToPluginEvent struct {
	plugin *mockPluginEvent
	done   chan struct{}
}

func newListenToPluginEvent() *listenToPluginEvent {
	return &listenToPluginEvent{
		done:   make(chan struct{}),
		plugin: &mockPluginEvent{},
	}
}

func (l *listenToPluginEvent) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		l.plugin.LoadedPluginName = v.Name
		l.plugin.LoadedPluginVersion = v.Version
		l.plugin.PluginType = v.Type
		l.done <- struct{}{}
	case *control_event.UnloadPluginEvent:
		l.plugin.UnloadedPluginName = v.Name
		l.plugin.UnloadedPluginVersion = v.Version
		l.plugin.PluginType = v.Type
		l.done <- struct{}{}
	case *control_event.SwapPluginsEvent:
		l.plugin.LoadedPluginName = v.LoadedPluginName
		l.plugin.LoadedPluginVersion = v.LoadedPluginVersion
		l.plugin.UnloadedPluginName = v.UnloadedPluginName
		l.plugin.UnloadedPluginVersion = v.UnloadedPluginVersion
		l.plugin.PluginType = v.PluginType
		l.done <- struct{}{}
	case *control_event.PluginSubscriptionEvent:
		l.done <- struct{}{}
	default:
		fmt.Println("Got an event you're not handling")
	}
}

var (
	AciPath = path.Join(strings.TrimRight(PulsePath, "build"), "pkg/unpackage/")
	AciFile = "pulse-collector-plugin-mock1.darwin-x86_64.aci"
)

type mocksigningManager struct {
	signed bool
}

func (ps *mocksigningManager) ValidateSignature([]string, string, string) error {
	if ps.signed {
		return nil
	}
	return errors.New("fake")
}

// Uses the mock collector plugin to simulate Loading
func TestLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		c := New()

		// Testing trying to load before starting pluginControl
		Convey("pluginControl before being started", t, func() {
			_, err := load(c, PluginPath)
			Convey("should return an error when loading a plugin", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("and there should be no plugin loaded", func() {
				So(len(c.pluginManager.all()), ShouldEqual, 0)
			})
		})

		// Start pluginControl and load our mock plugin
		c.Start()
		time.Sleep(100 * time.Millisecond)
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, err := load(c, PluginPath)
		<-lpe.done

		Convey("pluginControl.Load on successful load", t, func() {
			Convey("should not return an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("should emit a plugin event message", func() {
				Convey("with loaded plugin name is mock2", func() {
					So(lpe.plugin.LoadedPluginName, ShouldEqual, "mock2")
				})

				Convey("with loaded plugin version as 2", func() {
					So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 2)
				})
				Convey("with loaded plugin type as collector", func() {
					So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
				})
			})
		})

		// Test trying to load a non-existant plugin
		_, err = c.Load(PluginPath + "foo")
		Convey("pluginControl.Load on bad plugin path", t, func() {
			Convey("should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		//Unpackaging
		Convey("pluginControl.Load on untar error with package", t, func() {
			PackagePath := path.Join(AciPath, "mock.aci")
			_, err := c.Load(PackagePath)
			Convey("should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		PackagePath := path.Join(AciPath, "noExec.aci")
		_, err = c.Load(PackagePath)
		Convey("pluginControl.Load on package with no executable file", t, func() {
			Convey("Should return error", func() {
				So(err, ShouldNotBeNil)
				Convey("And error should say 'Error no executable files found'", func() {
					So(err.Error(), ShouldContainSubstring, "Error no executable files found")
				})
			})
			Convey("And no plugin should be added to the pluginCatalog", func() {
				So(len(c.pluginManager.all()), ShouldNotEqual, 2)
			})
		})
		// Stop our controller so the plugins are unloaded and cleaned up from the system
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestLoadWithSignedPlugins(t *testing.T) {
	if PulsePath != "" {
		Convey("pluginControl.Load should successufully load a signed plugin with trust enabled", t, func() {
			c := New()
			c.pluginTrust = PluginTrustEnabled
			c.signingManager = &mocksigningManager{signed: true}
			lpe := newListenToPluginEvent()
			c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, PluginPath, "mock.asc")
			<-lpe.done
			Convey("so error on loading a signed plugin should be nil", func() {
				So(err, ShouldBeNil)
			})
			// Stop our controller to clean up our plugin
			c.Stop()
			time.Sleep(100 * time.Millisecond)
		})
		Convey("pluginControl.Load should successfully load unsigned plugin when trust level set to warning", t, func() {
			c := New()
			c.pluginTrust = PluginTrustWarn
			c.signingManager = &mocksigningManager{signed: false}
			lpe := newListenToPluginEvent()
			c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, PluginPath)
			<-lpe.done
			Convey("so error on loading an unsigned plugin should be nil", func() {
				So(err, ShouldBeNil)
			})
			c.Stop()
			time.Sleep(100 * time.Millisecond)
		})
		Convey("pluginControl.Load returns error with trust enabled and signing not validated", t, func() {
			c := New()
			c.pluginTrust = PluginTrustEnabled
			c.signingManager = &mocksigningManager{signed: false}
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, PluginPath)
			Convey("so error should not be nil when loading an unsigned plugin with trust enabled", func() {
				So(err, ShouldNotBeNil)
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
		c := New()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestUnload", lpe)
		c.Start()
		time.Sleep(100 * time.Millisecond)
		load(c, PluginPath)
		<-lpe.done
		Convey("Before calling unload, a plugin is loaded", t, func() {
			Convey("So pluginCatalog should not be empty before unload is called", func() {
				So(len(c.pluginManager.all()), ShouldEqual, 1)
			})
		})

		// Test unloading the plugin we just loaded
		pc := c.PluginCatalog()
		_, err := c.Unload(pc[0])
		<-lpe.done
		Convey("pluginControl.Unload when unloading a loaded plugin", t, func() {
			Convey("should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("should generate an unloaded plugin event", func() {
				Convey("where unloaded plugin name is mock2", func() {
					So(lpe.plugin.UnloadedPluginName, ShouldEqual, "mock2")
				})
				Convey("where unloaded plugin version should equal 2", func() {
					So(lpe.plugin.UnloadedPluginVersion, ShouldEqual, 2)
				})
				Convey("where unloaded plugin type should equal collector", func() {
					So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
				})
			})
		})

		// Test unloading the plugin again should result in an error
		_, err = c.Unload(pc[0])
		Convey("pluginControl.Unload when unloading a plugin that does not exist or has already been unloaded", t, func() {
			Convey("should return an error", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("and error should say 'plugin not found'", func() {
				So(err.Error(), ShouldResemble, "plugin not found")
			})
		})

		// Stop our controller
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestStop(t *testing.T) {
	Convey("pluginControl.Stop", t, func() {
		c := New()
		lps := newLoadedPlugins()
		err := lps.add(&loadedPlugin{
			Type: plugin.CollectorPluginType,
			Meta: plugin.PluginMeta{
				Name:    "bad-swap",
				Version: 1,
			},
		})
		So(err, ShouldBeNil)
		c.pluginManager = &MockPluginManagerBadSwap{loadedPlugins: lps}
		c.Start()
		So(c.pluginManager.all(), ShouldNotBeEmpty)
		c.Stop()

		Convey("stops", func() {
			So(c.Started, ShouldBeFalse)
		})
	})
}

func TestPluginCatalog(t *testing.T) {
	ts := time.Now()

	c := New()

	// We need our own plugin manager to drop mock
	// loaded plugins into.  Aribitrarily adding
	// plugins from the pm is no longer supported.
	tpm := newPluginManager()
	c.pluginManager = tpm

	lp1 := new(loadedPlugin)
	lp1.Meta = plugin.PluginMeta{Name: "test1",
		Version:              1,
		AcceptedContentTypes: []string{"a", "b", "c"},
		ReturnedContentTypes: []string{"a", "b", "c"},
	}
	lp1.Type = 0
	lp1.State = "loaded"
	lp1.LoadedTime = ts
	tpm.loadedPlugins.add(lp1)

	lp2 := new(loadedPlugin)
	lp2.Meta = plugin.PluginMeta{Name: "test2", Version: 1}
	lp2.Type = 0
	lp2.State = "loaded"
	lp2.LoadedTime = ts
	tpm.loadedPlugins.add(lp2)

	lp3 := new(loadedPlugin)
	lp3.Meta = plugin.PluginMeta{Name: "test3", Version: 1}
	lp3.Type = 0
	lp3.State = "loaded"
	lp3.LoadedTime = ts
	tpm.loadedPlugins.add(lp3)

	lp4 := new(loadedPlugin)
	lp4.Meta = plugin.PluginMeta{Name: "test1",
		Version:              4,
		AcceptedContentTypes: []string{"d", "e", "f"},
		ReturnedContentTypes: []string{"d", "e", "f"},
	}
	lp4.Type = 0
	lp4.State = "loaded"
	lp4.LoadedTime = ts
	tpm.loadedPlugins.add(lp4)

	lp5 := new(loadedPlugin)
	lp5.Meta = plugin.PluginMeta{Name: "test1",
		Version:              0,
		AcceptedContentTypes: []string{"d", "e", "f"},
		ReturnedContentTypes: []string{"d", "e", "f"},
	}
	lp5.Type = 0
	lp5.State = "loaded"
	lp5.LoadedTime = ts
	tpm.loadedPlugins.add(lp5)

	pc := c.PluginCatalog()

	Convey("it returns a list of CatalogedPlugins (PluginCatalog)", t, func() {
		So(pc, ShouldHaveSameTypeAs, core.PluginCatalog{})
	})

	Convey("the loadedPlugins implement the interface CatalogedPlugin interface", t, func() {
		So(lp1.Name(), ShouldEqual, "test1")
	})

	Convey("GetPluginContentTypes", t, func() {
		Convey("Given a plugin that exists", func() {
			act, ret, err := c.GetPluginContentTypes("test1", core.PluginType(0), 1)
			So(err, ShouldBeNil)
			So(act, ShouldResemble, []string{"a", "b", "c"})
			So(ret, ShouldResemble, []string{"a", "b", "c"})
		})
		Convey("Given a plugin with a version that does NOT exist", func() {
			act, ret, err := c.GetPluginContentTypes("test1", core.PluginType(0), 5)
			So(err, ShouldNotBeNil)
			So(act, ShouldBeEmpty)
			So(ret, ShouldBeEmpty)
		})
		Convey("Given a plugin where the version provided is 0", func() {
			act, ret, err := c.GetPluginContentTypes("test1", core.PluginType(0), 0)
			So(err, ShouldBeNil)
			So(act, ShouldResemble, []string{"d", "e", "f"})
			So(ret, ShouldResemble, []string{"d", "e", "f"})
		})
		Convey("Given no plugins for the name and type", func() {
			act, ret, err := c.GetPluginContentTypes("test9", core.PluginType(0), 5)
			So(err, ShouldNotBeNil)
			So(act, ShouldBeEmpty)
			So(ret, ShouldBeEmpty)
		})
	})

}

type mc struct {
	e int
}

func (m *mc) Fetch(ns []string) ([]*metricType, perror.PulseError) {
	if m.e == 2 {
		return nil, perror.New(errors.New("test"))
	}
	return nil, nil
}

func (m *mc) resolvePlugin(mns []string, ver int) (*loadedPlugin, error) {
	return nil, nil
}

func (m *mc) GetPlugin([]string, int) (*loadedPlugin, perror.PulseError) {
	return nil, nil
}

func (m *mc) Get(ns []string, ver int) (*metricType, perror.PulseError) {
	if m.e == 1 {
		return &metricType{
			policy: &mockCDProc{},
		}, nil
	}
	return nil, perror.New(errorMetricNotFound(ns))
}

func (m *mc) Subscribe(ns []string, ver int) perror.PulseError {
	if ns[0] == "nf" {
		return perror.New(errorMetricNotFound(ns))
	}
	return nil
}

func (m *mc) Unsubscribe(ns []string, ver int) perror.PulseError {
	if ns[0] == "nf" {
		return perror.New(errorMetricNotFound(ns))
	}
	if ns[0] == "neg" {
		return errNegativeSubCount
	}
	return nil
}

func (m *mc) Add(*metricType)                 {}
func (m *mc) Table() map[string][]*metricType { return map[string][]*metricType{} }
func (m *mc) Item() (string, []*metricType)   { return "", []*metricType{} }

func (m *mc) Next() bool {
	m.e = 1
	return false
}

func (m *mc) AddLoadedMetricType(*loadedPlugin, core.Metric) {

}

func (m *mc) RmUnloadedPluginMetrics(lp *loadedPlugin) {

}

type mockCDProc struct {
}

func (m *mockCDProc) Process(in map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *cpolicy.ProcessingErrors) {
	if _, ok := in["fail"]; ok {
		pe := cpolicy.NewProcessingErrors()
		pe.AddError(errors.New("test fail"))
		return nil, pe
	}
	return &in, nil
}

func (m *mockCDProc) HasRules() bool {
	return true
}

// TODO move to metricCatalog
// func TestResolvePlugin(t *testing.T) {
// 	Convey(".resolvePlugin()", t, func() {
// 		c := New()
// 		lp := &loadedPlugin{}
// 		mt := newMetricType([]string{"foo", "bar"}, time.Now(), lp)
// 		c.metricCatalog.Add(mt)
// 		Convey("it resolves the plugin", func() {
// 			p, err := c.resolvePlugin([]string{"foo", "bar"}, -1)
// 			So(err, ShouldBeNil)
// 			So(p, ShouldEqual, lp)
// 		})
// 		Convey("it returns an error if the metricType cannot be found", func() {
// 			p, err := c.resolvePlugin([]string{"baz", "qux"}, -1)
// 			So(p, ShouldBeNil)
// 			So(err, ShouldResemble, errors.New("metric not found"))
// 		})
// 	})
// }

func TestExportedMetricCatalog(t *testing.T) {
	Convey(".MetricCatalog()", t, func() {
		c := New()
		lp := &loadedPlugin{}
		mt := newMetricType([]string{"foo", "bar"}, time.Now(), lp)
		c.metricCatalog.Add(mt)
		Convey("it returns a collection of core.MetricTypes", func() {
			t, err := c.MetricCatalog()
			So(err, ShouldBeNil)
			So(len(t), ShouldEqual, 1)
			So(t[0].Namespace(), ShouldResemble, []string{"foo", "bar"})
		})
		Convey("If metric catalog fetch fails", func() {
			c.metricCatalog = &mc{e: 2}
			mts, err := c.MetricCatalog()
			So(mts, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMetricExists(t *testing.T) {
	Convey("MetricExists()", t, func() {
		c := New()
		c.metricCatalog = &mc{}
		So(c.MetricExists([]string{"hi"}, -1), ShouldEqual, false)
	})
}

type MockMetricType struct {
	namespace []string
	cfg       *cdata.ConfigDataNode
	ver       int
}

func (m MockMetricType) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Namespace []string              `json:"namespace"`
		Config    *cdata.ConfigDataNode `json:"config"`
	}{
		Namespace: m.namespace,
		Config:    m.cfg,
	})
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) Source() string {
	return ""
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Timestamp() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return m.ver
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return m.cfg
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func (m MockMetricType) Labels() []core.Label    { return nil }
func (m MockMetricType) Tags() map[string]string { return nil }

func TestMetricConfig(t *testing.T) {
	Convey("required config provided by task", t, func() {
		c := New()
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		load(c, JSONRPC_PluginPath)
		<-lpe.done
		cd := cdata.NewNode()
		m1 := MockMetricType{
			namespace: []string{"intel", "mock", "foo"},
		}
		metric, errs := c.validateMetricTypeSubscription(m1, cd)
		Convey("So metric should not be valid without config", func() {
			So(metric, ShouldBeNil)
			So(errs, ShouldNotBeNil)
		})
		cd.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		metric, errs = c.validateMetricTypeSubscription(m1, cd)
		Convey("So metric should be valid with config", func() {
			So(errs, ShouldBeNil)
			So(metric, ShouldNotBeNil)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
	Convey("nil config provided by task", t, func() {
		config := NewConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(OptSetConfig(config))
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		load(c, JSONRPC_PluginPath)
		<-lpe.done
		var cd *cdata.ConfigDataNode
		m1 := MockMetricType{
			namespace: []string{"intel", "mock", "foo"},
		}
		metric, errs := c.validateMetricTypeSubscription(m1, cd)
		Convey("So metric should be valid with config", func() {
			So(errs, ShouldBeNil)
			So(metric, ShouldNotBeNil)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
	Convey("required config provided by global plugin config", t, func() {
		config := NewConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(OptSetConfig(config))
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		load(c, JSONRPC_PluginPath)
		<-lpe.done
		cd := cdata.NewNode()
		m1 := MockMetricType{
			namespace: []string{"intel", "mock", "foo"},
			ver:       1,
		}
		metric, errs := c.validateMetricTypeSubscription(m1, cd)
		Convey("So metric should be valid with config", func() {
			So(errs, ShouldBeNil)
			So(metric, ShouldNotBeNil)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestCollectDynamicMetrics(t *testing.T) {
	Convey("given a plugin using the native client", t, func() {
		config := NewConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(OptSetConfig(config), CacheExpiration(time.Second*1))
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		load(c, PluginPath)
		<-lpe.done
		load(c, JSONRPC_PluginPath)
		<-lpe.done
		cd := cdata.NewNode()
		metrics, err := c.metricCatalog.Fetch([]string{})
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 6)
		m, err := c.metricCatalog.Get([]string{"intel", "mock", "*", "baz"}, 2)
		So(err, ShouldBeNil)
		So(m, ShouldNotBeNil)
		jsonm, err := c.metricCatalog.Get([]string{"intel", "mock", "*", "baz"}, 1)
		So(err, ShouldBeNil)
		So(jsonm, ShouldNotBeNil)
		metric, errs := c.validateMetricTypeSubscription(m, cd)
		So(errs, ShouldBeNil)
		So(metric, ShouldNotBeNil)
		Convey("collects metrics from plugin using native client", func() {
			lp, err := c.pluginManager.get("collector:mock2:2")
			So(err, ShouldBeNil)
			So(lp, ShouldNotBeNil)
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock2:2")
			So(errp, ShouldBeNil)
			So(pool, ShouldNotBeNil)
			pool.subscribe("1", unboundSubscriptionType)
			err = c.pluginRunner.runPlugin(lp.Path)
			So(err, ShouldBeNil)
			mts, errs := c.CollectMetrics([]core.Metric{m}, time.Now().Add(time.Second*1))
			hits, err := pool.plugins[1].client.CacheHits(core.JoinNamespace(m.namespace), 2)
			So(err, ShouldBeNil)
			So(hits, ShouldEqual, 0)
			So(errs, ShouldBeNil)
			So(len(mts), ShouldEqual, 10)
			mts, errs = c.CollectMetrics([]core.Metric{m}, time.Now().Add(time.Second*1))
			hits, err = pool.plugins[1].client.CacheHits(core.JoinNamespace(m.namespace), 2)
			So(err, ShouldBeNil)
			So(hits, ShouldEqual, 1)
			So(errs, ShouldBeNil)
			So(len(mts), ShouldEqual, 10)
			pool.unsubscribe("1")
			Convey("collects metrics from plugin using httpjson client", func() {
				lp, err := c.pluginManager.get("collector:mock1:1")
				So(err, ShouldBeNil)
				So(lp, ShouldNotBeNil)
				pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock1:1")
				So(errp, ShouldBeNil)
				So(pool, ShouldNotBeNil)
				pool.subscribe("1", unboundSubscriptionType)
				err = c.pluginRunner.runPlugin(lp.Path)
				So(err, ShouldBeNil)
				mts, errs := c.CollectMetrics([]core.Metric{jsonm}, time.Now().Add(time.Second*1))
				hits, err := pool.plugins[1].client.CacheHits(core.JoinNamespace(m.namespace), 1)
				So(err, ShouldBeNil)
				So(hits, ShouldEqual, 0)
				So(errs, ShouldBeNil)
				So(len(mts), ShouldEqual, 10)
				mts, errs = c.CollectMetrics([]core.Metric{jsonm}, time.Now().Add(time.Second*1))
				hits, err = pool.plugins[1].client.CacheHits(core.JoinNamespace(m.namespace), 1)
				So(err, ShouldBeNil)
				So(hits, ShouldEqual, 1)
				So(errs, ShouldBeNil)
				So(len(mts), ShouldEqual, 10)
				hits = pool.plugins[1].client.AllCacheHits()
				So(hits, ShouldEqual, 2)
				misses := pool.plugins[1].client.AllCacheMisses()
				So(misses, ShouldEqual, 2)
				pool.unsubscribe("1")
				c.Stop()
				time.Sleep(100 * time.Millisecond)
			})
		})
	})
}

func TestCollectMetrics(t *testing.T) {

	Convey("given a new router", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		config := NewConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(OptSetConfig(config))
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		time.Sleep(100 * time.Millisecond)

		// Add a global plugin config
		c.Config.Plugins.Collector.Plugins["mock1"] = newPluginConfigItem(optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}))

		// Load plugin
		load(c, JSONRPC_PluginPath)
		<-lpe.done
		mts, err := c.MetricCatalog()
		So(err, ShouldBeNil)
		So(len(mts), ShouldEqual, 4)

		cd := cdata.NewNode()
		cd.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})

		m := []core.Metric{}
		m1 := MockMetricType{
			namespace: []string{"intel", "mock", "foo"},
			cfg:       cd,
		}
		m2 := MockMetricType{
			namespace: []string{"intel", "mock", "bar"},
			cfg:       cd,
		}
		m3 := MockMetricType{
			namespace: []string{"intel", "mock", "test"},
			cfg:       cd,
		}

		// retrieve loaded plugin
		lp, err := c.pluginManager.get("collector:mock1:1")
		So(err, ShouldBeNil)
		So(lp, ShouldNotBeNil)
		pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock1:1")
		So(errp, ShouldBeNil)
		pool.subscribe("1", unboundSubscriptionType)
		err = c.pluginRunner.runPlugin(lp.Path)
		So(err, ShouldBeNil)
		m = append(m, m1, m2, m3)
		time.Sleep(time.Millisecond * 1100)

		for x := 0; x < 5; x++ {
			cr, err := c.CollectMetrics(m, time.Now().Add(time.Second*1))
			So(err, ShouldBeNil)
			for i := range cr {
				So(cr[i].Data(), ShouldContainSubstring, "The mock collected data!")
				So(cr[i].Data(), ShouldContainSubstring, "test=true")
			}
		}
		ap := c.AvailablePlugins()
		So(ap, ShouldNotBeEmpty)
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})

	// Not sure what this was supposed to test, because it's actually testing nothing
	SkipConvey("Pool", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1
		// Create controller
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		load(c, PluginPath)
		m := []core.Metric{}
		c.CollectMetrics(m, time.Now().Add(time.Second*60))
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
}

type mockMetric struct {
	namespace []string
	data      int
}

func (m *mockMetric) Namespace() []string {
	return m.namespace
}

func (m *mockMetric) Data() interface{} {
	return m.data
}

type mockPlugin struct {
	pluginType core.PluginType
	name       string
	ver        int
	config     *cdata.ConfigDataNode
}

func (m mockPlugin) Name() string                  { return m.name }
func (m mockPlugin) TypeName() string              { return m.pluginType.String() }
func (m mockPlugin) Version() int                  { return m.ver }
func (m mockPlugin) Config() *cdata.ConfigDataNode { return m.config }

func TestPublishMetrics(t *testing.T) {
	Convey("Given an available file publisher plugin", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		config := NewConfig()
		c := New(OptSetConfig(config))
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestPublishMetrics", lpe)
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		time.Sleep(1 * time.Second)

		// Load plugin
		_, err := load(c, path.Join(PulsePath, "plugin", "pulse-publisher-file"))
		<-lpe.done
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("publisher:file:3")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "file")
		So(lp.ConfigPolicy, ShouldNotBeNil)

		Convey("Subscribe to file publisher with good config", func() {
			n := cdata.NewNode()
			config.Plugins.Publisher.Plugins[lp.Name()] = newPluginConfigItem(optAddPluginConfigItem("file", ctypes.ConfigValueStr{Value: "/tmp/pulse-TestPublishMetrics.out"}))
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("publisher:file:3")
			So(errp, ShouldBeNil)
			pool.subscribe("1", unboundSubscriptionType)
			err := c.pluginRunner.runPlugin(lp.Path)
			So(err, ShouldBeNil)
			time.Sleep(2500 * time.Millisecond)

			Convey("Publish to file", func() {
				metrics := []plugin.PluginMetricType{
					*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "", nil, nil, 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.PulseGOBContentType
				errs := c.PublishMetrics(contentType, buf.Bytes(), "file", 3, n.Table())
				So(errs, ShouldBeNil)
				ap := c.AvailablePlugins()
				So(ap, ShouldNotBeEmpty)
			})
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)

	})
}
func TestProcessMetrics(t *testing.T) {
	Convey("Given an available file processor plugin", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		c := New()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestProcessMetrics", lpe)
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		time.Sleep(1 * time.Second)
		c.Config.Plugins.Processor.Plugins["passthru"] = newPluginConfigItem(optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}))

		// Load plugin
		_, err := load(c, path.Join(PulsePath, "plugin", "pulse-processor-passthru"))
		<-lpe.done
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("processor:passthru:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "passthru")
		So(lp.ConfigPolicy, ShouldNotBeNil)

		Convey("Subscribe to passthru processor with good config", func() {
			n := cdata.NewNode()
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("processor:passthru:1")
			So(errp, ShouldBeNil)
			pool.subscribe("1", unboundSubscriptionType)
			err := c.pluginRunner.runPlugin(lp.Path)
			So(err, ShouldBeNil)
			time.Sleep(2500 * time.Millisecond)

			Convey("process metrics", func() {
				metrics := []plugin.PluginMetricType{
					*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "", nil, nil, 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.PulseGOBContentType
				_, ct, errs := c.ProcessMetrics(contentType, buf.Bytes(), "passthru", 1, n.Table())
				So(errs, ShouldBeEmpty)
				mts := []plugin.PluginMetricType{}
				dec := gob.NewDecoder(bytes.NewBuffer(ct))
				err := dec.Decode(&mts)
				So(err, ShouldBeNil)
				So(mts[0].Data_, ShouldEqual, 2)
				So(errs, ShouldBeNil)
			})
		})

		Convey("Count()", func() {
			pmt := &pluginMetricTypes{}
			count := pmt.Count()
			So(count, ShouldResemble, 0)

		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)

	})
}
