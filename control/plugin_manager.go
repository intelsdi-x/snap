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

// PluginManger manages loading, unloading, and swapping
// of plugins
package control

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/schema"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/client"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/perror"
)

const (
	// loadedPlugin States
	DetectedState pluginState = "detected"
	LoadingState  pluginState = "loading"
	LoadedState   pluginState = "loaded"
	UnloadedState pluginState = "unloaded"
)

var (
	ErrPluginNotFound         = errors.New("plugin not found")
	ErrPluginAlreadyLoaded    = errors.New("plugin is already loaded")
	ErrPluginNotInLoadedState = errors.New("Plugin must be in a LoadedState")

	pmLogger = log.WithField("_module", "control-plugin-mgr")
)

type pluginState string

type loadedPlugins struct {
	*sync.RWMutex

	table map[string]*loadedPlugin
}

func newLoadedPlugins() *loadedPlugins {
	return &loadedPlugins{
		RWMutex: &sync.RWMutex{},
		table:   make(map[string]*loadedPlugin),
	}
}

// add adds a loadedPlugin pointer to the table
func (l *loadedPlugins) add(lp *loadedPlugin) perror.SnapError {
	l.Lock()
	defer l.Unlock()

	if _, exists := l.table[lp.Key()]; exists {
		return perror.New(ErrPluginAlreadyLoaded, map[string]interface{}{
			"plugin-name":    lp.Meta.Name,
			"plugin-version": lp.Meta.Version,
			"plugin-type":    lp.Type.String(),
		})
	}
	l.table[lp.Key()] = lp
	return nil
}

// get retrieves a plugin from the table
func (l *loadedPlugins) get(key string) (*loadedPlugin, error) {
	l.RLock()
	defer l.RUnlock()

	lp, ok := l.table[key]
	if !ok {
		tnv := strings.Split(key, ":")
		if len(tnv) != 3 {
			return nil, ErrBadKey
		}

		v, err := strconv.Atoi(tnv[2])
		if err != nil {
			return nil, ErrBadKey
		}
		if v < 1 {
			pmLogger.Info("finding latest plugin")
			return l.findLatest(tnv[0], tnv[1])
		}
		return nil, ErrPluginNotFound
	}
	return lp, nil
}

func (l *loadedPlugins) remove(key string) {
	l.Lock()
	delete(l.table, key)
	l.Unlock()
}

func (l *loadedPlugins) findLatest(typeName, name string) (*loadedPlugin, error) {
	l.RLock()
	defer l.RUnlock()

	// quick check to see if there exists a plugin with the name/version we're looking for.
	// if not we just return ErrNotFound before we check versions.
	var latest *loadedPlugin
	for _, lp := range l.table {
		if lp.TypeName() == typeName && lp.Name() == name {
			latest = lp
			break
		}
	}

	if latest != nil {
		for _, lp := range l.table {
			if lp.TypeName() == typeName && lp.Name() == name && lp.Version() > latest.Version() {
				latest = lp
			}
		}
		return latest, nil
	}
	return nil, ErrPluginNotFound
}

// the struct representing a plugin that is loaded into snap
type pluginDetails struct {
	CheckSum  [sha256.Size]byte
	Exec      string
	ExecPath  string
	IsPackage bool
	Manifest  *schema.ImageManifest
	Path      string
	Signed    bool
	Signature []byte
}

type loadedPlugin struct {
	Meta         plugin.PluginMeta
	Details      *pluginDetails
	Type         plugin.PluginType
	State        pluginState
	Token        string
	LoadedTime   time.Time
	ConfigPolicy *cpolicy.ConfigPolicy
}

// returns plugin name
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) Name() string {
	return lp.Meta.Name
}

func (lp *loadedPlugin) PluginPath() string {
	return lp.Details.Path
}

func (l *loadedPlugin) Key() string {
	return fmt.Sprintf("%s:%s:%d", l.TypeName(), l.Name(), l.Version())
}

// returns plugin version
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) Version() int {
	return lp.Meta.Version
}

// returns plugin type as a string
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) TypeName() string {
	return lp.Type.String()
}

// returns current plugin state
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) Status() string {
	return string(lp.State)
}

// returns plugin signing as a bool
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) IsSigned() bool {
	return lp.Details.Signed
}

// returns a unix timestamp of the LoadTime of a plugin
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) LoadedTimestamp() *time.Time {
	return &lp.LoadedTime
}

// the struct representing the object responsible for
// loading and unloading plugins
type pluginManager struct {
	metricCatalog catalogsMetrics
	loadedPlugins *loadedPlugins
	logPath       string
	pluginConfig  *pluginConfig
}

