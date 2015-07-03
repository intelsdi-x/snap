package scheduler_event

import (
	"github.com/intelsdi-x/pulse/core"
)

const (
	MetricCollected        = "Control.MetricsCollected"
	MetricCollectionFailed = "Control.MetricCollectionFailed"
)

type MetricCollectedEvent struct {
	TaskID  uint64
	Metrics []core.Metric
}

func (e MetricCollectedEvent) Namespace() string {
	return MetricCollected
}

type MetricCollectionFailedEvent struct {
	TaskID uint64
	Errors []error
}

func (e MetricCollectionFailedEvent) Namespace() string {
	return MetricCollectionFailed
}
