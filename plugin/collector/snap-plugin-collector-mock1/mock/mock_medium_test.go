// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package mock

import (
	"math/rand"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/str"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCollectMetric(t *testing.T) {
	ns0 := core.NewNamespace("intel", "mock", "test")
	ns1 := core.NewNamespace("intel", "mock", "foo")
	ns2 := core.NewNamespace("intel", "mock", "bar")
	ns3 := core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")

	Convey("Testing CollectMetric", t, func() {

		newPlg := new(Mock)
		So(newPlg, ShouldNotBeNil)

		Convey("with 'test' config variable'", func() {

			node := cdata.NewNode()
			node.AddItem("test", ctypes.ConfigValueBool{Value: true})
			cfg := plugin.ConfigType{ConfigDataNode: node}

			Convey("testing specific metrics", func() {
				mTypes := []plugin.MetricType{
					plugin.MetricType{Namespace_: ns0, Config_: cfg.ConfigDataNode},
					plugin.MetricType{Namespace_: ns1, Config_: cfg.ConfigDataNode},
					plugin.MetricType{Namespace_: ns2, Config_: cfg.ConfigDataNode},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				for _, mt := range mts {
					_, ok := mt.Data_.(string)
					So(ok, ShouldBeTrue)
				}
			})

			Convey("testing dynamic metric", func() {

				mt := plugin.MetricType{Namespace_: ns3, Config_: cfg.ConfigDataNode}

				Convey("for none specified instance", func() {
					mts, _ := newPlg.CollectMetrics([]plugin.MetricType{mt})

					// there is 10 available hosts (host0, host1, ..., host9)
					So(len(mts), ShouldEqual, 10)

					Convey("returned metrics should have data type integer", func() {
						for _, mt := range mts {
							_, ok := mt.Data_.(int)
							So(ok, ShouldBeTrue)
						}
					})

					Convey("returned metrics should remain dynamic", func() {
						for _, mt := range mts {
							isDynamic, _ := mt.Namespace().IsDynamic()
							So(isDynamic, ShouldBeTrue)
						}
					})

				})

				Convey("for specified instance which is available - host0", func() {
					mt.Namespace()[2].Value = "host0"
					mts, _ := newPlg.CollectMetrics([]plugin.MetricType{mt})

					// only one metric for this specific hostname should be returned
					So(len(mts), ShouldEqual, 1)
					So(mts[0].Namespace().String(), ShouldEqual, "/intel/mock/host0/baz")

					Convey("returned metric should have data type integer", func() {
						_, ok := mts[0].Data_.(int)
						So(ok, ShouldBeTrue)
					})

					Convey("returned metric should remain dynamic", func() {
						isDynamic, _ := mt.Namespace().IsDynamic()
						So(isDynamic, ShouldBeTrue)
					})

				})

				Convey("for specified instance which is not available - host10", func() {
					mt.Namespace()[2].Value = "host10"
					mts, err := newPlg.CollectMetrics([]plugin.MetricType{mt})

					So(mts, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "requested hostname `host10` is not available")

				})
			})

		})

		Convey("without config variables", func() {

			node := cdata.NewNode()
			cfg := plugin.ConfigType{ConfigDataNode: node}

			Convey("testing specific metrics", func() {
				mTypes := []plugin.MetricType{
					plugin.MetricType{Namespace_: ns0, Config_: cfg.ConfigDataNode},
					plugin.MetricType{Namespace_: ns1, Config_: cfg.ConfigDataNode},
					plugin.MetricType{Namespace_: ns2, Config_: cfg.ConfigDataNode},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				for _, mt := range mts {
					_, ok := mt.Data_.(string)
					So(ok, ShouldBeTrue)
				}
			})

			Convey("testing dynamic metics", func() {
				mTypes := []plugin.MetricType{
					plugin.MetricType{Namespace_: ns3, Config_: cfg.ConfigDataNode},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				Convey("returned metrics should have data type integer", func() {
					for _, mt := range mts {
						_, ok := mt.Data_.(int)
						So(ok, ShouldBeTrue)
					}
				})

				Convey("returned metrics should remain dynamic", func() {
					for _, mt := range mts {
						isDynamic, _ := mt.Namespace().IsDynamic()
						So(isDynamic, ShouldBeTrue)
					}
				})

			})

		})
	})
}

func TestGetMetricTypes(t *testing.T) {
	Convey("Tesing GetMetricTypes", t, func() {

		newPlg := new(Mock)
		So(newPlg, ShouldNotBeNil)

		Convey("with missing on-load plugin config entry", func() {
			node := cdata.NewNode()
			node.AddItem("test-fail", ctypes.ConfigValueStr{Value: ""})

			_, err := newPlg.GetMetricTypes(plugin.ConfigType{ConfigDataNode: node})

			So(err, ShouldNotBeNil)
		})

		Convey("with 'test' config variable", func() {
			node := cdata.NewNode()
			node.AddItem("test", ctypes.ConfigValueStr{Value: ""})

			mts, err := newPlg.GetMetricTypes(plugin.ConfigType{ConfigDataNode: node})

			So(err, ShouldBeNil)
			So(len(mts), ShouldEqual, 5)

			Convey("checking namespaces", func() {
				metricNames := []string{}
				for _, m := range mts {
					metricNames = append(metricNames, m.Namespace().String())
				}

				ns := core.NewNamespace("intel", "mock", "test")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)

				ns = core.NewNamespace("intel", "mock", "foo")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)

				ns = core.NewNamespace("intel", "mock", "bar")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)

				ns = core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)
			})
		})

		Convey("without config variables", func() {
			node := cdata.NewNode()
			mts, err := newPlg.GetMetricTypes(plugin.ConfigType{ConfigDataNode: node})

			So(err, ShouldBeNil)
			So(len(mts), ShouldEqual, 4)

			Convey("checking namespaces", func() {
				metricNames := []string{}
				for _, m := range mts {
					metricNames = append(metricNames, m.Namespace().String())
				}

				ns := core.NewNamespace("intel", "mock", "foo")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)

				ns = core.NewNamespace("intel", "mock", "bar")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)

				ns = core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")
				So(str.Contains(metricNames, ns.String()), ShouldBeTrue)
			})
		})

	})
}

func TestMeta(t *testing.T) {
	Convey("Testing Meta", t, func() {
		meta := Meta()
		So(meta.Name, ShouldEqual, Name)
		So(meta.Version, ShouldEqual, Version)
		So(meta.Type, ShouldEqual, Type)
		So(meta.AcceptedContentTypes[0], ShouldEqual, plugin.SnapGOBContentType)
		So(meta.ReturnedContentTypes[0], ShouldEqual, plugin.SnapGOBContentType)
		So(meta.Unsecure, ShouldEqual, true)
		So(meta.RoutingStrategy, ShouldEqual, plugin.DefaultRouting)
		So(meta.CacheTTL, ShouldEqual, 1100*time.Millisecond)
	})
}

func TestRandInt(t *testing.T) {
	Convey("Testing randInt", t, func() {
		rand.Seed(time.Now().UTC().UnixNano())
		data := randInt(65, 90)
		So(data, ShouldBeBetween, 64, 91)
	})
}

func TestGetConfigPolicy(t *testing.T) {
	Convey("Testing GetConfigPolicy", t, func() {
		newPlg := new(Mock)
		So(newPlg, ShouldNotBeNil)

		configPolicy, err := newPlg.GetConfigPolicy()

		So(err, ShouldBeNil)
		So(configPolicy, ShouldNotBeNil)
	})
}
