package collection

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Collector struct {
	Caching    bool
	CachingTTL float64
}

type CollectorConfig interface {
	CachingEnabled() bool
	CacheTTL() float64
}

type collector interface {
	GetMetricList() []Metric
	GetMetricValues(metrics []Metric, things ...interface{}) []Metric
}

func GetHostname() string {
	x, _ := exec.Command("sh", "-c", "hostname -f").Output()
	return string(x[:len(x)-1])
}

func newMetricCache() metricCache {
	return metricCache{time.Now(), []Metric{}, true}
}

func NewCollectorByType(cType string, cConfig CollectorConfig) collector {
	switch cType {
	case "collectd":
		return NewCollectDCollector(cConfig.(*collectDConfig))
	default:
		panic(1)
	}
}

func NewCollectorMap() map[string]collector {
	return map[string]collector{}
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
