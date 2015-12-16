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
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/pkg/chrono"
)

// GlobalCacheExpiration the time limit for which a cache entry is valid.
var GlobalCacheExpiration time.Duration

var (
	metricCache = cache{
		table: make(map[string]*cachecell),
	}
	cacheLog = log.WithField("_module", "client-cache")

	// ErrCacheEntryDoesNotExist - error message when a cache error does not exist
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
}

func (c *cache) get(ns string, version int) interface{} {
	var (
		cell *cachecell
		ok   bool
	)

	key := fmt.Sprintf("%v:%v", ns, version)
	if cell, ok = c.table[key]; ok && chrono.Chrono.Now().Sub(cell.time) < GlobalCacheExpiration {
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
			c.table[key].time = time.Now()
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
