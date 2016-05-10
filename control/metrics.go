/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
)

var (
	errMetricNotFound   = errors.New("metric not found")
	errNegativeSubCount = serror.New(errors.New("subscription count cannot be < 0"))
	notAllowedChars     = map[string][]string{
		"brackets":     {"(", ")", "[", "]", "{", "}"},
		"spaces":       {" "},
		"punctuations": {".", ",", ";", "?", "!"},
		"slashes":      {"|", "\\", "/"},
		"carets":       {"^"},
		"quotations":   {"\"", "`", "'"},
	}
)

func errorMetricNotFound(ns []string, ver ...int) error {
	if len(ver) > 0 {
		return fmt.Errorf("Metric not found: %s (version: %d)", core.NewNamespace(ns...).String(), ver[0])
	}
	return fmt.Errorf("Metric not found: %s", core.NewNamespace(ns...).String())
}

func errorFetchMetricsNotFound(ns []string) error {
	if len(ns) == 0 {
		// when fetching all cataloged metrics failed
		return fmt.Errorf("Metric catalog is empty (no plugin loaded)")
	}
	return fmt.Errorf("Metrics not found below a given namespace: %s", core.NewNamespace(ns...).String())
}

func errorMetricEndsWithAsterisk(ns []string) error {
	return fmt.Errorf("Metric namespace %s ends with an asterisk is not allowed", core.NewNamespace(ns...).String())
}

func errorMetricNamespaceHasNotAllowedChar(ns []string) error {
	// presents each elements of namespace separately (list of not allowed characters contains slashes)
	return fmt.Errorf("Metric namespace %s contains not allowed characters. Avoid using %s", ns, listNotAllowedChars())
}

// listNotAllowedChars returns list of not allowed characters in metric's namespace as a string
// which is used in construct errorMetricContainsNotAllowedChars as a recommendation
// exemplary output: "brackets [( ) [ ] { }], spaces [ ], punctuations [. , ; ? !], slashes [| \ /], carets [^], quotations [" ` ']"
func listNotAllowedChars() string {
	var result string
	for groupName, chars := range notAllowedChars {
		result += fmt.Sprintf(" %s %s,", groupName, chars)
	}
	// trim the comma in the end
	return strings.TrimSuffix(result, ",")
}

type metricCatalogItem struct {
	namespace string
	versions  map[int]core.Metric
}

func (m *metricCatalogItem) Namespace() string {
	return m.namespace
}

func (m *metricCatalogItem) Versions() map[int]core.Metric {
	return m.versions
}

type metricType struct {
	Plugin             *loadedPlugin
	namespace          core.Namespace
	version            int
	lastAdvertisedTime time.Time
	subscriptions      int
	policy             processesConfigData
	config             *cdata.ConfigDataNode
	data               interface{}
	tags               map[string]string
	timestamp          time.Time
	description        string
	unit               string
}

type processesConfigData interface {
	Process(map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *cpolicy.ProcessingErrors)
	HasRules() bool
}

func newMetricType(ns core.Namespace, last time.Time, plugin *loadedPlugin) *metricType {
	return &metricType{
		Plugin:             plugin,
		namespace:          ns,
		lastAdvertisedTime: last,
	}
}

func (m *metricType) Key() string {
	return fmt.Sprintf("%s/%d", m.Namespace().String(), m.Version())
}

func (m *metricType) Namespace() core.Namespace {
	return m.namespace
}

func (m *metricType) Data() interface{} {
	return m.data
}

