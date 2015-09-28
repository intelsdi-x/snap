/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
)

func TestCache(t *testing.T) {
	CacheExpiration = time.Duration(500 * time.Millisecond)
	Convey("puts and gets a metric", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
		}
		mc.put("/foo/bar", foo)
		ret := mc.get("/foo/bar")
		So(ret, ShouldNotBeNil)
		So(ret, ShouldEqual, foo)
	})
	Convey("returns nil if the cache cell does not exist", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		ret := mc.get("/foo/bar")
		So(ret, ShouldBeNil)
	})
	Convey("returns nil if the cache cell has expired", t, func() {
		mc := &cache{
			table: make(map[string]*cachecell),
		}
		foo := &plugin.PluginMetricType{
			Namespace_: []string{"foo", "bar"},
		}
		mc.put("/foo/bar", foo)
		time.Sleep(501 * time.Millisecond)
		ret := mc.get("/foo/bar")
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
			mc.put("/foo/bar", foo)
			mc.get("/foo/bar")
			So(mc.table["/foo/bar"].hits, ShouldEqual, 1)
		})
		Convey("ticks miss count when a cache entry is missed", func() {
			mc := &cache{
				table: make(map[string]*cachecell),
			}
			foo := &plugin.PluginMetricType{
				Namespace_: []string{"foo", "bar"},
			}
			mc.put("/foo/bar", foo)
			time.Sleep(501 * time.Millisecond)
			mc.get("/foo/bar")
			So(mc.table["/foo/bar"].misses, ShouldEqual, 1)
		})
	})

}
