package core

type WorkflowState int

const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

type Workflow interface {
	Map() WfMap
	State() WorkflowState
}

type WfMap struct {
	Collect CollectStep
	Process ProcessStep
	Publish PublishStep
}

type CollectStep struct {
	MetricTypes []MetricType
	Process     []ProcessStep
	Publish     []PublishStep
}

type ProcessStep struct {
	Plugin  Plugin
	Process []ProcessStep
	Publish []PublishStep
}

type PublishStep struct {
	Plugin Plugin
}
