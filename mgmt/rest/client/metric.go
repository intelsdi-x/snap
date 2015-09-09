package client

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

var (
	ErrAPIResponseMetaType = errors.New("Received an invalid API response (META/TYPE)")
)

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

func (c *Client) FetchMetrics(ns string, ver int) *GetMetricsResult {
	r := &GetMetricsResult{}
	q := fmt.Sprintf("/metrics%s?ver=%d", ns, ver)
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

type GetMetricsResult struct {
	Catalog []*rbody.Metric
	Err     error
}

type GetMetricResult struct {
	Metric *rbody.Metric
	Err    error
}

func (g *GetMetricsResult) Len() int {
	return len(g.Catalog)
}

func convertCatalog(c *rbody.MetricsReturned) []*rbody.Metric {
	mci := make([]*rbody.Metric, len(*c))
	for i, _ := range *c {
		mci[i] = &(*c)[i]
	}
	return mci
}
