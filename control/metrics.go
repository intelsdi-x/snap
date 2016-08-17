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

	hostnameReader hostnamer
)

// hostnameReader, hostnamer created for mocking
func init() {
	hostnameReader = &hostnameReaderType{}
}

type hostnamer interface {
	Hostname() (name string, err error)
}

type hostnameReaderType struct{}

func (h *hostnameReaderType) Hostname() (name string, err error) {
	return os.Hostname()
}

func errorMetricNotFound(ns string, ver ...int) error {
	if len(ver) > 0 {
		return fmt.Errorf("Metric not found: %s (version: %d)", ns, ver[0])
	}
	return fmt.Errorf("Metric not found: %s", ns)
}

func errorMetricContainsNotAllowedChars(ns string) error {
	return fmt.Errorf("Metric namespace %s contains not allowed characters. Avoid using %s", ns, listNotAllowedChars())
}

func errorMetricEndsWithAsterisk(ns string) error {
	return fmt.Errorf("Metric namespace %s ends with an asterisk is not allowed", ns)
}

func errorMetricStaticElementHasName(value, name, ns string) error {
	return fmt.Errorf("A static element %s should not define name %s for namespace %s.", value, name, ns)
}

func errorMetricDynamicElementHasNoName(value, ns string) error {
	return fmt.Errorf("A dynamic element %s requires a name for namespace %s.", value, ns)
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

type metric struct {
	namespace core.Namespace
	version   int
	config    *cdata.ConfigDataNode
}

func (m *metric) Namespace() core.Namespace {
	return m.namespace
}

func (m *metric) Config() *cdata.ConfigDataNode {
	return m.config
}

func (m *metric) Version() int {
	return m.version
}

func (m *metric) Data() interface{}             { return nil }
func (m *metric) Description() string           { return "" }
func (m *metric) Unit() string                  { return "" }
func (m *metric) Tags() map[string]string       { return nil }
func (m *metric) LastAdvertisedTime() time.Time { return time.Unix(0, 0) }
func (m *metric) Timestamp() time.Time          { return time.Unix(0, 0) }

type processesConfigData interface {
	Process(map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *cpolicy.ProcessingErrors)
	HasRules() bool
}

func newMetricType(ns core.Namespace, last time.Time, plugin *loadedPlugin) *metricType {
	return &metricType{
		Plugin: plugin,

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

type metricCatalog struct {
	tree  *MTTrie
	mutex *sync.Mutex
	keys  []string

	// mKeys holds requested metric's keys which can include wildcards and matched to them the cataloged keys
	mKeys       map[string][]string
	currentIter int
}

func newMetricCatalog() *metricCatalog {
	return &metricCatalog{
		tree:        NewMTTrie(),
		mutex:       &sync.Mutex{},
		currentIter: 0,
		keys:        []string{},
		mKeys:       make(map[string][]string),
	}
}

func (mc *metricCatalog) Keys() []string {
	return mc.keys
}

// matchedNamespaces retrieves all matched items stored in mKey map under the key 'wkey' and converts them to namespaces
func (mc *metricCatalog) matchedNamespaces(wkey string) ([]core.Namespace, error) {
	// mkeys means matched metrics keys
	mkeys := mc.mKeys[wkey]

	if len(mkeys) == 0 {
		return nil, errorMetricNotFound(getMetricNamespace(wkey).String())
	}

	// convert matched keys to a slice of namespaces
	return convertKeysToNamespaces(mkeys), nil
}

// GetQueriedNamespaces returns all matched metrics namespaces for query 'ns' which can contain
// an asterisk or tuple (refer to query support)
func (mc *metricCatalog) GetQueriedNamespaces(ns core.Namespace) ([]core.Namespace, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// get metric key (might contain wildcard(s))
	wkey := ns.Key()

	return mc.matchedNamespaces(wkey)
}

// UpdateQueriedNamespaces matches given 'ns' which could contain an asterisk or a tuple and add them to matching map under key 'ns'
func (mc *metricCatalog) UpdateQueriedNamespaces(ns core.Namespace) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// get metric key (might contain wildcard(s))
	wkey := ns.Key()

	// adding matched namespaces to map
	mc.addItemToMatchingMap(wkey)
}

func convertKeysToNamespaces(keys []string) []core.Namespace {
	// nss is a slice of slices which holds metrics namespaces
	nss := []core.Namespace{}
	for _, key := range keys {
		ns := getMetricNamespace(key)
		if len(ns) != 0 {
			nss = append(nss, ns)
		}
	}
	return nss
}

// addItemToMatchingMap adds `wkey` to matching map (or updates if `wkey` exists) with corresponding cataloged keys as a content;
// if this 'wkey' does not match to any cataloged keys, it will be removed from matching map
func (mc *metricCatalog) addItemToMatchingMap(wkey string) {
	matchedKeys := mc.matchKeys(wkey)
	if len(matchedKeys) == 0 {
		mc.removeItemFromMatchingMap(wkey)
	} else {
		mc.mKeys[wkey] = matchedKeys
	}
}

// matchKeys returns all keys matching with provided key
func (mc *metricCatalog) matchKeys(wkey string) []string {
	matchedKeys := []string{}

	// wkey contains `.` which should not be interpreted as regexp tokens, but as a single character
	exp := strings.Replace(wkey, ".", "[.]", -1)

	// change `*` into regexp `.*` which matches any characters
	exp = strings.Replace(exp, "*", ".*", -1)

	regex := regexp.MustCompile("^" + exp + "$")
	for _, key := range mc.keys {
		match := regex.FindStringSubmatch(key)
		if match == nil {
			continue
		}
		matchedKeys = appendIfMissing(matchedKeys, key)
	}
	return matchedKeys
}

// removeItemFromMatchingMap removes `wkey` from matching map
func (mc *metricCatalog) removeItemFromMatchingMap(wkey string) {
	if _, exist := mc.mKeys[wkey]; exist {
		delete(mc.mKeys, wkey)
	}
}

// updateMatchingMap updates the contents of matching map
func (mc *metricCatalog) updateMatchingMap() {
	for wkey := range mc.mKeys {
		// add (or update if exist) item `wkey'
		mc.addItemToMatchingMap(wkey)
	}
}

// removeMatchedKey iterates over all items in the mKey and removes `key` from its content
func (mc *metricCatalog) removeMatchedKey(key string) {
	for wkey, mkeys := range mc.mKeys {
		for index, mkey := range mkeys {
			if mkey == key {
				// remove this key from slice
				mc.mKeys[wkey] = append(mkeys[:index], mkeys[index+1:]...)
			}
		}
		// if no matched key left, remove this item from map
		if len(mc.mKeys[wkey]) == 0 {
			mc.removeItemFromMatchingMap(wkey)
		}
	}
}

// validateMetricNamespace validates metric namespace in terms of containing not allowed characters and ending with an asterisk
func validateMetricNamespace(ns core.Namespace) error {
	value := ""
	for _, i := range ns {
		// A dynamic element requires the name while a static element does not.
		if i.Name != "" && i.Value != "*" {
			return errorMetricStaticElementHasName(i.Value, i.Name, ns.String())
		}
		if i.Name == "" && i.Value == "*" {
			return errorMetricDynamicElementHasNoName(i.Value, ns.String())
		}

		value += i.Value
	}

	for _, chars := range notAllowedChars {
		for _, ch := range chars {
			if strings.ContainsAny(value, ch) {
				return errorMetricContainsNotAllowedChars(ns.String())
			}
		}
	}
	// plugin should NOT advertise metrics ending with a wildcard
	if strings.HasSuffix(value, "*") {
		return errorMetricEndsWithAsterisk(ns.String())
	}

	return nil
}

func (mc *metricCatalog) AddLoadedMetricType(lp *loadedPlugin, mt core.Metric) error {
	if err := validateMetricNamespace(mt.Namespace()); err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "add-loaded-metric-type",
			"error":   fmt.Errorf("Metric namespace %s is invalid", mt.Namespace()),
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
	return nil
}

// RmUnloadedPluginMetrics removes plugin metrics which was unloaded,
// consequently cataloged metrics are changed, so matching map is being updated too
func (mc *metricCatalog) RmUnloadedPluginMetrics(lp *loadedPlugin) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.tree.DeleteByPlugin(lp)

	// Update metric catalog keys
	mc.keys = []string{}
	mts := mc.tree.gatherMetricTypes()
	for _, m := range mts {
		mc.keys = append(mc.keys, m.Namespace().Key())
	}

	// update the contents of matching map (mKeys)
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

// Get retrieves a metric given a namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) Get(ns core.Namespace, version int) (*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	return mc.get(ns.Strings(), version)
}

// GetVersions retrieves all versions of a given metric namespace.
func (mc *metricCatalog) GetVersions(ns core.Namespace) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	return mc.getVersions(ns.Strings())
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

	// remove all items from map mKey mapped for this 'ns'
	key := ns.Key()
	mc.removeMatchedKey(key)
}

