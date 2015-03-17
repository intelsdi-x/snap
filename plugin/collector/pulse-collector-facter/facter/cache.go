// internals of facter cache management
package facter

import (
	"errors"
	"time"
)

// getNamesToUpdate compares given fact names with cache state
// and prepare a list of stale or non-existing ones
// returns the names of metrics that should have to be updated
func (f *Facter) getNamesToUpdate(names []string) []string {

	now := time.Now()

	// check every cache entry is ok (stale/exists?)
	namesToUpdate := []string{}
	for _, name := range names {

		fact, exists := f.cache[name]

		// assume it is stale
		// stale also stays true for not existin ones
		stale := true
		if exists {
			stale = now.Sub(fact.lastUpdate) > f.cacheTTL
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
func (f *Facter) synchronizeCache(names []string) error {

	// check not empty argument
	if len(names) == 0 {
		return errors.New("I cannot synchronize cache for empty name list!")
	}

	// what is needed to be updated
	namesToUpdate := f.getNamesToUpdate(names)

	// if there is something that has to refreshed - refresh it
	if len(namesToUpdate) > 0 {

		err := f.updateCache(namesToUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateCache updates (refresh) cache entries (unconditionally) with current values
// from external binary
// you can pass empty names collection to force update everything
func (f *Facter) updateCache(names []string) error {

	// obtain actual facts (with default cmd config)
	facts, receviedAt, err := f.getFacts(
		names, // facts: what to update
		f.facterExecutionDeadline, // timeout
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

	// update cache with new fact values
	for _, name := range names {

		// update unconditionally value in cache
		entry := f.cache[name]
		entry.lastUpdate = *receviedAt
		entry.value = (*facts)[name] // extract raw fact value received from Facter
		f.cache[name] = entry

	}
	return nil
}

// updateCacheAll updates all cache entries unconditionally (just a wrapper for updateCache for all metrics)
func (f *Facter) updateCacheAll() error {
	return f.updateCache([]string{})
}

/**********************
 *  helper fact type  *
 **********************/

// helper type to deal with json values which additionally stores last update moment
type entry struct {
	value      interface{}
	lastUpdate time.Time
}
