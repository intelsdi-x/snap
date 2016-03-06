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
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/pkg/promise"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	SnapPath                     = os.Getenv("SNAP_PATH")
	snap_collector_mock2_path    = path.Join(SnapPath, "plugin", "snap-collector-mock2")
	snap_processor_passthru_path = path.Join(SnapPath, "plugin", "snap-processor-passthru")
	snap_publisher_file_path     = path.Join(SnapPath, "plugin", "snap-publisher-file")
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
			rp, err := core.NewRequestedPlugin(snap_collector_mock2_path)
			So(err, ShouldBeNil)
			_, err = c.Load(rp)
			So(err, ShouldBeNil)
			rp2, err := core.NewRequestedPlugin(snap_publisher_file_path)
			So(err, ShouldBeNil)
			_, err = c.Load(rp2)
			So(err, ShouldBeNil)
			rp3, err := core.NewRequestedPlugin(snap_processor_passthru_path)
			So(err, ShouldBeNil)
			_, err = c.Load(rp3)
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)

			metrics, err2 := c.MetricCatalog()
			So(err2, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			w := wmap.NewWorkflowMap()
			w.CollectNode.AddMetric("/intel/mock/foo", 2)
			w.CollectNode.AddConfigItem("/intel/mock/foo", "password", "secret")

			pu := wmap.NewPublishNode("file", 3)
			pu.AddConfigItem("file", "/tmp/snap-TestCollectPublishWorkflow.out")

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

// The mocks below are here for testing work submission
type Mock1 struct {
	count      int
	errorIndex int
	delay      time.Duration
	queue      map[string]int
}

func (m *Mock1) CollectMetrics([]core.Metric, time.Time, string) ([]core.Metric, []error) {
	return nil, nil
}

func (m *Mock1) Work(j job) queuedJob {
	m.queue[j.TypeString()]++
	return m
}

func (m *Mock1) Promise() promise.Promise {
	return m
}

func (m *Mock1) Await() []error {
	time.Sleep(m.delay)
	m.count++
	if m.count == m.errorIndex {
		return []error{errors.New("I am an error")}
	}
	return nil
}

func (m *Mock1) AwaitUntil(time.Duration) []error {
	return nil
}

func (m *Mock1) Complete([]error) {

}

func (m *Mock1) IsComplete() bool {
	return false
}

func (m *Mock1) Job() job {
	return nil
}

func (m *Mock1) AndThen(_ func([]error)) {
}

func (m *Mock1) AndThenUntil(_ time.Duration, _ func([]error)) {
}

func TestWorkJobs(t *testing.T) {
	Convey("Test speed and concurrency of TestWorkJobs\n", t, func() {
		Convey("submit multiple jobs\n", func() {
			m := &Mock1{queue: make(map[string]int)}
			m.delay = time.Millisecond * 100
			pj := newCollectorJob(nil, time.Second*1, m, nil, "")
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m, id: "1", name: "mock"}
			for x := 0; x < 3; x++ {
				n := cdata.NewNode()
				pr := &processNode{config: n, name: fmt.Sprintf("prjob%d", counter)}
				pu := &publishNode{config: n, name: fmt.Sprintf("pujob%d", counter)}
				counter++
				prs = append(prs, pr)
				pus = append(pus, pu)
			}
			workJobs(prs, pus, t, pj)
			So(m.queue["processor"], ShouldEqual, 3)
			So(m.queue["publisher"], ShouldEqual, 3)
			So(t.failedRuns, ShouldEqual, 0)
		})
		Convey("submit multiple jobs with nesting", func() {
			m := &Mock1{queue: make(map[string]int)}
			m.delay = time.Millisecond * 100
			pj := newCollectorJob(nil, time.Second*1, m, nil, "")
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m, id: "1", name: "mock"}
			// 3 proc + 3 pub
			for x := 0; x < 3; x++ {
				n := cdata.NewNode()
				pr := &processNode{config: n, name: fmt.Sprintf("prjob%d", counter)}
				pu := &publishNode{config: n, name: fmt.Sprintf("pujob%d", counter)}
				counter++
				prs = append(prs, pr)
				pus = append(pus, pu)
			}
			// 3 proc => 3 proc + 3 pub
			for _, pr := range prs {
				cprs := make([]*processNode, 0)
				cpus := make([]*publishNode, 0)
				for x := 0; x < 3; x++ {
					n := cdata.NewNode()
					cpr := &processNode{config: n, name: fmt.Sprintf("prjobchild%d", counter)}
					cpu := &publishNode{config: n, name: fmt.Sprintf("pujobchild%d", counter)}
					counter++
					cprs = append(cprs, cpr)
					cpus = append(cpus, cpu)
				}
				pr.ProcessNodes = cprs
				pr.PublishNodes = cpus
			}
			workJobs(prs, pus, t, pj)
			// (3*3)+3
			So(m.queue["processor"], ShouldEqual, 12)
			// (3*3)
			So(m.queue["publisher"], ShouldEqual, 12)
			So(t.failedRuns, ShouldEqual, 0)
		})
		Convey("submit multiple jobs where one has an error", func() {
			m := &Mock1{queue: make(map[string]int)}
			// make the 13th job fail
			m.errorIndex = 13
			m.delay = time.Millisecond * 100
			pj := newCollectorJob(nil, time.Second*1, m, nil, "")
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m, id: "1", name: "mock"}
			// 3 proc + 3 pub
			for x := 0; x < 3; x++ {
				n := cdata.NewNode()
				pr := &processNode{config: n, name: fmt.Sprintf("prjob%d", counter)}
				pu := &publishNode{config: n, name: fmt.Sprintf("pujob%d", counter)}
				counter++
				prs = append(prs, pr)
				pus = append(pus, pu)
			}
			// 3 proc => 3 proc + 3 pub
			for _, pr := range prs {
				cprs := make([]*processNode, 0)
				cpus := make([]*publishNode, 0)
				for x := 0; x < 3; x++ {
					n := cdata.NewNode()
					cpr := &processNode{config: n, name: fmt.Sprintf("prjobchild%d", counter)}
					cpu := &publishNode{config: n, name: fmt.Sprintf("pujobchild%d", counter)}
					counter++
					cprs = append(cprs, cpr)
					cpus = append(cpus, cpu)
				}
				pr.ProcessNodes = cprs
				pr.PublishNodes = cpus
			}
			workJobs(prs, pus, t, pj)
			// (3*3)+3
			So(m.queue["processor"], ShouldEqual, 12)
			// (3*3)
			So(m.queue["publisher"], ShouldEqual, 12)
			So(t.failedRuns, ShouldEqual, 1)
			So(t.lastFailureMessage, ShouldEqual, "I am an error")
		})

	})
}
