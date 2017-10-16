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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/snap/control/plugin"
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
	// ErrTaskDisabledNotStoppable - The error message for when a task is disabled and cannot be stopped
	ErrTaskDisabledNotStoppable = errors.New("Task is disabled. Only running tasks can be stopped.")
	// ErrTaskEndedNotStoppable - The error message for when a task is ended and cannot be stopped
	ErrTaskEndedNotStoppable = errors.New("Task is ended. Only running tasks can be stopped.")
	// ErrPluginIncompatibleWithScheduleType - The error message for when a streaming schedule type references a non streaming plugin or vice versa.
	ErrPluginIncompatibleWithScheduleType = errors.New("Plugin is incompatible with the tasks schedule type.")
	// ErrMultipleStreamingPlugins - The error message when a task with a streaming schedule refers to multiple streaming plugins.
	ErrMultipleStreamingPlugins = errors.New("Multiple streaming plugins within the same task is not supported.")
)

type schedulerState int

const (
	schedulerStopped schedulerState = iota
	schedulerStarted
)

type depGroupMap map[string]struct {
	requestedMetrics  []core.RequestedMetric
	subscribedPlugins []core.SubscribedPlugin
}

func newDepGroup() depGroupMap {
	return depGroupMap{}
}

// ManagesMetric is implemented by control
// On startup a scheduler will be created and passed a reference to control
type managesMetrics interface {
	collectsMetrics
	streamsMetrics
	publishesMetrics
	processesMetrics
	GetAutodiscoverPaths() []string
	ValidateDeps([]core.RequestedMetric, []core.SubscribedPlugin, *cdata.ConfigDataTree, ...core.SubscribedPluginAssert) []serror.SnapError
	SubscribeDeps(string, []core.RequestedMetric, []core.SubscribedPlugin, *cdata.ConfigDataTree) []serror.SnapError
	UnsubscribeDeps(string) []serror.SnapError
}

type collectsMetrics interface {
	CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error)
}

type streamsMetrics interface {
	StreamMetrics(string, map[string]map[string]string, time.Duration, int64) (chan []core.Metric, chan error, []error)
}

type publishesMetrics interface {
	PublishMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) []error
}

