package control

import (
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
			So(mc.table[getMetricKey(ns)], ShouldResemble, mt)
		})
	})

}
