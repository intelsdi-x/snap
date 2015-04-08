package core

type WorkflowState int

const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

type Workflow interface {
	Map()
	State() WorkflowState
}
