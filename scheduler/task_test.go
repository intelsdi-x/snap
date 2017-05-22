// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	emitter = gomit.NewEventController()
)

func TestTask(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("Task", t, func() {
		sampleWFMap := wmap.Sample()
		wf, errs := wmapToWorkflow(sampleWFMap)
		So(errs, ShouldBeEmpty)
		c := &mockMetricManager{}
		Convey("task + simple schedule", func() {
			// create a simple schedule which equals to windowed schedule without start and stop time
			sch := schedule.NewWindowedSchedule(time.Millisecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task.Spin()
			time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
			So(task.state, ShouldEqual, core.TaskSpinning)

			task.Stop()
		})

		Convey("Task specified-name test", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter, core.SetTaskName("My name is unique"))
			So(err, ShouldBeNil)
			task.Spin()
			So(task.GetName(), ShouldResemble, "My name is unique")

		})
		Convey("Task default-name test", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task.Spin()
			So(task.GetName(), ShouldResemble, "Task-"+task.ID())

		})

		Convey("Task deadline duration test", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter, core.TaskDeadlineDuration(20*time.Second))
			So(err, ShouldBeNil)
			task.Spin()
			So(task.deadlineDuration, ShouldEqual, 20*time.Second)
			task.Option(core.TaskDeadlineDuration(20 * time.Second))

			So(core.TaskDeadlineDuration(2*time.Second), ShouldNotBeEmpty)

		})

		Convey("Tasks are created and creation of task table is checked", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task1, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task1.Spin()
			task.Spin()
			tC := newTaskCollection()
			tC.add(task)
			tC.add(task1)
			taskTable := tC.Table()

			So(len(taskTable), ShouldEqual, 2)

		})

		Convey("Task is created and starts to spin", func() {
			sch := schedule.NewWindowedSchedule(time.Second*5, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task.Spin()
			So(task.state, ShouldEqual, core.TaskSpinning)
			Convey("Task is Stopped", func() {
				task.Stop()
				time.Sleep(time.Millisecond * 10) // it is a race so we slow down the test
				So(task.state, ShouldEqual, core.TaskStopped)
			})
		})

		Convey("task fires", func() {
			sch := schedule.NewWindowedSchedule(time.Nanosecond*100, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task.Spin()
			time.Sleep(time.Millisecond * 50)
			So(task.hitCount, ShouldBeGreaterThan, 2)
			So(task.missedIntervals, ShouldBeGreaterThan, 2)
			task.Stop()
		})

		Convey("Enable a running task", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*10, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)
			task.Spin()
			err = task.Enable()
			So(err, ShouldNotBeNil)
			So(task.State(), ShouldBeIn, []core.TaskState{core.TaskSpinning, core.TaskFiring})
		})

		Convey("Enable a disabled task", func() {
			sch := schedule.NewWindowedSchedule(time.Millisecond*10, nil, nil, 0)
			task, err := newTask(sch, wf, newWorkManager(), c, emitter)
			So(err, ShouldBeNil)

			task.state = core.TaskDisabled
			err = task.Enable()
			So(err, ShouldBeNil)
			So(task.State(), ShouldEqual, core.TaskStopped)
		})
	})

	Convey("Create task collection", t, func() {
		sampleWFMap := wmap.Sample()
		wf, errs := wmapToWorkflow(sampleWFMap)
		So(errs, ShouldBeEmpty)

		sch := schedule.NewWindowedSchedule(time.Millisecond*10, nil, nil, 0)
		task, err := newTask(sch, wf, newWorkManager(), &mockMetricManager{}, emitter)
		So(err, ShouldBeNil)
		So(task.id, ShouldNotBeEmpty)
		So(task.id, ShouldNotBeNil)
		taskCollection := newTaskCollection()

		Convey("Add task to collection", func() {

			err := taskCollection.add(task)
			So(err, ShouldBeNil)
			So(len(taskCollection.table), ShouldEqual, 1)

			Convey("Attempt to add the same task again", func() {
				err := taskCollection.add(task)
				So(err, ShouldNotBeNil)
			})

			Convey("Get task from collection", func() {
				t := taskCollection.Get(task.id)
				So(t, ShouldNotBeNil)
				So(t.ID(), ShouldEqual, task.id)
				So(t.CreationTime().Nanosecond(), ShouldBeLessThan, time.Now().Nanosecond())
				So(t.HitCount(), ShouldEqual, 0)
				So(t.MissedCount(), ShouldEqual, 0)
				So(t.State(), ShouldEqual, core.TaskStopped)
				So(t.Status(), ShouldEqual, core.WorkflowStopped)
				So(t.LastRunTime().IsZero(), ShouldBeTrue)
			})

			Convey("Attempt to get task with an invalid Id", func() {
				t := taskCollection.Get("1234")
				So(t, ShouldBeNil)
			})

			Convey("Create another task and compare the id", func() {
				task2, err := newTask(sch, wf, newWorkManager(), &mockMetricManager{}, emitter)
				So(err, ShouldBeNil)
				So(task2.id, ShouldNotEqual, task.ID())
			})

		})

	})
}
