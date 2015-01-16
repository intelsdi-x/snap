package control

import (
	"errors"

	"github.com/intelsdilabs/gomit"
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	HandlerRegistrationName = "control.runner"

	// availablePlugin States
	PluginRunning availablePluginState = iota - 1 // Default value (0) is Running
)

type availablePluginState int

// Handles events pertaining to plugins and control the runnning state accordingly.
type Runner struct {
	delegates []gomit.Delegator
}

// Representing a plugin running and available to execute calls against.
type availablePlugin struct {
	State availablePluginState
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

// Start and return an availablePlugin or error.
func startPlugin(p plugin.PluginExecutor) (availablePlugin, error) {
	// Start plugin in daemon mode

	// Wait for plugin response

	// Ask for metric inventory

	// build availablePlugin
	ap := availablePlugin{}
	// return availablePlugin

	return ap, nil
}

// Halt a RunnablePlugin
func stopPlugin() {}

// Empty handler acting as placeholder until implementation. This helps tests
// pass to ensure registration works.
func (r *Runner) HandleGomitEvent(e gomit.Event) {
	// to do
}
