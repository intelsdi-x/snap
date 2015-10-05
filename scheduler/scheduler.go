package scheduler

import (
	"errors"
	"fmt"
	// "strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/core/scheduler_event"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

var (
	HandlerRegistrationName = "scheduler"

	ErrMetricManagerNotSet = errors.New("MetricManager is not set.")
	ErrSchedulerNotStarted = errors.New("Scheduler is not started.")
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
	ValidateDeps([]core.Metric, []core.SubscribedPlugin) []perror.PulseError
	SubscribeDeps(string, []core.Metric, []core.Plugin) []perror.PulseError
	UnsubscribeDeps(string, []core.Metric, []core.Plugin) []perror.PulseError
}

// ManagesPluginContentTypes is an interface to a plugin manager that can tell us what content accept and returns are supported.
type managesPluginContentTypes interface {
	GetPluginContentTypes(n string, t core.PluginType, v int) ([]string, []string, error)
}

type collectsMetrics interface {
	CollectMetrics([]core.Metric, time.Time) ([]core.Metric, []error)
}

type publishesMetrics interface {
	PublishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue) []error
}

type processesMetrics interface {
	ProcessMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue) (string, []byte, []error)
}

type scheduler struct {
	workManager     *workManager
	metricManager   managesMetrics
	tasks           *taskCollection
	state           schedulerState
	logger          *log.Entry
	eventManager    *gomit.EventController
	taskWatcherColl *taskWatcherCollection
}

type managesWork interface {
	Work(job) job
}

