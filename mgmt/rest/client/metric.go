package client

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

var (
	ErrAPIResponseMetaType = errors.New("Received an invalid API response (META/TYPE)")
)

func (c *Client) GetMetricCatalog() *GetMetricCatalogResult {
	r := &GetMetricCatalogResult{}
	resp, err := c.do("GET", "/metrics", ContentTypeJSON)
	if err != nil {
		return &GetMetricCatalogResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricCatalogReturnedType:
		mc := resp.Body.(*rbody.MetricCatalogReturned)
		r.Catalog = convertCatalog(mc.Catalog)
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

func (c *Client) FetchMetrics(ns string, ver int) *FetchMetricsResult {
	r := &FetchMetricsResult{}
	q := fmt.Sprintf("/metrics%s?ver=%d", ns, ver)
	resp, err := c.do("GET", q, ContentTypeJSON)
	if err != nil {
		return &FetchMetricsResult{Err: err}
	}

	switch resp.Meta.Type {
	case rbody.MetricCatalogReturnedType:
		mc := resp.Body.(*rbody.MetricCatalogReturned)
		r.Catalog = convertCatalog(mc.Catalog)
	case rbody.ErrorType:
		r.Err = resp.Body.(*rbody.Error)
	default:
		r.Err = ErrAPIResponseMetaType
	}
	return r
}

type GetMetricCatalogResult struct {
	Catalog []MetricCatalogItem
	Err     error
}

func (g *GetMetricCatalogResult) Len() int {
	return len(g.Catalog)
}

type MetricCatalogItem struct {
	*rbody.CatalogItem
}

type FetchMetricsResult struct {
	Catalog []MetricCatalogItem
	Err     error
}

func (g *FetchMetricsResult) Len() int {
	return len(g.Catalog)
}

func convertCatalog(c []rbody.CatalogItem) []MetricCatalogItem {
	mci := make([]MetricCatalogItem, len(c))
	for i, _ := range c {
		mci[i] = MetricCatalogItem{&c[i]}
	}
	return mci
}
