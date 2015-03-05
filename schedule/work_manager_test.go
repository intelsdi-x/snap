package schedule

import (
	"testing"

	"github.com/intelsdilabs/pulse/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkerManager(t *testing.T) {
	Convey("WorkerManager", t, func() {
		manager = newWorkManager(int64(5), 1)
		Convey("Work", func() {
			metricTypes := make([]core.MetricType, 0)
			var j job
			j = newCollectorJob(metricTypes)
			j = manager.Work(j)
			So(j.Errors(), ShouldBeNil)
			So(j.(*collectorJob).Metrics(), ShouldBeNil)
		})
	})
}
