/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package control

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/pkg/aci"
)

var (
	runnerLog = log.WithField("_module", "control-runner")
)

type availablePluginState int

const (
	// HandlerRegistrationName is the registered name of the control runner
	HandlerRegistrationName = "control.runner"

	// PluginRunning is the running state of a plugin
	PluginRunning availablePluginState = iota - 1 // Default value (0) is Running
	// PluginStopped is the stopped state of a plugin
	PluginStopped
	// PluginDisabled is the disabled state of a plugin
	PluginDisabled

	// MaximumRestartOnDeadPluginEvent is the maximum count of restarting a plugin
	// after the event of control_event.DeadAvailablePluginEvent
	MaxPluginRestartCount = 3
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
	emitter          gomit.Emitter
	monitor          *monitor
	availablePlugins *availablePlugins
	metricCatalog    catalogsMetrics
	pluginManager    managesPlugins
}

func newRunner() *runner {
	r := &runner{
		monitor:          newMonitor(),
		availablePlugins: newAvailablePlugins(),
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
		err := errors.New("error while starting plugin: " + e.Error())
		defer runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, err
	}

	// Wait for plugin response
	resp, err := p.WaitForResponse(time.Second * 5)
	if err != nil {
		e := errors.New("error while waiting for response: " + err.Error())
		runnerLog.WithFields(log.Fields{
			"_block": "start-plugin",
			"error":  e.Error(),
		}).Error("error starting a plugin")
		return nil, e
	}

	if resp == nil {
		e := errors.New("no response object returned from plugin")
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

	if resp.Meta.Unsecure {
		err = ap.client.Ping()
	} else {
		err = ap.client.SetKey()
	}
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
	pool, err := r.availablePlugins.getPool(ap.key)
	if err != nil {
		return err
	}
	if pool != nil {
		pool.Kill(ap.id, reason)
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
			pool.Kill(v.Id, "plugin dead")
		}

		if pool.Eligible() {
			if pool.RestartCount() < MaxPluginRestartCount {
				e := r.restartPlugin(v.Key)
				if e != nil {
					runnerLog.WithFields(log.Fields{
						"_block":  "handle-events",
						"aplugin": v.String,
					}).Error(err.Error())
					return
				}
				pool.IncRestartCount()

				runnerLog.WithFields(log.Fields{
					"_block":        "handle-events",
					"event":         v.Name,
					"aplugin":       v.Version,
					"restart_count": pool.RestartCount(),
				}).Warning("plugin restarted")

				r.emitter.Emit(&control_event.RestartedAvailablePluginEvent{
					Id:      v.Id,
					Name:    v.Name,
					Version: v.Version,
					Key:     v.Key,
					Type:    v.Type,
				})
			} else {
				r.emitter.Emit(&control_event.MaxPluginRestartsExceededEvent{
					Id:      v.Id,
					Name:    v.Name,
					Version: v.Version,
					Key:     v.Key,
					Type:    v.Type,
				})
			}
		}
	case *control_event.PluginUnsubscriptionEvent:
		runnerLog.WithFields(log.Fields{
			"_block":         "subscribe-pool",
			"event":          v.Namespace(),
			"plugin-name":    v.PluginName,
			"plugin-version": v.PluginVersion,
			"plugin-type":    core.PluginType(v.PluginType).String(),
		}).Debug("handling plugin unsubscription event")

		err := r.handleUnsubscription(core.PluginType(v.PluginType).String(), v.PluginName, v.PluginVersion, v.TaskId)
		if err != nil {
			return
		}
	case *control_event.UnloadPluginEvent:
		// On plugin unload,  find the key and pool info for the plugin being unloaded.
		r.availablePlugins.RLock()
		var pool strategy.Pool
		var k string
		for key, p := range r.availablePlugins.table {
			tnv := strings.Split(key, ":")
			if core.PluginType(v.Type).String() == tnv[0] && v.Name == tnv[1] && v.Version == p.Version() {
				pool = p
				k = key
				break
			}
		}

		r.availablePlugins.RUnlock()
		if pool == nil {
			return
		}
		// Check for the highest lower version plugin and move subscriptions that
		// are not bound to a plugin version to this pool.
		plugin, err := r.pluginManager.get(fmt.Sprintf("%s:%s:%d", core.PluginType(v.Type).String(), v.Name, -1))
		if err != nil {
			return
		}
		newPool, err := r.availablePlugins.getOrCreatePool(plugin.Key())
		if err != nil {
			return
		}
		subs := pool.MoveSubscriptions(newPool)
		// Start new plugins in newPool if needed
		if newPool.Eligible() {
			e := r.restartPlugin(plugin.Key())
			if e != nil {
				runnerLog.WithFields(log.Fields{
					"_block": "handle-events",
				}).Error(err.Error())
				return
			}
		}
		// Remove the unloaded plugin from available plugins
		r.availablePlugins.Lock()
		delete(r.availablePlugins.table, k)
		r.availablePlugins.Unlock()
		pool.RLock()
		defer pool.RUnlock()
		if len(subs) != 0 {
			runnerLog.WithFields(log.Fields{
				"_block":         "subscribe-pool",
				"event":          v.Namespace(),
				"plugin-name":    v.Name,
				"plugin-version": v.Version,
				"plugin-type":    core.PluginType(v.Type).String(),
			}).Info("pool with subscriptions to move found")
			for _, sub := range subs {
				r.emitter.Emit(&control_event.PluginSubscriptionEvent{
					PluginName:       v.Name,
					PluginVersion:    v.Version,
					TaskId:           sub.TaskID,
					PluginType:       v.Type,
					SubscriptionType: int(strategy.UnboundSubscriptionType),
				})
				r.emitter.Emit(&control_event.PluginUnsubscriptionEvent{
					PluginName:    v.Name,
					PluginVersion: pool.Version(),
					TaskId:        sub.TaskID,
					PluginType:    v.Type,
				})
				r.emitter.Emit(&control_event.MovePluginSubscriptionEvent{
					PluginName:      v.Name,
					PreviousVersion: pool.Version(),
					NewVersion:      v.Version,
					TaskId:          sub.TaskID,
					PluginType:      v.Type,
				})
			}
		}
	case *control_event.LoadPluginEvent:
		// On loaded plugin event all subscriptions that are not bound to a specific version
		// need to moved to the loaded version if it's version is greater than the currently
		// available plugin.
		var pool strategy.Pool
		//k := fmt.Sprintf("%v:%v:%v", core.PluginType(v.Type).String(), v.Name, -1)
		//pool, _ = r.availablePlugins.getPool(k)
		fmt.Println("\n\none")
		fmt.Println()
		r.availablePlugins.RLock()
		currentHighestVersion := -1
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
					"plugin-type":    core.PluginType(v.Type).String(),
					"plugin-signed":  v.Signed,
				}).Info("pool has bad key ", key)
				continue
			}
			// attempt to find a pool whose type and name are the same, and whose version is
			// less than newly loaded plugin.
			if core.PluginType(v.Type).String() == tnv[0] && v.Name == tnv[1] && v.Version > p.Version() {
				// See if the pool version is higher than the current highest.
				// We only want to move subscriptions from the currentHighest
				// because that is where subscriptions that are bound to the
				// latest version will be.
				if p.Version() > currentHighestVersion {
					pool = p
					currentHighestVersion = p.Version()
				}
			}
		}
		fmt.Println("\n\ntwo")
		fmt.Println()
		r.availablePlugins.RUnlock()

		fmt.Println("\n\nthree")
		fmt.Println()
		// now check to see if anything was put where pool points.
		// if not, there are no older pools whose subscriptions need to be
		// moved.
		if pool == nil {
			runnerLog.WithFields(log.Fields{
				"_block":         "subscribe-pool",
				"event":          v.Namespace(),
				"plugin-name":    v.Name,
				"plugin-version": v.Version,
				"plugin-type":    core.PluginType(v.Type).String(),
			}).Info("No previous pool found for loaded plugin")
			return
		}
		plugin, err := r.pluginManager.get(fmt.Sprintf("%s:%s:%d", core.PluginType(v.Type).String(), v.Name, v.Version))
		if err != nil {
			return
		}
		fmt.Println("\n\nfour")
		fmt.Println()
		newPool, err := r.availablePlugins.getOrCreatePool(plugin.Key())
		if err != nil {
			return
		}

		fmt.Println("\n\nfive")
		fmt.Println()
		// Move subscriptions to the new, higher versioned pool
		subs := pool.MoveSubscriptions(newPool)
		fmt.Println("\n\nsix")
		fmt.Println()
		if newPool.Eligible() {
			e := r.restartPlugin(plugin.Key())
			if e != nil {
				runnerLog.WithFields(log.Fields{
					"_block": "handle-events",
				}).Error(err.Error())
				return
			}
			runnerLog.WithFields(log.Fields{
				"_block": "pool eligible",
			}).Info("starting plugin")
			fmt.Println("\n\neleventy-ten")
			fmt.Println()
		}

		fmt.Println("\n\nseven")
		fmt.Println()
		pool.RLock()
		fmt.Println("\n\neight")
		fmt.Println()
		defer pool.RUnlock()

		if len(subs) != 0 {
			fmt.Println("\n\nnine")
			fmt.Println()
			runnerLog.WithFields(log.Fields{
				"_block":         "subscribe-pool",
				"event":          v.Namespace(),
				"plugin-name":    v.Name,
				"plugin-version": v.Version,
				"plugin-type":    core.PluginType(v.Type).String(),
			}).Info("pool with subscriptions to move found")
			for _, sub := range subs {
				fmt.Println("\n\nten")
				fmt.Println()
				/*r.emitter.Emit(&control_event.PluginSubscriptionEvent{
					PluginName:       v.Name,
					PluginVersion:    v.Version,
					TaskId:           sub.TaskID,
					PluginType:       v.Type,
					SubscriptionType: int(strategy.UnboundSubscriptionType),
				})
				r.emitter.Emit(&control_event.PluginUnsubscriptionEvent{
					PluginName:    v.Name,
					PluginVersion: pool.Version(),
					TaskId:        sub.TaskID,
					PluginType:    v.Type,
				})*/
				r.emitter.Emit(&control_event.MovePluginSubscriptionEvent{
					PluginName:      v.Name,
					PreviousVersion: pool.Version(),
					NewVersion:      v.Version,
					TaskId:          sub.TaskID,
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

func (r *runner) runPlugin(details *pluginDetails) error {
	if details.IsPackage {
		f, err := os.Open(details.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		tempPath, err := aci.Extract(f)
		if err != nil {
			return err
		}
		details.ExecPath = path.Join(tempPath, "rootfs")
	}
	ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(details.Exec), path.Join(details.ExecPath, details.Exec))
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   path.Join(details.ExecPath, details.Exec),
			"error":  err,
		}).Error("error creating executable plugin")
		return err
	}
	ap, err := r.startPlugin(ePlugin)
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   path.Join(details.ExecPath, details.Exec),
			"error":  err,
		}).Error("error starting new plugin")
		return err
	}
	ap.exec = details.Exec
	ap.execPath = details.ExecPath
	if details.IsPackage {
		ap.fromPackage = true
	}
	return nil
}

func (r *runner) handleUnsubscription(pType, pName string, pVersion int, taskID string) error {
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
	if pool.SubscriptionCount() < pool.Count() {
		runnerLog.WithFields(log.Fields{
			"_block":                  "handle-unsubscription",
			"pool-count":              pool.Count(),
			"pool-subscription-count": pool.SubscriptionCount(),
		}).Debug(fmt.Sprintf("killing an available plugin in pool  %s:%s:%d", pType, pName, pVersion))
		pool.SelectAndKill(taskID, "unsubscription event")
	}
	return nil
}

func (r *runner) restartPlugin(key string) error {
	lp, err := r.pluginManager.get(key)
	if err != nil {
		return err
	}
	return r.runPlugin(lp.Details)
}
