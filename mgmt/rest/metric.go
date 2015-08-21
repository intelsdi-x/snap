package rest

import (
	"fmt"
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
	b.Metric = mb
	respond(200, b, w)
}

func respondWithMetrics(mets []core.CatalogedMetric, w http.ResponseWriter) {
	b := rbody.NewMetricCatalogReturned()

	for _, cm := range mets {
		ci := &rbody.CatalogItem{
			Namespace: cm.Namespace(),
			Versions:  make(map[string]*rbody.Metric, len(cm.Versions())),
		}
		for k, m := range cm.Versions() {
			ci.Versions[fmt.Sprintf("%d", k)] = &rbody.Metric{
				LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
			}
		}

		b.Catalog = append(b.Catalog, *ci)
	}
	sort.Sort(b)
	respond(200, b, w)
}
