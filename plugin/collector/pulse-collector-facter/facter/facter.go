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
	"fmt"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
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

/**********
 * Facter *
 **********/

// Facter implements API to communicate with Pulse
type Facter struct {
	// available metrics that can me returned
	metricTypes           []plugin.PluginMetricType
	metricTypesLastUpdate time.Time
	metricTypesTTL        time.Duration

	metricCache *metricCache
}

// NewFacter constructs new Facter with default values
func NewFacter() *Facter {
	return &Facter{
		metricTypesTTL: defaultMetricTypesTTL,
		metricCache:    newMetricCache(defaultCacheTTL, defaultFacterDeadline),
	}
}

// Pulse plugin interface implementation
// ----------------------------------------------------

// GetMetricTypes returns available metrics types
func (f *Facter) GetMetricTypes() ([]plugin.PluginMetricType, error) {

	// synchronize cache conditionally as a whole
	timeElapsed := time.Since(f.metricTypesLastUpdate)
	needUpdate := timeElapsed > f.metricTypesTTL
	if needUpdate {
		// synchronize cache conditionally for all fields
		err := f.metricCache.updateCacheAll() // fills internal f.fcache
		if err != nil {
			log.Println("Facter: synchronizeCache returned error: " + err.Error())
			return nil, err
		}

		// fill f.metricTypes based on f.cache
		f.prepareMetricTypes()

	}
	// return metricTypes prepared earlier
	metricTypes := []plugin.PluginMetricType{}
	for _, metricType := range f.metricTypes {
		metricTypes = append(metricTypes, metricType)
	}
	return metricTypes, nil
}

// Collect collects metrics from external binary a returns them in form
// acceptable by Pulse
func (f *Facter) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {

	// parse input

	// requested names
	names := []string{}
	for _, metricType := range metricTypes {
		namespace := metricType.Namespace()
		// check namespace intel(vendor)/facter(prefix)/FACTNAME
		if len(namespace) != 3 {
			return nil, errors.New(fmt.Sprintf("unknown metricType %s (should containt just 3 segments)", namespace))
		}
		if namespace[0] != vendor {
			return nil, errors.New(fmt.Sprintf("unknown metricType %s (expected vendor %s)", namespace, vendor))
		}

		if namespace[1] != prefix {
			return nil, errors.New(fmt.Sprintf("unknown metricType %s (expected prefix %s)", namespace, prefix))
		}

		// name of fact - last part of namespace
		name := namespace[2]
		names = append(names, name)
	}

	// synchronize cache (stale of missing data) to have all we need
	err := f.metricCache.synchronizeCache(names)
	if err != nil {
		return nil, err
	}

	// read data from cache and preapre PluginMetric slice
	ms := []plugin.PluginMetric{}
	for _, name := range names {
		entry := f.metricCache.getEntry(name)
		metric := plugin.NewPluginMetric(namespace(name), entry.value)
		ms = append(ms, *metric)
	}

	return ms, nil
}

// helper functions to support CollectMetrics & GetMetricTypes

// prepareMetricTypes fills metricTypes internal collection ready to send back to pulse
func (f *Facter) prepareMetricTypes() {

	// new temporary collection
	metricTypes := make([]plugin.PluginMetricType, 0, f.metricCache.size())

	// rewrite values from cache to another collection acceptable by Pulse
	entries := f.metricCache.entries()
	for name, _ := range entries {
		metricType := plugin.NewPluginMetricType(namespace(name))
		metricTypes = append(metricTypes, *metricType)
	}

	// update internal state
	f.metricTypes = metricTypes

	// remember the last the metricTypes was filled
	// to be confronted with f.metricTypesTTL
	f.metricTypesLastUpdate = time.Now()
}

// required by PulseAPI
func ConfigPolicyTree() *cpolicy.ConfigPolicyTree {
	c := cpolicy.NewTree()
	return c
}

// namspace returns namespace slice of strings
// composed from: vendor, prefix and fact name
func namespace(name string) []string {
	return []string{vendor, prefix, name}

}
