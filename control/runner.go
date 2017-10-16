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
	"time"

	"github.com/intelsdi-x/gomit"
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/client"
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
)

var (
	// MaximumRestartOnDeadPluginEvent is the maximum count of restarting a plugin
	// after the event of control_event.DeadAvailablePluginEvent
	MaxPluginRestartCount = 3

	defaultRunnerOpts = []pluginRunnerOpt{optDefaultRunnerSecurity()}
)

type executablePlugin interface {
	Run(time.Duration) (plugin.Response, error)
	Kill() error
}

// Handles events pertaining to plugins and control the runnning state accordingly.
type runner struct {
	delegates         []gomit.Delegator
	emitter           gomit.Emitter
	monitor           *monitor
	availablePlugins  *availablePlugins
	metricCatalog     catalogsMetrics
	pluginManager     managesPlugins
	grpcSecurity      client.GRPCSecurity
	pluginLoadTimeout int
}

func newRunner(opts ...pluginRunnerOpt) *runner {
	r := &runner{
		pluginLoadTimeout: defaultPluginLoadTimeout,
		monitor:           newMonitor(),
		availablePlugins:  newAvailablePlugins(),
	}
	mergedOpts := append([]pluginRunnerOpt{}, defaultRunnerOpts...)
	mergedOpts = append(mergedOpts, opts...)
	for _, opt := range append(mergedOpts) {
		opt(r)
	}
	return r
}

type pluginRunnerOpt func(*runner)

// OptEnableRunnerTLS enables the TLS configuration in runner
func OptEnableRunnerTLS(grpcSecurity client.GRPCSecurity) pluginRunnerOpt {
	return func(r *runner) {
		r.grpcSecurity = grpcSecurity
	}
}

func optDefaultRunnerSecurity() pluginRunnerOpt {
	return func(r *runner) {
		r.grpcSecurity = client.SecurityTLSOff()
	}
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

// SetPluginLoadTimeout sets plugin load timeout
func (r *runner) SetPluginLoadTimeout(timeout int) {
	r.pluginLoadTimeout = timeout
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
	type result struct {
		ap  *availablePlugin
		err error
	}
	resultChan := make(chan result)
	go func() {
		resp, err := p.Run(time.Second * time.Duration(r.pluginLoadTimeout))
		if err != nil {
			e := errors.New("error starting plugin: " + err.Error())
			runnerLog.WithFields(log.Fields{
				"_block": "start-plugin",
				"error":  e.Error(),
			}).Error("error starting a plugin")
			resultChan <- result{nil, e}
			return

		}

		if resp.State != plugin.PluginSuccess {
			e := errors.New("plugin could not start error: " + resp.ErrorMessage)
			runnerLog.WithFields(log.Fields{
				"_block": "start-plugin",
				"error":  e.Error(),
			}).Error("error starting a plugin")
			resultChan <- result{nil, e}
			return
		}

		// build availablePlugin
		ap, err := newAvailablePlugin(resp, r.emitter, p, r.grpcSecurity)
		if err != nil {
			resultChan <- result{nil, err}
			return
		}

		if resp.Meta.Unsecure {
			err = ap.client.Ping()
		} else {
			err = ap.client.SetKey()
		}
		if err != nil {
			resultChan <- result{nil, err}
			return
		}
		r.availablePlugins.insert(ap)

		runnerLog.WithFields(log.Fields{
			"_block":                "start-plugin",
			"available-plugin":      ap.String(),
			"available-plugin-type": ap.TypeName(),
		}).Info("available plugin started")
		resultChan <- result{ap, nil}

		defer r.emitter.Emit(&control_event.StartPluginEvent{
			Name:    ap.Name(),
			Version: ap.Version(),
			Type:    int(ap.Type()),
			Key:     ap.key,
			Id:      ap.ID(),
		})
	}()

	select {
	case results := <-resultChan:
		return results.ap, results.err
	case <-time.After(time.Second * time.Duration(r.pluginLoadTimeout)):
		e := errors.New("error starting plugin due to timeout")
		return nil, e
	}
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
			if pool.RestartCount() < MaxPluginRestartCount || MaxPluginRestartCount == -1 {
				e := r.restartPlugin(v.Key)
				if e != nil {
					runnerLog.WithFields(log.Fields{
						"_block":  "handle-events",
						"aplugin": v.String,
					}).Error(e.Error())
					return
				}
				pool.IncRestartCount()

				runnerLog.WithFields(log.Fields{
					"_block":        "handle-events",
					"aplugin":       v.String,
					"restart-count": pool.RestartCount(),
				}).Warning("plugin restarted")

				r.emitter.Emit(&control_event.RestartedAvailablePluginEvent{
					Id:      v.Id,
					Name:    v.Name,
					Version: v.Version,
					Key:     v.Key,
					Type:    v.Type,
				})
			} else {
				runnerLog.WithFields(log.Fields{
					"_block":  "handle-events",
					"aplugin": v.String,
				}).Warning("plugin disabled due to exceeding restart limit: ", MaxPluginRestartCount)

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
	default:
		runnerLog.WithFields(log.Fields{
			"_block": "handle-events",
			"event":  v.Namespace(),
		}).Info("Nothing to do for this event")
	}
}

func (r *runner) runPlugin(name string, details *pluginDetails) error {
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
	commands := make([]string, len(details.Exec))
	for i, e := range details.Exec {
		commands[i] = path.Join(details.ExecPath, e)
	}
	ePlugin, err := plugin.NewExecutablePlugin(r.pluginManager.GenerateArgs(int(log.GetLevel())).
		SetCertPath(details.CertPath).
		SetKeyPath(details.KeyPath).
		SetCACertPaths(details.CACertPaths).
		SetTLSEnabled(details.TLSEnabled), commands...)
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   commands,
			"error":  err,
		}).Error("error creating executable plugin")
		return err
	}
	ePlugin.SetName(name)
	ap, err := r.startPlugin(ePlugin)
	if err != nil {
		runnerLog.WithFields(log.Fields{
			"_block": "run-plugin",
			"path":   commands,
			"error":  err,
		}).Error("error starting new plugin")
		return err
	}
	ap.execPath = details.ExecPath
	if details.IsPackage {
		ap.fromPackage = true
	}
	return nil
}

func (r *runner) handleUnsubscription(pType, pName string, pVersion int, taskID string) error {
	pool, err := r.availablePlugins.getPool(fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", pType, pName, pVersion))
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
		_, err := r.pluginManager.get(fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", pType, pName, pVersion))
		if err != nil {
			runnerLog.WithFields(log.Fields{
				"_block":                  "handle-unsubscription",
				"pool-count":              pool.Count(),
				"pool-subscription-count": pool.SubscriptionCount(),
				"plugin-name":             pName,
				"plugin-version":          pVersion,
				"plugin-type":             pType,
				"error":                   err.Error(),
			}).Error("unable to get loaded plugin")
		}
		pool.SelectAndKill(taskID, "unsubscription event")
	}
	return nil
}

func (r *runner) restartPlugin(key string) error {
	lp, err := r.pluginManager.get(key)
	if err != nil {
		return err
	}
	return r.runPlugin(lp.Name(), lp.Details)
}
