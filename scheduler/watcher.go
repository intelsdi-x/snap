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

package scheduler

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/core"
)

var (
	watcherLog = log.WithField("_module", "scheduler-watcher")
)

// TaskWatcher struct type
type TaskWatcher struct {
	id      uint64
	taskIDs []string
	parent  *taskWatcherCollection
	stopped bool
	handler core.TaskWatcherHandler
}

// Close stops watching a task. Cannot be restarted.
func (t *TaskWatcher) Close() error {
	for _, x := range t.taskIDs {
		t.parent.rm(x, t)
	}
	return nil
}

type taskWatcherCollection struct {
	// Collection of task watchers by
	coll       map[string][]*TaskWatcher
	tIDCounter uint64
	mutex      *sync.Mutex
}

func newTaskWatcherCollection() *taskWatcherCollection {
	return &taskWatcherCollection{
		coll:       make(map[string][]*TaskWatcher),
		tIDCounter: 1,
		mutex:      &sync.Mutex{},
	}
}

func (t *taskWatcherCollection) rm(taskID string, tw *TaskWatcher) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.coll[taskID] != nil {
		for i, w := range t.coll[taskID] {
			if w == tw {
				watcherLog.WithFields(log.Fields{
					"task-id":         taskID,
					"task-watcher-id": tw.id,
				}).Debug("removing watch from task")
				t.coll[taskID] = append(t.coll[taskID][:i], t.coll[taskID][i+1:]...)
				if len(t.coll[taskID]) == 0 {
					delete(t.coll, taskID)
				}
			}
		}
	}
}

func (t *taskWatcherCollection) add(taskID string, twh core.TaskWatcherHandler) (*TaskWatcher, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// init map for task ID if it does not eist
	if t.coll[taskID] == nil {
		t.coll[taskID] = make([]*TaskWatcher, 0)
	}
	tw := &TaskWatcher{
		// Assign unique ID to task watcher
		id: t.tIDCounter,
		// Add ref to coll for cleanup later
		parent:  t,
		stopped: false,
		handler: twh,
	}
	// Increment number for next time
	t.tIDCounter++
	// Add task id to task watcher list
	tw.taskIDs = append(tw.taskIDs, taskID)
	// Add this task watcher in
	t.coll[taskID] = append(t.coll[taskID], tw)
	watcherLog.WithFields(log.Fields{
		"task-id":         taskID,
		"task-watcher-id": tw.id,
	}).Debug("Added to task watcher collection")
	return tw, nil
}

func (t *taskWatcherCollection) handleMetricCollected(taskID string, m []core.Metric) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskID] == nil || len(t.coll[taskID]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for _, v := range t.coll[taskID] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskID,
			"task-watcher-id": v.id,
		}).Debug("calling taskwatcher collection func")
		// Call the catcher
		v.handler.CatchCollection(m)
	}
}

func (t *taskWatcherCollection) handleTaskStarted(taskID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskID] == nil || len(t.coll[taskID]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskID,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for _, v := range t.coll[taskID] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskID,
			"task-watcher-id": v.id,
		}).Debug("calling taskwatcher task started func")
		// Call the catcher
		v.handler.CatchTaskStarted()
	}
}

func (t *taskWatcherCollection) handleTaskStopped(taskID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskID] == nil || len(t.coll[taskID]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for _, v := range t.coll[taskID] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskID,
			"task-watcher-id": v.id,
		}).Debug("calling taskwatcher task stopped func")
		// Call the catcher
		v.handler.CatchTaskStopped()
	}
}

func (t *taskWatcherCollection) handleTaskDisabled(taskID string, why string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskID] == nil || len(t.coll[taskID]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskID,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for _, v := range t.coll[taskID] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskID,
			"task-watcher-id": v.id,
		}).Debug("calling taskwatcher task disabled func")
		// Call the catcher
		v.handler.CatchTaskDisabled(why)
	}
}
