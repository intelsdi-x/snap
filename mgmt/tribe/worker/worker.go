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
	"os"
	"path"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

const (
	retryDelay = 500 * time.Millisecond
	retryLimit = 20
)

const (
	PluginLoadedType = iota
	PluginUnloadedType
)

const (
	TaskCreatedType = iota
	TaskStoppedType
	TaskStartedType
	TaskRemovedType
)

var (
	PluginRequestTypeLookup = map[PluginRequestType]string{
		PluginLoadedType:   "Loaded",
		PluginUnloadedType: "Unloaded",
	}

	TaskRequestTypeLookup = map[TaskRequestType]string{
		TaskCreatedType: "Created",
		TaskStoppedType: "Stopped",
		TaskStartedType: "Started",
		TaskRemovedType: "Removed",
	}
)

type PluginRequestType int

func (p PluginRequestType) String() string {
	return PluginRequestTypeLookup[p]
}

type TaskRequestType int

func (t TaskRequestType) String() string {
	return TaskRequestTypeLookup[t]
}

type PluginRequest struct {
	Plugin      core.Plugin
	RequestType PluginRequestType
	retryCount  int
}

type TaskRequest struct {
	Task        Task
	RequestType TaskRequestType
	retryCount  int
}

type Task struct {
	ID            string
	StartOnCreate bool
}

type ManagesPlugins interface {
	Load(*core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError)
	Unload(plugin core.Plugin) (core.CatalogedPlugin, serror.SnapError)
	PluginCatalog() core.PluginCatalog
}

type ManagesTasks interface {
	GetTask(id string) (core.Task, error)
	CreateTaskTribe(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors)
	StopTaskTribe(id string) []serror.SnapError
	StartTaskTribe(id string) []serror.SnapError
	RemoveTaskTribe(id string) error
}

