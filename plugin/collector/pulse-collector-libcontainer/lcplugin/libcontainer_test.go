package lcplugin

import (
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func SkipTestGetMetrics(t *testing.T) {

	//TODO unskip when fixtures are tarballed
	Convey("Libcontainer cache s", t, func() {

		lc := NewLibCntr()

		Convey("empty for start", func() {
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

		})

	})
}
