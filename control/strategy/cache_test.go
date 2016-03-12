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

package strategy

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/chrono"
)

func TestCache(t *testing.T) {
	GlobalCacheExpiration = time.Duration(300 * time.Millisecond)
	Convey("puts and gets a metric", t, func() {
		mc := NewCache(GlobalCacheExpiration)
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
			Version_:   1,
		}

		mc.put(genKeyFromMetric(foo), foo)
		ret := mc.get(genKeyFromMetric(foo))

		So(ret, ShouldNotBeNil)
		So(ret, ShouldEqual, foo)
	})
	Convey("returns nil if the cache cell does not exist", t, func() {
		mc := NewCache(GlobalCacheExpiration)
		ret := mc.get("/foo/bar:1")
		So(ret, ShouldBeNil)
	})
	Convey("returns nil if the cache cell has expired", t, func() {
		// Make sure global clock is restored after test.
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()

		// Use artificial time: pause to get base time.
		chrono.Chrono.Pause()

		mc := NewCache(400 * time.Millisecond)
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
			Version_:   1,
		}
		mc.put(genKeyFromMetric(foo), foo)
		chrono.Chrono.Forward(401 * time.Millisecond)

		ret := mc.get(genKeyFromMetric(foo))
		So(ret, ShouldBeNil)
	})
	Convey("hit and miss counts", t, func() {
		Convey("ticks hit count when a cache entry is hit", func() {
			mc := NewCache(400 * time.Millisecond)
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
				Version_:   1,
			}
			mc.put(genKeyFromMetric(foo), foo)
			mc.get(genKeyFromMetric(foo))
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks hit count when a cache entry is still a hit", func() {
			defer chrono.Chrono.Reset()
			defer chrono.Chrono.Continue()

			chrono.Chrono.Pause()

			mc := NewCache(400 * time.Millisecond)
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
				Version_:   1,
			}

			mc.put(genKeyFromMetric(foo), foo)
			chrono.Chrono.Forward(250 * time.Millisecond)
			mc.get(genKeyFromMetric(foo))
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is missed", func() {
			defer chrono.Chrono.Reset()
			defer chrono.Chrono.Continue()

			chrono.Chrono.Pause()

			mc := NewCache(GlobalCacheExpiration)
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
				Version_:   1,
			}
			mc.put(genKeyFromMetric(foo), foo)
			chrono.Chrono.Forward(301 * time.Millisecond)
			mc.get(genKeyFromMetric(foo))
			So(mc.table["/foo/bar:1"].misses, ShouldEqual, 1)
		})
	})

	Convey("Add and get metrics via updateCache and checkCache", t, func() {
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()

		chrono.Chrono.Pause()

		mc := NewCache(GlobalCacheExpiration)
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
			Version_:   1,
		}
		baz := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "baz"},
			Version_:   1,
		}
		metricList := []core.Metric{foo, baz}
		mc.updateCache(metricList)
		Convey("they should be retrievable via get", func() {
			ret := mc.get(genKeyFromMetric(foo))
			So(ret, ShouldEqual, foo)
			ret = mc.get(genKeyFromMetric(baz))
			So(ret, ShouldEqual, baz)
		})
		Convey("they should be retrievable via checkCache", func() {
			nonCached := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "fooer"},
				Version_:   1,
			}
			metricList = append(metricList, nonCached)
			toCollect, fromCache := mc.checkCache(metricList)
			Convey("Should return cached metrics", func() {
				So(len(fromCache), ShouldEqual, 2)
				So(fromCache[0], ShouldEqual, foo)
				So(fromCache[1], ShouldEqual, baz)
			})
			Convey("Should return non-cached metrics to be collect", func() {
				So(len(toCollect), ShouldEqual, 1)
				So(toCollect[0], ShouldEqual, nonCached)
			})
			Convey("cache should store individual", func() {
				Convey("hits", func() {

					hits, err := mc.cacheHits(foo)
					So(err, ShouldBeNil)
					So(hits, ShouldEqual, 1)
					hits, err = mc.cacheHits(nonCached)
					So(err, ShouldBeNil)
					So(hits, ShouldEqual, 0)
				})
				Convey("misses", func() {
					misses, err := mc.cacheMisses(foo)
					So(err, ShouldBeNil)
					So(misses, ShouldEqual, 0)
					misses, err = mc.cacheMisses(nonCached)
					So(err, ShouldBeNil)
					So(misses, ShouldEqual, 1)
				})

			})
			Convey("should error when using unknown metric for hits or misses", func() {
				unknown := &plugin.PluginMetricType{
					Namespace_: []string{"unknown"},
					Version_:   1,
				}
				_, err := mc.cacheHits(unknown)
				So(err, ShouldNotBeNil)
				_, err = mc.cacheMisses(unknown)
				So(err, ShouldNotBeNil)
			})
			Convey("cache should store total hits and misses", func() {
				So(mc.allCacheHits(), ShouldEqual, 2)
				So(mc.allCacheMisses(), ShouldEqual, 1)
			})
		})
	})

	Convey("Adding plugins with same namespace but different versions", t, func() {
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()
		chrono.Chrono.Pause()
		mc := NewCache(GlobalCacheExpiration)
		v1 := plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
			Version_:   1,
			Labels_:    []core.Label{{Index: 1, Name: "Hostname"}},
		}
		v2 := plugin.PluginMetricType{
			Namespace_: []string{"foo", "Baz"},
			Version_:   2,
			Labels_:    []core.Label{{Index: 1, Name: "Hostname"}},
		}
		metricList := []core.Metric{v1, v2}
		mc.updateCache(metricList)
		Convey("Should be cached separately", func() {
			Convey("so only 1 should be returned from the cache", func() {
				starMetric := &plugin.PluginMetricType{
					Namespace_: []string{"foo", "*"},
					Version_:   2,
				}
				// Check /foo/* with both versions
				toCollect, fromCache := mc.checkCache([]core.Metric{starMetric})
				So(len(toCollect), ShouldEqual, 0)
				So(len(fromCache), ShouldEqual, 1)
				starMetric.Version_ = 1
				toCollect, fromCache = mc.checkCache([]core.Metric{starMetric})
				So(len(toCollect), ShouldEqual, 0)
				So(len(fromCache), ShouldEqual, 1)
			})
		})
	})

	Convey("Having partial result in cache lookaside table should result in all metrics being collected", t, func() {
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()
		chrono.Chrono.Pause()
		mc := NewCache(GlobalCacheExpiration)
		v1 := plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
			Version_:   1,
			Labels_:    []core.Label{{Index: 1, Name: "Hostname"}},
		}
		v2 := plugin.PluginMetricType{
			Namespace_: []string{"foo", "Baz"},
			Version_:   1,
			Labels_:    []core.Label{{Index: 1, Name: "Hostname"}},
		}
		metricList := []core.Metric{v1, v2}
		mc.updateCache(metricList)
		//modify one entry in the table to make it invalid
		mc.table[genKeyFromMetric(v1)].time = chrono.Chrono.Now().Add(-1 * GlobalCacheExpiration)
		chrono.Chrono.Forward(100 * time.Millisecond)
		wildcard := plugin.PluginMetricType{
			Namespace_: []string{"foo", "*"},
			Version_:   1,
		}
		toCollect, fromCache := mc.checkCache([]core.Metric{wildcard})
		So(len(toCollect), ShouldEqual, 1)
		So(toCollect[0], ShouldResemble, wildcard)
		So(len(fromCache), ShouldEqual, 0)
	})
}
