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

package client

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
)

func TestCache(t *testing.T) {
	GlobalCacheExpiration = time.Duration(300 * time.Millisecond)
	Convey("puts and gets a metric", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
		}

		mc.put("/foo/bar", 1, foo)
		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldNotBeNil)
		So(ret, ShouldEqual, foo)
	})
	Convey("returns nil if the cache cell does not exist", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldBeNil)
	})
	Convey("returns nil if the cache cell has expired", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
		}
		mc.put("/foo/bar", 1, foo)
		time.Sleep(301 * time.Millisecond)

		ret := mc.get("/foo/bar", 1)
		So(ret, ShouldBeNil)
	})
	Convey("hit and miss counts", t, func() {
		Convey("ticks hit count when a cache entry is hit", func() {
			mc := &cache{
				table: make(map[string]*cachecell),
			}
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
			}
			mc.put("/foo/bar", 1, foo)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is still a hit", func() {
			// Make sure global clock is restored after test.
			defer core.Chrono.Reset()
			defer core.Chrono.Continue()

			// Use artificial time: pause to get base time.
			core.Chrono.Pause()

			mc := &cache{
				table: make(map[string]*cachecell),
			}
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
			}

			mc.put("/foo/bar", 1, foo)
			core.Chrono.Forward(250 * time.Millisecond)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is missed", func() {
			defer core.Chrono.Reset()
			defer core.Chrono.Continue()

			core.Chrono.Pause()

			mc := &cache{
				table: make(map[string]*cachecell),
			}
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
			}
			mc.put("/foo/bar", 1, foo)
			core.Chrono.Forward(301 * time.Millisecond)
			mc.get("/foo/bar", 1)
			So(mc.table["/foo/bar:1"].misses, ShouldEqual, 1)
		})
	})

}
