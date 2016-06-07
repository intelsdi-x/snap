// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	"net"
	"path"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginPath = path.Join(SnapPath, "plugin")
)

func TestDistributedWorkflow(t *testing.T) {
	Convey("Create a scheduler with 2 controls and load plugins", t, func() {
		l, _ := net.Listen("tcp", ":0")
		l.Close()
		cfg := control.GetDefaultConfig()
		cfg.ListenPort = l.Addr().(*net.TCPAddr).Port
		c1 := control.New(cfg)
		c1.Start()
		m, _ := net.Listen("tcp", ":0")
		m.Close()
		cfg.ListenPort = m.Addr().(*net.TCPAddr).Port
		port1 := cfg.ListenPort
		c2 := control.New(cfg)
		schcfg := GetDefaultConfig()
		sch := New(schcfg)
		c2.Start()
		sch.SetMetricManager(c1)
		err := sch.Start()
		So(err, ShouldBeNil)
		// Load appropriate plugins into each control.
		mock2Path := path.Join(PluginPath, "snap-collector-mock2")
		passthruPath := path.Join(PluginPath, "snap-processor-passthru")
		filePath := path.Join(PluginPath, "snap-publisher-file")

		// mock2 and file onto c1

		rp, err := core.NewRequestedPlugin(mock2Path)
		So(err, ShouldBeNil)
		_, err = c1.Load(rp)
		So(err, ShouldBeNil)
		rp, err = core.NewRequestedPlugin(filePath)
		So(err, ShouldBeNil)
		_, err = c1.Load(rp)
		So(err, ShouldBeNil)
		// passthru on c2
		rp, err = core.NewRequestedPlugin(passthruPath)
		So(err, ShouldBeNil)
		passthru, err := c2.Load(rp)
		So(err, ShouldBeNil)

		Convey("Test task with one local and one remote node", func() {
			//Create a task
			//Create a workflowmap
			wf := dsWFMap(port1)
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, true)
			So(len(errs.Errors()), ShouldEqual, 0)
			So(t, ShouldNotBeNil)
		})

		Convey("Test task with invalid remote port", func() {
			wf := dsWFMap(0)
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, true)
			So(len(errs.Errors()), ShouldEqual, 1)
			So(t, ShouldBeNil)
		})

		Convey("Test task without remote plugin", func() {
			_, err := c2.Unload(passthru)
			So(err, ShouldBeNil)
			wf := dsWFMap(port1)
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, true)
			So(len(errs.Errors()), ShouldEqual, 1)
			So(t, ShouldBeNil)
		})

		Convey("Test task failing when control is stopped while task is running", func() {
			wf := dsWFMap(port1)
			interval := time.Millisecond * 100
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(interval), wf, true)
			So(len(errs.Errors()), ShouldEqual, 0)
			So(t, ShouldNotBeNil)
			c2.Stop()
			// Give task time to fail
			time.Sleep(time.Second)
			tasks := sch.GetTasks()
			var task core.Task
			for _, v := range tasks {
				task = v
			}
			So(task.State(), ShouldEqual, core.TaskDisabled)
		})

	})

}

