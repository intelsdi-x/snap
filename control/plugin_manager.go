// PluginManger manages loading, unloading, and swapping
// of plugins
package control

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/client"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
)

const (
	// loadedPlugin States
	DetectedState pluginState = "detected"
	LoadingState  pluginState = "loading"
	LoadedState   pluginState = "loaded"
	UnloadedState pluginState = "unloaded"
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
func (l *loadedPlugins) Append(lp *loadedPlugin) error {

	l.Lock()
	defer l.Unlock()

	// make sure we don't already have this plugin in the table
	for i, pl := range *l.table {
		if lp.Meta.Name == pl.Meta.Name && lp.Meta.Version == pl.Meta.Version {
			return errors.New("plugin [" + lp.Meta.Name + ":" + strconv.Itoa(lp.Meta.Version) + "] already loaded at index " + strconv.Itoa(i))
		}
	}

	// append
	newLoadedPlugins := append(*l.table, lp)
	// overwrite
	l.table = &newLoadedPlugins

	return nil
}

// returns a copy of the table
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
func (lp *loadedPlugin) LoadedTimestamp() int64 {
	return lp.LoadedTime.Unix()
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
func (p *pluginManager) LoadPlugin(path string, emitter gomit.Emitter) (*loadedPlugin, error) {
	lPlugin := new(loadedPlugin)
	lPlugin.Path = path
	lPlugin.State = DetectedState

	log.WithFields(log.Fields{
		"module": "control-plugin-manager",
		"block":  "load-plugin",
		"path":   filepath.Base(lPlugin.Path),
	}).Info("plugin load called")
	ePlugin, err := plugin.NewExecutablePlugin(p.GenerateArgs(lPlugin.Path), lPlugin.Path)

	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	err = ePlugin.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	var resp *plugin.Response
	resp, err = ePlugin.WaitForResponse(time.Second * 3)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	ap, err := newAvailablePlugin(resp, -1, emitter, ePlugin)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	switch resp.Type {
	case plugin.CollectorPluginType:
		colClient := ap.Client.(client.PluginCollectorClient)

		// Get the ConfigPolicyTree and add it to the loaded plugin
		cpt, err := colClient.GetConfigPolicyTree()
		if err != nil {
			log.WithFields(log.Fields{
				"module":      "control-plugin-manager",
				"block":       "load-plugin",
				"plugin-type": "collector",
				"error":       err.Error(),
			}).Error("error in getting config policy tree")
			return nil, err
		}
		lPlugin.ConfigPolicyTree = &cpt

		// Get metric types
		metricTypes, err := colClient.GetMetricTypes()
		if err != nil {
			log.WithFields(log.Fields{
				"module":      "control-plugin-manager",
				"block":       "load-plugin",
				"plugin-type": "collector",
				"error":       err.Error(),
			}).Error("error in getting metric types")
			return nil, err
		}

		// Add metric types to metric catalog
		for _, mt := range metricTypes {
			p.metricCatalog.AddLoadedMetricType(lPlugin, mt)
		}

	case plugin.PublisherPluginType:
		pubClient := ap.Client.(client.PluginPublisherClient)
		cpn, err := pubClient.GetConfigPolicyNode()
		if err != nil {
			log.WithFields(log.Fields{
				"module":      "control-plugin-manager",
				"block":       "load-plugin",
				"plugin-type": "publisher",
				"error":       err.Error(),
			}).Error("error in getting config policy node")
			return nil, err
		}

		cpt := cpolicy.NewTree()
		cpt.Add([]string{""}, &cpn)
		lPlugin.ConfigPolicyTree = cpt

	case plugin.ProcessorPluginType:
		procClient := ap.Client.(client.PluginProcessorClient)

		cpn, err := procClient.GetConfigPolicyNode()
		if err != nil {
			log.WithFields(log.Fields{
				"module":      "control-plugin-manager",
				"block":       "load-plugin",
				"plugin-type": "processor",
				"error":       err.Error(),
			}).Error("error in getting config policy node")
			return nil, err
		}

		cpt := cpolicy.NewTree()
		cpt.Add([]string{""}, &cpn)
		lPlugin.ConfigPolicyTree = cpt

	default:
		return nil, fmt.Errorf("Unknown plugin type '%s'", resp.Type.String())
	}

	err = ePlugin.Kill()
	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	if resp.State != plugin.PluginSuccess {
		e := fmt.Errorf("Plugin loading did not succeed: %s\n", resp.ErrorMessage)
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  e,
		}).Error("load plugin error")
		return nil, e
	}

	lPlugin.Meta = resp.Meta
	lPlugin.Type = resp.Type
	lPlugin.Token = resp.Token
	lPlugin.LoadedTime = time.Now()
	lPlugin.State = LoadedState

	err = p.LoadedPlugins().Append(lPlugin)
	if err != nil {
		log.WithFields(log.Fields{
			"module": "control-plugin-manager",
			"block":  "load-plugin",
			"error":  err.Error(),
		}).Error("load plugin error")
		return nil, err
	}

	return lPlugin, nil
}

// unloads a plugin from the LoadedPlugins table
func (p *pluginManager) UnloadPlugin(pl core.CatalogedPlugin) error {

	// We hold the mutex here to safely splice out the plugin from the table.
	// Using a stale index can be slightly dangerous (unloading incorrect plugin).
	p.LoadedPlugins().Lock()
	defer p.LoadedPlugins().Unlock()

	var (
		index  int
		plugin *loadedPlugin
		found  bool
	)

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
		return errors.New("plugin [" + pl.Name() + "] -- [" + strconv.Itoa(pl.Version()) + "] not found (has it already been unloaded?)")
	}

	if plugin.State != LoadedState {
		return errors.New("Plugin must be in a LoadedState")
	}

	// splice out the given plugin
	p.LoadedPlugins().NonblockingSplice(index)

	return nil
}

func (p *pluginManager) GenerateArgs(pluginPath string) plugin.Arg {
	pluginLog := filepath.Join(p.logPath, filepath.Base(pluginPath)) + ".log"
	return plugin.NewArg(p.pubKey, pluginLog)
}
