/* This modules converts implements Pulse API to become an plugin.
Additionally it caches all data to protect the system against overuse.

Implementation details:
legend:
- metric - represents value of metric from Pulse side
- fact - represents a value about a system gathered from Facter
- name - is string identifier that refers to metric from the Pulse side, so name points to metric


    GetMetricTypes  +------------+
            +-------> typesCache |
            |       +-----+------+
Pulse   +---+----+        |
  +-----> Facter |        | getEntries
        +---+----+        |
            |       +-----v--------+   +----------------+
            +-------> metricsCache +--->   getFacts     |
    CollectMetric   +--------------+   +-------+--------+
                                               | run in goroutine
                                       +-------v--------+
                                       |   ./facter     |
                                       +----------------+
*/

package facter

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	// parts of returned namescape
	vendor = "intel"
	prefix = "facter"
	// how long we are caching the date from external binary to prevent overuse of resources
	defaultCacheTTL = 60 * time.Second
	// how long are we going to cache available types of metrics
	defaultMetricTypesTTL = defaultCacheTTL
	// deadline a.k.a. timeout we are ready to wait for external binary to gather the data
	defaultFacterDeadline = 5 * time.Second
)

/**********
 * Facter *
 **********/

// Facter implements API to communicate with Pulse
type Facter struct {
	typesCache  *typesCache
	metricCache *metricCache
}

// make sure that we actually satisify requierd interface
var _ plugin.CollectorPlugin = (*Facter)(nil)

// NewFacter constructs new Facter with default values
func NewFacter() *Facter {
	return &Facter{
		typesCache:  newTypesCache(defaultMetricTypesTTL),
		metricCache: newMetricCache(defaultCacheTTL, defaultFacterDeadline),
	}
}

// ------------ Pulse plugin interface implementation --------------

// GetMetricTypes returns available metrics types
// idea: if types cache is stale then update metrics cache and based on this fill cache for types
// and return types from cache
func (f *Facter) GetMetricTypes() ([]plugin.PluginMetricType, error) {

	// synchronize metrics cache conditionally if metrics cache is stale
	if f.typesCache.needUpdate() {
		// synchronize cache conditionally for all fields
		err := f.metricCache.updateCacheAll() // fills internal f.fcache
		if err != nil {
			log.Println("Facter: synchronizeCache returned error: " + err.Error())
			return nil, err
		}
		// fill typesCache.metricTypes cache
		f.typesCache.cacheMetricTypes(f.metricCache.entries())

	}
	return f.typesCache.getMetricTypes(), nil
}

// Collect collects metrics from external binary a returns them in form
// acceptable by Pulse
func (f *Facter) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {

	// requested names
	names := []string{}
	for _, metricType := range metricTypes {
		namespace := metricType.Namespace()

		err := validateNamespace(namespace)
		if err != nil {
			return nil, err
		}

		// name of fact - last part of namespace
		name := namespace[2]
		names = append(names, name)
	}

	if len(names) == 0 {
		// nothing request, none returned
		return []plugin.PluginMetric(nil), nil
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

// validateNamespace checks namespace intel(vendor)/facter(prefix)/FACTNAME
func validateNamespace(namespace []string) error {
	if len(namespace) != 3 {
		return errors.New(fmt.Sprintf("unknown metricType %s (should containt just 3 segments)", namespace))
	}
	if namespace[0] != vendor {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected vendor %s)", namespace, vendor))
	}

	if namespace[1] != prefix {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected prefix %s)", namespace, prefix))
	}
	return nil
}

// ------------ helper functions --------------

// namspace returns namespace slice of strings
// composed from: vendor, prefix and fact name
func namespace(name string) []string {
	return []string{vendor, prefix, name}

}
