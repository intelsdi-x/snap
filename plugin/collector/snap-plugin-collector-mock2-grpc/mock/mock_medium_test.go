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
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCollectMetric(t *testing.T) {
	ns0 := plugin.NewNamespace("intel", "mock", "test")
	ns1 := plugin.NewNamespace("intel", "mock", "foo")
	ns2 := plugin.NewNamespace("intel", "mock", "bar")
	ns3 := plugin.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")

	Convey("Testing CollectMetric", t, func() {

		newPlg := new(Mock)
		So(newPlg, ShouldNotBeNil)

		Convey("with 'test' config variable'", func() {

			cfg := plugin.Config{"test": true}

			Convey("testing specific metrics", func() {
				mTypes := []plugin.Metric{
					plugin.Metric{Namespace: ns0, Config: cfg},
					plugin.Metric{Namespace: ns1, Config: cfg},
					plugin.Metric{Namespace: ns2, Config: cfg},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				Convey("returned metrics should have data type integer", func() {
					for _, mt := range mts {
						_, ok := mt.Data.(int)
						So(ok, ShouldBeTrue)
					}
				})
			})

			Convey("testing dynamic metric", func() {

				mt := plugin.Metric{Namespace: ns3, Config: cfg}
				isDynamic, _ := mt.Namespace.IsDynamic()
				So(isDynamic, ShouldBeTrue)

				Convey("for none specified instance", func() {
					mts, _ := newPlg.CollectMetrics([]plugin.Metric{mt})

					// there is 10 available hosts (host0, host1, ..., host9)
					So(len(mts), ShouldEqual, 10)

					Convey("returned metrics should have data type integer", func() {
						for _, mt := range mts {
							_, ok := mt.Data.(int)
							So(ok, ShouldBeTrue)
						}
					})

					Convey("returned metrics should remain dynamic", func() {
						for _, mt := range mts {
							isDynamic, _ := mt.Namespace.IsDynamic()
							So(isDynamic, ShouldBeTrue)
						}
					})

				})

				Convey("for specified instance which is available - host0", func() {
					mt.Namespace[2].Value = "host0"
					mts, _ := newPlg.CollectMetrics([]plugin.Metric{mt})

					// only one metric for this specific hostname should be returned
					So(len(mts), ShouldEqual, 1)
					So(mts[0].Namespace.Strings(), ShouldResemble, []string{"intel", "mock", "host0", "baz"})

					Convey("returned metric should have data type integer", func() {
						_, ok := mts[0].Data.(int)
						So(ok, ShouldBeTrue)
					})

					Convey("returned metric should remain dynamic", func() {
						isDynamic, _ := mt.Namespace.IsDynamic()
						So(isDynamic, ShouldBeTrue)
					})

				})

				Convey("for specified instance which is not available - host10", func() {
					mt.Namespace[2].Value = "host10"
					mts, err := newPlg.CollectMetrics([]plugin.Metric{mt})

					So(mts, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "requested hostname `host10` is not available")

				})
			})

		})

		Convey("without config variables", func() {

			cfg := plugin.Config{}

			Convey("testing specific metrics", func() {
				mTypes := []plugin.Metric{
					plugin.Metric{Namespace: ns0, Config: cfg},
					plugin.Metric{Namespace: ns1, Config: cfg},
					plugin.Metric{Namespace: ns2, Config: cfg},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				Convey("returned metrics should have data type integer", func() {
					for _, mt := range mts {
						_, ok := mt.Data.(int)
						So(ok, ShouldBeTrue)
					}
				})
			})

			Convey("testing dynamic metics", func() {
				mTypes := []plugin.Metric{
					plugin.Metric{Namespace: ns3, Config: cfg},
				}
				mts, _ := newPlg.CollectMetrics(mTypes)

				Convey("returned metrics should have data type integer", func() {
					for _, mt := range mts {
						_, ok := mt.Data.(int)
						So(ok, ShouldBeTrue)
					}
				})

				Convey("returned metrics should remain dynamic", func() {
					for _, mt := range mts {
						isDynamic, _ := mt.Namespace.IsDynamic()
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
			cfg := plugin.Config{"test-fail": true}
			_, err := newPlg.GetMetricTypes(cfg)

			So(err, ShouldNotBeNil)
		})

		Convey("with 'test' config variable", func() {
			cfg := plugin.Config{"test": true}

			mts, err := newPlg.GetMetricTypes(cfg)

			So(err, ShouldBeNil)
			So(len(mts), ShouldEqual, 6)

			Convey("checking namespaces", func() {
				nsRes := getNsResultHavingConfig()
				for i, m := range mts {
					Convey("checking namespaces: "+strings.Join(m.Namespace.Strings(), "%"), func() {
						So(m.Namespace.Strings(), ShouldResemble, nsRes[i])
					})
				}
			})
		})

		Convey("without config variables", func() {
			mts, err := newPlg.GetMetricTypes(plugin.Config{})

			So(err, ShouldBeNil)
			So(len(mts), ShouldEqual, 5)

			Convey("checking namespaces", func() {
				nsRes := getNsResultWithoutConfig()
				for i, m := range mts {
					Convey("checking namespaces: "+strings.Join(m.Namespace.Strings(), "%"), func() {
						So(m.Namespace.Strings(), ShouldResemble, nsRes[i])
					})
				}
			})
		})
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

func getNsResultHavingConfig() [][]string {
	nss := [][]string{
		{"intel", "mock", "test%>"},
		{"intel", "mock", "/foo=㊽"},
		{"intel", "mock", "/bar⽔"},
		{"intel", "mock", "*", "/baz⽔"},
		{"intel", "mock", "*", "baz㊽", "/bar⽔"},
		{"intel", "mock", "*", "baz㊽", "|barᵹÄ☍"},
	}
	return nss
}

func getNsResultWithoutConfig() [][]string {
	nss := [][]string{
		{"intel", "mock", "/foo=㊽"},
		{"intel", "mock", "/bar⽔"},
		{"intel", "mock", "*", "/baz⽔"},
		{"intel", "mock", "*", "baz㊽", "/bar⽔"},
		{"intel", "mock", "*", "baz㊽", "|barᵹÄ☍"},
	}
	return nss
}
