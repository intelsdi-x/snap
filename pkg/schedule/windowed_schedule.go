package schedule

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	logger = log.WithField("_module", "schedule")
)

// A schedule that waits on an interval within a specific time window
type WindowedSchedule struct {
	Interval  time.Duration
	StartTime *time.Time
	StopTime  *time.Time
	state     ScheduleState
}

func NewWindowedSchedule(i time.Duration, start *time.Time, stop *time.Time) *WindowedSchedule {
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
	if w.StopTime != nil && time.Now().After(*w.StopTime) {
		return ErrInvalidStopTime
	}
	if w.StopTime != nil && w.StartTime != nil && w.StopTime.Before(*w.StartTime) {
		return ErrStopBeforeStart
	}
	if w.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

func (w *WindowedSchedule) Wait(last time.Time) Response {
	// Do we even have a specific start time?
	if w.StartTime != nil {
		// Wait till it is time to start if before the window start
		if time.Now().Before(*w.StartTime) {
			wait := w.StartTime.Sub(time.Now())
			logger.WithFields(log.Fields{
				"_block":         "windowed-wait",
				"sleep-duration": wait,
			}).Debug("Waiting for window to start")
			time.Sleep(wait)
		}
		if (last == time.Time{}) {
			logger.WithFields(log.Fields{
				"_block": "windowed-wait",
			}).Debug("Last was unset using start time")
			last = *w.StartTime
		}
	} else {
		if (last == time.Time{}) {
			logger.WithFields(log.Fields{
				"_block": "windowed-wait",
			}).Debug("Last was unset using start time")
			last = time.Now()
		}
	}

	// If within the window we wait our interval and return
	// otherwise we exit with a compleled state.
	var m uint
	// Do we even have a stop time?
	if w.StopTime != nil {
		if time.Now().Before(*w.StopTime) {
			logger.WithFields(log.Fields{
				"_block":           "windowed-wait",
				"time-before-stop": w.StopTime.Sub(time.Now()),
			}).Debug("Within window, calling interval")
			logger.WithFields(log.Fields{
				"_block":   "windowed-wait",
				"last":     last,
				"interval": w.Interval,
			}).Debug("waiting for interval")
			m, _ = waitOnInterval(last, w.Interval)
		} else {
			w.state = Ended
			m = 0
		}
	} else {
		logger.WithFields(log.Fields{
			"_block":   "windowed-wait",
			"last":     last,
			"interval": w.Interval,
		}).Debug("waiting for interval")
		// This has no end like a simple schedule
		m, _ = waitOnInterval(last, w.Interval)

	}
	return &WindowedScheduleResponse{
		state:    w.GetState(),
		missed:   m,
		lastTime: time.Now(),
	}
}

// A response from SimpleSchedule conforming to ScheduleResponse interface
type WindowedScheduleResponse struct {
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
