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

package strategy

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
)

var (
	// This defines the maximum running instances of a loaded plugin.
	// It is initialized at runtime via the cli.
	MaximumRunningPlugins = 3
)

var (
	ErrBadType     = errors.New("bad plugin type")
	ErrBadStrategy = errors.New("bad strategy")
	ErrPoolEmpty   = errors.New("plugin pool is empty")
)

type Pool interface {
	RoutingAndCaching
	Count() int
	Eligible() bool
	Insert(a AvailablePlugin) error
	Kill(id uint32, reason string)
	Plugins() MapAvailablePlugin
	RLock()
	RUnlock()
	SelectAndStop(taskID, reason string)
	SelectAndKill(taskID, reason string)
	SelectAP(taskID string, configID map[string]ctypes.ConfigValue) (AvailablePlugin, serror.SnapError)
	Strategy() RoutingAndCaching
	Subscribe(taskID string)
	SubscriptionCount() int
	Unsubscribe(taskID string)
	Version() int
	RestartCount() int
	IncRestartCount()
	KillAll(string)
}

type AvailablePlugin interface {
	core.AvailablePlugin
	CacheTTL() time.Duration
	CheckHealth()
	ConcurrencyCount() int
	Exclusive() bool
	Kill(r string) error
	RoutingStrategy() plugin.RoutingStrategyType
	SetID(id uint32)
	String() string
	Type() plugin.PluginType
	Stop(string) error
	IsRemote() bool
	SetIsRemote(bool)
}

type subscription struct {
	Version int
	TaskID  string
}

type pool struct {
	// used to coordinate changes to a pool
	*sync.RWMutex

	// the version of the plugins in the pool.
	// subscriptions uses this.
	version int
	// key is the primary key used in availablePlugins:
	// {plugin_type}:{plugin_name}:{plugin_version}
	key string

	// The subscriptions to this pool.
	subs map[string]*subscription

	// The plugins in the pool.
	// the primary key is an increasing --> uint from
	// snapteld epoch (`service snapteld start`).
	plugins    MapAvailablePlugin
	pidCounter uint32

	// The max size which this pool may grow.
	max int

	// The number of subscriptions per running instance
	concurrencyCount int

	// The routing and caching strategy declared by the plugin.
	// strategy RoutingAndCaching
	RoutingAndCaching

	// restartCount the restart count of available plugins
	// when the DeadAvailablePluginEvent occurs
	restartCount int
}

func NewPool(key string, plugins ...AvailablePlugin) (Pool, error) {
	versl := strings.Split(key, core.Separator)
	ver, err := strconv.Atoi(versl[len(versl)-1])
	if err != nil {
		return nil, err
	}
	p := &pool{
		RWMutex:          &sync.RWMutex{},
		version:          ver,
		key:              key,
		subs:             map[string]*subscription{},
		plugins:          MapAvailablePlugin{},
		max:              MaximumRunningPlugins,
		concurrencyCount: 1,
	}

	if len(plugins) > 0 {
		for _, plg := range plugins {
			p.Insert(plg)
		}
	}

	return p, nil
}

// Version returns the version
func (p *pool) Version() int {
	return p.version
}

// Plugins returns a map of plugin ids to the AvailablePlugin
func (p *pool) Plugins() MapAvailablePlugin {
	return p.plugins
}

// Strategy returns the routing strategy
func (p *pool) Strategy() RoutingAndCaching {
	return p.RoutingAndCaching
}

// RestartCount returns the restart count of a pool
func (p *pool) RestartCount() int {
	return p.restartCount
}

func (p *pool) IncRestartCount() {
	p.Lock()
	defer p.Unlock()
	p.restartCount++
}

// Insert inserts an AvailablePlugin into the pool
func (p *pool) Insert(a AvailablePlugin) error {
	if a.Type() != plugin.CollectorPluginType && a.Type() != plugin.ProcessorPluginType && a.Type() != plugin.PublisherPluginType && a.Type() != plugin.StreamCollectorPluginType {
		return ErrBadType
	}
	// If an empty pool is created, it does not have
	// any available plugins from which to retrieve
	// concurrency count or exclusivity.  We ensure it
	// is set correctly on an insert.
	if len(p.plugins) == 0 {
		if err := p.applyPluginMeta(a); err != nil {
			return err
		}
	}

	a.SetID(p.generatePID())
	p.plugins[a.ID()] = a
	return nil
}

// applyPluginMeta is called when the first plugin is added to the pool
func (p *pool) applyPluginMeta(a AvailablePlugin) error {
	// Checking if plugin is exclusive
	// (only one instance should be running).
	if a.Exclusive() {
		p.max = 1
	}

	// Set the cache TTL
	cacheTTL := GlobalCacheExpiration
	// if the plugin exposes a default TTL that is greater the the global default use it
	if a.CacheTTL() != 0 && a.CacheTTL() > GlobalCacheExpiration {
		cacheTTL = a.CacheTTL()
	}

	// Set the concurrency count
	p.concurrencyCount = a.ConcurrencyCount()

	// Set the routing and caching strategy
	switch a.RoutingStrategy() {
	case plugin.DefaultRouting:
		p.RoutingAndCaching = NewLRU(cacheTTL)
	case plugin.StickyRouting:
		p.RoutingAndCaching = NewSticky(cacheTTL)
		p.concurrencyCount = 1
	case plugin.ConfigRouting:
		p.RoutingAndCaching = NewConfigBased(cacheTTL)
	default:
		return ErrBadStrategy
	}

	return nil
}

