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
	"log"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

type MockPublisher struct {
	Meta PluginMeta
}

func (f *MockPublisher) Publish(_ string, _ []byte, _ map[string]ctypes.ConfigValue) error {
	return nil
}

func (f *MockPublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, nil
}

type MockPublisherSessionState struct {
	PingTimeoutDuration time.Duration
	Daemon              bool
	listenAddress       string
	listenPort          string
	token               string
	logger              *log.Logger
	killChan            chan int
}

func (s *MockPublisherSessionState) Ping(arg PingArgs, b *bool) error {
	return nil
}

func (s *MockPublisherSessionState) Kill(arg KillArgs, b *bool) error {
	s.killChan <- 0
	return nil
}

func (s *MockPublisherSessionState) Logger() *log.Logger {
	return s.logger
}

func (s *MockPublisherSessionState) ListenAddress() string {
	return s.listenAddress
}

func (s *MockPublisherSessionState) ListenPort() string {
	return s.listenPort
}

func (s *MockPublisherSessionState) SetListenAddress(a string) {
	s.listenAddress = a
}

func (s *MockPublisherSessionState) Token() string {
	return s.token
}

func (m *MockPublisherSessionState) ResetHeartbeat() {

}

func (s *MockPublisherSessionState) KillChan() chan int {
	return s.killChan
}

func (s *MockPublisherSessionState) isDaemon() bool {
	return s.Daemon
}

func (s *MockPublisherSessionState) generateResponse(r *Response) []byte {
	return []byte("mockResponse")
}

func (s *MockPublisherSessionState) heartbeatWatch(killChan chan int) {
	time.Sleep(time.Millisecond * 200)
	killChan <- 0
}

func TestStartPublisher(t *testing.T) {
	Convey("Publisher", t, func() {
		Convey("start with dynamic port", func() {
			c := new(MockPublisher)
			m := &PluginMeta{
				RPCType: NativeRPC,
				Type:    PublisherPluginType,
			}

			Convey("RPC service should not panic", func() {
				So(func() { Start(m, c, "{}") }, ShouldNotPanic)
			})

		})
	})
}
