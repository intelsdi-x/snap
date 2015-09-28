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

package plugin

import (
	"encoding/gob"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// Acts as a proxy for RPC calls to a CollectorPlugin. This helps keep the function signature simple
// within plugins vs. having to match required RPC patterns.

// Collector plugin
type CollectorPlugin interface {
	Plugin
	CollectMetrics([]PluginMetricType) ([]PluginMetricType, error)
	GetMetricTypes() ([]PluginMetricType, error)
	GetConfigPolicy() (cpolicy.ConfigPolicy, error)
}

func init() {
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
	gob.Register(&cpolicy.IntRule{})
	gob.Register(&cpolicy.FloatRule{})
}
