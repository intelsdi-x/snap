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
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
)

var (
	ErrorCouldNotSelect = errors.New("could not select a plugin (round robin strategy)")
)

// lru provides a stragey that selects the least recently used available plugin.
type lru struct {
	metricCache *cache
	logger      *log.Entry
}

func NewLRU(cacheTTL time.Duration) *lru {
	return &lru{
		metricCache: NewCache(cacheTTL),
		logger: log.WithFields(log.Fields{
			"_module": "control-routing",
		}),
	}
}

func (l *lru) Cache() *cache {
	return l.metricCache
}

func (l *lru) String() string {
	return "least-recently-used"
}

func (l *lru) CacheTTL() time.Duration {
	return l.Cache().ttl
}

func (l *lru) Select(spa []SelectablePlugin) (SelectablePlugin, error) {
	t := time.Now()
	index := -1
	for i, sp := range spa {
		// look for the least recently used
		if sp.LastHit().Before(t) || index == -1 {
			index = i
			t = sp.LastHit()
		}
	}
	if index > -1 {
		l.logger.WithFields(log.Fields{
			"block":     "select",
			"strategy":  l.String(),
			"pool size": len(spa),
			"index":     spa[index].String(),
			"hitcount":  spa[index].HitCount(),
		}).Debug("plugin selected")
		return spa[index], nil
	}
	l.logger.WithFields(log.Fields{
		"block":    "select",
		"strategy": "round-robin",
		"error":    ErrorCouldNotSelect,
	}).Error("error selecting")
	return nil, ErrorCouldNotSelect
}

// checkCache checks the cache for metric types.
// returns:
//  - array of metrics that need to be collected
//  - array of metrics that were returned from the cache
func (l *lru) CheckCache(mts []core.Metric) ([]core.Metric, []core.Metric) {
	var fromCache []core.Metric
	var metricsToCollect []core.Metric
	for _, mt := range mts {
		if m := l.metricCache.get(core.JoinNamespace(mt.Namespace()), mt.Version()); m != nil {
			switch metric := m.(type) {
			case core.Metric:
				fromCache = append(fromCache, metric)
			case []core.Metric:
				for _, met := range metric {
					fromCache = append(fromCache, met)
				}
			default:
				l.logger.WithFields(log.Fields{
					"_module": "client",
					"_block":  "checkCache",
				}).Error("unsupported type found in the cache")
			}
		} else {
			metricsToCollect = append(metricsToCollect, mt)
		}
	}
	return metricsToCollect, fromCache
}

// updateCache updates the cache with the given array of metrics.
func (l *lru) UpdateCache(mts []core.Metric) {
	results := []core.Metric{}
	dc := map[string][]core.Metric{}
	for _, mt := range mts {
		if mt.Labels() == nil {
			// cache the individual metric
			l.metricCache.put(core.JoinNamespace(mt.Namespace()), mt.Version(), mt)
			l.logger.Debugf("putting %v:%v in the cache", mt.Namespace(), mt.Version())
		} else {
			// collect the dynamic query results so we can cache
			ns := make([]string, len(mt.Namespace()))
			copy(ns, mt.Namespace())
			for _, label := range mt.Labels() {
				ns[label.Index] = "*"
			}
			if _, ok := dc[core.JoinNamespace(ns)]; !ok {
				dc[core.JoinNamespace(ns)] = []core.Metric{}
			}
			dc[core.JoinNamespace(ns)] = append(dc[core.JoinNamespace(ns)], mt)
			l.metricCache.put(core.JoinNamespace(ns), mt.Version(), dc[core.JoinNamespace(ns)])
			l.logger.Debugf("putting %v:%v in the cache", ns, mt.Version())
		}
		results = append(results, mt)
	}
}

func (l *lru) AllCacheHits() uint64 {
	var hits uint64
	for _, v := range l.metricCache.table {
		hits += v.hits
	}
	return hits
}

// AllCacheMisses returns cache misses across all metrics.
func (l *lru) AllCacheMisses() uint64 {
	var misses uint64
	for _, v := range l.metricCache.table {
		misses += v.misses
	}
	return misses
}

// CacheHits returns the cache hits for a given metric namespace and version.
func (l *lru) CacheHits(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := l.metricCache.table[key]; ok {
		return v.hits, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}

// CacheMisses returns the cache misses for a given metric namespace and version.
func (l *lru) CacheMisses(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := l.metricCache.table[key]; ok {
		return v.misses, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}
