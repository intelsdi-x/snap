package core

import (
	"time"
)

type MetricType interface {
	Version() int
	Namespace() []string
	LastAdvertisedTime() time.Time
}

type Metric interface {
	Namespace() []string
	Data() interface{}
}
