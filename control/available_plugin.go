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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/client"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
)

const (
	// DefaultClientTimeout - default timeout for a client connection attempt
	DefaultClientTimeout = time.Second * 3
	// DefaultHealthCheckTimeout - default timeout for a health check
	DefaultHealthCheckTimeout = time.Second * 1
	// DefaultHealthCheckFailureLimit - how any consecutive health check timeouts must occur to trigger a failure
	DefaultHealthCheckFailureLimit = 3
)

var (
	ErrPoolNotFound = errors.New("plugin pool not found")
	ErrBadKey       = errors.New("bad key")
)

// availablePlugin represents a plugin which is
// running and available to respond to requests
type availablePlugin struct {
	meta               plugin.PluginMeta
	key                string
	pluginType         plugin.PluginType
	client             client.PluginClient
	name               string
	version            int
	id                 uint32
	hitCount           int
	lastHitTime        time.Time
	emitter            gomit.Emitter
	failedHealthChecks int
	healthChan         chan error
	ePlugin            executablePlugin
	exec               string
	execPath           string
	fromPackage        bool
}

// newAvailablePlugin returns an availablePlugin with information from a
// plugin.Response
func newAvailablePlugin(resp *plugin.Response, emitter gomit.Emitter, ep executablePlugin) (*availablePlugin, error) {
	if resp.Type != plugin.CollectorPluginType && resp.Type != plugin.ProcessorPluginType && resp.Type != plugin.PublisherPluginType {
		return nil, strategy.ErrBadType
	}
	ap := &availablePlugin{
		meta:        resp.Meta,
		name:        resp.Meta.Name,
		version:     resp.Meta.Version,
		pluginType:  resp.Type,
		emitter:     emitter,
		healthChan:  make(chan error, 1),
		lastHitTime: time.Now(),
		ePlugin:     ep,
	}
	ap.key = fmt.Sprintf("%s:%s:%d", ap.pluginType.String(), ap.name, ap.version)

	listenURL := fmt.Sprintf("http://%v/rpc", resp.ListenAddress)
	// Create RPC Client
	switch resp.Type {
	case plugin.CollectorPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewCollectorHttpJSONRPCClient(listenURL, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewCollectorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		}
	case plugin.PublisherPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewPublisherHttpJSONRPCClient(listenURL, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewPublisherNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		}
	case plugin.ProcessorPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewProcessorHttpJSONRPCClient(listenURL, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewProcessorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		}
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + resp.Type.String())
	}

	return ap, nil
}

func (a *availablePlugin) ID() uint32 {
	return a.id
}

func (a *availablePlugin) SetID(id uint32) {
	a.id = id
}

func (a *availablePlugin) Exclusive() bool {
	return a.meta.Exclusive
}

func (a *availablePlugin) CacheTTL() time.Duration {
	return a.meta.CacheTTL
}

func (a *availablePlugin) RoutingStrategy() plugin.RoutingStrategyType {
	return a.meta.RoutingStrategy
}

func (a *availablePlugin) ConcurrencyCount() int {
	return a.meta.ConcurrencyCount
}

func (a *availablePlugin) String() string {
	return fmt.Sprintf("%s:%s:v%d:id%d", a.TypeName(), a.name, a.version, a.id)
}

func (a *availablePlugin) TypeName() string {
	return a.pluginType.String()
}

func (a *availablePlugin) Type() plugin.PluginType {
	return a.pluginType
}

func (a *availablePlugin) Name() string {
	return a.name
}

func (a *availablePlugin) Version() int {
	return a.version
}

func (a *availablePlugin) HitCount() int {
	return a.hitCount
}

func (a *availablePlugin) LastHit() time.Time {
	return a.lastHitTime
}

// Stop halts a running availablePlugin
func (a *availablePlugin) Stop(r string) error {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "stop",
		"aplugin": a,
	}).Info("stopping available plugin")
	return a.client.Kill(r)
}

// Kill assumes aplugin is not able to here a Kill RPC call
func (a *availablePlugin) Kill(r string) error {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "kill",
		"aplugin": a,
	}).Info("hard killing available plugin")
	if a.fromPackage {
		log.WithFields(log.Fields{
			"_module":    "control-aplugin",
			"block":      "kill",
			"aplugin":    a,
			"pluginPath": path.Join(a.execPath, a.exec),
		}).Debug("deleting available plugin path")
		os.RemoveAll(filepath.Dir(a.execPath))
	}
	return a.ePlugin.Kill()
}