// Item returns the current metricType in the collection.  The method Next()
// provides the  means to move the iterator forward.
func (mc *metricCatalog) Item() (string, []*metricType) {
	key := mc.keys[mc.currentIter-1]
	ns := strings.Split(key, ".")
	mtsi, _ := mc.tree.Get(ns)
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

	m, err := mc.get(ns, version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "subscribe",
			"error":   err,
		}).Error("error getting metrics")
		return err
	}

	m.Subscribe()
	return nil
}

// Unsubscribe atomically decrements a metric's count in the table
func (mc *metricCatalog) Unsubscribe(ns []string, version int) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	m, err := mc.get(ns, version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "unsubscribe",
			"error":   err,
		}).Error("error getting metrics")
		return err
	}

	return m.Unsubscribe()
}

func (mc *metricCatalog) GetPlugin(mns core.Namespace, ver int) (*loadedPlugin, error) {
	m, err := mc.Get(mns, ver)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-plugin",
			"error":   err,
		}).Error("error getting plugin")
		return nil, err
	}
	return m.Plugin, nil
}

func (mc *metricCatalog) get(ns []string, ver int) (*metricType, error) {
	mts, err := mc.getVersions(ns)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get",
			"error":   err,
		}).Error("error getting plugin version from metric catalog")
		return nil, err
	}
	// a version IS given
	if ver > 0 {
		l, err := getVersion(mts, ver)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control",
				"_file":   "metrics.go,",
				"_block":  "get",
				"error":   err,
			}).Error("error getting plugin version")
			return nil, errorMetricNotFound("/"+strings.Join(ns, "/"), ver)
		}
		return l, nil
	}
	// ver is less than or equal to 0 get the latest
	return getLatest(mts), nil
}

