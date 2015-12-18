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
	"github.com/intelsdi-x/snap/pkg/chrono"
)

func TestCache(t *testing.T) {
	GlobalCacheExpiration = time.Duration(300 * time.Millisecond)
	Convey("puts and gets a metric", t, func() {
		mc := NewCache(GlobalCacheExpiration)
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
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
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
		}
		mc.put("/foo/bar", 1, foo)
		chrono.Chrono.Forward(401 * time.Millisecond)

		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldBeNil)
	})
	Convey("hit and miss counts", t, func() {
		Convey("ticks hit count when a cache entry is hit", func() {
			mc := NewCache(400 * time.Millisecond)
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
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
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
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
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
			}
			mc.put("/foo/bar", 1, foo)
			chrono.Chrono.Forward(301 * time.Millisecond)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].misses, ShouldEqual, 1)
		})
	})

}
