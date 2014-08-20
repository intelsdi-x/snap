package collection

import (
	"time"
)

type MetricType int

const (
	Polling MetricType = iota
	Subscribing
)

type Metric struct {
	Host       string
	Namespace  []string
	LastUpdate time.Time
	Values	map[string]metricType
	Collector string
	MetricType MetricType
}

// Selects metrics based on host, namespace, or value filters
type MetricFilter struct {
	HostFilter *HostFilter
	NamespaceFilter *NamespaceFilter
	ValueFilter *ValueFilter
}

type HostFilter struct {}

type NamespaceFilter struct {}

type ValueFilter struct {}

type metricType interface {}

type metricCache struct {
	LastPull time.Time
	Metrics []Metric
	New bool
}

func newMetric(hostname string, namespace []string, last time.Time, values map[string]metricType, collector string, mType MetricType) Metric{
	return Metric{hostname, namespace, last, values, collector, mType}
}
