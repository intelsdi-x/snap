package core

type WorkflowState int

const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

type Workflow interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	State() WorkflowState
}
