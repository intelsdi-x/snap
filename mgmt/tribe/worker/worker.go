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

package worker

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

const (
	PluginLoadedType = iota
)

const (
	TaskCreatedType = iota
)

var workerLogger = log.WithFields(log.Fields{
	"_module": "worker",
})

type PluginRequest struct {
	Plugin      core.Plugin
	RequestType int
}

type TaskRequest struct {
	Task        Task
	RequestType int
}

type Task struct {
	ID            string
	StartOnCreate bool
}

type ManagesPlugins interface {
	Load(path string) (core.CatalogedPlugin, perror.PulseError)
	Unload(plugin core.Plugin) (core.CatalogedPlugin, perror.PulseError)
	PluginCatalog() core.PluginCatalog
}

type ManagesTasks interface {
	GetTask(id string) (core.Task, error)
	CreateTask(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors)
}

type getsMembers interface {
	GetPluginAgreementMembers() ([]Member, error)
	GetTaskAgreementMembers() ([]Member, error)
}

type Member interface {
	GetAddr() net.IP
	GetRestPort() string
	GetRestProto() string
	GetRestInsecureSkipVerify() bool
	GetName() string
}

// newPluginWorker
func newWorker(id int,
	pluginQueue chan PluginRequest,
	taskQueue chan TaskRequest,
	quitChan chan struct{},
	wg *sync.WaitGroup,
	pm ManagesPlugins,
	tm ManagesTasks,
	mm getsMembers) worker {
	worker := worker{
		pluginManager: pm,
		taskManager:   tm,
		memberManager: mm,
		id:            id,
		pluginWork:    pluginQueue,
		taskWork:      taskQueue,
		waitGroup:     wg,
		quitChan:      quitChan,
	}

	return worker
}

type worker struct {
	pluginManager ManagesPlugins
	memberManager getsMembers
	taskManager   ManagesTasks
	id            int
	pluginWork    chan PluginRequest
	taskWork      chan TaskRequest
	quitChan      chan struct{}
	waitGroup     *sync.WaitGroup
}

func DispatchWorkers(nworkers int, pluginQueue chan PluginRequest, taskQueue chan TaskRequest, quitChan chan struct{}, workerWaitGroup *sync.WaitGroup, cp ManagesPlugins, tm ManagesTasks, mm getsMembers) {

	for i := 0; i < nworkers; i++ {
		workerLogger.Infof("Starting tribe worker-%d", i+1)
		worker := newWorker(i+1, pluginQueue, taskQueue, quitChan, workerWaitGroup, cp, tm, mm)
		worker.start()
	}
}

