/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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
	"errors"
	"fmt"
	// "strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

var (
	// logger for the scheduler
	schedulerLogger = log.WithFields(log.Fields{
		"_module": "scheduler",
	})

	// HandlerRegistrationName registers a handler with the event manager
	HandlerRegistrationName = "scheduler"

	// ErrMetricManagerNotSet - The error message for metricManager is not set
	ErrMetricManagerNotSet = errors.New("MetricManager is not set.")
	// ErrSchedulerNotStarted - The error message for scheduler is not started
	ErrSchedulerNotStarted = errors.New("Scheduler is not started.")
	// ErrTaskAlreadyRunning - The error message for task is already running
	ErrTaskAlreadyRunning = errors.New("Task is already running.")
	// ErrTaskAlreadyStopped - The error message for task is already stopped
	ErrTaskAlreadyStopped = errors.New("Task is already stopped.")
	// ErrTaskDisabledNotRunnable - The error message for task is disabled and cannot be started
	ErrTaskDisabledNotRunnable = errors.New("Task is disabled. Cannot be started.")
)

type schedulerState int

const (
	schedulerStopped schedulerState = iota
	schedulerStarted
)

// ManagesMetric is implemented by control
// On startup a scheduler will be created and passed a reference to control
type managesMetrics interface {
	collectsMetrics
	publishesMetrics
	processesMetrics
	managesPluginContentTypes
	ValidateDeps([]core.Metric, *cdata.ConfigDataTree, []core.SubscribedPlugin) []serror.SnapError
	SubscribeDeps(string, []core.Metric, []core.Plugin) []serror.SnapError
	UnsubscribeDeps(string, []core.Metric, []core.Plugin) []serror.SnapError
}

// ManagesPluginContentTypes is an interface to a plugin manager that can tell us what content accept and returns are supported.
type managesPluginContentTypes interface {
	GetPluginContentTypes(n string, t core.PluginType, v int) ([]string, []string, error)
}

type collectsMetrics interface {
	CollectMetrics([]core.RequestedMetric, *cdata.ConfigDataTree, time.Time, string) ([]core.Metric, []error)
}

type publishesMetrics interface {
	PublishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error
}

type processesMetrics interface {
	ProcessMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) (string, []byte, []error)
}

type scheduler struct {
	workManager     *workManager
	metricManager   managesMetrics
	tasks           *taskCollection
	state           schedulerState
	eventManager    *gomit.EventController
	taskWatcherColl *taskWatcherCollection
}

type managesWork interface {
	Work(job) queuedJob
}

// New returns an instance of the scheduler
// The MetricManager must be set before the scheduler can be started.
// The MetricManager must be started before it can be used.
func New(cfg *Config) *scheduler {
	schedulerLogger.WithFields(log.Fields{
		"_block": "New",
		"value":  cfg.WorkManagerQueueSize,
	}).Info("Setting work manager queue size")
	schedulerLogger.WithFields(log.Fields{
		"_block": "New",
		"value":  cfg.WorkManagerPoolSize,
	}).Info("Setting work manager pool size")
	opts := []workManagerOption{
		CollectQSizeOption(cfg.WorkManagerQueueSize),
		CollectWkrSizeOption(cfg.WorkManagerPoolSize),
		PublishQSizeOption(cfg.WorkManagerQueueSize),
		PublishWkrSizeOption(cfg.WorkManagerPoolSize),
		ProcessQSizeOption(cfg.WorkManagerQueueSize),
		ProcessWkrSizeOption(cfg.WorkManagerPoolSize),
	}
	s := &scheduler{
		tasks:           newTaskCollection(),
		eventManager:    gomit.NewEventController(),
		taskWatcherColl: newTaskWatcherCollection(),
	}

	// we are setting the size of the queue and number of workers for
	// collect, process and publish consistently for now
	s.workManager = newWorkManager(opts...)
	s.workManager.Start()
	s.eventManager.RegisterHandler(HandlerRegistrationName, s)

	return s
}

type taskErrors struct {
	errs []serror.SnapError
}

func (t *taskErrors) Errors() []serror.SnapError {
	return t.errs
}

func (s *scheduler) Name() string {
	return "scheduler"
}

func (s *scheduler) RegisterEventHandler(name string, h gomit.Handler) error {
	return s.eventManager.RegisterHandler(name, h)
}

// CreateTask creates and returns task
func (s *scheduler) CreateTask(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	return s.createTask(sch, wfMap, startOnCreate, "user", opts...)
}

