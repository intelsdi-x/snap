package schedule

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkerManager(t *testing.T) {
	Convey("WorkerManager", t, func() {
		manager := new(managesWork)
		So(manager, ShouldNotBeNil)
		Convey("Work", func() {
			metricTypes := []core.MetricType{
				&MockMetricType{
					namespace:               []string{"foo", "bar"},
					version:                 1,
					lastAdvertisedTimestamp: 0,
				},
			}
			deadline := time.Now().Add(1 * time.Second)
			job := NewCollectorJob(metricTypes, deadline)
			So(job.(CollectorJob).Deadline(), ShouldResemble, deadline)
			job = manager.Work(job)
			So(job.Errors(), ShouldBeNil)
			So(job.(CollectorJob).Metrics(), ShouldBeNil)
		})
	})
}
