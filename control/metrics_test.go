package control

import (
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricType(t *testing.T) {
	Convey("newMetricType()", t, func() {
		Convey("returns a metricType", func() {
			mt := newMetricType([]string{"test"}, time.Now(), new(loadedPlugin))
			So(mt, ShouldHaveSameTypeAs, new(metricType))
		})
	})
	Convey("metricType.Namespace()", t, func() {
		Convey("returns the namespace of a metricType", func() {
			ns := []string{"test"}
			mt := newMetricType(ns, time.Now(), new(loadedPlugin))
			So(mt.Namespace(), ShouldHaveSameTypeAs, ns)
			So(mt.Namespace(), ShouldResemble, ns)
		})
	})
	Convey("metricType.Version()", t, func() {
		Convey("returns the namespace of a metricType", func() {
			ns := []string{"test"}
			lp := &loadedPlugin{Meta: plugin.PluginMeta{Version: 1}}
			mt := newMetricType(ns, time.Now(), lp)
			So(mt.Version(), ShouldEqual, 1)
		})
	})
	Convey("metricType.LastAdvertisedTimestamp()", t, func() {
		Convey("returns the LastAdvertisedTimestamp for the metricType", func() {
			ts := time.Now()
			mt := newMetricType([]string{"test"}, ts, new(loadedPlugin))
			So(mt.LastAdvertisedTime(), ShouldHaveSameTypeAs, ts)
			So(mt.LastAdvertisedTime(), ShouldResemble, ts)
		})
	})
}

func TestMetricCatalog(t *testing.T) {
	Convey("newMetricCatalog()", t, func() {
		Convey("returns a metricCatalog", func() {
			mc := newMetricCatalog()
			So(mc, ShouldHaveSameTypeAs, new(metricCatalog))
		})
	})
	Convey("metricCatalog.Add()", t, func() {
		Convey("adds a metricType to the metricCatalog", func() {
			ns := []string{"test"}
			mt := newMetricType(ns, time.Now(), new(loadedPlugin))
			mc := newMetricCatalog()
			mc.Add(mt)
			_mt, err := mc.Get(ns, -1)
			So(_mt, ShouldResemble, mt)
			So(err, ShouldBeNil)
		})
	})
	Convey("metricCatalog.Get()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now()
		Convey("add multiple metricTypes and get them back", func() {
			ns := [][]string{
				[]string{"test1"},
				[]string{"test2"},
				[]string{"test3"},
			}
			lp := new(loadedPlugin)
			mt := []*metricType{
				newMetricType(ns[0], ts, lp),
				newMetricType(ns[1], ts, lp),
				newMetricType(ns[2], ts, lp),
			}
			for _, v := range mt {
				mc.Add(v)
			}
			for k, v := range ns {
				_mt, err := mc.Get(v, -1)
				So(_mt, ShouldEqual, mt[k])
				So(err, ShouldBeNil)
			}
		})
		Convey("it returns the latest version", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType([]string{"foo", "bar"}, ts, lp2)
			mc.Add(m2)
			m35 := newMetricType([]string{"foo", "bar"}, ts, lp35)
			mc.Add(m35)
			m, err := mc.Get([]string{"foo", "bar"}, -1)
			So(err, ShouldBeNil)
			So(m, ShouldEqual, m35)
		})
		Convey("it returns the queried version", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType([]string{"foo", "bar"}, ts, lp2)
			mc.Add(m2)
			m35 := newMetricType([]string{"foo", "bar"}, ts, lp35)
			mc.Add(m35)
			m, err := mc.Get([]string{"foo", "bar"}, 2)
			So(err, ShouldBeNil)
			So(m, ShouldEqual, m2)
		})
		Convey("it returns metric not found if version doesn't exist", func() {
			lp2 := new(loadedPlugin)
			lp2.Meta.Version = 2
			lp35 := new(loadedPlugin)
			lp35.Meta.Version = 35
			m2 := newMetricType([]string{"foo", "bar"}, ts, lp2)
			mc.Add(m2)
			m35 := newMetricType([]string{"foo", "bar"}, ts, lp35)
			mc.Add(m35)
			_, err := mc.Get([]string{"foo", "bar"}, 7)
			So(err, ShouldResemble, errorMetricNotFound([]string{"foo", "bar"}, 7))
		})
	})
	Convey("metricCatalog.Table()", t, func() {
		Convey("returns a copy of the table", func() {
			mc := newMetricCatalog()
			mt := newMetricType([]string{"foo", "bar"}, time.Now(), &loadedPlugin{})
			mc.Add(mt)
			//TODO test tree
			//So(mc.Table(), ShouldHaveSameTypeAs, map[string][]*metricType{})
			//So(mc.Table()["foo.bar"], ShouldResemble, []*metricType{mt})
		})
	})
	Convey("metricCatalog.Remove()", t, func() {
		Convey("removes a metricType from the catalog", func() {
			ns := []string{"test"}
			mt := newMetricType(ns, time.Now(), new(loadedPlugin))
			mc := newMetricCatalog()
			mc.Add(mt)
			mc.Remove(ns)
			_mt, err := mc.Get(ns, -1)
			So(_mt, ShouldBeNil)
			So(err, ShouldResemble, errorMetricNotFound([]string{"test"}))
		})
	})
	Convey("metricCatalog.Next()", t, func() {
		ns := []string{"test"}
		mt := newMetricType(ns, time.Now(), new(loadedPlugin))
		mc := newMetricCatalog()
		Convey("returns false on empty table", func() {
			ok := mc.Next()
			So(ok, ShouldEqual, false)
		})
		Convey("returns true on populated table", func() {
			mc.Add(mt)
			ok := mc.Next()
			So(ok, ShouldEqual, true)
		})
	})
	Convey("metricCatalog.Item()", t, func() {
		ns := [][]string{
			[]string{"test1"},
			[]string{"test2"},
			[]string{"test3"},
		}
		lp := new(loadedPlugin)
		t := time.Now()
		mt := []*metricType{
			newMetricType(ns[0], t, lp),
			newMetricType(ns[1], t, lp),
			newMetricType(ns[2], t, lp),
		}
		mc := newMetricCatalog()
		for _, v := range mt {
			mc.Add(v)
		}
		Convey("return first key and item in table", func() {
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, getMetricKey(ns[0]))
			So(item, ShouldResemble, []*metricType{mt[0]})
		})
		Convey("return second key and item in table", func() {
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, getMetricKey(ns[1]))
			So(item, ShouldResemble, []*metricType{mt[1]})
		})
		Convey("return third key and item in table", func() {
			mc.Next()
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, getMetricKey(ns[2]))
			So(item, ShouldResemble, []*metricType{mt[2]})
		})
	})
}

