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

package v2

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/response"
	"github.com/intelsdi-x/snap/pkg/stringutils"
	"github.com/julienschmidt/httprouter"
)

func (s *V2) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ver := 0 // 0: get all metrics

	// If we are provided a parameter with the name 'ns' we need to
	// perform a query
	q := r.URL.Query()
	v := q.Get("ver")
	ns_query := q.Get("ns")
	if ns_query != "" {
		ver = 0 // 0: get all versions
		if v != "" {
			var err error
			ver, err = strconv.Atoi(v)
			if err != nil {
				response.Write(400, response.FromError(err), w)
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
			response.Write(404, response.FromError(err), w)
			return
		}
		respondWithMetrics(r.Host, mts, w)
		return
	}

	mts, err := s.metricManager.MetricCatalog()
	if err != nil {
		response.Write(500, response.FromError(err), w)
		return
	}
	respondWithMetrics(r.Host, mts, w)
}

func (s *V2) getMetricsFromTree(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	namespace := params.ByName("namespace")

	// we land here if the request contains a trailing slash, because it matches the tree
	// lookup URL: /v2/metrics/*namespace.  If the length of the namespace param is 1, we
	// redirect the request to getMetrics.  This results in GET /v2/metrics and
	// GET /v2/metrics/ behaving the same way.
	if len(namespace) <= 1 {
		s.getMetrics(w, r, params)
		return
	}

	ns := parseNamespace(namespace)

	q := r.URL.Query()
	v := q.Get("ver")

	var err error
	ver := 0 // returns all metrics
	if v != "" {
		ver, err = strconv.Atoi(v)
		if err != nil {
			response.Write(400, response.FromError(err), w)
			return
		}
	}

	if ns[len(ns)-1] == "*" || ver == 0 {
		mts, err := s.metricManager.FetchMetrics(core.NewNamespace(ns...), ver)
		if err != nil {
			response.Write(404, response.FromError(err), w)
			return
		}
		respondWithMetrics(r.Host, mts, w)
		return
	} else {
		mt, err := s.metricManager.GetMetric(core.NewNamespace(ns...), ver)
		if err != nil {
			response.Write(404, response.FromError(err), w)
			return
		}

		dyn, indexes := mt.Namespace().IsDynamic()
		mb := &response.Metric{
			Namespace:       mt.Namespace().String(),
			Version:         mt.Version(),
			Dynamic:         dyn,
			DynamicElements: getDynamicElements(mt.Namespace(), indexes),
			Description:     mt.Description(),
			Unit:            mt.Unit(),
			LastAdvertisedTimestamp: mt.LastAdvertisedTime().Unix(),
			Policy:                  response.PolicyTableSlice(mt.Policy().RulesAsTable()),
			Href:                    catalogedMetricURI(r.Host, version, mt),
		}
		response.Write(200, mb, w)
	}
}

func respondWithMetrics(host string, mts []core.CatalogedMetric, w http.ResponseWriter) {
	b := make(response.Metrics, 0)
	for _, m := range mts {
		policies := response.PolicyTableSlice(m.Policy().RulesAsTable())
		dyn, indexes := m.Namespace().IsDynamic()
		b = append(b, response.Metric{
			Namespace:               m.Namespace().String(),
			Version:                 m.Version(),
			LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
			Description:             m.Description(),
			Dynamic:                 dyn,
			DynamicElements:         getDynamicElements(m.Namespace(), indexes),
			Unit:                    m.Unit(),
			Policy:                  policies,
			Href:                    catalogedMetricURI(host, version, m),
		})
	}
	sort.Sort(b)
	response.Write(200, b, w)
}

func catalogedMetricURI(host, version string, mt core.CatalogedMetric) string {
	return fmt.Sprintf("%s://%s/%s/metrics?ns=%s&ver=%d", protocolPrefix, host, version, url.QueryEscape(mt.Namespace().String()), mt.Version())
}

func getDynamicElements(ns core.Namespace, indexes []int) []response.DynamicElement {
	elements := make([]response.DynamicElement, 0, len(indexes))
	for _, v := range indexes {
		e := ns.Element(v)
		elements = append(elements, response.DynamicElement{
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
