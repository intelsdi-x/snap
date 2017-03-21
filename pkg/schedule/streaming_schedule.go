/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package schedule

import "time"

// StreamingSchedule is a schedule that only implements an endless repeating interval
type StreamingSchedule struct {
	state ScheduleState
}

// NewStreamingSchedule returns the SimpleSchedule given the time interval
func NewStreamingSchedule() *StreamingSchedule {
	return &StreamingSchedule{}
}

// GetState returns the schedule state
func (s *StreamingSchedule) GetState() ScheduleState {
	return Active
}

// Validate returns an error if the interval of schedule is less
// or equals zero
func (s *StreamingSchedule) Validate() error {
	return nil
}

// Wait returns the StreamingSchedule state, misses and the last schedule ran
func (s *StreamingSchedule) Wait(last time.Time) Response {
	return &StreamingScheduleResponse{}
}

// StreamingScheduleResponse a response from SimpleSchedule conforming to ScheduleResponse interface
type StreamingScheduleResponse struct{}

// State returns the state of the Schedule
func (s *StreamingScheduleResponse) State() ScheduleState {
	return Active
}

// Error returns last error
func (s *StreamingScheduleResponse) Error() error {
	return nil
}

// Missed returns any missed intervals
func (s *StreamingScheduleResponse) Missed() uint {
	return 0
}

// LastTime returns the last response time
func (s *StreamingScheduleResponse) LastTime() time.Time {
	return time.Time{}
}
