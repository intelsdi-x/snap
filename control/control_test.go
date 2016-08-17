// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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
	"errors"
	"fmt"
	"math/rand"
	"net"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vrischmann/jsonutil"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
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

func (m *MockPluginManagerBadSwap) LoadPlugin(*pluginDetails, gomit.Emitter) (*loadedPlugin, serror.SnapError) {
	return new(loadedPlugin), nil
}
func (m *MockPluginManagerBadSwap) UnloadPlugin(c core.Plugin) (*loadedPlugin, serror.SnapError) {
	return nil, serror.New(errors.New("fake"))
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

func load(c *pluginControl, paths ...string) (core.CatalogedPlugin, serror.SnapError) {
	// This is a Travis optimized loading of plugins. From time to time, tests will error in Travis
	// due to a timeout when waiting for a response from a plugin. We are going to attempt loading a plugin
	// 3 times before letting the error through. Hopefully this cuts down on the number of Travis failures
	var e serror.SnapError
	var p core.CatalogedPlugin
	rp, err := core.NewRequestedPlugin(paths[0])
	if err != nil {
		return nil, serror.New(err)
	}
	if len(paths) > 1 {
		rp.SetSignature([]byte{00, 00, 00})
	}
	for i := 0; i < 3; i++ {
		p, e = c.Load(rp)
		if e == nil {
			break
		}
		if e != nil && i == 2 {
			return nil, e

		}
	}
	return p, nil
}

// Generates a config to use for testing by taking the a default config
// and setting the ListenPort to an available port on the system running
// the tests.
func getTestConfig() *Config {
	config := GetDefaultConfig()
	config.ListenPort = getPort()
	return config
}

// Returns an available port on the test system. If no port is available
// after 1000 tries, then the test will panic.
func getPort() int {
	count := 0
	for count < 1000 {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err == nil {
			p := ln.Addr().(*net.TCPAddr).Port
			ln.Close()
			return p
		}
		count++
	}
	panic("Could not find an available port")
}

func TestPluginControlGenerateArgs(t *testing.T) {
	c := New(getTestConfig())
	err := c.Start()
	Convey("pluginControl starts successfully", t, func() {
		So(c.Started, ShouldBeTrue)
		So(err, ShouldBeNil)
		So(c.Name(), ShouldResemble, "control")
	})
	c.SetMonitorOptions(MonitorDurationOption(time.Millisecond * 100))
	Convey("sets monitor duration", t, func() {
		So(c.pluginRunner.Monitor().duration, ShouldResemble, 100*time.Millisecond)
	})
	c.Stop()
}

func TestSwapPlugin(t *testing.T) {
	if fixtures.SnapPath != "" {
		c := New(getTestConfig())
		c.Start()
		time.Sleep(100 * time.Millisecond)
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginsSwapped", lpe)

		_, e := load(c, fixtures.PluginPath)
		Convey("Loading first plugin", t, func() {
			Convey("Should not error", func() {
				So(e, ShouldBeNil)
			})
		})
		// Travis optimization: If for some reason we can not load
		// the plugin three times, we will fail the test immediately
		// as we wait on a channel to be closed before proceeding with
		// additional tests. If the plugin never loads, the channel will not
		// close and just hang the test indefinitely.
		if e != nil {
			t.FailNow()
		}
		<-lpe.done
		Convey("First plugin in catalog", t, func() {
			Convey("Should have name mock", func() {
				So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock")
			})
		})
		mock1Path := strings.Replace(fixtures.PluginPath, "snap-plugin-collector-mock2", "snap-plugin-collector-mock1", 1)
		mockRP, mErr := core.NewRequestedPlugin(mock1Path)
		Convey("Loading a plugin should not error", t, func() {
			So(mErr, ShouldBeNil)
		})
		err := c.SwapPlugins(mockRP, c.PluginCatalog()[0])
		Convey("Swapping plugins", t, func() {
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
		})
		if err != nil {
			t.FailNow()
		}
		<-lpe.done

		// Swap plugin that was loaded with a different version of the plugin
		Convey("Swapping plugins", t, func() {
			Convey("Should generate a swapped plugins event", func() {
				Convey("So first plugin in catalog after swap should have name mock", func() {
					So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock")
				})
				Convey("So swapped plugins event should show loaded plugin name as mock", func() {
					So(lpe.plugin.LoadedPluginName, ShouldEqual, "mock")
				})
				Convey("So swapped plugins event should show loaded plugin version as 1", func() {
					So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 1)
				})
				Convey("So swapped plugins event should show unloaded plugin name as mock", func() {
					So(lpe.plugin.UnloadedPluginName, ShouldEqual, "mock")
				})
				Convey("So swapped plugins event should show unloaded plugin version as 2", func() {
					So(lpe.plugin.UnloadedPluginVersion, ShouldEqual, 2)
				})
				Convey("So swapped plugins event should show plugin type as collector", func() {
					So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
				})
			})
		})

		// Swap plugin with a different type of plugin
		Convey("First plugin in catalog", t, func() {
			Convey("Should have name mock", func() {
				So(c.PluginCatalog()[0].Name(), ShouldEqual, "mock")
			})
		})

		filePath := strings.Replace(fixtures.PluginPath, "snap-plugin-collector-mock2", "snap-plugin-publisher-mock-file", 1)
		fileRP, pErr := core.NewRequestedPlugin(filePath)
		Convey("Loading a plugin should not error", t, func() {
			So(pErr, ShouldBeNil)
		})
		err = c.SwapPlugins(fileRP, c.PluginCatalog()[0])
		Convey("Swapping mock and file plugins", t, func() {
			Convey("Should error", func() {
				So(err, ShouldNotBeNil)
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

			mockRP, mErr := core.NewRequestedPlugin(mock1Path)
			So(mErr, ShouldBeNil)
			err := c.SwapPlugins(mockRP, lp)
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
	EventNamespace        string
}

type listenToPluginEvent struct {
	plugin    *mockPluginEvent
	done      chan struct{}
	max       chan struct{}
	restarted chan struct{}
}

func newListenToPluginEvent() *listenToPluginEvent {
	return &listenToPluginEvent{
		done:      make(chan struct{}),
		restarted: make(chan struct{}),
		max:       make(chan struct{}),
		plugin:    &mockPluginEvent{},
	}
}

func (l *listenToPluginEvent) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.RestartedAvailablePluginEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.restarted <- struct{}{}
	case *control_event.MaxPluginRestartsExceededEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.max <- struct{}{}
	case *control_event.DeadAvailablePluginEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	case *control_event.HealthCheckFailedEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	case *control_event.LoadPluginEvent:
		l.plugin.LoadedPluginName = v.Name
		l.plugin.LoadedPluginVersion = v.Version
		l.plugin.PluginType = v.Type
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	case *control_event.UnloadPluginEvent:
		l.plugin.UnloadedPluginName = v.Name
		l.plugin.UnloadedPluginVersion = v.Version
		l.plugin.PluginType = v.Type
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	case *control_event.SwapPluginsEvent:
		l.plugin.LoadedPluginName = v.LoadedPluginName
		l.plugin.LoadedPluginVersion = v.LoadedPluginVersion
		l.plugin.UnloadedPluginName = v.UnloadedPluginName
		l.plugin.UnloadedPluginVersion = v.UnloadedPluginVersion
		l.plugin.PluginType = v.PluginType
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	case *control_event.PluginSubscriptionEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.done <- struct{}{}
	default:
		controlLogger.WithFields(log.Fields{
			"event:": v.Namespace(),
			"_block": "HandleGomit",
		}).Info("Unhandled Event")
	}
}

var (
	AciPath = path.Join(strings.TrimRight(fixtures.SnapPath, "build"), "pkg/unpackage/")
	AciFile = "snap-collector-plugin-mock1.darwin-x86_64.aci"
)

type mocksigningManager struct {
	signed bool
}

func (ps *mocksigningManager) ValidateSignature([]string, string, []byte) error {
	if ps.signed {
		return nil
	}
	return errors.New("fake")
}

// Uses the mock collector plugin to simulate Loading
func TestLoad(t *testing.T) {
	// These tests only work if SNAP_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if fixtures.SnapPath != "" {
		c := New(getTestConfig())

		// Testing trying to load before starting pluginControl
		Convey("pluginControl before being started", t, func() {
			_, err := load(c, fixtures.PluginPath)
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
		_, err := load(c, fixtures.PluginPath)
		Convey("pluginControl.Load on successful load", t, func() {
			Convey("Should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})
		if err != nil {
			t.FailNow()
		}
		<-lpe.done

		Convey("pluginControl.Load on successful load", t, func() {
			Convey("should emit a plugin event message", func() {
				Convey("with loaded plugin name is mock", func() {
					So(lpe.plugin.LoadedPluginName, ShouldEqual, "mock")
				})
				Convey("with loaded plugin version as 2", func() {
					So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 2)
				})
				Convey("with loaded plugin type as collector", func() {
					So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
				})
			})
		})

		// Stop our controller so the plugins are unloaded and cleaned up from the system
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	} else {
		fmt.Printf("SNAP_PATH not set. Cannot test %s plugin.\n", fixtures.PluginName)
	}
}

func TestLoadWithSignedPlugins(t *testing.T) {
	if fixtures.SnapPath != "" {
		Convey("pluginControl.Load should successufully load a signed plugin with trust enabled", t, func() {
			c := New(getTestConfig())
			c.pluginTrust = PluginTrustEnabled
			c.signingManager = &mocksigningManager{signed: true}
			lpe := newListenToPluginEvent()
			c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, fixtures.PluginPath, "mock.asc")
			Convey("so error on loading a signed plugin should be nil", func() {
				So(err, ShouldBeNil)
			})
			// Stop our controller to clean up our plugin
			c.Stop()
			time.Sleep(100 * time.Millisecond)
		})
		Convey("pluginControl.Load should successfully load unsigned plugin when trust level set to warning", t, func() {
			c := New(getTestConfig())
			c.pluginTrust = PluginTrustWarn
			c.signingManager = &mocksigningManager{signed: false}
			lpe := newListenToPluginEvent()
			c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, fixtures.PluginPath)
			Convey("so error on loading an unsigned plugin should be nil", func() {
				So(err, ShouldBeNil)
			})
			c.Stop()
			time.Sleep(100 * time.Millisecond)
		})
		Convey("pluginControl.Load returns error with trust enabled and signing not validated", t, func() {
			c := New(getTestConfig())
			c.pluginTrust = PluginTrustEnabled
			c.signingManager = &mocksigningManager{signed: false}
			c.Start()
			time.Sleep(100 * time.Millisecond)
			_, err := load(c, fixtures.PluginPath)
			Convey("so error should not be nil when loading an unsigned plugin with trust enabled", func() {
				So(err, ShouldNotBeNil)
			})
			c.Stop()
		})
	} else {
		fmt.Printf("SNAP_PATH not set. Cannot test %s plugin.\n", fixtures.PluginName)
	}
}

func TestUnload(t *testing.T) {
	// These tests only work if SNAP_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if fixtures.SnapPath != "" {
		c := New(getTestConfig())
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestUnload", lpe)
		c.Start()
		time.Sleep(100 * time.Millisecond)
		_, e := load(c, fixtures.PluginPath)
		Convey("Loading a plugin to test unload", t, func() {
			Convey("Should not error", func() {
				So(e, ShouldBeNil)
			})
		})
		if e != nil {
			t.FailNow()
		}
		<-lpe.done
		Convey("And our plugin catalog", t, func() {
			Convey("Should not be empty", func() {
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
				Convey("where unloaded plugin name is mock", func() {
					So(lpe.plugin.UnloadedPluginName, ShouldEqual, "mock")
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
		fmt.Printf("SNAP_PATH not set. Cannot test %s plugin.\n", fixtures.PluginName)
	}
}

func TestStop(t *testing.T) {
	Convey("pluginControl.Stop", t, func() {
		c := New(getTestConfig())
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

	c := New(getTestConfig())

	// We need our own plugin manager to drop mock
	// loaded plugins into.  Arbitrarily adding
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

func (m *mc) Fetch(ns core.Namespace) ([]*metricType, error) {
	if m.e == 2 {
		return nil, serror.New(errors.New("test"))
	}
	return nil, nil
}

func (m *mc) resolvePlugin(mns []string, ver int) (*loadedPlugin, error) {
	return nil, nil
}

func (m *mc) GetPlugin(core.Namespace, int) (*loadedPlugin, error) {
	return nil, nil
}

func (m *mc) GetVersions(core.Namespace) ([]*metricType, error) {
	return nil, nil
}

func (m *mc) Get(ns core.Namespace, ver int) (*metricType, error) {
	if m.e == 1 {
		return &metricType{
			policy: &mockCDProc{},
		}, nil
	}
	return nil, serror.New(errorMetricNotFound(ns.String()))
}

func (m *mc) Subscribe(ns []string, ver int) error {
	if ns[0] == "nf" {
		return serror.New(errorMetricNotFound("/" + strings.Join(ns, "/")))
	}
	return nil
}

func (m *mc) Unsubscribe(ns []string, ver int) error {
	if ns[0] == "nf" {
		return serror.New(errorMetricNotFound("/" + strings.Join(ns, "/")))
	}
	if ns[0] == "neg" {
		return errNegativeSubCount
	}
	return nil
}

func (m *mc) Add(*metricType)                 {}
func (m *mc) Table() map[string][]*metricType { return map[string][]*metricType{} }
func (m *mc) Item() (string, []*metricType)   { return "", []*metricType{} }
func (m *mc) Keys() []string                  { return []string{} }

func (m *mc) Next() bool {
	m.e = 1
	return false
}

func (m *mc) AddLoadedMetricType(*loadedPlugin, core.Metric) error {
	return nil

}

func (m *mc) RmUnloadedPluginMetrics(lp *loadedPlugin) {

}

func (m *mc) GetQueriedNamespaces(ns core.Namespace) ([]core.Namespace, error) {
	return []core.Namespace{ns}, nil
}

func (m *mc) UpdateQueriedNamespaces(ns core.Namespace) {
}

func (m *mc) MatchNamespaces(ns core.Namespace) ([]core.Namespace, error) {
	return []core.Namespace{ns}, nil
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
		c := New(getTestConfig())
		lp := &loadedPlugin{}
		mt := newMetricType(core.NewNamespace("foo", "bar"), time.Now(), lp)
		c.metricCatalog.Add(mt)
		Convey("it returns a collection of core.MetricTypes", func() {
			t, err := c.MetricCatalog()
			So(err, ShouldBeNil)
			So(len(t), ShouldEqual, 1)
			So(t[0].Namespace(), ShouldResemble, core.NewNamespace("foo", "bar"))
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
		c := New(getTestConfig())
		c.metricCatalog = &mc{}
		So(c.MetricExists(core.NewNamespace("hi"), -1), ShouldEqual, false)
	})
}

func TestMetricConfig(t *testing.T) {
	Convey("required config provided by task", t, func() {
		c := New(getTestConfig())
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, err := load(c, fixtures.JSONRPCPluginPath)
		So(err, ShouldBeNil)
		<-lpe.done
		cd := cdata.NewNode()
		m1 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
		}

		Convey("So metric should not be valid without config", func() {
			errs := c.subscriptionGroups.validateMetric(m1)
			So(errs, ShouldNotBeNil)
		})
		cd.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})

		Convey("So metric should be valid with config", func() {
			m1.Cfg = cdata.NewNode()
			m1.Cfg.AddItem("password", ctypes.ConfigValueStr{Value: "password"})
			errs := c.subscriptionGroups.validateMetric(m1)
			So(errs, ShouldBeNil)
		})

		Convey("So metric should not be valid if does not occur in the catalog", func() {
			m := fixtures.MockMetricType{
				Namespace_: core.NewNamespace("intel", "mock", "bad"),
			}
			errs := c.subscriptionGroups.validateMetric(m)
			So(errs, ShouldNotBeNil)
		})

		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})

	Convey("config provided by defaults", t, func() {
		c := New(getTestConfig())
		c.Config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "pwd"})
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, err := load(c, fixtures.JSONRPCPluginPath)
		So(err, ShouldBeNil)
		<-lpe.done
		m1 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
		}

		Convey("So metric should be valid with config", func() {
			errs := c.subscriptionGroups.validateMetric(m1)
			So(errs, ShouldBeNil)
		})
		Convey("So mock should have name: bob config from defaults", func() {
			So(c.Config.Plugins.pluginCache["0:mock:1"].Table()["name"], ShouldResemble, ctypes.ConfigValueStr{Value: "bob"})
		})

		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})

	Convey("nil config provided by task", t, func() {
		config := getTestConfig()
		c := New(config)
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, err := load(c, fixtures.JSONRPCPluginPath)
		So(err, ShouldBeNil)
		<-lpe.done

		cfg := cdata.NewNode()
		cfg.AddItem("password", ctypes.ConfigValueStr{Value: "password"})
		m1 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
			Cfg:        cfg,
		}

		Convey("So metric should be valid with config", func() {
			errs := c.subscriptionGroups.validateMetric(m1)
			So(errs, ShouldBeNil)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
	Convey("required config provided by global plugin config", t, func() {
		config := getTestConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(config)
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, err := load(c, fixtures.JSONRPCPluginPath)
		So(err, ShouldBeNil)
		<-lpe.done
		m1 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
			Ver:        1,
		}
		errs := c.subscriptionGroups.validateMetric(m1)
		Convey("So metric should be valid with config", func() {
			So(errs, ShouldBeNil)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestRoutingCachingStrategy(t *testing.T) {
	Convey("Given loaded plugins that use sticky routing", t, func() {
		config := getTestConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(config)
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, e := load(c, fixtures.PluginPath)
		So(e, ShouldBeNil)
		if e != nil {
			t.FailNow()
		}
		metric := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
			Ver:        2,
			Cfg:        cdata.NewNode(),
		}
		<-lpe.done

		cdt := cdata.NewTree()
		node := cdata.NewNode()
		node.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		cdt.Add([]string{"intel", "mock"}, node)

		Convey("Start the plugins", func() {
			lp, err := c.pluginManager.get("collector:mock:2")
			So(err, ShouldBeNil)
			So(lp, ShouldNotBeNil)
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			So(pool, ShouldNotBeNil)
			tasks := []string{
				uuid.New(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
			}
			for _, id := range tasks {
				pool.Subscribe(id)
				err = c.pluginRunner.runPlugin(lp.Details)
				So(err, ShouldBeNil)
				serr := c.subscriptionGroups.Add(id, []core.RequestedMetric{metric}, cdt, []core.SubscribedPlugin{})
				So(serr, ShouldBeNil)
			}
			// The cache ttl should be 500ms. The system default is 500ms, but the plugin exposed 100ms which is less than the system default.
			ttl, err := pool.CacheTTL(tasks[0])
			So(err, ShouldBeNil)
			So(ttl, ShouldResemble, 500*time.Millisecond)
			So(pool.Strategy(), ShouldHaveSameTypeAs, strategy.NewSticky(ttl))
			So(pool.Count(), ShouldEqual, len(tasks))
			So(pool.SubscriptionCount(), ShouldEqual, len(tasks))
			Convey("Collect metrics", func() {
				taskID := tasks[rand.Intn(len(tasks))]
				for i := 0; i < 10; i++ {
					_, errs := c.CollectMetrics(taskID, nil)
					So(errs, ShouldBeEmpty)
				}
				Convey("Check cache stats", func() {
					So(pool.AllCacheHits(), ShouldEqual, 9)
					So(pool.AllCacheMisses(), ShouldEqual, 1)
				})
			})
		})
		c.Stop()
	})

	Convey("Given loaded plugins that use least-recently-used routing", t, func() {
		c := New(getTestConfig())
		c.Start()
		c.Config.Plugins.Collector.Plugins["mock"] = newPluginConfigItem(
			optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}),
			optAddPluginConfigItem("user", ctypes.ConfigValueStr{Value: "jane"}),
			optAddPluginConfigItem("password", ctypes.ConfigValueStr{Value: "doe"}),
		)
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, e := load(c, fixtures.JSONRPCPluginPath)
		So(e, ShouldBeNil)
		if e != nil {
			t.FailNow()
		}
		metric, err := c.metricCatalog.Get(core.NewNamespace("intel", "mock", "foo"), 1)
		metric.config = cdata.NewNode()
		So(err, ShouldBeNil)
		So(metric.Namespace().String(), ShouldResemble, "/intel/mock/foo")
		So(err, ShouldBeNil)
		<-lpe.done

		cdt := cdata.NewTree()
		node := cdata.NewNode()
		node.AddItem("user", ctypes.ConfigValueStr{Value: "jane"})
		node.AddItem("test", ctypes.ConfigValueBool{Value: true})
		node.AddItem("password", ctypes.ConfigValueStr{Value: "doe"})
		cdt.Add([]string{"intel", "mock"}, node)

		Convey("Start the plugins", func() {
			lp, err := c.pluginManager.get("collector:mock:1")
			So(err, ShouldBeNil)
			So(lp, ShouldNotBeNil)
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			time.Sleep(1 * time.Second)
			So(errp, ShouldBeNil)
			So(pool, ShouldNotBeNil)
			tasks := []string{
				uuid.New(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
				uuid.New(),
			}
			for _, id := range tasks {
				pool.Subscribe(id)
				err = c.pluginRunner.runPlugin(lp.Details)
				So(err, ShouldBeNil)
				serrs := c.subscriptionGroups.Add(id, []core.RequestedMetric{metric}, cdt, []core.SubscribedPlugin{})
				So(serrs, ShouldBeNil)
			}
			// The cache ttl should be 100ms which is what the plugin exposed (no system default was provided)
			ttl, err := pool.CacheTTL(tasks[0])
			So(err, ShouldBeNil)
			So(ttl, ShouldResemble, 1100*time.Millisecond)
			So(pool.Strategy(), ShouldHaveSameTypeAs, strategy.NewLRU(ttl))
			So(pool.Count(), ShouldEqual, len(tasks))
			So(pool.SubscriptionCount(), ShouldEqual, len(tasks))
			Convey("Collect metrics", func() {
				taskID := tasks[rand.Intn(len(tasks))]
				for i := 0; i < 10; i++ {
					cr, errs := c.CollectMetrics(taskID, nil)
					So(errs, ShouldBeEmpty)
					for i := range cr {
						So(cr[i].Data(), ShouldContainSubstring, "The mock collected data!")
						So(cr[i].Data(), ShouldContainSubstring, "test=true")
					}
				}
				Convey("Check cache stats", func() {
					So(pool.AllCacheHits(), ShouldEqual, 9)
					So(pool.AllCacheMisses(), ShouldEqual, 1)
				})
			})
		})
		c.Stop()
	})
}

func TestCollectDynamicMetrics(t *testing.T) {
	Convey("given a plugin using the native client", t, func() {
		config := getTestConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		config.CacheExpiration = jsonutil.Duration{time.Second * 1}
		c := New(config)
		c.Start()
		So(strategy.GlobalCacheExpiration, ShouldResemble, time.Second*1)
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
		_, e := load(c, fixtures.PluginPath)
		Convey("Loading native client plugin", func() {
			Convey("Should not error", func() {
				So(e, ShouldBeNil)
			})
		})
		if e != nil {
			t.FailNow()
		}
		<-lpe.done
		_, e = load(c, fixtures.JSONRPCPluginPath)
		Convey("Loading JSONRPC client plugin", func() {
			Convey("Should not error", func() {
				So(e, ShouldBeNil)
			})
		})
		if e != nil {
			t.FailNow()
		}
		<-lpe.done
		metrics, err := c.metricCatalog.Fetch(core.NewNamespace())
		So(err, ShouldBeNil)
		So(len(metrics), ShouldEqual, 6)

		m, err := c.metricCatalog.Get(core.NewNamespace("intel", "mock", "*", "baz"), 2)
		So(err, ShouldBeNil)
		So(m, ShouldNotBeNil)

		errs := c.subscriptionGroups.validateMetric(m)
		So(errs, ShouldBeNil)
		cdt := cdata.NewTree()

		Convey("collects metrics from plugin using native client", func() {
			lp, err := c.pluginManager.get("collector:mock:2")
			So(err, ShouldBeNil)
			So(lp, ShouldNotBeNil)
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			So(pool, ShouldNotBeNil)
			taskID := uuid.New()

			ttl, err := pool.CacheTTL(taskID)
			So(err, ShouldResemble, strategy.ErrPoolEmpty)
			So(ttl, ShouldEqual, 0)
			So(pool.Count(), ShouldEqual, 0)
			So(pool.SubscriptionCount(), ShouldEqual, 0)

			serrs := c.SubscribeDeps(taskID, []core.RequestedMetric{m},
				[]core.SubscribedPlugin{subscribedPlugin{
					typeName: "collector",
					name:     "mock",
					version:  2,
					config:   cdata.NewNode(),
				}}, cdt)
			So(serrs, ShouldBeNil)

			ttl, err = pool.CacheTTL(taskID)
			So(err, ShouldBeNil)
			So(ttl, ShouldEqual, time.Second)
			So(pool.Count(), ShouldEqual, 1)
			So(pool.SubscriptionCount(), ShouldEqual, 1)

			// The minimum TTL advertised by the plugin is 100ms therefore the TTL for th			// pool should be the global cache expiration
			So(ttl, ShouldEqual, strategy.GlobalCacheExpiration)
			mts, errs := c.CollectMetrics(taskID, nil)
			hits, err := pool.CacheHits(m.namespace.String(), 2, taskID)
			So(err, ShouldBeNil)
			So(hits, ShouldEqual, 0)
			So(errs, ShouldBeNil)
			So(len(mts), ShouldEqual, 10)
			mts, errs = c.CollectMetrics(taskID, nil)
			hits, err = pool.CacheHits(m.namespace.String(), 2, taskID)
			So(err, ShouldBeNil)

			// todo resolve problem with caching for dynamic metrics
			// So(hits, ShouldEqual, 1)

			So(errs, ShouldBeNil)
			So(len(mts), ShouldEqual, 10)
			pool.Unsubscribe(taskID)
			pool.SelectAndKill(taskID, "unsubscription event")
			So(pool.Count(), ShouldEqual, 0)
			So(pool.SubscriptionCount(), ShouldEqual, 0)
		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestFailedPlugin(t *testing.T) {
	Convey("given a loaded plugin", t, func() {
		// Create controller
		c := New(getTestConfig())
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TEST", lpe)
		c.Config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})

		// Load plugin
		_, e := load(c, fixtures.PluginPath)
		So(e, ShouldBeNil)
		<-lpe.done
		_, err := c.MetricCatalog()
		So(err, ShouldBeNil)
		// metrics to collect
		cfg := cdata.NewNode()
		cfg.AddItem("panic", ctypes.ConfigValueBool{Value: true})
		mets := []core.Metric{
			fixtures.MockMetricType{
				Namespace_: core.NewNamespace("intel", "mock", "foo"),
				Cfg:        cfg,
			},
		}

		r := []core.RequestedMetric{}
		for _, m := range mets {
			r = append(r, m)
		}

		cps := []core.SubscribedPlugin{fixtures.NewMockPlugin(core.CollectorPluginType, "mock", 2)}
		cdt := cdata.NewTree()
		cdt.Add([]string{"intel", "mock"}, cfg)
		taskID := "taskID"
		serrs := c.SubscribeDeps(taskID, r, cps, cdt)
		So(serrs, ShouldBeNil)

		// retrieve loaded plugin
		lp, err := c.pluginManager.get("collector:mock:2")
		So(err, ShouldBeNil)
		So(lp, ShouldNotBeNil)

		Convey("create a pool, add subscriptions and start plugins", func() {
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			Convey("collect metrics against a plugin that will panic", func() {
				So(pool.Count(), ShouldEqual, 1)

				var err []error
				var cr []core.Metric
				eventMap := map[string]int{}
				for i := 0; i < MaxPluginRestartCount+1; i++ {
					cr, err = c.CollectMetrics(taskID, nil)
					So(err, ShouldNotBeNil)
					So(cr, ShouldBeNil)
					<-lpe.done

					if i < MaxPluginRestartCount {
						<-lpe.restarted
						eventMap[lpe.plugin.EventNamespace]++
						So(pool.RestartCount(), ShouldEqual, i+1)
						So(lpe.plugin.EventNamespace, ShouldEqual, control_event.AvailablePluginRestarted)
					}
				}
				<-lpe.max
				So(lpe.plugin.EventNamespace, ShouldEqual, control_event.PluginRestartsExceeded)
				So(eventMap[control_event.AvailablePluginRestarted], ShouldEqual, MaxPluginRestartCount)
				So(len(pool.Plugins()), ShouldEqual, 0)
				So(pool.RestartCount(), ShouldEqual, MaxPluginRestartCount)
			})
		})
		c.Stop()
	})
}

func TestCollectMetrics(t *testing.T) {
	Convey("given a loaded plugin", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		config := getTestConfig()
		config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})
		c := New(config)
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)

		// Add a global plugin config
		c.Config.Plugins.Collector.Plugins["mock"] = newPluginConfigItem(optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}))

		// Load plugin
		_, e := load(c, fixtures.JSONRPCPluginPath)
		So(e, ShouldBeNil)
		<-lpe.done
		mts, err := c.MetricCatalog()
		So(err, ShouldBeNil)
		So(len(mts), ShouldEqual, 4)

		cd := cdata.NewNode()
		cd.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})

		//m := []core.Metric{}
		m1 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
			Cfg:        cd,
		}
		m2 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "bar"),
			Cfg:        cd,
		}
		m3 := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "test"),
			Cfg:        cd,
		}

		// retrieve loaded plugin
		lp, err := c.pluginManager.get("collector:mock:1")
		So(err, ShouldBeNil)
		So(lp, ShouldNotBeNil)

		r := []core.RequestedMetric{}
		for _, m := range []fixtures.MockMetricType{m1, m2, m3} {
			r = append(r, m)
		}

		cdt := cdata.NewTree()
		cdt.Add([]string{"intel", "mock"}, cd)
		taskHit := "hitting"
		taskNonHit := "not-hitting"

		Convey("create a pool, add subscriptions and start plugins", func() {
			serrs := c.SubscribeDeps(taskHit, r, []core.SubscribedPlugin{subscribedPlugin{typeName: "collector", name: "mock", version: 1}}, cdt)
			So(serrs, ShouldBeNil)
			serrs = c.SubscribeDeps(taskNonHit, r, []core.SubscribedPlugin{subscribedPlugin{typeName: "collector", name: "mock", version: 1}}, cdt)
			So(serrs, ShouldBeNil)

			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)

			Convey("collect metrics", func() {
				for x := 0; x < 4; x++ {
					cr, err := c.CollectMetrics(taskHit, nil)
					So(err, ShouldBeNil)
					for i := range cr {
						So(cr[i].Data(), ShouldContainSubstring, "The mock collected data!")
						So(cr[i].Data(), ShouldContainSubstring, "test=true")
						So(cr[i].Data(), ShouldContainSubstring, "name={bob}")
						So(cr[i].Data(), ShouldContainSubstring, "password={testval}")
					}
				}
				ap := c.AvailablePlugins()
				So(ap, ShouldNotBeEmpty)
				So(pool.Strategy().String(), ShouldEqual, plugin.DefaultRouting.String())
				So(len(pool.Plugins()), ShouldEqual, 2)
				// when the first first plugin is hit the cache is populated the
				// cache satisfies the next 3 collect calls that come in within the
				// cache duration
				So(pool.Plugins()[1].HitCount(), ShouldEqual, 1)
				So(pool.Plugins()[2].HitCount(), ShouldEqual, 0)
				c.Stop()
			})
		})
	})
}

func TestPublishMetrics(t *testing.T) {
	Convey("Given an available file publisher plugin", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		c := New(getTestConfig())
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestPublishMetrics", lpe)
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		time.Sleep(1 * time.Second)

		// Load plugin
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-publisher-mock-file"))
		So(err, ShouldBeNil)
		<-lpe.done
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("publisher:mock-file:3")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "mock-file")
		So(lp.ConfigPolicy, ShouldNotBeNil)

		Convey("Subscribe to file publisher with good config", func() {
			n := cdata.NewNode()
			c.Config.Plugins.Publisher.Plugins[lp.Name()] = newPluginConfigItem(optAddPluginConfigItem("file", ctypes.ConfigValueStr{Value: "/tmp/snap-TestPublishMetrics.out"}))
			serrs := c.SubscribeDeps("1", []core.RequestedMetric{}, []core.SubscribedPlugin{subscribedPlugin{typeName: "publisher", name: "mock-file", version: 3}}, cdata.NewTree())
			So(serrs, ShouldBeNil)
			time.Sleep(2500 * time.Millisecond)

			Convey("Publish to file", func() {
				metrics := []plugin.MetricType{
					*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.SnapGOBContentType
				errs := c.PublishMetrics(contentType, buf.Bytes(), "mock-file", 3, n.Table(), "1")
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
		c := New(getTestConfig())
		lpe := newListenToPluginEvent()
		c.eventManager.RegisterHandler("TestProcessMetrics", lpe)
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		time.Sleep(1 * time.Second)
		c.Config.Plugins.Processor.Plugins["passthru"] = newPluginConfigItem(optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}))

		// Load plugin
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-processor-passthru"))
		So(err, ShouldBeNil)
		<-lpe.done
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("processor:passthru:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "passthru")
		So(lp.ConfigPolicy, ShouldNotBeNil)

		Convey("Subscribe to passthru processor with good config", func() {
			n := cdata.NewNode()
			serrs := c.SubscribeDeps("1", []core.RequestedMetric{}, []core.SubscribedPlugin{subscribedPlugin{typeName: "processor", name: "passthru", version: 1}}, cdata.NewTree())
			So(serrs, ShouldBeNil)
			time.Sleep(2500 * time.Millisecond)

			Convey("process metrics", func() {
				metrics := []plugin.MetricType{
					*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.SnapGOBContentType
				_, ct, errs := c.ProcessMetrics(contentType, buf.Bytes(), "passthru", 1, n.Table(), "1")
				So(errs, ShouldBeEmpty)
				mts := []plugin.MetricType{}
				dec := gob.NewDecoder(bytes.NewBuffer(ct))
				err := dec.Decode(&mts)
				So(err, ShouldBeNil)
				So(mts[0].Data_, ShouldEqual, 2)
				So(errs, ShouldBeNil)
			})
		})

		Convey("Count()", func() {
			pmt := &metricTypes{}
			count := pmt.Count()
			So(count, ShouldResemble, 0)

		})
		c.Stop()
		time.Sleep(100 * time.Millisecond)

	})
}

type listenToPluginEvents struct {
	plugin  *mockPluginEvent
	load    chan struct{}
	sub     chan struct{}
	unsub   chan struct{}
	started chan struct{}
}

func newListenToPluginEvents() *listenToPluginEvents {
	return &listenToPluginEvents{
		load:    make(chan struct{}),
		unsub:   make(chan struct{}),
		sub:     make(chan struct{}),
		started: make(chan struct{}),
		plugin:  &mockPluginEvent{},
	}
}

func (l *listenToPluginEvents) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.PluginSubscriptionEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.sub <- struct{}{}
	case *control_event.PluginUnsubscriptionEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.unsub <- struct{}{}
	case *control_event.LoadPluginEvent:
		l.plugin.EventNamespace = v.Namespace()
		l.load <- struct{}{}
	case *control_event.StartPluginEvent:
		l.started <- struct{}{}
	default:
		controlLogger.WithFields(log.Fields{
			"event:": v.Namespace(),
			"_block": "HandleGomit",
		}).Info("Unhandled Event")
	}
}

func TestMetricSubscriptionToNewVersion(t *testing.T) {
	Convey("Given a metric that is being collected at v1", t, func() {
		c := New(getTestConfig())
		lpe := newListenToPluginEvents()
		c.eventManager.RegisterHandler("TestMetricSubscriptionToNewVersion", lpe)
		c.Start()
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock1"))
		<-lpe.load
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("collector:mock:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "mock")
		//Subscribe deps to create pools.
		metric := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel", "mock", "foo"),
			Cfg:        cdata.NewNode(),
			Ver:        0,
		}
		So(metric.Version(), ShouldEqual, 0)
		ct := cdata.NewTree()
		n := cdata.NewNode()
		n.AddItem("pass", ctypes.ConfigValueBool{true})
		ct.Add([]string{""}, n)
		serr := c.SubscribeDeps("testTaskID", []core.RequestedMetric{metric}, []core.SubscribedPlugin{}, ct)
		<-lpe.sub // wait for subscription event
		<-lpe.started
		So(serr, ShouldBeNil)
		// collect metrics as a sanity check that everything is setup correctly
		mts, errs := c.CollectMetrics("testTaskID", nil)
		So(errs, ShouldBeNil)
		So(len(mts), ShouldEqual, 1)
		// ensure the data coming back is from v1. V1's data is type string
		_, ok := mts[0].Data().(string)
		So(ok, ShouldEqual, true)
		Convey("Loading v2 of that plugin should move subscriptions to newer version", func() {
			// Load version snap-plugin-collector-mock2
			_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock2"))
			So(err, ShouldBeNil)
			select {
			// Wait on subscriptionMovedEvent
			case <-lpe.sub:
			case <-time.After(3 * time.Second):
				fmt.Println("timeout waiting for subscription event")
				So(false, ShouldEqual, true)
			}

			pool1, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)
			So(pool1.SubscriptionCount(), ShouldEqual, 0)

			pool2, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			So(pool2.SubscriptionCount(), ShouldEqual, 1)

			mts, errs = c.CollectMetrics("testTaskID", nil)
			So(len(mts), ShouldEqual, 1)

			// ensure the data coming back is from v2, V2's data is type int
			_, ok = mts[0].Data().(int)
			So(ok, ShouldEqual, true)
		})
		c.Stop()
	})
}

func TestMetricSubscriptionToOlderVersion(t *testing.T) {
	Convey("Given a metric that is being collected at v2", t, func() {
		c := New(getTestConfig())
		lpe := newListenToPluginEvents()
		c.eventManager.RegisterHandler("TestMetricSubscriptionToOlderVersion", lpe)
		c.Start()
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock2"))
		<-lpe.load
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("collector:mock:2")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "mock")
		requestedMetric := fixtures.NewMockRequestedMetric(
			core.NewNamespace("intel", "mock", "bar"),
			0,
		)
		serr := c.SubscribeDeps("testTaskID", []core.RequestedMetric{requestedMetric}, []core.SubscribedPlugin{}, cdata.NewTree())
		<-lpe.sub // wait for subscription event
		<-lpe.started
		So(serr, ShouldBeNil)
		// collect metrics as a sanity check that everything is setup correctly
		mts, errs := c.CollectMetrics("testTaskID", nil)
		So(errs, ShouldBeNil)
		So(len(mts), ShouldEqual, 1)
		// ensure the data coming back is from v2. V2's data is type int
		_, ok := mts[0].Data().(int)
		So(ok, ShouldEqual, true)
		// grab plugin for mock v2
		pc := c.PluginCatalog()
		mockv2 := pc[0]
		Convey("Loading v1 of that plugin and unloading v2 should move subscriptions to older version", func() {
			// Load version snap-plugin-collector-mock2
			_, err = load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock1"))
			<-lpe.load
			So(err, ShouldBeNil)
			unloadedPlugin, err := c.Unload(mockv2)
			So(err, ShouldBeNil)
			So(unloadedPlugin, ShouldNotBeNil)
			<-lpe.unsub
			_, subscriptionErros, serr := c.subscriptionGroups.Get("testTaskID")
			So(subscriptionErros, ShouldBeNil)
			So(serr, ShouldBeNil)

			select {
			case <-lpe.sub:
			case <-time.After(3 * time.Second):
				fmt.Println("timeout waiting for subscription event")
				So(false, ShouldEqual, true)
			}

			// Check for subscription movement.
			// Give some time for subscription to be moved.
			var pool1 strategy.Pool
			var errp error
			ap := c.pluginRunner.AvailablePlugins()
			pool1, errp = ap.getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			So(pool1.SubscriptionCount(), ShouldEqual, 0)

			pool2, errp := ap.getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)
			So(pool2.SubscriptionCount(), ShouldEqual, 1)

			mts, errs := c.CollectMetrics("testTaskID", nil)
			So(errs, ShouldBeEmpty)
			So(len(mts), ShouldEqual, 1)

			// ensure the data coming back is from v1, V1's data is type string
			_, ok := mts[0].Data().(string)
			So(ok, ShouldEqual, true)
		})
		c.Stop()
	})

}

