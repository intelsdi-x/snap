/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

type PluginCacheClient interface {
	CacheHits(string, int) (uint64, error)
	CacheMisses(string, int) (uint64, error)
	AllCacheHits() uint64
	AllCacheMisses() uint64
}

// PluginClient A client providing common plugin method calls.
type PluginClient interface {
	PluginCacheClient
	SetKey() error
	Ping() error
	Kill(string) error
	GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
}

// PluginCollectorClient A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]core.Metric) ([]core.Metric, error)
	GetMetricTypes(plugin.PluginConfigType) ([]core.Metric, error)
}

// PluginProcessorClient A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error)
}

// PluginPublisherClient A client providing publishing specific plugin method calls.
type PluginPublisherClient interface {
	PluginClient
	Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error
}

type pluginCacheClient struct{}

// AllCacheHits returns cache hits across all metrics.
func (c *pluginCacheClient) AllCacheHits() uint64 {
	var hits uint64
	for _, v := range metricCache.table {
		hits += v.hits
	}
	return hits
}

// AllCacheMisses returns cache misses across all metrics.
func (c *pluginCacheClient) AllCacheMisses() uint64 {
	var misses uint64
	for _, v := range metricCache.table {
		misses += v.misses
	}
	return misses
}

// CacheHits returns the cache hits for a given metric namespace and version.
func (c *pluginCacheClient) CacheHits(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := metricCache.table[key]; ok {
		return v.hits, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}

// CacheMisses returns the cache misses for a given metric namespace and version.
func (c *pluginCacheClient) CacheMisses(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := metricCache.table[key]; ok {
		return v.misses, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}

// checkCache checks the cache for metric types.
// returns:
//  - array of metrics that need to be collected
//  - array of metrics that were returned from the cache
func checkCache(mts []core.Metric) ([]plugin.PluginMetricType, []core.Metric) {
	var fromCache []core.Metric
	var metricsToCollect []plugin.PluginMetricType
	for _, mt := range mts {
		if m := metricCache.get(core.JoinNamespace(mt.Namespace()), mt.Version()); m != nil {
			switch metric := m.(type) {
			case core.Metric:
				fromCache = append(fromCache, metric)
			case []core.Metric:
				for _, met := range metric {
					fromCache = append(fromCache, met)
				}
			default:
				log.WithFields(log.Fields{
					"_module": "client",
					"_block":  "checkCache",
				}).Error("unsupported type found in the cache")
			}
		} else {
			mt := plugin.PluginMetricType{
				Namespace_:          mt.Namespace(),
				LastAdvertisedTime_: mt.LastAdvertisedTime(),
				Version_:            mt.Version(),
				Tags_:               mt.Tags(),
				Labels_:             mt.Labels(),
				Config_:             mt.Config(),
			}
			metricsToCollect = append(metricsToCollect, mt)
		}
	}
	return metricsToCollect, fromCache
}

// updateCache updates the cache with the given array of metrics.
func updateCache(mts []plugin.PluginMetricType) {
	results := []core.Metric{}
	dc := map[string][]core.Metric{}
	for _, mt := range mts {
		if mt.Labels == nil {
			// cache the individual metric
			metricCache.put(core.JoinNamespace(mt.Namespace_), mt.Version(), mt)
		} else {
			// collect the dynamic query results so we can cache
			ns := make([]string, len(mt.Namespace()))
			copy(ns, mt.Namespace())
			for _, label := range mt.Labels_ {
				ns[label.Index] = "*"
			}
			if _, ok := dc[core.JoinNamespace(ns)]; !ok {
				dc[core.JoinNamespace(ns)] = []core.Metric{}
			}
			dc[core.JoinNamespace(ns)] = append(dc[core.JoinNamespace(ns)], mt)
			metricCache.put(core.JoinNamespace(ns), mt.Version(), dc[core.JoinNamespace(ns)])
		}
		results = append(results, mt)
	}
}
