package rbody

import (
	"fmt"
)

const (
	MetricCatalogReturnedType = "metric_catalog_returned"
)

type MetricCatalogReturned struct {
	Catalog []MetricType
}

type MetricType struct {
	Namespace               string `json:"namespace"`
	Version                 int    `json:"version"`
	LastAdvertisedTimestamp int64  `json:"last_advertised_timestamp,omitempty"`
}

func (m *MetricType) key() string {
	return fmt.Sprintf("%s/%d", m.Namespace, m.Version)
}

func (m *MetricCatalogReturned) Len() int {
	return len(m.Catalog)
}

func (m *MetricCatalogReturned) Less(i, j int) bool {
	return m.Catalog[i].key() < m.Catalog[j].key()
}

func (m *MetricCatalogReturned) Swap(i, j int) {
	m.Catalog[i], m.Catalog[j] = m.Catalog[j], m.Catalog[i]
}

func NewMetricCatalogReturned() *MetricCatalogReturned {
	return &MetricCatalogReturned{Catalog: make([]MetricType, 0)}
}

func (m *MetricCatalogReturned) ResponseBodyMessage() string {
	return "Metric catalog returned"
}

func (m *MetricCatalogReturned) ResponseBodyType() string {
	return MetricCatalogReturnedType
}
