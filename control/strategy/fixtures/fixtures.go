/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

package fixtures

import (
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
)

const (
	hitcount           = 43
	id                 = 1234
	cacheTTL           = time.Second
	ccount             = 1
	excl               = true
	nonexlc            = false
	configBasedRouting = plugin.ConfigRouting
	stickyRouting      = plugin.StickyRouting
	lruRouting         = plugin.DefaultRouting
	collectorType      = plugin.CollectorPluginType
	publisherType      = plugin.PublisherPluginType
	processorType      = plugin.ProcessorPluginType
	version            = 1
	name               = "mock"
)

var lastHit = time.Unix(1460027570, 0)

type MockAvailablePlugin struct {
	pluginName string
	hitCount   int
	lastHit    time.Time
	id         uint32
	ttl        time.Duration
	concount   int
	exclusive  bool
	strategy   plugin.RoutingStrategyType
	pluginType plugin.PluginType
	version    int
}

func NewMockAvailablePlugin() *MockAvailablePlugin {
	mock := &MockAvailablePlugin{
		pluginName: name,
		hitCount:   hitcount,
		lastHit:    lastHit,
		id:         id,
		ttl:        cacheTTL,
		exclusive:  nonexlc,
		concount:   ccount,
		strategy:   lruRouting,
		pluginType: plugin.CollectorPluginType,
		version:    version,
	}
	return mock
}

func (m *MockAvailablePlugin) WithName(name string) *MockAvailablePlugin {
	m.pluginName = name
	return m
}

func (m *MockAvailablePlugin) WithHitCount(count int) *MockAvailablePlugin {
	m.hitCount = count
	return m
}

func (m *MockAvailablePlugin) WithLastHit(last time.Time) *MockAvailablePlugin {
	m.lastHit = last
	return m
}

func (m *MockAvailablePlugin) WithID(id uint32) *MockAvailablePlugin {
	m.id = id
	return m
}

func (m *MockAvailablePlugin) WithTTL(ttl time.Duration) *MockAvailablePlugin {
	m.ttl = ttl
	return m
}

func (m *MockAvailablePlugin) WithConCount(count int) *MockAvailablePlugin {
	m.concount = count
	return m
}

func (m *MockAvailablePlugin) WithStrategy(strategy plugin.RoutingStrategyType) *MockAvailablePlugin {
	m.strategy = strategy
	return m
}

func (m *MockAvailablePlugin) WithPluginType(plgType plugin.PluginType) *MockAvailablePlugin {
	m.pluginType = plgType
	return m
}

func (m *MockAvailablePlugin) WithExclusive(excl bool) *MockAvailablePlugin {
	m.exclusive = excl
	return m
}

func (m *MockAvailablePlugin) WithVersion(ver int) *MockAvailablePlugin {
	m.version = ver
	return m
}

func (m MockAvailablePlugin) HitCount() int {
	return m.hitCount
}

func (m MockAvailablePlugin) LastHit() time.Time {
	return m.lastHit
}

func (m MockAvailablePlugin) String() string {
	return strings.Join([]string{m.pluginType.String(), m.pluginName, strconv.Itoa(m.Version())}, core.Separator)
}

func (m MockAvailablePlugin) Kill(string) error {
	return nil
}

func (m MockAvailablePlugin) Stop(string) error {
	return nil
}

func (m MockAvailablePlugin) ID() uint32 {
	return m.id
}

func (m MockAvailablePlugin) CacheTTL() time.Duration {
	return m.ttl
}

func (m MockAvailablePlugin) CheckHealth() {}

func (m MockAvailablePlugin) ConcurrencyCount() int {
	return m.concount
}

func (m MockAvailablePlugin) Exclusive() bool {
	return m.exclusive
}

func (m MockAvailablePlugin) RoutingStrategy() plugin.RoutingStrategyType {
	return m.strategy
}

func (m MockAvailablePlugin) SetID(id uint32) {
	m.id = id
}

func (m MockAvailablePlugin) Type() plugin.PluginType {
	return m.pluginType
}

func (m MockAvailablePlugin) TypeName() string {
	return m.pluginType.String()
}

func (m MockAvailablePlugin) Name() string {
	return m.pluginName
}

func (m MockAvailablePlugin) Version() int {
	return m.version
}
