package control

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"strings"

	"github.com/intelsdilabs/gomit"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/core/control_event"
)

const (
	HandlerRegistrationName = "control.runner"

	// availablePlugin States
	PluginRunning availablePluginState = iota - 1 // Default value (0) is Running
	PluginStopped
	PluginDisabled

	// Until more advanced decisioning on starting exists this is the max number to spawn.
	MaximumRunningPlugins = 3
)

// TBD
type executablePlugin interface {
	Start() error
	Kill() error
	WaitForResponse(time.Duration) (*plugin.Response, error)
}

// Handles events pertaining to plugins and control the runnning state accordingly.
type runner struct {
	delegates        []gomit.Delegator
	monitor          *monitor
	availablePlugins *availablePlugins
	metricCatalog    catalogsMetrics
	pluginManager    managesPlugins
	mutex            *sync.Mutex
}

func newRunner() *runner {
	r := &runner{
		monitor:          newMonitor(),
		availablePlugins: newAvailablePlugins(),
		mutex:            &sync.Mutex{},
	}
	return r
}

func (r *runner) SetMetricCatalog(c catalogsMetrics) {
	r.metricCatalog = c
}

func (r *runner) SetPluginManager(m managesPlugins) {
	r.pluginManager = m
}

func (r *runner) AvailablePlugins() *availablePlugins {
	return r.availablePlugins
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

	// Stop the monitor
	r.monitor.Stop()

	// TODO: Actually stop the plugins

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
	ap, err := newAvailablePlugin(resp)
	if err != nil {
		return nil, err
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
	err := ap.Stop(reason)
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

	switch v := e.Body.(type) {
	case *control_event.MetricSubscriptionEvent:
		fmt.Println("runner")
		fmt.Println(v.Namespace())
		fmt.Printf("Metric subscription (%s v%d)\n", strings.Join(v.MetricNamespace, "/"), v.Version)

		// Our logic here is simple for alpha. We should replace with parameter managed logic.
		//
		// 1. Get the loaded plugin for the subscription.
		// 2. Check that at least one available plugin of that type is running
		// 3. If not start one

		mt, err := r.metricCatalog.Get(v.MetricNamespace, v.Version)
		if err != nil {
			// log this error # TODO with logging
			fmt.Println(err)
		}
		fmt.Printf("Plugin is (%s)\n", mt.Plugin.Key())

		pool := r.availablePlugins.Collectors.GetPluginPool(mt.Plugin.Key())
		if pool != nil && pool.Count() >= MaximumRunningPlugins {
			// if r.availablePlugins.Collectors.PluginPoolHasAP(mt.Plugin.Key()) {
			fmt.Println("We have at least one running!")
			return
		}

		fmt.Println("No APs running!")
		ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(), mt.Plugin.Path)
		if err != nil {
			fmt.Println(err)
		}
		ap, err := r.startPlugin(ePlugin)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println("NEW AP")
		fmt.Println(ap)
	}
}
