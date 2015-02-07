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
	HandlerRegistrationName = "control.runner"

	// availablePlugin States
	PluginRunning availablePluginState = iota - 1 // Default value (0) is Running
	PluginStopped
	PluginDisabled

	DefaultClientTimeout           = time.Second * 3
	DefaultHealthCheckTimeout      = time.Second * 1
	DefaultHealthCheckFailureLimit = 3
)

type availablePluginState int

type availablePlugins struct {
	Collectors, Publishers, Processors *apCollection
}

func newAvailablePlugins() *availablePlugins {
	return &availablePlugins{
		Collectors: newAPCollection(),
		Processors: newAPCollection(),

		Publishers: newAPCollection(),
	}
}

func (a *availablePlugins) Insert(ap *availablePlugin) error {
	switch ap.Type {
	case plugin.CollectorPluginType:
		a.Collectors.Add(ap)
	case plugin.PublisherPluginType:
		a.Publishers.Add(ap)
	case plugin.ProcessorPluginType:
		a.Processors.Add(ap)
	default:
		return errors.New("cannot insert into available plugins, unknown plugin type")
	}
	return nil
}

func (a *availablePlugins) Remove(ap *availablePlugin) error {
	switch ap.Type {
	case plugin.CollectorPluginType:
		a.Collectors.Remove(ap)
	case plugin.PublisherPluginType:
		a.Publishers.Remove(ap)
	case plugin.ProcessorPluginType:
		a.Processors.Remove(ap)
	default:
		return errors.New("cannot remove from available plugins, unknown plugin type")
	}
	return nil
}

type apCollection struct {
	table       *map[string]*[]*availablePlugin
	mutex       *sync.Mutex
	keys        *[]string
	currentIter int
}

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

// adds an availablePlugin pointer to the collection
func (c *apCollection) Add(ap *availablePlugin) error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := (*c.table)[ap.Key]; !ok {
		*c.keys = append(*c.keys, ap.Key)
	}

	if (*c.table)[ap.Key] != nil {
		// make sure we don't already  have a pointer to this plugin in the table
		for i, pa := range *(*c.table)[ap.Key] {
			if ap == pa {
				return errors.New("plugin instance already available at index " + strconv.Itoa(i))
			}
		}
	} else {
		var col []*availablePlugin
		(*c.table)[ap.Key] = &col
	}

	// tell ap its index in the table
	ap.Index = len(*(*c.table)[ap.Key])

	// append
	newcollection := append(*(*c.table)[ap.Key], ap)

	// overwrite
	(*c.table)[ap.Key] = &newcollection

	return nil
}

// returns a copy of the table
func (c *apCollection) Table() map[string][]*availablePlugin {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	m := make(map[string][]*availablePlugin)
	for k, v := range *c.table {
		m[k] = *v
	}
	return m
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (c *apCollection) Lock() {
	c.mutex.Lock()
}

func (c *apCollection) Unlock() {
	c.mutex.Unlock()
}

func (c *apCollection) Remove(ap *availablePlugin) {

	c.mutex.Lock()
	splicedcoll := append((*(*c.table)[ap.Key])[:ap.Index], (*(*c.table)[ap.Key])[ap.Index+1:]...)
	(*c.table)[ap.Key] = &splicedcoll
	c.mutex.Unlock()

}

func (c *apCollection) Values() (string, *[]*availablePlugin) {
	key := (*c.keys)[c.currentIter-1]
	return key, (*c.table)[key]
}

func (c *apCollection) Next() bool {
	c.currentIter++
	if c.currentIter > len(*c.table) {
		c.currentIter = 0
		return false
	}
	return true
}

// Handles events pertaining to plugins and control the runnning state accordingly.
type runner struct {
	delegates        []gomit.Delegator
	monitor          *monitor
	availablePlugins *availablePlugins
}

func newRunner() *runner {
	r := &runner{
		monitor:          newMonitor(-1),
		availablePlugins: newAvailablePlugins(),
	}
	return r
}

// Representing a plugin running and available to execute calls against.
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

func newAvailablePlugin(resp *plugin.Response) *availablePlugin {
	ap := &availablePlugin{
		Name:    resp.Meta.Name,
		Version: resp.Meta.Version,
		Type:    resp.Type,

		eventManager: new(gomit.EventController),
		healthChan:   make(chan error, 1),
	}
	ap.makeKey()
	return ap
}

// Emits event and disables the plugin if the default limit is exceeded
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
		// TODO: this event should be handled by runner to
		// remove stop the ap, and remove it from the collection
		defer ap.eventManager.Emit(pde)
	}
	hcfe := &control_event.HealthCheckFailedEvent{
		Name:    ap.Name,
		Version: ap.Version,
		Type:    ap.Type,
	}
	defer ap.eventManager.Emit(hcfe)
}

