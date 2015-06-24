package request

import (
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type TaskCreationRequest struct {
	ID uint64 `json:"id"`
	// Config       map[string][]configItem `json:"config"`
	Name     string            `json:"name"`
	Deadline string            `json:"deadline"`
	Workflow *wmap.WorkflowMap `json:"workflow"`
	Schedule Schedule          `json:"schedule"`
	// CreationTime       *time.Time        `json:"creation_timestamp,omitempty"`
	// LastRunTime        *time.Time        `json:"last_run_timestamp,omitempty"`
	// HitCount           uint              `json:"hit_count,omitempty"`
	// MissCount          uint              `json:"miss_count,omitempty"`
	// FailedCount        uint              `json:"failed_count,omitempty"`
	// LastFailureMessage string            `json:"last_failure_message,omitempty"`
	// State              string            `json:"task_state"`
}

type Schedule struct {
	Type     string `json:"type"`
	Interval string `json:"interval"`
}
