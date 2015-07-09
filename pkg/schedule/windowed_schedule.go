package schedule

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

// A schedule that waits on an interval within a specific time window
type WindowedSchedule struct {
	Interval  time.Duration
	StartTime time.Time
	StopTime  time.Time
	state     ScheduleState
}

func NewWindowedSchedule(i time.Duration, start time.Time, stop time.Time) *WindowedSchedule {
	return &WindowedSchedule{
		Interval:  i,
		StartTime: start,
		StopTime:  stop,
	}
}

func (w *WindowedSchedule) GetState() ScheduleState {
	return w.state
}

func (w *WindowedSchedule) Validate() error {
	if time.Now().After(w.StopTime) {
		return ErrInvalidStopTime
	}
	if w.StopTime.Before(w.StartTime) {
		return ErrStopBeforeStart
	}
	if w.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

func (w *WindowedSchedule) Wait(last time.Time) Response {
	// Wait till it is time to start if before the window start
	if time.Now().Before(w.StartTime) {
		wait := w.StartTime.Sub(time.Now())
		log.WithFields(log.Fields{
			"sleep-duration": wait,
		}).Debug("Waiting for window to start")
		time.Sleep(wait)
	}
	if last.String() == "0001-01-01 00:00:00 +0000 UTC" {
		log.Debug("Last was unset using start time")
		last = w.StartTime
	}
	// If within the window we wait our interval and return
	// otherwise we exit with a compleled state.
	var m uint
	if time.Now().Before(w.StopTime) {
		log.WithFields(log.Fields{
			"time-before-stop": w.StopTime.Sub(time.Now()),
		}).Debug("Within window, calling interval")
		m, _ = waitOnInterval(last, w.Interval)
	} else {
		w.state = Ended
		m = 0
	}
	return &WindowedScheduleResponse{
		state:    w.GetState(),
		missed:   m,
		lastTime: time.Now(),
	}
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type WindowedScheduleResponse struct {
	// err    error
	state    ScheduleState
	missed   uint
	lastTime time.Time
}

// Returns the state of the Schedule
func (w *WindowedScheduleResponse) State() ScheduleState {
	return w.state
}

// Returns last error
func (w *WindowedScheduleResponse) Error() error {
	return nil
}

// Returns any missed intervals
func (w *WindowedScheduleResponse) Missed() uint {
	return w.missed
}

func (w *WindowedScheduleResponse) LastTime() time.Time {
	return w.lastTime
}
