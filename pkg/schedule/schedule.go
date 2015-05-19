package schedule

import (
	"time"
)

type ScheduleState int

const (
	// set by the scheduler once a schedule has been validated and is in use in a task
	ScheduleActive ScheduleState = iota
	// Schedule is ended
	ScheduleEnded
	// Schedule error state
	ScheduleError
)

type Schedule interface {
	GetState() ScheduleState
	Validate() error
	Wait(time.Time) ScheduleResponse
}

type ScheduleResponse interface {
	Error() error
	State() ScheduleState
	Missed() uint
}
