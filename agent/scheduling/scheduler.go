package scheduling

import (
	"time"
)

// Manages metric workers
type Scheduler struct {

}

type schedule struct {
	Start *time.Time
	Stop *time.Time
	Interval time.Duration
	// Interval will be fixed to static division inside start-stop range. Optionally maybe have it based on from task start
}

func NewSchedule(duration time.Duration, times...time.Time) schedule{
	s := schedule{}
	s.Interval = duration
	if len(times) > 0 {
		s.Start = times[0]
		if len(times) > 1 {
			s.End = times[1]
		}
	}
	return s
}

