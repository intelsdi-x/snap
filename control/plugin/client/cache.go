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
