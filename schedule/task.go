package schedule

import (
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/core"
)

const (
	//Task states
	TaskStopped TaskState = iota
	TaskSpinning
	TaskFiring

	DefaultDeadlineDuration = time.Second * 5
)

type Task struct {
	schResponseChan chan ScheduleResponse
	killChan        chan struct{}
	schedule        Schedule
	workflow        Workflow
	metricTypes     []core.MetricType
	mu              sync.Mutex //protects state
	state           TaskState
	creationTime    time.Time
	lastFireTime    time.Time
	manager         managesWork
	deadlineDuration time.Duration
}

type TaskState int

type option func(t *Task) option

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (t *Task) Option(opts ...option) option {
	var previous option
	for _, opt := range opts {
		previous = opt(t)
	}
	return previous
}

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func TaskDeadlineDuration(v time.Duration) option {
	return func(t *Task) option {
		previous := t.deadlineDuration
		t.deadlineDuration = v
		return TaskDeadlineDuration(previous)
	}
}

//NewTask creates a Task
func NewTask(s Schedule, mtc []core.MetricType, wf Workflow, opts ...option) *Task {
	task := &Task{
		schResponseChan:  make(chan ScheduleResponse),
		killChan:         make(chan struct{}),
		metricTypes:      mtc,
		schedule:         s,
		state:            TaskStopped,
		creationTime:     time.Now(),
		workflow:         wf,
		manager:         manager,
		deadlineDuration: DefaultDeadlineDuration,
	}
	//set options
	for _, opt := range opts {
		opt(task)
	}
	return task
}

func (t *Task) Spin() {
	// We need to lock long enough to change state
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == TaskStopped {
		t.state = TaskSpinning
		// spin in a goroutine
		go t.spin()
	}
}

func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != TaskStopped {
		t.killChan <- struct{}{}
	}
}

func (t *Task) spin() {
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
				t.lastFireTime = time.Now()
				t.fire()
			}
			// TODO stop task on schedule error state or end state
		case <-t.killChan:
			// Only here can it truly be stopped
			t.state = TaskStopped
			break
		}
	}
}

func (t *Task) fire() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state = TaskFiring
	t.workflow.Start(t)
	t.state = TaskSpinning
}

func (t *Task) waitForSchedule() {
	t.schResponseChan <- t.schedule.Wait(t.lastFireTime)
}
