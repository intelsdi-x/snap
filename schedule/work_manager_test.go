package schedule

import (
	"testing"

	"github.com/intelsdilabs/pulse/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkerManager(t *testing.T) {
	Convey("WorkerManager", t, func() {
		manager := new(managesWork)
		Convey("Work", func() {
			metricTypes := make([]core.MetricType, 0)
			job := NewCollectorJob(metricTypes)
			job = manager.Work(job)
			So(job.Errors(), ShouldBeNil)
			So(job.(CollectorJob).Metrics(), ShouldBeNil)
		})
	})
}
