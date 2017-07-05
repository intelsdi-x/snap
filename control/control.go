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
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/client"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
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
	ErrControllerNotStarted = errors.New("Must start Controller before use")
)

type pluginControl struct {
	// TODO, going to need coordination on changing of these
	Started bool
	Config  *Config

	autodiscoverPaths []string
	eventManager      *gomit.EventController

	pluginManager  managesPlugins
	metricCatalog  catalogsMetrics
	pluginRunner   runsPlugins
	signingManager managesSigning

	pluginTrust  int
	keyringFiles []string
	// used to cleanly shutdown the GRPC server
	grpcServer  *grpc.Server
	closingChan chan bool
	wg          sync.WaitGroup

	subscriptionGroups ManagesSubscriptionGroups
	grpcSecurity       client.GRPCSecurity
}

type subscribedPlugin struct {
	typeName string
	name     string
	version  int
	config   *cdata.ConfigDataNode
}

func (s subscribedPlugin) TypeName() string              { return s.typeName }
func (s subscribedPlugin) Name() string                  { return s.name }
func (s subscribedPlugin) Version() int                  { return s.version }
func (s subscribedPlugin) Config() *cdata.ConfigDataNode { return s.config }

type runsPlugins interface {
	Start() error
	Stop() []error
	AvailablePlugins() *availablePlugins
	AddDelegates(...gomit.Delegator)
	SetEmitter(gomit.Emitter)
	SetMetricCatalog(catalogsMetrics)
	SetPluginManager(managesPlugins)
	Monitor() *monitor
	runPlugin(string, *pluginDetails) error
	SetPluginLoadTimeout(int)
}

type managesPlugins interface {
	teardown()
	get(string) (*loadedPlugin, error)
	all() map[string]*loadedPlugin
	LoadPlugin(*pluginDetails, gomit.Emitter) (*loadedPlugin, serror.SnapError)
	UnloadPlugin(core.Plugin) (*loadedPlugin, serror.SnapError)
	SetMetricCatalog(catalogsMetrics)
	GenerateArgs(logLevel int) plugin.Arg
	SetPluginConfig(*pluginConfig)
	GetPluginConfig() *pluginConfig
	SetPluginTags(map[string]map[string]string)
	AddStandardAndWorkflowTags(core.Metric, map[string]map[string]string) core.Metric
	SetPluginLoadTimeout(int)
}

type catalogsMetrics interface {
	GetMetric(core.Namespace, int) (*metricType, error)
	GetMetrics(core.Namespace, int) ([]*metricType, error)
	Add(*metricType)
	AddLoadedMetricType(*loadedPlugin, core.Metric) error
	RmUnloadedPluginMetrics(lp *loadedPlugin)
	GetVersions(core.Namespace) ([]*metricType, error)
	Fetch(core.Namespace) ([]*metricType, error)
	Keys() []string
	Subscribe([]string, int) error
	Unsubscribe([]string, int) error
	GetPlugin(core.Namespace, int) (core.CatalogedPlugin, error)
	GetPlugins(core.Namespace) ([]core.CatalogedPlugin, error)
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
		c.pluginManager.SetPluginLoadTimeout(c.Config.PluginLoadTimeout)
		c.pluginRunner.SetPluginLoadTimeout(c.Config.PluginLoadTimeout)
	}
}

// OptSetTags sets the plugin control tags.
func OptSetTags(tags map[string]map[string]string) PluginControlOpt {
	return func(c *pluginControl) {
		c.pluginManager.SetPluginTags(tags)
	}
}

// MaximumPluginRestarts
func MaxPluginRestarts(cfg *Config) PluginControlOpt {
	return func(*pluginControl) {
		MaxPluginRestartCount = cfg.MaxPluginRestarts
	}
}

