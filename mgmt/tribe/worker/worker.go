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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

const (
	PluginLoadedType = iota
)

const (
	TaskCreatedType = iota
	TaskStoppedType
	TaskStartedType
	TaskRemovedType
)

const (
	retryDelay = 500 * time.Millisecond
	retryLimit = 20
)

var workerLogger = log.WithFields(log.Fields{
	"_module": "worker",
})

type PluginRequest struct {
	Plugin      core.Plugin
	RequestType int
	retryCount  int
}

type TaskRequest struct {
	Task        Task
	RequestType int
	retryCount  int
}

type Task struct {
	ID            string
	StartOnCreate bool
}

type ManagesPlugins interface {
	Load(...string) (core.CatalogedPlugin, perror.PulseError)
	Unload(plugin core.Plugin) (core.CatalogedPlugin, perror.PulseError)
	PluginCatalog() core.PluginCatalog
}

type ManagesTasks interface {
	GetTask(id string) (core.Task, error)
	CreateTaskTribe(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors)
	StopTaskTribe(id string) []perror.PulseError
	StartTaskTribe(id string) []perror.PulseError
	RemoveTaskTribe(id string) error
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
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"task":        work.Task.ID,
					"worker":      w.id,
					"requestType": work.RequestType,
				})
				wLogger.Debug("received task work")
				if work.RequestType == TaskStartedType {
					if err := w.startTask(work.Task.ID); err != nil {
						if work.retryCount < retryLimit {
							workerLogger.WithField("retryCount", work.retryCount).Debug("requeueing start request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.taskWork <- work
						}
					}
				}
				if work.RequestType == TaskStoppedType {
					if err := w.stopTask(work.Task.ID); err != nil {
						if work.retryCount < retryLimit {
							workerLogger.WithField("retryCount", work.retryCount).Debug("requeueing stop request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.taskWork <- work
						}
					}
				}
				if work.RequestType == TaskCreatedType {
					w.createTask(work.Task.ID, work.Task.StartOnCreate)
				}
				if work.RequestType == TaskRemovedType {
					if err := w.removeTask(work.Task.ID); err != nil {
						if work.retryCount < retryLimit {
							workerLogger.WithField("retryCount", work.retryCount).Debug("requeueing remove request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.taskWork <- work
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
							dir, err := ioutil.TempDir("", "")
							if err != nil {
								wLogger.Error(err)
								continue
							}
							f, err := os.Create(path.Join(dir, fmt.Sprintf("%s-%s-%d", work.Plugin.TypeName(), work.Plugin.Name(), work.Plugin.Version())))
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

func (w worker) createTask(taskID string, startOnCreate bool) {
	workerLogger = workerLogger.WithFields(log.Fields{
		"task":   taskID,
		"_block": "worker-createTask",
	})
	done := false
	_, err := w.taskManager.GetTask(taskID)
	if err == nil {
		return
	}
	for {
		members, err := w.memberManager.GetTaskAgreementMembers()
		if err != nil {
			workerLogger.Error(err)
			continue
		}
		for _, member := range shuffle(members) {
			uri := fmt.Sprintf("%s://%s:%s", member.GetRestProto(), member.GetAddr(), member.GetRestPort())
			workerLogger.Debugf("getting task %v from %v", taskID, uri)
			c := client.New(uri, "v1", member.GetRestInsecureSkipVerify())
			taskResult := c.GetTask(taskID)
			if taskResult.Err != nil {
				workerLogger.WithField("err", taskResult.Err.Error()).Debug("error getting task")
				continue
			}
			workerLogger.Debug("creating task")
			opt := core.SetTaskID(taskID)
			t, errs := w.taskManager.CreateTaskTribe(
				getSchedule(taskResult.ScheduledTaskReturned.Schedule),
				taskResult.Workflow,
				startOnCreate,
				opt)
			if errs != nil && len(errs.Errors()) > 0 {
				fields := log.Fields{
					"task":   taskID,
					"worker": w.id,
					"_block": "start",
				}
				for idx, e := range errs.Errors() {
					fields[fmt.Sprintf("err-%d", idx)] = e
				}
				workerLogger.WithFields(fields).Debug("error creating task")
				continue
			}
			workerLogger.WithField("id", t.ID()).Debugf("task created")
			done = true
			break
		}
		if done {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (w worker) startTask(taskID string) error {
	workerLogger = workerLogger.WithFields(log.Fields{
		"task":   taskID,
		"_block": "worker-startTask",
	})
	workerLogger.Debug("starting task")
	errs := w.taskManager.StartTaskTribe(taskID)
	if errs == nil || len(errs) == 0 {
		return nil
	}
	if errs != nil {
		for _, err := range errs {
			if err.Error() == scheduler.ErrTaskAlreadyRunning.Error() {
				workerLogger.WithFields(err.Fields()).Info(err)
				return nil
			} else {
				workerLogger.WithFields(err.Fields()).Error(err)
			}
		}
	}
	return errors.New("error starting task")
}

func (w worker) stopTask(taskID string) error {
	workerLogger = workerLogger.WithFields(log.Fields{
		"task":   taskID,
		"_block": "worker-stopTask",
	})
	errs := w.taskManager.StopTaskTribe(taskID)
	if errs == nil || len(errs) == 0 {
		return nil
	}
	for _, err := range errs {
		if err.Error() == scheduler.ErrTaskAlreadyStopped.Error() {
			workerLogger.WithFields(err.Fields()).Info(err)
			return nil
		} else {
			workerLogger.WithFields(err.Fields()).Error(err)
		}
	}
	return errors.New("error stopping task")
}

func (w worker) removeTask(taskID string) error {
	workerLogger = workerLogger.WithFields(log.Fields{
		"task":   taskID,
		"_block": "worker-removeTask",
	})
	err := w.taskManager.RemoveTaskTribe(taskID)
	if err == nil {
		return nil
	}
	workerLogger.WithField("task", taskID).Error(err)
	return err
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
		if item.TypeName() == t &&
			item.Name() == n &&
			item.Version() == v {
			workerLogger.WithFields(log.Fields{
				"name":    n,
				"version": v,
				"type":    t,
				"_block":  "isPluginLoaded",
			}).Debugf("Plugin already loaded")
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