// CheckHealth checks the health of a plugin and updates
// a.failedHealthChecks
func (a *availablePlugin) CheckHealth() {
	go func() {
		a.healthChan <- a.client.Ping()
	}()
	select {
	case err := <-a.healthChan:
		if err == nil {
			if a.failedHealthChecks > 0 {
				// only log on first ok health check
				log.WithFields(log.Fields{
					"_module": "control-aplugin",
					"block":   "check-health",
					"aplugin": a,
				}).Debug("health is ok")
			}
			a.failedHealthChecks = 0
		} else {
			a.healthCheckFailed()
		}
	case <-time.After(DefaultHealthCheckTimeout):
		a.healthCheckFailed()
	}
}

// healthCheckFailed increments a.failedHealthChecks and emits a DisabledPluginEvent
// and a HealthCheckFailedEvent
func (a *availablePlugin) healthCheckFailed() {
	log.WithFields(log.Fields{
		"_module": "control-aplugin",
		"block":   "check-health",
		"aplugin": a,
	}).Warning("heartbeat missed")
	a.failedHealthChecks++
	if a.failedHealthChecks >= DefaultHealthCheckFailureLimit {
		log.WithFields(log.Fields{
			"_module": "control-aplugin",
			"block":   "check-health",
			"aplugin": a,
		}).Warning("heartbeat failed")
		pde := &control_event.DeadAvailablePluginEvent{
			Name:    a.name,
			Version: a.version,
			Type:    int(a.pluginType),
			Key:     a.key,
			Id:      a.ID(),
			String:  a.String(),
		}
		defer a.emitter.Emit(pde)
	}
	hcfe := &control_event.HealthCheckFailedEvent{
		Name:    a.name,
		Version: a.version,
		Type:    int(a.pluginType),
	}
	defer a.emitter.Emit(hcfe)
}

type availablePlugins struct {
	// Used to coordinate operations on the table.
	*sync.RWMutex
	// table holds all the plugin pools.
	// The Pools' primary keys are equal to
	// {plugin_type}:{plugin_name}:{plugin_version}
	table map[string]strategy.Pool
}

func newAvailablePlugins() *availablePlugins {
	return &availablePlugins{
		RWMutex: &sync.RWMutex{},
		table:   make(map[string]strategy.Pool),
	}
}

func (ap *availablePlugins) insert(pl *availablePlugin) error {
	if pl.pluginType != plugin.CollectorPluginType && pl.pluginType != plugin.ProcessorPluginType && pl.pluginType != plugin.PublisherPluginType {
		return strategy.ErrBadType
	}

	ap.Lock()
	defer ap.Unlock()

	key := fmt.Sprintf("%s:%s:%d", pl.TypeName(), pl.name, pl.version)
	_, exists := ap.table[key]
	if !exists {
		p, err := strategy.NewPool(key, pl)
		if err != nil {
			return serror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}
		ap.table[key] = p
		return nil
	}
	ap.table[key].Insert(pl)
	return nil
}

func (ap *availablePlugins) getPool(key string) (strategy.Pool, serror.SnapError) {
	ap.RLock()
	defer ap.RUnlock()
	pool, ok := ap.table[key]
	if !ok {
		tnv := strings.Split(key, ":")
		if len(tnv) != 3 {
			return nil, serror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}

		v, err := strconv.Atoi(tnv[2])
		if err != nil {
			return nil, serror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}

		if v < 1 {
			return ap.findLatestPool(tnv[0], tnv[1])
		}
		// No key found
		return nil, serror.New(ErrBadKey, map[string]interface{}{"key": key})
	}

	return pool, nil
}

func (ap *availablePlugins) hasPool(key string) bool {
	if pool, _ := ap.getPool(key); pool != nil {
		return true
	}
	return false
}

