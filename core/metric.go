package core

type MetricType interface {
	Namespace() []string
	LastAdvertisedTimestamp() int64
}
