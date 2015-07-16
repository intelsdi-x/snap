package control

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/control_event"
)

var (
	runnerLog = log.WithField("_module", "control-runner")
)

type availablePluginState int

const (
	HandlerRegistrationName = "control.runner"

	// availablePlugin States
	PluginRunning availablePluginState = iota - 1 // Default value (0) is Running
	PluginStopped
	PluginDisabled
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
	routingStrategy  RoutingStrategy
}

func newRunner(routingStrategy RoutingStrategy) *runner {
	r := &runner{
		monitor:          newMonitor(),
		availablePlugins: newAvailablePlugins(routingStrategy),
		mutex:            &sync.Mutex{},
		apIdCounter:      &idCounter{mutex: &sync.Mutex{}},
		routingStrategy:  routingStrategy,
	}
	return r
}

func (r *runner) SetStrategy(rs RoutingStrategy) {
	r.routingStrategy = rs
}

func (r *runner) Strategy() RoutingStrategy {
	return r.routingStrategy
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
	runnerLog.WithFields(log.Fields{
		"_block": "start",
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
	defer runnerLog.WithFields(log.Fields{
		"_block": "start-plugin",
	}).Debug("stopped")
	return errs
}

func (r *runner) startPlugin(p executablePlugin) (*availablePlugin, error) {
	e := p.Start()
	if e != nil {
		e_ := errors.New("error while starting plugin: " + e.Error())
		defer runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e_
	}

	// Wait for plugin response
	resp, err := p.WaitForResponse(time.Second * 3)
	if err != nil {
		e := errors.New("error while waiting for response: " + err.Error())
		runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	if resp == nil {
		e := errors.New("no reponse object returned from plugin")
		runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	if resp.State != plugin.PluginSuccess {
		e := errors.New("plugin could not start error: " + resp.ErrorMessage)
		runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	// build availablePlugin
	ap, err := newAvailablePlugin(resp, r.emitter, p)
	if err != nil {
		return nil, err
	}

	// Ping through client
	err = ap.client.Ping()
	if err != nil {
		return nil, err
	}

	r.availablePlugins.insert(ap)
	runnerLog.WithFields(log.Fields{
		"_block":                "start-plugin",
		"available-plugin":      ap.String(),
		"available-plugin-type": ap.TypeName(),
	}).Info("available plugin started")

	return ap, nil
}

func (r *runner) stopPlugin(reason string, ap *availablePlugin) error {
	err := ap.Stop(reason)
	if err != nil {
		return err
	}
	pool, err := r.availablePlugins.getPool(ap.key)
	if err != nil {
		return err
	}
	if pool != nil {
		pool.remove(ap.id)
	}
	return nil
}

// Empty handler acting as placeholder until implementation. This helps tests
// pass to ensure registration works.
func (r *runner) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.DeadAvailablePluginEvent:
		runnerLog.WithFields(log.Fields{
			"_block":  "handle-events",
			"event":   v.Namespace(),
			"aplugin": v.String,
		}).Warning("handling dead available plugin event")
		pool, err := r.availablePlugins.getPool(v.Key)
		if err != nil {
			runnerLog.WithFields(log.Fields{
				"_block":  "handle-events",
				"aplugin": v.String,
			}).Error(err.Error())
			return
		}
		if pool != nil {
			pool.kill(v.Id, "plugin dead")
		}
	case *control_event.PluginSubscriptionEvent:
		runnerLog.WithFields(log.Fields{
			"_block":         "subscribe-pool",
			"event":          v.Namespace(),
			"plugin-name":    v.PluginName,
			"plugin-version": v.PluginVersion,
			"plugin-type":    v.PluginType,
		}).Debug("handling plugin subscription event")

		r.handleSubscription(
			core.PluginType(v.PluginType).String(),
			v.PluginName,
			v.PluginVersion,
			v.TaskId,
			subscriptionType(v.SubscriptionType),
		)

	case *control_event.PluginUnsubscriptionEvent:
		runnerLog.WithFields(log.Fields{
			"_block":         "subscribe-pool",
			"event":          v.Namespace(),
			"plugin-name":    v.PluginName,
			"plugin-version": v.PluginVersion,
			"plugin-type":    v.PluginType,
		}).Debug("handling plugin unsubscription event")

		err := r.handleUnsubscription(core.PluginType(v.PluginType).String(), v.PluginName, v.PluginVersion, v.TaskId)
		if err != nil {
			return
		}
	case *control_event.LoadPluginEvent:
		var pool *apPool
		r.availablePlugins.RLock()
		for key, p := range r.availablePlugins.pools() {
			// tuple of type name and version
			// type @ index 0, name @ index 1, version @ index 2
			tnv := strings.Split(key, ":")
			// make sure we don't panic and crash the service if a junk key is retrieved
			if len(tnv) != 3 {
				runnerLog.WithFields(log.Fields{
					"_block":         "subscribe-pool",
					"event":          v.Namespace(),
					"plugin-name":    v.Name,
					"plugin-version": v.Version,
					"plugin-type":    v.Type,
				}).Info("pool has bad key ", key)
				continue
			}

			// attempt to find a pool whose type and name are the same, and whose version is
			// less than newly loaded plugin.  If we find it, break out of loop.
			if core.PluginType(v.Type).String() == tnv[0] && v.Name == tnv[1] && v.Version > p.version {
				pool = p
				break
			}
		}
		r.availablePlugins.RUnlock()
		// now check to see if anything was put where pool points.
		// if not, there are no older pools whose subscriptions need to be
		// moved.
		if pool == nil {
			runnerLog.WithFields(log.Fields{
				"_block":         "subscribe-pool",
				"event":          v.Namespace(),
				"plugin-name":    v.Name,
				"plugin-version": v.Version,
				"plugin-type":    v.Type,
			}).Info("No previous pool found for loaded plugin")
			return
		}
		// walk through the subscriptions in the pool and move any subscriptions
		// which are unbound.
		for task, sub := range pool.subscriptions() {
			if sub.subType == unboundSubscriptionType && v.Version > sub.version {
				runnerLog.WithFields(log.Fields{
					"_block":         "subscribe-pool",
					"event":          v.Namespace(),
					"plugin-name":    v.Name,
					"plugin-version": v.Version,
					"plugin-type":    v.Type,
				}).Info("pool with subscriptions to move found")
				// subscribe to new pool
				plugin, err := r.pluginManager.get(fmt.Sprintf("%s:%s:%d", core.PluginType(v.Type).String(), v.Name, v.Version))
				if err != nil {
					return
				}
				newPool, err := r.availablePlugins.getOrCreatePool(plugin.Key())
				if err != nil {
					return
				}
				// subscribe new pool
				newPool.subscribe(task, unboundSubscriptionType)
				r.emitter.Emit(&control_event.PluginSubscriptionEvent{
					PluginName:       v.Name,
					PluginVersion:    v.Version,
					TaskId:           task,
					PluginType:       v.Type,
					SubscriptionType: int(unboundSubscriptionType),
				})
				// unsubscribe old pool
				pool.unsubscribe(task)
				r.emitter.Emit(&control_event.PluginUnsubscriptionEvent{
					PluginName:    v.Name,
					PluginVersion: sub.version,
					TaskId:        task,
					PluginType:    v.Type,
				})
				r.emitter.Emit(&control_event.MovePluginSubscriptionEvent{
					PluginName:      v.Name,
					PreviousVersion: sub.version,
					NewVersion:      v.Version,
					TaskId:          task,
					PluginType:      v.Type,
				})
			}
		}
	default:
		runnerLog.WithFields(log.Fields{
			"_block": "handle-events",
			"event":  v.Namespace(),
		}).Info("Nothing to do for this event")
	}
}

func (r *runner) runPlugin(path string) {
	ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(path), path)
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   path,
			"error":  err,
		}).Error("error creating executable plugin")
	}
	_, err = r.startPlugin(ePlugin)
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   path,
			"error":  err,
		}).Error("error starting new plugin")
	}
}

