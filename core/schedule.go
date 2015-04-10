package core

import (
	"errors"
	"time"
)

type ScheduleState int

const (
	// a schedule has been initialized but not made active by the scheduler
	ScheduleInitialized ScheduleState = iota
	// set by the scheduler once a schedule has been validated and is in use in a task
	ScheduleActive
	// TODO (danielscottt): FUTURE USE - for cron / windowed schedule
	ScheduleEnded
	// Schedule error state
	ScheduleError
)

type Schedule interface {
	GetState() ScheduleState
	Validate() error
}

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
