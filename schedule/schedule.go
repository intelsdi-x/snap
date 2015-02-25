package schedule

import (
	"time"
)

const (
	ScheduleActive ScheduleState = iota
	ScheduleEnded
	ScheduleError
)

type Schedule interface {
	Wait(time.Time) ScheduleResponse
}

type ScheduleState int

type ScheduleResponse interface {
	State() ScheduleState
	Error() error
	MissedIntervals() []time.Time
}
