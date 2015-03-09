package control

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

var (
	errMetricNotFound   = errors.New("metric not found")
	errNegativeSubCount = errors.New("subscription count cannot be < 0")
)

type metricType struct {
	Plugin *loadedPlugin

	namespace          []string
	lastAdvertisedTime time.Time
	subscriptions      int
}

type processesConfigData interface {
	Process(map[string]ctypes.ConfigValue) *map[string]ctypes.ConfigValue
}

func newMetricType(ns []string, last time.Time, plugin *loadedPlugin) *metricType {
	return &metricType{
		Plugin: plugin,

		namespace:          ns,
		lastAdvertisedTime: last,
	}
}

func (m *metricType) Namespace() []string {
	return m.namespace
}

func (m *metricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func (m *metricType) Subscribe() {
	m.subscriptions++
}

func (m *metricType) Unsubscribe() error {
	if m.subscriptions == 0 {
		return errNegativeSubCount
	}
	m.subscriptions--
	return nil
}

func (m *metricType) SubscriptionCount() int {
	return m.subscriptions
}

func (m *metricType) Version() int {
	return m.Plugin.Version()
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

func (m *metricCatalog) AddLoadedMetricType(lp *loadedPlugin, mt core.MetricType) {
	newMt := metricType{
		Plugin:             lp,
		namespace:          mt.Namespace(),
		lastAdvertisedTime: mt.LastAdvertisedTime(),
	}
	m.Add(&newMt)
}

// adds a metricType pointer to the loadedPlugins table
func (mc *metricCatalog) Add(m *metricType) {

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := getMetricKey(m.Namespace())

	if _, ok := (*mc.table)[key]; !ok {
		*mc.keys = append(*mc.keys, key)
		(*mc.table)[key] = &[]*metricType{m}
	} else {
		*(*mc.table)[key] = append(*(*mc.table)[key], m)
	}

	sort((*mc.table)[key])
}

// returns a copy of the table
func (mc *metricCatalog) Table() map[string][]*metricType {
	var m = map[string][]*metricType{}
	for k, v := range *mc.table {
		m[k] = *v
	}
	return m
}

// used to transactionally retrieve a loadedPlugin pointer from the table
func (mc *metricCatalog) Get(ns []string, version int) (*metricType, error) {
	mc.Lock()
	defer mc.Unlock()
	key := getMetricKey(ns)
	return mc.get(key, version)
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
func (mc *metricCatalog) Item() (string, []*metricType) {
	key := (*mc.keys)[mc.currentIter-1]
	return key, *(*mc.table)[key]
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

// Subscribe atomically increments a metric's subscription count in the table.
func (mc *metricCatalog) Subscribe(ns []string, version int) error {
	key := getMetricKey(ns)

	mc.Lock()
	defer mc.Unlock()

	m, err := mc.get(key, version)
	if err != nil {
		return err
	}

	m.Subscribe()
	return nil
}

// Unsubscribe atomically decrements a metric's count in the table
func (mc *metricCatalog) Unsubscribe(ns []string, version int) error {
	key := getMetricKey(ns)

	mc.Lock()
	defer mc.Unlock()

	m, err := mc.get(key, version)
	if err != nil {
		return err
	}

	return m.Unsubscribe()
}

func (mc *metricCatalog) resolvePlugin(mns []string, ver int) (*loadedPlugin, error) {
	m, err := mc.Get(mns, ver)
	if err != nil {
		return nil, err
	}
	return m.Plugin, nil
}

func (mc *metricCatalog) get(key string, ver int) (*metricType, error) {
	var (
		ok bool
		m  *[]*metricType
	)
	if m, ok = (*mc.table)[key]; !ok {
		return nil, errMetricNotFound
	}

	// handle cases where there are multiple versions of a metric type
	if len(*m) > 1 {

		// a version IS given
		if ver >= 0 {
			l, err := getVersion(m, ver)
			if err != nil {
				return nil, err
			}
			return l, nil
		}

		// multiple versions, but -1 was given for the version
		// meaning, just get the lastest
		return getLatest(m), nil
	}

	// only one metric type to return, so get it
	return (*m)[0], nil
}

func getMetricKey(metric []string) string {
	return strings.Join(metric, ".")
}

func sort(c *[]*metricType) {
	unsorted := len(*c)
	last := 0
	for sorted := 0; sorted < len(*c)-1; sorted++ {
		last = 0
		lastIndex := 0
		for i, mt := range (*c)[:unsorted] {
			if mt.Plugin.Version() > last {
				last = mt.Plugin.Version()
				lastIndex = i
			}
		}
		(*c)[lastIndex], (*c)[unsorted-1] = (*c)[unsorted-1], (*c)[lastIndex]
		unsorted--
	}
}

func getLatest(c *[]*metricType) *metricType {
	return (*c)[len(*c)-1]
}

func getVersion(c *[]*metricType, ver int) (*metricType, error) {
	for _, m := range *c {
		if m.Plugin.Version() == ver {
			return m, nil
		}
	}
	return nil, errMetricNotFound
}
