package collection

import (
	"time"
	"strconv"
	"strings"
	"os/exec"
)

type Collector struct {
	Caching bool
	CachingTTL float64
}

type CollectorConfig interface {
}

type collector interface {
	GetMetricList() []Metric
	GetMetricValues(metrics []Metric, things...interface{}) []Metric
}

func GetHostname() string{
	x, _ := exec.Command("sh", "-c", "hostname -f").Output()
	return string(x[:len(x)-1])
}

func newMetricCache() metricCache{
	return metricCache{time.Now(), []Metric{}, true}
}

func (m *metricCache) IsExpired(min_age float64) bool{
	return time.Since(m.LastPull).Seconds() > min_age || m.New
}

func (m *Metric) GetNamespaceString() string{
	return strings.Join(m.Namespace, "/")
}

func (m *Metric) GetFullNamespace() string{
	return m.Host + "/" + m.GetNamespaceString()
}

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}
