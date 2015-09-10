package rbody

import "fmt"

const (
	MetricsReturnedType = "metrics_returned"
	MetricReturnedType  = "metric_returned"
)

type PolicyTable struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default"`
	Required bool        `json:"required"`
}

type Metric struct {
	LastAdvertisedTimestamp int64         `json:"last_advertised_timestamp,omitempty"`
	Namespace               string        `json:"namespace,omitempty"`
	Version                 int           `json:"version,omitempty"`
	Policy                  []PolicyTable `json:"policy,omitempty"`
}

type MetricReturned struct {
	Metric *Metric
}

func (m *MetricReturned) ResponseBodyMessage() string {
	return "Metric returned"
}

func (m *MetricReturned) ResponseBodyType() string {
	return MetricReturnedType
}

type MetricsReturned []Metric

func (m MetricsReturned) Len() int {
	return len(m)
}

func (m MetricsReturned) Less(i, j int) bool {
	return (fmt.Sprintf("%s:%d", m[i].Namespace, m[i].Version)) < (fmt.Sprintf("%s:%d", m[j].Namespace, m[j].Version))
}

func (m MetricsReturned) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func NewMetricsReturned() MetricsReturned {
	return make([]Metric, 0)
}

func (m MetricsReturned) ResponseBodyMessage() string {
	return "Metric"
}

func (m MetricsReturned) ResponseBodyType() string {
	return MetricsReturnedType
}
