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
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/promise"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/plugin/helper"
	"github.com/intelsdi-x/snap/scheduler/wmap"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	SnapPath                     = helper.BuildPath
	snap_collector_mock1_path    = helper.PluginFilePath("snap-plugin-collector-mock1")
	snap_collector_mock2_path    = helper.PluginFilePath("snap-plugin-collector-mock2")
	snap_processor_passthru_path = helper.PluginFilePath("snap-plugin-processor-passthru")
	snap_publisher_file_path     = helper.PluginFilePath("snap-plugin-publisher-mock-file")

	metricsToCollect = 3
)

type MockMetricType struct {
	namespace []string
}

type mockPluginEvent struct {
	LoadedPluginName      string
	LoadedPluginVersion   int
	UnloadedPluginName    string
	UnloadedPluginVersion int
	PluginType            int
	EventNamespace        string
}

type eventListener struct {
	plugin                    *mockPluginEvent
	metricCollectCount        int
	MetricCollectFailureCount int
	metricsCollectionDone     chan bool
	done                      chan struct{}
}

func newEventListener() *eventListener {
	return &eventListener{
		done: make(chan struct{}),
	}
}

func (l *eventListener) HandleGomitEvent(e gomit.Event) {
	go func() {
		switch e.Body.(type) {
		case *control_event.LoadPluginEvent:
			l.done <- struct{}{}
		case *scheduler_event.MetricCollectedEvent:
			if l.metricCollectCount > metricsToCollect {
				l.done <- struct{}{}
			}
			l.metricCollectCount++
		case *scheduler_event.MetricCollectionFailedEvent:
			if l.MetricCollectFailureCount > 1 {
				l.done <- struct{}{}
			}
			l.MetricCollectFailureCount++
		default:
		}
	}()
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
		ccfg := control.GetDefaultConfig()
		ccfg.Plugins.Collector.Plugins = control.NewPluginsConfig()
		ccfg.Plugins.Collector.Plugins["mock"] = control.NewPluginConfigItem()
		ccfg.Plugins.Collector.Plugins["mock"].Versions = map[int]*cdata.ConfigDataNode{}
		ccfg.Plugins.Collector.Plugins["mock"].Versions[1] = cdata.NewNode()
		ccfg.Plugins.Collector.Plugins["mock"].Versions[1].AddItem("test", ctypes.ConfigValueBool{Value: true})
		c := control.New(ccfg)
		c.Start()
		cfg := GetDefaultConfig()
		s := New(cfg)
		s.SetMetricManager(c)
		Convey("create a workflow", func() {
			rp, err := core.NewRequestedPlugin(snap_collector_mock2_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp)
			So(err, ShouldBeNil)
			rp2, err := core.NewRequestedPlugin(snap_publisher_file_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			plugPublisher, err := c.Load(rp2)
			So(err, ShouldBeNil)
			rp3, err := core.NewRequestedPlugin(snap_processor_passthru_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp3)
			So(err, ShouldBeNil)
			rp4, err := core.NewRequestedPlugin(snap_collector_mock1_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp4)
			So(err, ShouldBeNil)

			metrics, err2 := c.MetricCatalog()
			So(err2, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			// The following two metrics will result in both versions (1 and 2) of
			// the mock plugin to be used.  '/intel/mock/test' will be coming from
			// mock version 1 due to the global config above.
			w := wmap.NewWorkflowMap()
			w.Collect.AddMetric("/intel/mock/foo", -1)
			w.Collect.AddMetric("/intel/mock/test", -1)
			w.Collect.AddConfigItem("/intel/mock/foo", "password", "secret")

			pu := wmap.NewPublishNode("mock-file", 3)
			pu.AddConfigItem("file", "/tmp/snap-TestCollectPublishWorkflow.out")
			pr := wmap.NewProcessNode("passthru", 1)

			pr.Add(pu)
			w.Collect.Add(pr)

			Convey("Start scheduler", func() {
				err := s.Start()
				So(err, ShouldBeNil)
				Convey("Create and start task", func() {
					el := newEventListener()
					s.RegisterEventHandler("TestCollectPublishWorkflow", el)
					// create a simple schedule which equals to windowed schedule
					// without start and stop time
					sch := schedule.NewWindowedSchedule(time.Millisecond*200, nil, nil, 0)
					t, err := s.CreateTask(sch, w, true)
					So(err.Errors(), ShouldBeEmpty)
					So(t, ShouldNotBeNil)
					<-el.done
					So(t.LastFailureMessage(), ShouldBeEmpty)
					So(t.FailedCount(), ShouldEqual, 0)
					So(t.HitCount(), ShouldBeGreaterThan, metricsToCollect)

					// check if task fails after unloading publisher
					c.Unload(plugPublisher)
					<-el.done
					So(t.LastFailureMessage(), ShouldNotBeEmpty)
					So(t.FailedCount(), ShouldBeGreaterThan, 0)
				})
			})
		})
	})
}

