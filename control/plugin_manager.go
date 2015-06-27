// PluginManger manages loading, unloading, and swapping
// of plugins
package control

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/client"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
)

const (
	// loadedPlugin States
	DetectedState pluginState = "detected"
	LoadingState  pluginState = "loading"
	LoadedState   pluginState = "loaded"
	UnloadedState pluginState = "unloaded"
)

var (
	ErrPluginNotFound         = errors.New("plugin not found (has it already been unloaded?)")
	ErrPluginAlreadyLoaded    = errors.New("plugin is already loaded")
	ErrPluginNotInLoadedState = errors.New("Plugin must be in a LoadedState")

	pmLogger = log.WithField("_module", "control-plugin-mgr")
)

type pluginState string

type loadedPlugins struct {
	table       *[]*loadedPlugin
	mutex       *sync.Mutex
	currentIter int
}

func newLoadedPlugins() *loadedPlugins {
	var t []*loadedPlugin
	return &loadedPlugins{
		table:       &t,
		mutex:       new(sync.Mutex),
		currentIter: 0,
	}
}

// adds a loadedPlugin pointer to the loadedPlugins table
func (l *loadedPlugins) Append(lp *loadedPlugin) perror.PulseError {

	l.Lock()
	defer l.Unlock()

	// make sure we don't already have this plugin in the table
	for i, pl := range *l.table {
		if lp.Meta.Name == pl.Meta.Name && lp.Meta.Version == pl.Meta.Version {
			pe := perror.New(ErrPluginAlreadyLoaded)
			f := map[string]interface{}{
				"plugin-name":         lp.Name(),
				"plugin-version":      lp.Version(),
				"loaded-plugin-index": i,
				"error":               pe,
			}
			pe.SetFields(f)
			pmLogger.WithFields(log.Fields{
				"_block": "load-plugin",
			}).WithFields(f).Warning(pe.Error())
			return pe
		}
	}

	// append
	newLoadedPlugins := append(*l.table, lp)
	// overwrite
	l.table = &newLoadedPlugins

	return nil
}

// Table returns a collection containing loadedPlugins
// The use of the Lock and Unlock methods is suggested with Table.
func (l *loadedPlugins) Table() []*loadedPlugin {
	return *l.table
}

