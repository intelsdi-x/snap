package client

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/pulse/control/plugin"
)

func TestCache(t *testing.T) {
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
