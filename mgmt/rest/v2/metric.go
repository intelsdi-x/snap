/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"net/url"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/stringutils"
	"github.com/julienschmidt/httprouter"
)

// MetricsResponse is the representation of metric operation response.
//
// swagger:response MetricsResponse
type MetricsResp struct {
	// in: body
	Body struct {
		Metrics []Metric `json:"metrics,omitempty"`
	}
}

// MetricParams defines the input query parameters for get a specific metric.
//
// swagger:parameters getMetrics
type MetricParams struct {
	// in: query
	Ns string `json:"ns"`
	// in: query
	Ver int `json:"ver"`
}

type MetricsResonse struct {
	Metrics Metrics `json:"metrics,omitempty"`
}

type Metrics []Metric

// Metric represents the metric type.
type Metric struct {
	LastAdvertisedTimestamp int64 `json:"last_advertised_timestamp,omitempty"`
	//required: true
	Namespace string `json:"namespace"`
	Version   int    `json:"version,omitempty"`
	// Dynamic boolean representation if the metric has dynamic element.
	Dynamic bool `json:"dynamic"`
	// DynamicElements a slice of dynamic elements.
	DynamicElements []DynamicElement `json:"dynamic_elements,omitempty"`
	Description     string           `json:"description,omitempty"`
	Unit            string           `json:"unit,omitempty"`
	// Policy a slice of metric rules.
	Policy PolicyTableSlice `json:"policy,omitempty"`
	Href   string           `json:"href"`
}

type DynamicElement struct {
	Index int `json:"index,omitempty"`
	// required: true
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Used to sort the metrics before marshalling the response
func (m Metrics) Len() int {
	return len(m)
}

func (m Metrics) Less(i, j int) bool {
	return (fmt.Sprintf("%s:%d", m[i].Namespace, m[i].Version)) < (fmt.Sprintf("%s:%d", m[j].Namespace, m[j].Version))
}

func (m Metrics) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (s *apiV2) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// If we are provided a parameter with the name 'ns' we need to
	// perform a query
	q := r.URL.Query()
	v := q.Get("ver")
	ns_query := q.Get("ns")
	if ns_query != "" {
		ver := 0 // 0: get all versions
		if v != "" {
			var err error
			ver, err = strconv.Atoi(v)
			if err != nil {
				Write(400, FromError(err), w)
				return
			}
		}
		// strip the leading char and split on the remaining.
		fc := stringutils.GetFirstChar(ns_query)
		ns := strings.Split(strings.TrimLeft(ns_query, fc), fc)
		if ns[len(ns)-1] == "*" {
			ns = ns[:len(ns)-1]
		}

		mts, err := s.metricManager.FetchMetrics(core.NewNamespace(ns...), ver)
		if err != nil {
			Write(404, FromError(err), w)
			return
		}
		respondWithMetrics(r.Host, mts, w)
		return
	}

	mts, err := s.metricManager.MetricCatalog()
	if err != nil {
		Write(500, FromError(err), w)
		return
	}
	respondWithMetrics(r.Host, mts, w)
}

func respondWithMetrics(host string, mts []core.CatalogedMetric, w http.ResponseWriter) {
	b := MetricsResonse{Metrics: make(Metrics, 0)}
	for _, m := range mts {
		policies := PolicyTableSlice(m.Policy().RulesAsTable())
		dyn, indexes := m.Namespace().IsDynamic()
		b.Metrics = append(b.Metrics, Metric{
			Namespace:               m.Namespace().String(),
			Version:                 m.Version(),
			LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
			Description:             m.Description(),
			Dynamic:                 dyn,
			DynamicElements:         getDynamicElements(m.Namespace(), indexes),
			Unit:                    m.Unit(),
			Policy:                  policies,
			Href:                    catalogedMetricURI(host, m),
		})
	}
	sort.Sort(b.Metrics)
	Write(200, b, w)
}

func catalogedMetricURI(host string, mt core.CatalogedMetric) string {
	return fmt.Sprintf("%s://%s/%s/metrics?ns=%s&ver=%d", protocolPrefix, host, version, url.QueryEscape(mt.Namespace().String()), mt.Version())
}

func getDynamicElements(ns core.Namespace, indexes []int) []DynamicElement {
	elements := make([]DynamicElement, 0, len(indexes))
	for _, v := range indexes {
		e := ns.Element(v)
		elements = append(elements, DynamicElement{
			Index:       v,
			Name:        e.Name,
			Description: e.Description,
		})
	}
	return elements
}

func parseNamespace(ns string) []string {
	fc := stringutils.GetFirstChar(ns)
	ns = strings.Trim(ns, fc)
	return strings.Split(ns, fc)
}
