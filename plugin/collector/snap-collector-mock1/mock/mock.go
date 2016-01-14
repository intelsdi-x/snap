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

package mock

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	// Name of plugin
	Name = "mock"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

// make sure that we actually satisify requierd interface
var _ plugin.CollectorPlugin = (*Mock)(nil)

// Mock collector implementation used for testing
type Mock struct {
}

// CollectMetrics collects metrics for testing
func (f *Mock) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := []plugin.PluginMetricType{}
	rand.Seed(time.Now().UTC().UnixNano())
	hostname, _ := os.Hostname()
	for i, p := range mts {
		if mts[i].Namespace()[2] == "*" {
			for j := 0; j < 10; j++ {
				v := fmt.Sprintf("host%d", j)
				data := randInt(65, 90)
				mt := plugin.PluginMetricType{
					Data_:      data,
					Namespace_: []string{"intel", "mock", v, "baz"},
					Source_:    hostname,
					Timestamp_: time.Now(),
					Labels_:    mts[i].Labels(),
					Version_:   mts[i].Version(),
				}
				metrics = append(metrics, mt)
			}
		} else {
			if cv, ok := p.Config().Table()["test"]; ok {
				p.Data_ = fmt.Sprintf("The mock collected data! config data: user=%s password=%s test=%v", p.Config().Table()["user"], p.Config().Table()["password"], cv.(ctypes.ConfigValueBool).Value)
			} else {
				p.Data_ = fmt.Sprintf("The mock collected data! config data: user=%s password=%s", p.Config().Table()["user"], p.Config().Table()["password"])
			}
			p.Timestamp_ = time.Now()
			p.Source_ = hostname
			metrics = append(metrics, p)
		}
	}
	return metrics, nil
}

//GetMetricTypes returns metric types for testing
func (f *Mock) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	mts := []plugin.PluginMetricType{}
	if _, ok := cfg.Table()["test-fail"]; ok {
		return mts, fmt.Errorf("missing on-load plugin config entry 'test'")
	}
	if _, ok := cfg.Table()["test"]; ok {
		mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "mock", "test"}})
	}
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "mock", "foo"}})
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "mock", "bar"}})
	mts = append(mts, plugin.PluginMetricType{
		Namespace_: []string{"intel", "mock", "*", "baz"},
		Labels_:    []core.Label{{Index: 2, Name: "host"}},
	})
	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicyTree for testing
func (f *Mock) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	rule, _ := cpolicy.NewStringRule("name", false, "bob")
	rule2, _ := cpolicy.NewStringRule("password", true)
	p := cpolicy.NewPolicyNode()
	p.Add(rule)
	p.Add(rule2)
	c.Add([]string{"intel", "mock", "foo"}, p)
	return c, nil
}

//Meta returns meta data for testing
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		Name,
		Version,
		Type,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Unsecure(true),
		plugin.RoutingStrategy(plugin.DefaultRouting),
		plugin.CacheTTL(1100*time.Millisecond),
	)
}

//Random number generator
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