// used to transactionally retrieve a loadedPlugin pointer from the table
func (l *loadedPlugins) Get(index int) (*loadedPlugin, error) {
	l.Lock()
	defer l.Unlock()

	if index > len(*l.table)-1 {
		return nil, errors.New("index out of range")
	}

	return (*l.table)[index], nil
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (l *loadedPlugins) Lock() {
	l.mutex.Lock()
}

func (l *loadedPlugins) Unlock() {
	l.mutex.Unlock()
}

/* we need an atomic read / write transaction for the splice when removing a plugin,
   as the plugin is found by its index in the table.  By having the default Splice
   method block, we protect against accidental use.  Using nonblocking requires explicit
   invocation.
*/
func (l *loadedPlugins) splice(index int) {
	if index < len(*l.table) {
		lp := append((*l.table)[:index], (*l.table)[index+1:]...)
		l.table = &lp
	}
}

// splice unsafely
func (l *loadedPlugins) NonblockingSplice(index int) {
	l.splice(index)
}

// atomic splice
func (l *loadedPlugins) Splice(index int) {

	l.Lock()
	l.splice(index)
	l.Unlock()

}

// returns the item of a certain index in the table.
// to be used when iterating over the table
func (l *loadedPlugins) Item() (int, *loadedPlugin) {
	i := l.currentIter - 1
	return i, (*l.table)[i]
}

// Returns true until the "end" of the table is reached.
// used to iterate over the table:
func (l *loadedPlugins) Next() bool {
	l.currentIter++
	if l.currentIter > len(*l.table) {
		l.currentIter = 0
		return false
	}
	return true
}

// get returns the loaded plugin matching the provided name, type and version.
// If the version provided is 0 or less the newest plugin by version will be
// returned.
func (l *loadedPlugins) get(n string, t plugin.PluginType, v int) (*loadedPlugin, error) {
	l.Lock()
	defer l.Unlock()

	pvd := make(map[int]*loadedPlugin)
	keys := make([]int, 0)

	for _, lp := range l.Table() {
		if lp.Name() == n && lp.Type == t {
			pvd[lp.Version()] = lp
			keys = append(keys, lp.Version())
		}
	}
	//return error if there are no matches
	if len(keys) == 0 {
		return nil, fmt.Errorf("There is no plugin matching {name: '%s' version: '%d' type: '%s'}.", n, v, t.String())
	}
	// a specific (greater than 0) version was provided
	if v > 0 {
		lp := pvd[v]
		if lp == nil {
			return nil, fmt.Errorf("There is no plugin matching {name: '%s' version: '%d' type: '%s'}.", n, v, t.String())
		}
		return lp, nil
	} else {
		// a version of 0 or less was provided meaning select the newest plugin
		var pv int
		for _, k := range keys {
			if k > pv {
				pv = k
			}
		}
		return pvd[pv], nil
	}
}

// the struct representing a plugin that is loaded into Pulse
type loadedPlugin struct {
	Meta             plugin.PluginMeta
	Path             string
	Type             plugin.PluginType
	State            pluginState
	Token            string
	LoadedTime       time.Time
	ConfigPolicyTree *cpolicy.ConfigPolicyTree
}

// returns plugin name
// implements the CatalogedPlugin interface
func (lp *loadedPlugin) Name() string {
	return lp.Meta.Name
}

func (l *loadedPlugin) Key() string {
	return fmt.Sprintf("%s:%d", l.Name(), l.Version())
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
	privKey       *rsa.PrivateKey
	pubKey        *rsa.PublicKey
	logPath       string
}

func newPluginManager() *pluginManager {
	p := &pluginManager{
		loadedPlugins: newLoadedPlugins(),
		logPath:       "/tmp",
	}
	return p
}

func (p *pluginManager) SetMetricCatalog(mc catalogsMetrics) {
	p.metricCatalog = mc
}

// Returns loaded plugins
func (p *pluginManager) LoadedPlugins() *loadedPlugins {
	return p.loadedPlugins
}

// Load is the method for loading a plugin and
// saving plugin into the LoadedPlugins array
func (p *pluginManager) LoadPlugin(path string, emitter gomit.Emitter) (*loadedPlugin, perror.PulseError) {
	lPlugin := new(loadedPlugin)
	lPlugin.Path = path
	lPlugin.State = DetectedState

	pmLogger.WithFields(log.Fields{
		"_block": "load-plugin",
		"path":   filepath.Base(lPlugin.Path),
	}).Info("plugin load called")
	ePlugin, err := plugin.NewExecutablePlugin(p.GenerateArgs(lPlugin.Path), lPlugin.Path)

	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, perror.New(err)
	}

	err = ePlugin.Start()
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, perror.New(err)
	}

	var resp *plugin.Response
	resp, err = ePlugin.WaitForResponse(time.Second * 3)
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, perror.New(err)
	}

	ap, err := newAvailablePlugin(resp, -1, emitter, ePlugin)
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, perror.New(err)
	}

	switch resp.Type {
	case plugin.CollectorPluginType:
		colClient := ap.Client.(client.PluginCollectorClient)

		// Get the ConfigPolicyTree and add it to the loaded plugin
		cpt, err := colClient.GetConfigPolicyTree()
		if err != nil {
			pmLogger.WithFields(log.Fields{
				"_block":      "load-plugin",
				"plugin-type": "collector",
				"error":       err.Error(),
			}).Error("error in getting config policy tree")
			return nil, perror.New(err)
		}
		lPlugin.ConfigPolicyTree = &cpt

		// Get metric types
		metricTypes, err := colClient.GetMetricTypes()
		if err != nil {
			pmLogger.WithFields(log.Fields{
				"_block":      "load-plugin",
				"plugin-type": "collector",
				"error":       err.Error(),
			}).Error("error in getting metric types")
			return nil, perror.New(err)
		}

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
					"plugin-path":      filepath.Base(lPlugin.Path),
					"metric-namespace": nmt.Namespace(),
					"metric-version":   nmt.Version(),
					"error":            err.Error(),
				}).Error("received metric with bad version")
				return nil, perror.New(err)
			}
			p.metricCatalog.AddLoadedMetricType(lPlugin, nmt)
		}

	case plugin.PublisherPluginType:
		pubClient := ap.Client.(client.PluginPublisherClient)
		cpn, err := pubClient.GetConfigPolicyNode()
		if err != nil {
			pmLogger.WithFields(log.Fields{
				"_block":      "load-plugin",
				"plugin-type": "publisher",
				"error":       err.Error(),
			}).Error("error in getting config policy node")
			return nil, perror.New(err)
		}

		cpt := cpolicy.NewTree()
		cpt.Add([]string{""}, &cpn)
		lPlugin.ConfigPolicyTree = cpt

	case plugin.ProcessorPluginType:
		procClient := ap.Client.(client.PluginProcessorClient)

		cpn, err := procClient.GetConfigPolicyNode()
		if err != nil {
			pmLogger.WithFields(log.Fields{
				"_block":      "load-plugin",
				"plugin-type": "processor",
				"error":       err.Error(),
			}).Error("error in getting config policy node")
			return nil, perror.New(err)
		}

		cpt := cpolicy.NewTree()
		cpt.Add([]string{""}, &cpn)
		lPlugin.ConfigPolicyTree = cpt

	default:
		return nil, perror.New(fmt.Errorf("Unknown plugin type '%s'", resp.Type.String()))
	}

	err = ePlugin.Kill()
	if err != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, perror.New(err)
	}

	if resp.State != plugin.PluginSuccess {
		e := fmt.Errorf("Plugin loading did not succeed: %s\n", resp.ErrorMessage)
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  e,
		}).Error("load plugin error")
		return nil, perror.New(e)
	}

	lPlugin.Meta = resp.Meta
	lPlugin.Type = resp.Type
	lPlugin.Token = resp.Token
	lPlugin.LoadedTime = time.Now()
	lPlugin.State = LoadedState

	aErr := p.LoadedPlugins().Append(lPlugin)
	if aErr != nil {
		pmLogger.WithFields(log.Fields{
			"_block": "load-plugin",
			"error":  aErr,
		}).Error("load plugin error")
		return nil, aErr
	}

	return lPlugin, nil
}

