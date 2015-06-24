package rest

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get current metric catalog
	mets, err := s.mm.MetricCatalog()
	// If error prepare and response with generic rest error
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	// Prepare metric catalog returned body
	b := rbody.NewMetricCatalogReturned()
	// Add catalog items to the body
	for _, m := range mets {
		b.Catalog = append(b.Catalog, rbody.MetricType{
			Namespace:               joinNamespace(m.Namespace()),
			Version:                 m.Version(),
			LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
		})
	}
	// return catalog and response with 200
	respond(200, b, w)
}

func (s *Server) getMetricsFromTree(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}