func (s *scheduler) CreateTaskTribe(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	return s.createTask(sch, wfMap, startOnCreate, "tribe", opts...)
}

func (s *scheduler) createTask(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, source string, opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	logger := schedulerLogger.WithFields(log.Fields{
		"_block": "create-task",
		"source": source,
	})
	// Create a container for task errors
	te := &taskErrors{
		errs: make([]serror.SnapError, 0),
	}

	// Return error if we are not started.
	if s.state != schedulerStarted {
		te.errs = append(te.errs, serror.New(ErrSchedulerNotStarted))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("scheduler not started")
		return nil, te
	}

	// Ensure the schedule is valid at this point and time.
	if err := sch.Validate(); err != nil {
		te.errs = append(te.errs, serror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("schedule passed not valid")
		return nil, te
	}

	// Generate a workflow from the workflow map
	wf, err := wmapToWorkflow(wfMap)
	if err != nil {
		te.errs = append(te.errs, serror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error(ErrSchedulerNotStarted.Error())
		return nil, te
	}

	// validate plugins and metrics
	mts, cdt, plugins := s.gatherMetricsAndPlugins(wf)
	errs := s.metricManager.ValidateDeps(mts, cdt, plugins)
	if len(errs) > 0 {
		te.errs = append(te.errs, errs...)
		return nil, te
	}

	// Bind plugin content type selections in workflow
	err = wf.BindPluginContentTypes(s.metricManager)
	if err != nil {
		te.errs = append(te.errs, serror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("unable to bind plugin content types")
		return nil, te
	}

	// Create the task object
	task := newTask(sch, wf, s.workManager, s.metricManager, s.eventManager, opts...)

	// Add task to taskCollection
	if err := s.tasks.add(task); err != nil {
		te.errs = append(te.errs, serror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("errors during task creation")
		return nil, te
	}

	logger.WithFields(log.Fields{
		"task-id":    task.ID(),
		"task-state": task.State(),
	}).Info("task created")

	event := &scheduler_event.TaskCreatedEvent{
		TaskID:        task.id,
		StartOnCreate: startOnCreate,
		Source:        source,
	}
	defer s.eventManager.Emit(event)

	if startOnCreate {
		logger.WithFields(log.Fields{
			"task-id": task.ID(),
			"source":  source,
		}).Info("starting task on creation")

		errs := s.StartTask(task.id)
		if errs != nil {
			te.errs = append(te.errs, errs...)
		}
	}

	return task, te
}

// RemoveTask given a tasks id.  The task must be stopped.
// Can return errors ErrTaskNotFound and ErrTaskNotStopped.
func (s *scheduler) RemoveTask(id string) error {
	return s.removeTask(id, "user")
}

func (s *scheduler) RemoveTaskTribe(id string) error {
	return s.removeTask(id, "tribe")
}

func (s *scheduler) removeTask(id, source string) error {
	logger := schedulerLogger.WithFields(log.Fields{
		"_block": "remove-task",
		"source": source,
	})
	t, err := s.getTask(id)
	if err != nil {
		logger.WithFields(log.Fields{
			"task id": id,
		}).Error(ErrTaskNotFound)
		return err
	}
	event := &scheduler_event.TaskDeletedEvent{
		TaskID: t.id,
		Source: source,
	}
	defer s.eventManager.Emit(event)
	return s.tasks.remove(t)
}

// GetTasks returns a copy of the tasks in a map where the task id is the key
func (s *scheduler) GetTasks() map[string]core.Task {
	tasks := make(map[string]core.Task)
	for id, t := range s.tasks.Table() {
		tasks[id] = t
	}
	return tasks
}

// GetTask provided the task id a task is returned
func (s *scheduler) GetTask(id string) (core.Task, error) {
	t, err := s.getTask(id)
	if err != nil {
		schedulerLogger.WithFields(log.Fields{
			"_block":  "get-task",
			"_error":  ErrTaskNotFound,
			"task-id": id,
		}).Error("error getting task")
		return nil, err // We do this to send back an explicit nil on the interface
	}
	return t, nil
}

// StartTask provided a task id a task is started
func (s *scheduler) StartTask(id string) []serror.SnapError {
	return s.startTask(id, "user")
}

func (s *scheduler) StartTaskTribe(id string) []serror.SnapError {
	return s.startTask(id, "tribe")
}

func (s *scheduler) startTask(id, source string) []serror.SnapError {
	logger := schedulerLogger.WithFields(log.Fields{
		"_block": "start-task",
		"source": source,
	})
	t, err := s.getTask(id)
	if err != nil {
		schedulerLogger.WithFields(log.Fields{
			"_block":  "start-task",
			"_error":  ErrTaskNotFound,
			"task-id": id,
		}).Error("error starting task")
		return []serror.SnapError{
			serror.New(err),
		}
	}

	if t.state == core.TaskDisabled {
		logger.WithFields(log.Fields{
			"task-id": t.ID(),
		}).Error("Task is disabled and must be enabled before starting")
		return []serror.SnapError{
			serror.New(ErrTaskDisabledNotRunnable),
		}
	}
	if t.state == core.TaskFiring || t.state == core.TaskSpinning {
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Info("task is already running")
		return []serror.SnapError{
			serror.New(ErrTaskAlreadyRunning),
		}
	}

	mts, _, plugins := s.gatherMetricsAndPlugins(t.workflow)
	cps := returnCorePlugin(plugins)
	serrs := s.metricManager.SubscribeDeps(t.ID(), mts, cps)
	if len(serrs) > 0 {
		// Tear down plugin processes started so far.
		uerrs := s.metricManager.UnsubscribeDeps(t.ID(), mts, cps)
		errs := append(serrs, uerrs...)
		logger.WithFields(log.Fields{
			"task-id": t.ID(),
			"_error":  errs,
		}).Error("task failed to start due to dependencies")
		return errs
	}

	event := &scheduler_event.TaskStartedEvent{
		TaskID: t.ID(),
		Source: source,
	}
	defer s.eventManager.Emit(event)
	t.Spin()
	logger.WithFields(log.Fields{
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task started")
	return nil
}

// StopTask provided a task id a task is stopped
func (s *scheduler) StopTask(id string) []serror.SnapError {
	return s.stopTask(id, "user")
}

func (s *scheduler) StopTaskTribe(id string) []serror.SnapError {
	return s.stopTask(id, "tribe")
}

func (s *scheduler) stopTask(id, source string) []serror.SnapError {
	logger := schedulerLogger.WithFields(log.Fields{
		"_block": "stop-task",
		"source": source,
	})
	t, err := s.getTask(id)
	if err != nil {
		logger.WithFields(log.Fields{
			"_error":  err.Error(),
			"task-id": id,
		}).Error("error stopping task")
		return []serror.SnapError{
			serror.New(err),
		}
	}

	if t.state == core.TaskStopped {
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Info("task is already stopped")
		return []serror.SnapError{
			serror.New(ErrTaskAlreadyStopped),
		}
	}

	mts, _, plugins := s.gatherMetricsAndPlugins(t.workflow)
	cps := returnCorePlugin(plugins)
	errs := s.metricManager.UnsubscribeDeps(t.ID(), mts, cps)
	if len(errs) > 0 {
		return errs
	}

	event := &scheduler_event.TaskStoppedEvent{
		TaskID: t.ID(),
		Source: source,
	}
	defer s.eventManager.Emit(event)
	t.Stop()
	logger.WithFields(log.Fields{
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task stopped")
	return nil
}

//EnableTask changes state from disabled to stopped
func (s *scheduler) EnableTask(id string) (core.Task, error) {
	t, e := s.getTask(id)
	if e != nil {
		schedulerLogger.WithFields(log.Fields{
			"_block":  "enable-task",
			"_error":  ErrTaskNotFound,
			"task-id": id,
		}).Error("error enabling task")
		return nil, e
	}

	err := t.Enable()
	if err != nil {
		schedulerLogger.WithFields(log.Fields{
			"_block":  "enable-task",
			"_error":  err.Error(),
			"task-id": id,
		}).Error("error enabling task")
		return nil, err
	}
	schedulerLogger.WithFields(log.Fields{
		"_block":     "enable-task",
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task enabled")
	return t, nil
}

// Start starts the scheduler
func (s *scheduler) Start() error {
	if s.metricManager == nil {
		schedulerLogger.WithFields(log.Fields{
			"_block": "start-scheduler",
			"_error": ErrMetricManagerNotSet.Error(),
		}).Error("error on scheduler start")
		return ErrMetricManagerNotSet
	}
	s.state = schedulerStarted
	schedulerLogger.WithFields(log.Fields{
		"_block": "start-scheduler",
	}).Info("scheduler started")
	return nil
}

func (s *scheduler) Stop() {
	s.state = schedulerStopped
	// stop all tasks that are not already stopped
	for _, t := range s.tasks.table {
		// Kill ensure another task can't turn it back on while we are shutting down
		t.Kill()
	}
	schedulerLogger.WithFields(log.Fields{
		"_block": "stop-scheduler",
	}).Info("scheduler stopped")
}

// Set metricManager for scheduler
func (s *scheduler) SetMetricManager(mm managesMetrics) {
	s.metricManager = mm
	schedulerLogger.WithFields(log.Fields{
		"_block": "set-metric-manager",
	}).Debug("metric manager linked")
}

//
func (s *scheduler) WatchTask(id string, tw core.TaskWatcherHandler) (core.TaskWatcherCloser, error) {
	task, err := s.getTask(id)
	if err != nil {
		schedulerLogger.WithFields(log.Fields{
			"_block":  "watch-task",
			"_error":  ErrTaskNotFound,
			"task-id": id,
		}).Error("error watching task")
		return nil, err
	}
	return s.taskWatcherColl.add(task.ID(), tw)
}

// Central handling for all async events in scheduler
func (s *scheduler) HandleGomitEvent(e gomit.Event) {

	switch v := e.Body.(type) {
	case *scheduler_event.MetricCollectedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
			"metric-count":    len(v.Metrics),
		}).Debug("event received")
		s.taskWatcherColl.handleMetricCollected(v.TaskID, v.Metrics)
	case *scheduler_event.MetricCollectionFailedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
			"errors-count":    v.Errors,
		}).Debug("event received")
	case *scheduler_event.TaskStartedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
		}).Debug("event received")
		s.taskWatcherColl.handleTaskStarted(v.TaskID)
	case *scheduler_event.TaskStoppedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
		}).Debug("event received")
		s.taskWatcherColl.handleTaskStopped(v.TaskID)
	case *scheduler_event.TaskDisabledEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
			"disabled-reason": v.Why,
		}).Debug("event received")
		// We need to unsubscribe from deps when a task goes disabled
		task, _ := s.getTask(v.TaskID)
		mts, _, plugins := s.gatherMetricsAndPlugins(task.workflow)
		cps := returnCorePlugin(plugins)
		s.metricManager.UnsubscribeDeps(task.ID(), mts, cps)
		s.taskWatcherColl.handleTaskDisabled(v.TaskID, v.Why)
	default:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
		}).Debug("event received")
	}
}

