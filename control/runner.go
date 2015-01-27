package control

import (
	"errors"
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

var availablePlugins []*availablePlugin

// Handles events pertaining to plugins and control the runnning state accordingly.
type Runner struct {
	delegates []gomit.Delegator
	monitor   *monitor
}

// Representing a plugin running and available to execute calls against.
type availablePlugin struct {
	State              availablePluginState
	Response           *plugin.Response
	client             client.PluginClient
	eventManager       *gomit.EventController
	failedHealthChecks int
}

func newAvailablePlugin() *availablePlugin {
	ap := new(availablePlugin)
	ap.eventManager = new(gomit.EventController)
	return ap
}

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

func (ap *availablePlugin) checkHealth() {
	hc := make(chan error, 1)
	go func() {
		hc <- ap.client.Ping()
	}()
	select {
	case err := <-hc:
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

	//Create an instance of the monitor add it the runner and start it
	monitor := newMonitor()
	r.monitor = monitor
	r.monitor.Start()

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

	availablePlugins = append(availablePlugins, ap)

	return ap, nil
}

// Halt a RunnablePlugin
func (r *Runner) stopPlugin() {}

// Empty handler acting as placeholder until implementation. This helps tests
// pass to ensure registration works.
func (r *Runner) HandleGomitEvent(e gomit.Event) {
	// to do
}
