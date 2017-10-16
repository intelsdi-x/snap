// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/fixtures"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type mockMetricManager struct {
	failValidatingMetrics      bool
	failValidatingMetricsAfter int
	failuredSoFar              int
	autodiscoverPaths          []string
	timeToWait                 time.Duration
}

func (m *mockMetricManager) StreamMetrics(string, map[string]map[string]string, time.Duration, int64) (chan []core.Metric, chan error, []error) {
	return nil, nil, nil
}

func (m *mockMetricManager) CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error) {
	time.Sleep(m.timeToWait)
	return nil, nil
}

func (m *mockMetricManager) PublishMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) []error {
	return nil
}

func (m *mockMetricManager) ProcessMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) ([]core.Metric, []error) {
	return nil, nil
}

func (m *mockMetricManager) ValidateDeps(mts []core.RequestedMetric, prs []core.SubscribedPlugin, ctree *cdata.ConfigDataTree, asserts ...core.SubscribedPluginAssert) []serror.SnapError {
	if m.failValidatingMetrics {
		return []serror.SnapError{
			serror.New(errors.New("metric validation error")),
		}
	}
	return nil
}
func (m *mockMetricManager) SubscribeDeps(taskID string, reqs []core.RequestedMetric, prs []core.SubscribedPlugin, ctree *cdata.ConfigDataTree) []serror.SnapError {
	return []serror.SnapError{
		serror.New(errors.New("metric validation error")),
	}
}

func (m *mockMetricManager) UnsubscribeDeps(taskID string) []serror.SnapError {
	return nil
}

func (m *mockMetricManager) SetAutodiscoverPaths(paths []string) {
	m.autodiscoverPaths = paths
}

func (m *mockMetricManager) GetAutodiscoverPaths() []string {
	return m.autodiscoverPaths
}

type mockMetricManagerError struct {
	errs []error
}

type mockMetricType struct {
	version            int
	namespace          []string
	lastAdvertisedTime time.Time
	config             *cdata.ConfigDataNode
}

func (m mockMetricType) Version() int {
	return m.version
}

func (m mockMetricType) Namespace() []string {
	return m.namespace
}

