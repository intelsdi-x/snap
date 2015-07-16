package control

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/control_event"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/core/perror"

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
func (m *MockPluginManagerBadSwap) SetMetricCatalog(catalogsMetrics)  {}
func (m *MockPluginManagerBadSwap) SetEmitter(gomit.Emitter)          {}
func (m *MockPluginManagerBadSwap) GenerateArgs(string) plugin.Arg    { return plugin.Arg{} }

func (m *MockPluginManagerBadSwap) all() map[string]*loadedPlugin {
	return m.loadedPlugins.table
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
		Convey("SwapPlugin", t, func() {
			c := New()
			lpe := new(listenToPluginEvent)
			c.eventManager.RegisterHandler("Control.PluginsSwapped", lpe)
			c.Start()
			_, e := c.Load(PluginPath)

			So(e, ShouldBeNil)

			dummy2Path := strings.Replace(PluginPath, "pulse-collector-dummy1", "pulse-collector-dummy2", 1)
			pc := c.PluginCatalog()
			dummy := pc[0]

			Convey("successfully swaps plugins", func() {
				err := c.SwapPlugins(dummy2Path, dummy)
				time.Sleep(100)
				pc = c.PluginCatalog()
				So(err, ShouldBeNil)
				So(pc[0].Name(), ShouldEqual, "dummy2")
				So(lpe.plugin.LoadedPluginName, ShouldEqual, "dummy2")
				So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 2)
				So(lpe.plugin.UnloadedPluginName, ShouldEqual, "dummy1")
				So(lpe.plugin.UnloadedPluginVersion, ShouldEqual, 1)
				So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
			})

			Convey("does not unload & returns an error if it cannot load a plugin", func() {
				err := c.SwapPlugins("/fake/plugin/path", pc[0])
				So(err, ShouldNotBeNil)
				So(pc[0].Name(), ShouldEqual, "dummy1")
			})

			Convey("rollsback loaded plugin & returns an error if it cannot unload a plugin", func() {
				dummy := pc[0]
				c.SwapPlugins(dummy2Path, dummy)
				pc = c.PluginCatalog()

				err := c.SwapPlugins(PluginPath+"oops", dummy)
				So(err, ShouldNotBeNil)
				So(pc[0].Name(), ShouldEqual, "dummy2")
			})

			Convey("rollback failure returns error", func() {
				dummy := pc[0]
				pm := new(MockPluginManagerBadSwap)
				pm.ExistingPlugin = dummy
				c.pluginManager = pm

				err := c.SwapPlugins(dummy2Path, dummy)
				So(err, ShouldNotBeNil)
			})
		})
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
	plugin mockPluginEvent
}

