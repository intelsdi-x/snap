package schedule

import (
	"time"
)

// A schedule that only implements an endless repeating interval
type SimpleSchedule struct {
	Interval time.Duration
	state    ScheduleState
}

func NewSimpleSchedule(i time.Duration) *SimpleSchedule {
	return &SimpleSchedule{
		Interval: i,
	}
}

func (s *SimpleSchedule) GetState() ScheduleState {
	return s.state
}

func (s *SimpleSchedule) Validate() error {
	if s.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

// Waits until net interval and returns true. Returning false signals a Schedule is no
// longer valid and should be halted. A SimpleSchedule has no end and as long as start
// is not in the future we will always in practice return true.
func (s *SimpleSchedule) Wait(last time.Time) Response {
	m, t := waitOnInterval(last, s.Interval)
	return &SimpleScheduleResponse{state: s.GetState(), missed: m, lastTime: t}
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type SimpleScheduleResponse struct {
	state    ScheduleState
	missed   uint
	lastTime time.Time
}

// Returns the state of the Schedule
func (s *SimpleScheduleResponse) State() ScheduleState {
	return s.state
}

// Returns last error
func (s *SimpleScheduleResponse) Error() error {
	return nil
}

// Returns any missed intervals
func (s *SimpleScheduleResponse) Missed() uint {
	return s.missed
}

func (s *SimpleScheduleResponse) LastTime() time.Time {
	return s.lastTime
}
