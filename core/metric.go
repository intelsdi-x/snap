package core

import (
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"
)

type MetricType interface {
	Version() int
	Namespace() []string
	LastAdvertisedTime() time.Time
	Config() *cdata.ConfigDataNode
}

type Metric interface {
	Namespace() []string
	Data() interface{}
}
