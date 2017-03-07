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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
)

// lru provides a strategy that selects the least recently used available plugin.
type lru struct {
	*cache
	logger *log.Entry
}

func NewLRU(cacheTTL time.Duration) *lru {
	return &lru{
		NewCache(cacheTTL),
		log.WithFields(log.Fields{
			"_module": "control-routing",
		}),
	}
}

// String returns the strategy name.
func (l *lru) String() string {
	return "least-recently-used"
}

// CacheTTL returns the TTL for the cache.
func (l *lru) CacheTTL(taskID string) (time.Duration, error) {
	return l.ttl, nil
}

// Select selects an available plugin using the least-recently-used strategy.
func (l *lru) Select(aps []AvailablePlugin, _ string) (AvailablePlugin, error) {
	t := time.Now()
	index := -1
	for i, ap := range aps {
		// look for the least recently used
		if ap.LastHit().Before(t) || index == -1 {
			index = i
			t = ap.LastHit()
		}
	}
	if index > -1 {
		l.logger.WithFields(log.Fields{
			"block":     "select",
			"strategy":  l.String(),
			"pool size": len(aps),
			"index":     aps[index].String(),
			"hitcount":  aps[index].HitCount(),
		}).Debug("plugin selected")
		return aps[index], nil
	}
	l.logger.WithFields(log.Fields{
		"block":    "select",
		"strategy": l.String(),
		"error":    ErrCouldNotSelect,
	}).Error("error selecting")
	return nil, ErrCouldNotSelect
}

// Remove selects a plugin
// Since there is no state to cleanup we only need to return the selected plugin
func (l *lru) Remove(aps []AvailablePlugin, taskID string) (AvailablePlugin, error) {
	ap, err := l.Select(aps, taskID)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

// checkCache checks the cache for metric types.
// returns:
//  - array of metrics that need to be collected
//  - array of metrics that were returned from the cache
func (l *lru) CheckCache(mts []core.Metric, _ string) ([]core.Metric, []core.Metric) {
	return l.checkCache(mts)
}

// updateCache updates the cache with the given array of metrics.
func (l *lru) UpdateCache(mts []core.Metric, _ string) {
	l.updateCache(mts)
}

// AllCacheHits returns cache hits across all metrics.
func (l *lru) AllCacheHits() uint64 {
	return l.allCacheHits()
}

// AllCacheMisses returns cache misses across all metrics.
func (l *lru) AllCacheMisses() uint64 {
	return l.allCacheMisses()
}

// CacheHits returns the cache hits for a given metric namespace and version.
func (l *lru) CacheHits(ns string, version int, _ string) (uint64, error) {
	return l.cacheHits(ns, version)
}

// CacheMisses returns the cache misses for a given metric namespace and version.
func (l *lru) CacheMisses(ns string, version int, _ string) (uint64, error) {
	return l.cacheMisses(ns, version)
}
