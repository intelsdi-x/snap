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
	"encoding/json"
	"fmt"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/plugin/helper"
)

var (
	SnapPath = helper.BuildPath

	PluginNameMock1 = "snap-plugin-collector-mock1"
	PluginPathMock1 = helper.PluginFilePath(PluginNameMock1)

	PluginNameMock2 = "snap-plugin-collector-mock2"
	PluginPathMock2 = helper.PluginFilePath(PluginNameMock2)

	PluginNameStreamingRand1 = "snap-plugin-streaming-collector-rand1"
	PluginPathStreamingRand1 = helper.PluginFilePath(PluginNameStreamingRand1)

	PluginNameMock2Grpc = "snap-plugin-collector-mock2-grpc"
	PluginPathMock2Grpc = helper.PluginFilePath(PluginNameMock2Grpc)
	PluginUriMock2Grpc  = "http://127.0.0.1:8183"
)

// mocks a metric type
type MockMetricType struct {
	Namespace_ core.Namespace
	Cfg        *cdata.ConfigDataNode
	Ver        int
}

func (m MockMetricType) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Namespace core.Namespace        `json:"namespace"`
		Config    *cdata.ConfigDataNode `json:"config"`
	}{
		Namespace: m.Namespace_,
		Config:    m.Cfg,
	})
}

func (m MockMetricType) Namespace() core.Namespace {
	return m.Namespace_
}

func (m MockMetricType) Description() string {
	return ""
}

func (m MockMetricType) Unit() string {
	return ""
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Timestamp() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return m.Ver
}

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return m.Cfg
}

func (m MockMetricType) Data() interface{} {
	return nil
}

func (m MockMetricType) Tags() map[string]string { return nil }

var ValidMetric = MockMetricType{
	Namespace_: core.NewNamespace([]string{"intel", "mock", "foo"}...),
	Cfg:        cdata.NewNode(),
	Ver:        0,
}
var InvalidMetric = MockMetricType{
	Namespace_: core.NewNamespace([]string{"this", "is", "invalid"}...),
	Cfg:        cdata.NewNode(),
	Ver:        1000,
}

// mocks a metric
type mockMetric struct {
	namespace []string
	data      int
}

func (m *mockMetric) Namespace() []string {
	return m.namespace
}

func (m *mockMetric) Data() interface{} {
	return m.data
}

func NewMockPlugin(plgType core.PluginType, name string, version int) MockPlugin {
	return MockPlugin{
		pluginType: plgType,
		name:       name,
		ver:        version,
	}
}

// mocks a plugin
type MockPlugin struct {
	pluginType core.PluginType
	name       string
	ver        int
	config     *cdata.ConfigDataNode
}

func (m MockPlugin) Name() string                  { return m.name }
func (m MockPlugin) TypeName() string              { return m.pluginType.String() }
func (m MockPlugin) Version() int                  { return m.ver }
func (m MockPlugin) Config() *cdata.ConfigDataNode { return m.config }
func (m MockPlugin) Key() string {
	return fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d", m.pluginType.String(), m.name, m.ver)
}

type MockRequestedMetric struct {
	namespace core.Namespace
	version   int
}

func NewMockRequestedMetric(ns core.Namespace, ver int) MockRequestedMetric {
	return MockRequestedMetric{namespace: ns, version: ver}
}

func (m MockRequestedMetric) Version() int {
	return m.version
}

func (m MockRequestedMetric) Namespace() core.Namespace {
	return m.namespace
}