func (m mockMetricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func (m mockMetricType) Config() *cdata.ConfigDataNode {
	return m.config
}

func (m mockMetricType) Data() interface{} {
	return nil
}

type mockScheduleResponse struct {
}

func (m mockScheduleResponse) state() schedule.ScheduleState {
	return schedule.Active
}

func (m mockScheduleResponse) err() error {
	return nil
}

func (m mockScheduleResponse) missedIntervals() uint {
	return 0
}

// Helper constructor functions for re-use amongst tests
func newMockMetricManager() *mockMetricManager {
	m := new(mockMetricManager)
	return m
}

func newScheduler() *scheduler {
	cfg := GetDefaultConfig()
	s := New(cfg)
	s.SetMetricManager(newMockMetricManager())
	return s
}

func newMockWorkflowMap() *wmap.WorkflowMap {
	w := wmap.NewWorkflowMap()
	// Collection node
	w.Collect.AddMetric("/foo/bar", 1)
	w.Collect.AddMetric("/foo/baz", 2)
	w.Collect.AddConfigItem("/foo/bar", "username", "root")
	w.Collect.AddConfigItem("/foo/bar", "port", 8080)
	w.Collect.AddConfigItem("/foo/bar", "ratio", 0.32)
	w.Collect.AddConfigItem("/foo/bar", "yesorno", true)

	// Add a process node
	pr1 := wmap.NewProcessNode("machine", 1)
	pr1.AddConfigItem("username", "wat")
	pr1.AddConfigItem("howmuch", 9999)

	// Add a process node
	pr12 := wmap.NewProcessNode("machine", 1)
	pr12.AddConfigItem("username", "wat2")
	pr12.AddConfigItem("howmuch", 99992)

	// Publish node for our process node
	pu1 := wmap.NewPublishNode("rmq", -1)
	pu1.AddConfigItem("birthplace", "dallas")
	pu1.AddConfigItem("monies", 2)

	// Publish node direct to collection
	pu2 := wmap.NewPublishNode("file", -1)
	pu2.AddConfigItem("color", "brown")
	pu2.AddConfigItem("purpose", 42)

	pr12.Add(pu2)
	pr1.Add(pr12)
	w.Collect.Add(pr1)
	w.Collect.Add(pu1)
	return w
}

var (
	startWait  = time.Millisecond * 50
	windowSize = time.Millisecond * 100
	interval   = time.Millisecond * 10
)

// ----------------------------- Medium Tests ----------------------------

func TestCreateTask(t *testing.T) {
	s := newScheduler()
	s.Start()
	w := newMockWorkflowMap()

	Convey("Calling CreateTask for a simple schedule", t, func() {
		Convey("returns an error when the schedule does not validate", func() {
			Convey("the interval is invalid", func() {
				Convey("the interval equals zero", func() {
					invalidInterval := 0 * time.Millisecond
					// create a simple schedule which equals to windowed schedule
					// without start and stop time
					sch := schedule.NewWindowedSchedule(invalidInterval, nil, nil, 0)
					tsk, errs := s.CreateTask(sch, w, false)
					So(errs, ShouldNotBeEmpty)
					So(tsk, ShouldBeNil)
					So(errs.Errors()[0].Error(), ShouldEqual, schedule.ErrInvalidInterval.Error())
				})
				Convey("the interval is less than zero", func() {
					invalidInterval := (-1) * time.Millisecond
					sch := schedule.NewWindowedSchedule(invalidInterval, nil, nil, 0)
					tsk, errs := s.CreateTask(sch, w, false)
					So(errs, ShouldNotBeEmpty)
					So(tsk, ShouldBeNil)
					So(errs.Errors()[0].Error(), ShouldEqual, schedule.ErrInvalidInterval.Error())
				})
			})
		})
		Convey("should not error when the schedule is valid", func() {
			// create a simple schedule which equals to windowed schedule
			// without start and stop time
			sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
			tsk, errs := s.CreateTask(sch, w, false)
			So(errs.Errors(), ShouldBeEmpty)
			So(tsk, ShouldNotBeNil)
		})
	}) //end of tests for a simple scheduler

	Convey("Calling CreateTask for a windowed schedule", t, func() {
		Convey("returns an error when the schedule does not validate", func() {
			Convey("the stop time was set in the past", func() {
				start := time.Now().Add(startWait)
				stop := time.Now().Add(time.Second * -10)
				sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
				tsk, errs := s.CreateTask(sch, w, false)
				So(errs, ShouldNotBeEmpty)
				So(tsk, ShouldBeNil)
				So(errs.Errors()[0].Error(), ShouldEqual, schedule.ErrInvalidStopTime.Error())
			})
			Convey("the stop time is before the start time", func() {
				start := time.Now().Add(startWait * 2)
				stop := time.Now().Add(startWait)
				sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
				tsk, errs := s.CreateTask(sch, w, false)
				So(errs, ShouldNotBeEmpty)
				So(tsk, ShouldBeNil)
				So(errs.Errors()[0].Error(), ShouldEqual, schedule.ErrStopBeforeStart.Error())
			})
		})
		Convey("should not error when the schedule is valid", func() {
			lse := fixtures.NewListenToSchedulerEvent()
			s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
			start := time.Now().Add(startWait)
			stop := time.Now().Add(startWait + windowSize)
			sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
			tsk, errs := s.CreateTask(sch, w, false)
			So(errs.Errors(), ShouldBeEmpty)
			So(tsk, ShouldNotBeNil)

			task := s.tasks.Get(tsk.ID())
			task.Spin()
			Convey("the task should be ended after reaching the end of window", func() {
				// wait for task ended event (or timeout)
				select {
				case <-lse.Ended:
				case <-time.After(stop.Add(interval + 1*time.Second).Sub(start)):
				}

				So(tsk.State(), ShouldEqual, core.TaskEnded)
			})
		})
	}) //end of tests for a windowed scheduler

	Convey("Calling CreateTask for a simple/windowed schedule with determined the count of runs", t, func() {
		Convey("Single run task firing immediately", func() {
			lse := fixtures.NewListenToSchedulerEvent()
			s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
			count := uint(1)
			sch := schedule.NewWindowedSchedule(interval, nil, nil, count)
			tsk, errs := s.CreateTask(sch, w, false)
			So(errs.Errors(), ShouldBeEmpty)
			So(tsk, ShouldNotBeNil)

			task := s.tasks.Get(tsk.ID())
			task.Spin()

			Convey("the task should be ended after reaching the end of window", func() {
				// wait for task ended event (or timeout)
				select {
				case <-lse.Ended:
				case <-time.After(time.Duration(int64(count)*interval.Nanoseconds()) + 1*time.Second):
				}
			})
		})
		Convey("Single run task firing on defined start time", func() {
			lse := fixtures.NewListenToSchedulerEvent()
			s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
			count := uint(1)
			start := time.Now().Add(startWait)
			sch := schedule.NewWindowedSchedule(interval, &start, nil, count)
			tsk, errs := s.CreateTask(sch, w, false)
			So(errs.Errors(), ShouldBeEmpty)
			So(tsk, ShouldNotBeNil)

			task := s.tasks.Get(tsk.ID())
			task.Spin()
			Convey("the task should be ended after reaching the end of window", func() {
				// wait for task ended event (or timeout)
				select {
				case <-lse.Ended:
				case <-time.After(time.Duration(int64(count)*interval.Nanoseconds()) + 1*time.Second):
				}
				// check if the task is ended
				So(tsk.State(), ShouldEqual, core.TaskEnded)
			})
		})
	}) //end of tests for simple/windowed schedule with determined the count

	Convey("Calling CreateTask for a cron schedule", t, func() {
		Convey("returns an error when the schedule does not validate", func() {
			Convey("the cron entry is empty", func() {
				cronEntry := ""
				tsk, errs := s.CreateTask(schedule.NewCronSchedule(cronEntry), w, false)
				So(errs, ShouldNotBeEmpty)
				So(tsk, ShouldBeNil)
				So(errs.Errors()[0].Error(), ShouldEqual, schedule.ErrMissingCronEntry.Error())
			})
			Convey("the cron entry is invalid", func() {
				cronEntry := "0 30"
				tsk, errs := s.CreateTask(schedule.NewCronSchedule(cronEntry), w, false)
				So(errs, ShouldNotBeEmpty)
				So(tsk, ShouldBeNil)
				So(errs.Errors()[0].Error(), ShouldStartWith, "Expected 5 or 6 fields")
			})
		})
		Convey("should not error when the schedule is valid", func() {
			cronEntry := "0 30 * * * *"
			tsk, errs := s.CreateTask(schedule.NewCronSchedule(cronEntry), w, false)
			So(errs.Errors(), ShouldBeEmpty)
			So(tsk, ShouldNotBeNil)
		})
	}) //end of tests for a cron scheduler

	s.Stop()
}

func TestStopTask(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	c := new(mockMetricManager)
	cfg := GetDefaultConfig()
	s := New(cfg)
	s.SetMetricManager(c)
	w := newMockWorkflowMap()
	s.Start()

	Convey("Calling StopTask on a running task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tsk, _ := s.CreateTask(sch, w, false)
		So(tsk, ShouldNotBeNil)
		task := s.tasks.Get(tsk.ID())
		task.Spin()
		// check if the task is running
		So(core.TaskStateLookup[task.State()], ShouldEqual, "Running")

		// stop the running task
		err := s.StopTask(tsk.ID())
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		time.Sleep(100 * time.Millisecond)
		Convey("State of the task should be TaskStopped", func() {
			So(tsk.State(), ShouldEqual, core.TaskStopped)
		})
	})

	Convey("Calling StopTask on a firing task", t, func() {
		Convey("Should allow last scheduled workfow execution to finish", func() {
			c.timeToWait = 500 * time.Millisecond
			lse := fixtures.NewListenToSchedulerEvent()
			s.eventManager.RegisterHandler("Scheduler.TaskStopped", lse)
			sc := schedule.NewWindowedSchedule(time.Second, nil, nil, 0)
			t, _ := s.CreateTask(sc, w, false)
			startTime := time.Now()
			t.(*task).Spin()
			// allowing things to settle and waiting for task state to change to firing
			time.Sleep(100 * time.Millisecond)
			So(t.State(), ShouldResemble, core.TaskFiring)

			// stop task when task state is firing
			t.(*task).Stop()
			// the last scheduled workflow execution should be allowed to finish
			// so we expect that stopping the task is going to happen not early than 500ms (set by by timeToWait)

			select {
			case <-lse.TaskStoppedEvents:
				// elapsed time should be greater than 500ms
				So(time.Since(startTime), ShouldBeGreaterThan, c.timeToWait)
				So(t.State(), ShouldResemble, core.TaskStopped)

				// the task should have fired once
				So(t.HitCount(), ShouldEqual, 1)
			case <-time.After(1 * time.Second):
			}
		})
	})

	Convey("Calling StopTask on a stopped task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tskStopped, _ := s.CreateTask(sch, w, false)
		So(tskStopped, ShouldNotBeNil)
		// check if the task is already stopped
		So(tskStopped.State(), ShouldEqual, core.TaskStopped)

		// try to stop the stopped task
		err := s.StopTask(tskStopped.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is already stopped.", func() {
			So(err[0].Error(), ShouldEqual, ErrTaskAlreadyStopped.Error())
		})
	})
	Convey("Calling StopTask on a disabled task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tskDisabled, _ := s.CreateTask(sch, w, false)
		So(tskDisabled, ShouldNotBeNil)
		taskDisabled := s.tasks.Get(tskDisabled.ID())
		taskDisabled.state = core.TaskDisabled

		// try to stop the disabled task
		err := s.StopTask(tskDisabled.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is disabled. Only running tasks can be stopped.", func() {
			So(err[0].Error(), ShouldEqual, ErrTaskDisabledNotStoppable.Error())
		})
		Convey("State of the task should be still TaskDisabled", func() {
			So(tskDisabled.State(), ShouldEqual, core.TaskDisabled)
		})
	})
	Convey("Calling StopTask on an ended task", t, func() {
		lse := fixtures.NewListenToSchedulerEvent()
		s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
		start := time.Now().Add(startWait)
		stop := time.Now().Add(startWait + windowSize)

		// create a task with windowed schedule
		sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
		tsk, errs := s.CreateTask(sch, w, false)
		So(errs.Errors(), ShouldBeEmpty)
		So(tsk, ShouldNotBeNil)

		task := s.tasks.Get(tsk.ID())
		task.Spin()

		// wait for task ended event (or timeout)
		select {
		case <-lse.Ended:
		case <-time.After(stop.Add(interval + 1*time.Second).Sub(start)):
		}

		// check if the task is ended
		So(tsk.State(), ShouldEqual, core.TaskEnded)

		// try to stop the ended task
		err := s.StopTask(tsk.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is ended. Only running tasks can be stopped.", func() {
			So(err[0].Error(), ShouldEqual, ErrTaskEndedNotStoppable.Error())
		})
		Convey("State of the task should be still TaskEnded", func() {
			So(tsk.State(), ShouldEqual, core.TaskEnded)
		})
	})

	s.Stop()
}