func newPluginManager(opts ...pluginManagerOpt) *pluginManager {
	p := &pluginManager{
		loadedPlugins: newLoadedPlugins(),
		logPath:       "/tmp",
		pluginConfig:  newPluginConfig(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type pluginManagerOpt func(*pluginManager)

func OptSetPluginConfig(cf *pluginConfig) pluginManagerOpt {
	return func(p *pluginManager) {
		p.pluginConfig = cf
	}
}

func (p *pluginManager) SetPluginConfig(cf *pluginConfig) {
	p.pluginConfig = cf
}

func (p *pluginManager) SetMetricCatalog(mc catalogsMetrics) {
	p.metricCatalog = mc
}

// Load is the method for loading a plugin and
// saving plugin into the LoadedPlugins array
func (p *pluginManager) LoadPlugin(details *pluginDetails, emitter gomit.Emitter) (*loadedPlugin, perror.SnapError) {
	lPlugin := new(loadedPlugin)
	lPlugin.Details = details
	lPlugin.State = DetectedState

	pmLogger.WithFields(log.Fields{
		"_block": "load-plugin",
		"path":   filepath.Base(lPlugin.Details.Exec),
	}).Info("plugin load called")
	ePlugin, err := plugin.NewExecutablePlugin(p.GenerateArgs(lPlugin.Details.Exec), path.Join(lPlugin.Details.ExecPath, lPlugin.Details.Exec))

	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while creating executable plugin")
		return nil, perror.New(err)
	}

	err = ePlugin.Start()
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while starting plugin")
		return nil, perror.New(err)
	}

	var resp *plugin.Response
	resp, err = ePlugin.WaitForResponse(time.Second * 3)

	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while waiting for response from plugin")
		return nil, perror.New(err)
	}

	ap, err := newAvailablePlugin(resp, emitter, ePlugin)
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while creating available plugin")
		return nil, perror.New(err)
	}

	if resp.Meta.Unsecure {
		err = ap.client.Ping()
	} else {
		err = ap.client.SetKey()
	}
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while pinging the plugin")
		return nil, perror.New(err)
	}

	// Get the ConfigPolicy and add it to the loaded plugin
	c, ok := ap.client.(plugin.Plugin)
	if !ok {
		return nil, perror.New(errors.New("missing GetConfigPolicy function"))
	}
	cp, err := c.GetConfigPolicy()
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block":         "load-plugin",
			"plugin-type":    "collector",
			"error":          err.Error(),
			"plugin-name":    ap.Name(),
			"plugin-version": ap.Version(),
			"plugin-id":      ap.ID(),
		}).Error("error in getting config policy")
		return nil, perror.New(err)
	}
	lPlugin.ConfigPolicy = cp

	if resp.Type == plugin.CollectorPluginType {
		colClient := ap.client.(client.PluginCollectorClient)

		cfg := plugin.PluginConfigType{
			ConfigDataNode: p.pluginConfig.getPluginConfigDataNode(core.PluginType(resp.Type), resp.Meta.Name, resp.Meta.Version),
		}

		metricTypes, err := colClient.GetMetricTypes(cfg)
		if err != nil {
			pmLogger.WithFields(log.Fields{
				"_block":      "load-plugin",
				"plugin-type": "collector",
				"error":       err.Error(),
			}).Error("error in getting metric types")
			return nil, perror.New(err)
		}

		// The plugin cache client will be integrated here later

		// Add metric types to metric catalog
		for _, nmt := range metricTypes {
			// If the version is 0 default it to the plugin version
			// This honors the plugins explicit version but falls back
			// to the plugin version as default
			if nmt.Version() < 1 {
				// Since we have to override version we convert to a internal struct
				nmt = &metricType{
					namespace:          nmt.Namespace(),
					version:            resp.Meta.Version,
					lastAdvertisedTime: nmt.LastAdvertisedTime(),
					config:             nmt.Config(),
					data:               nmt.Data(),
					tags:               nmt.Tags(),
					labels:             nmt.Labels(),
				}
			}
			// We quit and throw an error on bad metric versions (<1)
			// the is a safety catch otherwise the catalog will be corrupted
			if nmt.Version() < 1 {
				err := errors.New("Bad metric version from plugin")
				pmLogger.WithFields(log.Fields{
					"_block":           "load-plugin",
					"plugin-name":      resp.Meta.Name,
					"plugin-version":   resp.Meta.Version,
					"plugin-type":      resp.Meta.Type.String(),
					"plugin-path":      filepath.Base(lPlugin.Details.ExecPath),
					"metric-namespace": nmt.Namespace(),
					"metric-version":   nmt.Version(),
					"error":            err.Error(),
				}).Error("received metric with bad version")
				return nil, perror.New(err)
			}
			p.metricCatalog.AddLoadedMetricType(lPlugin, nmt)
		}
	}

	err = ePlugin.Kill()
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error while killing plugin executable plugin")
		return nil, perror.New(err)
	}

	if resp.State != plugin.PluginSuccess {
		e := fmt.Errorf("Plugin loading did not succeed: %s\n", resp.ErrorMessage)
		pmLogger.WithFields(log.Fields{
			"_block":          "load-plugin",
			"error":           e,
			"plugin response": resp.ErrorMessage,
		}).Error("load plugin error")
		return nil, perror.New(e)
	}

	lPlugin.Meta = resp.Meta
	lPlugin.Type = resp.Type
	lPlugin.Token = resp.Token
	lPlugin.LoadedTime = time.Now()
	lPlugin.State = LoadedState

	aErr := p.loadedPlugins.add(lPlugin)
	if aErr != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  aErr,
		}).Error("load plugin error while adding loaded plugin to load plugins collection")
		return nil, aErr
	}

	return lPlugin, nil
}

