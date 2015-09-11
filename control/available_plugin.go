/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/client"
	"github.com/intelsdi-x/pulse/control/routing"
	"github.com/intelsdi-x/pulse/core/control_event"
	"github.com/intelsdi-x/pulse/core/perror"
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
	ErrBadType      = errors.New("bad plugin type")

	// This defines the maximum running instances of a loaded plugin.
	// It is initialized at runtime via the cli.
	maximumRunningPlugins = 3
)

type subscriptionType int

const (
	// this subscription is bound to an explicit version
	boundSubscriptionType subscriptionType = iota
	// this subscription is akin to "latest" and must be moved if a newer version is loaded.
	unboundSubscriptionType
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
}

// newAvailablePlugin returns an availablePlugin with information from a
// plugin.Response
func newAvailablePlugin(resp *plugin.Response, privKey *rsa.PrivateKey, emitter gomit.Emitter, ep executablePlugin) (*availablePlugin, error) {
	if resp.Type != plugin.CollectorPluginType && resp.Type != plugin.ProcessorPluginType && resp.Type != plugin.PublisherPluginType {
		return nil, ErrBadType
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

	listenUrl := fmt.Sprintf("http://%v/rpc", resp.ListenAddress)
	// Create RPC Client
	switch resp.Type {
	case plugin.CollectorPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewCollectorHttpJSONRPCClient(listenUrl, DefaultClientTimeout, resp.PublicKey, privKey)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewCollectorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, privKey)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		}
	case plugin.PublisherPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewPublisherHttpJSONRPCClient(listenUrl, DefaultClientTimeout, resp.PublicKey, privKey)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewPublisherNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, privKey)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		}
	case plugin.ProcessorPluginType:
		switch resp.Meta.RPCType {
		case plugin.JSONRPC:
			c, e := client.NewProcessorHttpJSONRPCClient(listenUrl, DefaultClientTimeout, resp.PublicKey, privKey)
			if e != nil {
				return nil, errors.New("error while creating client connection: " + e.Error())
			}
			ap.client = c
		case plugin.NativeRPC:
			c, e := client.NewProcessorNativeClient(resp.ListenAddress, DefaultClientTimeout, resp.PublicKey, privKey)
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

func (a *availablePlugin) String() string {
	return fmt.Sprintf("%s:%s:v%d:id%d", a.TypeName(), a.name, a.version, a.id)
}

func (a *availablePlugin) TypeName() string {
	return a.pluginType.String()
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
	case <-time.After(time.Second * 1):
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

type apPool struct {
	// used to coordinate changes to a pool
	*sync.RWMutex

	// the version of the plugins in the pool.
	// subscriptions uses this.
	version int
	// key is the primary key used in availablePlugins:
	// {plugin_type}:{plugin_name}:{plugin_version}
	key string

	// The subscriptions to this pool.
	subs map[uint64]*subscription

	// The plugins in the pool.
	// the primary key is an increasing --> uint from
	// pulsed epoch (`service pulsed start`).
	plugins    map[uint32]*availablePlugin
	pidCounter uint32

	// The max size which this pool may grow.
	max int

	// The number of subscriptions per running instance
	concurrencyCount int
}

func newPool(key string, plugins ...*availablePlugin) (*apPool, error) {
	versl := strings.Split(key, ":")
	ver, err := strconv.Atoi(versl[len(versl)-1])
	if err != nil {
		return nil, err
	}
	p := &apPool{
		RWMutex:          &sync.RWMutex{},
		version:          ver,
		key:              key,
		subs:             map[uint64]*subscription{},
		plugins:          make(map[uint32]*availablePlugin),
		max:              maximumRunningPlugins,
		concurrencyCount: 1,
	}

	if len(plugins) > 0 {
		for _, plg := range plugins {
			plg.id = p.generatePID()
			p.plugins[plg.id] = plg
		}
		// Because plugin metadata is a singleton and immutable (in static code)
		// it is safe to take the first item.  Reloading an identical plugin
		// with new metadata is protected by plugin loading.

		// Checking if plugin is exclusive
		// (only one instance should be running).
		if plugins[0].meta.Exclusive {
			p.max = 1
		}
		// set concurrency count
		p.concurrencyCount = plugins[0].meta.ConcurrencyCount
	}

	return p, nil
}

func (p *apPool) insert(ap *availablePlugin) error {
	if ap.pluginType != plugin.CollectorPluginType && ap.pluginType != plugin.ProcessorPluginType && ap.pluginType != plugin.PublisherPluginType {
		return ErrBadType
	}
	ap.id = p.generatePID()
	p.plugins[ap.id] = ap

	// If an empty pool is created, it does not have
	// any available plugins from which to retrieve
	// concurrency count or exclusivity.  We ensure it
	// is set correctly on an insert.
	if ap.meta.Exclusive {
		p.max = 1
	}
	p.concurrencyCount = ap.meta.ConcurrencyCount
	return nil
}

// subscribe adds a subscription to the pool.
// Using subscribe is idempotent.
func (p *apPool) subscribe(taskId uint64, subType subscriptionType) {
	p.Lock()
	defer p.Unlock()

	if _, exists := p.subs[taskId]; !exists {
		// Version is the last item in the key, so we split here
		// to retrieve it for the subscription.
		p.subs[taskId] = &subscription{
			taskId:  taskId,
			subType: subType,
			version: p.version,
		}
	}
}

// subscribe adds a subscription to the pool.
// Using unsubscribe is idempotent.
func (p *apPool) unsubscribe(taskId uint64) {
	p.Lock()
	defer p.Unlock()
	delete(p.subs, taskId)
}

func (p *apPool) eligible() bool {
	p.RLock()
	defer p.RUnlock()

	// optimization: don't even bother with concurrency
	// count if we have already reached pool max
	if p.count() == p.max {
		return false
	}

	should := p.subscriptionCount() / p.concurrencyCount
	if should > p.count() && should <= p.max {
		return true
	}

	return false
}

// kill kills and removes the available plugin from its pool.
// Using kill is idempotent.
func (p *apPool) kill(id uint32, reason string) {
	p.Lock()
	defer p.Unlock()

	ap, ok := p.plugins[id]
	if ok {
		ap.Kill(reason)
		delete(p.plugins, id)
	}
}

// kills the plugin with the lowest hit count
func (p *apPool) killLeastUsed(reason string) {
	p.Lock()
	defer p.Unlock()

	var (
		id   uint32
		prev int
	)

	// grab details from the first item
	for _, p := range p.plugins {
		prev = p.hitCount
		id = p.id
		break
	}

	// walk through all and find the lowest hit count
	for _, p := range p.plugins {
		if p.hitCount < prev {
			prev = p.hitCount
			id = p.id
		}
	}

	// kill that ap
	ap, ok := p.plugins[id]
	if ok {
		// only log on first ok health check
		log.WithFields(log.Fields{
			"_module": "control-aplugin",
			"block":   "kill-least-used",
			"aplugin": ap,
		}).Debug("killing available plugin")
		ap.Kill(reason)
		delete(p.plugins, id)
	}
}

// remove removes an available plugin from the the pool.
// using remove is idempotent.
func (p *apPool) remove(id uint32) {
	p.Lock()
	defer p.Unlock()
	delete(p.plugins, id)
}

func (p *apPool) count() int {
	return len(p.plugins)
}

// NOTE: The data returned by subscriptions should be constant and read only.
func (p *apPool) subscriptions() map[uint64]*subscription {
	p.RLock()
	defer p.RUnlock()
	return p.subs
}

func (p *apPool) subscriptionCount() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.subs)
}

