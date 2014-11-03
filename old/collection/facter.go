/*
Facter collector

author: Nicholas Weaver
arch: x86
*/

package collection

import (
	"encoding/json"
	"os/exec"
	"time"
)

var (
	facter_cache metricCache = newMetricCache()
)

type facterCollector struct {
	CollectorBase
	FacterPath string
}

type facterConfig struct {
	ConfigBase
	Path string
}

type fact interface{}

// TODO refacter config to be stored and properties to me mapped.
func NewFacterCollector(f facterConfig) Collector {
	c := new(facterCollector)
	c.collectorType = "facter"
	c.FacterPath = f.Path
	c.Caching = f.CachingEnabled()
	c.CachingTTL = f.CacheTTL()
	return c
}

func NewFacterConfig(path string) CollectorConfig {
	return facterConfig{Path: path}
}

func (f *facterCollector) GetMetricList() []Metric {
	if !f.Caching || (f.Caching && facter_cache.IsExpired(f.CachingTTL)) {
		out, _ := exec.Command("sh", "-c", f.FacterPath+" -j").Output()
		update_time := time.Now()
		facter_map := new(map[string]metricType)
		json.Unmarshal(out, facter_map)

		hostname := (*facter_map)["fqdn"].(string)

		metric := Metric{hostname, []string{"facter", "facts"}, update_time, map[string]metricType{}, "facter", Polling}
		for k, v := range *facter_map {
			metric.Values[k] = v
		}
		facter_cache.Metrics = []Metric{metric}
		facter_cache.LastPull = time.Now()
		facter_cache.New = false
	}

	return facter_cache.Metrics
}

func (f *facterCollector) GetMetricValues(metrics []Metric, things ...interface{}) []Metric {
	return f.GetMetricList()
}
