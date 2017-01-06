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

type MockProcessor struct {
	Meta PluginMeta
}

func (f *MockProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	return "", nil, nil
}

func (f *MockProcessor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, nil
}

type MockProcessorSessionState struct {
	PingTimeoutDuration time.Duration
	Daemon              bool
	listenAddress       string
	listenPort          string
	token               string
	logger              *log.Logger
	killChan            chan int
}

func (s *MockProcessorSessionState) Ping(arg PingArgs, b *bool) error {
	return nil
}

func (s *MockProcessorSessionState) Kill(arg KillArgs, b *bool) error {
	s.killChan <- 0
	return nil
}

func (s *MockProcessorSessionState) Logger() *log.Logger {
	return s.logger
}

func (s *MockProcessorSessionState) ListenAddress() string {
	return s.listenAddress
}

func (s *MockProcessorSessionState) ListenPort() string {
	return s.listenPort
}

func (s *MockProcessorSessionState) SetListenAddress(a string) {
	s.listenAddress = a
}

func (s *MockProcessorSessionState) Token() string {
	return s.token
}

func (m *MockProcessorSessionState) ResetHeartbeat() {

}

func (s *MockProcessorSessionState) KillChan() chan int {
	return s.killChan
}

func (s *MockProcessorSessionState) isDaemon() bool {
	return !s.Daemon
}

func (s *MockProcessorSessionState) generateResponse(r *Response) []byte {
	return []byte("mockResponse")
}

func (s *MockProcessorSessionState) heartbeatWatch(killChan chan int) {
	time.Sleep(time.Millisecond * 200)
	killChan <- 0
}

func TestStartProcessor(t *testing.T) {
	Convey("Processor", t, func() {
		Convey("start with dynamic port", func() {
			c := new(MockProcessor)
			m := &PluginMeta{
				RPCType: NativeRPC,
				Type:    ProcessorPluginType,
			}
			Convey("RPC service should not panic", func() {
				So(func() { Start(m, c, "{}") }, ShouldNotPanic)
			})

		})
	})
}