func TestProcessChainingWorkflow(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("Given a started plugin control", t, func() {
		c := control.New(control.GetDefaultConfig())
		c.Start()
		cfg := GetDefaultConfig()
		s := New(cfg)
		s.SetMetricManager(c)
		Convey("create a workflow with chained processors", func() {
			lpe := newEventListener()
			c.RegisterEventHandler("Control.PluginLoaded", lpe)
			rp, err := core.NewRequestedPlugin(snap_collector_mock2_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp)
			So(err, ShouldBeNil)
			<-lpe.done
			rp2, err := core.NewRequestedPlugin(snap_publisher_file_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp2)
			So(err, ShouldBeNil)
			<-lpe.done
			rp3, err := core.NewRequestedPlugin(snap_processor_passthru_path, c.GetTempDir(), nil)
			So(err, ShouldBeNil)
			_, err = c.Load(rp3)
			So(err, ShouldBeNil)
			<-lpe.done

			metrics, err2 := c.MetricCatalog()
			So(err2, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			w := wmap.NewWorkflowMap()
			w.Collect.AddMetric("/intel/mock/foo", 2)
			w.Collect.AddConfigItem("/intel/mock/foo", "password", "secret")

			pu := wmap.NewPublishNode("mock-file", 3)
			pu.AddConfigItem("file", "/tmp/snap-TestCollectPublishWorkflow.out")

			pr1 := wmap.NewProcessNode("passthru", 1)

			pr2 := wmap.NewProcessNode("passthru", 1)

			pr2.Add(pu)
			pr1.Add(pr2)
			w.Collect.Add(pr1)

			Convey("Start scheduler", func() {
				err := s.Start()
				So(err, ShouldBeNil)
				Convey("Create task", func() {
					// create a simple schedule which equals to windowed schedule
					// without start and stop time
					sch := schedule.NewWindowedSchedule(time.Millisecond*200, nil, nil, 0)
					t, err := s.CreateTask(sch, w, true)
					s.RegisterEventHandler("TestProcessChainingWorkflow", lpe)
					So(err.Errors(), ShouldBeEmpty)
					So(t, ShouldNotBeNil)
					<-lpe.done
					So(t.LastFailureMessage(), ShouldBeEmpty)
					So(t.FailedCount(), ShouldEqual, 0)
					So(t.HitCount(), ShouldBeGreaterThan, metricsToCollect)
				})
			})
		})
	})
}

// The mocks below are here for testing work submission
type Mock1 struct {
	sync.Mutex
	count      int
	errorIndex int
	queue      map[string]int
}

func (m *Mock1) CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error) {
	return nil, nil
}

func (m *Mock1) ExpandWildcards(core.Namespace) ([]core.Namespace, serror.SnapError) {
	return nil, nil
}

func (m *Mock1) Work(j job) queuedJob {
	m.Lock()
	defer m.Unlock()
	m.queue[j.TypeString()]++
	return m
}

func (m *Mock1) Promise() promise.Promise {
	return m
}

func (m *Mock1) Await() []error {
	m.Lock()
	defer m.Unlock()
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

func (m *Mock1) IsError() bool {
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
	// log.SetLevel(log.DebugLevel)
	Convey("Test speed and concurrency of TestWorkJobs\n", t, func() {
		Convey("submit multiple jobs\n", func() {
			m1 := &Mock1{queue: make(map[string]int)}
			pj := newCollectorJob(nil, time.Second*1, m1, nil, "", nil)
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m1, id: "1", name: "mock"}
			for x := 0; x < 3; x++ {
				n := cdata.NewNode()
				pr := &processNode{config: n, name: fmt.Sprintf("prjob%d", counter)}
				pu := &publishNode{config: n, name: fmt.Sprintf("pujob%d", counter)}
				counter++
				prs = append(prs, pr)
				pus = append(pus, pu)
			}
			workJobs(prs, pus, t, pj)
			So(t.failedRuns, ShouldEqual, 0)
			So(m1.queue["processor"], ShouldEqual, 3)
			So(m1.queue["publisher"], ShouldEqual, 3)
		})
		Convey("submit multiple jobs with nesting", func() {
			m2 := &Mock1{queue: make(map[string]int)}
			pj := newCollectorJob(nil, time.Second*1, m2, nil, "", nil)
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m2, id: "1", name: "mock"}
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
			So(t.failedRuns, ShouldEqual, 0)
			// (3*3)+3
			So(m2.queue["processor"], ShouldEqual, 12)
			// (3*3)
			So(m2.queue["publisher"], ShouldEqual, 12)

		})
		Convey("submit multiple jobs where one has an error", func() {
			m3 := &Mock1{queue: make(map[string]int)}
			// make the 13th job fail
			m3.errorIndex = 13
			pj := newCollectorJob(nil, time.Second*1, m3, nil, "", nil)
			prs := make([]*processNode, 0)
			pus := make([]*publishNode, 0)
			counter := 0
			t := &task{manager: m3, id: "1", name: "mock"}
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
			So(t.failedRuns, ShouldEqual, 1)
			So(t.lastFailureMessage, ShouldEqual, "I am an error")
			// (3*3)+3
			So(m3.queue["processor"], ShouldEqual, 12)
			// (3*3)
			So(m3.queue["publisher"], ShouldEqual, 12)
		})

	})
}