type processesMetrics interface {
	ProcessMetrics([]core.Metric, map[string]ctypes.ConfigValue, string, string, int) ([]core.Metric, []error)
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

// Implemented as a separate function so that defer calls
// are properly handled and cleanup done properly
// (like for the removal of temporary Yaml to JSON conversion)
func autoDiscoverTasks(taskFiles []os.FileInfo, fullPath string,
	fp func(sch schedule.Schedule,
		wfMap *wmap.WorkflowMap,
		startOnCreate bool,
		opts ...core.TaskOption) (core.Task, core.TaskErrors)) {
	// Note that the list of files is sorted by name due to ioutil.ReadDir
	// default behaviour. See go doc ioutil.ReadDir
	for _, file := range taskFiles {
		f, err := os.Open(path.Join(fullPath, file.Name()))
		if err != nil {
			log.WithFields(log.Fields{
				"_block":           "autoDiscoverTasks",
				"_module":          "scheduler",
				"autodiscoverpath": fullPath,
				"task":             file.Name(),
			}).Error("Opening file ", err)
			continue
		}
		defer f.Close()
		if !strings.HasSuffix(file.Name(), ".json") {
			fc, err := ioutil.ReadAll(f)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "autoDiscoverTasks",
						"_module":          "scheduler",
						"autodiscoverpath": fullPath,
						"task":             file.Name(),
					}).Error("Reading Yaml file ", err)
				continue
			}
			js, err := yaml.YAMLToJSON(fc)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "autoDiscoverTasks",
						"_module":          "scheduler",
						"autodiscoverpath": fullPath,
						"task":             file.Name(),
					}).Error("Parsing Yaml file ", err)
				continue
			}
			tfile, err := ioutil.TempFile(os.TempDir(), "yaml2json")
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "autoDiscoverTasks",
						"_module":          "scheduler",
						"autodiscoverpath": fullPath,
						"task":             file.Name(),
					}).Error("Creating temporary file ", err)
				continue
			}
			defer os.Remove(tfile.Name())
			err = ioutil.WriteFile(tfile.Name(), js, 0644)
			if err != nil {
				log.WithFields(
					log.Fields{
						"_block":           "autoDiscoverTasks",
						"_module":          "scheduler",
						"autodiscoverpath": fullPath,
						"task":             file.Name(),
					}).Error("Writing JSON file from Yaml ", err)
				continue
			}
			f, err = os.Open(tfile.Name())
			if err != nil {
				log.WithFields(log.Fields{
					"_block":           "autoDiscoverTasks",
					"_module":          "scheduler",
					"autodiscoverpath": fullPath,
					"task":             file.Name(),
				}).Error("Opening temporary file ", err)
				continue
			}
			defer f.Close()
		}
		mode := true
		task, err := core.CreateTaskFromContent(f, &mode, fp)
		if err != nil {
			log.WithFields(log.Fields{
				"_block":           "autoDiscoverTasks",
				"_module":          "scheduler",
				"autodiscoverpath": fullPath,
				"task":             file.Name(),
			}).Error(err)
			continue
		}
		//TODO: see if the following is really mandatory
		//in which case mgmt/rest/response/task.go contents might also
		//move into pkg/task
		//response.AddSchedulerTaskFromTask(task)
		log.WithFields(log.Fields{
			"_block":           "autoDiscoverTasks",
			"_module":          "scheduler",
			"autodiscoverpath": fullPath,
			"task-file-name":   file.Name(),
			"task-ID":          task.ID(),
		}).Info("Loading task")
	}
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
		"_block":          "create-task",
		"source":          source,
		"start-on-create": startOnCreate,
	})
	// Create a container for task errors
	te := &taskErrors{
		errs: make([]serror.SnapError, 0),
	}

	// Return error if we are not started.
	if s.state != schedulerStarted {
		te.errs = append(te.errs, serror.New(ErrSchedulerNotStarted))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error(ErrSchedulerNotStarted.Error())
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
		f.Error("Unable to generate workflow from workflow map")
		return nil, te
	}

	// Create the task object
	task, err := newTask(sch, wf, s.workManager, s.metricManager, s.eventManager, opts...)
	if err != nil {
		te.errs = append(te.errs, serror.New(err))
		f := buildErrorsLog(te.Errors(), logger)
		f.Error("Unable to create task")
		return nil, te
	}

	// subscribedPluginAsserts includes rules that need to be evaluated once we
	// have mapped the metrics to specific collector plugins.  Examples include
	// asserting that streaming tasks don't reference non-streaming collectors.
	subscribedPluginAsserts := []core.SubscribedPluginAssert{}
	// Group dependencies by the node they live on
	// and validate them.
	depGroups := getWorkflowPlugins(wf.processNodes, wf.publishNodes, wf.metrics)
	for k, group := range depGroups {

		// populate subscribedPluginAsserts
		switch sch.(type) {
		case *schedule.StreamingSchedule:
			// assert no non-streaming plugins
			subscribedPluginAsserts = append(subscribedPluginAsserts, func(plugins []core.SubscribedPlugin) serror.SnapError {
				for _, plg := range plugins {
					if plg.TypeName() != plugin.StreamCollectorPluginType.String() {
						return serror.New(
							ErrPluginIncompatibleWithScheduleType,
							map[string]interface{}{
								"schedule_type": fmt.Sprintf("%T", sch),
								"plugin_name":   plg.Name(),
								"plugin_type":   plg.TypeName(),
							},
						)
					}
				}
				return nil
			})
			// assert only a single streaming plugin
			subscribedPluginAsserts = append(subscribedPluginAsserts, func(plugins []core.SubscribedPlugin) serror.SnapError {
				if len(plugins) > 1 {
					return serror.New(
						ErrMultipleStreamingPlugins,
						map[string]interface{}{
							"schedule_type":     fmt.Sprintf("%T", sch),
							"num_of_collectors": len(plugins),
						},
					)
				}
				return nil
			})
		default:
			// assert no streaming plugins
			subscribedPluginAsserts = append(subscribedPluginAsserts, func(plugins []core.SubscribedPlugin) serror.SnapError {
				for _, plg := range plugins {
					if plg.TypeName() == plugin.StreamCollectorPluginType.String() {
						return serror.New(
							ErrPluginIncompatibleWithScheduleType,
							map[string]interface{}{
								"schedule_type": fmt.Sprintf("%T", sch),
								"plugin_name":   plg.Name(),
								"plugin_type":   plg.TypeName(),
							},
						)
					}
				}
				return nil
			})
		}

		manager, err := task.RemoteManagers.Get(k)
		if err != nil {
			te.errs = append(te.errs, serror.New(err))
			return nil, te
		}
		var errs []serror.SnapError
		errs = manager.ValidateDeps(group.requestedMetrics, group.subscribedPlugins, wf.configTree, subscribedPluginAsserts...)

		if len(errs) > 0 {
			te.errs = append(te.errs, errs...)
			return nil, te
		}
	}

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

	// Ensure the schedule is valid at this point and time.
	if err := t.schedule.Validate(); err != nil {
		errs := []serror.SnapError{
			serror.New(err),
		}
		f := buildErrorsLog(errs, logger)
		f.Error("schedule passed not valid")
		return errs
	}

	// subscribe plugins to task
	if _, err := t.SubscribePlugins(); len(err) != 0 {
		return err
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

	switch t.state {
	case core.TaskStopped:
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Error("task is already stopped")
		return []serror.SnapError{
			serror.New(ErrTaskAlreadyStopped),
		}
	case core.TaskEnded:
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Error("task is already ended")
		return []serror.SnapError{
			serror.New(ErrTaskEndedNotStoppable),
		}
	case core.TaskDisabled:
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Error("invalid action (stop) called on disabled task")
		return []serror.SnapError{
			serror.New(ErrTaskDisabledNotStoppable),
		}
	default:
		t.Stop()
		logger.WithFields(log.Fields{
			"task-id":    t.ID(),
			"task-state": t.State(),
		}).Info("task stopped")
	}

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

	//Autodiscover
	autoDiscoverPaths := s.metricManager.GetAutodiscoverPaths()
	if autoDiscoverPaths != nil && len(autoDiscoverPaths) != 0 {
		schedulerLogger.WithFields(log.Fields{
			"_block": "start-scheduler",
		}).Info("auto discover path is enabled")
		for _, pa := range autoDiscoverPaths {
			fullPath, err := filepath.Abs(pa)
			if err != nil {
				schedulerLogger.WithFields(log.Fields{
					"_block":           "start-scheduler",
					"autodiscoverpath": pa,
				}).Fatal(err)
			}
			schedulerLogger.WithFields(log.Fields{
				"_block": "start-scheduler",
			}).Info("autoloading tasks from: ", fullPath)
			files, err := ioutil.ReadDir(fullPath)
			if err != nil {
				schedulerLogger.WithFields(log.Fields{
					"_block":           "start-scheduler",
					"autodiscoverpath": pa,
				}).Fatal(err)
			}
			var taskFiles []os.FileInfo
			for _, file := range files {
				if file.IsDir() {
					schedulerLogger.WithFields(log.Fields{
						"_block":           "start-scheduler",
						"autodiscoverpath": pa,
					}).Warning("Ignoring subdirectory: ", file.Name())
					continue
				}
				// tasks files (JSON and YAML)
				fname := strings.ToLower(file.Name())
				if !strings.HasSuffix(fname, ".json") && !strings.HasSuffix(fname, ".yaml") && !strings.HasSuffix(fname, ".yml") {
					continue
				}
				taskFiles = append(taskFiles, file)
			}
			autoDiscoverTasks(taskFiles, fullPath, s.CreateTask)
		}
	} else {
		schedulerLogger.WithFields(log.Fields{
			"_block": "start-scheduler",
		}).Info("auto discover path is disabled")
	}

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
		// We need to unsubscribe from deps when a task has stopped
		task, _ := s.getTask(v.TaskID)
		task.UnsubscribePlugins()
		s.taskWatcherColl.handleTaskStopped(v.TaskID)
	case *scheduler_event.TaskEndedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
		}).Debug("event received")
		// We need to unsubscribe from deps when a task has ended
		task, _ := s.getTask(v.TaskID)
		task.UnsubscribePlugins()
		s.taskWatcherColl.handleTaskEnded(v.TaskID)
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
		task.UnsubscribePlugins()
		s.taskWatcherColl.handleTaskDisabled(v.TaskID, v.Why)
	case *scheduler_event.PluginsUnsubscribedEvent:
		log.WithFields(log.Fields{
			"_module":         "scheduler-events",
			"_block":          "handle-events",
			"event-namespace": e.Namespace(),
			"task-id":         v.TaskID,
		}).Debug("event received")
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

