package v2

import (
	"fmt"

	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
)

type MetricList struct {
	Metrics []rbody.Metric `json:"metrics,omitempty"`
}

type MetricReturned struct {
	Metric *rbody.Metric `json:"metric,omitempty"`
}

func (m *MetricReturned) ResponseBodyMessage() string {
	return "Metric returned"
}

func (m *MetricReturned) ResponseBodyType() string {
	return "metric_returned"
}

type MetricsReturned MetricList

func (m MetricsReturned) Len() int {
	return len(m.Metrics)
}

func (m MetricsReturned) Less(i, j int) bool {
	return (fmt.Sprintf("%s:%d", m.Metrics[i].Namespace, m.Metrics[i].Version)) < (fmt.Sprintf("%s:%d", m.Metrics[j].Namespace, m.Metrics[j].Version))
}

func (m MetricsReturned) Swap(i, j int) {
	m.Metrics[i], m.Metrics[j] = m.Metrics[j], m.Metrics[i]
}

func NewMetricsReturned() MetricsReturned {
	return MetricsReturned{Metrics: []rbody.Metric{}}
}

func (m MetricsReturned) ResponseBodyMessage() string {
	return "Metrics returned"
}

func (m MetricsReturned) ResponseBodyType() string {
	return rbody.MetricsReturnedType
}
