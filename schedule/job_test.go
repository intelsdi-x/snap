package schedule

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCollectorJob(t *testing.T) {
	Convey("newCollectorJob()", t, func() {
		Convey("it returns an init-ed collectorJob", func() {
			cj := newCollectorJob([]core.MetricType{})
			So(cj, ShouldHaveSameTypeAs, &collectorJob{})
		})
	})
	Convey("StartTime()", t, func() {
		Convey("it should return the job starttime", func() {
			cj := newCollectorJob([]core.MetricType{})
			So(cj.StartTime(), ShouldHaveSameTypeAs, time.Now().Unix())
		})
	})
	Convey("Deadline()", t, func() {
		Convey("it should return the job daedline", func() {
			cj := newCollectorJob([]core.MetricType{})
			So(cj.Deadline(), ShouldEqual, defaultDeadline)
		})
	})
}
