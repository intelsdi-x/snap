/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

	"github.com/robfig/cron"
)

// ErrMissingCronEntry indicates missing cron entry
var ErrMissingCronEntry = errors.New("Cron entry is missing")

// CronSchedule is a schedule that waits as long as specified in cron entry
type CronSchedule struct {
	entry    string
	enabled  bool
	state    ScheduleState
	schedule *cron.Cron
}

// NewCronSchedule creates and starts new cron schedule and returns an instance of CronSchedule
func NewCronSchedule(entry string) *CronSchedule {
	schedule := cron.New()
	return &CronSchedule{
		entry:    entry,
		schedule: schedule,
		enabled:  false,
	}
}

// Entry returns the cron schedule entry
func (c *CronSchedule) Entry() string {
	return c.entry
}

// GetState returns state of CronSchedule
func (c *CronSchedule) GetState() ScheduleState {
	return c.state
}

// Validate returns error if cron entry dosn't match crontab format
func (c *CronSchedule) Validate() error {
	if c.entry == "" {
		return ErrMissingCronEntry
	}
	_, err := cron.Parse(c.entry)
	if err != nil {
		return err
	}
	return nil
}

// Wait waits as long as specified in cron entry
func (c *CronSchedule) Wait(last time.Time) Response {
	var err error
	now := time.Now()

	// first run
	if (last == time.Time{}) {
		last = now
	}
	// schedule not enabled, either due to first run or invalid cron entry
	if !c.enabled {
		err = c.schedule.AddFunc(c.entry, func() {})
		if err != nil {
			c.state = Error
		} else {
			c.enabled = true
		}
	}

	var misses uint
	if c.enabled {
		s := c.schedule.Entries()[0].Schedule

		// calculate misses
		for next := last; next.Before(now); {
			next = s.Next(next)
			if next.After(now) {
				break
			}
			misses++
		}

		// wait
		waitTime := s.Next(now)
		time.Sleep(waitTime.Sub(now))
	}

	return &CronScheduleResponse{
		state:    c.GetState(),
		err:      err,
		missed:   misses,
		lastTime: time.Now(),
	}
}

// CronScheduleResponse is the response from CronSchedule
type CronScheduleResponse struct {
	state    ScheduleState
	err      error
	missed   uint
	lastTime time.Time
}

// State returns the state of the Schedule
func (c *CronScheduleResponse) State() ScheduleState {
	return c.state
}

// Error returns last error
func (c *CronScheduleResponse) Error() error {
	return c.err
}

// Missed returns any missed intervals
func (c *CronScheduleResponse) Missed() uint {
	return c.missed
}

// LastTime returns the last response time
func (c *CronScheduleResponse) LastTime() time.Time {
	return c.lastTime
}
