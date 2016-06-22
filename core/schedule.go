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
	"time"

	"github.com/intelsdi-x/snap/pkg/schedule"
)

type Schedule struct {
	Type           string `json:"type,omitempty"`
	Interval       string `json:"interval,omitempty"`
	StartTimestamp *int64 `json:"start_timestamp,omitempty"`
	StopTimestamp  *int64 `json:"stop_timestamp,omitempty"`
}

func makeSchedule(s Schedule) (schedule.Schedule, error) {
	switch s.Type {
	case "simple":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}
		sch := schedule.NewSimpleSchedule(d)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "windowed":
		d, err := time.ParseDuration(s.Interval)
		if err != nil {
			return nil, err
		}

		var start, stop *time.Time
		if s.StartTimestamp != nil {
			t := time.Unix(*s.StartTimestamp, 0)
			start = &t
		}
		if s.StopTimestamp != nil {
			t := time.Unix(*s.StopTimestamp, 0)
			stop = &t
		}
		sch := schedule.NewWindowedSchedule(
			d,
			start,
			stop,
		)

		err = sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	case "cron":
		if s.Interval == "" {
			return nil, errors.New("missing cron entry")
		}
		sch := schedule.NewCronSchedule(s.Interval)

		err := sch.Validate()
		if err != nil {
			return nil, err
		}
		return sch, nil
	default:
		return nil, errors.New("unknown schedule type " + s.Type)
	}
}