type getsMembers interface {
	GetPluginAgreementMembers() ([]Member, error)
	GetTaskAgreementMembers() ([]Member, error)
	GetRequestPassword() string
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
	logger := log.WithFields(log.Fields{
		"_module":   "worker",
		"worker-id": id,
	})
	worker := worker{
		pluginManager: pm,
		taskManager:   tm,
		memberManager: mm,
		id:            id,
		pluginWork:    pluginQueue,
		taskWork:      taskQueue,
		waitGroup:     wg,
		quitChan:      quitChan,
		logger:        logger,
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
	logger        *log.Entry
}

func DispatchWorkers(nworkers int, pluginQueue chan PluginRequest, taskQueue chan TaskRequest, quitChan chan struct{}, workerWaitGroup *sync.WaitGroup, cp ManagesPlugins, tm ManagesTasks, mm getsMembers) {

	for i := 0; i < nworkers; i++ {
		log.WithFields(log.Fields{
			"_module": "worker",
			"_block":  "dispatch-workers",
		}).Infof("dispatching tribe worker-%d", i+1)
		worker := newWorker(i+1, pluginQueue, taskQueue, quitChan, workerWaitGroup, cp, tm, mm)
		worker.start()
	}
}

// Start "starts" the workers
func (w worker) start() {
	logger := w.logger.WithFields(log.Fields{"_block": "start"})
	// task worker
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		logger.Debug("starting task worker")
		for {
			select {
			case work := <-w.taskWork:
				// Receive a work request.
				logger := w.logger.WithFields(log.Fields{
					"task":         work.Task.ID,
					"request-type": work.RequestType.String(),
					"retries":      work.retryCount,
				})
				logger.Debug("received task work")
				if work.RequestType == TaskStartedType {
					if err := w.startTask(work.Task.ID); err != nil {
						if work.retryCount < retryLimit {
							logger.WithField("retry-count", work.retryCount).Debug("requeueing task start request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.taskWork <- work
						}
					}
				}
				if work.RequestType == TaskStoppedType {
					if err := w.stopTask(work.Task.ID); err != nil {
						if work.retryCount < retryLimit {
							logger.WithField("retry-count", work.retryCount).Debug("requeueing task stop request")
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
							logger.WithField("retry-count", work.retryCount).Debug("requeueing request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.taskWork <- work
						}
					}
				}
			case <-w.quitChan:
				logger.Infof("stopping tribe worker")
				return
			}
		}
	}()

	// plugin worker
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		logger.Debug("starting plugin worker")
		for {
			select {
			case work := <-w.pluginWork:
				// Receive a work request.
				logger := w.logger.WithFields(log.Fields{
					"plugin-name":    work.Plugin.Name(),
					"plugin-version": work.Plugin.Version(),
					"plugin-type":    work.Plugin.TypeName(),
					"request-type":   work.RequestType.String(),
				})
				logger.Debug("received plugin work")
				if work.RequestType == PluginLoadedType {
					if err := w.loadPlugin(work.Plugin); err != nil {
						if work.retryCount < retryLimit {
							logger.WithField("retry-count", work.retryCount).Debug("requeueing request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.pluginWork <- work
						}
					}
				}
				if work.RequestType == PluginUnloadedType {
					if err := w.unloadPlugin(work.Plugin); err != nil {
						if work.retryCount < retryLimit {
							logger.WithField("retry-count", work.retryCount).Debug("requeueing request")
							work.retryCount++
							time.Sleep(retryDelay)
							w.pluginWork <- work
						}
					}
				}
			case <-w.quitChan:
				w.logger.Debug("stop tribe plugin worker")
				return
			}
		}
	}()
}

func (w worker) unloadPlugin(plugin core.Plugin) error {
	logger := w.logger.WithFields(log.Fields{
		"plugin-name":    plugin.Name(),
		"plugin-version": plugin.Version(),
		"plugin-type":    plugin.TypeName(),
		"_block":         "unload-plugin",
	})
	if !w.isPluginLoaded(plugin.Name(), plugin.TypeName(), plugin.Version()) {
		return nil
	}
	if _, err := w.pluginManager.Unload(plugin); err != nil {
		logger.WithField("err", err).Info("failed to unload plugin")
		return err
	}
	return nil
}

func (w worker) loadPlugin(plugin core.Plugin) error {
	logger := w.logger.WithFields(log.Fields{
		"plugin-name":    plugin.Name(),
		"plugin-version": plugin.Version(),
		"plugin-type":    plugin.TypeName(),
		"_block":         "load-plugin",
	})
	if w.isPluginLoaded(plugin.Name(), plugin.TypeName(), plugin.Version()) {
		return nil
	}
	members, err := w.memberManager.GetPluginAgreementMembers()
	if err != nil {
		logger.Error(err)
		return err
	}
	for _, member := range shuffle(members) {
		url := fmt.Sprintf("%s://%s:%s/v1/plugins/%s/%s/%d?download=true", member.GetRestProto(), member.GetAddr(), member.GetRestPort(), plugin.TypeName(), plugin.Name(), plugin.Version())
		c, err := client.New(url, "v1", member.GetRestInsecureSkipVerify(), client.Password(w.memberManager.GetRequestPassword()))
		if err != nil {
			logger.WithFields(log.Fields{
				"err": err,
				"url": url,
			}).Info("unable to create client")
			continue
		}
		f, err := w.downloadPlugin(c, plugin)
		// If we can't download from this member, try the next
		if err != nil {
			logger.Error(err)
			continue
		}
		rp, err := core.NewRequestedPlugin(f.Name())
		if err != nil {
			logger.Error(err)
			return err
		}
		_, err = w.pluginManager.Load(rp)
		if err != nil {
			logger.Error(err)
			return err
		}
		if w.isPluginLoaded(plugin.Name(), plugin.TypeName(), plugin.Version()) {
			return nil
		}
		return errors.New("failed to load plugin")
	}
	return errors.New("failed to find a member with the plugin")
}

func (w worker) downloadPlugin(c *client.Client, plugin core.Plugin) (*os.File, error) {
	logger := w.logger.WithFields(log.Fields{
		"plugin-name":    plugin.Name(),
		"plugin-version": plugin.Version(),
		"plugin-type":    plugin.TypeName(),
		"url":            c.URL,
		"_block":         "download-plugin",
	})
	resp, err := c.TribeRequest()
	if err != nil {
		logger.WithFields(log.Fields{
			"err": err,
		}).Info("plugin not found")
		return nil, fmt.Errorf("Plugin not found at %s: %s", c.URL, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		if resp.Header.Get("Content-Type") != "application/x-gzip" {
			logger.WithField("content-type", resp.Header.Get("Content-Type")).Error("Expected application/x-gzip")
		}
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		fpath := path.Join(dir, fmt.Sprintf("%s-%s-%d", plugin.TypeName(), plugin.Name(), plugin.Version()))
		f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		io.Copy(f, resp.Body)
		f.Close()
		return f, nil
	}
	return nil, fmt.Errorf("Status code not 200 was %v: %s", resp.StatusCode, c.URL)
}
func (w worker) createTask(taskID string, startOnCreate bool) {
	logger := w.logger.WithFields(log.Fields{
		"task-id": taskID,
		"_block":  "create-task",
	})
	done := false
	_, err := w.taskManager.GetTask(taskID)
	if err == nil {
		return
	}
	for {
		members, err := w.memberManager.GetTaskAgreementMembers()
		if err != nil {
			logger.Error(err)
			continue
		}
		for _, member := range shuffle(members) {
			uri := fmt.Sprintf("%s://%s:%s", member.GetRestProto(), member.GetAddr(), member.GetRestPort())
			logger.Debugf("getting task %v from %v", taskID, uri)

			c, err := client.New(uri, "v1", member.GetRestInsecureSkipVerify(), client.Password(w.memberManager.GetRequestPassword()))
			if err != nil {
				logger.Error(err)
				continue
			}

			taskResult := c.GetTask(taskID)
			if taskResult.Err != nil {
				logger.WithField("err", taskResult.Err.Error()).Debug("error getting task")
				continue
			}
			// this block addresses the condition when we are creating and starting
			// a task and the task is created but fails to start (deps were not yet met)
			if startOnCreate {
				if _, err := w.taskManager.GetTask(taskID); err == nil {
					logger.Debug("starting task")
					if errs := w.taskManager.StartTaskTribe(taskID); errs != nil {
						fields := log.Fields{}
						for idx, e := range errs {
							fields[fmt.Sprintf("err-%d", idx)] = e.Error()
						}
						logger.WithFields(fields).Error("error starting task")
						continue
					}
					done = true
					break
				}
			}
			logger.Debug("creating task")
			opt := core.SetTaskID(taskID)
			_, errs := w.taskManager.CreateTaskTribe(
				getSchedule(taskResult.ScheduledTaskReturned.Schedule),
				taskResult.Workflow,
				startOnCreate,
				opt)
			if errs != nil && len(errs.Errors()) > 0 {
				fields := log.Fields{}
				for idx, e := range errs.Errors() {
					fields[fmt.Sprintf("err-%d", idx)] = e
				}
				logger.WithFields(fields).Debug("error creating task")
				continue
			}
			logger.Debugf("task created")
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
	logger := w.logger.WithFields(log.Fields{
		"task-id": taskID,
		"_block":  "start-task",
	})
	logger.Debug("starting task")
	errs := w.taskManager.StartTaskTribe(taskID)
	if errs == nil || len(errs) == 0 {
		return nil
	}
	if errs != nil {
		for _, err := range errs {
			if err.Error() == scheduler.ErrTaskAlreadyRunning.Error() {
				logger.WithFields(err.Fields()).Info(err)
				return nil
			} else {
				logger.WithFields(err.Fields()).Info(err)
			}
		}
	}
	return errors.New("error starting task")
}

func (w worker) stopTask(taskID string) error {
	logger := w.logger.WithFields(log.Fields{
		"task-id": taskID,
		"_block":  "stop-task",
	})
	errs := w.taskManager.StopTaskTribe(taskID)
	if errs == nil || len(errs) == 0 {
		return nil
	}
	for _, err := range errs {
		if err.Error() == scheduler.ErrTaskAlreadyStopped.Error() {
			logger.WithFields(err.Fields()).Info(err)
			return nil
		} else {
			logger.WithFields(err.Fields()).Info(err)
		}
	}
	return errors.New("error stopping task")
}

func (w worker) removeTask(taskID string) error {
	logger := w.logger.WithFields(log.Fields{
		"task-id": taskID,
		"_block":  "remove-task",
	})
	err := w.taskManager.RemoveTaskTribe(taskID)
	if err == nil {
		return nil
	}
	logger.Info(err)
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
			w.logger.WithFields(log.Fields{
				"name":    n,
				"version": v,
				"type":    t,
				"_block":  "is-plugin-loaded",
			}).Debugf("plugin already loaded")
			return true
		}
	}
	return false
}

func getSchedule(s *core.Schedule) schedule.Schedule {
	switch s.Type {
	case "simple":
		d, e := time.ParseDuration(s.Interval)
		if e != nil {
			log.WithField("_block", "get-schedule").Error(e)
			return nil
		}
		return schedule.NewSimpleSchedule(d)
	}
	return nil
}
