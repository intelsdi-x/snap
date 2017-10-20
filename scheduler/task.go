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
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/intelsdi-x/gomit"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/grpc/controlproxy"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

const (
	// DefaultDeadlineDuration - The default timeout is 5 second
	DefaultDeadlineDuration = time.Second * 5
	// DefaultStopOnFailure is used to set the number of failures before a task is disabled
	DefaultStopOnFailure = 10
)

var (
	taskLogger = schedulerLogger.WithField("_module", "scheduler-task")

	// ErrTaskNotFound - The error message for task not found
	ErrTaskNotFound = errors.New("Task not found")
	// ErrTaskNotStopped - The error message for task must be stopped
	ErrTaskNotStopped = errors.New("Task must be stopped")
	// ErrTaskHasAlreadyBeenAdded - The error message for task has already been added
	ErrTaskHasAlreadyBeenAdded = errors.New("Task has already been added")
	// ErrTaskDisabledOnFailures - The error message for task disabled due to consecutive failures
	ErrTaskDisabledOnFailures = errors.New("Task disabled due to consecutive failures")
	// ErrTaskNotDisabled - The error message for task must be disabled
	ErrTaskNotDisabled = errors.New("Task must be disabled")
)

type task struct {
	sync.Mutex //protects state

	id                 string
	name               string
	schResponseChan    chan schedule.Response
	killChan           chan struct{}
	schedule           schedule.Schedule
	workflow           *schedulerWorkflow
	state              core.TaskState
	creationTime       time.Time
	lastFireTime       time.Time
	manager            managesWork
	metricsManager     managesMetrics
	deadlineDuration   time.Duration
	hitCount           uint
	missedIntervals    uint
	failureMutex       sync.Mutex
	failedRuns         uint
	lastFailureMessage string
	lastFailureTime    time.Time
	stopOnFailure      int
	eventEmitter       gomit.Emitter
	RemoteManagers     managers
	isStream           bool

	maxCollectDuration time.Duration
	maxMetricsBuffer   int64
}

//NewTask creates a Task
func newTask(s schedule.Schedule, wf *schedulerWorkflow, m *workManager, mm managesMetrics, emitter gomit.Emitter, opts ...core.TaskOption) (*task, error) {

	//Task would always be given a default name.
	//However if a user want to change this name, she can pass optional arguments, in form of core.TaskOption
	//The new name then get over written.

	taskID := uuid.New()
	name := fmt.Sprintf("Task-%s", taskID)
	wf.eventEmitter = emitter
	mgrs := newManagers(mm)
	err := createTaskClients(&mgrs, wf)
	if err != nil {
		return nil, err
	}
	_, stream := s.(*schedule.StreamingSchedule)
	task := &task{
		id:               taskID,
		name:             name,
		schResponseChan:  make(chan schedule.Response),
		schedule:         s,
		state:            core.TaskStopped,
		creationTime:     time.Now(),
		workflow:         wf,
		manager:          m,
		metricsManager:   mm,
		deadlineDuration: DefaultDeadlineDuration,
		stopOnFailure:    DefaultStopOnFailure,
		eventEmitter:     emitter,
		RemoteManagers:   mgrs,
		isStream:         stream,
	}
	//set options
	for _, opt := range opts {
		opt(task)
	}
	return task, nil
}

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (t *task) Option(opts ...core.TaskOption) core.TaskOption {
	var previous core.TaskOption
	for _, opt := range opts {
		previous = opt(t)
	}
	return previous
}

func (t *task) MaxCollectDuration() time.Duration {
	return t.maxCollectDuration
}

func (t *task) SetMaxCollectDuration(ti time.Duration) {
	t.maxCollectDuration = ti
}

func (t *task) MaxMetricsBuffer() int64 {
	return t.maxMetricsBuffer
}

func (t *task) SetMaxMetricsBuffer(i int64) {
	t.maxMetricsBuffer = i
}

