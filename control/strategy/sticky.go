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

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core"
)

import "time"

var (
	ErrCacheDoesNotExist = errors.New("cache does not exist")
)

// sticky provides a strategy that ... concurrency count is 1
type sticky struct {
	plugins     map[string]AvailablePlugin
	metricCache map[string]*cache
	logger      *log.Entry
	cacheTTL    time.Duration
}

func NewSticky(cacheTTL time.Duration) *sticky {
	return &sticky{
		metricCache: make(map[string]*cache),
		plugins:     make(map[string]AvailablePlugin),
		cacheTTL:    cacheTTL,
		logger: log.WithFields(log.Fields{
			"_module": "control-routing",
		}),
	}
}

// Select selects an available plugin using the sticky plugin strategy.
func (s *sticky) Select(aps []AvailablePlugin, taskID string) (AvailablePlugin, error) {
	if ap, ok := s.plugins[taskID]; ok && ap != nil {
		return ap, nil
	}
	return s.selectPlugin(aps, taskID)
}

// Remove selects a plugin and and removes it from the cache
func (s *sticky) Remove(aps []AvailablePlugin, taskID string) (AvailablePlugin, error) {
	ap, err := s.Select(aps, taskID)
	if err != nil {
		return nil, err
	}
	delete(s.metricCache, taskID)
	delete(s.plugins, taskID)
	return ap, nil
}

// String returns the strategy name.
func (s *sticky) String() string {
	return "sticky"
}

// CacheTTL returns the TTL for the cache.
func (s *sticky) CacheTTL(taskID string) (time.Duration, error) {
	return s.cacheTTL, nil
}

// checkCache checks the cache for metric types.
// returns:
//  - array of metrics that need to be collected
//  - array of metrics that were returned from the cache
func (s *sticky) CheckCache(mts []core.Metric, taskID string) ([]core.Metric, []core.Metric) {
	if _, ok := s.metricCache[taskID]; !ok {
		s.metricCache[taskID] = NewCache(s.cacheTTL)
	}
	return s.metricCache[taskID].checkCache(mts)
}

// updateCache updates the cache with the given array of metrics.
func (s *sticky) UpdateCache(mts []core.Metric, taskID string) {
	if _, ok := s.metricCache[taskID]; !ok {
		s.metricCache[taskID] = NewCache(s.cacheTTL)
	}
	s.metricCache[taskID].updateCache(mts)
}

// AllCacheHits returns cache hits across all metrics.
func (s *sticky) AllCacheHits() uint64 {
	var total uint64
	for _, cache := range s.metricCache {
		total += cache.allCacheHits()
	}
	return total
}

// AllCacheMisses returns cache misses across all metrics.
func (s *sticky) AllCacheMisses() uint64 {
	var total uint64
	for _, cache := range s.metricCache {
		total += cache.allCacheMisses()
	}
	return total
}

// CacheHits returns the cache hits for a given metric namespace and version.
func (s *sticky) CacheHits(ns string, version int, taskID string) (uint64, error) {
	if cache, ok := s.metricCache[taskID]; ok {
		return cache.cacheHits(ns, version)
	}
	return 0, ErrCacheDoesNotExist
}

// CacheMisses returns the cache misses for a given metric namespace and version.
func (s *sticky) CacheMisses(ns string, version int, taskID string) (uint64, error) {
	if cache, ok := s.metricCache[taskID]; ok {
		return cache.cacheMisses(ns, version)
	}
	return 0, ErrCacheDoesNotExist
}

func (s *sticky) selectPlugin(aps []AvailablePlugin, taskID string) (AvailablePlugin, error) {
	for _, ap := range aps {
		available := true
		for _, busyPlugin := range s.plugins {
			if ap == busyPlugin {
				available = false
			}
		}
		if available {
			s.plugins[taskID] = ap
			return ap, nil
		}
	}
	s.logger.WithFields(log.Fields{
		"_block":   "findAvailablePlugin",
		"strategy": s.String(),
		"error":    fmt.Sprintf("%v of %v plugins are available", len(aps)-len(s.plugins), len(aps)),
	}).Error(ErrCouldNotSelect)
	return nil, ErrCouldNotSelect
}
