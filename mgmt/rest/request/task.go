/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
