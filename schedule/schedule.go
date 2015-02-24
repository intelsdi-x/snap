package schedule

import (
	"time"
)

type Schedule interface {
	Wait(time.Time) chan struct{}
}
