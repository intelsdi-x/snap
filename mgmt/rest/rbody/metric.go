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

import (
	"fmt"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	MetricsReturnedType = "metrics_returned"
	MetricReturnedType  = "metric_returned"
)

type PolicyTable cpolicy.RuleTable

type PolicyTableSlice []cpolicy.RuleTable

type Metric struct {
	LastAdvertisedTimestamp int64            `json:"last_advertised_timestamp,omitempty"`
	Namespace               string           `json:"namespace,omitempty"`
	Version                 int              `json:"version,omitempty"`
	Dynamic                 bool             `json:"dynamic"`
	DynamicElements         []DynamicElement `json:"dynamic_elements,omitempty"`
	Description             string           `json:"description,omitempty"`
	Unit                    string           `json:"unit,omitempty"`
	Policy                  PolicyTableSlice `json:"policy,omitempty"`
	Href                    string           `json:"href"`
}

type DynamicElement struct {
	Index       int    `json:"index,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
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
	return "Metrics returned"
}

func (m MetricsReturned) ResponseBodyType() string {
	return MetricsReturnedType
}
