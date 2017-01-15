package schedule

import (
	"time"
)

// SimpleSchedule is a schedule that only implements an endless repeating interval
type SimpleSchedule struct {
	Interval time.Duration
	state    ScheduleState
}

// NewSimpleSchedule returns the SimpleSchedule given the time interval
func NewSimpleSchedule(i time.Duration) *SimpleSchedule {
	return &SimpleSchedule{
		Interval: i,
	}
}

// GetState returns the schedule state
func (s *SimpleSchedule) GetState() ScheduleState {
	return s.state
}

// Validate returns an error if the interval of schedule is less
// or equals zero
func (s *SimpleSchedule) Validate() error {
	if s.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

// Wait returns the SimpleSchedule state, misses and the last schedule ran
func (s *SimpleSchedule) Wait(last time.Time) Response {
	m, t := waitOnInterval(last, s.Interval)
	return &SimpleScheduleResponse{state: s.GetState(), missed: m, lastTime: t}
}

// SimpleScheduleResponse a response from SimpleSchedule conforming to ScheduleResponse interface
type SimpleScheduleResponse struct {
	state    ScheduleState
	missed   uint
	lastTime time.Time
}

// State returns the state of the Schedule
func (s *SimpleScheduleResponse) State() ScheduleState {
	return s.state
}

// Error returns last error
func (s *SimpleScheduleResponse) Error() error {
	return nil
}

// Missed returns any missed intervals
func (s *SimpleScheduleResponse) Missed() uint {
	return s.missed
}

// LastTime returns the last response time
func (s *SimpleScheduleResponse) LastTime() time.Time {
	return s.lastTime
}
