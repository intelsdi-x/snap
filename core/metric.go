package core

import "github.com/intelsdilabs/pulse/core/cdata"

type MetricType interface {
	Version() int
	Namespace() []string
	LastAdvertisedTimestamp() int64
	Config() *cdata.ConfigDataNode
}

type Metric interface {
	Namespace() []string
	Data() interface{}
}
