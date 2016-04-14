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
	"encoding/json"
	"testing"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/internal/common"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func mockPluginReply(name, typeName string, version int64) *rpc.PluginReply {
	cp, _ := json.Marshal(cpolicy.New())
	return &rpc.PluginReply{
		Name:            name,
		TypeName:        typeName,
		Version:         version,
		IsSigned:        false,
		Status:          "loaded",
		LoadedTimestamp: &common.Time{Sec: 0, Nsec: 0},
		ConfigPolicy:    cp,
	}
}

func mockAvailablePluginReply(name, typeName string, version int64) *rpc.AvailablePluginReply {
	return &rpc.AvailablePluginReply{
		Name:             name,
		TypeName:         typeName,
		Version:          version,
		IsSigned:         false,
		HitCount:         0,
		ID:               0,
		LastHitTimestamp: &common.Time{Sec: 0, Nsec: 0},
	}
}

type MockManagesMetrics struct{}

func (m MockManagesMetrics) MetricCatalog(context.Context, *common.Empty, ...grpc.CallOption) (*rpc.MetricCatalogReply, error) {
	return nil, nil
}
func (m MockManagesMetrics) FetchMetrics(context.Context, *rpc.FetchMetricsRequest, ...grpc.CallOption) (*rpc.MetricCatalogReply, error) {
	return nil, nil
}
func (m MockManagesMetrics) GetMetricVersions(context.Context, *rpc.GetMetricVersionsRequest, ...grpc.CallOption) (*rpc.MetricCatalogReply, error) {
	return nil, nil
}
func (m MockManagesMetrics) GetMetric(context.Context, *rpc.FetchMetricsRequest, ...grpc.CallOption) (*rpc.MetricReply, error) {
	return nil, nil
}
func (m MockManagesMetrics) Load(context.Context, *rpc.PluginRequest, ...grpc.CallOption) (*rpc.PluginReply, error) {
	return nil, nil
}
func (m MockManagesMetrics) Unload(context.Context, *rpc.UnloadPluginRequest, ...grpc.CallOption) (*rpc.PluginReply, error) {
	return nil, nil
}

func (m MockManagesMetrics) PluginCatalog(context.Context, *common.Empty, ...grpc.CallOption) (*rpc.PluginCatalogReply, error) {
	return &rpc.PluginCatalogReply{
		Plugins: []*rpc.PluginReply{
			mockPluginReply("foo", "collector", 1),
			mockPluginReply("bar", "publisher", 1),
			mockPluginReply("foo", "collector", 2),
			mockPluginReply("baz", "publisher", 1),
			mockPluginReply("foo", "processor", 1),
			mockPluginReply("foobar", "processor", 1),
		},
	}, nil
}
func (m MockManagesMetrics) AvailablePlugins(context.Context, *common.Empty, ...grpc.CallOption) (*rpc.AvailablePluginsReply, error) {
	return &rpc.AvailablePluginsReply{
		Plugins: []*rpc.AvailablePluginReply{
			mockAvailablePluginReply("foo", "collector", 1),
			mockAvailablePluginReply("bar", "publisher", 1),
			mockAvailablePluginReply("foo", "collector", 2),
			mockAvailablePluginReply("baz", "publisher", 1),
			mockAvailablePluginReply("foo", "processor", 1),
			mockAvailablePluginReply("foobar", "processor", 1),
		},
	}, nil
}

func (m MockManagesMetrics) GetPlugin(context.Context, *rpc.GetPluginRequest, ...grpc.CallOption) (*rpc.GetPluginReply, error) {
	return nil, nil
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
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
		})
		Convey("With details", func() {
			detail := true
			Convey("Get All plugins", func() {
				plName := ""
				plType := ""
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 6)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 1)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins, _ := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
		})
	})

}
