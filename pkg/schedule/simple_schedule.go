package schedule

import (
	"errors"
	"time"
)

// A schedule that only implements an endless repeating interval
type SimpleSchedule struct {
	Interval time.Duration
	State    ScheduleState
}

func NewSimpleSchedule(interval time.Duration) *SimpleSchedule {
	return &SimpleSchedule{
		Interval: interval,
	}
}

func (s *SimpleSchedule) GetState() ScheduleState {
	return s.State
}

func (s *SimpleSchedule) Validate() error {
	if s.Interval <= 0 {
		return errors.New("Simple Schedule interval must be greater than 0")
	}
	return nil
}

// Waits until net interval and returns true. Returning false signals a Schedule is no
// longer valid and should be halted. A SimpleSchedule has no end and as long as start
// is not in the future we will always in practice return true.
func (s *SimpleSchedule) Wait(last time.Time) Response {
	// Get the difference in time.Duration since last in nanoseconds (int64)
	timeDiff := time.Since(last).Nanoseconds()
	// cache our schedule interval in nanseconds
	nanoInterval := s.Interval.Nanoseconds()
	// use modulo operation to obtain the remainder of time over last interval
	remainder := timeDiff % nanoInterval
	// substract remainder from
	missed := (timeDiff - remainder) / nanoInterval // timeDiff.Nanoseconds() % s.Interval.Nanoseconds()
	waitDuration := nanoInterval - remainder
	// Wait until predicted interval fires
	time.Sleep(time.Duration(waitDuration))
	return SimpleScheduleResponse{state: s.GetState(), missed: uint(missed)}
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type SimpleScheduleResponse struct {
	// err    error
	state  ScheduleState
	missed uint
}

// Returns the state of the Schedule
func (s SimpleScheduleResponse) State() ScheduleState {
	return s.state
}

// Returns last error
func (s SimpleScheduleResponse) Error() error {
	return nil
}

// Returns any missed intervals
func (s SimpleScheduleResponse) Missed() uint {
	return s.missed
}
