package scheduler

import (
	"time"

	"github.com/intelsdilabs/pulse/core"
)

// A schedule that only implements an endless repeating interval
type simpleSchedule struct {
	*core.SimpleSchedule
}

func newSimpleSchedule(ss *core.SimpleSchedule) *simpleSchedule {
	s := &simpleSchedule{SimpleSchedule: ss}
	s.State = core.ScheduleActive
	return s
}

// Waits until net interval and returns true. Returning false signals a Schedule is no
// longer valid and should be halted. A SimpleSchedule has no end and as long as start
// is not in the future we will always in practice return true.
func (s *simpleSchedule) Wait(last time.Time) scheduleResponse {
	// Get the difference in time.Duration since last in nanoseconds (int64)
	timeDiff := time.Now().Sub(last).Nanoseconds()
	// cache our schedule interval in nanseconds
	nanoInterval := s.Interval.Nanoseconds()
	// use modulo operation to obtain the remainder of time over last interval
	remainder := timeDiff % nanoInterval
	// substract remainder from
	missed := (timeDiff - remainder) / nanoInterval // timeDiff.Nanoseconds() % s.Interval.Nanoseconds()
	waitDuration := nanoInterval - remainder
	// Wait until predicted interval fires
	time.Sleep(time.Duration(waitDuration))
	return simpleScheduleResponse{st: s.State, miss: uint(missed)}
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type simpleScheduleResponse struct {
	er   error
	st   core.ScheduleState
	miss uint
}

// Returns the state of the Schedule
func (s simpleScheduleResponse) state() core.ScheduleState {
	return s.st
}

// Returns last error
func (s simpleScheduleResponse) err() error {
	return s.er
}

// Returns any missed intervals
func (s simpleScheduleResponse) missedIntervals() uint {
	return s.miss
}
