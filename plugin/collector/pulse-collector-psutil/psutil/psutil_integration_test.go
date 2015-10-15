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

package psutil

import (
	"runtime"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/cdata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPsutilCollectMetrics(t *testing.T) {
	Convey("psutil collector", t, func() {
		p := &Psutil{}
		Convey("collect metrics", func() {
			mts := []plugin.PluginMetricType{
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load1"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load5"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load15"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "vm", "total"},
				},
			}
			if runtime.GOOS != "darwin" {
				mts = append(mts, plugin.PluginMetricType{
					Namespace_: []string{"psutil", "cpu0", "user"},
				})
			}
			metrics, err := p.CollectMetrics(mts)
			//prettyPrint(metrics)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
		})
		Convey("get metric types", func() {
			mts, err := p.GetMetricTypes(plugin.PluginConfigType{cdata.NewNode()})
			//prettyPrint(mts)
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
		})

	})
}
