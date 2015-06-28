package request

import (
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type TaskCreationRequest struct {
	Name     string            `json:"name"`
	Deadline string            `json:"deadline"`
	Workflow *wmap.WorkflowMap `json:"workflow"`
	Schedule Schedule          `json:"schedule"`
}

type Schedule struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
}
