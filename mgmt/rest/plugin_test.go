// +build legacy

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

package rest

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	. "github.com/smartystreets/goconvey/convey"
)

type MockLoadedPlugin struct {
	MyName string
	MyType string
}

func (m MockLoadedPlugin) Name() string       { return m.MyName }
func (m MockLoadedPlugin) TypeName() string   { return m.MyType }
func (m MockLoadedPlugin) Version() int       { return 0 }
func (m MockLoadedPlugin) Plugin() string     { return "" }
func (m MockLoadedPlugin) IsSigned() bool     { return false }
func (m MockLoadedPlugin) Status() string     { return "" }
func (m MockLoadedPlugin) PluginPath() string { return "" }
func (m MockLoadedPlugin) LoadedTimestamp() *time.Time {
	now := time.Now()
	return &now
}
func (m MockLoadedPlugin) Policy() *cpolicy.ConfigPolicy { return nil }

// have my mock object also support AvailablePlugin
func (m MockLoadedPlugin) HitCount() int { return 0 }
func (m MockLoadedPlugin) LastHit() time.Time {
	return time.Now()
}
func (m MockLoadedPlugin) ID() uint32 { return 0 }

type MockManagesMetrics struct{}

func (m MockManagesMetrics) MetricCatalog() ([]core.CatalogedMetric, error) {
	return nil, nil
}
func (m MockManagesMetrics) FetchMetrics(core.Namespace, int) ([]core.CatalogedMetric, error) {
	return nil, nil
}
func (m MockManagesMetrics) GetMetricVersions(core.Namespace) ([]core.CatalogedMetric, error) {
	return nil, nil
}
func (m MockManagesMetrics) GetMetric(core.Namespace, int) (core.CatalogedMetric, error) {
	return nil, nil
}
func (m MockManagesMetrics) Load(*core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError) {
	return nil, nil
}
func (m MockManagesMetrics) Unload(core.Plugin) (core.CatalogedPlugin, serror.SnapError) {
	return nil, nil
}

func (m MockManagesMetrics) PluginCatalog() core.PluginCatalog {
	return []core.CatalogedPlugin{
		MockLoadedPlugin{MyName: "foo", MyType: "collector"},
		MockLoadedPlugin{MyName: "bar", MyType: "publisher"},
		MockLoadedPlugin{MyName: "foo", MyType: "collector"},
		MockLoadedPlugin{MyName: "baz", MyType: "publisher"},
		MockLoadedPlugin{MyName: "foo", MyType: "processor"},
		MockLoadedPlugin{MyName: "foobar", MyType: "processor"},
	}
}
func (m MockManagesMetrics) AvailablePlugins() []core.AvailablePlugin {
	return []core.AvailablePlugin{
		MockLoadedPlugin{MyName: "foo", MyType: "collector"},
		MockLoadedPlugin{MyName: "bar", MyType: "publisher"},
		MockLoadedPlugin{MyName: "foo", MyType: "collector"},
		MockLoadedPlugin{MyName: "baz", MyType: "publisher"},
		MockLoadedPlugin{MyName: "foo", MyType: "processor"},
		MockLoadedPlugin{MyName: "foobar", MyType: "processor"},
	}
}
func (m MockManagesMetrics) GetAutodiscoverPaths() []string {
	return nil
}

func TestGetPlugins(t *testing.T) {
	mm := MockManagesMetrics{}
	host := "localhost"
	Convey("Test getPlugns method", t, func() {
		Convey("Without details", func() {
			detail := false
			Convey("Get All plugins", func() {
				plName := ""
				plType := ""
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
		})
		Convey("With details", func() {
			detail := true
			Convey("Get All plugins", func() {
				plName := ""
				plType := ""
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 6)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 1)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
		})
	})

}
