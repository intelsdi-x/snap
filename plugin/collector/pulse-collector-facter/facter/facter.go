package facter

import (
	"errors"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

/*******************
 *  pulse plugin  *
 *******************/

// TODO: all this should be private
const (
	vendor = "intel"
	prefix = "facter"
	// how long we are caching the date from external binary to prevent overuse of resources
	defaultCacheTTL = 60 * time.Second
	// how long are we going to cacha avaiable types of metrics
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

	cache    map[string]fact
	cacheTTL time.Duration

	facterExecutionDeadline time.Duration

	// injects implementation for getting facts - defaults to use getFacts from cmd.go
	// but allows to replace with fake during tests
	getFacts func(keys []string, facterTimeout time.Duration, cmdConfig *cmdConfig) (*stringmap, *time.Time, error)
}

// NewFacter constructs new Facter with default values
func NewFacter() *Facter {
	return &Facter{
		metricTypesTTL: defaultMetricTypesTTL,
		cacheTTL:       defaultCacheTTL,
		cache:          map[string]fact{},
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
// acceptble by Pulse
func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	// TODO: INPUT: read CollectorArgs structure to extract requested metrics
	requestedNames := []string{"kernel", "uptime"}

	// synchronize cache (stale of missing data) to have all we need
	err := f.synchronizeCache(requestedNames)
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
// only if there a need for that
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

// updateCache updates (refresh) cache  entries (unconditionally) with current values
// from facter, pass empty collection to update all facts in cache
func (f *Facter) updateCache(names []string) error {

	// obtain actual facts (with default cmd config)
	facts, receviedAt, err := f.getFacts(
		names, // what
		f.facterExecutionDeadline, // timeout
		nil, // default options "facter --json"
	)
	if err != nil {
		return err
	}

	// if names was empty, we want update all facts
	if len(names) == 0 {
		for factName, _ := range *facts {
			names = append(names, factName)
		}
	}

	// merge cache with new facts
	for _, name := range names {
		// create fact if not exists yet
		value := (*facts)[name]
		if _, exists := f.cache[name]; !exists {
			f.cache[name] = fact{
				lastUpdate: *receviedAt,
				value:      value,
			}
		} else {
			// just update the value in cache
			fact := f.cache[name]
			fact.lastUpdate = *receviedAt
			fact.value = value
			f.cache[name] = fact // updating a field in non-addressable value in map golang defect #3117
		}
	}
	return nil
}

// updateCacheAll updates all cache entries unconditionally (just a wrapper for updateCache for all metrics)
func (f *Facter) updateCacheAll() error {
	return f.updateCache([]string{})
}

// prepareMetricTypes fills metricTypes internal collection ready to send back to pulse
func (f *Facter) prepareMetricTypes() {

	// new temporary collection (to not mess with
	metricTypes := make([]*plugin.MetricType, 0, len(f.cache))

	// rewrite values from cache to another colllection accetable by pulse
	for factName, value := range f.cache {
		metricTypes = append(metricTypes,
			plugin.NewMetricType(
				[]string{vendor, prefix, factName}, // namespace
				value.lastUpdate.Unix()),           // lastAdvertisedTimestamp TODO would be time.Now()
		)
	}

	// update Facter state
	f.metricTypes = metricTypes

	// remember the last the metricTypes was filled
	f.metricTypesLastUpdate = time.Now()
}

// helper type to deal with json that stores last update moment
// allows to implement a local cache in PluginFacter
type fact struct {
	value      interface{}
	lastUpdate time.Time
}
