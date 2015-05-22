package schedule

import (
	"time"
)

type ScheduleState int

const (
	// set by the scheduler once a schedule has been validated and is in use in a task
	Active ScheduleState = iota
	// Schedule is ended
	Ended
	// Schedule is halted with an error
	Error
)

type Schedule interface {
	// Returns the current state of the schedule
	GetState() ScheduleState
	// Returns where a schedule is still valid for this point in time.
	Validate() error
	// Blocks until time to fire and returns a schedule.Response
	Wait(time.Time) Response
}

type Response interface {
	// Contains any errors captured during a schedule.Wait()
	Error() error
	// Returns the schedule state
	State() ScheduleState
	// Returns any intervals that were missed since the call to Wait()
	Missed() uint
}