//Returns the name of the task
func (t *task) GetName() string {
	return t.name
}

func (t *task) SetName(name string) {
	t.name = name
}

// CreateTime returns the time the task was created.
func (t *task) CreationTime() *time.Time {
	return &t.creationTime
}

func (t *task) DeadlineDuration() time.Duration {
	return t.deadlineDuration
}

func (t *task) SetDeadlineDuration(d time.Duration) {
	t.deadlineDuration = d
}

func (t *task) SetTaskID(id string) {
	t.id = id
}

// HitCount returns the number of times the task has fired.
func (t *task) HitCount() uint {
	return t.hitCount
}

// Id returns the tasks Id.
func (t *task) ID() string {
	return t.id
}

// LastRunTime returns the time of the tasks last run.
func (t *task) LastRunTime() *time.Time {
	return &t.lastFireTime
}

// MissedCount returns the number of intervals missed.
func (t *task) MissedCount() uint {
	return t.missedIntervals
}

// FailedRuns returns the number of intervals missed.
func (t *task) FailedCount() uint {
	return t.failedRuns
}

// LastFailureMessage returns the last error from a task run
func (t *task) LastFailureMessage() string {
	return t.lastFailureMessage
}

// State returns state of the task.
func (t *task) State() core.TaskState {
	return t.state
}

// Status returns the state of the workflow.
func (t *task) Status() WorkflowState {
	return t.workflow.State()
}

func (t *task) SetStopOnFailure(v int) {
	t.stopOnFailure = v
}

func (t *task) SetID(id string) {
	t.id = id
}

func (t *task) GetStopOnFailure() int {
	return t.stopOnFailure
}

// Spin will start a task spinning in its own routine while it waits for its
// schedule.
func (t *task) Spin() {
	// We need to lock long enough to change state
	t.Lock()
	defer t.Unlock()
	// if this task is a streaming task
	if t.isStream {
		t.state = core.TaskSpinning
		t.killChan = make(chan struct{})
		go t.stream()
		return
	}

	// Reset the lastFireTime at each Spin.
	// This ensures misses are tracked only forward of the point
	// in time that a task starts spinning. E.g. stopping a task,
	// waiting a period of time, and starting the task won't show
	// misses for the interval while stopped.
	t.lastFireTime = time.Time{}

	if t.state == core.TaskStopped || t.state == core.TaskEnded {
		t.state = core.TaskSpinning
		t.killChan = make(chan struct{})
		// spin in a goroutine
		go t.spin()
	}
}

// Fork stream stuff here
func (t *task) stream() {
	var consecutiveFailures int
	resetTime := time.Second * 3
	for {
		metricsChan, errChan, err := t.metricsManager.StreamMetrics(
			t.id,
			t.workflow.tags,
			t.maxCollectDuration,
			t.maxMetricsBuffer)
		if err != nil {
			consecutiveFailures++
			// check task failures
			if t.stopOnFailure >= 0 && consecutiveFailures >= t.stopOnFailure {
				taskLogger.WithFields(log.Fields{
					"_block":               "stream",
					"task-id":              t.id,
					"task-name":            t.name,
					"consecutive failures": consecutiveFailures,
					"error":                t.lastFailureMessage,
				}).Error(ErrTaskDisabledOnFailures)
				// disable the task
				t.disable(t.lastFailureMessage)
				return
			}
			// If we are unsuccessful at setting up the stream
			// wait for a second and then try again until either
			// the connection is successful or we pass the
			// acceptable number of consecutive failures
			time.Sleep(resetTime)
			continue
		} else {
			consecutiveFailures = 0
		}
		done := false
		for !done {
			if errChan == nil {
				break
			}
			select {
			case <-t.killChan:
				t.Lock()
				t.state = core.TaskStopped
				t.Unlock()
				done = true
				event := new(scheduler_event.TaskStoppedEvent)
				event.TaskID = t.id
				defer t.eventEmitter.Emit(event)
				return
			case mts, ok := <-metricsChan:
				if !ok {
					metricsChan = nil
					break
				}
				if len(mts) == 0 {
					continue
				}
				t.hitCount++
				consecutiveFailures = 0
				t.workflow.StreamStart(t, mts)
			case err := <-errChan:
				taskLogger.WithFields(log.Fields{
					"_block":    "stream",
					"task-id":   t.id,
					"task-name": t.name,
				}).Error("Error: " + err.Error())
				consecutiveFailures++
				if err.Error() == "connection broken" {
					// Wait here before trying to reconnect to allow time
					// for plugin restarts.
					time.Sleep(resetTime)
					done = true
				}
				// check task failures
				if t.stopOnFailure >= 0 && consecutiveFailures >= t.stopOnFailure {
					taskLogger.WithFields(log.Fields{
						"_block":               "stream",
						"task-id":              t.id,
						"task-name":            t.name,
						"consecutive failures": consecutiveFailures,
						"error":                t.lastFailureMessage,
					}).Error(ErrTaskDisabledOnFailures)
					// disable the task
					t.disable(t.lastFailureMessage)
					return
				}
			}
		}
	}
}

