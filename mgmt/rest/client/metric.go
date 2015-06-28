package client

import (
	"errors"

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

func convertCatalog(c []rbody.CatalogItem) []MetricCatalogItem {
	mci := make([]MetricCatalogItem, len(c))
	for i, _ := range c {
		mci[i] = MetricCatalogItem{&c[i]}
	}
	return mci
}
