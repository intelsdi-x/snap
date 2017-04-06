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

package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/intelsdi-x/snap/pkg/schedule"
)

// Schedule defines a scheduler.
//
// swagger:model Schedule
type Schedule struct {
	// required: true
	// enum: simple, windowed, streaming, cron
	Type string `json:"type"`
	// required: true
	Interval       string     `json:"interval"`
	StartTimestamp *time.Time `json:"start_timestamp,omitempty"`
	StopTimestamp  *time.Time `json:"stop_timestamp,omitempty"`
	Count          uint       `json:"count,omitempty"`
}

var (
	ErrMissingScheduleInterval = errors.New("missing `interval` in configuration of schedule")
)

func makeSchedule(s Schedule) (schedule.Schedule, error) {
	switch s.Type {
	case "simple", "windowed":
		if s.Interval == "" {
			return nil, ErrMissingScheduleInterval
		}

		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}

		sch := schedule.NewWindowedSchedule(
			d,
			s.StartTimestamp,
			s.StopTimestamp,
			s.Count,
		)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "cron":
		if s.Interval == "" {
			return nil, ErrMissingScheduleInterval
		}
		sch := schedule.NewCronSchedule(s.Interval)

		err := sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "streaming":
		return schedule.NewStreamingSchedule(), nil
	default:
		return nil, fmt.Errorf("unknown schedule type `%s`", s.Type)
	}
}
