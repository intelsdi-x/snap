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
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	hostnameReader      hostnamer
)

// hostnameReader, hostnamer created for mocking
func init() {
	host, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "addStandardAndWorkflowTags",
			"error":   err.Error(),
		}).Error("Unable to determine hostname")
		host = "not_found"
	}
	hostnameReader = &hostnameReaderType{hostname: host, hostnameRefreshTTL: time.Hour, lastRefresh: time.Now()}
}

type hostnamer interface {
	Hostname() (name string)
}

type hostnameReaderType struct {
	hostname           string
	hostnameRefreshTTL time.Duration
	lastRefresh        time.Time
}

func (h *hostnameReaderType) Hostname() (name string) {
	if time.Now().After(h.lastRefresh.Add(h.hostnameRefreshTTL)) {
		host, err := os.Hostname()
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control",
				"_file":   "metrics.go,",
				"_block":  "addStandardAndWorkflowTags",
				"error":   err.Error(),
			}).Error("Unable to determine hostname")
			host = "not_found"
		}

		h.hostname = host
		h.lastRefresh = time.Now()
	}
	return h.hostname
}

func errorMetricNotFound(ns string, ver ...int) error {
	if len(ver) > 0 {
		return fmt.Errorf("Metric not found: %s (version: %d)", ns, ver[0])
	}
	return fmt.Errorf("Metric not found: %s", ns)
}

func errorMetricsNotFound(ns string, ver ...int) error {
	if len(ver) > 0 {
		return fmt.Errorf("No metric found below the given namespace: %s (version: %d)", ns, ver[0])
	}
	return fmt.Errorf("No metric found below the given namespace: %s", ns)
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

func errorMetricElementHasTuple(value, ns string) error {
	return fmt.Errorf("A element %s should not define tuple for namespace %s.", value, ns)
}

func errorEmptyNamespace() error {
	return fmt.Errorf("Incorrect format of requested metric, empty list of namespace elements")
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
	Plugin             core.CatalogedPlugin
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

func (m *metric) Data() interface{} {
	return nil
}
func (m *metric) Description() string {
	return ""
}
func (m *metric) Unit() string {
	return ""
}
func (m *metric) Tags() map[string]string {
	return nil
}
func (m *metric) LastAdvertisedTime() time.Time {
	return time.Unix(0, 0)
}
func (m *metric) Timestamp() time.Time {
	return time.Unix(0, 0)
}

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

type catalogedPlugin struct {
	name         string
	version      int
	signed       bool
	typeName     plugin.PluginType
	state        pluginState
	path         string
	loadedTime   time.Time
	configPolicy *cpolicy.ConfigPolicy
}

func (cp *catalogedPlugin) TypeName() string {
	return cp.typeName.String()
}

func (cp *catalogedPlugin) Name() string {
	return cp.name
}

func (cp *catalogedPlugin) Version() int {
	return cp.version
}

func (cp *catalogedPlugin) Key() string {
	return fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", cp.TypeName(), cp.Name(), cp.Version())
}

func (cp *catalogedPlugin) IsSigned() bool {
	return cp.signed
}

func (cp *catalogedPlugin) Status() string {
	return string(cp.state)
}

func (cp *catalogedPlugin) PluginPath() string {
	return cp.path
}

func (cp *catalogedPlugin) LoadedTimestamp() *time.Time {
	return &cp.loadedTime
}

func (cp *catalogedPlugin) Policy() *cpolicy.ConfigPolicy {
	return cp.configPolicy
}

func newCatalogedPlugin(lp *loadedPlugin) core.CatalogedPlugin {
	cp := cpolicy.New()
	for _, keyNode := range lp.Policy().GetAll() {
		node := cpolicy.NewPolicyNode()
		rules, err := keyNode.ConfigPolicyNode.CopyRules()
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control",
				"_file":   "metrics.go,",
				"_block":  "newCatalogedPlugin",
				"error":   err.Error(),
			}).Error("Unable to copy rules")
			return nil
		}
		node.Add(rules...)
		cp.Add(keyNode.Key, node)
	}
	return &catalogedPlugin{
		name:         lp.Name(),
		version:      lp.Version(),
		signed:       lp.IsSigned(),
		typeName:     lp.Type,
		state:        lp.State,
		path:         lp.PluginPath(),
		loadedTime:   lp.LoadedTime,
		configPolicy: cp,
	}
}

type metricCatalog struct {
	tree  *MTTrie
	mutex *sync.Mutex
	keys  []string
}

func newMetricCatalog() *metricCatalog {
	return &metricCatalog{
		tree:  NewMTTrie(),
		mutex: &sync.Mutex{},
		keys:  []string{},
	}
}

func (mc *metricCatalog) Keys() []string {
	return mc.keys
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
		Plugin:             newCatalogedPlugin(lp),
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
		mc.keys = append(mc.keys, m.Namespace().String())
	}
}