func (t *task) Stop() {
	t.Lock()
	defer t.Unlock()
	if t.state == core.TaskFiring || t.state == core.TaskSpinning {
		t.state = core.TaskStopping
		close(t.killChan)
	}
}

// UnsubscribePlugins groups task dependencies by the node they live in workflow and unsubscribe them
func (t *task) UnsubscribePlugins() []serror.SnapError {
	depGroups := getWorkflowPlugins(t.workflow.processNodes, t.workflow.publishNodes, t.workflow.metrics)
	var errs []serror.SnapError
	for k := range depGroups {
		event := &scheduler_event.PluginsUnsubscribedEvent{
			TaskID:  t.ID(),
			Plugins: depGroups[k].subscribedPlugins,
		}
		defer t.eventEmitter.Emit(event)
		mgr, err := t.RemoteManagers.Get(k)
		if err != nil {
			errs = append(errs, serror.New(err))
		} else {
			uerrs := mgr.UnsubscribeDeps(t.ID())
			if len(uerrs) > 0 {
				errs = append(errs, uerrs...)
			}
		}
	}
	for _, err := range errs {
		taskLogger.WithFields(log.Fields{
			"_block":     "UnsubscribePlugins",
			"task-id":    t.id,
			"task-name":  t.name,
			"task-state": t.state,
		}).Error(err)
	}
	return errs
}

// SubscribePlugins groups task dependencies by the node they live in workflow and subscribe them.
// If there are errors with subscribing any deps, manage unsubscribing all other deps that may have already been subscribed
// and then return the errors.
func (t *task) SubscribePlugins() ([]string, []serror.SnapError) {
	depGroups := getWorkflowPlugins(t.workflow.processNodes, t.workflow.publishNodes, t.workflow.metrics)
	var subbedDeps []string
	for k := range depGroups {
		var errs []serror.SnapError
		mgr, err := t.RemoteManagers.Get(k)
		if err != nil {
			errs = append(errs, serror.New(err))
		} else {
			errs = mgr.SubscribeDeps(t.ID(), depGroups[k].requestedMetrics, depGroups[k].subscribedPlugins, t.workflow.configTree)
		}
		// If there are errors with subscribing any deps, go through and unsubscribe all other
		// deps that may have already been subscribed then return the errors.
		if len(errs) > 0 {
			for _, key := range subbedDeps {
				mgr, err := t.RemoteManagers.Get(key)
				if err != nil {
					errs = append(errs, serror.New(err))
				} else {
					// sending empty mts to unsubscribe to indicate task should not start
					uerrs := mgr.UnsubscribeDeps(t.ID())
					errs = append(errs, uerrs...)
				}
			}
			return nil, errs
		}
		// If subscribed successfully add to subbedDeps
		subbedDeps = append(subbedDeps, k)
	}

	return subbedDeps, nil
}

