package schedule

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/intelsdilabs/pulse/core"
)

const (
	//Task states
	TaskStopped taskState = iota
	TaskSpinning
	TaskFiring

	DefaultDeadlineDuration = time.Second * 5
)

type Task interface {
	Id() string
	Status() workflowState
	State() taskState
	HitCount() uint
	MissedCount() uint
	LastRunTime() time.Time
	CreationTime() time.Time
}

type task struct {
	id               string
	schResponseChan  chan ScheduleResponse
	killChan         chan struct{}
	schedule         Schedule
	workflow         Workflow
	metricTypes      []core.MetricType
	mu               sync.Mutex //protects state
	state            taskState
	creationTime     time.Time
	lastFireTime     time.Time
	manager          managesWork
	deadlineDuration time.Duration
	hitCount         uint
	missedIntervals  uint
}

type taskState int

type option func(t *task) option

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (t *task) option(opts ...option) option {
	var previous option
	for _, opt := range opts {
		previous = opt(t)
	}
	return previous
}

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func taskDeadlineDuration(v time.Duration) option {
	return func(t *task) option {
		previous := t.deadlineDuration
		t.deadlineDuration = v
		return taskDeadlineDuration(previous)
	}
}

//NewTask creates a Task
func newTask(s Schedule, mtc []core.MetricType, wf Workflow, m *workManager, opts ...option) *task {
	task := &task{
		id:               uuid.New(),
		schResponseChan:  make(chan ScheduleResponse),
		killChan:         make(chan struct{}),
		metricTypes:      mtc,
		schedule:         s,
		state:            TaskStopped,
		creationTime:     time.Now(),
		workflow:         wf,
		manager:          m,
		deadlineDuration: DefaultDeadlineDuration,
	}
	//set options
	for _, opt := range opts {
		opt(task)
	}
	return task
}

// CreateTime returns the time the task was created.
func (t *task) CreationTime() time.Time {
	return t.creationTime
}

// HitCount returns the number of times the task has fired.
func (t *task) HitCount() uint {
	return t.hitCount
}

// Id returns the tasks Id.
func (t *task) Id() string {
	return t.id
}

// LastRunTime returns the time of the tasks last run.
func (t *task) LastRunTime() time.Time {
	return t.lastFireTime
}

// MissedCount retruns the number of intervals missed.
func (t *task) MissedCount() uint {
	return t.missedIntervals
}

// State returns state of the task.
func (t *task) State() taskState {
	return t.state
}

// Status returns the state of the workflow.
func (t *task) Status() workflowState {
	return t.workflow.State()
}

// Spin will start a task spinning in its own routine while it waits for its
// schedule.
func (t *task) Spin() {
	// We need to lock long enough to change state
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == TaskStopped {
		t.state = TaskSpinning
		// spin in a goroutine
		go t.spin()
	}
}

func (t *task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != TaskStopped {
		t.killChan <- struct{}{}
	}
}

func (t *task) spin() {
	for {
		// Start go routine to wait on schedule
		go t.waitForSchedule()
		// wait here on
		//  schResponseChan - response from schedule
		//  killChan - signals task needs to be stopped
		select {
		case sr := <-t.schResponseChan:
			// If response show this schedule is stil active we fire
			if sr.State() == ScheduleActive {
				t.missedIntervals += sr.MissedIntervals()
				t.lastFireTime = time.Now()
				t.fire()
				t.hitCount++
			}
			// TODO stop task on schedule error state or end state
		case <-t.killChan:
			// Only here can it truly be stopped
			t.state = TaskStopped
			break
		}
	}
}

func (t *task) fire() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state = TaskFiring
	t.workflow.Start(t)
	t.state = TaskSpinning
}

func (t *task) waitForSchedule() {
	t.schResponseChan <- t.schedule.Wait(t.lastFireTime)
}

type taskCollection struct {
	*sync.Mutex
	table map[string]Task
}

func newTaskCollection() *taskCollection {
	return &taskCollection{
		table: make(map[string]Task),
		Mutex: &sync.Mutex{},
	}
}

// Get given a task id returns a Task or nil if not found
func (t *taskCollection) Get(id string) Task {
	t.Lock()
	defer t.Unlock()

	if t, ok := t.table[id]; ok {
		return t
	}
	return nil
}

// Add given a reference to a task adds it to the collection of tasks.  An
// error is returned if the task alredy exists in the collection.
func (t *taskCollection) add(task *task) error {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.table[task.id]; !ok {
		//If we don't already have this task in the collection save it
		t.table[task.id] = task
	} else {
		return errors.New(fmt.Sprintf("A task with Id '%s' has already been added.", task.id))
	}

	return nil
}

// Table returns a copy of the taskCollection
func (t *taskCollection) Table() map[string]Task {
	t.Lock()
	defer t.Unlock()
	tasks := make(map[string]Task)
	for k, v := range t.table {
		tasks[k] = v
	}
	return tasks
}
