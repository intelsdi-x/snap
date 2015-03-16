package schedule

import (
	"errors"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

const (
	//Schedule states
	ScheduleActive ScheduleState = iota
	ScheduleEnded
	ScheduleError

	//Scheduler states
	SchedulerStopped SchedulerState = iota
	SchedulerStarted
)

var (
	MetricManagerNotSet = errors.New("MetricManager is not set.")
	SchedulerNotStarted = errors.New("Scheduler is not started.")
)

type managesWork interface {
	Work(job) job
}

// Schedule - Validate() will include ensure that the underlying schedule is
// still valid.  For example, it doesn't start in the past.
type Schedule interface {
	Wait(time.Time) ScheduleResponse
	Validate() error
}

type ScheduleState int

type ScheduleResponse interface {
	State() ScheduleState
	Error() error
	MissedIntervals() int
}

// ManagesMetric is implemented by control
// On startup a scheduler will be created and passed a reference to control
type managesMetric interface {
	SubscribeMetricType(mt core.MetricType, cd *cdata.ConfigDataNode) (core.MetricType, []error)
	UnsubscribeMetricType(mt core.MetricType)
}

type TaskErrors interface {
	Errors() []error
}

type taskErrors struct {
	errs []error
}

func (t *taskErrors) Errors() []error {
	return t.errs
}

type scheduler struct {
	workManager   *workManager
	metricManager managesMetric
	state         SchedulerState
}

type SchedulerState int

// CreateTask creates and returns task
func (scheduler *scheduler) CreateTask(mts []core.MetricType, s Schedule, cdt *cdata.ConfigDataTree, wf Workflow, opts ...option) (*Task, TaskErrors) {
	te := &taskErrors{
		errs: make([]error, 0),
	}

	if scheduler.state != SchedulerStarted {
		te.errs = append(te.errs, SchedulerNotStarted)
		return nil, te
	}

	//validate Schedule
	if err := s.Validate(); err != nil {
		te.errs = append(te.errs, err)
		return nil, te
	}

	//subscribe to MT
	//if we encounter an error we will unwind successful subscriptions
	subscriptions := make([]core.MetricType, 0)
	for _, m := range mts {
		cd := cdt.Get(m.Namespace())
		mt, err := scheduler.metricManager.SubscribeMetricType(m, cd)
		if err == nil {
			//mt := newMetricType(m, config)
			//mtc = append(mtc, mt)
			subscriptions = append(subscriptions, mt)
		} else {
			te.errs = append(te.errs, err...)
		}
	}

	if len(te.errs) > 0 {
		//unwind successful subscriptions
		for _, sub := range subscriptions {
			scheduler.metricManager.UnsubscribeMetricType(sub)
		}
		return nil, te
	}

	task := NewTask(s, subscriptions, wf, scheduler.workManager, opts...)

	return task, nil
}

// Start starts the scheduler
func (s *scheduler) Start() error {
	if s.metricManager == nil {
		return MetricManagerNotSet
	}
	s.state = SchedulerStarted
	return nil
}

func (s *scheduler) Stop() {
	s.state = SchedulerStopped
}

// Set metricManager for scheduler
func (s *scheduler) SetMetricManager(mm managesMetric) {
	s.metricManager = mm
}

// New returns an instance of the schduler
// The MetricManager must be set before the scheduler can be started.
// The MetricManager must be started before it can be used.

func New(poolSize, queueSize int) *scheduler {
	s := &scheduler{}

	s.workManager = newWorkManager(int64(queueSize), poolSize)
	s.workManager.Start()

	return s
}