func (l *listenToPluginEvent) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		l.plugin = mockPluginEvent{
			LoadedPluginName:    v.Name,
			LoadedPluginVersion: v.Version,
			PluginType:          v.Type,
		}
	case *control_event.UnloadPluginEvent:
		l.plugin = mockPluginEvent{
			UnloadedPluginName:    v.Name,
			UnloadedPluginVersion: v.Version,
			PluginType:            v.Type,
		}
	case *control_event.SwapPluginsEvent:
		l.plugin = mockPluginEvent{
			LoadedPluginName:      v.LoadedPluginName,
			LoadedPluginVersion:   v.LoadedPluginVersion,
			UnloadedPluginName:    v.UnloadedPluginName,
			UnloadedPluginVersion: v.UnloadedPluginVersion,
			PluginType:            v.PluginType,
		}
	default:
		fmt.Println("Got an event you're not handling")
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
				lpe := new(listenToPluginEvent)
				c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)
				c.Start()
				_, err := c.Load(PluginPath)
				time.Sleep(100)
				So(err, ShouldBeNil)
				So(c.pluginManager.all(), ShouldNotBeEmpty)
				So(lpe.plugin.LoadedPluginName, ShouldEqual, "dummy1")
				So(lpe.plugin.LoadedPluginVersion, ShouldEqual, 1)
				So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
			})

			Convey("returns error if not started", func() {
				c := New()
				_, err := c.Load(PluginPath)

				So(len(c.pluginManager.all()), ShouldEqual, 0)
				So(err, ShouldNotBeNil)
			})

			Convey("adds to pluginControl.pluginManager.LoadedPlugins on successful load", func() {
				c := New()
				c.Start()
				_, err := c.Load(PluginPath)

				So(err, ShouldBeNil)
				So(len(c.pluginManager.all()), ShouldBeGreaterThan, 0)
			})

			Convey("returns error from pluginManager.LoadPlugin()", func() {
				c := New()
				c.Start()
				_, err := c.Load(PluginPath + "foo")

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
				lpe := new(listenToPluginEvent)
				c.eventManager.RegisterHandler("Control.PluginUnloaded", lpe)
				c.Start()
				_, err := c.Load(PluginPath)

				So(c.pluginManager.all(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)

				pc := c.PluginCatalog()

				So(len(pc), ShouldEqual, 1)
				_, err2 := c.Unload(pc[0])
				time.Sleep(100)
				So(err2, ShouldBeNil)
				So(lpe.plugin.UnloadedPluginName, ShouldEqual, "dummy1")
				So(lpe.plugin.UnloadedPluginVersion, ShouldEqual, 1)
				So(lpe.plugin.PluginType, ShouldEqual, int(plugin.CollectorPluginType))
			})

			Convey("returns error on unload for unknown plugin(or already unloaded)", func() {
				c := New()
				c.Start()
				_, err := c.Load(PluginPath)

				So(c.pluginManager.all(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)

				pc := c.PluginCatalog()

				So(len(pc), ShouldEqual, 1)
				_, err2 := c.Unload(pc[0])
				So(err2, ShouldBeNil)
				_, err3 := c.Unload(pc[0])
				So(err3.Error(), ShouldResemble, "plugin not found")
			})
			Convey("Listen for PluginUnloaded event", func() {
				c := New()
				lpe := new(listenToPluginEvent)
				c.eventManager.RegisterHandler("Control.PluginUnloaded", lpe)
				c.Start()
				c.Load(PluginPath)
				pc := c.PluginCatalog()
				c.Unload(pc[0])
				time.Sleep(100)
				So(lpe.plugin.UnloadedPluginName, ShouldEqual, "dummy1")
			})

		})

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

func (m *mc) Fetch(ns []string) ([]*metricType, error) {
	if m.e == 2 {
		return nil, errors.New("test")
	}
	return nil, nil
}

func (m *mc) resolvePlugin(mns []string, ver int) (*loadedPlugin, error) {
	return nil, nil
}

func (m *mc) GetPlugin([]string, int) (*loadedPlugin, error) {
	return nil, nil
}

func (m *mc) Get(ns []string, ver int) (*metricType, error) {
	if m.e == 1 {
		return &metricType{
			policy: &mockCDProc{},
		}, nil
	}
	return nil, errMetricNotFound
}

func (m *mc) Subscribe(ns []string, ver int) error {
	if ns[0] == "nf" {
		return errMetricNotFound
	}
	return nil
}

func (m *mc) Unsubscribe(ns []string, ver int) error {
	if ns[0] == "nf" {
		return errMetricNotFound
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
			So(t[0].Namespace(), ShouldResemble, "/foo/bar")
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
		// c.metricCatalog.Next()
		// So(c.MetricExists([]string{"hi"}, -1), ShouldEqual, true)
	})
}

