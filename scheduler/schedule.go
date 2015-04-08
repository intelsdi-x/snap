package scheduler

import (
	"time"

	"github.com/intelsdilabs/pulse/core"
)

type scheduleState int

const (
	scheduleActive scheduleState = iota
	scheduleEnded
	scheduleError
)

// schedule - Validate() will include ensure that the underlying schedule is
// still valid.  For example, it doesn't start in the past.
type schedule interface {
	core.Schedule

	Wait(time.Time) scheduleResponse
}

type scheduleResponse interface {
	err() error
	state() scheduleState
	missedIntervals() uint
}
