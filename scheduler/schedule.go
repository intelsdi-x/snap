package scheduler

import (
	"errors"
	"time"

	"github.com/intelsdilabs/pulse/core"
)

type schedule interface {
	core.Schedule

	Wait(time.Time) scheduleResponse
}

func assertSchedule(sched core.Schedule) (schedule, error) {
	switch val := sched.(type) {
	case *core.SimpleSchedule:
		return newSimpleSchedule(val), nil
	}
	return nil, errors.New("unknown schedule type")
}

type scheduleResponse interface {
	err() error
	state() core.ScheduleState
	missedIntervals() uint
}
