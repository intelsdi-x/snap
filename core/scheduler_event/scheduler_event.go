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

package schedulerevent

import (
	"github.com/intelsdi-x/snap/core"
)

const (
	// TaskCreated represents the state of scheduler task created
	TaskCreated = "Scheduler.TaskCreated"
	// TaskDeleted represents the state of scheduler task deleted
	TaskDeleted = "Scheduler.TaskDeleted"
	// TaskStarted represents the state of scheduler task started
	TaskStarted = "Scheduler.TaskStarted"
	// TaskStopped represents the state of scheduler task stopped
	TaskStopped = "Scheduler.TaskStopped"
	// TaskDisabled represents the state of scheduler task disabled
	TaskDisabled = "Scheduler.TaskDisabled"
	// MetricCollected represents the state of scheduler task collected
	MetricCollected = "Scheduler.MetricsCollected"
	// MetricCollectionFailed represents the state of scheuler metric collection failed
	MetricCollectionFailed = "Scheduler.MetricCollectionFailed"
)

// TaskStartedEvent struct type describing task id and source
type TaskStartedEvent struct {
	TaskID string
	Source string
}

// Namespace returns TaskState string message
// after starting a task
func (e TaskStartedEvent) Namespace() string {
	return TaskStarted
}

// TaskCreatedEvent struct type describing the created task id,
// source and if start task on creation
type TaskCreatedEvent struct {
	TaskID        string
	StartOnCreate bool
	Source        string
}

// Namespace returns TaskCreated string message
// after creating a task
func (e TaskCreatedEvent) Namespace() string {
	return TaskCreated
}

// TaskDeletedEvent struct type describing deleted task id and source
type TaskDeletedEvent struct {
	TaskID string
	Source string
}

// Namespace returns TaskDeleted string message
// after deleting a task
func (e TaskDeletedEvent) Namespace() string {
	return TaskDeleted
}

// TaskStoppedEvent struct type describing stopped task id and source
type TaskStoppedEvent struct {
	TaskID string
	Source string
}

// Namespace returns TaskStopped string message
// after stopping a task
func (e TaskStoppedEvent) Namespace() string {
	return TaskStopped
}

// TaskDisabledEvent struct type describing disabled task id and reason
type TaskDisabledEvent struct {
	TaskID string
	Why    string
}

// Namespace returns TaskDisabled string message
// after disabling a task
func (e TaskDisabledEvent) Namespace() string {
	return TaskDisabled
}

// MetricCollectedEvent struct type describing
// the task id and collected metrics
type MetricCollectedEvent struct {
	TaskID  string
	Metrics []core.Metric
}

// Namespace returns MetricCollected string constant
// after collecting the metrics
func (e MetricCollectedEvent) Namespace() string {
	return MetricCollected
}

// MetricCollectionFailedEvent struct type
// describing failed metric collection task id
// and errors
type MetricCollectionFailedEvent struct {
	TaskID string
	Errors []error
}

// Namespace returns MetricCollectionFailed string message
// after the metric collection failure
func (e MetricCollectionFailedEvent) Namespace() string {
	return MetricCollectionFailed
}
