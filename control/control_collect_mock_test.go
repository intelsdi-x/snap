package control

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/intelsdilabs/gomit"
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/routing"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"

	. "github.com/smartystreets/goconvey/convey"
)

// ------------------- mock plugin manager -------------
type MockPluginManager struct{}

func (_ *MockPluginManager) LoadPlugin(string, gomit.Emitter) (*loadedPlugin, error) { return nil, nil }
func (_ *MockPluginManager) UnloadPlugin(CatalogedPlugin) error                      { return nil }
func (_ *MockPluginManager) LoadedPlugins() *loadedPlugins                           { return nil }
func (_ *MockPluginManager) SetMetricCatalog(catalogsMetrics)                        {}
func (_ *MockPluginManager) GenerateArgs() plugin.Arg                                { return plugin.Arg{} }

// ------------------ mock catalogs  -----------------

type EmptyMockMetricCatalog struct{}

func (_ *EmptyMockMetricCatalog) Get([]string, int) (*metricType, error)             { return nil, nil }
func (_ *EmptyMockMetricCatalog) Add(*metricType)                                    {}
func (_ *EmptyMockMetricCatalog) AddLoadedMetricType(*loadedPlugin, core.MetricType) {}
func (_ *EmptyMockMetricCatalog) Item() (string, []*metricType)                      { return "", nil }
func (_ *EmptyMockMetricCatalog) Next() bool                                         { return false }
func (_ *EmptyMockMetricCatalog) Subscribe([]string, int) error                      { return nil }
func (_ *EmptyMockMetricCatalog) Unsubscribe([]string, int) error                    { return nil }
func (_ *EmptyMockMetricCatalog) Table() map[string][]*metricType                    { return nil }
func (_ *EmptyMockMetricCatalog) GetPlugin([]string, int) (*loadedPlugin, error)     { return nil, nil }

// catalog with just one foo loaded plugin
type FooMetricCatalog struct{ EmptyMockMetricCatalog }

func (_ *FooMetricCatalog) GetPlugin([]string, int) (*loadedPlugin, error) {
	return fooLoadedPlugin, nil
}

// catalog with 2 loaded plugins
type FooBarMetricCatalog struct{ EmptyMockMetricCatalog }

func (_ *FooBarMetricCatalog) GetPlugin(ns []string, v int) (*loadedPlugin, error) {
	if ns[0] == "foo" {
		return fooLoadedPlugin, nil
	}
	if ns[0] == "bar" {
		return barLoadedPlugin, nil
	}
	return nil, nil
}

type ReturnErrorMockMetricCatalog struct{ EmptyMockMetricCatalog }

func (_ *ReturnErrorMockMetricCatalog) GetPlugin([]string, int) (*loadedPlugin, error) {
	return nil, errors.New("some error")
}

// ----------------------- mock metric types ----------------------------

type FooMetricType struct{}

func (_ FooMetricType) Version() int                  { return 0 }
func (_ FooMetricType) Namespace() []string           { return []string{"foo"} }
func (_ FooMetricType) LastAdvertisedTime() time.Time { return time.Now() }
func (_ FooMetricType) Config() *cdata.ConfigDataNode { return cdata.NewNode() }

type BarMetricType struct{ FooMetricType }

func (_ BarMetricType) Namespace() []string { return []string{"bar"} }

// ----------------------- mock metric ----------------------------

type FooMetric struct{}

func (_ FooMetric) Namespace() []string { return []string{"foo"} }
func (_ FooMetric) Data() interface{}   { return 1 }

// ----------------------- Mock Collector Client ---------------

type MockPluginCollectorClient struct{}

func (_ MockPluginCollectorClient) Ping() error       { return nil }
func (_ MockPluginCollectorClient) Kill(string) error { return nil }
func (_ MockPluginCollectorClient) CollectMetrics([]core.MetricType) ([]core.Metric, error) {
	return nil, nil
}
func (_ MockPluginCollectorClient) GetMetricTypes() ([]core.MetricType, error) { return nil, nil }

// ---------------------- mock strategy ------------------------------

type FooMockStrategy struct{}

func (_ *FooMockStrategy) Select(routing.SelectablePluginPool, []routing.SelectablePlugin) (routing.SelectablePlugin, error) {
	return fooAvailablePlugin, nil
}

type ErrorMockStrategy struct{}

func (_ *ErrorMockStrategy) Select(routing.SelectablePluginPool, []routing.SelectablePlugin) (routing.SelectablePlugin, error) {
	return nil, errors.New("some error")
}

var _ RoutingStrategy = (*FooMockStrategy)(nil) // check interface implementation

// ------------------ bunch of entities to just test CollectMetrics -----------

