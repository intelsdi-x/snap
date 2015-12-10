/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbody

import "fmt"

// const A list of constants
const (
	MetricsReturnedType = "metrics_returned"
	MetricReturnedType  = "metric_returned"
)

// PolicyTable struct type
type PolicyTable struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default,omitempty"`
	Required bool        `json:"required"`
	Minimum  interface{} `json:"minimum,omitempty"`
	Maximum  interface{} `json:"maximum,omitempty"`
}

// Metric struct type
type Metric struct {
	LastAdvertisedTimestamp int64         `json:"last_advertised_timestamp,omitempty"`
	Namespace               string        `json:"namespace,omitempty"`
	Version                 int           `json:"version,omitempty"`
	Policy                  []PolicyTable `json:"policy,omitempty"`
	Href                    string        `json:"href"`
}

// MetricReturned struct type
type MetricReturned struct {
	Metric *Metric
}

// ResponseBodyMessage returns a string response message
func (m *MetricReturned) ResponseBodyMessage() string {
	return "Metric returned"
}

// ResponseBodyType returns a string response type message
func (m *MetricReturned) ResponseBodyType() string {
	return MetricReturnedType
}

// MetricsReturned Array of metrics
type MetricsReturned []Metric

// Len The length of metric list
func (m MetricsReturned) Len() int {
	return len(m)
}

// Less returns true or false by comparing two metric namespaces and versions.
func (m MetricsReturned) Less(i, j int) bool {
	return (fmt.Sprintf("%s:%d", m[i].Namespace, m[i].Version)) < (fmt.Sprintf("%s:%d", m[j].Namespace, m[j].Version))
}

// Swap two metrics
func (m MetricsReturned) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// NewMetricsReturned returns array of metrics
func NewMetricsReturned() MetricsReturned {
	return make([]Metric, 0)
}

// ResponseBodyMessage returns a string
func (m MetricsReturned) ResponseBodyMessage() string {
	return "Metric"
}

// ResponseBodyType returns a string response type
func (m MetricsReturned) ResponseBodyType() string {
	return MetricsReturnedType
}
