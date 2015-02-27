package core

type MetricType interface {
	Namespace() []string
	LastAdvertisedTimestamp() int64
}

type Metric interface {
	Namespace() []string
	Data() interface{}
}
