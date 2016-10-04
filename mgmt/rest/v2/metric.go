package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/codegangsta/negroni"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody/v2"
)

func RespondWithMetrics(host string, mets []core.CatalogedMetric, w http.ResponseWriter) {
	b := v2.NewMetricsReturned()

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
		dyn, indexes := met.Namespace().IsDynamic()
		var dynamicElements []rbody.DynamicElement
		if dyn {
			dynamicElements = getDynamicElements(met.Namespace(), indexes)
		}

		b.Metrics = append(b.Metrics, rbody.Metric{
			Namespace:               met.Namespace().String(),
			Version:                 met.Version(),
			LastAdvertisedTimestamp: met.LastAdvertisedTime().Unix(),
			Description:             met.Description(),
			Dynamic:                 dyn,
			DynamicElements:         dynamicElements,
			Unit:                    met.Unit(),
			Policy:                  policies,
			Href:                    catalogedMetricURI(host, met),
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
