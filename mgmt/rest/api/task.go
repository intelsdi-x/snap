package api

import (
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

type Tasks interface {
	CreateTask(schedule.Schedule, *wmap.WorkflowMap, bool, ...core.TaskOption) (core.Task, core.TaskErrors)
	GetTasks() map[string]core.Task
	GetTask(string) (core.Task, error)
	StartTask(string) []serror.SnapError
	StopTask(string) []serror.SnapError
	RemoveTask(string) error
	WatchTask(string, core.TaskWatcherHandler) (core.TaskWatcherCloser, error)
	EnableTask(string) (core.Task, error)
}
