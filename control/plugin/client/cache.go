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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/core"
)

// the time limit for which a cache entry is valid.
var CacheExpiration time.Duration

var (
	metricCache = cache{
		table: make(map[string]*cachecell),
	}
	cacheLog = log.WithField("_module", "client-cache")
)

type cachecell struct {
	time   time.Time
	metric core.Metric
	hits   uint64
	misses uint64
}

type cache struct {
	table map[string]*cachecell
}

func (c *cache) get(key string) core.Metric {
	var (
		cell *cachecell
		ok   bool
	)
	if cell, ok = c.table[key]; ok && time.Since(cell.time) < CacheExpiration {
		cell.hits++
		cacheLog.WithFields(log.Fields{
			"namespace": key,
			"hits":      cell.hits,
			"misses":    cell.misses,
		}).Debug(fmt.Sprintf("cache hit [%s]", key))
		return cell.metric
	}
	if ok {
		cell.misses++
		cacheLog.WithFields(log.Fields{
			"namespace": key,
			"hits":      cell.hits,
			"misses":    cell.misses,
		}).Debug(fmt.Sprintf("cache miss [%s]", key))
	}
	return nil
}

func (c *cache) put(key string, metric core.Metric) {
	if _, ok := c.table[key]; ok {
		c.table[key].time = time.Now()
		c.table[key].metric = metric
	} else {
		c.table[key] = &cachecell{
			time:   time.Now(),
			metric: metric,
		}
	}
}
