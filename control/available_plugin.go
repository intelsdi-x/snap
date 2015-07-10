package control

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/client"
	"github.com/intelsdi-x/pulse/control/routing"
	"github.com/intelsdi-x/pulse/core/control_event"
)

const (
	// DefaultClientTimeout - default timeout for a client connection attempt
	DefaultClientTimeout = time.Second * 3
	// DefaultHealthCheckTimeout - default timeout for a health check
	DefaultHealthCheckTimeout = time.Second * 1
	// DefaultHealthCheckFailureLimit - how any consecutive health check timeouts must occur to trigger a failure
	DefaultHealthCheckFailureLimit = 3
)

type availablePluginState int

// availablePlugin represents a plugin running and available to execute calls against
type availablePlugin struct {
	Key    string
	Type   plugin.PluginType
	Client client.PluginClient
	Index  int

	name               string
	version            int
	id                 int
	hitCount           int
	lastHitTime        time.Time
	emitter            gomit.Emitter
	failedHealthChecks int
	healthChan         chan error
	ePlugin            executablePlugin
}

// newAvailablePlugin returns an availablePlugin with information from a
// plugin.Response
func newAvailablePlugin(resp *plugin.Response, id int, emitter gomit.Emitter, ep executablePlugin) (*availablePlugin, error) {
	ap := &availablePlugin{
		name:    resp.Meta.Name,
		version: resp.Meta.Version,
		Type:    resp.Type,

		emitter:     emitter,
		healthChan:  make(chan error, 1),
		lastHitTime: time.Now(),
		id:          id,
		ePlugin:     ep,
	}

	// Create RPC Client
	switch resp.Type {
	case plugin.CollectorPluginType:
		c, e := client.NewCollectorNativeClient(resp.ListenAddress, DefaultClientTimeout)
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
		ap.Client = c
	case plugin.PublisherPluginType:
		c, e := client.NewPublisherNativeClient(resp.ListenAddress, DefaultClientTimeout)
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
		ap.Client = c
	case plugin.ProcessorPluginType:
		c, e := client.NewProcessorNativeClient(resp.ListenAddress, DefaultClientTimeout)
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
		ap.Client = c
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + resp.Type.String())
	}

	ap.makeKey()
	return ap, nil
}

func (a *availablePlugin) ID() int {
	return a.id
}

func (a *availablePlugin) String() string {
	return fmt.Sprintf("%s:v%d:id%d", a.name, a.version, a.id)
}

func (a *availablePlugin) TypeName() string {
	return a.Type.String()
}

func (a *availablePlugin) Name() string {
	return a.name
}

func (a *availablePlugin) Version() int {
	return a.version
}

// Stop halts a running availablePlugin
func (a *availablePlugin) Stop(r string) error {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "stop",
		"aplugin": a,
	}).Info("stopping available plugin")
	return a.Client.Kill(r)
}

// Kill assumes aplugin is not able to here a Kill RPC call
func (a *availablePlugin) Kill(r string) error {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "kill",
		"aplugin": a,
	}).Info("hard killing available plugin")
	return a.ePlugin.Kill()
}

// CheckHealth checks the health of a plugin and updates
// a.failedHealthChecks
func (a *availablePlugin) CheckHealth() {
	go func() {
		a.healthChan <- a.Client.Ping()
	}()
	select {
	case err := <-a.healthChan:
		if err == nil {
			if a.failedHealthChecks > 0 {
				// only log on first ok health check
				log.WithFields(log.Fields{
					"_module": "control-aplugin",
					"block":   "check-health",
					"aplugin": a,
				}).Debug("health is ok")
			}
			a.failedHealthChecks = 0
		} else {
			a.healthCheckFailed()
		}
	case <-time.After(time.Second * 1):
		a.healthCheckFailed()
	}
}

// healthCheckFailed increments a.failedHealthChecks and emits a DisabledPluginEvent
// and a HealthCheckFailedEvent
func (a *availablePlugin) healthCheckFailed() {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "check-health",
		"aplugin": a,
	}).Warning("heartbeat missed")
	a.failedHealthChecks++
	if a.failedHealthChecks >= DefaultHealthCheckFailureLimit {
		log.WithFields(log.Fields{
			"_module": "control-aplugin",
			"block":   "check-health",
			"aplugin": a,
		}).Warning("heartbeat failed")
		pde := &control_event.DeadAvailablePluginEvent{
			Name:    a.name,
			Version: a.version,
			Type:    int(a.Type),
			Key:     a.Key,
			Id:      a.ID(),
			String:  a.String(),
		}
		defer a.emitter.Emit(pde)
	}
	hcfe := &control_event.HealthCheckFailedEvent{
		Name:    a.name,
		Version: a.version,
		Type:    int(a.Type),
	}
	defer a.emitter.Emit(hcfe)
}

