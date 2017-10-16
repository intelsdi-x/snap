/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/serror"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrSubscriptionGroupAlreadyExists - error message when the subscription
	// group already exists
	ErrSubscriptionGroupAlreadyExists = core.ErrSubscriptionGroupAlreadyExists

	// ErrSubscriptionGroupDoesNotExist - error message when the subscription
	// group does not exist
	ErrSubscriptionGroupDoesNotExist = core.ErrSubscriptionGroupDoesNotExist

	ErrConfigRequiredForMetric = errors.New("config required")
)

// ManagesSubscriptionGroups is the interface implemented by an object that can
// manage subscription groups.
type ManagesSubscriptionGroups interface {
	Process() (errs []serror.SnapError)
	Add(id string, requested []core.RequestedMetric,
		configTree *cdata.ConfigDataTree,
		plugins []core.SubscribedPlugin) []serror.SnapError
	Get(id string) (map[string]metricTypes, []serror.SnapError, error)
	Remove(id string) []serror.SnapError
	ValidateDeps(requested []core.RequestedMetric,
		plugins []core.SubscribedPlugin,
		configTree *cdata.ConfigDataTree, asserts ...core.SubscribedPluginAssert) (serrs []serror.SnapError)
	validateMetric(metric core.Metric) (serrs []serror.SnapError)
	validatePluginUnloading(*loadedPlugin) (errs []serror.SnapError)
}

type subscriptionGroup struct {
	*pluginControl
	// requested metrics - never updated
	requestedMetrics []core.RequestedMetric
	// requested plugins - contains only processors and publishers;
	// never updated
	requestedPlugins []core.SubscribedPlugin
	// config from request - never updated
	configTree *cdata.ConfigDataTree
	// resulting metrics - updated after plugin load/unload events; they are grouped by plugin
	metrics map[string]metricTypes
	// resulting plugins - updated after plugin load/unload events
	plugins []core.SubscribedPlugin
	// errors generated the last time the subscription was processed
	// subscription groups are processed when the subscription group is added
	// and when plugins are loaded/unloaded
	errors []serror.SnapError
}

type subscriptionMap map[string]*subscriptionGroup

type subscriptionGroups struct {
	subscriptionMap
	*sync.Mutex
	*pluginControl
}

func newSubscriptionGroups(control *pluginControl) *subscriptionGroups {
	return &subscriptionGroups{
		make(map[string]*subscriptionGroup),
		&sync.Mutex{},
		control,
	}
}

// Add adds a subscription group provided a subscription group id, requested
// metrics, config tree and plugins. The requested metrics are mapped to
// collector plugins which are then combined with the provided (processor and
// publisher) plugins.  The provided config map is used to construct the
// []core.Metric which will be used during collect calls made against the
// subscription group.
// Returns an array of errors ([]serror.SnapError).
// `ErrSubscriptionGroupAlreadyExists` is returned if the subscription already
// exists.  Also, if there are errors mapping the requested metrics to plugins
// those are returned.
func (s subscriptionGroups) Add(id string, requested []core.RequestedMetric,
	configTree *cdata.ConfigDataTree,
	plugins []core.SubscribedPlugin) []serror.SnapError {
	s.Lock()
	defer s.Unlock()
	errs := s.add(id, requested, configTree, plugins)
	return errs
}

func (s subscriptionGroups) add(id string, requested []core.RequestedMetric,
	configTree *cdata.ConfigDataTree,
	plugins []core.SubscribedPlugin) []serror.SnapError {
	if _, ok := s.subscriptionMap[id]; ok {
		return []serror.SnapError{serror.New(ErrSubscriptionGroupAlreadyExists)}
	}

	subscriptionGroup := &subscriptionGroup{
		requestedMetrics: requested,
		requestedPlugins: plugins,
		configTree:       configTree,
		pluginControl:    s.pluginControl,
	}

	errs := subscriptionGroup.process(id)
	if errs != nil {
		return errs
	}
	s.subscriptionMap[id] = subscriptionGroup
	return nil
}

// Remove removes a subscription group given a subscription group ID.
func (s subscriptionGroups) Remove(id string) []serror.SnapError {
	s.Lock()
	defer s.Unlock()
	return s.remove(id)
}

func (s subscriptionGroups) remove(id string) []serror.SnapError {
	subscriptionGroup, ok := s.subscriptionMap[id]
	if !ok {
		return []serror.SnapError{serror.New(ErrSubscriptionGroupDoesNotExist)}
	}
	serrs := subscriptionGroup.unsubscribePlugins(id, s.subscriptionMap[id].plugins)
	delete(s.subscriptionMap, id)
	return serrs
}

