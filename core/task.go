package core

import "time"

type TaskState int

const (
	//Task states
	TaskStopped TaskState = iota
	TaskSpinning
	TaskFiring
)

type Task interface {
	Id() uint64
	Status() WorkflowState
	State() TaskState
	HitCount() uint
	MissedCount() uint
	LastRunTime() time.Time
	CreationTime() time.Time
}

type TaskErrors interface {
	Errors() []error
}
