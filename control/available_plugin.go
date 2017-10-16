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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	// DefaultClientTimeout - default timeout for RPC method completion
	DefaultClientTimeout = time.Second * 10
	// DefaultHealthCheckTimeout - default timeout for a health check
	DefaultHealthCheckTimeout = time.Second * 10
	// DefaultHealthCheckFailureLimit - how any consecutive health check timeouts must occur to trigger a failure
	DefaultHealthCheckFailureLimit = 3
)

var (
	ErrPoolNotFound      = errors.New("plugin pool not found")
	ErrBadKey            = errors.New("bad key")
	ErrMsgInsecurePlugin = "secure framework can't connect to insecure plugin"
	ErrMsgInsecureClient = "insecure framework can't connect to secure plugin"
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
	execPath           string
	fromPackage        bool
	pprofPort          string
	isRemote           bool
}

// newAvailablePlugin returns an availablePlugin with information from a
// plugin.Response
func newAvailablePlugin(resp plugin.Response, emitter gomit.Emitter, ep executablePlugin, security client.GRPCSecurity) (*availablePlugin, error) {
	if security.TLSEnabled && !resp.Meta.TLSEnabled {
		return nil, errors.New(ErrMsgInsecurePlugin + "; plugin_name: " + resp.Meta.Name)
	}
	if !security.TLSEnabled && resp.Meta.TLSEnabled {
		return nil, errors.New(ErrMsgInsecureClient + "; plugin_name: " + resp.Meta.Name)
	}
	if resp.Type != plugin.CollectorPluginType && resp.Type != plugin.ProcessorPluginType && resp.Type != plugin.PublisherPluginType && resp.Type != plugin.StreamCollectorPluginType {
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
		pprofPort:   resp.PprofAddress,
		isRemote:    false,
	}
	ap.key = fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", ap.pluginType.String(), ap.name, ap.version)

	// Create RPC Client
	switch resp.Type {
	case plugin.CollectorPluginType:
		switch resp.Meta.RPCType {
		case plugin.NativeRPC:
			log.WithFields(log.Fields{
				"_module":     "control-aplugin",
				"_block":      "newAvailablePlugin",
				"plugin_name": ap.name,
			}).Warning("This plugin is using a deprecated RPC protocol. Find more information here: https://github.com/intelsdi-x/snap/issues/1289 ")
			c, e := client.NewCollectorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.GRPC:
			c, e := client.NewCollectorGrpcClient(resp.ListenAddress, DefaultClientTimeout, security)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.STREAMGRPC:
			c, e := client.NewStreamCollectorGrpcClient(
				resp.ListenAddress,
				DefaultClientTimeout,
				security)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		default:
			return nil, errors.New("Invalid RPCTYPE")
		}
	case plugin.PublisherPluginType:
		switch resp.Meta.RPCType {
		case plugin.NativeRPC:
			c, e := client.NewPublisherNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.GRPC:
			c, e := client.NewPublisherGrpcClient(resp.ListenAddress, DefaultClientTimeout, security)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		default:
			return nil, errors.New("Invalid RPCTYPE")
		}
	case plugin.ProcessorPluginType:
		switch resp.Meta.RPCType {
		case plugin.NativeRPC:
			c, e := client.NewProcessorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, !resp.Meta.Unsecure)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.GRPC:
			c, e := client.NewProcessorGrpcClient(resp.ListenAddress, DefaultClientTimeout, security)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		default:
			return nil, errors.New("Invalid RPCTYPE")
		}
	case plugin.StreamCollectorPluginType:
		switch resp.Meta.RPCType {
		case plugin.STREAMGRPC:
			c, e := client.NewStreamCollectorGrpcClient(
				resp.ListenAddress,
				DefaultClientTimeout,
				security)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		default:
			return nil, errors.New("Invalid RPCTYPE")
		}
	default:
		return nil, errors.New("Cannot create a client for a plugin of the type: " + resp.Type.String())
	}

	return ap, nil
}

