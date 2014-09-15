package collection

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type CollectorBase struct {
	Caching    bool
	CachingTTL float64

	collectorType string
}

type ConfigBase struct {
	isCaching  bool
	cachingTTL float64
}

type CollectorConfig interface {
	CachingEnabled() bool
	CacheTTL() float64
}

type Collector interface {
	CollectorType() string
	GetMetricList() []Metric
	GetMetricValues(metrics []Metric, things ...interface{}) []Metric
}

func (c *CollectorBase) CollectorType() string {
	return c.collectorType
}

func (c ConfigBase) CachingEnabled() bool {
	return c.isCaching
}

func (c ConfigBase) CacheTTL() float64 {
	return c.cachingTTL
}

func GetHostname() string {
	x, _ := exec.Command("sh", "-c", "hostname -f").Output()
	return string(x[:len(x)-1])
}

func newMetricCache() metricCache {
	return metricCache{time.Now(), []Metric{}, true}
}

func NewCollectorByType(cType string, cConfig CollectorConfig) Collector {
	switch cType {
	case "collectd":
		return NewCollectDCollector(cConfig.(collectDConfig))
	case "facter":
		return NewFacterCollector(cConfig.(facterConfig))
	case "container":
		return NewContainerCollector(cConfig.(containerConfig))
	default:
		panic(1)
	}
}

func NewCollectorMap() map[string]Collector {
	return map[string]Collector{}
}

func (m *metricCache) IsExpired(min_age float64) bool {
	return time.Since(m.LastPull).Seconds() > min_age || m.New
}

func (m *Metric) GetNamespaceString() string {
	return strings.Join(m.Namespace, "/")
}

func (m *Metric) GetFullNamespace() string {
	return m.Host + "/" + m.GetNamespaceString()
}

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}