// unloads a plugin from the LoadedPlugins table
func (p *pluginManager) UnloadPlugin(pl core.Plugin) (*loadedPlugin, perror.SnapError) {

	plugin, err := p.loadedPlugins.get(fmt.Sprintf("%s:%s:%d", pl.TypeName(), pl.Name(), pl.Version()))
	if err != nil {
		pe := perror.New(ErrPluginNotFound, map[string]interface{}{
			"plugin-name":    pl.Name(),
			"plugin-version": pl.Version(),
			"plugin-type":    pl.TypeName(),
		})
		return nil, pe
	}

	if plugin.State != LoadedState {
		pe := perror.New(ErrPluginNotInLoadedState, map[string]interface{}{
			"plugin-name":    plugin.Name(),
			"plugin-version": plugin.Version(),
			"plugin-type":    pl.TypeName(),
		})
		return nil, pe
	}

	// If the plugin was loaded from os.TempDir() clean up
	if strings.Contains(plugin.Details.Path, os.TempDir()) {
		pmLogger.WithFields(log.Fields{
			"plugin-type":    plugin.TypeName(),
			"plugin-name":    plugin.Name(),
			"plugin-version": plugin.Version(),
			"plugin-path":    plugin.Details.Path,
		}).Debugf("Removing plugin")
		if err := os.RemoveAll(filepath.Dir(plugin.Details.Path)); err != nil {
			pmLogger.WithFields(log.Fields{
				"plugin-type":    plugin.TypeName(),
				"plugin-name":    plugin.Name(),
				"plugin-version": plugin.Version(),
				"plugin-path":    plugin.Details.Path,
			}).Error(err)
			pe := perror.New(err)
			pe.SetFields(map[string]interface{}{
				"plugin-type":    plugin.TypeName(),
				"plugin-name":    plugin.Name(),
				"plugin-version": plugin.Version(),
				"plugin-path":    plugin.Details.Path,
			})
			return nil, pe
		}
	}

	p.loadedPlugins.remove(plugin.Key())

	// Remove any metrics from the catalog if this was a collector
	if plugin.TypeName() == "collector" {
		p.metricCatalog.RmUnloadedPluginMetrics(plugin)
	}

	return plugin, nil
}

func (p *pluginManager) GenerateArgs(pluginPath string) plugin.Arg {
	pluginLog := filepath.Join(p.logPath, filepath.Base(pluginPath)) + ".log"
	return plugin.NewArg(pluginLog)
}

func (p *pluginManager) teardown() {
	for _, lp := range p.loadedPlugins.table {
		_, err := p.UnloadPlugin(lp)
		if err != nil {
			runnerLog.WithFields(log.Fields{
				"plugin-type":    lp.TypeName(),
				"plugin-name":    lp.Name(),
				"plugin-version": lp.Version(),
				"plugin-path":    lp.Details.Path,
			}).Warn("error removing plugin in teardown:", err)
		}
	}
}

func (p *pluginManager) get(key string) (*loadedPlugin, error) {
	return p.loadedPlugins.get(key)
}

func (p *pluginManager) all() map[string]*loadedPlugin {
	p.loadedPlugins.RLock()
	defer p.loadedPlugins.RUnlock()
	return p.loadedPlugins.table
}