// New returns an instance of the scheduler
// The MetricManager must be set before the scheduler can be started.
// The MetricManager must be started before it can be used.
func New(opts ...workManagerOption) *scheduler {
	s := &scheduler{
		tasks: newTaskCollection(),
		logger: log.WithFields(log.Fields{
			"_module": "scheduler",
		}),
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
	errs []perror.PulseError
}

func (t *taskErrors) Errors() []perror.PulseError {
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
	logger := s.logger.WithField("_block", "create-task")
	// Create a container for task errors
	te := &taskErrors{
		errs: make([]perror.PulseError, 0),
	}

	// Return error if we are not started.
	if s.state != schedulerStarted {
		te.errs = append(te.errs, perror.New(ErrSchedulerNotStarted))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("scheduler not started")
		return nil, te
	}

	// Ensure the schedule is valid at this point and time.
	if err := sch.Validate(); err != nil {
		te.errs = append(te.errs, perror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("schedule passed not valid")
		return nil, te
	}

	// Generate a workflow from the workflow map
	wf, err := wmapToWorkflow(wfMap)
	if err != nil {
		te.errs = append(te.errs, perror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error(ErrSchedulerNotStarted.Error())
		return nil, te
	}

	// Bind plugin content type selections in workflow
	err = wf.BindPluginContentTypes(s.metricManager)

	// validate plugins and metrics
	mts, plugins := s.gatherPlugins(wf)
	errs := s.metricManager.ValidateDeps(mts, plugins)
	if len(errs) > 0 {
		te.errs = append(te.errs, errs...)
		return nil, te
	}

	// Create the task object
	task := newTask(sch, wf, s.workManager, s.metricManager, s.eventManager, opts...)

	// Add task to taskCollection
	if err := s.tasks.add(task); err != nil {
		te.errs = append(te.errs, perror.New(err))
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
	}
	defer s.eventManager.Emit(event)

	if startOnCreate {
		logger.WithFields(log.Fields{
			"task-id": task.ID(),
		}).Info("starting task on creation")

		cps := make([]core.Plugin, len(plugins))
		for i, plugin := range plugins {
			cps[i] = plugin
		}
		errs := s.metricManager.SubscribeDeps(task.ID(), mts, cps)
		if len(errs) > 0 {
			te.errs = append(te.errs, errs...)
			return nil, te
		}

		task.Spin()
	}

	return task, te
}

// RemoveTask given a tasks id.  The task must be stopped.
// Can return errors ErrTaskNotFound and ErrTaskNotStopped.
func (s *scheduler) RemoveTask(id string) error {
	t := s.tasks.Get(id)
	if t == nil {
		log.WithFields(log.Fields{
			"_module": "scheduler",
			"block":   "RemoveTask",
			"task id": id,
		}).Error(ErrTaskNotFound)
		return fmt.Errorf("No task found with id '%v'", id)
	}
	event := &scheduler_event.TaskDeletedEvent{
		TaskID: t.id,
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
	task := s.tasks.Get(id)
	if task == nil {
		return nil, fmt.Errorf("No task with Id '%v'", id)
	}
	return task, nil
}

// StartTask provided a task id a task is started
func (s *scheduler) StartTask(id string) []perror.PulseError {
	t := s.tasks.Get(id)
	if t == nil {
		return []perror.PulseError{
			perror.New(fmt.Errorf("No task found with id '%v'", id)),
		}
	}

	mts, plugins := s.gatherPlugins(t.workflow)
	cps := make([]core.Plugin, len(plugins))
	for i, plugin := range plugins {
		cps[i] = plugin
	}
	errs := s.metricManager.SubscribeDeps(id, mts, cps)
	if len(errs) > 0 {
		return errs
	}

	event := new(scheduler_event.TaskStartedEvent)
	event.TaskID = t.ID()
	defer s.eventManager.Emit(event)
	t.Spin()
	s.logger.WithFields(log.Fields{
		"_block":     "start-task",
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task started")
	return nil
}

// StopTask provided a task id a task is stopped
func (s *scheduler) StopTask(id string) []perror.PulseError {
	t := s.tasks.Get(id)
	if t == nil {
		e := fmt.Errorf("No task found with id '%v'", id)
		s.logger.WithFields(log.Fields{
			"_block":  "stop-task",
			"_error":  e.Error(),
			"task-id": id,
		}).Warning("error on stopping of task")
		return []perror.PulseError{
			perror.New(e),
		}
	}

	mts, plugins := s.gatherPlugins(t.workflow)
	cps := make([]core.Plugin, len(plugins))
	for i, plugin := range plugins {
		cps[i] = plugin
	}
	errs := s.metricManager.UnsubscribeDeps(t.ID(), mts, cps)
	if len(errs) > 0 {
		return errs
	}

	event := new(scheduler_event.TaskStoppedEvent)
	event.TaskID = t.ID()
	defer s.eventManager.Emit(event)
	t.Stop()
	s.logger.WithFields(log.Fields{
		"_block":     "stop-task",
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task stopped")
	return nil
}

// Start starts the scheduler
func (s *scheduler) Start() error {
	if s.metricManager == nil {
		s.logger.WithFields(log.Fields{
			"_block": "start-scheduler",
			"_error": ErrMetricManagerNotSet.Error(),
		}).Error("error on scheduler start")
		return ErrMetricManagerNotSet
	}
	s.state = schedulerStarted
	s.logger.WithFields(log.Fields{
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
	s.logger.WithFields(log.Fields{
		"_block": "stop-scheduler",
	}).Info("scheduler stopped")
}

// Set metricManager for scheduler
func (s *scheduler) SetMetricManager(mm managesMetrics) {
	s.metricManager = mm
	s.logger.WithFields(log.Fields{
		"_block": "set-metric-manager",
	}).Debug("metric manager linked")
}

//
func (s *scheduler) WatchTask(id string, tw core.TaskWatcherHandler) (core.TaskWatcherCloser, error) {
	if task := s.tasks.Get(id); task != nil {
		a, b := s.taskWatcherColl.add(task.ID(), tw)
		return a, b
	}
	return nil, ErrTaskNotFound
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
		s.taskWatcherColl.handleTaskDisabled(v.TaskID, v.Why)
	default:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
		}).Debug("event received")
	}
}

func (s *scheduler) gatherPlugins(wf *schedulerWorkflow) ([]core.Metric, []core.SubscribedPlugin) {
	var (
		mts     []core.Metric
		plugins []core.SubscribedPlugin
	)

	for _, m := range wf.metrics {
		mts = append(mts, &metric{
			namespace: m.Namespace(),
			version:   m.Version(),
			config:    wf.configTree.Get(m.Namespace()),
		})
	}
	s.walkWorkflow(wf.processNodes, wf.publishNodes, &plugins)

	return mts, plugins
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

func buildErrorsLog(errs []perror.PulseError, logger *log.Entry) *log.Entry {
	for i, e := range errs {
		logger = logger.WithField(fmt.Sprintf("%s[%d]", "error", i), e.Error())
	}
	return logger
}
