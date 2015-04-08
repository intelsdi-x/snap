package schedule

import (
	"time"

	"github.com/intelsdilabs/pulse/core"
)

type workflowState int

const (
	//Workflow states
	WorkflowStopped workflowState = iota
	WorkflowStarted
)

type Workflow interface {
	Start(task *task)
	State() workflowState
}

type workflow struct {
	rootStep *collectorStep
	state    workflowState
}

// NewWorkflow creates and returns a workflow
func NewWorkflow() *workflow {
	return &workflow{
		rootStep: new(collectorStep),
	}
}

// State returns current workflow state
func (w *workflow) State() workflowState {
	return w.state
}

// Start starts a workflow
func (w *workflow) Start(task *task) {
	w.state = WorkflowStarted
	j := w.rootStep.CreateJob(task.metricTypes, task.deadlineDuration)

	// dispatch 'collect' job to be worked
	j = task.manager.Work(j)

	//process through additional steps (processors, publishers, ...)
	for _, step := range w.rootStep.Steps() {
		w.processStep(step, j, task.manager)
	}
}

func (w *workflow) processStep(step Step, j job, m managesWork) {
	//do work for current step
	j = step.CreateJob(j)
	j = m.Work(j)
	//do work for child steps
	for _, step := range step.Steps() {
		w.processStep(step, j, m)
	}
}

// Step interface for a workflow step
type Step interface {
	Steps() []Step
	AddStep(s Step) Step
	CreateJob(j job) job
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

func (s *step) CreateJob(j job) job {
	//modifyJob for publish step and return
	return j
}

type ProcessStep interface {
	Step
}

type processStep struct {
	step
}

type PublishStep interface {
	Step
}

type publishStep struct {
	step
}

type CollectorStep interface {
}

type collectorStep struct {
	step
}

func (c *collectorStep) CreateJob(metricTypes []core.MetricType, deadlineDuration time.Duration) job {
	return newCollectorJob(metricTypes, deadlineDuration)
}
