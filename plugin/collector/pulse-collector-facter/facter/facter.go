/*
# testing
go test github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/intelsdilabs/pulse/plugin/helper"
)

/*******************
 *  pulse plugin  *
 *******************/

var (
	namespace = []string{"intel", "facter"}
	Name      = GetPluginName(&namespace) //preprocessor needed
)

const (
	//	Name    = "facter" //should it be intel/facter ?
	Version                   = 1
	Type                      = plugin.CollectorPluginType
	DefaultCacheTTL           = 60 * time.Second
	DefaultAvailableMetricTTL = DefaultCacheTTL
)

// helper type to deal with json that stores last update moment
// for a given fact
type fact struct {
	value      interface{}
	lastUpdate time.Time
}

// wrapper to use facter binnary, responsible for running external binnary and caching
type Facter struct {
	availableMetricTypes     *[]*plugin.MetricType //map[string]interface{}
	availableMetricTimestamp time.Time
	availableMetricTTL       time.Time

	cacheTTL time.Duration
	cache    map[string]*fact
}

// fullfill the availableMetricTypes with data from facter
func (f *Facter) loadAvailableMetricTypes() error {

	// get all facts (empty slice)
	facterMap, timestamp, err := getFacts([]string{})
	if err != nil {
		log.Println("FacterPlugin: getting facts fatal error: ", err)
		return err
	}

	avaibleMetrics := make([]*plugin.MetricType, 0, len(*facterMap))
	for key := range *facterMap {
		avaibleMetrics = append(
			avaibleMetrics,
			plugin.NewMetricType(
				append(namespace, key),
				timestamp.Unix()))
	}

	f.availableMetricTypes = &avaibleMetrics

	return nil
}

// construct new Facter
func NewFacterPlugin() *Facter {
	f := new(Facter)
	//TODO read from config
	f.cacheTTL = DefaultCacheTTL
	f.cache = make(map[string]*fact)
	return f
}

// responsible for update cache with given list of metrics
func (f *Facter) updateCache(names []string) error {
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

	// are cached facts outdated ?
	if len(namesToUpdate) > 0 {
		// update cache

		// obtain actual facts
		facts, receviedAt, err := getFacts(namesToUpdate)
		if err != nil {
			return err
		}

		// merge cache with new facts
		for _, name := range namesToUpdate {
			// create fact if not exists yet
			value := (*facts)[name]
			if _, exists := f.cache[name]; !exists {
				f.cache[name] = &fact{
					lastUpdate: *receviedAt,
					value:      value,
				}
			} else {
				// just update the value in cache
				f.cache[name].value = value
			}

		}
	}
	return nil
}

/******************************************
 *  Pulse plugin interface implmentation  *
 ******************************************/

// get available metrics types
func (f *Facter) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	if time.Since(f.availableMetricTimestamp) > f.cacheTTL {

		f.loadAvailableMetricTypes()
		reply.MetricTypes = *f.availableMetricTypes

		return nil
	} else {
		reply.MetricTypes = *f.availableMetricTypes
		return nil
	}
}

// collect metrics
func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	// TODO: INPUT: read CollectorArgs structure to extrac requested metrics
	requestedNames := []string{"kernel", "uptime"}

	// prepare cache to have all we need
	err := f.updateCache(requestedNames)
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

/****************************
 *  external facter cmd api *
 ****************************/

// helper type to deal with json parsing
type stringmap map[string]interface{}

// get facts from facter (external command)
// returns all keys if none requested
func getFacts(keys []string) (*stringmap, *time.Time, error) {

	var timestamp time.Time

	// default options
	args := []string{"--json"}
	args = append(args, keys...)

	// execute command and capture the output
	out, err := exec.Command("facter", args...).Output()
	if err != nil {
		log.Println("exec returned " + err.Error())
		return nil, nil, err
	}
	timestamp = time.Now()

	var facterMap stringmap
	err = json.Unmarshal(out, &facterMap)
	if err != nil {
		log.Println("Unmarshal failed " + err.Error())
		return nil, nil, err
	}
	return &facterMap, &timestamp, nil
}