func getWorkflowPlugins(prnodes []*processNode, pbnodes []*publishNode, requestedMetrics []core.RequestedMetric) depGroupMap {
	depGroup := depGroupMap{}
	// Add metrics to depGroup map under local host(signified by empty string)
	// for now since remote collection not supported
	depGroup[""] = struct {
		requestedMetrics  []core.RequestedMetric
		subscribedPlugins []core.SubscribedPlugin
	}{requestedMetrics: requestedMetrics,
		subscribedPlugins: nil}
	return walkWorkflowForDeps(prnodes, pbnodes, requestedMetrics, depGroup)
}

func walkWorkflowForDeps(prnodes []*processNode, pbnodes []*publishNode, requestedMetrics []core.RequestedMetric, depGroup depGroupMap) depGroupMap {
	for _, pr := range prnodes {
		processors := depGroup[pr.Target]
		if _, ok := depGroup[pr.Target]; ok {
			processors.subscribedPlugins = append(processors.subscribedPlugins, pr)
		} else {
			processors.subscribedPlugins = []core.SubscribedPlugin{pr}
		}
		depGroup[pr.Target] = processors
		walkWorkflowForDeps(pr.ProcessNodes, pr.PublishNodes, requestedMetrics, depGroup)
	}
	for _, pb := range pbnodes {
		publishers := depGroup[pb.Target]
		if _, ok := depGroup[pb.Target]; ok {
			publishers.subscribedPlugins = append(publishers.subscribedPlugins, pb)
		} else {
			publishers.subscribedPlugins = []core.SubscribedPlugin{pb}
		}
		depGroup[pb.Target] = publishers
	}
	return depGroup
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