// Start "starts" the workers
func (w worker) start() {
	// task worker
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		workerLogger.Debugf("Starting task worker-%d", w.id)
		for {
			select {
			case work := <-w.taskWork:
				done := false
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"task":   work.Task.ID,
					"worker": w.id,
					"_block": "start",
				})
				wLogger.Error("received task work")
				if work.RequestType == TaskCreatedType {
					_, err := w.taskManager.GetTask(work.Task.ID)
					if err != nil {
						for {
							members, err := w.memberManager.GetTaskAgreementMembers()
							if err != nil {
								wLogger.Error(err)
								continue
							}
							for _, member := range shuffle(members) {
								uri := fmt.Sprintf("%s://%s:%s", member.GetRestProto(), member.GetAddr(), member.GetRestPort())
								wLogger.Debugf("getting task %v from %v", work.Task.ID, uri)
								c := client.New(uri, "v1", member.GetRestInsecureSkipVerify())
								taskResult := c.GetTask(work.Task.ID)
								if taskResult.Err != nil {
									wLogger.WithField("err", taskResult.Err.Error()).Debug("error getting task")
									continue
								}
								wLogger.Debug("creating task")
								opt := core.SetTaskID(work.Task.ID)
								t, errs := w.taskManager.CreateTask(
									getSchedule(taskResult.ScheduledTaskReturned.Schedule),
									taskResult.Workflow,
									work.Task.StartOnCreate,
									opt)
								if errs != nil && len(errs.Errors()) > 0 {
									fields := log.Fields{
										"task":   work.Task.ID,
										"worker": w.id,
										"_block": "start",
									}
									for idx, e := range errs.Errors() {
										fields[fmt.Sprintf("err-%d", idx)] = e
									}
									wLogger.WithFields(fields).Debug("error creating task")
									continue
								}
								wLogger.WithField("id", t.ID()).Debugf("task created")
								done = true
								break
							}
							if done {
								break
							}
							time.Sleep(500 * time.Millisecond)
						}
					}
				}

			case <-w.quitChan:
				workerLogger.Infof("Tribe plugin worker-%d is stopping\n", w.id)
				return

			}
		}
	}()

	// plugin worker
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		workerLogger.Debugf("Starting plugin worker-%d", w.id)
		for {
			select {
			case work := <-w.pluginWork:
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"plugin_name":    work.Plugin.Name(),
					"plugin_version": work.Plugin.Version(),
					"plugin_type":    work.Plugin.TypeName(),
					"worker":         w.id,
					"_block":         "start",
				})
				wLogger.Debug("received plugin work")
				done := false
				for {
					if w.isPluginLoaded(work.Plugin.Name(), work.Plugin.TypeName(), work.Plugin.Version()) {
						break
					}
					members, err := w.memberManager.GetPluginAgreementMembers()
					if err != nil {
						wLogger.Error(err)
						continue
					}
					for _, member := range shuffle(members) {
						url := fmt.Sprintf("%s://%s:%s/v1/plugins/%s/%s/%d?download=true", member.GetRestProto(), member.GetAddr(), member.GetRestPort(), work.Plugin.TypeName(), work.Plugin.Name(), work.Plugin.Version())
						resp, err := http.Get(url)
						if err != nil {
							wLogger.Error(err)
							continue
						}
						if resp.StatusCode == 200 {
							if resp.Header.Get("Content-Type") != "application/x-gzip" {
								wLogger.WithField("content-type", resp.Header.Get("Content-Type")).Error("Expected application/x-gzip")
							}
							f, err := ioutil.TempFile("", fmt.Sprintf("%s-%s-%d", work.Plugin.TypeName(), work.Plugin.Name(), work.Plugin.Version()))
							if err != nil {
								wLogger.Error(err)
								f.Close()
								continue
							}
							io.Copy(f, resp.Body)
							f.Close()
							err = os.Chmod(f.Name(), 0700)
							if err != nil {
								wLogger.Error(err)
								continue
							}
							_, err = w.pluginManager.Load(f.Name())
							if err != nil {
								wLogger.Error(err)
								continue
							}
							if w.isPluginLoaded(work.Plugin.Name(), work.Plugin.TypeName(), work.Plugin.Version()) {
								done = true
								break
							}
						}
					}
					if done {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}
			case <-w.quitChan:
				workerLogger.Debugf("Tribe plugin worker-%d is stopping\n", w.id)
				return
			}
		}
	}()
}

func shuffle(m []Member) []Member {
	result := make([]Member, len(m))
	perm := rand.Perm(len(m))
	for i, v := range perm {
		result[v] = m[i]
	}
	return result
}

func (w worker) isPluginLoaded(n, t string, v int) bool {
	catalog := w.pluginManager.PluginCatalog()
	for _, item := range catalog {
		workerLogger.WithFields(log.Fields{
			"name":    n,
			"version": v,
			"type":    t,
		}).Errorf("loaded plugin.. looking for %v %v %v", item.Name(), item.Version(), item.TypeName())
		if item.TypeName() == t &&
			item.Name() == n &&
			item.Version() == v {
			workerLogger.WithField("_block", "isPluginLoaded").Error("Plugin already loaded")
			return true
		}
	}
	return false
}

func getSchedule(s *request.Schedule) schedule.Schedule {
	switch s.Type {
	case "simple":
		d, e := time.ParseDuration(s.Interval)
		if e != nil {
			workerLogger.WithField("_block", "getSchedule").Error(e)
			return nil
		}
		return schedule.NewSimpleSchedule(d)
	}
	return nil
}
