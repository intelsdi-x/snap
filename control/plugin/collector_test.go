// +build legacy

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

package plugin

import (
	"testing"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	. "github.com/smartystreets/goconvey/convey"
)

type MockPlugin struct {
	Meta PluginMeta
}

func (f *MockPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, nil
}

func (f *MockPlugin) CollectMetrics(_ []MetricType) ([]MetricType, error) {
	return []MetricType{}, nil
}

func (c *MockPlugin) GetMetricTypes(_ ConfigType) ([]MetricType, error) {
	return []MetricType{
		{Namespace_: core.NewNamespace("foo", "bar")},
	}, nil
}

func TestStartCollector(t *testing.T) {
	Convey("Collector", t, func() {
		Convey("start with dynamic port", func() {
			m := &PluginMeta{
				RPCType: NativeRPC,
				Type:    CollectorPluginType,
			}
			c := new(MockPlugin)

			err, rc := Start(m, c, "{}")
			So(err, ShouldBeNil)
			So(rc, ShouldEqual, 0)

			Convey("RPC service should not panic", func() {
				So(func() { Start(m, c, "{}") }, ShouldNotPanic)
			})
		})
	})
}
