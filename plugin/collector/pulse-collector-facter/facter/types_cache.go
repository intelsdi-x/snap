// cache of metric types - requires a entries passed from "metrics cache"
// protected by mutexes, so can be accessed from mulitple goroutines
package facter

import (
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

// typesCache is a cache that stores metrics types
type typesCache struct {
	data       []plugin.PluginMetricType
	mutex      sync.Mutex
	lastUpdate time.Time
	ttl        time.Duration
}

// newTypesCache is constuctor for typesCache
func newTypesCache(ttl time.Duration) *typesCache {
	return &typesCache{
		data:  []plugin.PluginMetricType{},
		ttl:   ttl,
		mutex: sync.Mutex{},
	}
}

// getMetricTypes returns a copy of cache, that can be returned to Pulse
func (t *typesCache) getMetricTypes() []plugin.PluginMetricType {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	mts := make([]plugin.PluginMetricType, len(t.data))
	copy(mts, t.data)
	return mts
}

// prepareMetricTypes fills metricTypes internal collection ready to send back to pulse
func (t *typesCache) cacheMetricTypes(entries map[string]entry) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// new temporary collection
	metricTypes := make([]plugin.PluginMetricType, 0, len(entries))

	// rewrite values from cache to another collection acceptable by Pulse
	for name, _ := range entries {
		metricType := plugin.NewPluginMetricType(namespace(name))
		metricTypes = append(metricTypes, *metricType)
	}

	// update internal state
	t.data = metricTypes

	// remember the last the metricTypes was filled
	// to be confronted with f.metricTypesTTL
	t.lastUpdate = time.Now()
}

// needUpdate returns true if types stored in cache are stale
// (older that ttl allows to be)
func (t *typesCache) needUpdate() bool {
	timeElapsed := time.Since(t.lastUpdate)
	return timeElapsed > t.ttl
}