func (s *scheduler) getTask(id string) (*task, error) {
	task := s.tasks.Get(id)
	if task == nil {
		return nil, fmt.Errorf("%v: ID(%v)", ErrTaskNotFound, id)
	}
	return task, nil
}

func (s *scheduler) gatherMetricsAndPlugins(wf *schedulerWorkflow) ([]core.Metric, *cdata.ConfigDataTree, []core.SubscribedPlugin) {
	var (
		mts     []core.Metric
		plugins []core.SubscribedPlugin
	)

	log.WithFields(log.Fields{
		"_module": "scheduler",
		"_file":   "scheduler.go,",
		"_block":  "gather-metrics-and-plugins",
	}).Debug("gathering metrics and plugins from workflow")

	for _, mt := range wf.metrics {
		mts = append(mts, &metric{
			namespace: mt.Namespace(),
			version:   mt.Version(),
			config:    cdata.NewNode(),
		})
	}

	s.walkWorkflow(wf.processNodes, wf.publishNodes, &plugins)

	return mts, wf.configTree, plugins
}

func (s *scheduler) walkWorkflow(prnodes []*processNode, pbnodes []*publishNode, plugins *[]core.SubscribedPlugin) {
	for _, pr := range prnodes {
		*plugins = append(*plugins, pr)
		s.walkWorkflow(pr.ProcessNodes, pr.PublishNodes, plugins)
	}
	for _, pb := range pbnodes {
		*plugins = append(*plugins, pb)
	}
}

func returnCorePlugin(plugins []core.SubscribedPlugin) []core.Plugin {
	cps := make([]core.Plugin, len(plugins))
	for i, plugin := range plugins {
		cps[i] = plugin
	}
	return cps
}

func buildErrorsLog(errs []serror.SnapError, logger *log.Entry) *log.Entry {
	for i, e := range errs {
		logger = logger.WithField(fmt.Sprintf("%s[%d]", "error", i), e.Error())
	}
	return logger
}
