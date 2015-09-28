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
	"log"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	. "github.com/smartystreets/goconvey/convey"
)

type MockSessionState struct {
	PingTimeoutDuration time.Duration
	Daemon              bool
	listenAddress       string
	listenPort          string
	token               string
	logger              *log.Logger
	killChan            chan int
}

func (s *MockSessionState) Ping(arg PingArgs, b *bool) error {
	return nil
}

func (s *MockSessionState) Kill(arg KillArgs, b *bool) error {
	s.killChan <- 0
	return nil
}

func (s *MockSessionState) Logger() *log.Logger {
	return s.logger
}

func (s *MockSessionState) ListenAddress() string {
	return s.listenAddress
}

func (s *MockSessionState) ListenPort() string {
	return s.listenPort
}

func (s *MockSessionState) SetListenAddress(a string) {
	s.listenAddress = a
}

func (s *MockSessionState) Token() string {
	return s.token
}

func (m *MockSessionState) ResetHeartbeat() {

}

func (s *MockSessionState) KillChan() chan int {
	return s.killChan
}

func (s *MockSessionState) isDaemon() bool {
	return s.Daemon
}

func (s *MockSessionState) generateResponse(r *Response) []byte {
	return []byte("mockResponse")
}

func (s *MockSessionState) heartbeatWatch(killChan chan int) {
	time.Sleep(time.Millisecond * 200)
	killChan <- 0
}

type MockPlugin struct {
	Meta PluginMeta
}

func (f *MockPlugin) GetConfigPolicy() (cpolicy.ConfigPolicy, error) {
	return cpolicy.ConfigPolicy{}, nil
}

func (f *MockPlugin) CollectMetrics(_ []PluginMetricType) ([]PluginMetricType, error) {
	return []PluginMetricType{}, nil
}

func (c *MockPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return []PluginMetricType{
		PluginMetricType{Namespace_: []string{"foo", "bar"}},
	}, nil
}

func TestStartCollector(t *testing.T) {
	Convey("Collector", t, func() {
		Convey("start with dynamic port", func() {
			m := &PluginMeta{
				RPCType: JSONRPC,
				Type:    CollectorPluginType,
			}
			c := new(MockPlugin)
			err, rc := Start(m, c, "{}")
			So(err, ShouldBeNil)
			So(rc, ShouldEqual, 0)
			Convey("RPC service already registered", func() {
				So(func() { Start(m, c, "{}") }, ShouldPanic)
			})
		})
	})
}
