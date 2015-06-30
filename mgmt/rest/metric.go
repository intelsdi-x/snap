package rest

import (
	"fmt"
	"net/http"
	"sort"

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

	mets, err := s.mm.FetchMetrics(ns)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	respondWithMetrics(mets, w)
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
