// internals of facter cache management
package facter

import (
	"errors"
	"sync"
	"time"
)

// metricCache caches values received from cmd to not overuse system under monitor
type metricCache struct {
	ttl   time.Duration
	data  map[string]entry
	mutex sync.Mutex
	// injects implementation for getting facts - defaults to use getFacts from cmd.go
	// but allows to replace with fake during tests
	getFacts func(
		names []string,
		facterTimeout time.Duration,
		cmdConfig *cmdConfig,
	) (*facts, *time.Time, error)
	// how much time we are ready to wait for getting result from cmd
	facterExecutionDeadline time.Duration
}

// newMetricCache create new metricCache object with given ttl and deadline for cmd
func newMetricCache(ttl, facterExecutionDeadline time.Duration) *metricCache {
	return &metricCache{
		ttl:                     ttl,
		data:                    map[string]entry{},
		mutex:                   sync.Mutex{},
		getFacts:                getFacts,
		facterExecutionDeadline: facterExecutionDeadline,
	}
}

// getEntry allow to extrac just one entry from cache
func (c *metricCache) getEntry(name string) entry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.data[name]
}

// size returns number of all entries actully in cache
func (c *metricCache) size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data)
}

// entries returns a copy of cache as map of entries
func (c *metricCache) entries() map[string]entry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	m := make(map[string]entry, len(c.data))
	for name, entry := range c.data {
		m[name] = entry
	}
	return m
}

// getNamesToUpdate compares given fact names with cache state
// and prepare a list of stale or non-existing ones
// returns the names of metrics that should have to be updated
func (c *metricCache) getNamesToUpdate(names []string) []string {
	// protect the c.data
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	// check every cache entry is ok (stale/exists?)
	namesToUpdate := []string{}
	for _, name := range names {

		fact, exists := c.data[name]

		// assume it is stale
		// stale also stays true for not existin ones
		stale := true
		if exists {
			stale = now.Sub(fact.lastUpdate) > c.ttl
		}
		if stale {
			namesToUpdate = append(namesToUpdate, name)
		}
	}
	return namesToUpdate
}

// synchronizeCache is responsible for updating metrics in cache (conditionally)
// only if there is a need for that
// names is slice with list of metrics to synchronize
// names cannot be empty
func (c *metricCache) synchronizeCache(names []string) error {

	// check not empty argument
	if len(names) == 0 {
		return errors.New("I cannot synchronize cache for empty name list!")
	}

	// what is needed to be updated
	namesToUpdate := c.getNamesToUpdate(names)

	// if there is something that has to refreshed - refresh it
	if len(namesToUpdate) > 0 {

		err := c.updateCache(namesToUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateCache updates (refresh) cache entries (unconditionally) with current values
// from external binary
// you can pass empty names collection to force update everything
func (c *metricCache) updateCache(
	names []string,
) error {

	// obtain actual facts (with default cmd config)
	facts, receviedAt, err := c.getFacts(
		names, // facts: what to update
		c.facterExecutionDeadline, // timeout
		nil, // default options "facter --json"
	)
	if err != nil {
		return err
	}

	// if names was empty, we want update all facts
	// so extract all fact names
	if len(names) == 0 {
		for name, _ := range *facts {
			names = append(names, name)
		}
	}

	// protect the c.data
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update cache with new fact values
	for _, name := range names {

		// update unconditionally value in cache
		entry := c.data[name]
		entry.lastUpdate = *receviedAt
		entry.value = (*facts)[name] // extract raw fact value received from Facter
		c.data[name] = entry

	}
	return nil
}

// updateCacheAll updates all cache entries unconditionally (just a wrapper for updateCache for all metrics)
func (c *metricCache) updateCacheAll() error {
	return c.updateCache([]string{})
}

// --------------- helper types for metricCache ----------------

// helper type to deal with json values which additionally stores last update moment
type entry struct {
	value      interface{}
	lastUpdate time.Time
}
