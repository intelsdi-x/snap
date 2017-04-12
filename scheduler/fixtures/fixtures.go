// + build medium legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package fixtures

import "github.com/intelsdi-x/gomit"
import "github.com/intelsdi-x/snap/core/scheduler_event"

type listenToSchedulerEvent struct {
	Ended                    chan struct{}
	UnsubscribedPluginEvents chan *scheduler_event.PluginsUnsubscribedEvent
	TaskStoppedEvents        chan struct{}
}

// NewListenToSchedulerEvent
func NewListenToSchedulerEvent() *listenToSchedulerEvent {
	return &listenToSchedulerEvent{
		Ended: make(chan struct{}),
		UnsubscribedPluginEvents: make(chan *scheduler_event.PluginsUnsubscribedEvent),
		TaskStoppedEvents:        make(chan struct{}),
	}
}

func (l *listenToSchedulerEvent) HandleGomitEvent(e gomit.Event) {
	switch msg := e.Body.(type) {
	case *scheduler_event.TaskEndedEvent:
		l.Ended <- struct{}{}
	case *scheduler_event.PluginsUnsubscribedEvent:
		l.UnsubscribedPluginEvents <- msg
	case *scheduler_event.TaskStoppedEvent:
		l.TaskStoppedEvents <- struct{}{}
	}
}
