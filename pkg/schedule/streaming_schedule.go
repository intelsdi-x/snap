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

// LastTime retruns the last response time
func (s *StreamingScheduleResponse) LastTime() time.Time {
	return time.Time{}
}
