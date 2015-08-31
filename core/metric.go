package core

import (
	"time"

	"github.com/intelsdi-x/pulse/core/cdata"
)

// Metric represents a Pulse metric collected or to be collected
type Metric interface {
	RequestedMetric
	LastAdvertisedTime() time.Time
	Config() *cdata.ConfigDataNode
	Data() interface{}
	Source() string
	Timestamp() time.Time
}

// RequestedMetric is a metric requested for collection
type RequestedMetric interface {
	Namespace() []string
	Version() int
}

type CatalogedMetric interface {
	Namespace() string
	Versions() map[int]Metric
}
