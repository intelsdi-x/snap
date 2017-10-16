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

package strategy

import (
	"github.com/intelsdi-x/snap/core"
	log "github.com/sirupsen/logrus"
)

import (
	"fmt"
	"time"
)

// config-based provides a strategy that selects plugin based on given config
type configBased struct {
	plugins     map[string]AvailablePlugin
	metricCache map[string]*cache
	logger      *log.Entry
	cacheTTL    time.Duration
}

func NewConfigBased(cacheTTL time.Duration) *configBased {
	return &configBased{
		metricCache: make(map[string]*cache),
		plugins:     make(map[string]AvailablePlugin),
		cacheTTL:    cacheTTL,
		logger: log.WithFields(log.Fields{
			"_module": "control-routing",
		}),
	}
}

// Select selects an available plugin using the config based plugin strategy.
func (cb *configBased) Select(aps []AvailablePlugin, id string) (AvailablePlugin, error) {
	if ap, ok := cb.plugins[id]; ok && ap != nil {
		return ap, nil
	}

	// add first one in case it's new id
	for _, ap := range aps {
		available := true
		for _, busyPlugin := range cb.plugins {
			if ap == busyPlugin {
				available = false
			}
		}
		if available {
			cb.plugins[id] = ap
			return ap, nil
		}
	}
	cb.logger.WithFields(log.Fields{
		"_block":   "findAvailablePlugin",
		"strategy": cb.String(),
		"error":    fmt.Sprintf("%v of %v plugins are available", len(aps)-len(cb.plugins), len(aps)),
	}).Error(ErrCouldNotSelect)
	return nil, ErrCouldNotSelect
}

// Remove selects a plugin and and removes it from the cache
func (cb *configBased) Remove(aps []AvailablePlugin, id string) (AvailablePlugin, error) {
	ap, err := cb.Select(aps, id)
	if err != nil {
		return nil, err
	}
	delete(cb.metricCache, id)
	delete(cb.plugins, id)
	return ap, nil
}

// String returns the strategy name.
func (cb *configBased) String() string {
	return "config-based"
}

// CacheTTL returns the TTL for the cache.
func (cb *configBased) CacheTTL(id string) (time.Duration, error) {
	return cb.cacheTTL, nil
}

// checkCache checks the cache for metric types.
// returns:
//  - array of metrics that need to be collected
//  - array of metrics that were returned from the cache
func (cb *configBased) CheckCache(mts []core.Metric, id string) ([]core.Metric, []core.Metric) {
	if _, ok := cb.metricCache[id]; !ok {
		cb.metricCache[id] = NewCache(cb.cacheTTL)
	}
	return cb.metricCache[id].checkCache(mts)
}

// updateCache updates the cache with the given array of metrics.
func (cb *configBased) UpdateCache(mts []core.Metric, id string) {
	if _, ok := cb.metricCache[id]; !ok {
		cb.metricCache[id] = NewCache(cb.cacheTTL)
	}
	cb.metricCache[id].updateCache(mts)
}

// AllCacheHits returns cache hits across all metrics.
func (cb *configBased) AllCacheHits() uint64 {
	var total uint64
	for _, cache := range cb.metricCache {
		total += cache.allCacheHits()
	}
	return total
}

// AllCacheMisses returns cache misses across all metrics.
func (cb *configBased) AllCacheMisses() uint64 {
	var total uint64
	for _, cache := range cb.metricCache {
		total += cache.allCacheMisses()
	}
	return total
}

// CacheHits returns the cache hits for a given metric namespace and version.
func (cb *configBased) CacheHits(ns string, version int, id string) (uint64, error) {
	if cache, ok := cb.metricCache[id]; ok {
		return cache.cacheHits(ns, version)
	}
	return 0, ErrCacheDoesNotExist
}

// CacheMisses returns the cache misses for a given metric namespace and version.
func (cb *configBased) CacheMisses(ns string, version int, id string) (uint64, error) {
	if cache, ok := cb.metricCache[id]; ok {
		return cache.cacheMisses(ns, version)
	}
	return 0, ErrCacheDoesNotExist
}
