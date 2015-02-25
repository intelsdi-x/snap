package core

type MetricType interface {
	Version() int
	Namespace() []string
	LastAdvertisedTimestamp() int64
}
