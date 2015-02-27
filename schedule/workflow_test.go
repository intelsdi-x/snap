package schedule

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type mockSchedule struct{}

func (m *mockSchedule) Wait() chan struct{} {
	return nil
}

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
				schedule := new(mockSchedule)
				task := NewTask(schedule)
				manager := new(managesWork)
				wf.Start(task, manager)
				So(wf.State(), ShouldEqual, WorkflowStarted)
			})
		})
	})
}