// New returns a new pluginControl instance
func New(cfg *Config) *pluginControl {
	// construct a slice of options from the input configuration
	opts := []PluginControlOpt{
		MaxRunningPlugins(cfg.MaxRunningPlugins),
		CacheExpiration(cfg.CacheExpiration.Duration),
		OptSetConfig(cfg),
		OptSetTags(cfg.Tags),
		MaxPluginRestarts(cfg),
	}
	c := &pluginControl{}
	c.Config = cfg
	// Initialize components
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

	managerOpts := []pluginManagerOpt{
		OptSetPprof(cfg.Pprof),
		OptSetTempDirPath(cfg.TempDirPath),
	}
	runnerOpts := []pluginRunnerOpt{}
	if cfg.IsTLSEnabled() {
		if cfg.CACertPaths != "" {
			certPaths := filepath.SplitList(cfg.CACertPaths)
			c.grpcSecurity = client.SecurityTLSExtended(cfg.TLSCertPath, cfg.TLSKeyPath, client.SecureClient, certPaths)
		} else {
			c.grpcSecurity = client.SecurityTLSEnabled(cfg.TLSCertPath, cfg.TLSKeyPath, client.SecureClient)
		}
		managerOpts = append(managerOpts, OptEnableManagerTLS(c.grpcSecurity))
		runnerOpts = append(runnerOpts, OptEnableRunnerTLS(c.grpcSecurity))
	}
	// Plugin Manager
	c.pluginManager = newPluginManager(managerOpts...)
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
	c.pluginRunner = newRunner(runnerOpts...)
	controlLogger.WithFields(log.Fields{
		"_block": "new",
	}).Debug("runner created")
	c.pluginRunner.AddDelegates(c.eventManager)
	c.pluginRunner.SetEmitter(c.eventManager)
	c.pluginRunner.SetMetricCatalog(c.metricCatalog)
	c.pluginRunner.SetPluginManager(c.pluginManager)

	// Pass runner events to control main module
	c.eventManager.RegisterHandler(c.Name(), c)

	// Create subscription group - used for managing a group of subscriptions
	c.subscriptionGroups = newSubscriptionGroups(c)

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

func (p *pluginControl) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		serrs := p.subscriptionGroups.Process()
		if serrs != nil {
			for _, err := range serrs {
				controlLogger.WithFields(log.Fields{
					"_block": "LoadPluginEvent",
				}).Error(err)
			}
		}
	case *control_event.UnloadPluginEvent:
		serrs := p.subscriptionGroups.Process()
		if serrs != nil {
			for _, err := range serrs {
				controlLogger.WithFields(log.Fields{
					"_block": "UnloadPluginEvent",
				}).Error(err)
			}
		}
	default:
		runnerLog.WithFields(log.Fields{
			"_block": "handle-events",
			"event":  v.Namespace(),
		}).Info("Nothing to do for this event")
	}
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

	//Autodiscover
	if p.Config.AutoDiscoverPath != "" {
		controlLogger.WithFields(log.Fields{
			"_block": "start",
		}).Info("auto discover path is enabled")

		paths := filepath.SplitList(p.Config.AutoDiscoverPath)
		p.SetAutodiscoverPaths(paths)
		for _, pa := range paths {
			fullPath, err := filepath.Abs(pa)
			if err != nil {
				controlLogger.WithFields(log.Fields{
					"_block":           "start",
					"autodiscoverpath": pa,
				}).Fatal(err)
			}
			controlLogger.WithFields(log.Fields{
				"_block": "start",
			}).Info("autoloading plugins from: ", fullPath)
			files, err := ioutil.ReadDir(fullPath)
			if err != nil {
				controlLogger.WithFields(log.Fields{
					"_block":           "start",
					"autodiscoverpath": pa,
				}).Fatal(err)
			}
			for _, file := range files {
				fileName := file.Name()

				statCheck := file
				if file.Mode()&os.ModeSymlink != 0 {
					realPath, err := filepath.EvalSymlinks(filepath.Join(fullPath, fileName))
					if err != nil {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": pa,
							"error":            err,
							"plugin":           fileName,
						}).Error("Cannot follow symlink")
						continue
					}
					statCheck, err = os.Stat(realPath)
					if err != nil {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": pa,
							"error":            err,
							"plugin":           fileName,
							"target-path":      realPath,
						}).Error("Target of symlink inacessible")
						continue
					}
				}

				if statCheck.IsDir() {
					controlLogger.WithFields(log.Fields{
						"_block":           "start",
						"autodiscoverpath": pa,
					}).Warning("Ignoring subdirectory: ", fileName)
					continue
				}
				// Ignore tasks files (JSON and YAML)
				fname := strings.ToLower(fileName)
				if strings.HasSuffix(fname, ".json") || strings.HasSuffix(fname, ".yaml") || strings.HasSuffix(fname, ".yml") {
					controlLogger.WithFields(log.Fields{
						"_block":           "start",
						"autodiscoverpath": pa,
					}).Warning("Ignoring JSON/Yaml file: ", fileName)
					continue
				}
				// if the file is a plugin package (which would have a suffix of '.aci') or if the file
				// is not a plugin signing file (which would have a suffix of '.asc'), then attempt to
				// automatically load the file as a plugin
				if strings.HasSuffix(fileName, ".aci") || !(strings.HasSuffix(fileName, ".asc")) {
					// check to makd sure the file is executable by someone (even if it isn't you); if no one
					// can execute this file then skip it (and include a warning in the log output)
					if (statCheck.Mode() & 0111) == 0 {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": pa,
							"plugin":           fileName,
						}).Warn("Auto-loading of plugin '", fileName, "' skipped (plugin not executable)")
						continue
					}
					rp, err := core.NewRequestedPlugin(path.Join(fullPath, fileName), p.GetTempDir(), nil)
					if err != nil {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": pa,
							"plugin":           fileName,
						}).Error(err)
					}
					signatureFile := fileName + ".asc"
					if _, err := os.Stat(path.Join(fullPath, signatureFile)); err == nil {
						err = rp.ReadSignatureFile(path.Join(fullPath, signatureFile))
						if err != nil {
							controlLogger.WithFields(log.Fields{
								"_block":           "start",
								"autodiscoverpath": pa,
								"plugin":           fileName + ".asc",
							}).Error(err)
						}
					}
					pl, err := p.Load(rp)
					if err != nil {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": fullPath,
							"plugin":           fileName,
						}).Error(err)
					} else {
						controlLogger.WithFields(log.Fields{
							"_block":           "start",
							"autodiscoverpath": fullPath,
							"plugin-file-name": fileName,
							"plugin-name":      pl.Name(),
							"plugin-version":   pl.Version(),
							"plugin-type":      pl.TypeName(),
						}).Info("Loading plugin")
					}
				}
			}
		}
	} else {
		controlLogger.WithFields(log.Fields{
			"_block": "start",
		}).Info("auto discover path is disabled")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", p.Config.ListenAddr, p.Config.ListenPort))
	if err != nil {
		controlLogger.WithField("error", err.Error()).Error("Failed to start control grpc listener")
		return err
	}

	opts := []grpc.ServerOption{}
	p.closingChan = make(chan bool, 1)
	p.grpcServer = grpc.NewServer(opts...)
	rpc.RegisterMetricManagerServer(p.grpcServer, &ControlGRPCServer{p})
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		err := p.grpcServer.Serve(lis)
		if err != nil {
			select {
			case <-p.closingChan:
			// If we called Stop() then there will be a value in p.closingChan, so
			// we'll get here and we can exit without showing the error.
			default:
				controlLogger.Fatal(err)
			}
		}
	}()

	return nil
}

