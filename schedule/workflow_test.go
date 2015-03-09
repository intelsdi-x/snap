package schedule

import (
	"testing"

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
				schedule := new(MockSchedule)
				task := NewTask(schedule, nil)
				manager := new(managesWork)
				wf.Start(task, manager)
				So(wf.State(), ShouldEqual, WorkflowStarted)
			})
		})
	})
}