func TestDistributedSubscriptions(t *testing.T) {

	Convey("Load control/scheduler with a mock remote scheduler", t, func() {
		l, _ := net.Listen("tcp", ":0")
		l.Close()
		cfg := control.GetDefaultConfig()
		cfg.ListenPort = l.Addr().(*net.TCPAddr).Port
		c1 := control.New(cfg)
		c1.Start()
		m, _ := net.Listen("tcp", ":0")
		m.Close()
		cfg.ListenPort = m.Addr().(*net.TCPAddr).Port
		port1 := cfg.ListenPort
		c2 := control.New(cfg)
		schcfg := GetDefaultConfig()
		sch := New(schcfg)
		c2.Start()
		sch.SetMetricManager(c1)
		err := sch.Start()
		So(err, ShouldBeNil)
		// Load appropriate plugins into each control.
		mock2Path := path.Join(PluginPath, "snap-collector-mock2")
		passthruPath := path.Join(PluginPath, "snap-processor-passthru")
		filePath := path.Join(PluginPath, "snap-publisher-file")

		// mock2 and file onto c1

		rp, err := core.NewRequestedPlugin(mock2Path)
		So(err, ShouldBeNil)
		_, err = c1.Load(rp)
		So(err, ShouldBeNil)
		rp, err = core.NewRequestedPlugin(filePath)
		So(err, ShouldBeNil)
		_, err = c1.Load(rp)
		So(err, ShouldBeNil)
		// passthru on c2
		rp, err = core.NewRequestedPlugin(passthruPath)
		So(err, ShouldBeNil)
		_, err = c2.Load(rp)
		So(err, ShouldBeNil)

		Convey("Starting task should not succeed if remote dep fails to subscribe", func() {
			//Create a task
			//Create a workflowmap
			wf := dsWFMap(port1)
			// Create a task that is not started immediately so we can
			// validate deps correctly.
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, false)
			So(len(errs.Errors()), ShouldEqual, 0)
			So(t, ShouldNotBeNil)
			schTask := t.(*task)
			remoteMockManager := &subscriptionManager{Fail: true}
			schTask.RemoteManagers.Add(fmt.Sprintf("127.0.0.1:%v", port1), remoteMockManager)
			localMockManager := &subscriptionManager{Fail: false}
			schTask.RemoteManagers.Add("", localMockManager)
			// Start task. We expect it to fail while subscribing deps
			terrs := sch.StartTask(t.ID())
			So(terrs, ShouldNotBeNil)
			Convey("So dependencies should have been unsubscribed", func() {
				// Ensure that unsubscribe call count is equal to subscribe call count
				// i.e that every subscribe call was followed by an unsubscribe since
				// we errored
				So(remoteMockManager.UnsubscribeCallCount, ShouldEqual, remoteMockManager.SubscribeCallCount)
				So(localMockManager.UnsubscribeCallCount, ShouldEqual, localMockManager.UnsubscribeCallCount)
			})
		})

		Convey("Starting task should not succeed if missing local dep fails to subscribe", func() {
			//Create a task
			//Create a workflowmap
			wf := dsWFMap(port1)
			// Create a task that is not started immediately so we can
			// validate deps correctly.
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, false)
			So(len(errs.Errors()), ShouldEqual, 0)
			So(t, ShouldNotBeNil)
			schTask := t.(*task)
			localMockManager := &subscriptionManager{Fail: true}
			schTask.RemoteManagers.Add("", localMockManager)
			remoteMockManager := &subscriptionManager{Fail: false}
			schTask.RemoteManagers.Add(fmt.Sprintf("127.0.0.1:%v", port1), remoteMockManager)

			// Start task. We expect it to fail while subscribing deps
			terrs := sch.StartTask(t.ID())
			So(terrs, ShouldNotBeNil)
			Convey("So dependencies should have been unsubscribed", func() {
				// Ensure that unsubscribe call count is equal to subscribe call count
				// i.e that every subscribe call was followed by an unsubscribe since
				// we errored
				So(remoteMockManager.UnsubscribeCallCount, ShouldEqual, remoteMockManager.SubscribeCallCount)
				So(localMockManager.UnsubscribeCallCount, ShouldEqual, localMockManager.UnsubscribeCallCount)
			})
		})

		Convey("Starting task should suceed if all deps are available", func() {
			//Create a task
			//Create a workflowmap
			wf := dsWFMap(port1)
			// Create a task that is not started immediately so we can
			// validate deps correctly.
			t, errs := sch.CreateTask(schedule.NewSimpleSchedule(time.Second), wf, false)
			So(len(errs.Errors()), ShouldEqual, 0)
			So(t, ShouldNotBeNil)
			schTask := t.(*task)
			localMockManager := &subscriptionManager{Fail: false}
			schTask.RemoteManagers.Add("", localMockManager)
			remoteMockManager := &subscriptionManager{Fail: false}
			schTask.RemoteManagers.Add(fmt.Sprintf("127.0.0.1:%v", port1), remoteMockManager)
			terrs := sch.StartTask(t.ID())
			So(terrs, ShouldBeNil)
			Convey("So all depndencies should have been subscribed to", func() {
				// Ensure that unsubscribe call count is equal to subscribe call count
				// i.e that every subscribe call was followed by an unsubscribe since
				// we errored
				So(localMockManager.SubscribeCallCount, ShouldBeGreaterThan, 0)
				So(remoteMockManager.SubscribeCallCount, ShouldBeGreaterThan, 0)
			})
		})
	})
}

func dsWFMap(port int) *wmap.WorkflowMap {
	wf := new(wmap.WorkflowMap)

	c := wmap.NewCollectWorkflowMapNode()
	c.Config["/intel/mock/foo"] = make(map[string]interface{})
	c.Config["/intel/mock/foo"]["password"] = "required"
	pr := &wmap.ProcessWorkflowMapNode{
		Name:    "passthru",
		Version: -1,
		Config:  make(map[string]interface{}),
		Target:  fmt.Sprintf("127.0.0.1:%v", port),
	}
	pu := &wmap.PublishWorkflowMapNode{
		Name:    "file",
		Version: -1,
		Config:  make(map[string]interface{}),
	}
	pu.Config["file"] = "/dev/null"
	pr.Add(pu)
	c.Add(pr)
	e := c.AddMetric("/intel/mock/foo", 2)
	if e != nil {
		panic(e)
	}
	wf.CollectNode = c

	return wf
}

type subscriptionManager struct {
	mockMetricManager
	Fail                 bool
	SubscribeCallCount   int
	UnsubscribeCallCount int
}

func (m *subscriptionManager) SubscribeDeps(taskID string, mts []core.Metric, prs []core.Plugin) []serror.SnapError {
	if m.Fail {
		return []serror.SnapError{serror.New(errors.New("error"))}
	}
	m.SubscribeCallCount += 1
	return nil
}

func (m *subscriptionManager) UnsubscribeDeps(taskID string, mts []core.Metric, prs []core.Plugin) []serror.SnapError {
	m.UnsubscribeCallCount += 1
	return nil
}