func TestStartTask(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	s := newScheduler()
	s.Start()
	w := newMockWorkflowMap()

	Convey("Calling StartTask a running task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tsk, _ := s.CreateTask(sch, w, false)
		So(tsk, ShouldNotBeNil)

		task := s.tasks.Get(tsk.ID())
		task.Spin()
		// check if the task is running
		So(core.TaskStateLookup[task.State()], ShouldEqual, "Running")

		// try to start the running task
		err := s.StartTask(tsk.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is already running.", func() {
			So(err[0].Error(), ShouldEqual, ErrTaskAlreadyRunning.Error())
		})
		Convey("State of the task should be still Running", func() {
			So(core.TaskStateLookup[task.State()], ShouldEqual, "Running")
		})

		task.Stop()
	})
	Convey("Calling StartTask on a disabled task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tskDisabled, _ := s.CreateTask(sch, w, false)
		So(tskDisabled, ShouldNotBeNil)
		taskDisabled := s.tasks.Get(tskDisabled.ID())
		taskDisabled.state = core.TaskDisabled

		// try to start the disabled task
		err := s.StartTask(tskDisabled.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is disabled. Cannot be started", func() {
			So(err[0].Error(), ShouldEqual, ErrTaskDisabledNotRunnable.Error())
		})
		Convey("State of the task should be still TaskDisabled", func() {
			So(tskDisabled.State(), ShouldEqual, core.TaskDisabled)
		})
	})
	Convey("Calling StartTask on an ended windowed task", t, func() {
		lse := fixtures.NewListenToSchedulerEvent()
		s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
		start := time.Now().Add(startWait)
		stop := time.Now().Add(startWait + windowSize)

		//create a task with windowed schedule
		sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
		tsk, errs := s.CreateTask(sch, w, false)
		So(errs.Errors(), ShouldBeEmpty)
		So(tsk, ShouldNotBeNil)

		task := s.tasks.Get(tsk.ID())
		task.Spin()

		// wait for task ended event (or timeout)
		select {
		case <-lse.Ended:
		case <-time.After(stop.Add(interval + 1*time.Second).Sub(start)):
		}

		// check if the task is ended
		So(tsk.State(), ShouldEqual, core.TaskEnded)

		// try to restart the ended windowed task for which the stop time is in the past
		err := s.StartTask(tsk.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		// the schedule is not longer valid at this point of time
		Convey("Error should read: Stop time is in the past", func() {
			So(err[0].Error(), ShouldEqual, schedule.ErrInvalidStopTime.Error())
		})
		Convey("State of the task should be still TaskEnded", func() {
			So(tsk.State(), ShouldEqual, core.TaskEnded)
		})
	})

	s.Stop()
}

