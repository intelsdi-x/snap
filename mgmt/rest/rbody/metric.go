package rbody

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

func NewMetricCatalogReturned() *MetricCatalogReturned {
	return &MetricCatalogReturned{Catalog: make([]MetricType, 0)}
}

func (m *MetricCatalogReturned) ResponseBodyMessage() string {
	return "Metric catalog returned"
}

func (m *MetricCatalogReturned) ResponseBodyType() string {
	return MetricCatalogReturnedType
}
