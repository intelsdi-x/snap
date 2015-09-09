package core

import (
	"strings"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/cdata"
)

// Metric represents a Pulse metric collected or to be collected
type Metric interface {
	RequestedMetric
	Config() *cdata.ConfigDataNode
	LastAdvertisedTime() time.Time
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
	RequestedMetric
	LastAdvertisedTime() time.Time
	Policy() *cpolicy.ConfigPolicyNode
}

func JoinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
