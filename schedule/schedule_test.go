package schedule

import (
	"errors"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

type MockMetricManager struct {
	failValidatingMetrics      bool
	failValidatingMetricsAfter int
	failuredSoFar              int
}

func (m *MockMetricManager) SubscribeMetricType(mt core.MetricType, cd *cdata.ConfigDataNode) (core.MetricType, []error) {
	if m.failValidatingMetrics {
		if m.failValidatingMetricsAfter > m.failuredSoFar {
			m.failuredSoFar++
			return nil, nil
		}
		return nil, []error{
			errors.New("metric validation error"),
		}
	}
	return nil, nil
}

func (m *MockMetricManager) UnsubscribeMetricType(mt core.MetricType) {

}

type MockMetricManagerError struct {
	errs []error
}

type MockMetricType struct {
	version            int
	namespace          []string
	lastAdvertisedTime time.Time
	config             *cdata.ConfigDataNode
}

func (m MockMetricType) Version() int {
	return m.version
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return m.config
}

type mockWorkflow struct {
	state workflowState
}

func (w *mockWorkflow) Start(t *task) {
	w.state = WorkflowStarted
	time.Sleep(15 * time.Millisecond)
}

func (w *mockWorkflow) State() workflowState {
	return w.state
}

func TestScheduler(t *testing.T) {
	Convey("new", t, func() {
		c := new(MockMetricManager)
		mockSchedule := &MockSchedule{
			tick: false,
			failValidatingSchedule: false,
		}
		mt := []core.MetricType{
			&MockMetricType{
				namespace:          []string{"foo", "bar"},
				version:            1,
				lastAdvertisedTime: time.Now(),
			},
			&MockMetricType{
				namespace:          []string{"foo2", "bar2"},
				version:            1,
				lastAdvertisedTime: time.Now(),
			},
			&MockMetricType{
				namespace:          []string{"foo2", "bar2"},
				version:            1,
				lastAdvertisedTime: time.Now(),
			},
		}
		scheduler := New(1, 5)
		cdt := cdata.NewTree()
		cd := cdata.NewNode()
		cd.AddItem("foo", ctypes.ConfigValueInt{Value: 1})
		cdt.Add([]string{"foo", "bar"}, cd)
		mockWF := new(mockWorkflow)

		Convey("returns errors when metrics do not validate", func() {
			c.failValidatingMetrics = true
			c.failValidatingMetricsAfter = 2
			scheduler := New(1, 5)
			scheduler.metricManager = c
			scheduler.Start()
			mockSchedule := &MockSchedule{
				tick: false,
				failValidatingSchedule: false,
			}
			_, err := scheduler.CreateTask(mt, mockSchedule, cdt, mockWF)
			So(err, ShouldNotBeNil)
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, errors.New("metric validation error"))

		})

		Convey("returns an error when scheduler started and MetricManager is not set", func() {
			err := scheduler.Start()
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, MetricManagerNotSet)

		})

		Convey("returns an error when a schedule does not validate", func() {
			mockSchedule.failValidatingSchedule = true
			_, err := scheduler.CreateTask(mt, mockSchedule, cdt, mockWF)
			So(err, ShouldNotBeNil)
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, SchedulerNotStarted)
			scheduler.metricManager = c
			scheduler.Start()
			_, err = scheduler.CreateTask(mt, mockSchedule, cdt, mockWF)
			So(err.Errors()[0], ShouldResemble, errors.New("schedule error"))

		})

		Convey("create a task", func() {
			scheduler.metricManager = c
			scheduler.Start()
			task, err := scheduler.CreateTask(mt, mockSchedule, cdt, mockWF)
			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			So(task.deadlineDuration, ShouldResemble, DefaultDeadlineDuration)
			So(len(scheduler.GetTasks()), ShouldEqual, 1)
			Convey("error when attempting to add duplicate task", func() {
				err := scheduler.tasks.add(task)
				So(err, ShouldNotBeNil)
			})
			Convey("get created task", func() {
				t, err := scheduler.GetTask(task.Id())
				So(err, ShouldBeNil)
				So(t, ShouldEqual, task)
			})
			Convey("error when attempting to get a task that doesn't exist", func() {
				t, err := scheduler.GetTask(uint64(1234))
				So(err, ShouldNotBeNil)
				So(t, ShouldBeNil)
			})
		})

		Convey("returns a task with a 6 second deadline duration", func() {
			scheduler.metricManager = c
			scheduler.Start()
			task, err := scheduler.CreateTask(mt, mockSchedule, cdt, mockWF, taskDeadlineDuration(6*time.Second))
			So(err, ShouldBeNil)
			So(task.deadlineDuration, ShouldResemble, time.Duration(6*time.Second))
			prev := task.option(taskDeadlineDuration(1 * time.Second))
			So(task.deadlineDuration, ShouldResemble, time.Duration(1*time.Second))
			task.option(prev)
			So(task.deadlineDuration, ShouldResemble, time.Duration(6*time.Second))
		})

	})
	Convey("Stop()", t, func() {
		Convey("Should set scheduler state to SchedulerStopped", func() {
			scheduler := New(1, 5)
			c := new(MockMetricManager)
			scheduler.metricManager = c
			scheduler.Start()
			scheduler.Stop()
			So(scheduler.state, ShouldEqual, SchedulerStopped)
		})
	})
	Convey("SetMetricManager()", t, func() {
		Convey("Should set metricManager for scheduler", func() {
			scheduler := New(1, 5)
			c := new(MockMetricManager)
			scheduler.SetMetricManager(c)
			So(scheduler.metricManager, ShouldEqual, c)
		})
	})
}
