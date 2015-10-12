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

package client

import (
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// PluginClient A client providing common plugin method calls.
type PluginClient interface {
	SetKey() error
	Ping() error
	Kill(string) error
	GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
}

// PluginCollectorClient A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]core.Metric) ([]core.Metric, error)
	GetMetricTypes() ([]core.Metric, error)
}

// PluginProcessorClient A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error)
}

// PluginPublisherClient A client providing publishing specific plugin method calls.
type PluginPublisherClient interface {
	PluginClient
	Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error
}
