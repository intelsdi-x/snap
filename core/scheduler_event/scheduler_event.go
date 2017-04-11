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

package scheduler_event

import (
	"github.com/intelsdi-x/snap/core"
)

const (
	PluginsUnsubscribed    = "Scheduler.PluginUnsubscribed"
	TaskCreated            = "Scheduler.TaskCreated"
	TaskDeleted            = "Scheduler.TaskDeleted"
	TaskStarted            = "Scheduler.TaskStarted"
	TaskStopped            = "Scheduler.TaskStopped"
	TaskEnded              = "Scheduler.TaskEnded"
	TaskDisabled           = "Scheduler.TaskDisabled"
	MetricCollected        = "Scheduler.MetricsCollected"
	MetricCollectionFailed = "Scheduler.MetricCollectionFailed"
)

type PluginsUnsubscribedEvent struct {
	TaskID  string
	Plugins []core.SubscribedPlugin
}

func (e PluginsUnsubscribedEvent) Namespace() string {
	return PluginsUnsubscribed
}

type TaskStartedEvent struct {
	TaskID string
	Source string
}

func (e TaskStartedEvent) Namespace() string {
	return TaskStarted
}

type TaskCreatedEvent struct {
	TaskID        string
	StartOnCreate bool
	Source        string
}

func (e TaskCreatedEvent) Namespace() string {
	return TaskCreated
}

type TaskDeletedEvent struct {
	TaskID string
	Source string
}

func (e TaskDeletedEvent) Namespace() string {
	return TaskDeleted
}

type TaskStoppedEvent struct {
	TaskID string
	Source string
}

func (e TaskStoppedEvent) Namespace() string {
	return TaskStopped
}

type TaskEndedEvent struct {
	TaskID string
	Source string
}

func (e TaskEndedEvent) Namespace() string {
	return TaskEnded
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
