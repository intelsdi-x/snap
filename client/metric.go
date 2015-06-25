package pulse

import (
	"errors"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

var (
	ErrAPIResponseMetaType = errors.New("Received an invalid API response (META/TYPE)")
)

// MetricConfig is the type for incoming configuation data.
// This data must be in compliance with a MetricPolicy
// Ex:
//     metric type: /intel/facter/kernel
//     MetricConfig {
//         Key:   "path",
//         Value: "/usr/local/bin/facter",
//     }
type MetricConfig struct {
	Table map[string]interface{}
}

func (m *MetricConfig) Merge(mc *MetricConfig) {
}

/*
   Generally, MetricPolicy should be read only.
   This is data which is created by a plugin writer when
   describing their metrics.  It is provided in the pulse
   client in this manner so that a consumer of the metric
   catalog can read these values should they need to.
*/

// MetricPolicy is the requirements for collecting a metric.
// Ex:
//    metric type: /intel/facter/kernel
//    MetricPolicy {
//        Key:      "path",
//        Type:     "string,
//        Required: true,
//        Default:  "/bin/facter",
//    }
type MetricPolicy struct {
}

type MetricType struct {
	Namespace               string        `json:"namespace"`
	Version                 int           `json:"version,omitempty"`
	LastAdvertisedTimestamp int64         `json:"last_advertised_timestamp,omitempty"`
	Config                  *MetricConfig `json:"config,omitempty"`
	Policy                  *MetricPolicy `json:"policy,omitempty"`
}

func (c *Client) GetMetricCatalog() *GetMetricCatalogResult {
	resp, err := c.do("GET", "/metrics")
	if err != nil {
		return &GetMetricCatalogResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricCatalogReturnedType:
		return &GetMetricCatalogResult{resp.Body.(*rbody.MetricCatalogReturned), nil}
	case rbody.ErrorType:
		return &GetMetricCatalogResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &GetMetricCatalogResult{Err: ErrAPIResponseMetaType}
	}
}

type GetMetricCatalogResult struct {
	*rbody.MetricCatalogReturned
	Err error
}

// type getMetricTypesReply struct {
// 	respBody
// 	Data getMetricTypesData `json:"data"`
// }

// type getMetricTypesData struct {
// 	MetricTypes []*MetricType `json:"metric_types"`
// }
