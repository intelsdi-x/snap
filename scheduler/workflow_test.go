package scheduler

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PulsePath = os.Getenv("PULSE_PATH")
)

type MockMetricType struct {
	namespace []string
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return 1
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return nil
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func TestCollectPublishWorkflow(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("Given a started plugin control", t, func() {

		c := control.New()
		c.Start()
		s := New()
		s.SetMetricManager(c)
		Convey("create a workflow", func() {
			_, err := c.Load(path.Join(PulsePath, "plugin", "pulse-collector-dummy2"))
			So(err, ShouldBeNil)
			_, err = c.Load(path.Join(PulsePath, "plugin", "pulse-publisher-file"))
			So(err, ShouldBeNil)
			_, err = c.Load(path.Join(PulsePath, "plugin", "pulse-processor-passthru"))
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)

			metrics, err2 := c.MetricCatalog()
			So(err2, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			w := wmap.NewWorkflowMap()
			w.CollectNode.AddMetric("/intel/dummy/foo", 2)
			w.CollectNode.AddConfigItem("/intel/dummy/foo", "password", "secret")

			pu := wmap.NewPublishNode("file", 1)
			pu.AddConfigItem("file", "/tmp/pulse-TestCollectPublishWorkflow.out")

			pr := wmap.NewProcessNode("passthru", 1)
			time.Sleep(100 * time.Millisecond)

			pr.Add(pu)
			w.CollectNode.Add(pr)

			Convey("Start scheduler", func() {
				err := s.Start()
				So(err, ShouldBeNil)
				Convey("Create task", func() {
					t, err := s.CreateTask(schedule.NewSimpleSchedule(time.Millisecond*500), w, false)
					So(err.Errors(), ShouldBeEmpty)
					So(t, ShouldNotBeNil)
					t.(*task).Spin()
					time.Sleep(3 * time.Second)

				})
			})
		})
	})
}