func (r *runner) handleSubscription(pType, pName string, pVersion int, taskId uint64, subType subscriptionType) {
	pool, err := r.availablePlugins.getPool(fmt.Sprintf("%s:%s:%d", pType, pName, pVersion))
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block":         "handle-subscription",
			"plugin-name":    pName,
			"plugin-version": pVersion,
			"plugin-type":    pType,
		}).Error("error retrieving pool")
		return
	}
	if pool == nil {
		runnerLog.WithFields(log.Fields{
			"_block":         "handle-subscription",
			"plugin-name":    pName,
			"plugin-version": pVersion,
			"plugin-type":    pType,
		}).Error("pool not found")
		return
	}
	runnerLog.WithFields(log.Fields{
		"_block":         "handle-subscription",
		"plugin-name":    pName,
		"plugin-version": pVersion,
		"plugin-type":    pType,
	}).Debug(fmt.Sprintf("found pool: version %d", pool.version))
	runnerLog.WithFields(log.Fields{
		"_block":                  "handle-subscription",
		"pool-count":              pool.count(),
		"pool-subscription-count": pool.subscriptionCount(),
		"pool-max":                pool.max,
		"pool-eligibility":        pool.eligible(),
	}).Debug("checking is pool is eligible to grow.")
	if pool.eligible() {
		runnerLog.WithFields(log.Fields{
			"_block":         "handle-subscription",
			"plugin-name":    pName,
			"plugin-version": pVersion,
			"plugin-type":    pType,
		}).Debug("pool is eligible. starting a new available plugin")
		plugin, err := r.pluginManager.get(fmt.Sprintf("%s:%s:%d", pType, pName, pVersion))
		if err != nil {
			runnerLog.WithFields(log.Fields{
				"_block":         "handle-subscription",
				"plugin-name":    pName,
				"plugin-version": pVersion,
				"plugin-type":    pType,
			}).Error("plugin not found")
			return
		}
		r.runPlugin(plugin.Path)
	}
}
func (r *runner) handleUnsubscription(pType, pName string, pVersion int, taskId uint64) error {
	pool, err := r.availablePlugins.getPool(fmt.Sprintf("%s:%s:%d", pType, pName, pVersion))
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block":         "handle-unsubscription",
			"plugin-name":    pName,
			"plugin-version": pVersion,
			"plugin-type":    pType,
		}).Error("error retrieving pool")
		return errors.New("error retrieving pool")
	}
	if pool == nil {
		runnerLog.WithFields(log.Fields{
			"_block":         "handle-unsubscription",
			"plugin-name":    pName,
			"plugin-version": pVersion,
			"plugin-type":    pType,
		}).Error("pool not found")
		return errors.New("pool not found")
	}
	if pool.subscriptionCount() < pool.count() {
		runnerLog.WithFields(log.Fields{
			"_block":                  "handle-unsubscription",
			"pool-count":              pool.count(),
			"pool-subscription-count": pool.subscriptionCount(),
		}).Debug(fmt.Sprintf("killing an available plugin in pool  %s:%s:%d", pType, pName, pVersion))
		pool.killLeastUsed("unsubscription event")
	}
	return nil
}
