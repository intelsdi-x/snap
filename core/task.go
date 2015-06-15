package core

import "time"

type TaskState int

const (
	TaskStopped TaskState = iota
	TaskSpinning
	TaskFiring
)

var (
	TaskStateLookup = map[TaskState]string{
		TaskStopped:  "Stopped",
		TaskSpinning: "Spinning",
		TaskFiring:   "Firing",
	}
)

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
	LastRunTime() *time.Time
	CreationTime() *time.Time
	DeadlineDuration() time.Duration
	SetDeadlineDuration(time.Duration)
	Option(...TaskOption) TaskOption
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
	Errors() []error
}
