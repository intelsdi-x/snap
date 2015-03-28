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
	"github.com/intelsdilabs/pulse/pkg/logger"
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

type idCounter struct {
	id    int
	mutex *sync.Mutex
}

func (i *idCounter) Next() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.id++
	return i.id
}

// Handles events pertaining to plugins and control the runnning state accordingly.
type runner struct {
	delegates        []gomit.Delegator
	monitor          *monitor
	availablePlugins *availablePlugins
	metricCatalog    catalogsMetrics
	pluginManager    managesPlugins
	mutex            *sync.Mutex
	apIdCounter      *idCounter
}

func newRunner() *runner {
	r := &runner{
		monitor:          newMonitor(),
		availablePlugins: newAvailablePlugins(),
		mutex:            &sync.Mutex{},
		apIdCounter:      &idCounter{mutex: &sync.Mutex{}},
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

func (r *runner) Monitor() *monitor {
	return r.monitor
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

	logger.Debug("runner.start", "started")
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
	defer logger.Debug("runner.stop", "stopped")
	return errs
}

func (r *runner) startPlugin(p executablePlugin) (*availablePlugin, error) {
	e := p.Start()
	if e != nil {
		e_ := errors.New("error while starting plugin: " + e.Error())
		defer logger.Error("runner.startplugin", e_.Error())
		return nil, e_
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
	ap, err := newAvailablePlugin(resp, r.apIdCounter.Next())
	if err != nil {
		return nil, err
	}

	// Ping through client
	err = ap.Client.Ping()
	if err != nil {
		return nil, err
	}

	r.availablePlugins.Insert(ap)
	logger.Infof("runner.events", "available plugin started (%s)", ap.String())

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
		r.mutex.Lock()
		defer r.mutex.Unlock()
		logger.Debugf("runner.events", "handling metric subscription event (%s v%d)", strings.Join(v.MetricNamespace, "/"), v.Version)

		// Our logic here is simple for alpha. We should replace with parameter managed logic.
		//
		// 1. Get the loaded plugin for the subscription.
		// 2. Check that at least one available plugin of that type is running
		// 3. If not start one

		mt, err := r.metricCatalog.Get(v.MetricNamespace, v.Version)
		if err != nil {
			// log this error # TODO with logging
			fmt.Println(err)
			return
		}
		logger.Debugf("runner.events", "plugin is (%s) for (%s v%d)", mt.Plugin.Key(), strings.Join(v.MetricNamespace, "/"), v.Version)

		pool := r.availablePlugins.Collectors.GetPluginPool(mt.Plugin.Key())
		if pool != nil && pool.Count() >= MaximumRunningPlugins {
			// if r.availablePlugins.Collectors.PluginPoolHasAP(mt.Plugin.Key()) {
			logger.Debugf("runner.events", "(%s) has %d available plugin running (need %d)", mt.Plugin.Key(), pool.Count(), MaximumRunningPlugins)
			return
		}
		if pool == nil {
			logger.Debugf("runner.events", "not enough available plugins (%d) running for (%s) need %d", 0, mt.Plugin.Key(), MaximumRunningPlugins)
		} else {
			logger.Debugf("runner.events", "not enough available plugins (%d) running for (%s) need %d", pool.Count(), mt.Plugin.Key(), MaximumRunningPlugins)
		}

		ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(), mt.Plugin.Path)
		if err != nil {
			fmt.Println(err)
		}
		_, err = r.startPlugin(ePlugin)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case *control_event.PublisherSubscriptionEvent:
		r.mutex.Lock()
		defer r.mutex.Unlock()
		logger.Debugf("runner.events", "handling publisher subscription event (%s v%d)", v.PublisherPlugin, v.Version)

		pl := getPluginByName(v.PublisherPlugin, v.Version, r.pluginManager.LoadedPlugins().Table())
		if pl == nil {
			logger.Debugf("runner.events", "plugin %s, version %d has not been found in available plugin collection\n", v.PublisherPlugin, v.Version)
			return
		}

		// TODO this is common for Metrics and Publisher subscription, refactor
		pool := r.availablePlugins.Publishers.GetPluginPool(v.PublisherPlugin)
		if pool != nil && pool.Count() >= MaximumRunningPlugins {
			// if r.availablePlugins.Collectors.PluginPoolHasAP(mt.Plugin.Key()) {
			logger.Debugf("runner.events", "(%s) has %d available plugin running (need %d)", pl.Meta.Name, pool.Count(), MaximumRunningPlugins)
			return
		}
		if pool == nil {
			logger.Debugf("runner.events", "not enough available plugins (%d) running for (%s) need %d", 0, pl.Meta.Name, MaximumRunningPlugins)
		} else {
			logger.Debugf("runner.events", "not enough available plugins (%d) running for (%s) need %d", pool.Count(), pl.Meta.Name, MaximumRunningPlugins)
		}

		ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(), pl.Path)
		if err != nil {
			fmt.Println(err)
		}
		_, err = r.startPlugin(ePlugin)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}

func getPluginByName(name string, version int, loadedPlugins []*loadedPlugin) *loadedPlugin {
	for _, lp := range loadedPlugins {
		if lp.Name() == name && lp.Version() == version {
			return lp
		}
	}
	return nil
}
