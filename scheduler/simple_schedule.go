package scheduler

import (
	"errors"
	"time"
)

// A schedule that only implements an endless repeating interval
type SimpleSchedule struct {
	interval time.Duration
}

func NewSimpleSchedule(interval time.Duration) *SimpleSchedule {
	return &SimpleSchedule{interval}
}

// Waits until net interval and returns true. Returning false signals a Schedule is no
// longer valid and should be halted. A SimpleSchedule has no end and as long as start
// is not in the future we will always in practice return true.
func (s *SimpleSchedule) Wait(last time.Time) ScheduleResponse {
	// Get the difference in time.Duration since last in nanoseconds (int64)
	timeDiff := time.Now().Sub(last).Nanoseconds()
	// cache our schedule interval in nanseconds
	nanoInterval := s.interval.Nanoseconds()
	// use modulo operation to obtain the remainder of time over last interval
	remainder := timeDiff % nanoInterval
	// substract remainder from
	missed := (timeDiff - remainder) / nanoInterval // timeDiff.Nanoseconds() % s.interval.Nanoseconds()
	waitDuration := nanoInterval - remainder
	// Wait until predicted interval fires
	time.Sleep(time.Duration(waitDuration))
	return SimpleScheduleResponse{state: ScheduleActive, misses: uint(missed)}
}

func (s *SimpleSchedule) Validate() error {
	if s.interval <= 0 {
		return errors.New("Interval must be greater than 0.")
	}
	return nil
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type SimpleScheduleResponse struct {
	err    error
	state  ScheduleState
	misses uint
}

// Returns the state of the Schedule
func (s SimpleScheduleResponse) State() ScheduleState {
	return s.state
}

// Returns last error
func (s SimpleScheduleResponse) Error() error {
	return s.err
}

// Returns any missed intervals
func (s SimpleScheduleResponse) MissedIntervals() uint {
	return s.misses
}