func TestEnableTask(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	s := newScheduler()
	s.Start()
	w := newMockWorkflowMap()

	Convey("Calling EnableTask on a disabled task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tskDisabled, _ := s.CreateTask(sch, w, false)
		So(tskDisabled, ShouldNotBeNil)
		taskDisabled := s.tasks.Get(tskDisabled.ID())
		taskDisabled.state = core.TaskDisabled

		// enable the disabled task
		tskEnabled, err := s.EnableTask(tskDisabled.ID())
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		Convey("State of the task should be TaskStopped after enabling", func() {
			So(tskEnabled, ShouldNotBeNil)
			// EnableTask changes state from disabled to stopped
			So(tskEnabled.State(), ShouldEqual, core.TaskStopped)
		})
	})
	Convey("Calling EnableTask on a running task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tsk, _ := s.CreateTask(sch, w, false)
		So(tsk, ShouldNotBeNil)
		task := s.tasks.Get(tsk.ID())
		task.Spin()
		// check if the task is running
		So(core.TaskStateLookup[task.State()], ShouldEqual, "Running")

		// try to enable the running task
		_, err := s.EnableTask(tsk.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task must be disabled.", func() {
			So(err, ShouldEqual, ErrTaskNotDisabled)
		})
	})
	Convey("Calling EnableTask on a stopped task", t, func() {
		sch := schedule.NewWindowedSchedule(interval, nil, nil, 0)
		tskStopped, _ := s.CreateTask(sch, w, false)
		So(tskStopped, ShouldNotBeNil)
		// check if the task is already stopped
		So(tskStopped.State(), ShouldEqual, core.TaskStopped)

		// try to enable the stopped task
		_, err := s.EnableTask(tskStopped.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task must be disabled.", func() {
			So(err, ShouldEqual, ErrTaskNotDisabled)
		})
	})
	Convey("Calling EnableTask on an ended task", t, func() {
		lse := fixtures.NewListenToSchedulerEvent()
		s.eventManager.RegisterHandler("Scheduler.TaskEnded", lse)
		start := time.Now().Add(startWait)
		stop := time.Now().Add(startWait + windowSize)

		// create a task with windowed schedule
		sch := schedule.NewWindowedSchedule(interval, &start, &stop, 0)
		tsk, errs := s.CreateTask(sch, w, false)
		So(errs.Errors(), ShouldBeEmpty)
		So(tsk, ShouldNotBeNil)

		task := s.tasks.Get(tsk.ID())
		task.Spin()

		// wait for task ended event (or timeout)
		select {
		case <-lse.Ended:
		case <-time.After(stop.Add(interval + 1*time.Second).Sub(start)):
		}

		// check if the task is ended
		So(tsk.State(), ShouldEqual, core.TaskEnded)

		// try to enable the ended task
		_, err := s.EnableTask(tsk.ID())
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task must be disabled.", func() {
			So(err, ShouldEqual, ErrTaskNotDisabled)
		})
	})

	s.Stop()
}
