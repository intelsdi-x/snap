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
		foo := &plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
		}

		mc.put("/foo/bar", 1, foo)
		ret := mc.get("/foo/bar", 1)

		So(ret, ShouldNotBeNil)
		So(ret, ShouldEqual, foo)
	})
	Convey("returns nil if the cache cell does not exist", t, func() {
		mc := NewCache(GlobalCacheExpiration)
		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldBeNil)
	})
	Convey("returns nil if the cache cell has expired", t, func() {
		// Make sure global clock is restored after test.
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()

		// Use artificial time: pause to get base time.
		chrono.Chrono.Pause()

		mc := NewCache(400 * time.Millisecond)
		foo := &plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
		}
		mc.put("/foo/bar", 1, foo)
		chrono.Chrono.Forward(401 * time.Millisecond)

		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldBeNil)
	})
	Convey("hit and miss counts", t, func() {
		Convey("ticks hit count when a cache entry is hit", func() {
			mc := NewCache(400 * time.Millisecond)
			foo := &plugin.MetricType{
				Namespace_: core.NewNamespace("foo", "bar"),
			}
			mc.put("/foo/bar", 1, foo)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is still a hit", func() {
			defer chrono.Chrono.Reset()
			defer chrono.Chrono.Continue()

			chrono.Chrono.Pause()

			mc := NewCache(400 * time.Millisecond)
			foo := &plugin.MetricType{
				Namespace_: core.NewNamespace("foo", "bar"),
			}

			mc.put("/foo/bar", 1, foo)
			chrono.Chrono.Forward(250 * time.Millisecond)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is missed", func() {
			defer chrono.Chrono.Reset()
			defer chrono.Chrono.Continue()

			chrono.Chrono.Pause()

			mc := NewCache(GlobalCacheExpiration)
			foo := &plugin.MetricType{
				Namespace_: core.NewNamespace("foo", "bar"),
			}
			mc.put("/foo/bar", 1, foo)
			chrono.Chrono.Forward(301 * time.Millisecond)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].misses, ShouldEqual, 1)
		})
	})

	Convey("Add and get metrics via updateCache and checkCache", t, func() {
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()

		chrono.Chrono.Pause()

		mc := NewCache(GlobalCacheExpiration)
		foo := &plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
		}
		baz := &plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "baz"),
		}
		metricList := []core.Metric{foo, baz}
		mc.updateCache(metricList)
		Convey("they should be retrievable via get", func() {
			ret := mc.get(foo.Namespace().String(), foo.Version())
			So(ret, ShouldEqual, foo)
			ret = mc.get(baz.Namespace().String(), baz.Version())
			So(ret, ShouldEqual, baz)
		})
		Convey("they should be retrievable via checkCache", func() {
			nonCached := &plugin.MetricType{
				Namespace_: core.NewNamespace("foo", "fooer"),
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
		})
	})

	Convey("Adding plugins with same namespace but different versions", t, func() {
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()
		chrono.Chrono.Pause()
		mc := NewCache(GlobalCacheExpiration)
		v1 := plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
			Version_:   1,
		}
		v2 := plugin.MetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
			Version_:   2,
		}
		metricList := []core.Metric{v1, v2}
		mc.updateCache(metricList)
		Convey("Should be cached separately", func() {
			Convey("so only 1 should be returned from the cache", func() {
				starMetric := &plugin.MetricType{
					Namespace_: core.NewNamespace("foo", "bar"),
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
}
