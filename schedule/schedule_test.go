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
		scheduler := New(c)

		Convey("returns errors when metrics do not validate", func() {
			c.failValidatingMetrics = true
			c.failValidatingMetricsAfter = 2
			scheduler := New(c)
			mockSchedule := &MockSchedule{
				tick: false,
				failValidatingSchedule: false,
			}
			cdt := cdata.NewTree()
			cd := cdata.NewNode()
			cd.AddItem("foo", ctypes.ConfigValueInt{Value: 1})
			cdt.Add([]string{"foo", "bar"}, cd)
			_, err := scheduler.CreateTask(mt, mockSchedule, cdt)
			So(err, ShouldNotBeNil)
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, errors.New("metric validation error"))

		})

		Convey("returns an error when a schedule does not validate", func() {
			mockSchedule.failValidatingSchedule = true
			_, err := scheduler.CreateTask(nil, mockSchedule, nil)
			So(err, ShouldNotBeNil)
			So(len(err.Errors()), ShouldBeGreaterThan, 0)
			So(err.Errors()[0], ShouldResemble, errors.New("schedule error"))
		})

		Convey("returns an a task", func() {
			task, err := scheduler.CreateTask(nil, mockSchedule, nil)
			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
		})

	})
}
