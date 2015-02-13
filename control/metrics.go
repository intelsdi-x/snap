package control

import (
	"errors"
	"strings"
	"sync"
)

type metricType struct {
	Plugin *loadedPlugin

	namespace               []string
	lastAdvertisedTimestamp int64
}

func newMetricType(ns []string, last int64, plugin *loadedPlugin) *metricType {
	return &metricType{
		Plugin: plugin,

		namespace:               ns,
		lastAdvertisedTimestamp: last,
	}
}

func (m *metricType) Namespace() []string {
	return m.namespace
}

func (m *metricType) LastAdvertisedTimestamp() int64 {
	return m.lastAdvertisedTimestamp
}

type metricCatalog struct {
	table       *map[string]*[]*metricType
	mutex       *sync.Mutex
	keys        *[]string
	currentIter int
}

func newMetricCatalog() *metricCatalog {
	t := make(map[string]*[]*metricType)
	var k []string
	return &metricCatalog{
		table:       &t,
		mutex:       &sync.Mutex{},
		currentIter: 0,
		keys:        &k,
	}
}

// adds a metricType pointer to the loadedPlugins table
func (mc *metricCatalog) Add(m *metricType) {

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	key := getMetricKey(m.Namespace())

	if _, ok := (*mc.table)[key]; !ok {
		*mc.keys = append(*mc.keys, key)
		(*mc.table)[key] = &[]*metricType{m}
		return
	}

	*(*mc.table)[key] = append(*(*mc.table)[key], m)
}

// returns a copy of the table
func (mc *metricCatalog) Table() map[string]*metricType {
	var m = map[string]*metricType{}
	for k, v := range *mc.table {
		m[k] = getLatest(*v)
	}
	return m
}

// used to transactionally retrieve a loadedPlugin pointer from the table
func (mc *metricCatalog) Get(ns []string) (*metricType, error) {
	mc.Lock()
	defer mc.Unlock()

	key := getMetricKey(ns)
	var (
		ok bool
		m  *[]*metricType
		l  *metricType
	)
	if m, ok = (*mc.table)[key]; !ok {
		return nil, errors.New("metric not found")
	}

	if len(*m) > 1 {
		l = getLatest(*m)
	} else {
		l = (*m)[0]
	}

	return l, nil
}

func (mc *metricCatalog) Exists(mns []string) bool {
	_, ok := (*mc.table)[getMetricKey(mns)]
	return ok
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (mc *metricCatalog) Lock() {
	mc.mutex.Lock()
}

func (mc *metricCatalog) Unlock() {
	mc.mutex.Unlock()
}

func (mc *metricCatalog) Remove(ns []string) {
	mc.mutex.Lock()
	delete(*mc.table, getMetricKey(ns))
	mc.mutex.Unlock()
}

// returns the item of a certain index in the table.
// to be used when iterating over the table
func (mc *metricCatalog) Item() (string, *metricType) {
	key := (*mc.keys)[mc.currentIter-1]
	mts := (*mc.table)[key]
	m := getLatest(*mts)
	return key, m
}

// Returns true until the "end" of the table is reached.
// used to iterate over the table:
func (mc *metricCatalog) Next() bool {
	mc.currentIter++
	if mc.currentIter > len(*mc.table) {
		mc.currentIter = 0
		return false
	}
	return true
}

func getMetricKey(metric []string) string {
	return strings.Join(metric, ".")
}

func getLatest(c []*metricType) *metricType {
	unsorted := len(c)
	last := 0
	for sorted := 0; sorted < len(c)-1; sorted++ {
		last = 0
		lastIndex := 0
		for i, mt := range c[:unsorted] {
			if mt.Plugin.Version() > last {
				last = mt.Plugin.Version()
				lastIndex = i
			}
		}
		c[lastIndex], c[unsorted-1] = c[unsorted-1], c[lastIndex]
		unsorted--
	}
	return c[len(c)-1]
}