// Add adds a metricType
func (mc *metricCatalog) Add(m *metricType) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := m.Namespace().String()

	// adding key as a cataloged keys (mc.keys)
	mc.keys = appendIfMissing(mc.keys, key)
	mc.tree.Add(m)
}

// GetMetric retrieves a metric for a given requested namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) GetMetric(requested core.Namespace, version int) (*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	var ns core.Namespace

	catalogedmt, err := mc.tree.GetMetric(requested.Strings(), version)
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-metric",
			"error":   err,
		}).Error("error getting metric")
		return nil, err
	}

	ns = catalogedmt.Namespace()

	if isDynamic, _ := ns.IsDynamic(); isDynamic {
		// when namespace is dynamic and the cataloged namespace (e.g. ns=/intel/mock/*/bar) is different than
		// the requested (e.g. requested=/intel/mock/host0/bar), than specify an instance of dynamic element,
		// so as a result the dynamic element will have set a value (e.g. ns[2].Value equals "host0")
		if ns.String() != requested.String() {
			ns = specifyInstanceOfDynamicMetric(ns, requested)
		}
	}

	returnedmt := &metricType{
		Plugin:             catalogedmt.Plugin,
		namespace:          ns,
		version:            catalogedmt.Version(),
		lastAdvertisedTime: catalogedmt.LastAdvertisedTime(),
		tags:               catalogedmt.Tags(),
		policy:             catalogedmt.Plugin.Policy().Get(catalogedmt.Namespace().Strings()),
		config:             catalogedmt.Config(),
		unit:               catalogedmt.Unit(),
		description:        catalogedmt.Description(),
		subscriptions:      catalogedmt.SubscriptionCount(),
	}
	return returnedmt, nil
}

// GetMetrics retrieves all metrics which fulfill a given requested namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) GetMetrics(requested core.Namespace, version int) ([]*metricType, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	returnedmts := []*metricType{}

	// resolve queried tuples in metric namespace
	requestedNss := findTuplesMatches(requested)
	for _, rns := range requestedNss {
		catalogedmts, err := mc.tree.GetMetrics(rns.Strings(), version)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control",
				"_file":   "metrics.go,",
				"_block":  "get-metrics",
				"error":   err,
			}).Error("error getting metric")
			return nil, err
		}
		for _, catalogedmt := range catalogedmts {
			ns := catalogedmt.Namespace()

			if isDynamic, _ := ns.IsDynamic(); isDynamic {
				// when namespace is dynamic and the cataloged namespace (e.g. ns=/intel/mock/*/bar) is different than
				// the requested (e.g. requested=/intel/mock/host0/bar), than specify an instance of dynamic element,
				// so as a result the dynamic element will have set a value (e.g. ns[2].Value equals "host0")
				if ns.String() != rns.String() {
					ns = specifyInstanceOfDynamicMetric(ns, rns)
				}
			}

			returnedmt := &metricType{
				Plugin:             catalogedmt.Plugin,
				namespace:          ns,
				version:            catalogedmt.Version(),
				lastAdvertisedTime: catalogedmt.LastAdvertisedTime(),
				tags:               catalogedmt.Tags(),
				policy:             catalogedmt.Plugin.Policy().Get(catalogedmt.Namespace().Strings()),
				config:             catalogedmt.Config(),
				unit:               catalogedmt.Unit(),
				description:        catalogedmt.Description(),
				subscriptions:      catalogedmt.SubscriptionCount(),
			}
			returnedmts = append(returnedmts, returnedmt)
		}
	}

	return returnedmts, nil
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

