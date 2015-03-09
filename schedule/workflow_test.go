package schedule

import (
	"testing"

	"github.com/intelsdilabs/pulse/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflow(t *testing.T) {
	Convey("Workflow", t, func() {
		workManager := new(managesWork)
		wf := NewWorkflow(workManager)
		So(wf.workManager, ShouldNotBeNil)
		So(wf.state, ShouldNotBeNil)
		Convey("Add steps", func() {
			pubStep := new(publishStep)
			procStep := new(processStep)
			wf.rootStep.AddStep(pubStep).AddStep(procStep)
			So(wf.rootStep, ShouldNotBeNil)
			So(wf.rootStep.Steps(), ShouldNotBeNil)
			Convey("Start", func() {
				schedule := new(MockSchedule)
				mts := make([]core.MetricType, 0)
				task := NewTask(schedule, mts, wf)
				wf.Start(task)
				So(wf.State(), ShouldEqual, WorkflowStarted)
			})
		})
	})
}