// Ping the client resetting the failedHealthCheks on success.
// On failure healthCheckFailed() is called.
func (ap *availablePlugin) CheckHealth() {
	go func() {
		ap.healthChan <- ap.Client.Ping()
	}()
	select {
	case err := <-ap.healthChan:
		if err == nil {
			//if res is ok - do nothing
			ap.failedHealthChecks = 0
		} else {
			ap.healthCheckFailed()
		}
	case <-time.After(time.Second * 1):
		ap.healthCheckFailed()
	}
}

// Halt an availablePlugin
func (ap *availablePlugin) Stop(r string) error {
	return ap.Client.Kill(r)
}

func (ap *availablePlugin) makeKey() {
	s := []string{ap.Name, strconv.Itoa(ap.Version)}
	ap.Key = strings.Join(s, ":")
}

// TBD
type executablePlugin interface {
	Start() error
	Kill() error
	WaitForResponse(time.Duration) (*plugin.Response, error)
}

// Adds Delegates (gomit.Delegator) for adding Runner handlers to on Start and
// unregistration on Stop.
func (r *runner) AddDelegates(delegates ...gomit.Delegator) {
	// Append the variadic collection of gomit.RegisterHanlders to r.delegates
	r.delegates = append(r.delegates, delegates...)
}

// Begin handing events and managing available plugins
func (r *runner) Start() error {
	// Delegates must be added before starting if none exist
	// then this Runner can do nothing and should not start.
	if len(r.delegates) == 0 {
		return errors.New("No delegates added before called Start()")
	}

	// For each delegate register needed handlers
	for _, del := range r.delegates {
		e := del.RegisterHandler(HandlerRegistrationName, r)
		if e != nil {
			return e
		}
	}

	// Start the monitor
	r.monitor.Start(r.availablePlugins)

	return nil
}

// Stop handling, gracefully stop all plugins.
func (r *runner) Stop() []error {
	var errs []error
	// For each delegate unregister needed handlers
	for _, del := range r.delegates {
		e := del.UnregisterHandler(HandlerRegistrationName)
		if e != nil {
			errs = append(errs, e)
		}
	}
	return errs
}

func (r *runner) startPlugin(p executablePlugin) (*availablePlugin, error) {
	// Start plugin in daemon mode
	e := p.Start()
	if e != nil {
		return nil, errors.New("error while starting plugin: " + e.Error())
	}

	// Wait for plugin response
	resp, err := p.WaitForResponse(time.Second * 3)
	if err != nil {
		return nil, errors.New("error while waiting for response: " + err.Error())
	}

	if resp == nil {
		return nil, errors.New("no reponse object returned from plugin")
	}

	if resp.State != plugin.PluginSuccess {
		return nil, errors.New("plugin could not start error: " + resp.ErrorMessage)
	}

	// build availablePlugin
	ap := newAvailablePlugin(resp)

	// var pluginClient plugin.
	// Create RPC client
	switch resp.Type {
	case plugin.CollectorPluginType:
		c, e := client.NewCollectorClient(resp.ListenAddress, DefaultClientTimeout)
		ap.Client = c
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + resp.Type.String())
	}

	// Ping through client
	err = ap.Client.Ping()
	if err != nil {
		return nil, err
	}

	r.availablePlugins.Insert(ap)

	return ap, nil
}

func (r *runner) stopPlugin(reason string, ap *availablePlugin) error {
	err := ap.Client.Kill(reason)
	if err != nil {
		return err
	}
	err = r.availablePlugins.Remove(ap)
	if err != nil {
		return err
	}
	return nil
}

// Empty handler acting as placeholder until implementation. This helps tests
// pass to ensure registration works.
func (r *runner) HandleGomitEvent(e gomit.Event) {
	// to do
}
