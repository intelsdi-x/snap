package control

import (
	"errors"
	"strconv"
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
	PluginDisabled

	DefaultClientTimeout           = time.Second * 3
	DefaultHealthCheckTimeout      = time.Second * 1
	DefaultHealthCheckFailureLimit = 3
)

type availablePluginState int

//JC remove me - var availablePlugins []*availablePlugin

type availablePlugins struct {
	table       *[]*availablePlugin
	mutex       *sync.Mutex
	currentIter int
}

func newAvailablePlugins() *availablePlugins {
	var t []*availablePlugin
	return &availablePlugins{
		table:       &t,
		mutex:       new(sync.Mutex),
		currentIter: 0,
	}
}

// adds an availablePlugin pointer to the availablePlugins table
func (a *availablePlugins) Append(ap *availablePlugin) error {

	a.mutex.Lock()
	defer a.mutex.Unlock()

	// make sure we don't already  have a pointer to this plugin in the table
	for i, pa := range *a.table {
		if ap == pa {
			return errors.New("plugin instance already available at index " + strconv.Itoa(i))
		}
	}

	// append
	newAvailablePluginsTable := append(*a.table, ap)
	// overwrite
	a.table = &newAvailablePluginsTable

	return nil
}

// returns a copy of the table
func (a *availablePlugins) Table() []*availablePlugin {
	return *a.table
}

// used to transactionally retrieve a loadedPlugin pointer from the table
func (a *availablePlugins) Get(index int) (*availablePlugin, error) {
	a.Lock()
	defer a.Unlock()

	if index > len(*a.table)-1 {
		return nil, errors.New("index out of range")
	}

	return (*a.table)[index], nil
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (a *availablePlugins) Lock() {
	a.mutex.Lock()
}

func (a *availablePlugins) Unlock() {
	a.mutex.Unlock()
}

/* we need an atomic read / write transaction for the splice when removing a plugin,
   as the plugin is found by its index in the table.  By having the default Splice
   method block, we protect against accidental use.  Using nonblocking requires explicit
   invocation.
*/
func (a *availablePlugins) splice(index int) {
	ap := append((*a.table)[:index], (*a.table)[index+1:]...)
	a.table = &ap
}

// splice unsafely
func (a *availablePlugins) NonblockingSplice(index int) {
	a.splice(index)
}

// atomic splice
func (a *availablePlugins) Splice(index int) {

	a.mutex.Lock()
	a.splice(index)
	a.mutex.Unlock()

}

// Handles events pertaining to plugins and control the runnning state accordingly.
type Runner struct {
	delegates        []gomit.Delegator
	monitor          *monitor
	availablePlugins *availablePlugins
}

func NewRunner() *Runner {
	r := &Runner{
		monitor:          newMonitor(),
		availablePlugins: newAvailablePlugins(),
	}
	return r
}

// Representing a plugin running and available to execute calls against.
type availablePlugin struct {
	State              availablePluginState
	Response           *plugin.Response
	client             client.PluginClient
	eventManager       *gomit.EventController
	failedHealthChecks int
	healthChan         chan error
}

func newAvailablePlugin() *availablePlugin {
	ap := new(availablePlugin)
	ap.eventManager = new(gomit.EventController)
	ap.healthChan = make(chan error, 1)
	return ap
}

// Emits event and disables the plugin if the default limit is exceeded
func (ap *availablePlugin) healthCheckFailed() {
	ap.failedHealthChecks++
	if ap.failedHealthChecks >= DefaultHealthCheckFailureLimit {
		ap.State = PluginDisabled
		pde := &control_event.DisabledPluginEvent{
			Type: ap.Response.Type,
			Meta: ap.Response.Meta,
		}
		defer ap.eventManager.Emit(pde)
	}
	hcfe := &control_event.HealthCheckFailedEvent{
		Type: ap.Response.Type,
		Meta: ap.Response.Meta,
	}
	defer ap.eventManager.Emit(hcfe)
}

// Ping the client resetting the failedHealthCheks on success.
// On failure healthCheckFailed() is called.
func (ap *availablePlugin) checkHealth() {
	go func() {
		ap.healthChan <- ap.client.Ping()
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

// TBD
type executablePlugin interface {
	Start() error
	Kill() error
	WaitForResponse(time.Duration) (*plugin.Response, error)
}

// Adds Delegates (gomit.Delegator) for adding Runner handlers to on Start and
// unregistration on Stop.
func (r *Runner) AddDelegates(delegates ...gomit.Delegator) {
	// Append the variadic collection of gomit.RegisterHanlders to r.delegates
	r.delegates = append(r.delegates, delegates...)
}

// Begin handing events and managing available plugins
func (r *Runner) Start() error {
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
func (r *Runner) Stop() []error {
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

func (r *Runner) startPlugin(p executablePlugin) (*availablePlugin, error) {
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
	ap := newAvailablePlugin()
	ap.Response = resp

	// var pluginClient plugin.
	// Create RPC client
	switch t := resp.Type; t {
	case plugin.CollectorPluginType:
		c, e := client.NewCollectorClient(resp.ListenAddress, DefaultClientTimeout)
		ap.client = c
		if e != nil {
			return nil, errors.New("error while creating client connection: " + e.Error())
		}
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + t.String())
	}

	// Ping through client

	// Ask for metric inventory

	ap.State = PluginRunning

	r.availablePlugins.Append(ap)

	return ap, nil
}

// Halt a RunnablePlugin
func (r *Runner) stopPlugin() {}

// Empty handler acting as placeholder until implementation. This helps tests
// pass to ensure registration works.
func (r *Runner) HandleGomitEvent(e gomit.Event) {
	// to do
}
