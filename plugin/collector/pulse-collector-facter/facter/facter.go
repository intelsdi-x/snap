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

const (
	Name   = "Intel Facter Plugin (c) 2015 Intel Corporation"
	Vendor = "intel"
	prefix = "facter"

	Version               = 1
	Type                  = plugin.CollectorPluginType
	DefaultCacheTTL       = 60 * time.Second
	DefaultMetricTypesTTL = DefaultCacheTTL
	DefautlFacterDeadline = 5 * time.Second
)

/*****************************************
 *  pulse public methods implementation  *
 *****************************************/

// returns PluginMeta
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

// returns ConfigPolicy
func ConfigPolicy() *plugin.ConfigPolicy {
	//TODO What is plugin policy?

	c := new(plugin.ConfigPolicy)
	return c
}

/*******************
 *  Facter struct  *
 *******************/

// wrapper to use facter binary, responsible for running external binary and caching
type Facter struct {
	metricTypes           []*plugin.MetricType //map[string]interface{}
	metricTypesLastUpdate time.Time
	metricTypesTTL        time.Time

	cacheTTL time.Duration
	cache    map[string]fact

	facterExecutionDeadline time.Duration

	// injects implementation for getting facts - defaults to use getFacts from cmd.go
	// but allows to replace with fake during tests
	getFacts func(keys []string, facterTimeout time.Duration) (*stringmap, *time.Time, error)
}

// construct new Facter
func NewFacter() *Facter {
	f := new(Facter)
	//TODO read from config
	f.cacheTTL = DefaultCacheTTL
	f.cache = make(map[string]fact)
	f.facterExecutionDeadline = DefautlFacterDeadline
	f.getFacts = getFacts
	return f
}

// Pulse plugin interface implementation
// TODO: to be rewritten to be compatible to pulse API
// ----------------------------------------------------

// get available metrics types
func (f *Facter) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	// TODO: handle plugin.GetMetricTypesArgs somehow

	// synchronize cache conditionally as a whole
	timeElapsed := time.Since(f.metricTypesLastUpdate)
	needUpdate := timeElapsed > DefaultMetricTypesTTL
	if needUpdate {
		// synchronize cache conditionally for fields
		err := f.updateCacheAll()
		if err != nil {
			log.Println("Facter: synchronizeCache returned error: " + err.Error())
			return err
		}

		// transform internal cache to metricTypes
		f.prepareMetricTypes()

	}
	// return metricTypes prepared earlier
	reply.MetricTypes = f.metricTypes

	return nil
}

// collect metrics
func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	// TODO: INPUT: read CollectorArgs structure to extract requested metrics
	requestedNames := []string{"kernel", "uptime"}

	// prepare cache to have all we need
	err := f.synchronizeCache(requestedNames)
	if err != nil {
		return err
	}

	// TODO: OUTPUT: fulfill reply structure with requested metrics
	// for _, name := range requestedNames {
	// 	// convert it some how if required
	// 	reply.metrics[name] = f.cache[name].value
	// }

	return nil
}

// internals (cache management)
// ------------------------------------

// compare given facts with cache state and prepare a list of stale or non-existing ones
// returns the names of metrics that have to be updated
func (f *Facter) getNamesToUpdate(names []string) []string {

	now := time.Now()

	// list of facts that have to be updated because are old
	// or acquired first time
	// so collect stale or not existing facts
	namesToUpdate := []string{}
	for _, name := range names {

		fact, exists := f.cache[name]
		// fact doesn't exist or is stale
		stale := false
		if exists {
			// is it stale ?
			stale = now.Sub(fact.lastUpdate) > f.cacheTTL
		}
		if !exists || stale {
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

	// obtain actual facts
	facts, receviedAt, err := f.getFacts(names, f.facterExecutionDeadline)
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
				[]string{Vendor, prefix, factName}, // namespace
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
