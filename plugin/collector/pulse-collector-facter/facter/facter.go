package facter

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
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

/*************
 *  facter   *
 *************/

// wrapper to use facter binnary, responsible for running external binnary and caching
type Facter struct {
	metricTypes          []*plugin.MetricType //map[string]interface{}
	metricTypesLastCheck time.Time
	metricTypesTTL       time.Time

	cacheTTL time.Duration
	cache    map[string]fact

	facterExecutionDeadline time.Duration
}

// construct new Facter
func NewFacterPlugin() *Facter {
	f := new(Facter)
	//TODO read from config
	f.cacheTTL = DefaultCacheTTL
	f.cache = make(map[string]fact)
	f.facterExecutionDeadline = DefautlFacterDeadline
	return f
}

//// fullfill the availableMetricTypes with data from facter
//func (f *Facter) loadAvailableMetricTypes() error {

//	// get all facts (empty slice)
//	facterMap, timestamp, err := getFacts([]string{}, f.facterExecutionDeadline)
//	if err != nil {
//		log.Println("FacterPlugin: getting facts fatal error: ", err)
//		return err
//	}

//	avaibleMetrics := make([]*plugin.MetricType, 0, len(*facterMap))
//	for key := range *facterMap {
//		avaibleMetrics = append(
//			avaibleMetrics,
//			plugin.NewMetricType(
//				[]string{Vendor, prefix, key},
//				timestamp.Unix()))
//	}

//	f.availableMetricTypes = &avaibleMetrics

//	return nil
//}

// responsible for updating stale metrics in cache
// names is slice with list of metrics to synchronize
// empty names means synchronize all facts
func (f *Facter) synchronizeCache(names []string) error {
	now := time.Now()

	// list of facts that have to be updated or acquired first time
	namesToUpdate := []string{}

	// collect stale or not existings facts
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

	// update outdated cache
	if len(namesToUpdate) > 0 {
		err := f.updateCache(namesToUpdate)
		if err != nil {
			return err
		}

	}
	return nil
}

// Updates all cache entries unconditionally (just a wrapper for updateCache)
func (f *Facter) updateCacheAll() error {
	return f.updateCache([]string{})
}

// Updates cache entries with current values
// pass empty names to update all entries
func (f *Facter) updateCache(names []string) error {

	// obtain actual facts
	facts, receviedAt, err := getFacts(names, f.facterExecutionDeadline)
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
			f.cache[name] = fact // golang defect #3117
		}
	}
	return nil
}

//cache    map[string]fact

func prepareMetricTypes(factMap *map[string]fact) ([]*plugin.MetricType, time.Time) {
	metricTypes := make([]*plugin.MetricType, 0, len(*factMap))
	var timestamp time.Time
	for factName, value := range *factMap {
		timestamp = value.lastUpdate
		metricTypes = append(metricTypes,
			plugin.NewMetricType([]string{Vendor, prefix, factName},
				timestamp.Unix()))
	}

	return metricTypes, timestamp
}

/******************************************
 *  Pulse plugin interface implmentation  *
 ******************************************/

// get available metrics types
func (f *Facter) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	if DefaultMetricTypesTTL > time.Since(f.metricTypesLastCheck) {
		err := f.synchronizeCache([]string{})
		if err != nil {
			log.Println("Facter: synchronizeCache returned error: " + err.Error())
			return err
		}
		var timestamp time.Time
		f.metricTypes, timestamp = prepareMetricTypes(&f.cache)
		f.metricTypesLastCheck = timestamp
	}
	reply.MetricTypes = f.metricTypes

	return nil
}

// collect metrics
func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	// TODO: INPUT: read CollectorArgs structure to extrac requested metrics
	requestedNames := []string{"kernel", "uptime"}

	// prepare cache to have all we need
	err := f.synchronizeCache(requestedNames)
	if err != nil {
		return err
	}

	// TODO: OUTPUT: fullfill reply structure with requested metrics
	// for _, name := range requestedNames {
	// 	// convert it some how if required
	// 	reply.metrics[name] = f.cache[name].value
	// }

	return nil
}

/****************************
 *  external facter cmd api *
 ****************************/
// helper type to deal with json that stores last update moment
// for a given fact
type fact struct {
	value      interface{}
	lastUpdate time.Time
}

// helper type to deal with json parsing
type stringmap map[string]interface{}

// get facts from facter (external command)
// returns all keys if none expliclty requested
func getFacts(keys []string, facterTimeout time.Duration) (*stringmap, *time.Time, error) {

	var timestamp time.Time

	// default options
	args := []string{"--json"}
	args = append(args, keys...)

	// execute command and capture the output
	jobCompletedChan := make(chan struct{})
	timeoutChan := time.After(facterTimeout)

	var err error
	output := make([]byte, 0, 1024)

	go func(jobCompletedChan chan<- struct{}, output *[]byte, err *error) {
		*output, *err = exec.Command("facter", args...).Output()
		jobCompletedChan <- struct{}{}
	}(jobCompletedChan, &output, &err)

	select {
	case <-timeoutChan:
		return nil, nil, errors.New("Facter plugin: fact gathering timeout")
	case <-jobCompletedChan:
		// success
		break
	}

	if err != nil {
		log.Println("exec returned " + err.Error())
		return nil, nil, err
	}
	timestamp = time.Now()

	var facterMap stringmap
	err = json.Unmarshal(output, &facterMap)
	if err != nil {
		log.Println("Unmarshal failed " + err.Error())
		return nil, nil, err
	}
	return &facterMap, &timestamp, nil
}
