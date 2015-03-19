package schedule

import (
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/pkg/logger"
)

const (
	//Task states
	TaskStopped TaskState = iota
	TaskSpinning
	TaskFiring

	DefaultDeadlineDuration = time.Second * 5
)

type Task struct {
	ID               string
	DeadlineDuration time.Duration
	Schedule         Schedule
	Workflow         Workflow
	MetricTypes      []core.MetricType
	CreationTime     time.Time
	LastFireTime     time.Time
	State            TaskState

	schResponseChan chan ScheduleResponse
	killChan        chan struct{}
	mu              sync.Mutex //protects state
	manager         managesWork
}

type TaskState int

type TaskOption func(t *Task) TaskOption

// Option sets the options specified.
// Returns an option to optionally restore the last arg's previous value.
func (t *Task) Option(opts ...TaskOption) TaskOption {
	var previous TaskOption
	for _, opt := range opts {
		previous = opt(t)
	}
	return previous
}

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func TaskDeadlineDuration(v time.Duration) TaskOption {
	return func(t *Task) TaskOption {
		previous := t.DeadlineDuration
		t.DeadlineDuration = v
		return TaskDeadlineDuration(previous)
	}
}

//NewTask creates a Task
func NewTask(s Schedule, mtc []core.MetricType, wf Workflow, m *workManager, opts ...TaskOption) *Task {
	task := &Task{
		ID:               uuid.New(),
		DeadlineDuration: DefaultDeadlineDuration,
		Schedule:         s,
		Workflow:         wf,
		MetricTypes:      mtc,
		CreationTime:     time.Now(),
		State:            TaskStopped,

		schResponseChan: make(chan ScheduleResponse),
		killChan:        make(chan struct{}),
		manager:         m,
	}
	//set options
	for _, opt := range opts {
		opt(task)
	}
	return task
}

func (t *Task) Spin() {
	// We need to lock long enough to change state
	logger.Debugf("scheduler", "%s spinning", t.ID)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.State == TaskStopped {
		t.State = TaskSpinning
		// spin in a goroutine
		go t.spin()
	}
}

func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.State != TaskStopped {
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
				t.LastFireTime = time.Now()
				t.fire()
			}
			// TODO stop task on schedule error state or end state
		case <-t.killChan:
			// Only here can it truly be stopped
			t.State = TaskStopped
			break
		}
	}
}

func (t *Task) fire() {
	logger.Debugf("scheduler", "%s firing", t.ID)
	t.mu.Lock()
	defer t.mu.Unlock()

	t.State = TaskFiring
	t.Workflow.Start(t)
	t.State = TaskSpinning
}

func (t *Task) waitForSchedule() {
	t.schResponseChan <- t.Schedule.Wait(t.LastFireTime)
}
