package schedule

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkerManager(t *testing.T) {
	Convey("WorkerManager", t, func() {
		manager := new(workManager)
		Convey("Work", func() {
			var job Job = new(job)
			job = manager.Work(job)
			So(job.Errors(), ShouldBeNil)
			So(job.Metrics(), ShouldBeNil)
		})
	})
}