// subscribe adds a subscription to the pool.
// Using subscribe is idempotent.
func (p *pool) Subscribe(taskID string) {
	p.Lock()
	defer p.Unlock()

	if _, exists := p.subs[taskID]; !exists {
		// Version is the last item in the key, so we split here
		// to retrieve it for the subscription.
		p.subs[taskID] = &subscription{
			TaskID:  taskID,
			Version: p.version,
		}
	}
}

// unsubscribe removes a subscription from the pool.
// Using unsubscribe is idempotent.
func (p *pool) Unsubscribe(taskID string) {
	p.Lock()
	defer p.Unlock()
	delete(p.subs, taskID)
}

// Eligible returns a bool indicating whether the pool is eligible to grow
func (p *pool) Eligible() bool {
	p.RLock()
	defer p.RUnlock()

	// optimization: don't even bother with concurrency
	// count if we have already reached pool max
	if len(p.plugins) >= p.max {
		return false
	}

	// Check if pool is eligible and number of plugins is less than maximum allowed
	if len(p.subs) > p.concurrencyCount*len(p.plugins) {
		return true
	}

	return false
}

// kill kills and removes the available plugin from its pool.
// Using kill is idempotent.
func (p *pool) Kill(id uint32, reason string) {
	p.Lock()
	defer p.Unlock()

	ap, ok := p.plugins[id]
	if ok {
		ap.Kill(reason)
		delete(p.plugins, id)
	}
}

// Kill all instances of a plugin
func (p *pool) KillAll(reason string) {
	for id, rp := range p.plugins {
		log.WithFields(log.Fields{
			"_block": "KillAll",
			"reason": reason,
		}).Debug(fmt.Sprintf("handling 'KillAll' for pool '%v', killing plugin '%v:%v'", p.String(), rp.Name(), rp.Version()))
		if err := rp.Stop(reason); err != nil {
			log.WithFields(log.Fields{
				"_block": "KillAll",
				"reason": reason,
			}).Error(err)
		}
		p.Kill(id, reason)
	}
}

// SelectAndStop selects, stops and removes the available plugin from the pool
func (p *pool) SelectAndStop(id, reason string) {
	rp, err := p.Remove(p.plugins.Values(), id)
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "SelectAndStop",
			"taskID": id,
			"reason": reason,
		}).Error(err)
		return
	}
	if err := rp.Stop(reason); err != nil {
		log.WithFields(log.Fields{
			"_block": "SelectAndStop",
			"taskID": id,
			"reason": reason,
		}).Error(err)
	}
	p.remove(rp.ID())
}

// SelectAndKill selects, kills and removes the available plugin from the pool
func (p *pool) SelectAndKill(id, reason string) {
	rp, err := p.Remove(p.plugins.Values(), id)
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "SelectAndKill",
			"taskID": id,
			"reason": reason,
		}).Error(err)
		return
	}
	if err := rp.Stop(reason); err != nil {
		log.WithFields(log.Fields{
			"_block": "SelectAndKill",
			"taskID": id,
			"reason": reason,
		}).Error(err)
	}
	if err := rp.Kill(reason); err != nil {
		log.WithFields(log.Fields{
			"_block": "SelectAndKill",
			"taskID": id,
			"reason": reason,
		}).Error(err)
	}
	p.remove(rp.ID())
}

// remove removes an available plugin from the the pool.
// using remove is idempotent.
func (p *pool) remove(id uint32) {
	p.Lock()
	defer p.Unlock()
	delete(p.plugins, id)
}

// Count returns the number of plugins in the pool
func (p *pool) Count() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.plugins)
}

// NOTE: The data returned by subscriptions should be constant and read only.
func (p *pool) subscriptions() map[string]*subscription {
	p.RLock()
	defer p.RUnlock()
	return p.subs
}

// SubscriptionCount returns the number of subscriptions in the pool
func (p *pool) SubscriptionCount() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.subs)
}

// SelectAP selects an available plugin from the pool
// the method is not thread safe, it should be protected outside of the body
func (p *pool) SelectAP(taskID string, config map[string]ctypes.ConfigValue) (AvailablePlugin, serror.SnapError) {
	aps := p.plugins.Values()

	var id string
	switch p.Strategy().String() {
	case "least-recently-used":
		id = ""
	case "sticky":
		id = taskID
	case "config-based":
		id = idFromCfg(config)
	default:
		return nil, serror.New(ErrBadStrategy)
	}

	ap, err := p.Select(aps, id)
	if err != nil {
		return nil, serror.New(err)
	}
	return ap, nil
}

func idFromCfg(cfg map[string]ctypes.ConfigValue) string {
	//TODO: check for nil map
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(cfg)
	if err != nil {
		return ""
	}
	return string(buff.Bytes())
}

// generatePID returns the next available pid for the pool
func (p *pool) generatePID() uint32 {
	atomic.AddUint32(&p.pidCounter, 1)
	return p.pidCounter
}

// CacheTTL returns the cacheTTL for the pool
func (p *pool) CacheTTL(taskID string) (time.Duration, error) {
	if len(p.plugins) == 0 {
		return 0, ErrPoolEmpty
	}
	return p.RoutingAndCaching.CacheTTL(taskID)
}
