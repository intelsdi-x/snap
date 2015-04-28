package scheduler

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/intelsdilabs/pulse/core"
)

const (
	DefaultDeadlineDuration = time.Second * 5
)

type task struct {
	sync.Mutex //protects state

	id               uint64
	schResponseChan  chan scheduleResponse
	killChan         chan struct{}
	schedule         schedule
	workflow         workflow
	metricTypes      []core.Metric
	state            core.TaskState
	creationTime     time.Time
	lastFireTime     time.Time
	manager          managesWork
	metricsManager   managesMetric
	deadlineDuration time.Duration
	hitCount         uint
	missedIntervals  uint
}

type option func(t *task) option

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func TaskDeadlineDuration(v time.Duration) option {
	return func(t *task) option {
		previous := t.deadlineDuration
		t.deadlineDuration = v
		return TaskDeadlineDuration(previous)
	}
}

//NewTask creates a Task
func newTask(s schedule, mtc []core.Metric, wf workflow, m *workManager, mm managesMetric, opts ...core.TaskOption) *task {
	task := &task{
		id:               id(),
		schResponseChan:  make(chan scheduleResponse),
		killChan:         make(chan struct{}),
		metricTypes:      mtc,
		schedule:         s,
		state:            core.TaskStopped,
		creationTime:     time.Now(),
		workflow:         wf,
		manager:          m,
		metricsManager:   mm,
		deadlineDuration: DefaultDeadlineDuration,
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

// CreateTime returns the time the task was created.
func (t *task) CreationTime() time.Time {
	return t.creationTime
}

func (t *task) DeadlineDuration() time.Duration {
	return t.deadlineDuration
}

func (t *task) SetDeadlineDuration(d time.Duration) {
	t.deadlineDuration = d
}

// HitCount returns the number of times the task has fired.
func (t *task) HitCount() uint {
	return t.hitCount
}

// Id returns the tasks Id.
func (t *task) Id() uint64 {
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
func (t *task) State() core.TaskState {
	return t.state
}

// Status returns the state of the workflow.
func (t *task) Status() core.WorkflowState {
	return t.workflow.State()
}

// Spin will start a task spinning in its own routine while it waits for its
// schedule.
func (t *task) Spin() {
	// We need to lock long enough to change state
	t.Lock()
	defer t.Unlock()
	if t.state == core.TaskStopped {
		t.state = core.TaskSpinning
		// spin in a goroutine
		go t.spin()
	}
}

func (t *task) Stop() {
	t.Lock()
	defer t.Unlock()
	if t.state != core.TaskStopped {
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
			if sr.state() == core.ScheduleActive {
				t.missedIntervals += sr.missedIntervals()
				t.lastFireTime = time.Now()
				t.fire()
				t.hitCount++
			}
			// TODO stop task on schedule error state or end state
		case <-t.killChan:
			// Only here can it truly be stopped
			t.state = core.TaskStopped
			break
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
	t.schResponseChan <- t.schedule.Wait(t.lastFireTime)
}

type taskCollection struct {
	*sync.Mutex

	table map[uint64]*task
}

func newTaskCollection() *taskCollection {
	return &taskCollection{
		Mutex: &sync.Mutex{},

		table: make(map[uint64]*task),
	}
}

// Get given a task id returns a Task or nil if not found
func (t *taskCollection) Get(id uint64) *task {
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
		return errors.New(fmt.Sprintf("A task with Id '%v' has already been added.", task.id))
	}

	return nil
}

// Table returns a copy of the taskCollection
func (t *taskCollection) Table() map[uint64]*task {
	t.Lock()
	defer t.Unlock()
	tasks := make(map[uint64]*task)
	for id, t := range t.table {
		tasks[id] = t
	}
	return tasks
}

var idCounter uint64

// id generates the sequential next id (starting from 0)
func id() uint64 {
	return atomic.AddUint64(&idCounter, 1)
}
