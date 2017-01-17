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

package client

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
)

var (
	// The default response error.
	ErrAPIResponseMetaType = errors.New("Received an invalid API response (META/TYPE)")
)

// GetMetricCatalog retrieves the metric catalog from a snap/client by issuing an HTTP GET request.
// A slice of metric catalogs returns if succeeded. Otherwise an error is returned.
func (c *Client) GetMetricCatalog() *GetMetricsResult {
	r := &GetMetricsResult{}
	resp, err := c.do("GET", "/metrics", ContentTypeJSON)
	if err != nil {
		return &GetMetricsResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricsReturnedType:
		mc := resp.Body.(*rbody.MetricsReturned)
		r.Catalog = convertCatalog(mc)
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// FetchMetrics retrieves the metric catalog given metric namespace and version through an HTTP GET request.
// It returns the corresponding metric catalog if succeeded. Otherwise, an error is returned.
func (c *Client) FetchMetrics(ns string, ver int) *GetMetricsResult {
	r := &GetMetricsResult{}
	q := fmt.Sprintf("/metrics?ns=%s&ver=%d", ns, ver)
	resp, err := c.do("GET", q, ContentTypeJSON)
	if err != nil {
		return &GetMetricsResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricsReturnedType:
		mc := resp.Body.(*rbody.MetricsReturned)
		r.Catalog = convertCatalog(mc)
	case rbody.MetricReturnedType:
		mc := resp.Body.(*rbody.MetricReturned)
		r.Catalog = []*rbody.Metric{mc.Metric}
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetMetricVersions retrieves all versions of a metric at a given namespace.
func (c *Client) GetMetricVersions(ns string) *GetMetricsResult {
	r := &GetMetricsResult{}
	q := fmt.Sprintf("/metrics?ns=%s", ns)
	resp, err := c.do("GET", q, ContentTypeJSON)
	if err != nil {
		return &GetMetricsResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricsReturnedType:
		mc := resp.Body.(*rbody.MetricsReturned)
		r.Catalog = convertCatalog(mc)
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetMetric retrieves a metric at a given namespace and version.
// If the version is < 1, the latest version is returned.
// Returns an interface as several return types are possible
// (array of metrics).
func (c *Client) GetMetric(ns string, ver int) interface{} {
	r := &GetMetricResult{}

	q := fmt.Sprintf("/metrics?ns=%s&ver=%d", ns, ver)
	resp, err := c.do("GET", q, ContentTypeJSON)
	if err != nil {
		return &GetMetricResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricReturnedType:
		mc := resp.Body.(*rbody.MetricReturned)
		r.Metric = mc.Metric
	case rbody.MetricsReturnedType:
		mc := resp.Body.(*rbody.MetricsReturned)
		r := convertMetrics(mc)
		return r
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

// GetMetricsResult is the response from snap/client on a GetMetricCatalog call.
type GetMetricsResult struct {
	Catalog []*rbody.Metric
	Err     error
}

// GetMetricResult is the response from snap/client on a GetMetricCatalog call.
type GetMetricResult struct {
	Metric *rbody.Metric
	Err    error
}

// Len returns the slice's length of GetMetricsResult.Catalog.
func (g *GetMetricsResult) Len() int {
	return len(g.Catalog)
}

func convertCatalog(c *rbody.MetricsReturned) []*rbody.Metric {
	mci := make([]*rbody.Metric, len(*c))
	for i := range *c {
		mci[i] = &(*c)[i]
	}
	return mci
}

func convertMetrics(c *rbody.MetricsReturned) []*GetMetricResult {
	mci := make([]*GetMetricResult, len(*c))
	for i, _ := range *c {
		r := &GetMetricResult{Metric: &(*c)[i]}
		mci[i] = r
	}
	return mci
}