func (a *availablePlugin) HitCount() int {
	return a.hitCount
}

func (a *availablePlugin) LastHit() time.Time {
	return a.lastHitTime
}

// makeKey creates the a.Key from the a.Name and a.Version
func (a *availablePlugin) makeKey() {
	s := []string{a.name, strconv.Itoa(a.version)}
	a.Key = strings.Join(s, ":")
}

// apCollection is a collection of availablePlugin
type apCollection struct {
	table       *map[string]*availablePluginPool
	mutex       *sync.Mutex
	keys        *[]string
	currentIter int
}

// newAPCollection returns an apCollection capable of storing availblePlugin
func newAPCollection() *apCollection {
	m := make(map[string]*availablePluginPool)
	var k []string
	return &apCollection{
		table:       &m,
		mutex:       &sync.Mutex{},
		keys:        &k,
		currentIter: 0,
	}
}

func (c *apCollection) GetPluginPool(key string) *availablePluginPool {
	c.Lock()
	defer c.Unlock()

	if ap, ok := (*c.table)[key]; ok {
		return ap
	}
	return nil
}

func (c *apCollection) PluginPoolHasAP(key string) bool {
	a := c.GetPluginPool(key)
	if a != nil && a.Count() > 0 {
		return true
	}
	return false
}

// Table returns a copy of the apCollection table
func (c *apCollection) Table() map[string][]*availablePlugin {
	c.Lock()
	defer c.Unlock()

	m := make(map[string][]*availablePlugin)
	for k, v := range *c.table {
		m[k] = *v.Plugins
	}
	return m
}

// Add adds an availablePlugin to the apCollection table
func (c *apCollection) Add(ap *availablePlugin) error {
	log.WithFields(log.Fields{
		"_module":        "control-aplugin",
		"block":          "apcollection",
		"plugin-name":    ap.Name(),
		"plugin-version": ap.Version(),
	}).Debug("available plugin added")
	c.Lock()
	defer c.Unlock()

	if _, ok := (*c.table)[ap.Key]; !ok {
		*c.keys = append(*c.keys, ap.Key)
	}

	if (*c.table)[ap.Key] != nil {
		// make sure we don't already have a pointer to this plugin in the table
		if exist, i := c.Exists(ap); exist {
			return errors.New("plugin instance already available at index " + strconv.Itoa(i))
		}
	} else {
		(*c.table)[ap.Key] = newAvailablePluginPool()
	}

	(*c.table)[ap.Key].Add(ap)
	return nil
}

// Remove removes an availablePlugin from the apCollection table
func (c *apCollection) Remove(ap *availablePlugin) error {
	c.Lock()
	defer c.Unlock()

	if exists, _ := c.Exists(ap); !exists {
		return errors.New("Warning: plugin does not exist in table")
	}

	(*c.table)[ap.Key].Remove(ap)
	log.WithFields(log.Fields{
		"_module":        "control-aplugin",
		"block":          "apcollection",
		"plugin-name":    ap.Name(),
		"plugin-version": ap.Version(),
	}).Debug("available plugin removed")
	return nil
}

// Lock locks the mutex and is exported for external operations that may be unsafe
func (c *apCollection) Lock() {
	c.mutex.Lock()
}

// Unlock unlocks the mutex
func (c *apCollection) Unlock() {
	c.mutex.Unlock()
}

// Item returns the item at current position in the apCollection table
func (c *apCollection) Item() (string, *availablePluginPool) {
	key := (*c.keys)[c.currentIter-1]
	return key, (*c.table)[key]
}

// Next moves iteration position in the apCollection table
func (c *apCollection) Next() bool {
	c.currentIter++
	if c.currentIter > len(*c.table) {
		c.currentIter = 0
		return false
	}
	return true
}

// exists checks the table to see if a pointer for the availablePlugin specified
// already exists
func (c *apCollection) Exists(ap *availablePlugin) (bool, int) {
	return (*c.table)[ap.Key].Exists(ap)
}

// availablePlugins is a collection of availablePlugins by type
type availablePlugins struct {
	Collectors, Publishers, Processors *apCollection
}

// newAvailablePlugins returns an availablePlugins pointer
func newAvailablePlugins() *availablePlugins {
	return &availablePlugins{
		Collectors: newAPCollection(),
		Processors: newAPCollection(),
		Publishers: newAPCollection(),
	}
}