func TestDynamicMetricSubscriptionLoad(t *testing.T) {
	Convey("Given a dynamic metric that is being collected", t, func() {
		log.SetLevel(log.DebugLevel)
		c := New(getTestConfig())
		lpe := newListenToPluginEvents()
		c.eventManager.RegisterHandler("TestDynamicMetricSubscriptionLoad", lpe)
		c.Start()
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("collector:mock:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "mock")
		//Subscribe deps to create pools.
		metric := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel").AddDynamicElement("*", "dynamic request"),
		}
		ct := cdata.NewTree()
		n := cdata.NewNode()
		n.AddItem("pass", ctypes.ConfigValueBool{true})
		ct.Add([]string{""}, n)
		serr := c.SubscribeDeps("testTaskID", []core.RequestedMetric{metric}, []core.SubscribedPlugin{}, ct)
		<-lpe.load // wait for load event
		<-lpe.sub  // wait for subscription event
		So(serr, ShouldBeNil)
		// collect metrics as a sanity check that everything is setup correctly
		mts1, errs := c.CollectMetrics("testTaskID", nil)
		So(errs, ShouldBeNil)
		So(len(mts1), ShouldBeGreaterThan, 1)
		// ensure the data coming back is from v1. V1's data is type string
		_, ok := mts1[0].Data().(string)
		So(ok, ShouldEqual, true)
		Convey("Loading another plugin should add subscriptions", func() {
			// Load version snap-plugin-collector-anothermock
			_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-anothermock1"))
			So(err, ShouldBeNil)
			<-lpe.load // wait for load event
			<-lpe.sub  // wait for subscription event

			pool1, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)
			So(pool1.SubscriptionCount(), ShouldEqual, 1)

			pool2, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:anothermock:1")
			So(errp, ShouldBeNil)
			So(pool2.SubscriptionCount(), ShouldEqual, 1)

			mts2, errs := c.CollectMetrics("testTaskID", nil)
			So(errs, ShouldBeNil)
			So(len(mts2), ShouldBeGreaterThan, len(mts1))

		})
		c.Stop()
	})
}

