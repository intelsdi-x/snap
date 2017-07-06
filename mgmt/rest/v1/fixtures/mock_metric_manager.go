// +build legacy small medium large

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
	"errors"
	"fmt"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
)

var pluginCatalog []core.CatalogedPlugin = []core.CatalogedPlugin{
	MockLoadedPlugin{MyName: "foo", MyType: "collector", MyVersion: 2},
	MockLoadedPlugin{MyName: "bar", MyType: "publisher", MyVersion: 3},
	MockLoadedPlugin{MyName: "foo", MyType: "collector", MyVersion: 4},
	MockLoadedPlugin{MyName: "baz", MyType: "publisher", MyVersion: 5},
	MockLoadedPlugin{MyName: "foo", MyType: "processor", MyVersion: 6},
	MockLoadedPlugin{MyName: "foobar", MyType: "processor", MyVersion: 1},
}

var metricCatalog []core.CatalogedMetric = []core.CatalogedMetric{
	MockCatalogedMetric{},
}

//////MockLoadedPlugin/////

type MockLoadedPlugin struct {
	MyName    string
	MyType    string
	MyVersion int
}

func (m MockLoadedPlugin) Name() string     { return m.MyName }
func (m MockLoadedPlugin) Port() string     { return "" }
func (m MockLoadedPlugin) TypeName() string { return m.MyType }
func (m MockLoadedPlugin) Version() int     { return m.MyVersion }
func (m MockLoadedPlugin) Key() string {
	return fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", m.MyType, m.MyName, m.MyVersion)
}
func (m MockLoadedPlugin) Plugin() string     { return "" }
func (m MockLoadedPlugin) IsSigned() bool     { return false }
func (m MockLoadedPlugin) Status() string     { return "" }
func (m MockLoadedPlugin) PluginPath() string { return "" }
func (m MockLoadedPlugin) LoadedTimestamp() *time.Time {
	t := time.Date(2016, time.September, 6, 0, 0, 0, 0, time.UTC)
	return &t
}
func (m MockLoadedPlugin) Policy() *cpolicy.ConfigPolicy { return cpolicy.New() }
func (m MockLoadedPlugin) HitCount() int                 { return 0 }
func (m MockLoadedPlugin) LastHit() time.Time            { return time.Now() }
func (m MockLoadedPlugin) ID() uint32                    { return 0 }

//////MockCatalogedMetric/////

type MockCatalogedMetric struct{}

func (m MockCatalogedMetric) Namespace() core.Namespace {
	return core.NewNamespace("one", "two", "three")
}
func (m MockCatalogedMetric) Version() int                      { return 5 }
func (m MockCatalogedMetric) LastAdvertisedTime() time.Time     { return time.Time{} }
func (m MockCatalogedMetric) Policy() *cpolicy.ConfigPolicyNode { return cpolicy.NewPolicyNode() }
func (m MockCatalogedMetric) Description() string               { return "This Is A Description" }
func (m MockCatalogedMetric) Unit() string                      { return "" }

//////MockManagesMetrics/////

type MockManagesMetrics struct{}

func (m MockManagesMetrics) MetricCatalog() ([]core.CatalogedMetric, error) {
	return metricCatalog, nil
}
func (m MockManagesMetrics) FetchMetrics(core.Namespace, int) ([]core.CatalogedMetric, error) {
	return metricCatalog, nil
}
func (m MockManagesMetrics) GetMetricVersions(core.Namespace) ([]core.CatalogedMetric, error) {
	return metricCatalog, nil
}
func (m MockManagesMetrics) GetMetric(core.Namespace, int) (core.CatalogedMetric, error) {
	return MockCatalogedMetric{}, nil
}
func (m MockManagesMetrics) Load(*core.RequestedPlugin) (core.CatalogedPlugin, serror.SnapError) {
	return MockLoadedPlugin{"foo", "collector", 1}, nil
}
func (m MockManagesMetrics) Unload(plugin core.Plugin) (core.CatalogedPlugin, serror.SnapError) {
	for _, pl := range pluginCatalog {
		if plugin.Name() == pl.Name() &&
			plugin.Version() == pl.Version() &&
			plugin.TypeName() == pl.TypeName() {
			return pl, nil
		}
	}
	return nil, serror.New(errors.New("plugin not found"))
}

