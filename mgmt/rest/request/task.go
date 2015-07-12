package request

import (
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type TaskCreationRequest struct {
	Name     string            `json:"name"`
	Deadline string            `json:"deadline"`
	Workflow *wmap.WorkflowMap `json:"workflow"`
	Schedule Schedule          `json:"schedule"`
	Start    bool              `json:"start"`
}

type Schedule struct {
	Type           string `json:"type,omitempty"`
	Interval       string `json:"interval,omitempty"`
	StartTimestamp *int64 `json:"start_timestamp,omitempty"`
	StopTimestamp  *int64 `json:"stop_timestamp,omitempty"`
}
