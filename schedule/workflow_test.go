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
				workerKillChan = make(chan struct{})
				manager := newWorkManager(int64(5), 1)
				schedule := new(mockSchedule)
				task := NewTask(schedule, manager)
				wf.Start(task, manager)
				So(wf.State(), ShouldEqual, WorkflowStarted)
			})
		})
	})
}