func (p *pluginControl) Stop() {
	// set the Started flag to false (since we're stopping the server)
	p.Started = false

	// and add a boolean to the p.closingChan (used for error handling in the
	// goroutine that is listening for connections)
	p.closingChan <- true

	// stop GRPC server
	p.grpcServer.Stop()
	p.wg.Wait()

	// stop runner
	err := p.pluginRunner.Stop()
	if err != nil {
		controlLogger.Error(err)
	}

	// stop running plugins
	for _, rp := range p.pluginRunner.AvailablePlugins().all() {
		controlLogger.Debug("Stopping running plugin")
		rp.Stop("daemon exiting")
	}

	// unload plugins
	p.pluginManager.teardown()

	// log that we've stopped the control module
	controlLogger.WithFields(log.Fields{
		"_block": "stop",
	}).Info("control stopped")

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
	details.CertPath = rp.CertPath()
	details.KeyPath = rp.KeyPath()
	details.CACertPaths = rp.CACertPaths()
	details.TLSEnabled = rp.TLSEnabled()
	details.Uri = rp.Uri()

	if rp.Uri() != nil {
		// Is a standalone plugin
	} else if filepath.Ext(rp.Path()) == ".aci" {
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
		details.Exec = details.Manifest.App.Exec
		details.IsPackage = true
	} else {
		details.IsPackage = false
		details.Exec = []string{filepath.Base(rp.Path())}
		details.ExecPath = filepath.Dir(rp.Path())
	}

	return details, nil
}

