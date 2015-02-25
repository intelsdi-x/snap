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

type ScheduleResponse struct {
	State                   ScheduleState
	Error                   error
	Waited                  time.Duration
	PreviousMissedIntervals int
}
