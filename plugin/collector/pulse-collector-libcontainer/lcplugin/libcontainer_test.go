package lcplugin

import (
	"strings"
	"testing"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMetrics(t *testing.T) {

	//TODO unskip when fixtures are tarballed
	SkipConvey("Libcontainer cache s", t, func() {

		lc := NewLibcontainerPlugin()

		Convey("empty for start", func() {
			So(lc.cache, ShouldBeEmpty)
		})

		Convey("Collect metrics", func() {
			var args plugin.GetMetricTypesArgs
			var reply plugin.GetMetricTypesReply
			//			beforeTimestamp := time.Now()
			lc.GetMetricTypes(args, &reply)
			//			afterTimestamp := time.Now()

			So(reply.MetricTypes, ShouldNotBeNil)

			expectedNS := []string{vendor, prefix, common, "count"}
			var expectedIdx int
			for idx, val := range reply.MetricTypes {
				ns := strings.Join(val.Namespace(), nsSeparator)
				exNs := strings.Join(expectedNS, nsSeparator)
				if ns == exNs {
					expectedIdx = idx
				}
			}
			So(reply.MetricTypes[expectedIdx].Namespace(), ShouldResemble, expectedNS)
			// TODO Uncomment when LastAdvertisedTimestamp will be time.Time
			//			So(reply.MetricTypes[expectedIdx].LastAdvertisedTimestamp(),
			//				ShouldHappenBetween,
			//				beforeTimestamp, afterTimestamp)

		})

	})
}
