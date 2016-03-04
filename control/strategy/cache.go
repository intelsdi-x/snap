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
	"github.com/intelsdi-x/snap/pkg/chrono"
)

// GlobalCacheExpiration the default time limit for which a cache entry is valid.
// A plugin can override the GlobalCacheExpiration (default).
var GlobalCacheExpiration time.Duration

var (
	cacheLog = log.WithField("_module", "routing-cache")

	ErrCacheEntryDoesNotExist = errors.New("cache entry does not exist")
)

type cachecell struct {
	// Time this cachecell was created
	time time.Time
	// Metric stored in this cachecell
	metric core.Metric
	// Number of cache hits for this entry
	hits uint64
	// Number of cache misses for this entry
	misses uint64
}

type cache struct {
	// Stores invidual metrics with a concrete namespace
	table map[string]*cachecell
	// Stores a list of individual metrics with a wildcard namespace.
	lookaside map[string][]string
	// Length of time cache entry is valid
	ttl time.Duration
	// Total cache hits
	hits uint64
	// Total cache misses
	misses uint64
}

func NewCache(expiration time.Duration) *cache {
	return &cache{
		table:     make(map[string]*cachecell),
		lookaside: map[string][]string{},
		ttl:       expiration,
	}
}

// Given a metric returns the key to be used to access
// that metric in the cache
func genKeyFromMetric(m core.Metric) string {
	return fmt.Sprintf("%v:%v", core.JoinNamespace(m.Namespace()), m.Version())
}

func (c *cache) get(key string) core.Metric {
	var (
		cell *cachecell
		ok   bool
	)

	if cell, ok = c.table[key]; ok && chrono.Chrono.Now().Sub(cell.time) < c.ttl {
		cell.hits++
		c.hits++
		cacheLog.WithFields(log.Fields{
			"namespace": key,
			"hits":      cell.hits,
			"misses":    cell.misses,
		}).Debug(fmt.Sprintf("cache hit [%s]", key))
		return cell.metric
	}
	if !ok {
		c.table[key] = &cachecell{
			time:   time.Time{},
			metric: nil,
		}
	}
	c.table[key].misses++
	c.misses++
	cacheLog.WithFields(log.Fields{
		"namespace": key,
		"hits":      c.table[key].hits,
		"misses":    c.table[key].misses,
	}).Debug(fmt.Sprintf("cache miss [%s]", key))
	return nil
}

func (c *cache) put(key string, m core.Metric) {
	if _, ok := c.table[key]; !ok {
		c.table[key] = &cachecell{
			time:   time.Time{},
			metric: nil,
		}
	}
	c.table[key].metric = m
	c.table[key].time = chrono.Chrono.Now()
}

func (c *cache) checkCache(mts []core.Metric) (metricsToCollect []core.Metric, fromCache []core.Metric) {
	for _, mt := range mts {
		if hasStar(mt.Namespace()) {
			// If the namespace has a star it is a wildcard metric and we will check lookaside table
			lookups, ok := c.lookaside[genKeyFromMetric(mt)]
			if !ok {
				// If we have a wildcarded entry that is not in the lookaside table, add
				// it to the metrics to be collected and increment the global miss counter.
				metricsToCollect = append(metricsToCollect, mt)
				c.misses++
			} else {
				// Look for all entries that we received from the lookaside for the
				// given wildcard metric. If we find them all in the cache add them to
				// fromCache. If not add the wildcard metrics to metricsToCollect.
				// This enforces 'all or nothing' for wildcard metrics w/r/t cache lookups.
				metrics := make([]core.Metric, 0, len(lookups))
				for _, lookup := range lookups {
					if m := c.get(lookup); m != nil {
						metrics = append(metrics, m)
					} else {
						break
					}
				}
				if len(metrics) != len(lookups) {
					metricsToCollect = append(metricsToCollect, mt)
				} else {
					fromCache = append(fromCache, metrics...)
				}
			}
		} else if m := c.get(genKeyFromMetric(mt)); m != nil {
			fromCache = append(fromCache, m)
		} else {
			metricsToCollect = append(metricsToCollect, mt)
		}
	}
	return metricsToCollect, fromCache
}

func hasStar(s []string) bool {
	for _, ch := range s {
		if ch == "*" {
			return true
		}
	}
	return false
}

func (c *cache) updateCache(mts []core.Metric) {
	dc := map[string][]string{}
	for _, mt := range mts {
		// cache the individual metric
		c.put(genKeyFromMetric(mt), mt)
		// if we have labels, get labled info to put in lookaside
		if mt.Labels() != nil {
			// collect the dynamic query results so we can cache
			ns := make([]string, len(mt.Namespace()))
			copy(ns, mt.Namespace())
			for _, label := range mt.Labels() {
				ns[label.Index] = "*"
			}
			// Generate key here instead of genKeyFromMetric because we have modified
			// the namespace to replace labeled indexes with *.
			key := fmt.Sprintf("%v:%v", core.JoinNamespace(ns), mt.Version())
			if _, ok := dc[key]; !ok {
				dc[key] = []string{}
			}
			dc[key] = append(dc[key], genKeyFromMetric(mt))
		}
	}
	for k, v := range dc {
		c.lookaside[k] = v
	}
}

func (c *cache) allCacheHits() uint64 {
	return c.hits
}

func (c *cache) allCacheMisses() uint64 {
	return c.misses
}

func (c *cache) cacheHits(mt core.Metric) (uint64, error) {
	key := genKeyFromMetric(mt)
	if v, ok := c.table[key]; ok {
		return v.hits, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}

func (c *cache) cacheMisses(mt core.Metric) (uint64, error) {
	key := genKeyFromMetric(mt)
	if v, ok := c.table[key]; ok {
		return v.misses, nil
	}
	return 0, ErrCacheEntryDoesNotExist
}
