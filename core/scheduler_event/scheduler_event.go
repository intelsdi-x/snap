package scheduler_event

import (
	"github.com/intelsdi-x/pulse/core"
)

const (
	TaskCreated            = "Control.TaskCreated"
	TaskDeleted            = "Control.TaskDeleted"
	TaskStarted            = "Control.TaskStarted"
	TaskStopped            = "Control.TaskStopped"
	TaskDisabled           = "Control.TaskDisabled"
	MetricCollected        = "Control.MetricsCollected"
	MetricCollectionFailed = "Control.MetricCollectionFailed"
)

type TaskStartedEvent struct {
	TaskID uint64
}

func (e TaskStartedEvent) Namespace() string {
	return TaskStarted
}

type TaskCreatedEvent struct {
	TaskID        uint64
	StartOnCreate bool
}

func (e TaskCreatedEvent) Namespace() string {
	return TaskCreated
}

type TaskDeletedEvent struct {
	TaskID uint64
}

func (e TaskDeletedEvent) Namespace() string {
	return TaskDeleted
}

type TaskStoppedEvent struct {
	TaskID uint64
}

func (e TaskStoppedEvent) Namespace() string {
	return TaskStopped
}

type TaskDisabledEvent struct {
	TaskID uint64
	Why    string
}

func (e TaskDisabledEvent) Namespace() string {
	return TaskDisabled
}

type MetricCollectedEvent struct {
	TaskID  uint64
	Metrics []core.Metric
}

func (e MetricCollectedEvent) Namespace() string {
	return MetricCollected
}

type MetricCollectionFailedEvent struct {
	TaskID uint64
	Errors []error
}

func (e MetricCollectionFailedEvent) Namespace() string {
	return MetricCollectionFailed
}
