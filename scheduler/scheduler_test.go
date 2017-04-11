// +build legacy

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
	"fmt"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type mockMetricManager struct {
	failValidatingMetrics      bool
	failValidatingMetricsAfter int
	failuredSoFar              int
	autodiscoverPaths          []string
}

func (m *mockMetricManager) StreamMetrics(string, map[string]map[string]string, time.Duration, int64) (chan []core.Metric, chan error, []error) {
	return nil, nil, nil
}

func (m *mockMetricManager) CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error) {
	return nil, nil
}

func (m *mockMetricManager) PublishMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) []error {
	return nil
}

func (m *mockMetricManager) ProcessMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) ([]core.Metric, []error) {
	return nil, nil
}
func (m *mockMetricManager) ValidateDeps(mts []core.RequestedMetric, prs []core.SubscribedPlugin, cdt *cdata.ConfigDataTree) []serror.SnapError {
	if m.failValidatingMetrics {
		return []serror.SnapError{
			serror.New(errors.New("metric validation error")),
		}
	}
	return nil
}
func (m *mockMetricManager) SubscribeDeps(taskID string, req []core.RequestedMetric, prs []core.SubscribedPlugin, cft *cdata.ConfigDataTree) []serror.SnapError {
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

type listenToPluginEvent struct {
	pluginLoaded  chan struct{}
	pluginStarted chan struct{}
}

func newListenToPluginEvent() *listenToPluginEvent {
	return &listenToPluginEvent{
		pluginLoaded:  make(chan struct{}),
		pluginStarted: make(chan struct{}),
	}
}

func (l *listenToPluginEvent) HandleGomitEvent(e gomit.Event) {
	switch e.Body.(type) {
	case *control_event.LoadPluginEvent:
		l.pluginLoaded <- struct{}{}
	case *control_event.StartPluginEvent:
		l.pluginStarted <- struct{}{}
	}
}

func TestScheduler(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("NewTask", t, func() {
		c := new(mockMetricManager)
		cfg := GetDefaultConfig()
		s := New(cfg)
		s.SetMetricManager(c)
		w := wmap.NewWorkflowMap()
		// Collection node
		w.CollectNode.AddMetric("/foo/bar", 1)
		w.CollectNode.AddMetric("/foo/baz", 2)
		w.CollectNode.AddConfigItem("/foo/bar", "username", "root")
		w.CollectNode.AddConfigItem("/foo/bar", "port", 8080)
		w.CollectNode.AddConfigItem("/foo/bar", "ratio", 0.32)
		w.CollectNode.AddConfigItem("/foo/bar", "yesorno", true)

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
		pu2 := wmap.NewPublishNode("mock-file", -1)
		pu2.AddConfigItem("color", "brown")
		pu2.AddConfigItem("purpose", 42)

		pr12.Add(pu2)
		pr1.Add(pr12)
		w.CollectNode.Add(pr1)
		w.CollectNode.Add(pu1)

		e := s.Start()
		So(e, ShouldBeNil)
		// create a simple schedule which equals to windowed schedule without start and stop time
		_, te := s.CreateTask(schedule.NewWindowedSchedule(time.Second, nil, nil, 0), w, false)
		So(te.Errors(), ShouldBeEmpty)

		Convey("returns errors when metrics do not validate", func() {
			c.failValidatingMetrics = true
			c.failValidatingMetricsAfter = 1
			// create a simple schedule which equals to windowed schedule without start and stop time
			sch := schedule.NewWindowedSchedule(time.Second, nil, nil, 0)
			_, err := s.CreateTask(sch, w, false)
			So(err, ShouldNotBeNil)
			fmt.Printf("%d", len(err.Errors()))
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, serror.New(errors.New("metric validation error")))

		})
		Convey("returns an error when scheduler started and MetricManager is not set", func() {
			s1 := New(GetDefaultConfig())
			err := s1.Start()
			So(err, ShouldNotBeNil)
			fmt.Printf("%v", err)
			So(err, ShouldResemble, ErrMetricManagerNotSet)

		})
		Convey("returns an error when a schedule does not validate", func() {
			s1 := New(GetDefaultConfig())
			s1.Start()
			sch := schedule.NewWindowedSchedule(time.Second, nil, nil, 0)
			_, err := s1.CreateTask(sch, w, false)
			So(err, ShouldNotBeNil)
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, serror.New(ErrSchedulerNotStarted))
			s1.metricManager = c
			s1.Start()
			_, err1 := s1.CreateTask(schedule.NewWindowedSchedule(time.Second*0, nil, nil, 0), w, false)
			So(err1.Errors()[0].Error(), ShouldResemble, "Interval must be greater than 0")

		})
		Convey("create a task", func() {
			sch := schedule.NewWindowedSchedule(time.Second*5, nil, nil, 0)
			tsk, err := s.CreateTask(sch, w, false)
			So(len(err.Errors()), ShouldEqual, 0)
			So(tsk, ShouldNotBeNil)
			So(tsk.(*task).deadlineDuration, ShouldResemble, DefaultDeadlineDuration)
			So(len(s.GetTasks()), ShouldEqual, 2)
			Convey("error when attempting to add duplicate task", func() {
				err := s.tasks.add(tsk.(*task))
				So(err, ShouldNotBeNil)

			})
			Convey("get created task", func() {
				t, err := s.GetTask(tsk.ID())
				So(err, ShouldBeNil)
				So(t, ShouldEqual, tsk)
			})
			Convey("error when attempting to get a task that doesn't exist", func() {
				t, err := s.GetTask("1234")
				So(err, ShouldNotBeNil)
				So(t, ShouldBeNil)
			})
			Convey("stop a stopped task", func() {
				err := s.StopTask(tsk.ID())
				So(len(err), ShouldEqual, 1)
				So(err[0].Error(), ShouldEqual, "Task is already stopped.")
			})
		})
		Convey("returns a task with a 6 second deadline duration", func() {
			sch := schedule.NewWindowedSchedule(6*time.Second, nil, nil, 0)
			tsk, err := s.CreateTask(sch, w, false, core.TaskDeadlineDuration(6*time.Second))
			So(len(err.Errors()), ShouldEqual, 0)
			So(tsk.(*task).deadlineDuration, ShouldResemble, time.Duration(6*time.Second))
			prev := tsk.(*task).Option(core.TaskDeadlineDuration(1 * time.Second))
			So(tsk.(*task).deadlineDuration, ShouldResemble, time.Duration(1*time.Second))
			tsk.(*task).Option(prev)
			So(tsk.(*task).deadlineDuration, ShouldResemble, time.Duration(6*time.Second))
		})
		Convey("returns a task with a 1m collectDuration", func() {
			tsk, err := s.CreateTask(schedule.NewStreamingSchedule(), w, false, core.SetMaxCollectDuration(time.Minute))
			So(len(err.Errors()), ShouldEqual, 0)
			So(tsk.(*task).maxCollectDuration, ShouldResemble, time.Duration(time.Minute))
		})
		Convey("Returns a task with a metric buffer of 100", func() {
			tsk, err := s.CreateTask(schedule.NewStreamingSchedule(), w, false, core.SetMaxMetricsBuffer(100))
			So(len(err.Errors()), ShouldEqual, 0)
			So(tsk.(*task).maxMetricsBuffer, ShouldEqual, 100)
		})
	})
	Convey("Stop()", t, func() {
		Convey("Should set scheduler state to SchedulerStopped", func() {
			scheduler := New(GetDefaultConfig())
			c := new(mockMetricManager)
			scheduler.metricManager = c
			scheduler.Start()
			scheduler.Stop()
			So(scheduler.state, ShouldEqual, schedulerStopped)
		})
	})
	Convey("SetMetricManager()", t, func() {
		Convey("Should set metricManager for scheduler", func() {
			scheduler := New(GetDefaultConfig())
			c := new(mockMetricManager)
			scheduler.SetMetricManager(c)
			So(scheduler.metricManager, ShouldEqual, c)
		})
	})

}