func (m *metricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func (m *metricType) Subscribe() {
	m.subscriptions++
}

func (m *metricType) Unsubscribe() serror.SnapError {
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
	if m.version > 0 {
		return m.version
	}
	if m.Plugin == nil {
		return -1
	}
	return m.Plugin.Version()
}

func (m *metricType) Config() *cdata.ConfigDataNode {
	return m.config
}

func (m *metricType) Policy() *cpolicy.ConfigPolicyNode {
	return m.policy.(*cpolicy.ConfigPolicyNode)
}

func (m *metricType) Tags() map[string]string {
	return m.tags
}

func (m *metricType) Timestamp() time.Time {
	return m.timestamp
}

func (m *metricType) Description() string {
	return m.description
}

func (m *metricType) Unit() string {
	return m.unit
}

// mapKey distinguishes items in metricCatalog.mTree map
// based on metric key and version
type mapKey struct {
	mtKey     string
	mtVersion int
}

func newMapKey(metricKey string, version int) mapKey {
	return mapKey{metricKey, version}
}

func (mk *mapKey) metricNamespace() []string {
	return strings.Split(mk.mtKey, ".")
}

func (mk *mapKey) metricVersion() int {
	return mk.mtVersion
}

type metricCatalog struct {
	tree        *MTTrie
	mutex       *sync.Mutex
	currentIter int
	keys        []string

	// mTree holds requested metrics and maps them to the cataloged metrics types
	mTree map[mapKey][]*metricType
}

func newMetricCatalog() *metricCatalog {
	return &metricCatalog{
		tree:        NewMTTrie(),
		mutex:       &sync.Mutex{},
		currentIter: 0,
		keys:        []string{},
		mTree:       make(map[mapKey][]*metricType),
	}
}

func (mc *metricCatalog) Keys() []string {
	return mc.keys
}

// GetMatchedMetricTypes returns all stored matched metrics types for requested 'ns' where 'ns' might represent metric namespace(s) explicitly
// or via query by using an asterisk or a tuple
func (mc *metricCatalog) GetMatchedMetricTypes(ns core.Namespace, ver int) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mkey := newMapKey(ns.Key(), ver)

	if _, exist := mc.mTree[mkey]; !exist {
		//add item if not exist
		mc.addItemToMatchingMap(mkey)
	}

	return mc.getMatchedMetricTypes(mkey)
}

// getMatchedMetricTypes returns all matched metric types stored under the key 'mkey' in map 'mTree'
func (mc *metricCatalog) getMatchedMetricTypes(mkey mapKey) ([]*metricType, error) {
	mts := mc.mTree[mkey]
	if len(mts) == 0 {
		return nil, errorMetricNotFound(mkey.metricNamespace(), mkey.metricVersion())
	}

	return mts, nil
}

// findTupleSubmatch returns all matched combination of queried tuples
// where as a tuple there is a mean of `(a|b)`
func findTupleSubmatch(ns []string) [][]string {
	/*
		Example:
		findTupleSubmatch([]string{"intel", "mock", "(host0|host1)", "(baz|bar)"})

		would return four slices:
			[]string{"intel", "mock", "host0", "baz"}
			[]string{"intel", "mock", "host1", "baz"}
			[]string{"intel", "mock", "host0", "bar"}
			[]string{"intel", "mock", "host1", "bar"}
	*/

	numOfPossibleCombinations := 1
	matchedItems := make(map[int][]string)

	for index, n := range ns {
		match := []string{}

		if strings.ContainsAny(n, "*") {
			// an asterisk covers all tuples cases
			match = []string{"*"} // to avoid retrieving the same metric more than once
		} else {
			// a tuple is equivalent to regexp token (e.q match either a or b: '(a|b)')
			// so it provides regular expression by itself
			regex := regexp.MustCompile(n)
			match = regex.FindAllString(n, -1)
		}

		if match == nil {
			continue
		}

		matchedItems[index] = append(matchedItems[index], match...)

		// number of possible combinations increases N=len(match) times
		numOfPossibleCombinations = numOfPossibleCombinations * len(match)

	}

	//prepare two dimensional slice representing namespaces
	nss := make([][]string, numOfPossibleCombinations)

	for index := 0; index < len(ns); index++ {
		items := matchedItems[index]
		fillThreshold := len(items)

		for i := 0; i < numOfPossibleCombinations; i++ {
			// iterate over items and start from
			// the beginning when 'i' exceeds the threshold (the length of items)
			item := items[i%(fillThreshold)]
			nss[i] = append(nss[i], item)
		}
	}

	return nss
}

// specifyInstanceOfDynamicMetric returns specified namespace of incoming metric 'mt'
// based on requested metric namespace 'ns' and indexes pointing to dynamic elements of namespace
func specifyInstanceOfDynamicMetric(mt *metricType, ns []string, indexes []int) core.Namespace {
	specifiedNamespace := make(core.Namespace, len(mt.Namespace()))
	copy(specifiedNamespace, mt.Namespace())

	for _, index := range indexes {
		if len(ns) > index {
			// use namespace's element of requested metric declared in task manifest
			// to specify a dynamic instance of the cataloged metric
			specifiedNamespace[index].Value = ns[index]
		}
	}

	return specifiedNamespace
}