func TestDynamicMetricSubscriptionUnload(t *testing.T) {
	Convey("Given a dynamic metric that is being collected", t, func() {
		c := New(getTestConfig())
		lpe := newListenToPluginEvents()
		c.eventManager.RegisterHandler("TestDynamicMetricSubscriptionUnload", lpe)
		c.Start()
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		_, err = load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-anothermock1"))
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 2)
		lpMock, err2 := c.pluginManager.get("collector:mock:1")
		So(err2, ShouldBeNil)
		So(lpMock.Name(), ShouldResemble, "mock")
		lpAMock, err3 := c.pluginManager.get("collector:anothermock:1")
		So(err3, ShouldBeNil)
		So(lpAMock.Name(), ShouldResemble, "anothermock")

		//Subscribe deps to create pools.
		metric := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel").AddDynamicElement("*", "dynamic request"),
		}
		ct := cdata.NewTree()
		n := cdata.NewNode()
		n.AddItem("pass", ctypes.ConfigValueBool{true})
		ct.Add([]string{""}, n)
		serr := c.SubscribeDeps("testTaskID", []core.RequestedMetric{metric}, []core.SubscribedPlugin{}, ct)
		<-lpe.sub
		subsCount := 0
	L:
		for {
			select {
			case <-lpe.load:
				subsCount += 1
			case <-time.After(6 * time.Second):
				fmt.Println("timeout waiting for subscription event")
				So(false, ShouldEqual, true)
			default:
				if subsCount == 2 {
					break L
				}
			}
		}
		So(serr, ShouldBeNil)
		// collect metrics as a sanity check that everything is setup correctly
		mts1, errs := c.CollectMetrics("testTaskID", nil)
		So(errs, ShouldBeNil)
		So(len(mts1), ShouldBeGreaterThan, 1)
		Convey("Unloading mock plugin should remove its subscriptions", func() {
			pool1, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)
			So(pool1.SubscriptionCount(), ShouldEqual, 1)
			_, err = c.Unload(lpMock)
			So(err, ShouldBeNil)
			<-lpe.unsub
			<-lpe.sub
			So(pool1.SubscriptionCount(), ShouldEqual, 0)
			pool2, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:anothermock:1")
			So(errp, ShouldBeNil)
			So(pool2.SubscriptionCount(), ShouldEqual, 1)
			mts2, errs := c.CollectMetrics("testTaskID", nil)
			So(errs, ShouldBeNil)
			So(len(mts2), ShouldBeLessThan, len(mts1))

			// ensure the data coming back is from another mock (values over 9000)
			val, ok := mts2[0].Data().(int)
			So(ok, ShouldEqual, true)
			So(val, ShouldBeGreaterThan, 9000)
		})
		c.Stop()
	})
}