func TestSubscribe(t *testing.T) {
	ns := [][]string{
		[]string{"test1"},
		[]string{"test2"},
		[]string{"test3"},
	}
	lp := new(loadedPlugin)
	ts := time.Now()
	mt := []*metricType{
		newMetricType(ns[0], ts, lp),
		newMetricType(ns[1], ts, lp),
		newMetricType(ns[2], ts, lp),
	}
	mc := newMetricCatalog()
	for _, v := range mt {
		mc.Add(v)
	}
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns an error", func() {
			err := mc.Subscribe([]string{"test4"}, -1)
			So(err, ShouldResemble, errorMetricNotFound([]string{"test4"}))
		})
	})
	Convey("when the metric is in the table", t, func() {
		Convey("then it gets correctly increments the count", func() {
			err := mc.Subscribe([]string{"test1"}, -1)
			So(err, ShouldBeNil)
			m, err2 := mc.Get([]string{"test1"}, -1)
			So(err2, ShouldBeNil)
			So(m.subscriptions, ShouldEqual, 1)
		})
	})
}

func TestUnsubscribe(t *testing.T) {
	ns := [][]string{
		[]string{"test1"},
		[]string{"test2"},
		[]string{"test3"},
	}
	lp := new(loadedPlugin)
	ts := time.Now()
	mt := []*metricType{
		newMetricType(ns[0], ts, lp),
		newMetricType(ns[1], ts, lp),
		newMetricType(ns[2], ts, lp),
	}
	mc := newMetricCatalog()
	for _, v := range mt {
		mc.Add(v)
	}
	Convey("when the metric is in the table", t, func() {
		Convey("then its subscription count is decremented", func() {
			err := mc.Subscribe([]string{"test1"}, -1)
			So(err, ShouldBeNil)
			err1 := mc.Unsubscribe([]string{"test1"}, -1)
			So(err1, ShouldBeNil)
			m, err2 := mc.Get([]string{"test1"}, -1)
			So(err2, ShouldBeNil)
			So(m.subscriptions, ShouldEqual, 0)
		})
	})
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns metric not found error", func() {
			err := mc.Unsubscribe([]string{"test4"}, -1)
			So(err, ShouldResemble, errorMetricNotFound([]string{"test4"}))
		})
	})
	Convey("when the metric's count is already 0", t, func() {
		Convey("then it returns negative subscription count error", func() {
			err := mc.Unsubscribe([]string{"test1"}, -1)
			So(err, ShouldResemble, errNegativeSubCount)
		})
	})
}

func TestSubscriptionCount(t *testing.T) {
	m := newMetricType([]string{"test"}, time.Now(), &loadedPlugin{})
	Convey("it returns the subscription count", t, func() {
		m.Subscribe()
		So(m.SubscriptionCount(), ShouldEqual, 1)
		m.Subscribe()
		m.Subscribe()
		So(m.SubscriptionCount(), ShouldEqual, 3)
		m.Unsubscribe()
		So(m.SubscriptionCount(), ShouldEqual, 2)
	})
}
