package pulse

import (
	"encoding/json"
	"errors"
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

func (c *Client) GetMetricCatalog() ([]*MetricType, error) {
	resp, err := c.do("GET", "/metrics", ContentTypeJSON)
	var rb getMetricTypesReply
	err = json.Unmarshal(resp.body, &rb)
	if err != nil {
		return nil, err
	}
	if rb.Meta.Code != 200 {
		return nil, errors.New(rb.Meta.Message)
	}
	return rb.Data.MetricTypes, err
}

type getMetricTypesReply struct {
	respBody
	Data getMetricTypesData `json:"data"`
}

type getMetricTypesData struct {
	MetricTypes []*MetricType `json:"metric_types"`
}
