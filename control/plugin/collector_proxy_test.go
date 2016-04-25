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
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	"github.com/intelsdi-x/snap/core"

	. "github.com/smartystreets/goconvey/convey"
)

type mockPlugin struct {
}

var mockMetricType = []MetricType{
	*NewMetricType(core.NewNamespace([]string{"foo"}).AddDynamicElement("test", "something dynamic here").AddStaticElement("bar"), time.Now(), nil, "", 1),
	*NewMetricType(core.NewNamespace([]string{"foo", "baz"}), time.Now(), nil, "", 2),
}

func (p *mockPlugin) GetMetricTypes(cfg ConfigType) ([]MetricType, error) {
	return mockMetricType, nil
}

func (p *mockPlugin) CollectMetrics(mockMetricTypes []MetricType) ([]MetricType, error) {
	for i := range mockMetricTypes {
		if mockMetricTypes[i].Namespace().String() == "/foo/*/bar" {
			mockMetricTypes[i].Namespace_[1].Value = "test"
		}
	}
	return mockMetricTypes, nil
}

func (p *mockPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	cpn := cpolicy.NewPolicyNode()
	r1, _ := cpolicy.NewStringRule("username", false, "root")
	r2, _ := cpolicy.NewStringRule("password", true)
	cpn.Add(r1, r2)
	ns := []string{"one", "two", "potato"}
	cp.Add(ns, cpn)
	cp.Freeze()

	return cp, nil
}

type mockErrorPlugin struct {
}

func (p *mockErrorPlugin) GetMetricTypes(cfg ConfigType) ([]MetricType, error) {
	return nil, errors.New("Error in get Metric Type")
}

func (p *mockErrorPlugin) CollectMetrics(_ []MetricType) ([]MetricType, error) {
	return nil, errors.New("Error in collect Metric")
}

func (p *mockErrorPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, errors.New("Error in get config policy")
}

func TestCollectorProxy(t *testing.T) {
	Convey("Test collector plugin proxy for get metric types ", t, func() {

		logger := log.New(os.Stdout,
			"test: ",
			log.Ldate|log.Ltime|log.Lshortfile)
		mockPlugin := &mockPlugin{}

		mockSessionState := &MockSessionState{
			Encoder:             encoding.NewGobEncoder(),
			listenPort:          "0",
			token:               "abcdef",
			logger:              logger,
			PingTimeoutDuration: time.Millisecond * 100,
			killChan:            make(chan int),
		}
		c := &collectorPluginProxy{
			Plugin:  mockPlugin,
			Session: mockSessionState,
		}
		Convey("Get Metric Types", func() {
			var reply []byte
			c.GetMetricTypes([]byte{}, &reply)
			var mtr GetMetricTypesReply
			err := c.Session.Decode(reply, &mtr)
			So(err, ShouldBeNil)
			So(mtr.MetricTypes[0].Namespace().String(), ShouldResemble, "/foo/*/bar")

		})
		Convey("Get error in Get Metric Type", func() {
			mockErrorPlugin := &mockErrorPlugin{}
			errC := &collectorPluginProxy{
				Plugin:  mockErrorPlugin,
				Session: mockSessionState,
			}
			var reply []byte
			err := errC.GetMetricTypes([]byte{}, &reply)
			So(err.Error(), ShouldResemble, "GetMetricTypes call error : Error in get Metric Type")
		})
		Convey("Collect Metric ", func() {
			args := CollectMetricsArgs{
				MetricTypes: mockMetricType,
			}
			out, err := c.Session.Encode(args)
			So(err, ShouldBeNil)
			var reply []byte
			c.CollectMetrics(out, &reply)
			var mtr CollectMetricsReply
			err = c.Session.Decode(reply, &mtr)
			So(mtr.PluginMetrics[0].Namespace().String(), ShouldResemble, "/foo/test/bar")
			So(mtr.PluginMetrics[0].Namespace()[1].Name, ShouldEqual, "test")

			Convey("Get error in Collect Metric ", func() {
				args := CollectMetricsArgs{
					MetricTypes: mockMetricType,
				}
				mockErrorPlugin := &mockErrorPlugin{}
				errC := &collectorPluginProxy{
					Plugin:  mockErrorPlugin,
					Session: mockSessionState,
				}
				out, err := errC.Session.Encode(args)
				So(err, ShouldBeNil)
				var reply []byte
				err = errC.CollectMetrics(out, &reply)
				So(err, ShouldNotBeNil)
			})

		})

	})

}