type MockMetricType struct {
	namespace []string
	cfg       *cdata.ConfigDataNode
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return 1
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return m.cfg
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func TestCollectMetrics(t *testing.T) {

	Convey("given a new router", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		// Load plugin
		c.Load(PluginPath)

		cd := cdata.NewNode()
		cd.AddItem("password", ctypes.ConfigValueStr{Value: "testval"})

		m := []core.Metric{}
		m1 := MockMetricType{
			namespace: []string{"intel", "dummy", "foo"},
			cfg:       cd,
		}
		m2 := MockMetricType{
			namespace: []string{"intel", "dummy", "bar"},
			cfg:       cd,
		}

		// retrieve loaded plugin
		lp, err := c.pluginManager.get("collector:dummy1:1")
		So(err, ShouldBeNil)
		So(lp, ShouldNotBeNil)
		pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("collector:dummy1:1")
		So(errp, ShouldBeNil)
		pool.subscribe(1, unboundSubscriptionType)
		err = c.sendPluginSubscriptionEvent(1, lp)
		So(err, ShouldBeNil)
		m = append(m, m1, m2)

		time.Sleep(time.Millisecond * 200)

		for x := 0; x < 5; x++ {
			cr, err := c.CollectMetrics(m, time.Now().Add(time.Second*60))
			So(err, ShouldBeNil)
			for i := range cr {
				So(cr[i].Data(), ShouldContainSubstring, "The dummy collected data!")
			}
			// fmt.Printf(" *  Collect Response: %+v\n", cr)
		}
		time.Sleep(time.Millisecond * 500)
		ap := c.AvailablePlugins()
		So(ap, ShouldNotBeEmpty)
	})

	Convey("Pool", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1
		// Create controller
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		c.Load(PluginPath)
		m := []core.Metric{}
		c.CollectMetrics(m, time.Now().Add(time.Second*60))
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
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()

		// Load plugin
		_, err := c.Load(path.Join(PulsePath, "plugin", "pulse-publisher-file"))
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("publisher:file:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "file")
		So(lp.ConfigPolicyTree, ShouldNotBeNil)

		Convey("Subscribe to file publisher with good config", func() {
			n := cdata.NewNode()
			n.AddItem("file", ctypes.ConfigValueStr{Value: "/tmp/pulse-TestPublishMetrics.out"})
			p := mockPlugin{
				name:       "file",
				pluginType: core.PublisherPluginType,
				ver:        1,
				config:     n,
			}
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("publisher:file:1")
			So(errp, ShouldBeNil)
			pool.subscribe(1, unboundSubscriptionType)
			errs := c.sendPluginSubscriptionEvent(1, p)
			So(errs, ShouldBeNil)
			time.Sleep(1 * time.Second)

			Convey("Publish to file", func() {
				metrics := []plugin.PluginMetricType{
					*plugin.NewPluginMetricType([]string{"foo"}, 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.PulseGOBContentType
				errs := c.PublishMetrics(contentType, buf.Bytes(), "file", 1, n.Table())
				So(errs, ShouldBeNil)
				ap := c.AvailablePlugins()
				So(ap, ShouldNotBeEmpty)
			})
		})

	})
}
func TestProcessMetrics(t *testing.T) {
	Convey("Given an available file processor plugin", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDurationDefault = time.Second * 1

		// Create controller
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()

		// Load plugin
		_, err := c.Load(path.Join(PulsePath, "plugin", "pulse-processor-passthru"))
		So(err, ShouldBeNil)
		So(len(c.pluginManager.all()), ShouldEqual, 1)
		lp, err2 := c.pluginManager.get("processor:passthru:1")
		So(err2, ShouldBeNil)
		So(lp.Name(), ShouldResemble, "passthru")
		So(lp.ConfigPolicyTree, ShouldNotBeNil)

		Convey("Subscribe to passthru processor with good config", func() {
			n := cdata.NewNode()
			p := mockPlugin{
				name:       "passthru",
				pluginType: core.ProcessorPluginType,
				ver:        1,
				config:     n,
			}
			pool, errp := c.pluginRunner.AvailablePlugins().getOrCreatePool("processor:passthru:1")
			So(errp, ShouldBeNil)
			pool.subscribe(1, unboundSubscriptionType)
			errs := c.sendPluginSubscriptionEvent(1, p)
			So(errs, ShouldBeNil)
			time.Sleep(1 * time.Second)

			Convey("process metrics", func() {
				metrics := []plugin.PluginMetricType{
					*plugin.NewPluginMetricType([]string{"foo"}, 1),
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				enc.Encode(metrics)
				contentType := plugin.PulseGOBContentType
				cnt, ct, errs := c.ProcessMetrics(contentType, buf.Bytes(), "passthru", 1, n.Table())
				fmt.Printf("%v %v", cnt, ct)
				So(errs, ShouldBeNil)
			})
		})

		Convey("Count()", func() {
			pmt := &pluginMetricTypes{}
			count := pmt.Count()
			So(count, ShouldResemble, 0)

		})

	})
}
