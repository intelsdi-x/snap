package facter

import (
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

type typesCache struct {

	// TODO; extrac metricTypes functionallity to anotther struct
	// typesCache:
	data       []plugin.PluginMetricType
	mutex      sync.Mutex
	lastUpdate time.Time
	ttl        time.Duration
}

func newTypesCache(ttl time.Duration) *typesCache {
	return &typesCache{
		data:  []plugin.PluginMetricType{},
		ttl:   ttl,
		mutex: sync.Mutex{},
	}
}

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
	return
}

func (t *typesCache) needUpdate() bool {
	timeElapsed := time.Since(t.lastUpdate)
	return timeElapsed > t.ttl
}
