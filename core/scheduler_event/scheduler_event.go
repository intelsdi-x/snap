package scheduler_event

import (
	"github.com/intelsdi-x/pulse/core"
)

const (
	TaskCreated            = "Scheduler.TaskCreated"
	TaskDeleted            = "Scheduler.TaskDeleted"
	TaskStarted            = "Scheduler.TaskStarted"
	TaskStopped            = "Scheduler.TaskStopped"
	TaskDisabled           = "Scheduler.TaskDisabled"
	MetricCollected        = "Scheduler.MetricsCollected"
	MetricCollectionFailed = "Scheduler.MetricCollectionFailed"
)

type TaskStartedEvent struct {
	TaskID string
}

func (e TaskStartedEvent) Namespace() string {
	return TaskStarted
}

type TaskCreatedEvent struct {
	TaskID        string
	StartOnCreate bool
}

func (e TaskCreatedEvent) Namespace() string {
	return TaskCreated
}

type TaskDeletedEvent struct {
	TaskID string
}

func (e TaskDeletedEvent) Namespace() string {
	return TaskDeleted
}

type TaskStoppedEvent struct {
	TaskID string
}

func (e TaskStoppedEvent) Namespace() string {
	return TaskStopped
}

type TaskDisabledEvent struct {
	TaskID string
	Why    string
}

func (e TaskDisabledEvent) Namespace() string {
	return TaskDisabled
}

type MetricCollectedEvent struct {
	TaskID  string
	Metrics []core.Metric
}

func (e MetricCollectedEvent) Namespace() string {
	return MetricCollected
}

type MetricCollectionFailedEvent struct {
	TaskID string
	Errors []error
}

func (e MetricCollectionFailedEvent) Namespace() string {
	return MetricCollectionFailed
}
