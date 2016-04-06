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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"
	"github.com/pborman/uuid"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

const (
	// DefaultDeadlineDuration - The default timeout is 5 second
	DefaultDeadlineDuration = time.Second * 5
	// DefaultStopOnFailure - The default stopping a failure is after three tries
	DefaultStopOnFailure = 3
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
	stopOnFailure      uint
	eventEmitter       gomit.Emitter
}

//NewTask creates a Task
func newTask(s schedule.Schedule, wf *schedulerWorkflow, m *workManager, mm managesMetrics, emitter gomit.Emitter, opts ...core.TaskOption) *task {

	//Task would always be given a default name.
	//However if a user want to change this name, she can pass optional arguments, in form of core.TaskOption
	//The new name then get over written.

	taskID := uuid.New()
	name := fmt.Sprintf("Task-%s", taskID)
	wf.eventEmitter = emitter
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
	}
	//set options
	for _, opt := range opts {
		opt(task)
	}
	return task
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

func (t *task) SetStopOnFailure(v uint) {
	t.stopOnFailure = v
}

func (t *task) SetID(id string) {
	t.id = id
}

func (t *task) GetStopOnFailure() uint {
	return t.stopOnFailure
}

// Spin will start a task spinning in its own routine while it waits for its
// schedule.
func (t *task) Spin() {
	// We need to lock long enough to change state
	t.Lock()
	defer t.Unlock()
	// Reset the lastFireTime at each Spin.
	// This ensures misses are tracked only forward of the point
	// in time that a task starts spinning. E.g. stopping a task,
	// waiting a period of time, and starting the task won't show
	// misses for the interval while stopped.
	t.lastFireTime = time.Now()
	if t.state == core.TaskStopped {
		t.state = core.TaskSpinning
		t.killChan = make(chan struct{})
		// spin in a goroutine
		go t.spin()
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
	var consecutiveFailures uint
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
			// If response show this schedule is stil active we fire
			case schedule.Active:
				t.missedIntervals += sr.Missed()
				t.lastFireTime = time.Now()
				t.hitCount++
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
				if consecutiveFailures >= t.stopOnFailure {
					taskLogger.WithFields(log.Fields{
						"_block":               "spin",
						"task-id":              t.id,
						"task-name":            t.name,
						"consecutive failures": consecutiveFailures,
						"error":                t.lastFailureMessage,
					}).Error(ErrTaskDisabledOnFailures)
					// You must lock on state change for tasks
					t.Lock()
					t.state = core.TaskDisabled
					t.Unlock()
					// Send task disabled event
					event := new(scheduler_event.TaskDisabledEvent)
					event.TaskID = t.id
					event.Why = fmt.Sprintf("Task disabled with error: %s", t.lastFailureMessage)
					defer t.eventEmitter.Emit(event)
					return
				}
			// Schedule has ended
			case schedule.Ended:
				// You must lock task to change state
				t.Lock()
				t.state = core.TaskEnded
				t.Unlock()
				return //spin

			// Schedule has errored
			case schedule.Error:
				// You must lock task to change state
				t.Lock()
				t.state = core.TaskDisabled
				t.Unlock()
				return //spin

			}
		case <-t.killChan:
			// Only here can it truly be stopped
			t.Lock()
			t.state = core.TaskStopped
			t.lastFireTime = time.Time{}
			t.Unlock()
			return
		}
	}
}

func (t *task) fire() {
	t.Lock()
	defer t.Unlock()

	t.state = core.TaskFiring
	t.workflow.Start(t)
	t.state = core.TaskSpinning
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
		if task.state != core.TaskStopped {
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