//Enable changes the state from Disabled to Stopped
func (t *task) Enable() error {
	t.Lock()
	defer t.Unlock()

	if t.state != core.TaskDisabled {
		return ErrTaskNotDisabled
	}
	t.state = core.TaskStopped

	return nil
}

func (t *task) Kill() {
	t.Lock()
	defer t.Unlock()
	if t.state == core.TaskFiring || t.state == core.TaskSpinning {
		close(t.killChan)
		t.state = core.TaskDisabled
	}
}

func (t *task) WMap() *wmap.WorkflowMap {
	return t.workflow.workflowMap
}

func (t *task) Schedule() schedule.Schedule {
	return t.schedule
}

func (t *task) spin() {
	var consecutiveFailures int
	for {
		taskLogger.Debug("task spin loop")
		// Start go routine to wait on schedule
		go t.waitForSchedule()
		// wait here on
		//  schResponseChan - response from schedule
		//  killChan - signals task needs to be stopped
		select {
		case sr := <-t.schResponseChan:
			switch sr.State() {
			// If response show this schedule is still active we fire
			case schedule.Active:
				t.missedIntervals += sr.Missed()
				t.fire()
				if t.lastFailureTime == t.lastFireTime {
					consecutiveFailures++
					taskLogger.WithFields(log.Fields{
						"_block":                    "spin",
						"task-id":                   t.id,
						"task-name":                 t.name,
						"consecutive failures":      consecutiveFailures,
						"consecutive failure limit": t.stopOnFailure,
						"error":                     t.lastFailureMessage,
					}).Warn("Task failed")
				} else {
					consecutiveFailures = 0
				}
				if t.stopOnFailure >= 0 && consecutiveFailures >= t.stopOnFailure {
					taskLogger.WithFields(log.Fields{
						"_block":               "spin",
						"task-id":              t.id,
						"task-name":            t.name,
						"consecutive failures": consecutiveFailures,
						"error":                t.lastFailureMessage,
					}).Error(ErrTaskDisabledOnFailures)

					// disable the task
					t.disable(t.lastFailureMessage)
					return
				}

			// Schedule has ended
			case schedule.Ended:
				// You must lock task to change state
				t.Lock()
				t.state = core.TaskEnded
				t.Unlock()
				// Send task ended event
				event := new(scheduler_event.TaskEndedEvent)
				event.TaskID = t.id
				defer t.eventEmitter.Emit(event)
				return //spin

			// Schedule has errored
			case schedule.Error:
				// disable the task
				failureMessage := sr.Error().Error()
				t.disable(failureMessage)
				return //spin

			}
		case <-t.killChan:
			// Only here can it truly be stopped
			t.Lock()
			t.state = core.TaskStopped
			t.lastFireTime = time.Time{}
			t.Unlock()
			event := new(scheduler_event.TaskStoppedEvent)
			event.TaskID = t.id
			defer t.eventEmitter.Emit(event)
			return
		}
	}
}

func (t *task) fire() {
	t.Lock()
	defer t.Unlock()

	t.state = core.TaskFiring
	t.lastFireTime = time.Now()
	t.workflow.Start(t)
	t.hitCount++
	t.state = core.TaskSpinning
}

// disable proceeds disabling a task which consists of changing task state to disabled and emitting an appropriate event
func (t *task) disable(failureMsg string) {
	t.Lock()
	t.state = core.TaskDisabled
	t.Unlock()

	// Send task disabled event
	event := new(scheduler_event.TaskDisabledEvent)
	event.TaskID = t.id
	event.Why = fmt.Sprintf("Task disabled with error: %s", failureMsg)
	defer t.eventEmitter.Emit(event)
}