func TestDynamicMetricSubscriptionLoadLessMetrics(t *testing.T) {
	Convey("Given a dynamic metric that is being collected", t, func() {
		log.SetLevel(log.DebugLevel)
		c := New(getTestConfig())
		testLessMetrics := cdata.NewNode()
		testLessMetrics.AddItem("test-less", ctypes.ConfigValueBool{true})
		c.Config.Plugins.Collector.All = testLessMetrics

		lpe := newListenToPluginEvents()
		c.eventManager.RegisterHandler("TestDynamicMetricSubscriptionLoadLessMetrics", lpe)
		c.Start()
		_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("collector:mock:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "mock")
		//Subscribe deps to create pools.
		metric := fixtures.MockMetricType{
			Namespace_: core.NewNamespace("intel").AddDynamicElement("*", "dynamic request"),
		}
		ct := cdata.NewTree()
		n := cdata.NewNode()
		n.AddItem("pass", ctypes.ConfigValueBool{true})
		n.AddItem("test-less", ctypes.ConfigValueBool{true})
		ct.Add([]string{""}, n)
		serr := c.SubscribeDeps("testTaskID", []core.RequestedMetric{metric}, []core.SubscribedPlugin{}, ct)
		<-lpe.load // wait for load event
		<-lpe.sub  // wait for subscription event
		So(serr, ShouldBeNil)
		lpMock, err2 := c.pluginManager.get("collector:mock:1")
		So(err2, ShouldBeNil)
		So(lpMock.Name(), ShouldResemble, "mock")
		// collect metrics as a sanity check that everything is setup correctly
		mts1, errs := c.CollectMetrics("testTaskID", nil)
		So(errs, ShouldBeNil)
		So(len(mts1), ShouldBeGreaterThan, 1)
		Convey("metrics are collected from mock1", func() {
			for _, m := range mts1 {
				if strings.Contains(m.Namespace().String(), "host") {
					val, ok := m.Data().(int64)
					So(ok, ShouldEqual, true)
					So(val, ShouldBeLessThan, 100)
				} else {
					_, ok := m.Data().(string)
					So(ok, ShouldEqual, true)
				}
			}
		})
		// ensure the data coming back is from v1. V1's data is type string
		_, ok := mts1[0].Data().(string)
		So(ok, ShouldEqual, true)
		Convey("Loading higher plugin version with less metrics", func() {
			// Load version snap-plugin-collector-mock2
			_, err := load(c, path.Join(fixtures.SnapPath, "plugin", "snap-plugin-collector-mock2"))
			So(err, ShouldBeNil)
			<-lpe.load // wait for load event
			<-lpe.sub  // wait for subscription event

			pool1, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
			So(errp, ShouldBeNil)
			So(pool1.SubscriptionCount(), ShouldEqual, 1)

			pool2, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
			So(errp, ShouldBeNil)
			So(pool2.SubscriptionCount(), ShouldEqual, 1)

			mts2, errs := c.CollectMetrics("testTaskID", nil)
			So(errs, ShouldBeNil)
			So(len(mts2), ShouldEqual, len(mts1))

			Convey("metrics are collected from mock1 and mock2", func() {
				// ensure the data coming back is from mock 1 and mock 2
				for _, m := range mts2 {
					if strings.Contains(m.Namespace().String(), "host") || strings.Contains(m.Namespace().String(), "bar") {
						val, ok := m.Data().(int)
						So(ok, ShouldEqual, true)
						So(val, ShouldBeGreaterThan, 1000)
					} else {
						_, ok := m.Data().(string)
						So(ok, ShouldEqual, true)
					}
				}
			})

			Convey("Unloading lower plugin version", func() {
				_, err = c.Unload(lpMock)
				So(err, ShouldBeNil)
				<-lpe.unsub

				pool1, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:1")
				So(errp, ShouldBeNil)
				So(pool1.SubscriptionCount(), ShouldEqual, 0)

				pool2, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:mock:2")
				So(errp, ShouldBeNil)
				So(pool2.SubscriptionCount(), ShouldEqual, 1)

				mts3, errs := c.CollectMetrics("testTaskID", nil)
				So(errs, ShouldBeNil)
				So(len(mts3), ShouldBeLessThan, len(mts2))

				// ensure the data coming back is from mock 2 (values over 1000)
				for _, m := range mts3 {
					val, ok := m.Data().(int)
					So(ok, ShouldEqual, true)
					So(val, ShouldBeGreaterThan, 1000)
				}
			})
		})
		c.Stop()
	})
}