func (m MockManagesMetrics) PluginCatalog() core.PluginCatalog {
	return pluginCatalog
}
func (m MockManagesMetrics) AvailablePlugins() []core.AvailablePlugin {
	return []core.AvailablePlugin{
		MockLoadedPlugin{MyName: "foo", MyType: "collector", MyVersion: 2},
		MockLoadedPlugin{MyName: "bar", MyType: "publisher", MyVersion: 3},
		MockLoadedPlugin{MyName: "foo", MyType: "collector", MyVersion: 4},
		MockLoadedPlugin{MyName: "baz", MyType: "publisher", MyVersion: 5},
		MockLoadedPlugin{MyName: "foo", MyType: "processor", MyVersion: 6},
		MockLoadedPlugin{MyName: "foobar", MyType: "processor", MyVersion: 1},
	}
}
func (m MockManagesMetrics) GetAutodiscoverPaths() []string {
	return nil
}

func (m MockManagesMetrics) GetTempDir() string {
	return ""
}

// These constants are the expected plugin responses from running
// rest_v1_test.go on the plugin routes found in mgmt/rest/server.go
const (
	GET_PLUGINS_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Plugin list returned",
    "type": "plugin_list_returned",
    "version": 1
  },
  "body": {
    "loaded_plugins": [
      {
        "name": "foo",
        "version": 2,
        "type": "collector",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/collector/foo/2"
      },
      {
        "name": "bar",
        "version": 3,
        "type": "publisher",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/publisher/bar/3"
      },
      {
        "name": "foo",
        "version": 4,
        "type": "collector",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/collector/foo/4"
      },
      {
        "name": "baz",
        "version": 5,
        "type": "publisher",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/publisher/baz/5"
      },
      {
        "name": "foo",
        "version": 6,
        "type": "processor",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/processor/foo/6"
      },
      {
        "name": "foobar",
        "version": 1,
        "type": "processor",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/processor/foobar/1"
      }
    ]
  }
}`

	GET_PLUGINS_RESPONSE_TYPE = `{
  "meta": {
    "code": 200,
    "message": "Plugin list returned",
    "type": "plugin_list_returned",
    "version": 1
  },
  "body": {
    "loaded_plugins": [
      {
        "name": "foo",
        "version": 2,
        "type": "collector",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/collector/foo/2"
      },
      {
        "name": "foo",
        "version": 4,
        "type": "collector",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/collector/foo/4"
      }
    ]
  }
}`

	GET_PLUGINS_RESPONSE_TYPE_NAME = `{
  "meta": {
    "code": 200,
    "message": "Plugin list returned",
    "type": "plugin_list_returned",
    "version": 1
  },
  "body": {
    "loaded_plugins": [
      {
        "name": "bar",
        "version": 3,
        "type": "publisher",
        "signed": false,
        "status": "",
        "loaded_timestamp": 1473120000,
        "href": "http://localhost:%d/v1/plugins/publisher/bar/3"
      }
    ]
  }
}`

	GET_PLUGINS_RESPONSE_TYPE_NAME_VERSION = `{
  "meta": {
    "code": 200,
    "message": "Plugin returned",
    "type": "plugin_returned",
    "version": 1
  },
  "body": {
    "name": "bar",
    "version": 3,
    "type": "publisher",
    "signed": false,
    "status": "",
    "loaded_timestamp": 1473120000,
    "href": "http://localhost:%d/v1/plugins/publisher/bar/3"
  }
}`

	GET_METRICS_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Metrics returned",
    "type": "metrics_returned",
    "version": 1
  },
  "body": [
    {
      "last_advertised_timestamp": -62135596800,
      "namespace": "/one/two/three",
      "version": 5,
      "dynamic": false,
      "description": "This Is A Description",
      "href": "http://localhost:%d/v1/metrics?ns=/one/two/three&ver=5"
    }
  ]
}`

	UNLOAD_PLUGIN_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Plugin successfully unloaded (foov2)",
    "type": "plugin_unloaded",
    "version": 1
  },
  "body": {
    "name": "foo",
    "version": 2,
    "type": "collector"
  }
}`
)