func (a *availablePlugin) Port() string {
	return a.pprofPort
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

func (a *availablePlugin) IsRemote() bool {
	return a.isRemote
}

func (a *availablePlugin) SetIsRemote(isRemote bool) {
	a.isRemote = isRemote
}

// Stop halts a running availablePlugin
func (a *availablePlugin) Stop(r string) error {
	log.WithFields(log.Fields{
		"_module":     "control-aplugin",
		"block":       "stop",
		"plugin_name": a,
	}).Info("stopping available plugin")
	if a.IsRemote() {
		return a.client.Close()
	}
	return a.client.Kill(r)
}

// Kill assumes a plugin is not able to hear a Kill RPC call
func (a *availablePlugin) Kill(r string) error {
	log.WithFields(log.Fields{
		"_module":     "control-aplugin",
		"block":       "kill",
		"plugin_name": a,
	}).Info("hard killing available plugin")
	if a.fromPackage {
		log.WithFields(log.Fields{
			"_module":     "control-aplugin",
			"block":       "kill",
			"plugin_name": a,
			"pluginPath":  a.execPath,
		}).Debug("deleting available plugin package")
		os.RemoveAll(filepath.Dir(a.execPath))
	}

	// If it's a streaming plugin, we need to signal the scheduler that
	// this plugin is being killed.
	if c, ok := a.client.(client.PluginStreamCollectorClient); ok {
		c.Killed()
	}

	if a.ePlugin != nil {
		return a.ePlugin.Kill()
	}
	return nil
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
					"_module":     "control-aplugin",
					"block":       "check-health",
					"plugin_name": a,
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
		"_module":     "control-aplugin",
		"block":       "check-health",
		"plugin_name": a,
	}).Warning("heartbeat missed")
	a.failedHealthChecks++
	if a.failedHealthChecks >= DefaultHealthCheckFailureLimit {
		log.WithFields(log.Fields{
			"_module":     "control-aplugin",
			"block":       "check-health",
			"plugin_name": a,
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
	if pl.pluginType != plugin.CollectorPluginType && pl.pluginType != plugin.ProcessorPluginType && pl.pluginType != plugin.PublisherPluginType && pl.pluginType != plugin.StreamCollectorPluginType {
		return strategy.ErrBadType
	}

	ap.Lock()
	defer ap.Unlock()

	key := fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", pl.TypeName(), pl.name, pl.version)
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
		tnv := strings.Split(key, core.Separator)
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

	config := metricTypes[0].Config()
	cfg := map[string]ctypes.ConfigValue{}
	if config != nil {
		cfg = config.Table()
	}

	pool.RLock()
	defer pool.RUnlock()
	p, serr := pool.SelectAP(taskID, cfg)
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

func (ap *availablePlugins) streamMetrics(
	pluginKey string,
	metricTypes []core.Metric,
	taskID string,
	maxCollectDuration time.Duration,
	maxMetricsBuffer int64) (chan []core.Metric, chan error, error) {

	pool, serr := ap.getPool(pluginKey)
	if serr != nil {
		return nil, nil, serr
	}
	if pool == nil {
		return nil, nil, serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": pluginKey})
	}

	if pool.Strategy() == nil {
		return nil, nil, errors.New("Plugin strategy not set")
	}

	config := metricTypes[0].Config()
	cfg := map[string]ctypes.ConfigValue{}
	if config != nil {
		cfg = config.Table()
	}

	pool.RLock()
	defer pool.RUnlock()
	p, serr := pool.SelectAP(taskID, cfg)
	if serr != nil {
		return nil, nil, serr
	}

	cli, ok := p.(*availablePlugin).client.(client.PluginStreamCollectorClient)
	if !ok {
		return nil, nil, serror.New(errors.New("Invalid streaming client"))
	}

	metricChan, errChan, err := cli.StreamMetrics(taskID, metricTypes)
	if err != nil {
		return nil, nil, serror.New(err)
	}
	err = cli.UpdateCollectDuration(maxCollectDuration)
	if err != nil {
		return nil, nil, serror.New(err)
	}
	err = cli.UpdateMetricsBuffer(maxMetricsBuffer)
	if err != nil {
		return nil, nil, serror.New(err)
	}

	return metricChan, errChan, nil
}

func (ap *availablePlugins) publishMetrics(metrics []core.Metric, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error {
	key := fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", plugin.PublisherPluginType.String(), pluginName, pluginVersion)
	pool, serr := ap.getPool(key)
	if serr != nil {
		return []error{serr}
	}
	if pool == nil {
		return []error{serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": key})}
	}

	pool.RLock()
	defer pool.RUnlock()

	p, serr := pool.SelectAP(taskID, config)
	if serr != nil {
		return []error{serr}
	}

	cli, ok := p.(*availablePlugin).client.(client.PluginPublisherClient)
	if !ok {
		return []error{errors.New("unable to cast client to PluginPublisherClient")}
	}

	err := cli.Publish(metrics, config)
	if err != nil {
		return []error{err}
	}
	p.(*availablePlugin).hitCount++
	p.(*availablePlugin).lastHitTime = time.Now()
	return nil
}

func (ap *availablePlugins) processMetrics(metrics []core.Metric, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) ([]core.Metric, []error) {
	var errs []error
	key := fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", plugin.ProcessorPluginType.String(), pluginName, pluginVersion)
	pool, serr := ap.getPool(key)
	if serr != nil {
		errs = append(errs, serr)
		return nil, errs
	}
	if pool == nil {
		return nil, []error{serror.New(ErrPoolNotFound, map[string]interface{}{"pool-key": key})}
	}

	pool.RLock()
	defer pool.RUnlock()
	p, err := pool.SelectAP(taskID, config)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	cli, ok := p.(*availablePlugin).client.(client.PluginProcessorClient)
	if !ok {
		return nil, []error{errors.New("unable to cast client to PluginProcessorClient")}
	}

	mts, errp := cli.Process(metrics, config)
	if errp != nil {
		return nil, []error{errp}
	}
	p.(*availablePlugin).hitCount++
	p.(*availablePlugin).lastHitTime = time.Now()
	return mts, nil
}

func (ap *availablePlugins) findLatestPool(pType, name string) (strategy.Pool, serror.SnapError) {
	// see if there exists a pool at all which matches name version.
	var latest strategy.Pool
	for key, pool := range ap.table {
		tnv := strings.Split(key, core.Separator)
		if tnv[0] == pType && tnv[1] == name && pool.Count() > 0 {
			latest = pool
			break
		}
	}
	if latest != nil {
		for key, pool := range ap.table {
			tnv := strings.Split(key, core.Separator)
			if tnv[0] == pType && tnv[1] == name && pool.Version() > latest.Version() && pool.Count() > 0 {
				latest = pool
			}
		}
		return latest, nil
	}

	return nil, nil
}

func (ap *availablePlugins) getOrCreatePool(key string) (strategy.Pool, error) {
	ap.Lock()
	defer ap.Unlock()
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
