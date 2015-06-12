package control

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/control_event"
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
	emitter          gomit.Emitter
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

func (r *runner) SetEmitter(e gomit.Emitter) {
	r.emitter = e
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
	log.WithFields(log.Fields{
		"module": "control-runner",
		"block":  "start",
	}).Debug("started")
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
	defer log.WithFields(log.Fields{
		"module": "control-runner",
		"block":  "start-plugin",
	}).Debug("stopped")
	return errs
}

func (r *runner) startPlugin(p executablePlugin) (*availablePlugin, error) {
	e := p.Start()
	if e != nil {
		e_ := errors.New("error while starting plugin: " + e.Error())
		defer log.WithFields(log.Fields{
			"module": "control-runner",
			"block":  "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e_
	}

	// Wait for plugin response
	resp, err := p.WaitForResponse(time.Second * 3)
	if err != nil {
		e := errors.New("error while waiting for response: " + err.Error())
		log.WithFields(log.Fields{
			"module": "control-runner",
			"block":  "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	if resp == nil {
		e := errors.New("no reponse object returned from plugin")
		log.WithFields(log.Fields{
			"module": "control-runner",
			"block":  "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	if resp.State != plugin.PluginSuccess {
		e := errors.New("plugin could not start error: " + resp.ErrorMessage)
		log.WithFields(log.Fields{
			"module": "control-runner",
			"block":  "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	// build availablePlugin
	ap, err := newAvailablePlugin(resp, r.apIdCounter.Next(), r.emitter)
	if err != nil {
		return nil, err
	}

	// Ping through client
	err = ap.Client.Ping()
	if err != nil {
		return nil, err
	}

	r.availablePlugins.Insert(ap)
	log.WithFields(log.Fields{
		"module":           "control-runner",
		"block":            "start-plugin",
		"available-plugin": ap.String(),
	}).Info("available plugin started")

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
	case *control_event.ProcessorSubscriptionEvent:
		log.WithFields(log.Fields{
			"module":         "control-runner",
			"block":          "handle-events",
			"event":          v.Namespace(),
			"plugin-name":    v.PluginName,
			"plugin-version": v.PluginVersion,
			"plugin-type":    "processor",
		}).Debug("handling processor subscription event")
		r.mutex.Lock()
		defer r.mutex.Unlock()
		for r.pluginManager.LoadedPlugins().Next() {
			_, lp := r.pluginManager.LoadedPlugins().Item()
			if lp.TypeName() == "processor" && lp.Name() == v.PluginName && lp.Version() == v.PluginVersion {
				pool := r.availablePlugins.Processors.GetPluginPool(lp.Key())
				ok := checkPool(pool, lp.Key())
				if !ok {
					return
				}

				ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(lp.Path), lp.Path)
				_, err = r.startPlugin(ePlugin)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}

		}
	case *control_event.PublisherSubscriptionEvent:
		log.WithFields(log.Fields{
			"module":         "control-runner",
			"block":          "handle-events",
			"event":          v.Namespace(),
			"plugin-name":    v.PluginName,
			"plugin-version": v.PluginVersion,
			"plugin-type":    "publisher",
		}).Debug("handling processor subscription event")
		r.mutex.Lock()
		defer r.mutex.Unlock()
		for r.pluginManager.LoadedPlugins().Next() {
			_, lp := r.pluginManager.LoadedPlugins().Item()
			if lp.TypeName() == "publisher" && lp.Name() == v.PluginName && lp.Version() == v.PluginVersion {
				pool := r.availablePlugins.Publishers.GetPluginPool(lp.Key())
				ok := checkPool(pool, lp.Key())
				if !ok {
					return
				}

				ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(lp.Path), lp.Path)
				_, err = r.startPlugin(ePlugin)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}

		}
	case *control_event.MetricSubscriptionEvent:
		log.WithFields(log.Fields{
			"module":           "control-runner",
			"block":            "handle-events",
			"event":            v.Namespace(),
			"metric-namespace": v.MetricNamespace,
			"metric-version":   v.Version,
		}).Debug("handling metric subscription event")
		r.mutex.Lock()
		defer r.mutex.Unlock()

		// Our logic here is simple for alpha. We should replace with parameter managed logic.
		//
		// 1. Get the loaded plugin for the subscription.
		// 2. Check that at least one available plugin of that type is running
		// 3. If not start one

		mt, err := r.metricCatalog.Get(v.MetricNamespace, v.Version)
		if err != nil {
			// log this error # TODO with logging
			log.WithFields(log.Fields{
				"module": "control-runner",
				"block":  "handle-events",
				"event":  v.Namespace(),
				"error":  err,
			}).Error("error on getting metric from metric catalog")
			return
		}
		log.WithFields(log.Fields{
			"module":           "control-runner",
			"block":            "handle-events",
			"event":            v.Namespace(),
			"metric-namespace": v.MetricNamespace,
			"metric-version":   v.Version,
			"plugin":           mt.Plugin.Key(),
		}).Debug("plugin found for metric")

		pool := r.availablePlugins.Collectors.GetPluginPool(mt.Plugin.Key())
		ok := checkPool(pool, mt.Plugin.Key())
		if !ok {
			return
		}

		ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(mt.Plugin.Path), mt.Plugin.Path)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "control-runner",
				"block":  "handle-events",
				"event":  v.Namespace(),
				"plugin": mt.Plugin.Key(),
				"path":   mt.Plugin.Path,
				"error":  err,
			}).Error("error creating executable plugin")
		}
		_, err = r.startPlugin(ePlugin)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "control-runner",
				"block":  "handle-events",
				"event":  v.Namespace(),
				"plugin": mt.Plugin.Key(),
				"error":  err,
			}).Error("error starting new plugin")
		}
	}
}

func checkPool(pool *availablePluginPool, key string) bool {
	if pool != nil && pool.Count() >= MaximumRunningPlugins {
		log.WithFields(log.Fields{
			"module":     "control-runner",
			"block":      "check-pool",
			"plugin":     key,
			"pool-count": pool.Count(),
			"max":        MaximumRunningPlugins,
		}).Debug("pool is large enough")
		return false
	}
	if pool == nil {
		log.WithFields(log.Fields{
			"module": "control-runner",
			"block":  "check-pool",
			"plugin": key,
			"max":    MaximumRunningPlugins,
		}).Debug("pool is not created")
	} else {
		log.WithFields(log.Fields{
			"module":     "control-runner",
			"block":      "check-pool",
			"plugin":     key,
			"pool-count": pool.Count(),
			"max":        MaximumRunningPlugins,
		}).Debug("pool is not large enough")
	}
	return true
}