func (p *apPool) selectAP(strat RoutingStrategy) (*availablePlugin, perror.PulseError) {
	p.RLock()
	defer p.RUnlock()

	sp := make([]routing.SelectablePlugin, p.count())
	i := 0
	for _, plg := range p.plugins {
		sp[i] = plg
		i++
	}
	sap, err := strat.Select(p, sp)
	if err != nil || sap == nil {
		return nil, perror.New(err)
	}
	return sap.(*availablePlugin), nil
}

func (p *apPool) generatePID() uint32 {
	atomic.AddUint32(&p.pidCounter, 1)
	return p.pidCounter
}

func (p *apPool) release() {
	p.RUnlock()
}

func (p *apPool) moveSubscriptions(to *apPool) []subscription {
	var subs []subscription

	p.Lock()
	defer p.Unlock()

	for task, sub := range p.subs {
		if sub.subType == unboundSubscriptionType && to.version > p.version {
			subs = append(subs, *sub)
			to.subscribe(task, unboundSubscriptionType)
			delete(p.subs, task)
		}
	}
	return subs
}

type subscription struct {
	subType subscriptionType
	version int
	taskId  uint64
}

type availablePlugins struct {
	// Used to coordinate operations on the table.
	*sync.RWMutex

	// the strategy used to select a plugin for execution
	routingStrategy RoutingStrategy

	// table holds all the plugin pools.
	// The Pools' primary keys are equal to
	// {plugin_type}:{plugin_name}:{plugin_version}
	table map[string]*apPool
}