func (mc *metricCatalog) GetPlugin(mns core.Namespace, ver int) (core.CatalogedPlugin, error) {
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

func (mc *metricCatalog) GetPlugins(mns core.Namespace) ([]core.CatalogedPlugin, error) {
	plugins := []core.CatalogedPlugin{}
	pluginsMap := map[string]core.CatalogedPlugin{}

	mts, err := mc.tree.GetVersions(mns.Strings())
	if err != nil {
		log.WithFields(log.Fields{
			"_module": "control",
			"_file":   "metrics.go,",
			"_block":  "get-plugins",
			"error":   err,
		}).Error("error getting plugin")
		return nil, err
	}
	for _, mt := range mts {
		// iterate over metrics and add the plugin which exposes the following metric to a map
		// under plugin key to ensure that plugins do not repeat
		key := mt.Plugin.Key()
		pluginsMap[key] = mt.Plugin
	}
	for _, plg := range pluginsMap {
		plugins = append(plugins, plg)
	}

	return plugins, nil
}

func appendIfMissing(keys []string, ns string) []string {
	for _, key := range keys {
		if ns == key {
			// do not append if the key was found in keys
			return keys
		}
	}
	return append(keys, ns)
}

// isTuple returns true when incoming namespace's element has been recognized as a tuple, otherwise returns false
// notice, that the tuple is a string which starts with `core.TuplePrefix`, ends with `core.TupleSuffix`
// and contains at least one `core.TupleSeparator`, e.g. (host0;host1)
func isTuple(element string) bool {
	if strings.HasPrefix(element, core.TuplePrefix) && strings.HasSuffix(element, core.TupleSuffix) && strings.Contains(element, core.TupleSeparator) {
		return true
	}
	return false
}

// containsTuple checks if a given element of namespace has a tuple; if yes, returns true and recognized tuple's items
func containsTuple(nsElement string) (bool, []string) {
	tupleItems := []string{}
	if isTuple(nsElement) {
		if strings.ContainsAny(nsElement, "*") {
			// an asterisk covers all tuples cases (eg. /intel/mock/(host0;host1;*)/baz)
			// so to avoid retrieving the same metric more than once, return only '*' as a tuple's items
			tupleItems = []string{"*"}
		} else {
			tuple := strings.TrimSuffix(strings.TrimPrefix(nsElement, core.TuplePrefix), core.TupleSuffix)
			items := strings.Split(tuple, core.TupleSeparator)
			// removing all leading and trailing white space
			for _, item := range items {
				tupleItems = append(tupleItems, strings.TrimSpace(item))
			}
		}
		return true, tupleItems
	}
	return false, nil
}

// findTuplesMatches returns all matched combination of queried tuples in incoming namespace,
// where a tuple is in the form of `(host0;host1;host3)`. If the incoming namespace:
// - does not contain any tuple, return the incoming namespace as the only item in output slice.
// - contains a tuple, return the copies of incoming namespace with appropriate values set to namespaces' elements
func findTuplesMatches(incomingNs core.Namespace) []core.Namespace {
	// How it works, exemplary incoming namespace:
	// 	"intel", "mock", "(host0;host1)", "(baz;bar)"
	//
	// the following 4 namespaces will be returned:
	// 	"intel", "mock", "host0", "baz"
	// 	"intel", "mock", "host1", "baz"
	// 	"intel", "mock", "host0", "bar"
	// 	"intel", "mock", "host1", "bar"

	matchedItems := make(map[int][]string)
	numOfPossibleCombinations := 1

	for index, element := range incomingNs.Strings() {
		match := []string{}
		if ok, tupleItems := containsTuple(element); ok {
			match = tupleItems
		} else {
			match = []string{element}
		}
		// store matched items under current index of incoming namespace element
		matchedItems[index] = append(matchedItems[index], match...)

		// number of possible combinations increases N=len(match) times
		numOfPossibleCombinations = numOfPossibleCombinations * len(match)
	}

	//prepare slice for returned namespaces (results of tuple find)
	returnedNss := make([]core.Namespace, numOfPossibleCombinations)

	// initialize each of returned namespaces as a copy of incoming namespace
	// (copied original value, name and description of their elements)
	for i := 0; i < numOfPossibleCombinations; i++ {
		returnedNs := make([]core.NamespaceElement, len(incomingNs.Strings()))
		copy(returnedNs, incomingNs)
		returnedNss[i] = returnedNs
	}
	// set appropriate value to namespace's elements
	for index, items := range matchedItems {
		for i := range returnedNss {
			// retrieve the matched item (when 'i' exceeds the number of matched items, start from beginning)
			item := items[i%len(items)]
			returnedNss[i][index].Value = item
		}
	}
	return returnedNss
}

// specifyInstanceOfDynamicMetric returns specified namespace of incoming cataloged metric's namespace
// based on requested metric namespace
func specifyInstanceOfDynamicMetric(catalogedNamespace core.Namespace, requestedNamespace core.Namespace) core.Namespace {
	specifiedNamespace := make(core.Namespace, len(catalogedNamespace))
	copy(specifiedNamespace, catalogedNamespace)

	_, indexes := catalogedNamespace.IsDynamic()

	for _, index := range indexes {
		if len(requestedNamespace) > index {
			// use namespace's element of requested metric declared in task manifest
			// to specify a dynamic instance of the cataloged metric
			specifiedNamespace[index].Value = requestedNamespace[index].Value
		}
	}
	return specifiedNamespace
}

// validateMetricNamespace validates metric namespace in terms of containing properly defined dynamic elements,
// not ending with an asterisk and not contain elements which might be erroneously recognized as a tuple
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
		if isTuple(i.Value) {
			return errorMetricElementHasTuple(i.Value, ns.String())
		}
		value += i.Value
	}
	// plugin should NOT advertise metrics ending with a wildcard
	if strings.HasSuffix(value, "*") {
		return errorMetricEndsWithAsterisk(ns.String())
	}
	return nil
}
