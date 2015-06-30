package rbody

const (
	MetricCatalogReturnedType = "metric_catalog_returned"
	MetricReturnedType        = "metric_returned"
)

type CatalogItem struct {
	Namespace string             `json:"namespace"`
	Versions  map[string]*Metric `json:"versions"`
}

func (m *CatalogItem) key() string {
	return m.Namespace
}

type Metric struct {
	LastAdvertisedTimestamp int64  `json:"last_advertised_timestamp,omitempty"`
	Namespace               string `json:"namespace,omitempty"`
	Version                 int    `json:"version,omitempty"`
}

type MetricReturned struct {
	Metric *Metric
}

func (m *MetricReturned) ResponseBodyMessage() string {
	return "Metric returned"
}

func (m *MetricReturned) ResponseBodyType() string {
	return MetricReturnedType
}

type MetricCatalogReturned struct {
	Catalog []CatalogItem
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
	return &MetricCatalogReturned{Catalog: make([]CatalogItem, 0)}
}

func (m *MetricCatalogReturned) ResponseBodyMessage() string {
	return "Metric catalog returned"
}

func (m *MetricCatalogReturned) ResponseBodyType() string {
	return MetricCatalogReturnedType
}