// addItemToMatchingMap adds `mkey` to matching map (or updates if `mkey` exists) with corresponding cataloged metrics as a content;
// if this 'mkey' does not match to any cataloged metrics, it will be removed from matching map
func (mc *metricCatalog) addItemToMatchingMap(mkey mapKey) {
	returnedmts := []*metricType{}

	if availablemts := mc.tree.gatherMetricTypes(); len(availablemts) == 0 {
		// no metric in the catalog
		mc.removeItemFromMatchingMap(mkey)
		return
	}

	// resolve queried tuples in metric namespace
	tnss := findTupleSubmatch(mkey.metricNamespace())

	for _, tns := range tnss {
		catalogedmts, err := mc.tree.GetMetrics(tns, mkey.metricVersion())
		if err != nil {
			// tuple e.q. `(a|b)` works like logic OR
			// return error only if neither 'a' nor 'b' cannot be found in the metric catalog
			// log error and check the next tuple
			log.WithFields(log.Fields{
				"_module": "control",
				"_file":   "metrics.go,",
				"_block":  "add-item-to-matching-map",
				"error":   err,
			}).Error("error getting metric")

			continue
		}

		for _, catalogedmt := range catalogedmts {
			var ns core.Namespace
			if ok, indexes := catalogedmt.Namespace().IsDynamic(); ok {
				// specify instance of dynamicMetric
				ns = specifyInstanceOfDynamicMetric(catalogedmt, tns, indexes)
			} else {
				ns = catalogedmt.Namespace()
			}

			returnedmt := &metricType{
				Plugin:             catalogedmt.Plugin,
				namespace:          ns,
				version:            catalogedmt.Version(),
				lastAdvertisedTime: catalogedmt.LastAdvertisedTime(),
				tags:               catalogedmt.Tags(),
				policy:             catalogedmt.Plugin.ConfigPolicy.Get(catalogedmt.Namespace().Strings()),
				config:             catalogedmt.Config(),
				unit:               catalogedmt.Unit(),
				description:        catalogedmt.Description(),
			}
			returnedmts = appendIfUnique(returnedmts, returnedmt)
		}
	}

	if len(returnedmts) == 0 {
		mc.removeItemFromMatchingMap(mkey)
		return
	}

	// add or update item(s) in map under key 'mkey'
	mc.mTree[mkey] = returnedmts
}

// removeItemFromMatchingMap removes items under the given key from matching map
func (mc *metricCatalog) removeItemFromMatchingMap(mkey mapKey) {
	if _, exist := mc.mTree[mkey]; exist {
		log.WithFields(log.Fields{
			"_module":   "core",
			"_file":     "metrics.go,",
			"_block":    "remove-item-from-matching-map",
			"_item-key": mkey,
		}).Debug("removing item from matching map under key")

		delete(mc.mTree, mkey)
	}
}

// updateMatchingMap updates the entire contents of matching map
func (mc *metricCatalog) updateMatchingMap() {
	for mkey := range mc.mTree {
		mc.addItemToMatchingMap(mkey)
	}
}

// validateMetricNamespace validates metric namespace in terms of containing not allowed characters and ending with an asterisk
func validateMetricNamespace(ns core.Namespace) error {
	name := ""
	for _, i := range ns {
		name += i.Value
	}
	for _, chars := range notAllowedChars {
		for _, ch := range chars {
			if strings.ContainsAny(name, ch) {
				return errorMetricNamespaceHasNotAllowedChar(ns.Strings())
			}
		}
	}
	// plugin should NOT advertise metrics ending with a wildcard
	if strings.HasSuffix(name, "*") {
		return errorMetricEndsWithAsterisk(ns.Strings())
	}

	return nil
}

func (mc *metricCatalog) AddLoadedMetricType(lp *loadedPlugin, mt core.Metric) error {
	if err := validateMetricNamespace(mt.Namespace()); err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "add-loaded-metric-type",
			"error":   fmt.Errorf("Metric namespace %s contains not allowed characters", mt.Namespace()),
		}).Error("error adding loaded metric type")
		return err
	}
	if lp.ConfigPolicy == nil {
		err := errors.New("Config policy is nil")
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "add-loaded-metric-type",
			"error":   err,
		}).Error("error adding loaded metric type")
		return err
	}
	newMt := metricType{
		Plugin:             lp,
		namespace:          mt.Namespace(),
		version:            mt.Version(),
		lastAdvertisedTime: mt.LastAdvertisedTime(),
		tags:               mt.Tags(),
		policy:             lp.ConfigPolicy.Get(mt.Namespace().Strings()),
		description:        mt.Description(),
		unit:               mt.Unit(),
	}
	mc.Add(&newMt)
	// the catalog has been changed, update content of matching map too
	mc.updateMatchingMap()
	return nil
}

// RmUnloadedPluginMetrics removes plugin metrics which was unloaded,
// consequently cataloged metrics are changed, so matching map is being updated too
func (mc *metricCatalog) RmUnloadedPluginMetrics(lp *loadedPlugin) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.tree.DeleteByPlugin(lp)

	// update content of matching map
	mc.updateMatchingMap()
}