func (mc *metricCatalog) getVersions(ns []string) ([]*metricType, error) {
	mts, err := mc.tree.Get(ns)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "getVersions",
			"error":   err,
		}).Error("error getting plugin version")
		return nil, err
	}
	if len(mts) == 0 {
		return nil, errorMetricNotFound("/" + strings.Join(ns, "/"))
	}
	return mts, nil
}

func getMetricNamespace(key string) core.Namespace {
	return core.NewNamespace(strings.Split(key, ".")...)
}

func getLatest(c []*metricType) *metricType {
	cur := c[0]
	for _, mt := range c {
		if mt.Version() > cur.Version() {
			cur = mt
		}
	}
	return cur
}

func appendIfMissing(keys []string, ns string) []string {
	for _, key := range keys {
		if ns == key {
			return keys
		}
	}
	return append(keys, ns)
}

func getVersion(c []*metricType, ver int) (*metricType, error) {
	for _, m := range c {
		if m.Plugin.Version() == ver {
			return m, nil
		}
	}
	return nil, errMetricNotFound
}

func addStandardAndWorkflowTags(m core.Metric, allTags map[string]map[string]string) core.Metric {
	hostname, err := hostnameReader.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "addStandardAndWorkflowTags",
			"error":   err.Error(),
		}).Error("Unable to determine hostname")
	}
	tags := m.Tags()
	if tags == nil {
		tags = map[string]string{}
	}
	// apply tags from workflow
	for ns, nsTags := range allTags {
		if strings.HasPrefix(m.Namespace().String(), ns) {
			for k, v := range nsTags {
				tags[k] = v
			}
		}
	}
	// apply standard tag
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
