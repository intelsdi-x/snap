package control

import (
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricType(t *testing.T) {
	Convey("newMetricType()", t, func() {
		Convey("returns a metricType", func() {
			mt := newMetricType([]string{"test"}, time.Now().Unix(), new(loadedPlugin))
			So(mt, ShouldHaveSameTypeAs, new(metricType))
		})
	})
	Convey("metricType.Namespace()", t, func() {
		Convey("returns the namespace of a metricType", func() {
			ns := []string{"test"}
			mt := newMetricType(ns, time.Now().Unix(), new(loadedPlugin))
			So(mt.Namespace(), ShouldHaveSameTypeAs, ns)
			So(mt.Namespace(), ShouldResemble, ns)
		})
	})
	Convey("metricType.LastAdvertisedTimestamp()", t, func() {
		Convey("returns the LastAdvertisedTimestamp for the metricType", func() {
			ts := time.Now().Unix()
			mt := newMetricType([]string{"test"}, ts, new(loadedPlugin))
			So(mt.LastAdvertisedTimestamp(), ShouldHaveSameTypeAs, ts)
			So(mt.LastAdvertisedTimestamp(), ShouldResemble, ts)
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
			mt := newMetricType(ns, time.Now().Unix(), new(loadedPlugin))
			mc := newMetricCatalog()
			mc.Add(mt)
			_mt, err := mc.Get(ns, -1)
			So(_mt, ShouldResemble, mt)
			So(err, ShouldBeNil)
		})
	})
	Convey("metricCatalog.Get()", t, func() {
		mc := newMetricCatalog()
		ts := time.Now().Unix()
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
	})
	Convey("metricCatalog.Table()", t, func() {
		Convey("returns a copy of the table", func() {
			mc := newMetricCatalog()
			mt := newMetricType([]string{"foo", "bar"}, time.Now().Unix(), &loadedPlugin{})
			mc.Add(mt)
			So(mc.Table(), ShouldHaveSameTypeAs, map[string]*metricType{})
			So(mc.Table()["foo.bar"], ShouldEqual, mt)
		})
	})
	Convey("metricCatalog.Remove()", t, func() {
		Convey("removes a metricType from the catalog", func() {
			ns := []string{"test"}
			mt := newMetricType(ns, time.Now().Unix(), new(loadedPlugin))
			mc := newMetricCatalog()
			mc.Add(mt)
			mc.Remove(ns)
			_mt, err := mc.Get(ns, -1)
			So(_mt, ShouldBeNil)
			So(err, ShouldResemble, errors.New("metric not found"))
		})
	})
	Convey("metricCatalog.Next()", t, func() {
		ns := []string{"test"}
		mt := newMetricType(ns, time.Now().Unix(), new(loadedPlugin))
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
		t := time.Now().Unix()
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
			So(item, ShouldResemble, mt[0])
		})
		Convey("return second key and item in table", func() {
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, getMetricKey(ns[1]))
			So(item, ShouldResemble, mt[1])
		})
		Convey("return third key and item in table", func() {
			mc.Next()
			mc.Next()
			mc.Next()
			key, item := mc.Item()
			So(key, ShouldEqual, getMetricKey(ns[2]))
			So(item, ShouldResemble, mt[2])
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
	ts := time.Now().Unix()
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
		Convey("then it gets added to the table", func() {
		})
	})
	Convey("when the metric is in the table", t, func() {
		Convey("then it gets correctly increments the count", func() {
		})
		Convey("then it does not add it twice to the keys array", func() {
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
	ts := time.Now().Unix()
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
		})
	})
	Convey("when the metric is not in the table", t, func() {
		Convey("then it returns the correct error", func() {
		})
	})
	Convey("when the metric's count is already 0", t, func() {
		Convey("then it returns the correct error", func() {
		})
	})
}