// Get returns the metrics (core.Metric) and an array of serror.SnapError when
// provided a subscription ID. The array of serror.SnapError returned was
// produced the last time `process` was run which is important since
// unloading/loading a plugin may produce errors when the requested metrics
// are looked up in the metric catalog.  Those errors will be provided back to
// the caller of the subscription group on the next `CollectMetrics`.
// Returns `ErrSubscriptionGroupDoesNotExist` when the subscription group
// does not exist.
func (s subscriptionGroups) Get(id string) (map[string]metricTypes, []serror.SnapError, error) {
	s.Lock()
	defer s.Unlock()
	return s.get(id)
}

func (s subscriptionGroups) get(id string) (map[string]metricTypes, []serror.SnapError, error) {
	if _, ok := s.subscriptionMap[id]; !ok {
		return nil, nil, ErrSubscriptionGroupDoesNotExist
	}
	sg := s.subscriptionMap[id]
	return sg.metrics, sg.errors, nil
}

// Process compares the new set of plugins with the previous set of plugins
// for the given subscription group subscribing to plugins that were added
// and unsubscribing to those that were removed since the last time the
// subscription group was processed.
// Returns an array of errors ([]serror.SnapError) which can occur when
// mapping requested metrics to collector plugins and getting a core.Plugin
// from a core.Requested.Plugin.

// When processing a subscription group the resulting metrics grouped by plugin
// (subscriptionGroup.metrics) for all subscription groups are updated based
// on the requested metrics (subscriptionGroup.requestedMetrics).  Similarly
// the required plugins (subscriptionGroup.plugins) are also updated.
func (s *subscriptionGroups) Process() (errs []serror.SnapError) {
	s.Lock()
	defer s.Unlock()
	for id, group := range s.subscriptionMap {
		if serrs := group.process(id); serrs != nil {
			errs = append(errs, serrs...)
		}
	}
	return errs
}

func (s *subscriptionGroups) ValidateDeps(requested []core.RequestedMetric,
	plugins []core.SubscribedPlugin,
	configTree *cdata.ConfigDataTree, asserts ...core.SubscribedPluginAssert) (serrs []serror.SnapError) {

	// resolve requested metrics and map to collectors
	pluginToMetricMap, collectors, errs := s.getMetricsAndCollectors(requested, configTree)
	if errs != nil {
		serrs = append(serrs, errs...)
	}

	// Validate if schedule type is streaming and we have a non-streaming plugin or vice versa
	for _, assert := range asserts {
		if serr := assert(collectors); serr != nil {
			serrs = append(serrs, serr)
		}
	}
	if len(serrs) > 0 {
		return serrs
	}

	// validateMetricsTypes
	for _, pmt := range pluginToMetricMap {
		for _, mt := range pmt.Metrics() {
			errs := s.validateMetric(mt)
			if len(errs) > 0 {
				serrs = append(serrs, errs...)
			}
		}
	}
	// add collectors to plugins (processors and publishers)
	for _, collector := range collectors {
		plugins = append(plugins, collector)
	}

	// validate plugins
	for _, plg := range plugins {
		typ, err := core.ToPluginType(plg.TypeName())
		if err != nil {
			return []serror.SnapError{serror.New(err)}
		}
		mergedConfig := plg.Config().ReverseMerge(
			s.Config.Plugins.getPluginConfigDataNode(
				typ, plg.Name(), plg.Version()))
		errs := s.validatePluginSubscription(plg, mergedConfig)
		if len(errs) > 0 {
			serrs = append(serrs, errs...)
			return serrs
		}
	}
	return
}

