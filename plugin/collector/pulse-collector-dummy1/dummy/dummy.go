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

package dummy

import (
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	// Name of plugin
	Name = "dummy1"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

// make sure that we actually satisify requierd interface
var _ plugin.CollectorPlugin = (*Dummy)(nil)

// Dummy collector implementation used for testing
type Dummy struct {
}

// CollectMetrics collects metrics for testing
func (f *Dummy) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := make([]plugin.PluginMetricType, len(mts))
	for i, p := range mts {
		var data string
		if cv, ok := p.Config().Table()["test"]; ok {
			data = fmt.Sprintf("The dummy collected data! config data: user=%s password=%s test=%v", p.Config().Table()["user"], p.Config().Table()["password"], cv.(ctypes.ConfigValueBool).Value)
		} else {
			data = fmt.Sprintf("The dummy collected data! config data: user=%s password=%s", p.Config().Table()["user"], p.Config().Table()["password"])
		}
		metrics[i] = plugin.PluginMetricType{
			Namespace_: p.Namespace(),
			Data_:      data,
			Timestamp_: time.Now(),
		}
		metrics[i].Source_, _ = os.Hostname()
	}
	return metrics, nil
}

//GetMetricTypes returns metric types for testing
func (f *Dummy) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	mts := []plugin.PluginMetricType{}
	if _, ok := cfg.Data.Table()["test-fail"]; ok {
		return mts, fmt.Errorf("missing on-load plugin config entry 'test'")
	}
	if _, ok := cfg.Data.Table()["test"]; ok {
		mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "dummy", "test"}})
	}
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "dummy", "foo"}})
	mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"intel", "dummy", "bar"}})
	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicyTree for testing
func (f *Dummy) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	rule, _ := cpolicy.NewStringRule("name", false, "bob")
	rule2, _ := cpolicy.NewStringRule("password", true)
	p := cpolicy.NewPolicyNode()
	p.Add(rule)
	p.Add(rule2)
	c.Add([]string{"intel", "dummy", "foo"}, p)
	return c, nil
}

//Meta returns meta data for testing
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType}, plugin.Unsecure(true))
}