func newAvailablePlugins(routingStrategy RoutingStrategy) *availablePlugins {
	return &availablePlugins{
		RWMutex:         &sync.RWMutex{},
		table:           make(map[string]*apPool),
		routingStrategy: routingStrategy,
	}
}

func (ap *availablePlugins) insert(pl *availablePlugin) error {
	if pl.pluginType != plugin.CollectorPluginType && pl.pluginType != plugin.ProcessorPluginType && pl.pluginType != plugin.PublisherPluginType {
		return ErrBadType
	}

	ap.Lock()
	defer ap.Unlock()

	key := fmt.Sprintf("%s:%s:%d", pl.TypeName(), pl.name, pl.version)
	_, exists := ap.table[key]
	if !exists {
		p, err := newPool(key, pl)
		if err != nil {
			return perror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}
		ap.table[key] = p
		return nil
	}
	ap.table[key].insert(pl)
	return nil
}

func (ap *availablePlugins) getPool(key string) (*apPool, perror.PulseError) {
	ap.RLock()
	defer ap.RUnlock()
	pool, ok := ap.table[key]
	if !ok {
		tnv := strings.Split(key, ":")
		if len(tnv) != 3 {
			return nil, perror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}

		v, err := strconv.Atoi(tnv[2])
		if err != nil {
			return nil, perror.New(ErrBadKey, map[string]interface{}{
				"key": key,
			})
		}

		if v < 1 {
			return ap.findLatestPool(tnv[0], tnv[1])
		}

		return nil, nil
	}

	return pool, nil
}

func (ap *availablePlugins) holdPool(key string) (*apPool, perror.PulseError) {
	pool, err := ap.getPool(key)
	if err != nil {
		return nil, err
	}

	if pool != nil {
		pool.RLock()
	}
	return pool, nil
}

func (ap *availablePlugins) findLatestPool(pType, name string) (*apPool, perror.PulseError) {
	// see if there exists a pool at all which matches name version.
	var latest *apPool
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
			if tnv[0] == pType && tnv[1] == name && pool.version > latest.version {
				latest = pool
			}
		}
		return latest, nil
	}

	return nil, nil
}

func (ap *availablePlugins) getOrCreatePool(key string) (*apPool, error) {
	var err error
	pool, ok := ap.table[key]
	if ok {
		return pool, nil
	}
	pool, err = newPool(key)
	if err != nil {
		return nil, err
	}
	ap.table[key] = pool
	return pool, nil
}

func (ap *availablePlugins) selectAP(key string) (*availablePlugin, perror.PulseError) {
	ap.RLock()
	defer ap.RUnlock()

	pool, err := ap.getPool(key)
	if err != nil {
		return nil, err
	}

	return pool.selectAP(ap.routingStrategy)
}

func (ap *availablePlugins) pools() map[string]*apPool {
	ap.RLock()
	defer ap.RUnlock()
	return ap.table
}

func (ap *availablePlugins) all() []*availablePlugin {
	var aps = []*availablePlugin{}
	ap.RLock()
	defer ap.RUnlock()
	for _, pool := range ap.table {
		for _, ap := range pool.plugins {
			aps = append(aps, ap)
		}
	}
	return aps
}
