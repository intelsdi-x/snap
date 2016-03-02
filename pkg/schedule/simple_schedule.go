package schedule

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

// SimpleSchedule is a schedule that only implements an endless repeating interval
type SimpleSchedule struct {
	Interval time.Duration `json:"interval"`
	state    ScheduleState
}

// NewSimpleSchedule returns the SimpleSchedule given the time interval
func NewSimpleSchedule(i time.Duration) *SimpleSchedule {
	return &SimpleSchedule{
		Interval: i,
	}
}

func (s *SimpleSchedule) UnmarshalJSON(data []byte) error {
	t := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&t); err != nil {
		return err
	}

	if v, ok := t["interval"]; ok {
		switch typ := v.(type) {
		case string:
			dur, err := time.ParseDuration(typ)
			if err != nil {
				return err
			}
			s.Interval = dur
		default:
			return errors.New("Unsupported interval value")
		}
	}
	return nil
}

func (s *SimpleSchedule) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Interval string `json:"interval"`
		Type     string `json:"type"`
	}{
		Interval: s.Interval.String(),
		Type:     "simple",
	})
}

// GetState returns the schedule state
func (s *SimpleSchedule) GetState() ScheduleState {
	return s.state
}

// Validate returns an error if the interval of schedule is less
// or equals zero
func (s *SimpleSchedule) Validate() error {
	if s.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

// Wait returns the SimpleSchedule state, misses and the last schedule ran
func (s *SimpleSchedule) Wait(last time.Time) Response {
	m, t := waitOnInterval(last, s.Interval)
	return &SimpleScheduleResponse{state: s.GetState(), missed: m, lastTime: t}
}

// SimpleScheduleResponse a response from SimpleSchedule conforming to ScheduleResponse interface
type SimpleScheduleResponse struct {
	state    ScheduleState
	missed   uint
	lastTime time.Time
}

// State returns the state of the Schedule
func (s *SimpleScheduleResponse) State() ScheduleState {
	return s.state
}

// Error returns last error
func (s *SimpleScheduleResponse) Error() error {
	return nil
}

// Missed returns any missed intervals
func (s *SimpleScheduleResponse) Missed() uint {
	return s.missed
}

// LastTime retruns the last response time
func (s *SimpleScheduleResponse) LastTime() time.Time {
	return s.lastTime
}