// validatePluginUnloading checks if process of unloading the plugin is safe for existing running tasks.
// If the plugin is used by running task and there is no replacements, return an error with appropriate message
// containing ids of tasks which use the plugin, what blocks unloading process until they are stopped
func (s *subscriptionGroups) validatePluginUnloading(pluginToUnload *loadedPlugin) (errs []serror.SnapError) {
	s.Lock()
	defer s.Unlock()
	for id, group := range s.subscriptionMap {
		if err := group.validatePluginUnloading(id, pluginToUnload); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (p *subscriptionGroups) validatePluginSubscription(pl core.SubscribedPlugin, mergedConfig *cdata.ConfigDataNode) []serror.SnapError {
	var serrs = []serror.SnapError{}
	controlLogger.WithFields(log.Fields{
		"_block": "validate-plugin-subscription",
		"plugin": fmt.Sprintf("%s:%d", pl.Name(), pl.Version()),
	}).Info(fmt.Sprintf("validating dependencies for plugin %s:%d", pl.Name(), pl.Version()))
	lp, err := p.pluginManager.get(key(pl))
	if err != nil {
		serrs = append(serrs, pluginNotFoundError(pl))
		return serrs
	}

	if lp.ConfigPolicy != nil {
		ncd := lp.ConfigPolicy.Get([]string{""})
		_, errs := ncd.Process(mergedConfig.Table())
		if errs != nil && errs.HasErrors() {
			for _, e := range errs.Errors() {
				se := serror.New(e)
				se.SetFields(map[string]interface{}{"name": pl.Name(), "version": pl.Version()})
				serrs = append(serrs, se)
			}
		}
	}
	return serrs
}

func (s *subscriptionGroups) validateMetric(
	metric core.Metric) (serrs []serror.SnapError) {
	mts, err := s.metricCatalog.GetMetrics(metric.Namespace(), metric.Version())
	if err != nil {
		serrs = append(serrs, serror.New(err, map[string]interface{}{
			"name":    metric.Namespace().String(),
			"version": metric.Version(),
		}))
		return serrs
	}
	for _, m := range mts {

		// No metric found return error.
		if m == nil {
			serrs = append(
				serrs, serror.New(
					fmt.Errorf("no metric found cannot subscribe: (%s) version(%d)",
						metric.Namespace(), metric.Version())))
			continue
		}

		m.config = metric.Config()

		typ, serr := core.ToPluginType(m.Plugin.TypeName())
		if serr != nil {
			serrs = append(serrs, serror.New(err))
			continue
		}

		// merge global plugin config
		if m.config != nil {
			m.config.ReverseMergeInPlace(
				s.Config.Plugins.getPluginConfigDataNode(typ,
					m.Plugin.Name(), m.Plugin.Version()))
		} else {
			m.config = s.Config.Plugins.getPluginConfigDataNode(typ,
				m.Plugin.Name(), m.Plugin.Version())
		}

		// When a metric is added to the MetricCatalog, the policy of rules defined by the plugin is added to the metric's policy.
		// If no rules are defined for a metric, we set the metric's policy to an empty ConfigPolicyNode.
		// Checking m.policy for nil will not work, we need to check if rules are nil.
		if m.policy.HasRules() {
			if m.Config() == nil {
				fields := log.Fields{
					"metric":  m.Namespace(),
					"version": m.Version(),
					"plugin":  m.Plugin.Name(),
				}
				serrs = append(serrs, serror.New(ErrConfigRequiredForMetric, fields))
				continue
			}
			ncdTable, errs := m.policy.Process(m.Config().Table())
			if errs != nil && errs.HasErrors() {
				for _, e := range errs.Errors() {
					serrs = append(serrs, serror.New(e))
				}
				continue
			}
			m.config = cdata.FromTable(*ncdTable)
		}
	}

	return serrs
}

// pluginIsSubscribed returns true if a provided plugin has been found among subscribed plugins
// in the following subscription group
func (s *subscriptionGroup) pluginIsSubscribed(plugin *loadedPlugin) bool {
	// range over subscribed plugins to find if the plugin is there
	for _, sp := range s.plugins {
		if sp.TypeName() == plugin.TypeName() && sp.Name() == plugin.Name() && sp.Version() == plugin.Version() {
			return true
		}
	}
	return false
}

// validatePluginUnloading verifies if a given plugin might be unloaded without causing running task failures
func (s *subscriptionGroup) validatePluginUnloading(id string, plgToUnload *loadedPlugin) (serr serror.SnapError) {
	impacted := false
	if !s.pluginIsSubscribed(plgToUnload) {
		// the plugin is not subscribed, so the task is not impacted by its unloading
		return nil
	}
	controlLogger.WithFields(log.Fields{
		"_block":           "subscriptionGroup.validatePluginUnloading",
		"task-id":          id,
		"plugin-to-unload": plgToUnload.Key(),
	}).Debug("validating impact of unloading the plugin")

	for _, requestedMetric := range s.requestedMetrics {
		// get all plugins exposing the requested metric
		plgs, _ := s.GetPlugins(requestedMetric.Namespace())
		// when requested version is fixed (greater than 0), take into account only plugins in the requested version
		if requestedMetric.Version() > 0 {
			// skip those which are not impacted by unloading (version different than plgToUnload.Version())
			if requestedMetric.Version() == plgToUnload.Version() {
				plgsInVer := []core.CatalogedPlugin{}

				for _, plg := range plgs {
					if plg.Version() == requestedMetric.Version() {
						plgsInVer = append(plgsInVer, plg)
					}
				}
				// set plugins only in the requested version
				plgs = plgsInVer
			}
		}
		if len(plgs) == 1 && plgs[0].Key() == plgToUnload.Key() {
			// the requested metric is exposed only by the single plugin and there is no replacement
			impacted = true
			controlLogger.WithFields(log.Fields{
				"_block":           "subscriptionGroup.validatePluginUnloading",
				"task-id":          id,
				"plugin-to-unload": plgToUnload.Key(),
				"requested-metric": fmt.Sprintf("%s:%d", requestedMetric.Namespace(), requestedMetric.Version()),
			}).Errorf("unloading the plugin would cause missing in collection the requested metric")
		}
	}
	if impacted {
		serr = serror.New(ErrPluginCannotBeUnloaded, map[string]interface{}{
			"task-id":          id,
			"plugin-to-unload": plgToUnload.Key(),
		})
	}
	return serr
}

func (s *subscriptionGroup) process(id string) (serrs []serror.SnapError) {
	// gathers collectors based on requested metrics
	pluginToMetricMap, plugins, serrs := s.getMetricsAndCollectors(s.requestedMetrics, s.configTree)
	controlLogger.WithFields(log.Fields{
		"collectors": fmt.Sprintf("%+v", plugins),
		"metrics":    fmt.Sprintf("%+v", s.requestedMetrics),
	}).Debug("gathered collectors")

	// notice that requested plugins contains only processors and publishers
	for _, plugin := range s.requestedPlugins {
		// add defaults to plugins (exposed in a plugins ConfigPolicy)
		if lp, err := s.pluginManager.get(
			fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d",
				plugin.TypeName(),
				plugin.Name(),
				plugin.Version())); err == nil && lp.ConfigPolicy != nil {
			if policy := lp.ConfigPolicy.Get([]string{""}); policy != nil && len(policy.Defaults()) > 0 {
				// set defaults to plugin config
				plugin.Config().ApplyDefaults(policy.Defaults())
			}

			// update version info for subscribed processor or publisher
			version := plugin.Version()
			if version < 1 {
				version = lp.Version()
			}
			s := subscribedPlugin{
				name:     plugin.Name(),
				typeName: plugin.TypeName(),
				version:  version,
				config:   plugin.Config(),
			}

			// add processors and publishers to collectors just gathered
			plugins = append(plugins, s)
		}
	}
	// calculates those plugins that need to be subscribed and unsubscribed to
	subs, unsubs := comparePlugins(plugins, s.plugins)
	controlLogger.WithFields(log.Fields{
		"subs":   fmt.Sprintf("%+v", subs),
		"unsubs": fmt.Sprintf("%+v", unsubs),
	}).Debug("subscriptions")
	if len(subs) > 0 {
		if errs := s.subscribePlugins(id, subs); errs != nil {
			serrs = append(serrs, errs...)
		}
	}
	if len(unsubs) > 0 {
		if errs := s.unsubscribePlugins(id, unsubs); errs != nil {
			serrs = append(serrs, errs...)
		}
	}

	// updating view
	// metrics are grouped by plugin
	s.metrics = pluginToMetricMap
	s.plugins = plugins
	s.errors = serrs

	return serrs
}

func (s *subscriptionGroup) subscribePlugins(id string,
	plugins []core.SubscribedPlugin) (serrs []serror.SnapError) {
	plgs := make([]*loadedPlugin, len(plugins))
	// First range through plugins to verify if all required plugins
	// are available
	for i, sub := range plugins {
		plg, err := s.pluginManager.get(key(sub))
		if err != nil {
			serrs = append(serrs, pluginNotFoundError(sub))
			return serrs
		}
		plgs[i] = plg
	}

	// If all plugins are available, subscribe to pools and start
	// plugins as needed
	for _, plg := range plgs {
		controlLogger.WithFields(log.Fields{
			"name":    plg.Name(),
			"type":    plg.TypeName(),
			"version": plg.Version(),
			"_block":  "subscriptionGroup.subscribePlugins",
		}).Debug("plugin subscription")
		if plg.Details.Uri != nil {
			// this is a remote plugin
			pool, err := s.pluginRunner.AvailablePlugins().getOrCreatePool(plg.Key())
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			if pool.Count() < 1 {
				var resp plugin.Response
				res, err := http.Get(plg.Details.Uri.String())
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				err = json.Unmarshal(body, &resp)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				ap, err := newAvailablePlugin(resp, s.eventManager, nil, s.grpcSecurity)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				ap.SetIsRemote(true)
				err = pool.Insert(ap)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
			}
		} else {
			pool, err := s.pluginRunner.AvailablePlugins().getOrCreatePool(plg.Key())
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			pool.Subscribe(id)
			if pool.Eligible() {
				err = s.verifyPlugin(plg)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				err = s.pluginRunner.runPlugin(plg.Name(), plg.Details)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
			}
		}

		serr := s.sendPluginSubscriptionEvent(id, plg)
		if serr != nil {
			serrs = append(serrs, serr)
			return serrs
		}
	}
	return serrs
}

func (p *subscriptionGroup) unsubscribePlugins(id string,
	plugins []core.SubscribedPlugin) (serrs []serror.SnapError) {
	for _, plugin := range plugins {
		controlLogger.WithFields(log.Fields{
			"name":    plugin.Name(),
			"type":    plugin.TypeName(),
			"version": plugin.Version(),
			"_block":  "subscriptionGroup.unsubscribePlugins",
		}).Debug("plugin unsubscription")
		pool, err := p.pluginRunner.AvailablePlugins().getPool(key(plugin))
		if err != nil {
			serrs = append(serrs, err)
			return serrs
		}
		if pool != nil {
			pool.Unsubscribe(id)
		}
		serr := p.sendPluginUnsubscriptionEvent(id, plugin)
		if serr != nil {
			serrs = append(serrs, serr)
		}
	}
	return
}

func (p *subscriptionGroup) sendPluginSubscriptionEvent(taskID string,
	pl core.Plugin) serror.SnapError {
	pt, err := core.ToPluginType(pl.TypeName())
	if err != nil {
		return serror.New(err)
	}
	e := &control_event.PluginSubscriptionEvent{
		TaskId:        taskID,
		PluginType:    int(pt),
		PluginName:    pl.Name(),
		PluginVersion: pl.Version(),
	}

	if _, err := p.eventManager.Emit(e); err != nil {
		return serror.New(err)
	}
	return nil
}

func (p *subscriptionGroup) sendPluginUnsubscriptionEvent(taskID string,
	pl core.Plugin) serror.SnapError {
	pt, err := core.ToPluginType(pl.TypeName())
	if err != nil {
		return serror.New(err)
	}
	e := &control_event.PluginUnsubscriptionEvent{
		TaskId:        taskID,
		PluginType:    int(pt),
		PluginName:    pl.Name(),
		PluginVersion: pl.Version(),
	}
	if _, err := p.eventManager.Emit(e); err != nil {
		return serror.New(err)
	}
	return nil
}

// comparePlugins compares the new state of plugins with the previous state.
// It returns an array of plugins that need to be subscribed and an array of
// plugins that need to be unsubscribed.
func comparePlugins(newPlugins,
	oldPlugins []core.SubscribedPlugin) (adds,
	removes []core.SubscribedPlugin) {
	newMap := make(map[string]int)
	oldMap := make(map[string]int)

	for _, n := range newPlugins {
		newMap[key(n)]++
	}
	for _, o := range oldPlugins {
		oldMap[key(o)]++
	}

	for _, n := range newPlugins {
		if oldMap[key(n)] > 0 {
			oldMap[key(n)]--
			continue
		}
		adds = append(adds, n)
	}

	for _, o := range oldPlugins {
		if newMap[key(o)] > 0 {
			newMap[key(o)]--
			continue
		}
		removes = append(removes, o)
	}

	return
}

func pluginNotFoundError(pl core.SubscribedPlugin) serror.SnapError {
	se := serror.New(fmt.Errorf("Plugin not found: type(%s) name(%s) version(%d)", pl.TypeName(), pl.Name(), pl.Version()))
	se.SetFields(map[string]interface{}{
		"name":    pl.Name(),
		"version": pl.Version(),
		"type":    pl.TypeName(),
	})
	return se
}

func key(p core.SubscribedPlugin) string {
	return fmt.Sprintf("%v"+core.Separator+"%v"+core.Separator+"%v", p.TypeName(), p.Name(), p.Version())
}
