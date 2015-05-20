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
	// Schedule error state
	Error
)

type Schedule interface {
	GetState() ScheduleState
	Validate() error
	Wait(time.Time) Response
}

type Response interface {
	Error() error
	State() ScheduleState
	Missed() uint
}
