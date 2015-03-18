package lcplugin

import (
	"strings"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCache(t *testing.T) {
	Convey("Libcontainer cache tests", t, func() {

		lc := NewLibCntr()

		Convey("Cache should be empty for start", func() {
			So(lc.cache, ShouldBeEmpty)
		})

	})
}

func TestGetMetricsTypes(t *testing.T) {

	//TODO unskip when fixtures are tarballed
	SkipConvey("Libcontainer TestGetMetricsTypes", t, func() {

		lc := NewLibCntr()

		Convey("Cache should be empty for start", func() {
			So(lc.cache, ShouldBeEmpty)
		})

		Convey("Collect metrics", func() {
			beforeTimestamp := time.Now()
			mt, err := lc.GetMetricTypes()
			afterTimestamp := time.Now()

			So(mt, ShouldNotBeNil)
			So(err, ShouldBeNil)

			expectedNS := []string{vendor, prefix, common, "count"}
			var expectedIdx int
			for idx, val := range mt {
				ns := strings.Join(val.Namespace(), nsSeparator)
				exNs := strings.Join(expectedNS, nsSeparator)
				if ns == exNs {
					expectedIdx = idx
				}
			}
			So(mt[expectedIdx].Namespace(), ShouldResemble, expectedNS)
			So(mt[expectedIdx].LastAdvertisedTime(),
				ShouldHappenBetween,
				beforeTimestamp, afterTimestamp)

			Convey("Cache should contain key \"intel/libcontainer/common/count\"", func() {
				So(lc.cache["intel/libcontainer/common/count"].namespace, ShouldResemble, expectedNS)
				So(lc.cache["intel/libcontainer/common/count"].lastUpdate, ShouldHappenBetween,
					beforeTimestamp, afterTimestamp)
			})
		})

	})
}

func TestCollectMetrics(t *testing.T) {

	SkipConvey("Libcontainer cache s", t, func() {

		lc := NewLibCntr()

		Convey("empty for start", func() {
			So(lc.cache, ShouldBeEmpty)
		})

		SkipConvey("Get non-stale metric from cache", func() {

			mval := 558
			mtmsp := time.Now()
			mns := []string{vendor, prefix, common, "fake_metric"}
			lc.cache["fake_metric"] = newMetric(mval, mtmsp, mns)

			input := make([]plugin.PluginMetricType, 0, 1)
			input = append(input, plugin.PluginMetricType{Namespace_: mns, Version_: 1})

			expectedVal := plugin.PluginMetric{Namespace_: mns, Data_: 558}

			ret, err := lc.CollectMetrics(input)
			So(err, ShouldBeNil)
			So(ret[0], ShouldResemble, expectedVal)

		})

		Convey("Get container count metric from cache (needs refresh)", func() {
			mns := []string{vendor, prefix, common, "count"}
			input := make([]plugin.PluginMetricType, 0, 1)
			input = append(input, plugin.PluginMetricType{Namespace_: mns, Version_: 1})

			retVal, err := lc.CollectMetrics(input)
			So(err, ShouldBeNil)
			So(len(retVal), ShouldEqual, 1)
			So(retVal[0].Namespace(), ShouldResemble, mns)
			//			var intVal interface{}
			//			intVal = 1
			//			So(retVal[0].Data(), ShouldEqual, intVal)

		})
	})

	//TODO unskip when fixtures are tarballed
}
