package scheduler

import (
	"time"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/pkg/logger"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type WorkflowState int

const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

func renderWorkflow(wfMap wmap.WorkflowMap) (*schedulerWorkflow, error) {
	wf := &schedulerWorkflow{}
	return wf, nil
}

type schedulerWorkflow struct {
	state WorkflowState
}

// Start starts a workflow
func (w *schedulerWorkflow) Start(task *task) {
	w.state = WorkflowStarted
	// j := w.rootStep.createJob(task.metricTypes, task.deadlineDuration, task.metricsManager)

	// dispatch 'collect' job to be worked
	// j = task.manager.Work(j)

	//process through additional steps (processors, publishers, ...)
	// for _, step := range w.rootStep.Steps() {
	// w.processStep(step, j, task.manager, task.metricsManager)
	// }
}

func (s *schedulerWorkflow) State() WorkflowState {
	return s.state
}

// func (s *schedulerWorkflow) State() core.WorkflowState {
// 	return w.state
// }

// func (s *schedulerWorkflow) Marshal() ([]byte, error) {
// 	return []byte{}, nil
// }

// func (s *schedulerWorkflow) Unmarshal([]byte) error {
// 	return nil
// }

type wf struct {
	rootStep *collectStep
	state    core.WorkflowState
}

// NewWorkflow creates and returns a workflow
func newWorkflow() *wf {
	return &wf{
		rootStep: new(collectStep),
	}
}

// State returns current workflow state
func (w *wf) State() core.WorkflowState {
	return w.state
}

func (w *wf) Marshal() ([]byte, error) {
	return []byte{}, nil
}

func (w *wf) Unmarshal([]byte) error {
	return nil
}

func (w *wf) processStep(step Step, j job, m managesWork, metricManager managesMetric) {
	//do work for current step
	j = step.createJob(j, metricManager)
	j = m.Work(j)
	//do work for child steps
	for _, step := range step.Steps() {
		w.processStep(step, j, m, metricManager)
	}
}

// Step interface for a workflow step
type Step interface {
	Steps() []Step
	AddStep(s Step) Step
	createJob(job, managesMetric) job
}

type step struct {
	steps []Step
}

// AddStep adds a child Step
func (s *step) AddStep(step Step) Step {
	s.steps = append(s.steps, step)
	return step
}

// Steps returns child Steps
func (s *step) Steps() []Step {
	return s.steps
}

type ProcessStep interface {
	Step
}

type processStep struct {
	step
}

func (p *processStep) createJob(j job, metricManager managesMetric) job {
	return j
}

type PublishStep interface {
	Step
}

type publishStep struct {
	step
	name        string
	version     int
	config      map[string]ctypes.ConfigValue
	contentType string
}

func NewPublishStep(name string, version int, contentType string, config map[string]ctypes.ConfigValue) *publishStep {
	return &publishStep{
		name:        name,
		version:     version,
		config:      config,
		contentType: contentType,
	}
}

func (p *publishStep) createJob(j job, metricManager managesMetric) job {
	logger.Debugf("Scheduler.PublishStep.CreateJob", "creating job!")
	switch j.Type() {
	case collectJobType:
		return newPublishJob(j.(*collectorJob), p.name, p.version, p.contentType, p.config, metricManager.(publishesMetrics))
	default:
		panic("Unknown type of job")
	}
}

type CollectStep interface {
}

type collectStep struct {
	step
}

func (c *collectStep) createJob(metricTypes []core.Metric, deadlineDuration time.Duration, collector collectsMetrics) job {
	return newCollectorJob(metricTypes, deadlineDuration, collector)
}
