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
	"fmt"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type mockMetricManager struct {
	failValidatingMetrics      bool
	failValidatingMetricsAfter int
	failuredSoFar              int
	acceptedContentTypes       map[string][]string
	returnedContentTypes       map[string][]string
	autodiscoverPaths          []string
}

func (m *mockMetricManager) lazyContentType(key string) {
	if m.acceptedContentTypes == nil {
		m.acceptedContentTypes = make(map[string][]string)
	}
	if m.returnedContentTypes == nil {
		m.returnedContentTypes = make(map[string][]string)
	}
	if m.acceptedContentTypes[key] == nil {
		m.acceptedContentTypes[key] = []string{"snap.gob"}
	}
	if m.returnedContentTypes[key] == nil {
		m.returnedContentTypes[key] = []string{}
	}
}

// Used to mock type from plugin
func (m *mockMetricManager) setAcceptedContentType(n string, t core.PluginType, v int, s []string) {
	key := fmt.Sprintf("%s:%d:%d", n, t, v)
	m.lazyContentType(key)
	m.acceptedContentTypes[key] = s
}

func (m *mockMetricManager) setReturnedContentType(n string, t core.PluginType, v int, s []string) {
	key := fmt.Sprintf("%s:%d:%d", n, t, v)
	m.lazyContentType(key)
	m.returnedContentTypes[key] = s
}

func (m *mockMetricManager) GetPluginContentTypes(n string, t core.PluginType, v int) ([]string, []string, error) {
	key := fmt.Sprintf("%s:%d:%d", n, t, v)
	m.lazyContentType(key)

	return m.acceptedContentTypes[key], m.returnedContentTypes[key], nil
}

func (m *mockMetricManager) CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error) {
	return nil, nil
}

func (m *mockMetricManager) PublishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error {
	return nil
}

func (m *mockMetricManager) ProcessMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) (string, []byte, []error) {
	return "", nil, nil
}

func (m *mockMetricManager) ValidateDeps(mts []core.RequestedMetric, prs []core.SubscribedPlugin, ctree *cdata.ConfigDataTree) []serror.SnapError {
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

// Helper constructor functions for resuse amongst tests
func newMockMetricManager() *mockMetricManager {
	m := new(mockMetricManager)
	m.setAcceptedContentType("machine", core.ProcessorPluginType, 1, []string{"snap.*", "snap.gob", "foo.bar"})
	m.setReturnedContentType("machine", core.ProcessorPluginType, 1, []string{"snap.gob"})
	m.setAcceptedContentType("rmq", core.PublisherPluginType, -1, []string{"snap.json", "snap.gob"})
	m.setAcceptedContentType("file", core.PublisherPluginType, -1, []string{"snap.json"})
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
	pu2 := wmap.NewPublishNode("file", -1)
	pu2.AddConfigItem("color", "brown")
	pu2.AddConfigItem("purpose", 42)

	pr12.Add(pu2)
	pr1.Add(pr12)
	w.CollectNode.Add(pr1)
	w.CollectNode.Add(pu1)
	return w
}

// ----------------------------- Medium Tests ----------------------------
func TestStopTask(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	s := newScheduler()
	s.Start()
	w := newMockWorkflowMap()
	tsk, _ := s.CreateTask(schedule.NewSimpleSchedule(time.Millisecond*100), w, false)
	task := s.tasks.Get(tsk.ID())
	task.Spin()
	err := s.StopTask(tsk.ID())

	Convey("Calling StopTask a running task", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		time.Sleep(100 * time.Millisecond)
		Convey("State of the task should be TaskStopped", func() {
			So(task.state, ShouldEqual, core.TaskStopped)
		})
	})

	tskStopped, _ := s.CreateTask(schedule.NewSimpleSchedule(time.Millisecond*100), w, false)
	err = s.StopTask(tskStopped.ID())
	Convey("Calling StopTask on a stopped task", t, func() {
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is already stopped.", func() {
			So(err[0].Error(), ShouldResemble, "Task is already stopped.")
		})
	})

	tskDisabled, _ := s.CreateTask(schedule.NewSimpleSchedule(time.Millisecond*100), w, false)
	taskDisabled := s.tasks.Get(tskDisabled.ID())
	taskDisabled.state = core.TaskDisabled
	err = s.StopTask(tskDisabled.ID())
	Convey("Calling StopTask on a disabled task", t, func() {
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should read: Task is disabled. Only running tasks can be stopped.", func() {
			So(err[0].Error(), ShouldResemble, "Task is disabled. Only running tasks can be stopped.")
		})
	})

	s.Stop()
}