func (t *task) waitForSchedule() {
	select {
	case <-t.killChan:
		return
	case t.schResponseChan <- t.schedule.Wait(t.lastFireTime):
	}
}

// RecordFailure updates the failed runs and last failure properties
func (t *task) RecordFailure(e []error) {
	// We synchronize this update to ensure it is atomic
	t.failureMutex.Lock()
	defer t.failureMutex.Unlock()
	t.failedRuns++
	t.lastFailureTime = t.lastFireTime
	t.lastFailureMessage = e[len(e)-1].Error()
}

type taskCollection struct {
	*sync.Mutex

	table map[string]*task
}

func newTaskCollection() *taskCollection {
	return &taskCollection{
		Mutex: &sync.Mutex{},

		table: make(map[string]*task),
	}
}

// Get given a task id returns a Task or nil if not found
func (t *taskCollection) Get(id string) *task {
	t.Lock()
	defer t.Unlock()

	if t, ok := t.table[id]; ok {
		return t
	}
	return nil
}

// Add given a reference to a task adds it to the collection of tasks.  An
// error is returned if the task already exists in the collection.
func (t *taskCollection) add(task *task) error {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.table[task.id]; !ok {
		//If we don't already have this task in the collection save it
		t.table[task.id] = task
	} else {
		taskLogger.WithFields(log.Fields{
			"_module": "scheduler-taskCollection",
			"_block":  "add",
			"task id": task.id,
		}).Error(ErrTaskHasAlreadyBeenAdded.Error())
		return ErrTaskHasAlreadyBeenAdded
	}

	return nil
}

// remove will remove a given task from tasks.  The task must be stopped.
// Can return errors ErrTaskNotFound and ErrTaskNotStopped.
func (t *taskCollection) remove(task *task) error {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.table[task.id]; ok {
		if task.state != core.TaskStopped && task.state != core.TaskDisabled && task.state != core.TaskEnded {
			taskLogger.WithFields(log.Fields{
				"_block":  "remove",
				"task id": task.id,
			}).Error(ErrTaskNotStopped)
			return ErrTaskNotStopped
		}
		delete(t.table, task.id)
	} else {
		taskLogger.WithFields(log.Fields{
			"_block":  "remove",
			"task id": task.id,
		}).Error(ErrTaskNotFound)
		return ErrTaskNotFound
	}
	return nil
}

// Table returns a copy of the taskCollection
func (t *taskCollection) Table() map[string]*task {
	t.Lock()
	defer t.Unlock()
	tasks := make(map[string]*task)
	for id, t := range t.table {
		tasks[id] = t
	}
	return tasks
}

// createTaskClients walks the workflowmap and creates clients for this task
// remoteManagers so that nodes that require proxy request can make them.
func createTaskClients(mgrs *managers, wf *schedulerWorkflow) error {
	return walkWorkflow(wf.processNodes, wf.publishNodes, mgrs)
}

func walkWorkflow(prnodes []*processNode, pbnodes []*publishNode, mgrs *managers) error {
	for _, pr := range prnodes {
		if pr.Target != "" {
			host, port, err := net.SplitHostPort(pr.Target)
			if err != nil {
				return err
			}
			p, err := strconv.Atoi(port)
			if err != nil {
				return err
			}
			proxy, err := controlproxy.New(host, p)
			if err != nil {
				return err
			}
			mgrs.Add(pr.Target, proxy)
		}
		err := walkWorkflow(pr.ProcessNodes, pr.PublishNodes, mgrs)
		if err != nil {
			return err
		}

	}
	for _, pu := range pbnodes {
		if pu.Target != "" {
			host, port, err := net.SplitHostPort(pu.Target)
			if err != nil {
				return err
			}
			p, err := strconv.Atoi(port)
			if err != nil {
				return err
			}
			proxy, err := controlproxy.New(host, p)
			if err != nil {
				return err
			}
			mgrs.Add(pu.Target, proxy)
		}
	}
	return nil
}
