/* This modules converts implements Pulse API to become an plugin.
Additionally it caches all data to protect the system against overuse.

Implementation details:
legend:
- metric - represents value of metric from Pulse side
- fact - represents a value about a system gathered from Facter
- name - is string identifier that refers to metric from the Pulse side, so name points to metric

*/

package facter

import (
	"errors"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	vendor = "intel"
	prefix = "facter"
	// how long we are caching the date from external binary to prevent overuse of resources
	defaultCacheTTL = 60 * time.Second
	// how long are we going to cache available types of metrics
	defaultMetricTypesTTL = defaultCacheTTL
	// deadline a.k.a. timeout we are ready to wait for external binary to gather the data
	defaultFacterDeadline = 5 * time.Second
)

/*****************************************
 *  pulse public methods implementation  *
 *****************************************/

// returns PluginMeta
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		"Intel Fact Gathering Plugin", // name
		1, // version
		plugin.CollectorPluginType, // pluginType
	)
}

// returns ConfigPolicy
//TODO What is plugin policy? just mock for now
func ConfigPolicy() *plugin.ConfigPolicy {
	return new(plugin.ConfigPolicy)
}

/**********
 * Facter *
 **********/

// Facter implements API to communicate with Pulse
type Facter struct {
	// available metrics that can me returned
	metricTypes           []*plugin.MetricType
	metricTypesLastUpdate time.Time
	metricTypesTTL        time.Duration

	cache    map[string]entry
	cacheTTL time.Duration

	facterExecutionDeadline time.Duration

	// injects implementation for getting facts - defaults to use getFacts from cmd.go
	// but allows to replace with fake during tests
	getFacts func(
		names []string,
		facterTimeout time.Duration,
		cmdConfig *cmdConfig,
	) (*facts, *time.Time, error)
}

// NewFacter constructs new Facter with default values
func NewFacter() *Facter {
	return &Facter{
		metricTypesTTL: defaultMetricTypesTTL,
		cacheTTL:       defaultCacheTTL,
		cache:          map[string]entry{},
		facterExecutionDeadline: defaultFacterDeadline,
		getFacts:                getFacts,
	}
}

// Pulse plugin interface implementation
// TODO: to be rewritten to be compatible to pulse API
// ----------------------------------------------------

// GetMetricTypes returns available metrics types
func (f *Facter) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	// TODO: handle plugin.GetMetricTypesArgs somehow

	// synchronize cache conditionally as a whole
	timeElapsed := time.Since(f.metricTypesLastUpdate)
	needUpdate := timeElapsed > f.metricTypesTTL
	if needUpdate {
		// synchronize cache conditionally for all fields
		err := f.updateCacheAll() // fills internal f.fcache
		if err != nil {
			log.Println("Facter: synchronizeCache returned error: " + err.Error())
			return err
		}

		// fill f.metricTypes based on f.cache
		f.prepareMetricTypes()

	}
	// return metricTypes prepared earlier
	reply.MetricTypes = f.metricTypes

	return nil
}

// Collect collects metrics from external binary a returns them in form
// acceptable by Pulse
func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	// TODO: INPUT: read CollectorArgs structure to extract requested metrics
	names := []string{"kernel", "uptime"}

	// synchronize cache (stale of missing data) to have all we need
	err := f.synchronizeCache(names)
	if err != nil {
		return err
	}

	// TODO: OUTPUT: fulfill reply structure with requested metrics
	// for _, name := range requestedNames {
	// 	// convert it some how to required format
	// 	reply.metrics[name] = f.cache[name].value
	// }

	return nil
}

// internals (cache management)
// ------------------------------------

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

// prepareMetricTypes fills metricTypes internal collection ready to send back to pulse
func (f *Facter) prepareMetricTypes() {

	// new temporary collection
	metricTypes := make([]*plugin.MetricType, 0, len(f.cache))

	// rewrite values from cache to another collection acceptable by Pulse
	for factName, value := range f.cache {
		metricTypes = append(metricTypes,
			plugin.NewMetricType(
				[]string{vendor, prefix, factName}, // namespace
				value.lastUpdate.Unix(),            // lastAdvertisedTimestamp TODO would be time.Now()
			),
		)
	}

	// update internal state
	f.metricTypes = metricTypes

	// remember the last the metricTypes was filled
	// to be confronted with f.metricTypesTTL
	f.metricTypesLastUpdate = time.Now()
}

/**********************
 *  helper fact type  *
 **********************/

// helper type to deal with json values which additionally stores last update moment
type entry struct {
	value      interface{}
	lastUpdate time.Time
}
