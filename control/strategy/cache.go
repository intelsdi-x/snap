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

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/chrono"
	log "github.com/sirupsen/logrus"
)

// GlobalCacheExpiration the default time limit for which a cache entry is valid.
// A plugin can override the GlobalCacheExpiration (default).
var GlobalCacheExpiration time.Duration

var (
	cacheLog = log.WithField("_module", "routing-cache")

	ErrCacheEntryDoesNotExist = errors.New("cache entry does not exist")
)

type cachecell struct {
	time    time.Time
	metric  core.Metric
	metrics []core.Metric
	hits    uint64
	misses  uint64
}

type cache struct {
	table map[string]*cachecell
	ttl   time.Duration
}

func NewCache(expiration time.Duration) *cache {
	return &cache{
		table: make(map[string]*cachecell),
		ttl:   expiration,
	}
}

func (c *cache) get(ns string, version int) interface{} {
	var (
		cell *cachecell
		ok   bool
	)

	key := fmt.Sprintf("%v:%v", ns, version)
	if cell, ok = c.table[key]; ok && chrono.Chrono.Now().Sub(cell.time) < c.ttl {
		cell.hits++
		cacheLog.WithFields(log.Fields{
			"namespace": key,
			"hits":      cell.hits,
			"misses":    cell.misses,
		}).Debug(fmt.Sprintf("cache hit [%s]", key))
		if cell.metric != nil {
			return cell.metric
		}
		return cell.metrics
	}
	if !ok {
		c.table[key] = &cachecell{
			time:    time.Time{},
			metrics: nil,
		}
	}
	c.table[key].misses++
	cacheLog.WithFields(log.Fields{
		"namespace": key,
		"hits":      c.table[key].hits,
		"misses":    c.table[key].misses,
	}).Debug(fmt.Sprintf("cache miss [%s]", key))
	return nil
}

func (c *cache) put(ns string, version int, m interface{}) {
	key := fmt.Sprintf("%v:%v", ns, version)
	switch metric := m.(type) {
	case core.Metric:
		if _, ok := c.table[key]; ok {
			c.table[key].time = chrono.Chrono.Now()
			c.table[key].metric = metric
		} else {
			c.table[key] = &cachecell{
				time:   chrono.Chrono.Now(),
				metric: metric,
			}
		}
	case []core.Metric:
		if _, ok := c.table[key]; ok {
			c.table[key].time = chrono.Chrono.Now()
			c.table[key].metrics = metric
		} else {
			c.table[key] = &cachecell{
				time:    chrono.Chrono.Now(),
				metrics: metric,
			}
		}
	default:
		cacheLog.WithFields(log.Fields{
			"namespace": key,
			"_block":    "put",
		}).Error("unsupported type")
	}
}

func (c *cache) checkCache(mts []core.Metric) (metricsToCollect []core.Metric, fromCache []core.Metric) {
	for _, mt := range mts {
		if m := c.get(mt.Namespace().String(), mt.Version()); m != nil {
			switch metric := m.(type) {
			case core.Metric:
				fromCache = append(fromCache, metric)
			case []core.Metric:
				for _, met := range metric {
					fromCache = append(fromCache, met)
				}
			default:
				cacheLog.WithFields(log.Fields{
					"_block": "checkCache",
				}).Error("unsupported type found in the cache")
			}
		} else {
			metricsToCollect = append(metricsToCollect, mt)
		}
	}
	return metricsToCollect, fromCache
}

type listMetricInfo struct {
	metrics   []core.Metric
	namespace string
	version   int
}

func (c *cache) updateCache(mts []core.Metric) {
	dc := map[string]*listMetricInfo{}
	for _, mt := range mts {
		isDynamic, idx := mt.Namespace().IsDynamic()
		if isDynamic {
			// cache dynamic metrics
			dynNS := make(core.Namespace, len(mt.Namespace()))
			copy(dynNS, mt.Namespace())
			for _, v := range idx {
				dynNS[v].Value = "*"
			}
			key := fmt.Sprintf("%v:%v", dynNS.String(), mt.Version())
			if _, ok := dc[key]; !ok {
				dc[key] = &listMetricInfo{
					metrics:   []core.Metric{},
					namespace: dynNS.String(),
					version:   mt.Version(),
				}
			}
			dc[key].metrics = append(dc[key].metrics, mt)
			continue
		}
		// cache the individual metric
		c.put(mt.Namespace().String(), mt.Version(), mt)
	}
	// write our dynamic metrics to the cache.
	for _, v := range dc {
		c.put(v.namespace, v.version, v.metrics)
	}
}

func (c *cache) allCacheHits() uint64 {
	var hits uint64
	for _, v := range c.table {
		hits += v.hits
	}
	return hits
}

func (c *cache) allCacheMisses() uint64 {
	var misses uint64
	for _, v := range c.table {
		misses += v.misses
	}
	return misses
}

func (c *cache) cacheHits(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := c.table[key]; ok {
		return v.hits, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}

func (c *cache) cacheMisses(ns string, version int) (uint64, error) {
	key := fmt.Sprintf("%v:%v", ns, version)
	if v, ok := c.table[key]; ok {
		return v.misses, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}
