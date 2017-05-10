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

package client

import (
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

// PluginClient A client providing common plugin method calls.
type PluginClient interface {
	SetKey() error
	Ping() error
	Kill(string) error
	Close() error
	GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
}

// PluginCollectorClient A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	CollectMetrics([]core.Metric) ([]core.Metric, error)
	GetMetricTypes(plugin.ConfigType) ([]core.Metric, error)
}

type PluginStreamCollectorClient interface {
	PluginClient
	StreamMetrics([]core.Metric) (chan []core.Metric, chan error, error)
	GetMetricTypes(plugin.ConfigType) ([]core.Metric, error)
	UpdateCollectedMetrics([]core.Metric) error
	UpdatePluginConfig([]byte) error
	UpdateMetricsBuffer(int64) error
	UpdateCollectDuration(time.Duration) error
	Killed()
}

// PluginProcessorClient A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	Process([]core.Metric, map[string]ctypes.ConfigValue) ([]core.Metric, error)
}

// PluginPublisherClient A client providing publishing specific plugin method calls.
type PluginPublisherClient interface {
	PluginClient
	Publish([]core.Metric, map[string]ctypes.ConfigValue) error
}
