package core

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type TaskState int

const (
	TaskDisabled TaskState = iota - 1
	TaskStopped
	TaskSpinning
	TaskFiring
)

var (
	TaskStateLookup = map[TaskState]string{
		TaskDisabled: "Disabled",
		TaskStopped:  "Stopped",
		TaskSpinning: "Spinning",
		TaskFiring:   "Firing",
	}
)

type TaskWatcherCloser interface {
	Close() error
}

type TaskWatcherHandler interface {
	CatchCollection([]Metric)
}

func (t TaskState) String() string {
	return TaskStateLookup[t]
}

type Task interface {
	ID() uint64
	// Status() WorkflowState TODO, switch to string
	State() TaskState
	HitCount() uint
	GetName() string
	SetName(string)
	MissedCount() uint
	FailedCount() uint
	LastFailureMessage() string
	LastRunTime() *time.Time
	CreationTime() *time.Time
	DeadlineDuration() time.Duration
	SetDeadlineDuration(time.Duration)
	SetStopOnFailure(uint)
	GetStopOnFailure() uint
	Option(...TaskOption) TaskOption
	WMap() *wmap.WorkflowMap
	Schedule() schedule.Schedule
}

type TaskOption func(Task) TaskOption

// TaskDeadlineDuration sets the tasks deadline.
// The deadline is the amount of time that can pass before a worker begins
// processing the tasks collect job.
func TaskDeadlineDuration(v time.Duration) TaskOption {
	return func(t Task) TaskOption {
		previous := t.DeadlineDuration()
		t.SetDeadlineDuration(v)
		return TaskDeadlineDuration(previous)
	}
}

// TaskStopOnFailure sets the tasks stopOnFailure
// The stopOnFailure is the number of consecutive task failures that will
// trigger disabling the task
func OptionStopOnFailure(v uint) TaskOption {
	return func(t Task) TaskOption {
		previous := t.GetStopOnFailure()
		t.SetStopOnFailure(v)
		log.WithFields(log.Fields{
			"_module":                   "core",
			"_block":                    "OptionStopOnFailure",
			"task-id":                   t.ID(),
			"task-name":                 t.GetName(),
			"consecutive failure limit": t.GetStopOnFailure(),
		}).Debug("Setting stop-on-failure limit for task")
		return OptionStopOnFailure(previous)
	}
}

//SetTaskName sets the name of the task.
//This is optional.
//If task name is not set, the task name is then defaulted to "Task-<task-id>"
func SetTaskName(name string) TaskOption {
	return func(t Task) TaskOption {
		previous := t.GetName()
		t.SetName(name)
		return SetTaskName(previous)
	}
}

type TaskErrors interface {
	Errors() []perror.PulseError
}
