package rest

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/intelsdilabs/pulse/core/cdata"
)

type metricType struct {
	Ns  string `json:"namespace"`
	Ver int    `json:"version"`
	LAT int64  `json:"last_advertised_timestamp,omitempty"`
}

func (m *metricType) Namespace() []string {
	return parseNamespace(m.Ns)
}

func (m *metricType) Version() int                  { return m.Ver }
func (m *metricType) LastAdvertisedTime() time.Time { return time.Unix(m.LAT, 0) }
func (m *metricType) Config() *cdata.ConfigDataNode { return nil }

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rmets := []metricType{}
	mets, err := s.mm.MetricCatalog()
	if err != nil {
		replyError(500, w, err)
		return
	}
	for _, m := range mets {
		rmets = append(rmets, metricType{
			Ns:  joinNamespace(m.Namespace()),
			Ver: m.Version(),
			LAT: m.LastAdvertisedTime().Unix(),
		})
	}
	mtsmap := make(map[string]interface{})
	mtsmap["metric_types"] = rmets
	replySuccess(200, w, mtsmap)
}

func (s *Server) getMetricsFromTree(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}
