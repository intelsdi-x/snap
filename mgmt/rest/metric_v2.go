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

package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/rbody"
	"github.com/intelsdi-x/snap/pkg/stringutils"
	"github.com/julienschmidt/httprouter"
)

func (s *Server) getMetricsV2(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
				respondV2(400, rbody.FromError(err), w)
				return
			}
		}
		// strip the leading char and split on the remaining.
		fc := stringutils.GetFirstChar(ns_query)
		ns := strings.Split(strings.TrimLeft(ns_query, fc), fc)
		if ns[len(ns)-1] == "*" {
			ns = ns[:len(ns)-1]
		}

		mts, err := s.mm.FetchMetrics(core.NewNamespace(ns...), ver)
		if err != nil {
			respondV2(404, rbody.FromError(err), w)
			return
		}
		respondWithMetricsV2(r.Host, mts, w)
		return
	}

	mts, err := s.mm.MetricCatalog()
	if err != nil {
		respondV2(500, rbody.FromError(err), w)
		return
	}
	respondWithMetricsV2(r.Host, mts, w)
}

func respondWithMetricsV2(host string, mts []core.CatalogedMetric, w http.ResponseWriter) {
	b := rbody.NewMetricsReturned()
	for _, m := range mts {
		policies := rbody.PolicyTableSlice(m.Policy().RulesAsTable())
		dyn, indexes := m.Namespace().IsDynamic()
		var dynamicElements []rbody.DynamicElement
		if dyn {
			dynamicElements = getDynamicElementsV2(m.Namespace(), indexes)
		}
		b = append(b, rbody.Metric{
			Namespace:               m.Namespace().String(),
			Version:                 m.Version(),
			LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
			Description:             m.Description(),
			Dynamic:                 dyn,
			DynamicElements:         dynamicElements,
			Unit:                    m.Unit(),
			Policy:                  policies,
			Href:                    catalogedMetricURI(host, "v2", m),
		})
	}
	sort.Sort(b)
	respondV2(200, b, w)
}

func catalogedMetricURI(host, version string, mt core.CatalogedMetric) string {
	return fmt.Sprintf("%s://%s/%s/metrics?ns=%s&ver=%d", protocolPrefix, host, version, url.QueryEscape(mt.Namespace().String()), mt.Version())
}

func getDynamicElementsV2(ns core.Namespace, indexes []int) []rbody.DynamicElement {
	elements := make([]rbody.DynamicElement, 0, len(indexes))
	for _, v := range indexes {
		e := ns.Element(v)
		elements = append(elements, rbody.DynamicElement{
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