// Add adds a metricType
func (mc *metricCatalog) Add(m *metricType) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := m.Namespace().Key()

	// adding key as a cataloged keys (mc.keys)
	mc.keys = appendIfMissing(mc.keys, key)

	mc.tree.Add(m)
}

// GetMetric retrieves a metric with given namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) GetMetric(ns core.Namespace, version int) (*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mt, err := mc.tree.GetMetric(ns.Strings(), version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-metric",
			"error":   err,
		}).Error("error getting metric")
		return nil, err
	}

	return mt, err
}

// GetMetrics retrieves all metrics which fulfill a given namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) GetMetrics(ns core.Namespace, version int) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mts, err := mc.tree.GetMetrics(ns.Strings(), version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-metrics",
			"error":   err,
		}).Error("error getting metrics")
		return nil, err
	}

	return mts, err
}

// GetVersions retrieves all versions of a given metric namespace.
func (mc *metricCatalog) GetVersions(ns core.Namespace) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mts, err := mc.tree.GetVersions(ns.Strings())
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-versions",
			"error":   err,
		}).Error("error getting plugin version")
		return nil, err
	}

	return mts, nil
}

// Fetch transactionally retrieves all metrics which fall under namespace ns
func (mc *metricCatalog) Fetch(ns core.Namespace) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mtsi, err := mc.tree.Fetch(ns.Strings())
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "fetch",
			"error":   err,
		}).Error("error fetching metrics")
		return nil, err
	}
	return mtsi, nil
}

// Remove removes a metricType from the catalog and from matching map
func (mc *metricCatalog) Remove(ns core.Namespace) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.tree.Remove(ns.Strings())
	mc.updateMatchingMap()
}

// Item returns the current metricType in the collection. The method Next()
// provides the  means to move the iterator forward.
func (mc *metricCatalog) Item() (string, []*metricType) {
	key := mc.keys[mc.currentIter-1]
	ns := strings.Split(key, ".")
	mtsi, _ := mc.tree.GetVersions(ns)
	var mts []*metricType
	for _, mt := range mtsi {
		mts = append(mts, mt)
	}
	return key, mts
}

// Next returns true until the "end" of the collection is reached.  When
// the end of the collection is reached the iterator is reset back to the
// head of the collection.
func (mc *metricCatalog) Next() bool {
	mc.currentIter++
	if mc.currentIter > len(mc.keys) {
		mc.currentIter = 0
		return false
	}
	return true
}

// Subscribe atomically increments a metric's subscription count in the table.
func (mc *metricCatalog) Subscribe(ns []string, version int) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	m, err := mc.tree.GetMetric(ns, version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "subscribe",
			"error":   err,
		}).Error("error getting metric")
		return err
	}

	m.Subscribe()
	return nil
}

// Unsubscribe atomically decrements a metric's count in the table
func (mc *metricCatalog) Unsubscribe(ns []string, version int) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	m, err := mc.tree.GetMetric(ns, version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "unsubscribe",
			"error":   err,
		}).Error("error getting metric")
		return err
	}

	return m.Unsubscribe()
}

func (mc *metricCatalog) GetPlugin(mns core.Namespace, ver int) (*loadedPlugin, error) {
	mt, err := mc.tree.GetMetric(mns.Strings(), ver)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-plugin",
			"error":   err,
		}).Error("error getting plugin")
		return nil, err
	}
	return mt.Plugin, nil
}

func appendIfMissing(keys []string, ns string) []string {
	for _, key := range keys {
		if ns == key {
			return keys
		}
	}
	return append(keys, ns)
}

func appendIfUnique(mts []*metricType, mt *metricType) []*metricType {
	unique := true
	for i := range mts {
		if reflect.DeepEqual(mts[i], mt) {
			// set unique to false and break this loop,
			// do not check the next one
			unique = false
			break
		}
	}
	if unique {
		// append if unique
		mts = append(mts, mt)
	}
	return mts
}

func addStandardTags(m core.Metric) core.Metric {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "addStandardTags",
			"error":   err.Error(),
		}).Error("Unable to determine hostname")
	}
	tags := m.Tags()
	if tags == nil {
		tags = map[string]string{}
	}
	tags[core.STD_TAG_PLUGIN_RUNNING_ON] = hostname
	metric := plugin.MetricType{
		Namespace_:          m.Namespace(),
		Version_:            m.Version(),
		LastAdvertisedTime_: m.LastAdvertisedTime(),
		Config_:             m.Config(),
		Data_:               m.Data(),
		Tags_:               tags,
		Description_:        m.Description(),
		Unit_:               m.Unit(),
		Timestamp_:          m.Timestamp(),
	}
	return metric
}