var (
	// ----------- mock instances
	fooMetricType       = FooMetricType{}
	barMetricType       = BarMetricType{}
	fooMetricTypes      = []core.MetricType{fooMetricType}
	foobarMetricTypes   = []core.MetricType{fooMetricType, barMetricType}
	manager             = &MockPluginManager{}
	emptyCatalog        = &EmptyMockMetricCatalog{}
	fooCatalog          = &FooMetricCatalog{}
	foobarCatalog       = &FooBarMetricCatalog{}
	returnErrorsCatalog = &ReturnErrorMockMetricCatalog{}
	emptyClient         = MockPluginCollectorClient{}

	// --------- loaded plugins

	fooLoadedPlugin = &loadedPlugin{
		Meta: plugin.PluginMeta{
			Name:    "foo",
			Version: 0,
			Type:    plugin.CollectorPluginType,
		},
	}
	barLoadedPlugin = &loadedPlugin{
		Meta: plugin.PluginMeta{
			Name:    "bar",
			Version: 0,
			Type:    plugin.CollectorPluginType,
		},
	}

	// ------- available plugin & plugins

	fooAvailablePlugin = &availablePlugin{
		Name:   "foo",
		Key:    "foo:0",
		Client: emptyClient,
	}

	// availaablePlugin without pools
	noPoolAvailablePlugins = availablePlugins{
		Collectors: &apCollection{
			table: &map[string]*availablePluginPool{},
			mutex: &sync.Mutex{},
		},
	}

	// availaablePlugin with pool but without availablePlugins
	emptyPoolAvailablePlugins = availablePlugins{
		Collectors: &apCollection{
			table: &map[string]*availablePluginPool{
				"foo:0": &availablePluginPool{
					Plugins: &[]*availablePlugin{},
				},
			},
			mutex: &sync.Mutex{},
		},
	}

	fooAvailablePluginPool = &availablePluginPool{
		Plugins: &[]*availablePlugin{
			fooAvailablePlugin,
		},
		mutex: &sync.Mutex{},
	}

	// availablePlugins with pool and one availablePlugin
	fooAvailablePlugins = availablePlugins{
		Collectors: &apCollection{
			table: &map[string]*availablePluginPool{
				"foo:0": fooAvailablePluginPool,
			},
			mutex: &sync.Mutex{},
		},
	}

	// ---- strategies
	fooMockStrategy   = &FooMockStrategy{}
	errorMockStrategy = &ErrorMockStrategy{}
)

func TestGroupMetricTypesByPlugin(t *testing.T) {
	Convey("group metrics types", t, func() {

		Convey("returns errors from catalog", func() {
			pmts, err := groupMetricTypesByPlugin(returnErrorsCatalog, fooMetricTypes)
			So(err, ShouldNotBeNil)
			So(pmts, ShouldBeNil)
		})

		Convey("when loaded plugin not found return err", func() {
			pmts, err := groupMetricTypesByPlugin(emptyCatalog, fooMetricTypes)
			So(err, ShouldNotBeNil)
			So(pmts, ShouldBeNil)
		})

		Convey("with foo1 as loaded plugin", func() {
			pmts, err := groupMetricTypesByPlugin(fooCatalog, fooMetricTypes)
			So(err, ShouldBeNil)
			So(pmts, ShouldNotBeNil)
			So(len(pmts), ShouldEqual, 1)

			pmt := pmts[fooLoadedPlugin.Key()]
			So(pmt, ShouldNotBeNil)
			So(pmt.plugin, ShouldEqual, fooLoadedPlugin)
			So(len(pmt.metricTypes), ShouldEqual, 1)
			So(pmt.metricTypes[0], ShouldResemble, fooMetricType)

		})
		Convey("can group two loaded plugins for two metric types", func() {
			pmts, err := groupMetricTypesByPlugin(foobarCatalog, foobarMetricTypes)
			So(err, ShouldBeNil)
			So(pmts, ShouldNotBeNil)
			So(len(pmts), ShouldEqual, 2)

			pmt := pmts[fooLoadedPlugin.Key()]
			So(pmt, ShouldNotBeNil)
			So(pmt.plugin, ShouldEqual, fooLoadedPlugin)
			So(len(pmt.metricTypes), ShouldEqual, 1)
			So(pmt.metricTypes[0], ShouldResemble, fooMetricType)

			pmt = pmts[barLoadedPlugin.Key()]
			So(pmt, ShouldNotBeNil)
			So(pmt.plugin, ShouldEqual, barLoadedPlugin)
			So(len(pmt.metricTypes), ShouldEqual, 1)
			So(pmt.metricTypes[0], ShouldResemble, barMetricType)
		})
	})
}

func TestGetPool(t *testing.T) {
	Convey("get pool", t, func() {

		Convey("no pool found gives error", func() {
			_, err := getPool("foo:0", &noPoolAvailablePlugins)
			So(err, ShouldNotBeNil)
		})

		Convey("empty pool gives error", func() {
			_, err := getPool("foo:0", &emptyPoolAvailablePlugins)
			So(err, ShouldNotBeNil)
		})

		Convey("pool with ap return pool", func() {
			pool, err := getPool("foo:0", &fooAvailablePlugins)
			So(err, ShouldBeNil)
			So(pool, ShouldHaveSameTypeAs, &availablePluginPool{})
		})

	})
}

func TestGetAvailablePlugin(t *testing.T) {

	Convey("get available plugin", t, func() {
		Convey("returns error", func() {
			ap, err := getAvailablePlugin(fooAvailablePluginPool, errorMockStrategy)
			So(err, ShouldNotBeNil)
			So(ap, ShouldBeNil)
		})
		Convey("returns availplugin", func() {
			ap, err := getAvailablePlugin(fooAvailablePluginPool, fooMockStrategy)
			So(err, ShouldBeNil)
			So(ap, ShouldNotBeNil)
		})
	})
}

func TestCollectMetrics2(t *testing.T) {
	Convey("full scenario", t, func() {

		c := New()
		c.pluginManager = manager
		c.metricCatalog = emptyCatalog

		metrics, errs := c.CollectMetrics([]core.MetricType{FooMetricType{}}, nil, time.Now())
		So(metrics, ShouldNotBeEmpty)
		So(errs, ShouldBeEmpty)
	})
}
