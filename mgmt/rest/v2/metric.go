package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody/v2"
	"github.com/urfave/negroni"
)

func RespondWithMetrics(host string, mts []core.CatalogedMetric, w http.ResponseWriter) {
	b := v2.NewMetricsReturned()
	for _, m := range mts {
		var policies rbody.PolicyTableSlice
		p := m.Policy().RulesAsTable()
		policies = rbody.PolicyTableSlice(p)
		dyn, indexes := m.Namespace().IsDynamic()
		var dynamicElements []rbody.DynamicElement
		if dyn {
			dynamicElements = getDynamicElements(m.Namespace(), indexes)
		}
		b.Metrics = append(b.Metrics, rbody.Metric{
			Namespace:               m.Namespace().String(),
			Version:                 m.Version(),
			LastAdvertisedTimestamp: m.LastAdvertisedTime().Unix(),
			Description:             m.Description(),
			Dynamic:                 dyn,
			DynamicElements:         dynamicElements,
			Unit:                    m.Unit(),
			Policy:                  policies,
			Href:                    catalogedMetricURI(host, m),
		})
	}
	sort.Sort(b)
	respond(200, b, w)
}

func catalogedMetricURI(host string, mt core.CatalogedMetric) string {
	return fmt.Sprintf("%s://%s/v1/metrics?ns=%s&ver=%d", "http", host, url.QueryEscape(mt.Namespace().String()), mt.Version())
}

func getDynamicElements(ns core.Namespace, indexes []int) []rbody.DynamicElement {
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

func respond(code int, b rbody.Body, w http.ResponseWriter) {
	resp := &rbody.APIResponse{
		Meta: &rbody.APIResponseMeta{
			Code:    code,
			Message: b.ResponseBodyMessage(),
			Type:    b.ResponseBodyType(),
			Version: 2,
		},
		Body: b,
	}
	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

	j, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic(err)
	}
	j = bytes.Replace(j, []byte("\\u0026"), []byte("&"), -1)
	fmt.Fprint(w, string(j))
}