// Insert adds an availablePlugin into the correct collection based on type
func (a *availablePlugins) Insert(ap *availablePlugin) error {
	switch ap.Type {
	case plugin.CollectorPluginType:
		err := a.Collectors.Add(ap)
		return err
	case plugin.PublisherPluginType:
		err := a.Publishers.Add(ap)
		return err
	case plugin.ProcessorPluginType:
		err := a.Processors.Add(ap)
		return err
	default:
		return errors.New("cannot insert into available plugins, unknown plugin type")
	}
}

// Remove removes an availablePlugin from the correct collection based on type
func (a *availablePlugins) Remove(ap *availablePlugin) error {
	switch ap.Type {
	case plugin.CollectorPluginType:
		err := a.Collectors.Remove(ap)
		return err
	case plugin.PublisherPluginType:
		err := a.Publishers.Remove(ap)
		return err
	case plugin.ProcessorPluginType:
		err := a.Processors.Remove(ap)
		return err
	default:
		return errors.New("cannot remove from available plugins, unknown plugin type")
	}
}

type availablePluginPool struct {
	Plugins       *[]*availablePlugin
	subscriptions int
	mutex         *sync.Mutex
}

func newAvailablePluginPool() *availablePluginPool {
	var app []*availablePlugin
	return &availablePluginPool{
		Plugins: &app,
		mutex:   &sync.Mutex{},
	}
}

func (a *availablePluginPool) Lock() {
	a.mutex.Lock()
}

func (a *availablePluginPool) Unlock() {
	a.mutex.Unlock()
}

func (a *availablePluginPool) Count() int {
	return len((*a.Plugins))
}

func (a *availablePluginPool) Subscribe() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.subscriptions++
	log.WithFields(log.Fields{
		"_module":            "control-avaialble-plugin-pool",
		"_block":             "subscribe",
		"subscription count": a.subscriptions,
	})
}

func (a *availablePluginPool) Unsubscribe() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.subscriptions--
	log.WithFields(log.Fields{
		"_module":            "control-avaialble-plugin-pool",
		"_block":             "unsubscribe",
		"subscription count": a.subscriptions,
	})
}

func (a *availablePluginPool) Add(ap *availablePlugin) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	// tell ap its index in the table
	ap.Index = len((*a.Plugins))
	// append
	newCollection := append((*a.Plugins), ap)
	// overwrite
	a.Plugins = &newCollection
	log.WithFields(
		log.Fields{
			"_module": "control-aplugin",
			"block":   "aplugin-pool-add",
			"aplugin": ap.String(),
		}).Info("added aplugin to pool")
}

func (a *availablePluginPool) Remove(ap *availablePlugin) {
	a.Lock()
	defer a.Unlock()
	// Place nil here to allow GC per : https://github.com/golang/go/wiki/SliceTricks
	(*a.Plugins)[ap.Index] = nil
	splicedColl := append((*a.Plugins)[:ap.Index], (*a.Plugins)[ap.Index+1:]...)
	a.Plugins = &splicedColl
	//reset indexes
	a.resetIndexes()
	log.WithFields(
		log.Fields{
			"_module": "control-aplugin",
			"block":   "aplugin-pool-remove",
			"aplugin": ap.String(),
		}).Info("removed aplugin from pool")
}

// Calls Kill on a aplugin if it exists in the pool and returns a pointer to the killed plugin
func (a *availablePluginPool) Kill(k, r string) (*availablePlugin, error) {
	a.Lock()
	defer a.Unlock()
	for _, ap := range *a.Plugins {
		if k == ap.String() {
			err := ap.Kill(r)
			log.WithFields(
				log.Fields{
					"_module": "control-aplugin",
					"block":   "aplugin-pool-kill",
					"aplugin": ap.String(),
				}).Info("killing aplugin")
			return ap, err
		}
	}
	return nil, nil
}

func (a *availablePluginPool) Exists(ap *availablePlugin) (bool, int) {
	for i, _ap := range *a.Plugins {
		if ap == _ap {
			return true, i
		}
	}
	return false, -1
}

func (a *availablePluginPool) resetIndexes() {
	for i, ap := range *a.Plugins {
		ap.Index = i
	}
}

func (a *availablePluginPool) SelectUsingStrategy(strat RoutingStrategy) (*availablePlugin, error) {
	a.Lock()
	defer a.Unlock()

	sp := make([]routing.SelectablePlugin, len(*a.Plugins))
	for i := range *a.Plugins {
		sp[i] = (*a.Plugins)[i]
	}
	sap, err := strat.Select(a, sp)
	if err != nil || sap == nil {
		return nil, err
	}
	return sap.(*availablePlugin), err
}