func (p *pluginControl) Unload(pl core.Plugin) (core.CatalogedPlugin, serror.SnapError) {
	up, err := p.pluginManager.get(fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", pl.TypeName(), pl.Name(), pl.Version()))
	if err != nil {
		se := serror.New(ErrPluginNotFound, map[string]interface{}{
			"plugin-name":    pl.Name(),
			"plugin-version": pl.Version(),
			"plugin-type":    pl.TypeName(),
		})
		return nil, se
	}

	if errs := p.subscriptionGroups.validatePluginUnloading(up); errs != nil {
		impactOnTasks := []string{}
		for _, err := range errs {
			taskId := err.Fields()["task-id"].(string)
			impactOnTasks = append(impactOnTasks, taskId)
		}
		se := serror.New(errorPluginCannotBeUnloaded(impactOnTasks), map[string]interface{}{
			"plugin-name":    pl.Name(),
			"plugin-version": pl.Version(),
			"plugin-type":    pl.TypeName(),
			"impacted-tasks": impactOnTasks,
		})
		return nil, se
	}

	// unload the plugin means removing it from plugin catalog
	// and, for collector plugins, removing its metrics from metric catalog
	if _, err := p.pluginManager.UnloadPlugin(pl); err != nil {
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

func (p *pluginControl) ValidateDeps(requested []core.RequestedMetric, plugins []core.SubscribedPlugin, configTree *cdata.ConfigDataTree, asserts ...core.SubscribedPluginAssert) []serror.SnapError {
	return p.subscriptionGroups.ValidateDeps(requested, plugins, configTree, asserts...)
}

// SubscribeDeps will subscribe to collectors, processors and publishers.  The collectors are subscribed by mapping the provided
// array of core.RequestedMetrics to the corresponding plugins while processors and publishers provided in the array of core.Plugin
// will be subscribed directly.  The ID provides a logical grouping of subscriptions.
func (p *pluginControl) SubscribeDeps(id string, requested []core.RequestedMetric, plugins []core.SubscribedPlugin, configTree *cdata.ConfigDataTree) (serrs []serror.SnapError) {
	return p.subscriptionGroups.Add(id, requested, configTree, plugins)
}

// UnsubscribeDeps unsubscribes a group of dependencies provided the subscription group ID
func (p *pluginControl) UnsubscribeDeps(id string) []serror.SnapError {
	// update view and unsubscribe to plugins
	return p.subscriptionGroups.Remove(id)
}

func (p *pluginControl) verifyPlugin(lp *loadedPlugin) error {
	if lp.Details.Uri != nil {
		// remote plugin
		if core.IsUri(lp.Details.Uri.String()) {
			return fmt.Errorf(fmt.Sprintf("Remote plugin failed to load: bad uri: (%x)", lp.Details.Uri))
		}
		return nil
	}
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

// getMetricsAndCollectors returns metrics to be collected grouped by plugin and collectors which are used to collect all of them
func (p *pluginControl) getMetricsAndCollectors(requested []core.RequestedMetric, configTree *cdata.ConfigDataTree) (map[string]metricTypes, []core.SubscribedPlugin, []serror.SnapError) {
	newMetricsGroupedByPlugin := make(map[string]metricTypes)
	newPlugins := []core.SubscribedPlugin{}
	var serrs []serror.SnapError
	for _, r := range requested {
		// get all metric types available in metricCatalog which fulfill the requested namespace and version (if ver <=0 the latest version will be taken)
		newMetrics, err := p.metricCatalog.GetMetrics(r.Namespace(), r.Version())
		if err != nil {
			log.WithFields(log.Fields{
				"_block": "control",
				"action": "expanding-requested-metrics",
				"query":  r.Namespace(),
				"err":    err,
			}).Error("error matching requested namespace with metric catalog")
			serrs = append(serrs, serror.New(err))
			continue
		}

		if controlLogger.Level >= log.DebugLevel {
			for _, m := range newMetrics {
				controlLogger.WithFields(log.Fields{
					"_block":  "control",
					"ns":      m.Namespace().String(),
					"version": m.Version(),
				}).Debug("Expanded namespaces found in metric catalog")
			}
		}

		for _, mt := range newMetrics {
			// in case config tree doesn't have any configuration for current namespace
			// it's needed to initialize config, otherwise it will stay nil and panic later on
			cfg := configTree.Get(mt.Namespace().Strings())
			if cfg == nil {
				cfg = cdata.NewNode()
			}
			// set config to metric
			mt.config = cfg

			// apply the defaults from the global (plugin) config
			cfgNode := p.pluginManager.GetPluginConfig().getPluginConfigDataNode(core.CollectorPluginType, mt.Plugin.Name(), mt.Plugin.Version())
			cfg.ApplyDefaults(cfgNode.Table())

			// apply defaults to the metric that may be present in the plugins
			// configpolicy
			if pluginCfg := mt.Plugin.Policy().Get(mt.Namespace().Strings()); pluginCfg != nil {
				mt.config.ApplyDefaults(pluginCfg.Defaults())
			}

			// cataloged plugin which exposes the metric
			cp := mt.Plugin
			key := fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", cp.TypeName(), cp.Name(), cp.Version())

			// groups metricTypes by a plugin.Key()
			pmt, _ := newMetricsGroupedByPlugin[key]

			// pmt (plugin-metric-type) contains plugin and metrics types grouped to this plugin
			pmt.plugin = cp
			pmt.metricTypes = append(pmt.metricTypes, mt)
			newMetricsGroupedByPlugin[key] = pmt

			plugin := subscribedPlugin{
				name:     cp.Name(),
				typeName: cp.TypeName(),
				version:  cp.Version(),
				config:   cdata.NewNode(),
			}

			if !containsPlugin(newPlugins, plugin) {
				newPlugins = append(newPlugins, plugin)
			}
		}
	}
	if controlLogger.Level >= log.DebugLevel {
		for _, pmt := range newMetricsGroupedByPlugin {
			for _, m := range pmt.Metrics() {
				log.WithFields(log.Fields{
					"_block": "control",
					"action": "gather",
					"metric": fmt.Sprintf("%s:%d", m.Namespace().String(), m.Version()),
				}).Debug("gathered metrics from workflow request")
			}

		}
		for _, p := range newPlugins {
			log.WithFields(log.Fields{
				"_block": "control",
				"action": "gather",
				"metric": fmt.Sprintf("%s:%s:%d", p.TypeName(), p.Name(), p.Version()),
			}).Debug("gathered plugins from workflow request")
		}
	}

	return newMetricsGroupedByPlugin, newPlugins, serrs
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
	nsMap := map[string]struct{}{}
	for _, mt := range mts {
		if version > 0 {
			// a version is specified
			if mt.version == version {
				cmt = append(cmt, mt)
			}
		} else if version < 0 {
			// -1 (or less) is specified meaning return the latest
			if _, ok := nsMap[mt.Namespace().String()]; !ok {
				mt, err = p.metricCatalog.GetMetric(mt.Namespace(), version)
				if err != nil {
					return nil, err
				}
				cmt = append(cmt, mt)
				nsMap[mt.Namespace().String()] = struct{}{}
			}
		} else {
			// no version is specified return all metric versions
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

func (p *pluginControl) GetPlugins(ns core.Namespace) ([]core.CatalogedPlugin, error) {
	return p.metricCatalog.GetPlugins(ns)
}

func (p *pluginControl) MetricExists(mns core.Namespace, ver int) bool {
	_, err := p.metricCatalog.GetMetric(mns, ver)
	if err == nil {
		return true
	}
	return false
}

// CollectMetrics is a blocking call to collector plugins returning a collection
// of metrics and errors.  If an error is encountered no metrics will be
// returned.
func (p *pluginControl) CollectMetrics(id string, allTags map[string]map[string]string) (metrics []core.Metric, errs []error) {
	// If control is not started we don't want tasks to be able to
	// go through a workflow.
	if !p.Started {
		return nil, []error{ErrControllerNotStarted}
	}

	// Subscription groups are processed anytime a plugin is loaded/unloaded.
	pluginToMetricMap, serrs, err := p.subscriptionGroups.Get(id)
	if err != nil {
		controlLogger.WithFields(log.Fields{
			"_block":                "CollectorMetrics",
			"subscription-group-id": id,
		}).Error(err)
		errs = append(errs, err)
		return
	}
	// If We received errors when the requested metrics were last processed
	// against the metric catalog we need to return them to the caller.
	if serrs != nil {
		for _, e := range serrs {
			errs = append(errs, e)
		}
	}

	for ns, nsTags := range allTags {
		for k, v := range nsTags {
			log.WithFields(log.Fields{
				"_module": "control",
				"block":   "CollectMetrics",
				"type":    "pluginCollector",
				"ns":      ns,
				"tag-key": k,
				"tag-val": v,
			}).Debug("Tags in CollectMetrics")
		}
	}

	cMetrics := make(chan []core.Metric)
	cError := make(chan error)
	var wg sync.WaitGroup

	// For each available plugin call available plugin using RPC client and wait for response (goroutines)
	for pluginKey, pmt := range pluginToMetricMap {
		// merge global plugin config into the config for the metric
		for _, mt := range pmt.metricTypes {
			if mt.Config() != nil {
				mt.Config().ReverseMergeInPlace(p.Config.Plugins.getPluginConfigDataNode(core.CollectorPluginType, pmt.plugin.Name(), pmt.plugin.Version()))
			}
		}

		wg.Add(1)

		go func(pluginKey string, mt []core.Metric) {
			mts, err := p.pluginRunner.AvailablePlugins().collectMetrics(pluginKey, mt, id)
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
				m[i] = p.pluginManager.AddStandardAndWorkflowTags(m[i], allTags)
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

func (p *pluginControl) StreamMetrics(
	id string,
	allTags map[string]map[string]string,
	maxCollectDuration time.Duration,
	maxMetricsBuffer int64) (chan []core.Metric, chan error, []error) {
	if !p.Started {
		return nil, nil, []error{ErrControllerNotStarted}
	}
	errs := make([]error, 0)
	pluginToMetricMap, serrs, err := p.subscriptionGroups.Get(id)
	if err != nil {
		controlLogger.WithFields(log.Fields{
			"_block":                "StreamMetrics",
			"subscription-group-id": id,
		}).Error(err)
		errs = append(errs, err)
		return nil, nil, errs
	}

	if serrs != nil {
		for _, e := range serrs {
			errs = append(errs, e)
		}
	}
	if len(pluginToMetricMap) > 1 {
		return nil, nil, append(errs, errors.New("Only 1 streaming collecting plugin per task"))
	}
	var metricChan chan []core.Metric
	var errChan chan error
	for pluginKey, pmt := range pluginToMetricMap {
		for _, mt := range pmt.metricTypes {
			if mt.Config() != nil {
				mt.Config().ReverseMergeInPlace(
					p.Config.Plugins.getPluginConfigDataNode(
						core.CollectorPluginType,
						pmt.plugin.Name(),
						pmt.plugin.Version()))
			}
		}
		metricChan, errChan, err = p.pluginRunner.AvailablePlugins().streamMetrics(pluginKey, pmt.metricTypes, id, maxCollectDuration, maxMetricsBuffer)
		if err != nil {
			errs = append(errs, err)
			return nil, nil, errs
		}
	}
	return metricChan, errChan, nil
}

// PublishMetrics
func (p *pluginControl) PublishMetrics(metrics []core.Metric, config map[string]ctypes.ConfigValue, taskID, pluginName string, pluginVersion int) []error {
	// If control is not started we don't want tasks to be able to
	// go through a workflow.
	if !p.Started {
		return []error{ErrControllerNotStarted}
	}
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

	return p.pluginRunner.AvailablePlugins().publishMetrics(metrics, pluginName, pluginVersion, merged, taskID)
}

// ProcessMetrics
func (p *pluginControl) ProcessMetrics(metrics []core.Metric, config map[string]ctypes.ConfigValue, taskID, pluginName string, pluginVersion int) ([]core.Metric, []error) {
	// If control is not started we don't want tasks to be able to
	// go through a workflow.
	if !p.Started {
		return nil, []error{ErrControllerNotStarted}
	}
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

	return p.pluginRunner.AvailablePlugins().processMetrics(metrics, pluginName, pluginVersion, merged, taskID)
}

func (p *pluginControl) SetAutodiscoverPaths(paths []string) {
	p.autodiscoverPaths = paths
}

func (p *pluginControl) GetAutodiscoverPaths() []string {
	return p.autodiscoverPaths
}

func (p *pluginControl) GetTempDir() string {
	return p.Config.TempDirPath
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
	plugin      core.CatalogedPlugin
	metricTypes []core.Metric
}

func (mts metricTypes) Count() int {
	return len(mts.metricTypes)
}

func (mts metricTypes) Metrics() []core.Metric {
	return mts.metricTypes
}

func (mts metricTypes) Plugin() core.CatalogedPlugin {
	return mts.plugin
}

func containsPlugin(slice []core.SubscribedPlugin, lookup subscribedPlugin) bool {
	for _, plugin := range slice {
		if plugin.Name() == lookup.Name() &&
			plugin.Version() == lookup.Version() &&
			plugin.TypeName() == lookup.TypeName() {
			return true
		}
	}
	return false
}
