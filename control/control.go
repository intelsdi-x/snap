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
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/aci"
	"github.com/intelsdi-x/snap/pkg/psigning"
)

const (
	// PluginTrustDisabled - enum representing plugin trust disabled
	PluginTrustDisabled int = iota
	// PluginTrustEnabled - enum representing plugin trust enabled
	PluginTrustEnabled
	// PluginTrustWarn - enum representing plugin trust warning
	PluginTrustWarn
)

var (
	controlLogger = log.WithFields(log.Fields{
		"_module": "control",
	})

	// ErrLoadedPluginNotFound - error message when a loaded plugin is not found
	ErrLoadedPluginNotFound = errors.New("Loaded plugin not found")

	// ErrControllerNotStarted - error message when the Controller was not started
	ErrControllerNotStarted = errors.New("Must start Controller before calling Load()")
)

type executablePlugins []plugin.ExecutablePlugin

type pluginControl struct {
	// TODO, going to need coordination on changing of these
	RunningPlugins executablePlugins
	Started        bool
	Config         *Config

	autodiscoverPaths []string
	eventManager      *gomit.EventController

	pluginManager  managesPlugins
	metricCatalog  catalogsMetrics
	pluginRunner   runsPlugins
	signingManager managesSigning

	pluginTrust  int
	keyringFiles []string
}

type runsPlugins interface {
	Start() error
	Stop() []error
	AvailablePlugins() *availablePlugins
	AddDelegates(...gomit.Delegator)
	SetEmitter(gomit.Emitter)
	SetMetricCatalog(catalogsMetrics)
	SetPluginManager(managesPlugins)
	Monitor() *monitor
	runPlugin(*pluginDetails) error
}

type managesPlugins interface {
	teardown()
	get(string) (*loadedPlugin, error)
	all() map[string]*loadedPlugin
	LoadPlugin(*pluginDetails, gomit.Emitter) (*loadedPlugin, serror.SnapError)
	UnloadPlugin(core.Plugin) (*loadedPlugin, serror.SnapError)
	SetMetricCatalog(catalogsMetrics)
	GenerateArgs(pluginPath string) plugin.Arg
	SetPluginConfig(*pluginConfig)
}

type catalogsMetrics interface {
	GetMetric(core.Namespace, int) (*metricType, error)
	GetMetrics(core.Namespace, int) ([]*metricType, error)
	GetMatchedMetricTypes(core.Namespace, int) ([]*metricType, error)
	Add(*metricType)
	AddLoadedMetricType(*loadedPlugin, core.Metric) error
	RmUnloadedPluginMetrics(lp *loadedPlugin)
	GetVersions(core.Namespace) ([]*metricType, error)
	Fetch(core.Namespace) ([]*metricType, error)
	Item() (string, []*metricType)
	Next() bool
	Subscribe([]string, int) error
	Unsubscribe([]string, int) error
	GetPlugin(core.Namespace, int) (*loadedPlugin, error)
}

type managesSigning interface {
	ValidateSignature([]string, string, []byte) error
}

// PluginControlOpt is used to set optional parameters on the pluginControl struct
type PluginControlOpt func(*pluginControl)

// MaxRunningPlugins sets the maximum number of plugins to run per pool
func MaxRunningPlugins(m int) PluginControlOpt {
	return func(c *pluginControl) {
		strategy.MaximumRunningPlugins = m
	}
}

// CacheExpiration is the PluginControlOpt which sets the global metric cache TTL
func CacheExpiration(t time.Duration) PluginControlOpt {
	return func(c *pluginControl) {
		strategy.GlobalCacheExpiration = t
	}
}

// OptSetConfig sets the plugin control configuration.
func OptSetConfig(cfg *Config) PluginControlOpt {
	return func(c *pluginControl) {
		c.Config = cfg
		c.pluginManager.SetPluginConfig(cfg.Plugins)
	}
}

