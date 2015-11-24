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

package tribe

import (
	"sync"
	"time"

	"github.com/intelsdi-x/snap/core"
)

const (
	TaskStateQueryResponseSizeLimit int = 1024
)

type taskStateResponses []taskStateResponse

type taskStateResponse struct {
	From  string
	State core.TaskState
}

func (t taskStateResponses) State() core.TaskState {
	states := map[core.TaskState]int{}
	for _, r := range t {
		if _, ok := states[r.State]; !ok {
			states[r.State] = 1
		}
		states[r.State]++
	}
	state := struct {
		count int
		state core.TaskState
	}{
		count: 0,
		state: core.TaskStopped,
	}
	for k, v := range states {
		if v > state.count {
			state.count = v
			state.state = k
		}
	}
	return state.state
}

type taskStateQueryResponse struct {
	uuid     string
	ltime    LTime
	deadline time.Time
	isClosed bool
	from     map[string]core.TaskState
	resp     chan taskStateResponse
	lock     sync.Mutex
}

func newStateQueryResponse(n int, q *taskStateQueryMsg) *taskStateQueryResponse {
	return &taskStateQueryResponse{
		uuid:  q.UUID,
		ltime: q.LTime,
		from:  map[string]core.TaskState{},
		resp:  make(chan taskStateResponse, n),
	}
}

func (s *taskStateQueryResponse) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isClosed {
		return
	}
	if s.resp != nil {
		close(s.resp)
	}
}