func (ap *availablePlugins) collectMetrics(pluginKey string, metricTypes []core.Metric, taskID string) ([]core.Metric, error) {
	var results []core.Metric
	pool, serr := ap.getPool(pluginKey)
	if serr != nil {
		return nil, serr
	}
	if pool == nil {
		return nil, serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": pluginKey})
	}
	// If the strategy is nil but the pool exists we likely are waiting on the pool to be fully initialized
	// because of a plugin load/unload event that is currently being processed. Prevents panic from using nil
	// RoutingAndCaching.
	if pool.Strategy() == nil {
		return nil, errors.New("Plugin strategy not set")
	}

	metricsToCollect, metricsFromCache := pool.CheckCache(metricTypes, taskID)

	if len(metricsToCollect) == 0 {
		return metricsFromCache, nil
	}

	pool.RLock()
	defer pool.RUnlock()
	p, serr := pool.SelectAP(taskID)
	if serr != nil {
		return nil, serr
	}

	// cast client to PluginCollectorClient
	cli, ok := p.(*availablePlugin).client.(client.PluginCollectorClient)
	if !ok {
		return nil, serror.New(errors.New("unable to cast client to PluginCollectorClient"))
	}

	// collect metrics
	metrics, err := cli.CollectMetrics(metricsToCollect)
	if err != nil {
		return nil, serror.New(err)
	}

	pool.UpdateCache(metrics, taskID)

	results = make([]core.Metric, len(metricsFromCache)+len(metrics))
	idx := 0
	for _, m := range metrics {
		results[idx] = m
		idx++
	}
	for _, m := range metricsFromCache {
		results[idx] = m
		idx++
	}

	// update plugin stats
	p.(*availablePlugin).hitCount++
	p.(*availablePlugin).lastHitTime = time.Now()

	return results, nil
}

func (ap *availablePlugins) publishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error {
	var errs []error
	key := strings.Join([]string{plugin.PublisherPluginType.String(), pluginName, strconv.Itoa(pluginVersion)}, ":")
	pool, serr := ap.getPool(key)
	if serr != nil {
		errs = append(errs, serr)
		return errs
	}
	if pool == nil {
		return []error{serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": key})}
	}

	pool.RLock()
	defer pool.RUnlock()
	p, err := pool.SelectAP(taskID)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	cli, ok := p.(*availablePlugin).client.(client.PluginPublisherClient)
	if !ok {
		return []error{errors.New("unable to cast client to PluginPublisherClient")}
	}

	errp := cli.Publish(contentType, content, config)
	if errp != nil {
		return []error{errp}
	}
	p.(*availablePlugin).hitCount++
	p.(*availablePlugin).lastHitTime = time.Now()
	return nil
}

func (ap *availablePlugins) processMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) (string, []byte, []error) {
	var errs []error
	key := strings.Join([]string{plugin.ProcessorPluginType.String(), pluginName, strconv.Itoa(pluginVersion)}, ":")
	pool, serr := ap.getPool(key)
	if serr != nil {
		errs = append(errs, serr)
		return "", nil, errs
	}
	if pool == nil {
		return "", nil, []error{serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": key})}
	}

	pool.RLock()
	defer pool.RUnlock()
	p, err := pool.SelectAP(taskID)
	if err != nil {
		errs = append(errs, err)
		return "", nil, errs
	}

	cli, ok := p.(*availablePlugin).client.(client.PluginProcessorClient)
	if !ok {
		return "", nil, []error{errors.New("unable to cast client to PluginProcessorClient")}
	}

	ct, c, errp := cli.Process(contentType, content, config)
	if errp != nil {
		return "", nil, []error{errp}
	}
	p.(*availablePlugin).hitCount++
	p.(*availablePlugin).lastHitTime = time.Now()
	return ct, c, nil
}

func (ap *availablePlugins) findLatestPool(pType, name string) (strategy.Pool, serror.SnapError) {
	// see if there exists a pool at all which matches name version.
	var latest strategy.Pool
	for key, pool := range ap.table {
		tnv := strings.Split(key, ":")
		if tnv[0] == pType && tnv[1] == name {
			latest = pool
			break
		}
	}
	if latest != nil {
		for key, pool := range ap.table {
			tnv := strings.Split(key, ":")
			if tnv[0] == pType && tnv[1] == name && pool.Version() > latest.Version() {
				latest = pool
			}
		}
		return latest, nil
	}

	return nil, nil
}

func (ap *availablePlugins) getOrCreatePool(key string) (strategy.Pool, error) {
	var err error
	pool, ok := ap.table[key]
	if ok {
		return pool, nil
	}
	pool, err = strategy.NewPool(key)
	if err != nil {
		return nil, err
	}
	ap.table[key] = pool
	return pool, nil
}

func (ap *availablePlugins) pools() map[string]strategy.Pool {
	ap.RLock()
	defer ap.RUnlock()
	return ap.table
}

func (ap *availablePlugins) all() []strategy.AvailablePlugin {
	var aps = []strategy.AvailablePlugin{}
	ap.RLock()
	defer ap.RUnlock()
	for _, pool := range ap.table {
		for _, ap := range pool.Plugins() {
			aps = append(aps, ap)
		}
	}
	return aps
}
