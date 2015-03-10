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
	Start(task *Task)
	State() workflowState
}

type workflow struct {
	workManager ManagesWork
	rootStep    *collectorStep
	state       workflowState
}

// NewWorkflow creates and returns a workflow
func NewWorkflow(workManager ManagesWork) *workflow {
	return &workflow{
		rootStep:    new(collectorStep),
		workManager: workManager,
	}
}

// State returns current workflow state
func (w *workflow) State() workflowState {
	return w.state
}

// Start starts a workflow
func (w *workflow) Start(task *Task) {
	w.state = WorkflowStarted
	deadline := time.Now().Add(task.deadlineDuration)
	job := w.rootStep.CreateJob(task.metricTypes, deadline)

	// dispatch 'collect' job to be worked
	job = w.workManager.Work(job)

	//process through additional steps (processors, publishers, ...)
	for _, step := range w.rootStep.Steps() {
		w.processStep(step, job)
	}
}

func (w *workflow) processStep(step Step, job Job) {
	//do work for current step
	job = step.CreateJob(job)
	job = w.workManager.Work(job)
	//do work for child steps
	for _, step := range step.Steps() {
		w.processStep(step, job)
	}
}

// Step interface for a workflow step
type Step interface {
	Steps() []Step
	AddStep(s Step) Step
	CreateJob(job Job) Job
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

func (s *step) CreateJob(j Job) Job {
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

func (c *collectorStep) CreateJob(metricTypes []core.MetricType, deadline time.Time) Job {
	return NewCollectorJob(metricTypes, deadline)
}
