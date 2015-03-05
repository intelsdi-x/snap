package plugin

import (
	"time"
)

// The metric used by plugin will probably have to be typed by metric data to
// allow passing over RPC. This can on the agent side meet core.Metric interface
// by functions supporting conversion.
type Metric struct {
	Namespace []string
	Data      interface{}
}

type MetricType struct {
	Namespace_          []string
	LastAdvertisedTime_ time.Time
}

func (m *MetricType) Namespace() []string {
	return m.Namespace_
}

func (m *MetricType) LastAdvertisedTime() time.Time {
	return m.LastAdvertisedTime_
}

func NewMetricType(ns []string) *MetricType {
	return &MetricType{
		Namespace_:          ns,
		LastAdvertisedTime_: time.Now(),
	}
}