// unloads a plugin from the LoadedPlugins table
func (p *pluginManager) UnloadPlugin(pl core.Plugin) (*loadedPlugin, perror.PulseError) {

	// We hold the mutex here to safely splice out the plugin from the table.
	// Using a stale index can be slightly dangerous (unloading incorrect plugin).
	p.LoadedPlugins().Lock()
	defer p.LoadedPlugins().Unlock()

	var (
		index  int
		plugin *loadedPlugin
		found  bool
	)

	// reset the iterator
	p.LoadedPlugins().currentIter = 0

	// find it in the list
	for p.LoadedPlugins().Next() {
		if !found {
			i, lp := p.LoadedPlugins().Item()
			// plugin key is its name && version
			if pl.Name() == lp.Meta.Name && pl.Version() == lp.Meta.Version {
				index = i
				plugin = lp
				// use bool for found becase we cannot check against default type values
				// index of given plugin may be 0
				found = true
			}
		} else {
			// break out of the loop once we find the plugin we're looking for
			break
		}
	}

	if !found {
		pe := perror.New(ErrPluginNotFound)
		pe.SetFields(map[string]interface{}{
			"plugin-name":    pl.Name(),
			"plugin-version": pl.Version(),
		})
		return nil, pe
	}

	if plugin.State != LoadedState {
		pe := perror.New(ErrPluginNotInLoadedState)
		pe.SetFields(map[string]interface{}{
			"plugin-name":    plugin.Name(),
			"plugin-version": plugin.Version(),
		})
		return nil, pe
	}

	// If the plugin was loaded from os.TempDir() clean up
	if strings.Contains(plugin.Path, os.TempDir()) {
		runnerLog.WithFields(log.Fields{
			"plugin-type":    plugin.TypeName(),
			"plugin-name":    plugin.Name(),
			"plugin-version": plugin.Version(),
			"plugin-path":    plugin.Path,
		}).Debugf("Removing plugin")
		if err := os.Remove(plugin.Path); err != nil {
			runnerLog.WithFields(log.Fields{
				"plugin-type":    plugin.TypeName(),
				"plugin-name":    plugin.Name(),
				"plugin-version": plugin.Version(),
				"plugin-path":    plugin.Path,
			}).Error(err)
			pe := perror.New(err)
			pe.SetFields(map[string]interface{}{
				"plugin-type":    plugin.TypeName(),
				"plugin-name":    plugin.Name(),
				"plugin-version": plugin.Version(),
				"plugin-path":    plugin.Path,
			})
			return nil, pe
		}
	}

	// splice out the given plugin
	p.LoadedPlugins().NonblockingSplice(index)

	// Remove any metrics from the catalog if this was a collector
	if plugin.TypeName() == "collector" {
		p.metricCatalog.RmUnloadedPluginMetrics(plugin)
	}

	return plugin, nil
}

func (p *pluginManager) GenerateArgs(pluginPath string) plugin.Arg {
	pluginLog := filepath.Join(p.logPath, filepath.Base(pluginPath)) + ".log"
	return plugin.NewArg(p.pubKey, pluginLog)
}
