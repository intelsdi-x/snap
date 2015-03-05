package control

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intelsdilabs/gomit"
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/client"
	"github.com/intelsdilabs/pulse/core/control_event"
)

const (
	DefaultClientTimeout           = time.Second * 3
	DefaultHealthCheckTimeout      = time.Second * 1
	DefaultHealthCheckFailureLimit = 3
)

type availablePluginState int

// availablePlugin represents a plugin running and available to execute calls against
type availablePlugin struct {
	Name    string
	Key     string
	Type    plugin.PluginType
	Version int
	Client  client.PluginClient
	Index   int

	eventManager       *gomit.EventController
	failedHealthChecks int
	healthChan         chan error
}

// newAvailablePlugin returns an availablePlugin with information from a
// plugin.Response
func newAvailablePlugin(resp *plugin.Response) (*availablePlugin, error) {
	ap := &availablePlugin{
		Name:    resp.Meta.Name,
		Version: resp.Meta.Version,
		Type:    resp.Type,

		eventManager: new(gomit.EventController),
		healthChan:   make(chan error, 1),
	}

	// Create RPC Client
	switch resp.Type {
	case plugin.CollectorPluginType:
		c, e := client.NewCollectorNativeClient(resp.ListenAddress, DefaultClientTimeout)
		ap.Client = c
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + resp.Type.String())
	}

	ap.makeKey()
	return ap, nil
}

// Stop halts a running availablePlugin
func (ap *availablePlugin) Stop(r string) error {
	return ap.Client.Kill(r)
}

// CheckHealth checks the health of a plugin and updates
// ap.failedHealthChecks
func (ap *availablePlugin) CheckHealth() {
	go func() {
		ap.healthChan <- ap.Client.Ping()
	}()
	select {
	case err := <-ap.healthChan:
		if err == nil {
			ap.failedHealthChecks = 0
		} else {
			ap.healthCheckFailed()
		}
	case <-time.After(time.Second * 1):
		ap.healthCheckFailed()
	}
}

// healthCheckFailed increments ap.failedHealthChecks and emits a DisabledPluginEvent
// and a HealthCheckFailedEvent
func (ap *availablePlugin) healthCheckFailed() {
	ap.failedHealthChecks++
	if ap.failedHealthChecks >= DefaultHealthCheckFailureLimit {
		pde := &control_event.DisabledPluginEvent{
			Name:    ap.Name,
			Version: ap.Version,
			Type:    ap.Type,
			Key:     ap.Key,
			Index:   ap.Index,
		}
		defer ap.eventManager.Emit(pde)
	}
	hcfe := &control_event.HealthCheckFailedEvent{
		Name:    ap.Name,
		Version: ap.Version,
		Type:    ap.Type,
	}
	defer ap.eventManager.Emit(hcfe)
}

// makeKey creates the ap.Key from the ap.Name and ap.Version
func (ap *availablePlugin) makeKey() {
	s := []string{ap.Name, strconv.Itoa(ap.Version)}
	ap.Key = strings.Join(s, ":")
}

// apCollection is a collection of availablePlugin
type apCollection struct {
	table       *map[string]*[]*availablePlugin
	mutex       *sync.Mutex
	keys        *[]string
	currentIter int
}

// newAPCollection returns an apCollection capable of storing availblePlugin
func newAPCollection() *apCollection {
	m := make(map[string]*[]*availablePlugin)
	var k []string
	return &apCollection{
		table:       &m,
		mutex:       &sync.Mutex{},
		keys:        &k,
		currentIter: 0,
	}
}

// Table returns a copy of the apCollection table
func (c *apCollection) Table() map[string][]*availablePlugin {
	c.Lock()
	defer c.Unlock()

	m := make(map[string][]*availablePlugin)
	for k, v := range *c.table {
		m[k] = *v
	}
	return m
}

// Add adds an availablePlugin to the apCollection table
func (c *apCollection) Add(ap *availablePlugin) error {
	c.Lock()
	defer c.Unlock()

	if _, ok := (*c.table)[ap.Key]; !ok {
		*c.keys = append(*c.keys, ap.Key)
	}

	if (*c.table)[ap.Key] != nil {
		// make sure we don't already have a pointer to this plugin in the table
		if exist, i := c.exists(ap); exist {
			return errors.New("plugin instance already available at index " + strconv.Itoa(i))
		}
	} else {
		var col []*availablePlugin
		(*c.table)[ap.Key] = &col
	}

	// tell ap its index in the table
	ap.Index = len(*(*c.table)[ap.Key])

	// append
	newCollection := append(*(*c.table)[ap.Key], ap)

	// overwrite
	(*c.table)[ap.Key] = &newCollection

	return nil
}

// Remove removes an availablePlugin from the apCollection table
func (c *apCollection) Remove(ap *availablePlugin) error {
	c.Lock()
	defer c.Unlock()
	if exists, _ := c.exists(ap); !exists {
		return errors.New("Warning: plugin does not exist in table")
	}
	splicedColl := append((*(*c.table)[ap.Key])[:ap.Index], (*(*c.table)[ap.Key])[ap.Index+1:]...)
	(*c.table)[ap.Key] = &splicedColl
	//reset indexes
	for i, ap := range *(*c.table)[ap.Key] {
		ap.Index = i
	}
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
func (c *apCollection) Item() (string, *[]*availablePlugin) {
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
func (c *apCollection) exists(ap *availablePlugin) (bool, int) {
	for i, _ap := range *(*c.table)[ap.Key] {
		if ap == _ap {
			return true, i
		}
	}
	return false, -1
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
