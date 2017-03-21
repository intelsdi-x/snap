/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schedule

import (
	"errors"
	"time"
)

var (
	// ErrInvalidInterval - Error message for the valid schedule interval must be greater than 0
	ErrInvalidInterval = errors.New("Interval must be greater than 0")
	// ErrInvalidStopTime - Error message for the stop tome is in the past
	ErrInvalidStopTime = errors.New("Stop time is in the past")
	// ErrStopBeforeStart - Error message for the stop time cannot occur before start time
	ErrStopBeforeStart = errors.New("Stop time cannot occur before start time")
)

// ScheduleState int type
type ScheduleState int

const (
	// Active - set by the scheduler once a schedule has been validated and is in use in a task
	Active ScheduleState = iota
	// Ended - Schedule is ended
	Ended
	// Error - Schedule is halted with an error
	Error
)

// Schedule interface
type Schedule interface {
	// Returns the current state of the schedule
	GetState() ScheduleState
	// Returns where a schedule is still valid for this point in time.
	Validate() error
	// Blocks until time to fire and returns a schedule.Response
	Wait(time.Time) Response
}

// Response interface defines the behavior of schedule response
type Response interface {
	// Contains any errors captured during a schedule.Wait()
	Error() error
	// Returns the schedule state
	State() ScheduleState
	// Returns any intervals that were missed since the call to Wait()
	Missed() uint
	// The time the interval fired
	LastTime() time.Time
}

func waitOnInterval(last time.Time, i time.Duration) (uint, time.Time) {
	// first run
	if (last == time.Time{}) {
		// for the first run, do not wait on interval
		// and schedule workflow execution immediately
		return uint(0), time.Now()
	}
	// Get the difference in time.Duration since last in nanoseconds (int64)
	timeDiff := time.Since(last).Nanoseconds()
	// cache our schedule interval in nanoseconds
	nanoInterval := i.Nanoseconds()
	// use modulo operation to obtain the remainder of time over last interval
	remainder := timeDiff % nanoInterval
	// subtract remainder from
	missed := (timeDiff - remainder) / nanoInterval // timeDiff.Nanoseconds() % s.Interval.Nanoseconds()
	waitDuration := nanoInterval - remainder
	// Wait until predicted interval fires
	time.Sleep(time.Duration(waitDuration))
	return uint(missed), time.Now()
}
