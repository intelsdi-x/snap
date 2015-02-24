package schedule

import (
	"testing"

	"github.com/intelsdilabs/pulse/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflow(t *testing.T) {
	Convey("Workflow", t, func() {
		wf := NewWorkflow()
		Convey("Add steps", func() {
			pubStep := new(publishStep)
			procStep := new(processStep)
			wf.rootStep.AddStep(pubStep).AddStep(procStep)
			So(wf.rootStep, ShouldNotBeNil)
			So(wf.rootStep.Steps(), ShouldNotBeNil)
			Convey("Start", func() {
				metricTypes := make([]core.MetricType, 0)
				manager := new(workManager)
				wf.Start(metricTypes, manager)
				So(wf.State(), ShouldEqual, WorkflowStarted)
			})
		})
	})
}
