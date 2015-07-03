package scheduler

import (
	"errors"
	"fmt"
	// "strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/core/scheduler_event"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

var (
	HandlerRegistrationName  = "scheduler"
	schedulerEventController = gomit.NewEventController()

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
	SubscribeMetricType(mt core.RequestedMetric, cd *cdata.ConfigDataNode) (core.Metric, []perror.PulseError)
	UnsubscribeMetricType(mt core.Metric)
	SubscribeProcessor(name string, ver int, config map[string]ctypes.ConfigValue) []perror.PulseError
	SubscribePublisher(name string, ver int, config map[string]ctypes.ConfigValue) []perror.PulseError
	UnsubscribeProcessor(name string, ver int) error
	UnsubscribePublisher(name string, ver int) error
	Subscribe(mts []core.Metric, prs []core.SubscribedPlugin, pus []core.SubscribedPlugin) []perror.PulseError
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
		eventManager:    schedulerEventController,
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

// CreateTask creates and returns task
func (s *scheduler) CreateTask(sch schedule.Schedule, wfMap *wmap.WorkflowMap, opts ...core.TaskOption) (core.Task, core.TaskErrors) {
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
		te.errs = append(te.errs, perror.New(ErrSchedulerNotStarted))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error(ErrSchedulerNotStarted.Error())
		return nil, te
	}

	// Bind plugin content type selections in workflow
	err = wf.BindPluginContentTypes(s.metricManager)

	// create metric type subscription requests
	var mtsubs []core.Metric
	var subscriptions []core.Metric
	for _, m := range wf.metrics {
		cdt, er := wfMap.CollectNode.GetConfigTree()
		if er != nil {
			te.errs = append(te.errs, perror.New(er))
			continue
		}
		mtsubs = append(mtsubs, &metric{
			namespace: m.Namespace(),
			version:   m.Version(),
			config:    cdt.Get(m.Namespace()),
		})
	}

	var (
		pusubs []core.SubscribedPlugin
		prsubs []core.SubscribedPlugin
	)

	// Subscribe
	s.subscribe(wf.processNodes, wf.publishNodes, &prsubs, &pusubs)
	errs := s.metricManager.Subscribe(mtsubs, prsubs, pusubs)
	if len(errs) > 0 {
		te.errs = append(te.errs, errs...)
		return nil, te
	}

	// Create the task object
	task := newTask(sch, subscriptions, wf, s.workManager, s.metricManager, opts...)

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

	return task, te
}

// RemoveTask given a tasks id.  The task must be stopped.
// Can return errors ErrTaskNotFound and ErrTaskNotStopped.
func (s *scheduler) RemoveTask(id uint64) error {
	t := s.tasks.Get(id)
	if t == nil {
		log.WithFields(log.Fields{
			"module":  "scheduler",
			"block":   "RemoveTask",
			"task id": id,
		}).Error(ErrTaskNotFound)
		return fmt.Errorf("No task found with id '%v'", id)
	}
	return s.tasks.remove(t)
}

// GetTasks returns a copy of the tasks in a map where the task id is the key
func (s *scheduler) GetTasks() map[uint64]core.Task {
	tasks := make(map[uint64]core.Task)
	for id, t := range s.tasks.Table() {
		tasks[id] = t
	}
	return tasks
}

// GetTask provided the task id a task is returned
func (s *scheduler) GetTask(id uint64) (core.Task, error) {
	task := s.tasks.Get(id)
	if task == nil {
		return nil, fmt.Errorf("No task with Id '%v'", id)
	}
	return task, nil
}

// StartTask provided a task id a task is started
func (s *scheduler) StartTask(id uint64) error {
	t := s.tasks.Get(id)
	if t == nil {
		return fmt.Errorf("No task found with id '%v'", id)
	}
	t.Spin()
	s.logger.WithFields(log.Fields{
		"_block":     "start-task",
		"task-id":    t.ID(),
		"task-state": t.State(),
	}).Info("task started")
	return nil
}

// StopTask provided a task id a task is stopped
func (s *scheduler) StopTask(id uint64) error {
	t := s.tasks.Get(id)
	if t == nil {
		e := fmt.Errorf("No task found with id '%v'", id)
		s.logger.WithFields(log.Fields{
			"_block":  "stop-task",
			"_error":  e.Error(),
			"task-id": id,
		}).Warning("error on stopping of task")
		return e
	}
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
		if t.state > core.TaskStopped {
			t.Stop()
		}
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
func (s *scheduler) WatchTask(id uint64, tw core.TaskWatcherHandler) (core.TaskWatcherCloser, error) {
	for _, t := range s.tasks.table {
		if id == t.ID() {
			a, b := s.taskWatcherColl.add(id, tw)
			return a, b
		}
	}
	return nil, ErrTaskNotFound
}

// Central handling for all async events in scheduler
func (s *scheduler) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *scheduler_event.MetricCollectedEvent:
		// println(fmt.Sprintf("MetricCollectedEvent: %d\n", v.TaskID))
		s.taskWatcherColl.handleMetricCollected(v.TaskID, v.Metrics)
	case *scheduler_event.MetricCollectionFailedEvent:
		// println(fmt.Sprintf("MetricCollectionFailedEvent: %d\n", v.TaskID))
	default:
		log.WithFields(log.Fields{
			"_module": "scheduler",
			"_block":  "handle-events",
			"event":   v.Namespace(),
		}).Debug("Nothing to do for this event")
	}
}

// subscribe subscribes to all processors and publishers recursively
func (s *scheduler) subscribe(prnodes []*processNode, punodes []*publishNode, prsubs *[]core.SubscribedPlugin, pusubs *[]core.SubscribedPlugin) {
	for _, pr := range prnodes {
		*prsubs = append(*prsubs, pr)
		s.subscribe(pr.ProcessNodes, pr.PublishNodes, prsubs, pusubs)
	}
	for _, pu := range punodes {
		*pusubs = append(*pusubs, pu)
	}
}

func buildErrorsLog(errs []perror.PulseError, logger *log.Entry) *log.Entry {
	for i, e := range errs {
		logger = logger.WithField(fmt.Sprintf("%s[%d]", "error", i), e.Error())
	}
	return logger
}
