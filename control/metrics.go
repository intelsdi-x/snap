package control

import "github.com/intelsdilabs/pulse/core"

type MetricType struct {
	Plugin *loadedPlugin

	namespace               []string
	lastAdvertisedTimestamp int64
}

func newMetricType(ns []string, last int64, plugin *loadedPlugin) *MetricType {
	return &MetricType{
		Plugin: plugin,

		namespace:               ns,
		lastAdvertisedTimestamp: last,
	}
}

func (m *MetricType) Namespace() []string {
	return m.namespace
}

func (m *MetricType) LastAdvertisedTimestamp() int64 {
	return m.lastAdvertisedTimestamp
}

type metricCatalog struct {
	table       *[]*MetricType
	mutex       *sync.Mutex
	currentIter int
}

func newMetricCatalog() *metricCatalog {
	var t []*MetricType
	return &metricCatalog{
		table:       &t,
		mutex:       &sync.Mutex{},
		currentIter: 0,
	}
}

// adds a loadedPlugin pointer to the loadedPlugins table
func (mc *MetricCatalog) Add(m *MetricType) error {

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// make sure we don't already  have a pointer to this plugin in the table
	for i, pl := range *mc.table {
		if lp == pl {
			return errors.New("plugin already loaded at index " + strconv.Itoa(i))
		}
	}

	// append
	newMetricCatalog := append(*mc.table, lp)
	// overwrite
	mc.table = &newMetricCatalog

	return nil
}

// returns a copy of the table
func (mc *MetricCatalog) Table() []*metricCatalog {
	return *mc.table
}

// used to transactionally retrieve a loadedPlugin pointer from the table
func (mc *MetricCatalog) Get(index int) (*MetricType, error) {
	mc.Lock()
	defer mc.Unlock()

	if index > len(*mc.table)-1 {
		return nil, errors.New("index out of range")
	}

	return (*mc.table)[index], nil
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (mc *MetricCatalog) Lock() {
	mc.mutex.Lock()
}

func (mc *MetricCatalog) Unlock() {
	mc.mutex.Unlock()
}

/* we need an atomic read / write transaction for the splice when removing a plugin,
   as the plugin is found by its index in the table.  By having the default Splice
   method block, we protect against accidental use.  Using nonblocking requires explicit
   invocation.
*/
func (mc *MetricCatalog) splice(index int) {
	m := append((*mc.table)[:index], (*mc.table)[index+1:]...)
	mc.table = &m
}

// splice unsafely
func (mc *MetricCatalog) NonblockingSplice(index int) {
	mc.splice(index)
}

// atomic splice
func (mc *MetricCatalog) Splice(index int) {

	mc.mutex.Lock()
	mc.splice(index)
	mc.mutex.Unlock()

}

// returns the item of a certain index in the table.
// to be used when iterating over the table
func (mc *MetricCatalog) Item() (int, *MetricType) {
	i := mc.currentIter - 1
	return i, (*mc.table)[i]
}

// Returns true until the "end" of the table is reached.
// used to iterate over the table:
func (mc *MetricCatalog) Next() bool {
	mc.currentIter++
	if mc.currentIter > len(*mc.table) {
		mc.currentIter = 0
		return false
	}
	return true
}