// New returns a new pluginControl instance
func New(cfg *Config) *pluginControl {
	// construct a slice of options from the input configuration
	opts := []PluginControlOpt{
		MaxRunningPlugins(cfg.MaxRunningPlugins),
		CacheExpiration(cfg.CacheExpiration.Duration),
		OptSetConfig(cfg),
	}
	c := &pluginControl{}
	c.Config = cfg
	// Initialize components
	//
	// Event Manager
	c.eventManager = gomit.NewEventController()

	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("pevent controller created")

	// Metric Catalog
	c.metricCatalog = newMetricCatalog()
	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("metric catalog created")

	// Plugin Manager
	c.pluginManager = newPluginManager()
	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("plugin manager created")
	// Plugin Manager needs a reference to the metric catalog
	c.pluginManager.SetMetricCatalog(c.metricCatalog)

	// Signing Manager
	c.signingManager = &psigning.SigningManager{}
	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("signing manager created")

	// Plugin Runner
	c.pluginRunner = newRunner()
	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("runner created")
	c.pluginRunner.AddDelegates(c.eventManager)
	c.pluginRunner.SetEmitter(c.eventManager)
	c.pluginRunner.SetMetricCatalog(c.metricCatalog)
	c.pluginRunner.SetPluginManager(c.pluginManager)

	// Start stuff
	err := c.pluginRunner.Start()
	if err != nil {
		panic(err)
	}

	// apply options

	// it is important that this happens last, as an option may
	// require that an internal member of c be constructed.
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (p *pluginControl) Name() string {
	return "control"
}

func (p *pluginControl) RegisterEventHandler(name string, h gomit.Handler) error {
	return p.eventManager.RegisterHandler(name, h)
}

// Begin handling load, unload, and inventory
func (p *pluginControl) Start() error {
	// Start pluginManager when pluginControl starts
	p.Started = true
	controlLogger.WithFields(log.Fields{
		"_block": "start",
	}).Info("control started")
	return nil
}

func (p *pluginControl) Stop() {
	p.Started = false
	controlLogger.WithFields(log.Fields{
		"_block": "stop",
	}).Info("control stopped")

	// stop runner
	err := p.pluginRunner.Stop()
	if err != nil {
		controlLogger.Error(err)
	}

	// stop running plugins
	for _, rp := range p.RunningPlugins {
		controlLogger.Debug("Stopping running plugin")
		rp.Kill()
	}

	// unload plugins
	p.pluginManager.teardown()
}

// Load is the public method to load a plugin into
// the LoadedPlugins array and issue an event when
// successful.
func (p *pluginControl) Load(rp *core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError) {
	f := map[string]interface{}{
		"_block": "load",
	}

	details, serr := p.returnPluginDetails(rp)
	if serr != nil {
		return nil, serr
	}
	if details.IsPackage {
		defer os.RemoveAll(filepath.Dir(details.ExecPath))
	}

	controlLogger.WithFields(f).Info("plugin load called")
	if !p.Started {
		se := serror.New(ErrControllerNotStarted)
		se.SetFields(f)
		controlLogger.WithFields(f).Error(se)
		return nil, se
	}

	pl, se := p.pluginManager.LoadPlugin(details, p.eventManager)
	if se != nil {
		return nil, se
	}

	// If plugin was loaded from a package, remove ExecPath for
	// the temporary plugin that was used for load
	if pl.Details.IsPackage {
		pl.Details.ExecPath = ""
	}

	// defer sending event
	event := &control_event.LoadPluginEvent{
		Name:    pl.Meta.Name,
		Version: pl.Meta.Version,
		Type:    int(pl.Meta.Type),
		Signed:  pl.Details.Signed,
	}
	defer p.eventManager.Emit(event)
	return pl, nil
}

func (p *pluginControl) verifySignature(rp *core.RequestedPlugin) (bool, serror.SnapError) {
	f := map[string]interface{}{
		"_block": "verifySignature",
	}
	switch p.pluginTrust {
	case PluginTrustDisabled:
		return false, nil
	case PluginTrustEnabled:
		err := p.signingManager.ValidateSignature(p.keyringFiles, rp.Path(), rp.Signature())
		if err != nil {
			return false, serror.New(err)
		}
	case PluginTrustWarn:
		if rp.Signature() == nil {
			controlLogger.WithFields(f).Warn("Loading unsigned plugin ", rp.Path())
			return false, nil
		}
		err := p.signingManager.ValidateSignature(p.keyringFiles, rp.Path(), rp.Signature())
		if err != nil {
			return false, serror.New(err)
		}
	}
	return true, nil

}

func (p *pluginControl) returnPluginDetails(rp *core.RequestedPlugin) (*pluginDetails, serror.SnapError) {
	details := &pluginDetails{}
	var serr serror.SnapError
	//Check plugin signing
	details.Signed, serr = p.verifySignature(rp)
	if serr != nil {
		return nil, serr
	}

	details.Path = rp.Path()
	details.CheckSum = rp.CheckSum()
	details.Signature = rp.Signature()

	if filepath.Ext(rp.Path()) == ".aci" {
		f, err := os.Open(rp.Path())
		if err != nil {
			return nil, serror.New(err)
		}
		defer f.Close()
		if err := aci.Validate(f); err != nil {
			return nil, serror.New(err)
		}
		tempPath, err := aci.Extract(f)
		if err != nil {
			return nil, serror.New(err)
		}
		details.ExecPath = path.Join(tempPath, "rootfs")
		if details.Manifest, err = aci.Manifest(f); err != nil {
			return nil, serror.New(err)
		}
		details.Exec = details.Manifest.App.Exec[0]
		details.IsPackage = true
	} else {
		details.IsPackage = false
		details.Exec = filepath.Base(rp.Path())
		details.ExecPath = filepath.Dir(rp.Path())
	}

	return details, nil
}

func (p *pluginControl) Unload(pl core.Plugin) (core.CatalogedPlugin, serror.SnapError) {
	up, err := p.pluginManager.UnloadPlugin(pl)
	if err != nil {
		return nil, err
	}

	event := &control_event.UnloadPluginEvent{
		Name:    up.Meta.Name,
		Version: up.Meta.Version,
		Type:    int(up.Meta.Type),
	}
	defer p.eventManager.Emit(event)
	return up, nil
}

func (p *pluginControl) SwapPlugins(in *core.RequestedPlugin, out core.CatalogedPlugin) serror.SnapError {
	details, serr := p.returnPluginDetails(in)
	if serr != nil {
		return serr
	}
	if details.IsPackage {
		defer os.RemoveAll(filepath.Dir(details.ExecPath))
	}

	lp, err := p.pluginManager.LoadPlugin(details, p.eventManager)
	if err != nil {
		return err
	}

	// Make sure plugin types and names are the same
	if lp.TypeName() != out.TypeName() || lp.Name() != out.Name() {
		serr := serror.New(errors.New("Plugin types and names must match."))
		serr.SetFields(map[string]interface{}{
			"in-type":  lp.TypeName(),
			"out-type": out.TypeName(),
			"in-name":  lp.Name(),
			"out-name": out.Name(),
		})
		_, err := p.pluginManager.UnloadPlugin(lp)
		if err != nil {
			se := serror.New(errors.New("Failed to rollback after error"))
			se.SetFields(map[string]interface{}{
				"original-unload-error": serr.Error(),
				"rollback-unload-error": err.Error(),
			})
			return se
		}
		return serr
	}

	up, err := p.pluginManager.UnloadPlugin(out)
	if err != nil {
		_, err2 := p.pluginManager.UnloadPlugin(lp)
		if err2 != nil {
			se := serror.New(errors.New("Failed to rollback after error"))
			se.SetFields(map[string]interface{}{
				"original-unload-error": err.Error(),
				"rollback-unload-error": err2.Error(),
			})
			return se
		}
		return err
	}

	event := &control_event.SwapPluginsEvent{
		LoadedPluginName:      lp.Meta.Name,
		LoadedPluginVersion:   lp.Meta.Version,
		UnloadedPluginName:    up.Meta.Name,
		UnloadedPluginVersion: up.Meta.Version,
		PluginType:            int(lp.Meta.Type),
	}
	defer p.eventManager.Emit(event)

	return nil
}

func (p *pluginControl) ValidateDeps(mts []core.Metric, cdt *cdata.ConfigDataTree, plugins []core.SubscribedPlugin) []serror.SnapError {
	var serrs []serror.SnapError

	errs := p.validateMetricsTypeSubscriptions(mts, cdt)
	if len(errs) > 0 {
		serrs = append(serrs, errs...)
	}

	//validate plugins (publishers and processors)
	for _, plg := range plugins {
		typ, err := core.ToPluginType(plg.TypeName())
		if err != nil {
			return []serror.SnapError{serror.New(err)}
		}
		plg.Config().ReverseMerge(p.Config.Plugins.getPluginConfigDataNode(typ, plg.Name(), plg.Version()))
		errs := p.validatePluginSubscription(plg)
		if len(errs) > 0 {
			serrs = append(serrs, errs...)
			return serrs
		}
	}

	return serrs
}

func (p *pluginControl) validatePluginSubscription(pl core.SubscribedPlugin) []serror.SnapError {
	var serrs = []serror.SnapError{}
	controlLogger.WithFields(log.Fields{
		"_block": "validate-plugin-subscription",
		"plugin": fmt.Sprintf("%s:%d", pl.Name(), pl.Version()),
	}).Info(fmt.Sprintf("validating dependencies for plugin %s:%d", pl.Name(), pl.Version()))
	lp, err := p.pluginManager.get(fmt.Sprintf("%s:%s:%d", pl.TypeName(), pl.Name(), pl.Version()))
	if err != nil {
		se := serror.New(fmt.Errorf("Plugin not found: type(%s) name(%s) version(%d)", pl.TypeName(), pl.Name(), pl.Version()))
		se.SetFields(map[string]interface{}{
			"name":    pl.Name(),
			"version": pl.Version(),
			"type":    pl.TypeName(),
		})
		serrs = append(serrs, se)
		return serrs
	}

	if lp.ConfigPolicy != nil {
		ncd := lp.ConfigPolicy.Get([]string{""})
		_, errs := ncd.Process(pl.Config().Table())
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

// validateMetricTypeSubscription validates the subscription for given requested metric
func (p *pluginControl) validateMetricTypeSubscription(rmt core.RequestedMetric, cdt *cdata.ConfigDataTree) []serror.SnapError {
	serrs := []serror.SnapError{}

	controlLogger.WithFields(log.Fields{
		"_block":    "validate-metric-subscription",
		"namespace": rmt.Namespace(),
		"version":   rmt.Version(),
	}).Info("subscription called on metric")

	// match the requested metric with the cataloged metrics types
	mts, err := p.metricCatalog.GetMatchedMetricTypes(rmt.Namespace(), rmt.Version())

	if err != nil {
		serrs = append(serrs, serror.New(err, map[string]interface{}{
			"name":    rmt.Namespace().String(),
			"version": rmt.Version(),
		}))
		return serrs
	}

	if cdt == nil {
		// initialize config tree if nil is provided
		cdt = cdata.NewTree()
	}

	for _, m := range mts {
		config := cdt.Get(m.Namespace().Strings())

		if config == nil {
			config = cdata.NewNode()
		}
		m.config = config

		typ, serr := core.ToPluginType(m.Plugin.TypeName())
		if serr != nil {
			return []serror.SnapError{serror.New(err)}
		}

		// merge global plugin config
		if m.config != nil {
			m.config.ReverseMerge(p.Config.Plugins.getPluginConfigDataNode(typ, m.Plugin.Name(), m.Plugin.Version()))
		} else {
			m.config = p.Config.Plugins.getPluginConfigDataNode(typ, m.Plugin.Name(), m.Plugin.Version())
		}

		// When a metric is added to the MetricCatalog, the policy of rules defined by the plugin is added to the metric's policy.
		// If no rules are defined for a metric, we set the metric's policy to an empty ConfigPolicyNode.
		// Checking m.policy for nil will not work, we need to check if rules are nil.
		if m.policy.HasRules() {

			ncdTable, errs := m.policy.Process(m.Config().Table())
			if errs != nil && errs.HasErrors() {
				for _, e := range errs.Errors() {
					serrs = append(serrs, serror.New(e))
				}
				return serrs
			}
			m.config = cdata.FromTable(*ncdTable)
		}
	}

	return serrs
}

// validateMetricsTypeSubscriptions validates subscriptions for all requested metrics
func (p *pluginControl) validateMetricsTypeSubscriptions(mts []core.Metric, cdt *cdata.ConfigDataTree) []serror.SnapError {
	var serrs []serror.SnapError
	for _, mt := range mts {
		errs := p.validateMetricTypeSubscription(mt, cdt)
		if len(errs) > 0 {
			serrs = append(serrs, errs...)
		}
	}
	return serrs
}

type gatheredPlugin struct {
	plugin           core.Plugin
	subscriptionType strategy.SubscriptionType
}

func (p *pluginControl) gatherCollectors(rmts []core.Metric) ([]gatheredPlugin, []serror.SnapError) {
	var (
		plugins []gatheredPlugin
		serrs   []serror.SnapError
	)

	// here we resolve and retrieve plugins for each metric type.
	// if the requested metric type version is < 1, we treat that as
	// latest as with plugins.  The following loops create a set
	// of plugins with proper versions needed to discern the subscription
	// types.

	colPlugins := make(map[string]gatheredPlugin)

	for _, rmt := range rmts {
		// get matched metrics types for requested metrics
		mts, err := p.metricCatalog.GetMatchedMetricTypes(rmt.Namespace(), rmt.Version())
		if err != nil {
			serrs = append(serrs, serror.New(err, map[string]interface{}{
				"name":    rmt.Namespace().String(),
				"version": rmt.Version(),
			}))
			continue
		}

		subType := strategy.BoundSubscriptionType
		if rmt.Version() < 1 {
			subType = strategy.UnboundSubscriptionType
		}

		for _, mt := range mts {
			lp := mt.Plugin

			if lp == nil {
				serrs = append(serrs, serror.New(errMetricNotFound, map[string]interface{}{
					"name":    mt.Namespace().String(),
					"version": mt.Version(),
				}))
				continue
			}

			colPlugins[fmt.Sprintf("%s:%d", lp.Key(), subType)] = gatheredPlugin{
				plugin:           lp,
				subscriptionType: subType,
			}
		}
	}

	if len(serrs) > 0 {
		return plugins, serrs
	}

	for _, lp := range colPlugins {
		plugins = append(plugins, lp)
	}
	if len(plugins) == 0 {
		serrs = append(serrs, serror.New(errors.New("No plugins found")))
		return nil, serrs
	}
	return plugins, nil
}

func (p *pluginControl) subscribeCollectorDeps(taskID string, mts []core.Metric) []serror.SnapError {
	var serrs []serror.SnapError
	collectors, errs := p.gatherCollectors(mts)
	if len(errs) > 0 {
		serrs = append(serrs, errs...)
	}

	for _, gc := range collectors {
		pool, err := p.pluginRunner.AvailablePlugins().getOrCreatePool(fmt.Sprintf("%s:%s:%d", gc.plugin.TypeName(), gc.plugin.Name(), gc.plugin.Version()))
		if err != nil {
			serrs = append(serrs, serror.New(err))
			return serrs
		}
		pool.Subscribe(taskID, gc.subscriptionType)
		if pool.Eligible() {
			err = p.verifyPlugin(gc.plugin.(*loadedPlugin))
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			err = p.pluginRunner.runPlugin(gc.plugin.(*loadedPlugin).Details)
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
		}
		serr := p.sendPluginSubscriptionEvent(taskID, gc.plugin)
		if serr != nil {
			serrs = append(serrs, serr)
		}
	}
	return serrs
}

func (p *pluginControl) SubscribeDeps(taskID string, mts []core.Metric, plugins []core.Plugin) []serror.SnapError {
	var serrs []serror.SnapError

	errs := p.subscribeCollectorDeps(taskID, mts)
	if len(errs) > 0 {
		serrs = append(serrs, errs...)
	}

	for _, sub := range plugins {
		// pools are created statically, not with keys like "publisher:foo:-1"
		// here we check to see if the version of the incoming plugin is -1, and
		// if it is, we look up the latest in loaded plugins, and use that key to
		// create the pool.
		if sub.Version() < 1 {
			latest, err := p.pluginManager.get(fmt.Sprintf("%s:%s:%d", sub.TypeName(), sub.Name(), sub.Version()))
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			pool, err := p.pluginRunner.AvailablePlugins().getOrCreatePool(latest.Key())
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			pool.Subscribe(taskID, strategy.UnboundSubscriptionType)
			if pool.Eligible() {
				err = p.verifyPlugin(latest)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				err = p.pluginRunner.runPlugin(latest.Details)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
			}
		} else {
			pool, err := p.pluginRunner.AvailablePlugins().getOrCreatePool(fmt.Sprintf("%s:%s:%d", sub.TypeName(), sub.Name(), sub.Version()))
			if err != nil {
				serrs = append(serrs, serror.New(err))
				return serrs
			}
			pool.Subscribe(taskID, strategy.BoundSubscriptionType)
			if pool.Eligible() {
				pl, err := p.pluginManager.get(fmt.Sprintf("%s:%s:%d", sub.TypeName(), sub.Name(), sub.Version()))
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				err = p.verifyPlugin(pl)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
				err = p.pluginRunner.runPlugin(pl.Details)
				if err != nil {
					serrs = append(serrs, serror.New(err))
					return serrs
				}
			}
		}
		serr := p.sendPluginSubscriptionEvent(taskID, sub)
		if serr != nil {
			serrs = append(serrs, serr)
		}
	}

	return serrs
}

func (p *pluginControl) verifyPlugin(lp *loadedPlugin) error {
	b, err := ioutil.ReadFile(lp.Details.Path)
	if err != nil {
		return err
	}
	cs := sha256.Sum256(b)
	if lp.Details.CheckSum != cs {
		return fmt.Errorf(fmt.Sprintf("Current plugin checksum (%x) does not match checksum when plugin was first loaded (%x).", cs, lp.Details.CheckSum))
	}
	if lp.Details.Signed {
		return p.signingManager.ValidateSignature(p.keyringFiles, lp.Details.Path, lp.Details.Signature)
	}
	return nil
}

func (p *pluginControl) sendPluginSubscriptionEvent(taskID string, pl core.Plugin) serror.SnapError {
	pt, err := core.ToPluginType(pl.TypeName())
	if err != nil {
		return serror.New(err)
	}
	e := &control_event.PluginSubscriptionEvent{
		TaskId:           taskID,
		PluginType:       int(pt),
		PluginName:       pl.Name(),
		PluginVersion:    pl.Version(),
		SubscriptionType: int(strategy.UnboundSubscriptionType),
	}
	if pl.Version() > 0 {
		e.SubscriptionType = int(strategy.BoundSubscriptionType)
	}
	if _, err := p.eventManager.Emit(e); err != nil {
		return serror.New(err)
	}
	return nil
}

func (p *pluginControl) UnsubscribeDeps(taskID string, mts []core.Metric, plugins []core.Plugin) []serror.SnapError {
	var serrs []serror.SnapError

	collectors, errs := p.gatherCollectors(mts)
	if len(errs) > 0 {
		serrs = append(serrs, errs...)
	}
	for _, gc := range collectors {
		plugins = append(plugins, gc.plugin)
	}

	for _, sub := range plugins {
		pool, err := p.pluginRunner.AvailablePlugins().getPool(fmt.Sprintf("%s:%s:%d", sub.TypeName(), sub.Name(), sub.Version()))
		if err != nil {
			serrs = append(serrs, err)
			return serrs
		}
		if pool != nil {
			pool.Unsubscribe(taskID)
		}
		serr := p.sendPluginUnsubscriptionEvent(taskID, sub)
		if serr != nil {
			serrs = append(serrs, serr)
		}
	}

	return serrs
}

func (p *pluginControl) sendPluginUnsubscriptionEvent(taskID string, pl core.Plugin) serror.SnapError {
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

// SetMonitorOptions exposes monitors options
func (p *pluginControl) SetMonitorOptions(options ...monitorOption) {
	p.pluginRunner.Monitor().Option(options...)
}

// returns the loaded plugin collection
// NOTE: The returned data from this function should be considered constant and read only
func (p *pluginControl) PluginCatalog() core.PluginCatalog {
	table := p.pluginManager.all()
	plugins := make([]core.CatalogedPlugin, len(table))
	i := 0
	for _, plugin := range table {
		plugins[i] = plugin
		i++
	}
	return plugins
}

// AvailablePlugins returns pointers to all the running plugins in the pools
// NOTE: The returned data from this function should be considered constant and read only
func (p *pluginControl) AvailablePlugins() []core.AvailablePlugin {
	var caps []core.AvailablePlugin
	for _, ap := range p.pluginRunner.AvailablePlugins().all() {
		caps = append(caps, ap)
	}
	return caps
}

// MetricCatalog returns the entire metric catalog
// NOTE: The returned data from this function should be considered constant and read only
func (p *pluginControl) MetricCatalog() ([]core.CatalogedMetric, error) {
	return p.FetchMetrics(core.Namespace{}, 0)
}

// FetchMetrics returns the metrics which fall under the given namespace
// NOTE: The returned data from this function should be considered constant and read only
func (p *pluginControl) FetchMetrics(ns core.Namespace, version int) ([]core.CatalogedMetric, error) {
	mts, err := p.metricCatalog.Fetch(ns)
	if err != nil {
		return nil, err
	}
	cmt := make([]core.CatalogedMetric, 0, len(mts))
	for _, mt := range mts {
		if version > 0 {
			if mt.version == version {
				cmt = append(cmt, mt)
			}
		} else {
			cmt = append(cmt, mt)
		}
	}
	return cmt, nil
}

func (p *pluginControl) GetMetric(ns core.Namespace, ver int) (core.CatalogedMetric, error) {
	return p.metricCatalog.GetMetric(ns, ver)
}

func (p *pluginControl) GetMetrics(ns core.Namespace, ver int) ([]core.CatalogedMetric, error) {
	mts, err := p.metricCatalog.GetMetrics(ns, ver)
	if err != nil {
		return nil, err
	}

	rmts := make([]core.CatalogedMetric, len(mts))
	for i, m := range mts {
		rmts[i] = m
	}
	return rmts, nil

}

func (p *pluginControl) GetMetricVersions(ns core.Namespace) ([]core.CatalogedMetric, error) {
	mts, err := p.metricCatalog.GetVersions(ns)
	if err != nil {
		return nil, err
	}

	rmts := make([]core.CatalogedMetric, len(mts))
	for i, m := range mts {
		rmts[i] = m
	}
	return rmts, nil
}

func (p *pluginControl) MetricsExists(mns core.Namespace, ver int) bool {
	_, err := p.metricCatalog.GetMetric(mns, ver)
	if err == nil {
		return true
	}
	return false
}

// gatherMetricTypesGroupedByPlugin gathers matched metrics types for requested metrics and groups them by plugins to call
func (p *pluginControl) gatherMetricTypesGroupedByPlugin(rmts []core.RequestedMetric, cdt *cdata.ConfigDataTree) (map[string]metricTypes, []error) {
	pmts := make(map[string]metricTypes)
	var errs []error

	if cdt == nil {
		// initialize config tree if nil is provided
		cdt = cdata.NewTree()
	}

	for _, rmt := range rmts {
		// for each requested metric get matched metrics types on appropriate version
		// (if version < 1 choose the latest)
		mts, err := p.metricCatalog.GetMatchedMetricTypes(rmt.Namespace(), rmt.Version())

		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, mt := range mts {
			lp := mt.Plugin
			if lp == nil {
				errs = append(errs, errorMetricNotFound(mt.Namespace().Strings(), mt.Version()))
				continue
			}

			// get config
			config := cdt.Get(mt.Namespace().Strings())

			if config == nil {
				config = cdata.NewNode()
			}
			returnedmt := plugin.MetricType{
				Namespace_:          mt.Namespace(),
				LastAdvertisedTime_: mt.LastAdvertisedTime(),
				Version_:            mt.Version(),
				Tags_:               mt.Tags(),
				Config_:             config,
				Unit_:               mt.Unit(),
			}

			// group matched metrics by plugin
			pluginKey := lp.Key()
			pmt, _ := pmts[pluginKey]
			pmt.plugin = lp
			pmt.metricTypes = append(pmt.metricTypes, returnedmt)
			pmts[pluginKey] = pmt

		}

	}

	return pmts, errs
}

// CollectMetrics is a blocking call to collector plugins returning a collection
// of metrics and errors.  If an error is encountered no metrics will be
// returned.
func (p *pluginControl) CollectMetrics(rmts []core.RequestedMetric, cdt *cdata.ConfigDataTree, deadline time.Time, taskID string) (metrics []core.Metric, errs []error) {

	// gathers metrics to collect grouped by plugin
	pluginToMetricMap, errs := p.gatherMetricTypesGroupedByPlugin(rmts, cdt)

	if len(errs) > 0 {
		return
	}

	cMetrics := make(chan []core.Metric)
	cError := make(chan error)
	var wg sync.WaitGroup

	// For each available plugin call available plugin using RPC client and wait for response (goroutines)
	for pluginKey, pmt := range pluginToMetricMap {

		// if plugin does not have a pool, try to make new subscription
		if !p.pluginRunner.AvailablePlugins().hasPool(pluginKey) {
			// this happens only when the queried metric in task manifest matches with metrics
			// from several plugins (e.q. "/intel/*") and during running the task, a new plugin was loaded
			// and its metrics are below the requested metric's namespace

			// check if metrics policy is fulfilled by task config
			serrs := p.validateMetricsTypeSubscriptions(pmt.metricTypes, cdt)

			if serrs == nil {
				serrs = p.subscribeCollectorDeps(taskID, pmt.metricTypes)
			} else {
				// metrics policy is not fulfilled
				controlLogger.WithFields(log.Fields{
					"_module": "control",
					"_file":   "control.go",
					"_block":  "collect-metrics",
					"plugin":  pluginKey,
					"task-id": taskID,
					"error":   serrs,
				}).Error("additional subscription is invalid")
			}

			for _, serr := range serrs {
				errs = append(errs, errors.New(serr.Error()))
			}

		}

		if len(errs) > 0 {
			return
		}

		// merge global plugin config into the config for the metric
		for _, mt := range pmt.metricTypes {

			if mt.Config() != nil {
				mt.Config().ReverseMerge(p.Config.Plugins.getPluginConfigDataNode(core.CollectorPluginType, pmt.plugin.Name(), pmt.plugin.Version()))
			}
		}

		wg.Add(1)

		go func(pluginKey string, mt []core.Metric) {
			mts, err := p.pluginRunner.AvailablePlugins().collectMetrics(pluginKey, mt, taskID)
			if err != nil {
				cError <- err
			} else {
				cMetrics <- mts
			}
		}(pluginKey, pmt.metricTypes)
	}

	go func() {
		for m := range cMetrics {
			// Reapply standard tags after collection as a precaution.  It is common for
			// plugin authors to inadvertently overwrite or not pass along the data
			// passed to CollectMetrics so we will help them out here.
			for i := range m {
				m[i] = addStandardTags(m[i])
			}
			metrics = append(metrics, m...)
			wg.Done()
		}
	}()

	go func() {
		for e := range cError {
			errs = append(errs, e)
			wg.Done()
		}
	}()

	wg.Wait()
	close(cMetrics)
	close(cError)

	if len(errs) > 0 {
		return nil, errs
	}
	return
}

// PublishMetrics
func (p *pluginControl) PublishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error {
	// merge global plugin config into the config for this request
	// without over-writing the task specific config
	cfg := p.Config.Plugins.getPluginConfigDataNode(core.PublisherPluginType, pluginName, pluginVersion).Table()
	merged := make(map[string]ctypes.ConfigValue)
	for k, v := range cfg {
		merged[k] = v
	}
	for k, v := range config {
		merged[k] = v
	}

	return p.pluginRunner.AvailablePlugins().publishMetrics(contentType, content, pluginName, pluginVersion, merged, taskID)
}

// ProcessMetrics
func (p *pluginControl) ProcessMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) (string, []byte, []error) {
	// merge global plugin config into the config for this request
	// without over-writing the task specific config
	cfg := p.Config.Plugins.getPluginConfigDataNode(core.ProcessorPluginType, pluginName, pluginVersion).Table()
	merged := make(map[string]ctypes.ConfigValue)
	for k, v := range cfg {
		merged[k] = v
	}
	for k, v := range config {
		merged[k] = v
	}

	return p.pluginRunner.AvailablePlugins().processMetrics(contentType, content, pluginName, pluginVersion, merged, taskID)
}

// GetPluginContentTypes returns accepted and returned content types for the
// loaded plugin matching the provided name, type and version.
// If the version provided is 0 or less the newest plugin by version will be
// returned.
func (p *pluginControl) GetPluginContentTypes(n string, t core.PluginType, v int) ([]string, []string, error) {
	lp, err := p.pluginManager.get(fmt.Sprintf("%s:%s:%d", t.String(), n, v))
	if err != nil {
		return nil, nil, err
	}
	return lp.Meta.AcceptedContentTypes, lp.Meta.ReturnedContentTypes, nil
}

func (p *pluginControl) SetAutodiscoverPaths(paths []string) {
	p.autodiscoverPaths = paths
}

func (p *pluginControl) GetAutodiscoverPaths() []string {
	return p.autodiscoverPaths
}

func (p *pluginControl) SetPluginTrustLevel(trust int) {
	p.pluginTrust = trust
}

func (p *pluginControl) SetKeyringFile(keyring string) {
	p.keyringFiles = append(p.keyringFiles, keyring)
}

type requestedPlugin struct {
	name    string
	version int
	config  *cdata.ConfigDataNode
}

func (r *requestedPlugin) Name() string {
	return r.name
}

func (r *requestedPlugin) Version() int {
	return r.version
}

func (r *requestedPlugin) Config() *cdata.ConfigDataNode {
	return r.config
}

// ------------------- helper struct and function for grouping metrics types ------

// just a tuple of loadedPlugin and metricType slice
type metricTypes struct {
	plugin      *loadedPlugin
	metricTypes []core.Metric
}

func (p *metricTypes) Count() int {
	return len(p.metricTypes)
}
