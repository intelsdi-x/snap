package scheduler

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/core"
)

var (
	watcherLog = log.WithField("_module", "scheduler-watcher")
)

type TaskWatcher struct {
	id      uint64
	taskIds []uint64
	parent  *taskWatcherCollection
	stopped bool
	handler core.TaskWatcherHandler
}

// Stops watching a task. Cannot be restarted.
func (t *TaskWatcher) Close() error {
	for _, x := range t.taskIds {
		t.parent.rm(x, t)
	}
	return nil
}

type taskWatcherCollection struct {
	// Collection of task watchers by
	coll       map[uint64][]*TaskWatcher
	tIdCounter uint64
	mutex      *sync.Mutex
}

func newTaskWatcherCollection() *taskWatcherCollection {
	return &taskWatcherCollection{
		coll:       make(map[uint64][]*TaskWatcher),
		tIdCounter: 1,
		mutex:      &sync.Mutex{},
	}
}

func (t *taskWatcherCollection) rm(taskId uint64, tw *TaskWatcher) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.coll[taskId] != nil {
		for i, w := range t.coll[taskId] {
			if w == tw {
				watcherLog.WithFields(log.Fields{
					"task-id":         taskId,
					"task-watcher-id": tw.id,
				}).Debug("removing watch from task")
				t.coll[taskId] = append(t.coll[taskId][:i], t.coll[taskId][i+1:]...)
				if len(t.coll[taskId]) == 0 {
					delete(t.coll, taskId)
				}
			}
		}
	}
}

func (t *taskWatcherCollection) add(taskId uint64, twh core.TaskWatcherHandler) (*TaskWatcher, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// init map for task ID if it does not eist
	if t.coll[taskId] == nil {
		t.coll[taskId] = make([]*TaskWatcher, 0)
	}
	tw := &TaskWatcher{
		// Assign unique ID to task watcher
		id: t.tIdCounter,
		// Add ref to coll for cleanup later
		parent:  t,
		stopped: false,
		handler: twh,
	}
	// Increment number for next time
	t.tIdCounter++
	// Add task id to task watcher list
	tw.taskIds = append(tw.taskIds, taskId)
	// Add this task watcher in
	t.coll[taskId] = append(t.coll[taskId], tw)
	watcherLog.WithFields(log.Fields{
		"task-id":         taskId,
		"task-watcher-id": tw.id,
	}).Debug("Added to task watcher collection")
	return tw, nil
}

func (t *taskWatcherCollection) handleMetricCollected(taskId uint64, m []core.Metric) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskId] == nil || len(t.coll[taskId]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for i, v := range t.coll[taskId] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskId,
			"task-watcher-id": i,
		}).Debug("calling taskwatcher collection func")
		// Call the catcher
		v.handler.CatchCollection(m)
	}
}

func (t *taskWatcherCollection) handleTaskStarted(taskId uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskId] == nil || len(t.coll[taskId]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for i, v := range t.coll[taskId] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskId,
			"task-watcher-id": i,
		}).Debug("calling taskwatcher task started func")
		// Call the catcher
		v.handler.CatchTaskStarted()
	}
}

func (t *taskWatcherCollection) handleTaskStopped(taskId uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskId] == nil || len(t.coll[taskId]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for i, v := range t.coll[taskId] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskId,
			"task-watcher-id": i,
		}).Debug("calling taskwatcher task stopped func")
		// Call the catcher
		v.handler.CatchTaskStopped()
	}
}

func (t *taskWatcherCollection) handleTaskDisabled(taskId uint64, why string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// no taskID means no watches, early exit
	if t.coll[taskId] == nil || len(t.coll[taskId]) == 0 {
		// Uncomment this debug line if needed. Otherwise this is too verbose for even debug level.
		// watcherLog.WithFields(log.Fields{
		// 	"task-id": taskId,
		// }).Debug("no watchers")
		return
	}
	// Walk all watchers for a task ID
	for i, v := range t.coll[taskId] {
		// Check if they have a catcher assigned
		watcherLog.WithFields(log.Fields{
			"task-id":         taskId,
			"task-watcher-id": i,
		}).Debug("calling taskwatcher task disabled func")
		// Call the catcher
		v.handler.CatchTaskDisabled(why)
	}
}
