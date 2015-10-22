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
	"net/http"
	"sort"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mets, err := s.mm.MetricCatalog()
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	respondWithMetrics(mets, w)
}

func (s *Server) getMetricsFromTree(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ns := parseNamespace(params.ByName("namespace"))

	var (
		ver int
		err error
	)
	q := r.URL.Query()
	v := q.Get("ver")
	if v == "" {
		ver = -1
	} else {
		ver, err = strconv.Atoi(v)
		if err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	}

	if ns[len(ns)-1] == "*" {
		mets, err := s.mm.FetchMetrics(ns[:len(ns)-1], ver)
		if err != nil {
			respond(404, rbody.FromError(err), w)
			return
		}
		respondWithMetrics(mets, w)
		return
	}

	mt, err := s.mm.GetMetric(ns, ver)
	if err != nil {
		respond(404, rbody.FromError(err), w)
		return
	}

	b := &rbody.MetricReturned{}
	mb := &rbody.Metric{
		Namespace:               core.JoinNamespace(mt.Namespace()),
		Version:                 mt.Version(),
		LastAdvertisedTimestamp: mt.LastAdvertisedTime().Unix(),
	}
	rt := mt.Policy().RulesAsTable()
	policies := make([]rbody.PolicyTable, 0, len(rt))
	for _, r := range rt {
		policies = append(policies, rbody.PolicyTable{
			Name:     r.Name,
			Type:     r.Type,
			Default:  r.Default,
			Required: r.Required,
			Minimum:  r.Minimum,
			Maximum:  r.Maximum,
		})
	}
	mb.Policy = policies
	b.Metric = mb
	respond(200, b, w)
}

func respondWithMetrics(mets []core.CatalogedMetric, w http.ResponseWriter) {
	b := rbody.NewMetricsReturned()

	for _, met := range mets {
		rt := met.Policy().RulesAsTable()
		policies := make([]rbody.PolicyTable, 0, len(rt))
		for _, r := range rt {
			policies = append(policies, rbody.PolicyTable{
				Name:     r.Name,
				Type:     r.Type,
				Default:  r.Default,
				Required: r.Required,
				Minimum:  r.Minimum,
				Maximum:  r.Maximum,
			})
		}
		b = append(b, rbody.Metric{
			Namespace:               core.JoinNamespace(met.Namespace()),
			Version:                 met.Version(),
			LastAdvertisedTimestamp: met.LastAdvertisedTime().Unix(),
			Policy:                  policies,
		})
	}
	sort.Sort(b)
	respond(200, b, w)
}
