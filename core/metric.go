package core

type MetricType interface {
	Version() int
	Namespace() []string
	LastAdvertisedTimestamp() int64
}

type Metric interface {
	Namespace() []string
	Data() interface{}
}
