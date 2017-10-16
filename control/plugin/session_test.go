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
	"encoding/json"
	"errors"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	. "github.com/smartystreets/goconvey/convey"
)

type MockSessionState struct {
	encoding.Encoder

	PingTimeoutDuration time.Duration
	Daemon              bool
	listenAddress       string
	listenPort          string
	token               string
	logger              *log.Logger
	killChan            chan int
}

func (s *MockSessionState) Ping(arg []byte, reply *[]byte) error {
	return nil
}

func (s *MockSessionState) Kill(arg []byte, reply *[]byte) error {
	s.killChan <- 0
	return nil
}

func (s *MockSessionState) GetConfigPolicy(arg []byte, reply *[]byte) error {
	out := GetConfigPolicyReply{Policy: cpolicy.New()}
	*reply, _ = s.Encode(out)
	return nil
}

func (s *MockSessionState) SetKey(SetKeyArgs, *[]byte) error { return nil }

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

func (s *MockSessionState) setKey(key []byte) {
}

func (s *MockSessionState) DecryptKey(in []byte) ([]byte, error) {
	return []byte{}, nil
}

type errSessionState struct {
	*MockSessionState
}

func (e *errSessionState) GetConfigPolicy(arg []byte, reply *[]byte) error {
	return errors.New("GetConfigPolicy call error : Error in get config policy")
}

func TestSessionState(t *testing.T) {
	Convey("SessionState", t, func() {
		now := time.Now()
		ss := &SessionState{
			LastPing: now,
			Arg:      &Arg{PingTimeoutDuration: 500 * time.Millisecond},
			Encoder:  encoding.NewJsonEncoder(),
		}
		ss.logger = log.New()
		Convey("Ping", func() {

			ss.Ping([]byte{}, &[]byte{})
			So(ss.LastPing.Nanosecond(), ShouldBeGreaterThan, now.Nanosecond())
		})
		Convey("Kill", func() {
			args := KillArgs{Reason: "testing"}
			out, err := ss.Encode(args)
			err = ss.Kill(out, &[]byte{})
			So(err, ShouldBeNil)
		})
		Convey("GenerateResponse", func() {
			r := &Response{}
			ss.listenAddress = "1234"
			ss.token = "asdf"
			response := ss.generateResponse(r)
			So(response, ShouldHaveSameTypeAs, []byte{})
			json.Unmarshal(response, &r)
			So(r.ListenAddress, ShouldEqual, "1234")
			So(r.Token, ShouldEqual, "asdf")
		})
		Convey("InitSessionState", func() {
			var mockPluginArgs string = "{\"RunAsDaemon\": true, \"PingTimeoutDuration\": 2000000000}"
			m := PluginMeta{
				RPCType: NativeRPC,
				Type:    CollectorPluginType,
			}
			sessionState, err, rc := NewSessionState(mockPluginArgs, &MockPlugin{Meta: m}, &m)
			So(sessionState.ListenAddress(), ShouldEqual, "")
			So(rc, ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(sessionState, ShouldNotBeNil)
			So(sessionState.PingTimeoutDuration, ShouldResemble, 2*time.Second)
		})
		Convey("InitSessionState with invalid args", func() {
			var mockPluginArgs string
			m := PluginMeta{
				RPCType: NativeRPC,
				Type:    CollectorPluginType,
			}
			_, err, _ := NewSessionState(mockPluginArgs, &MockPlugin{Meta: m}, &m)
			So(err, ShouldNotBeNil)
		})
		Convey("InitSessionState with a custom log path", func() {
			var mockPluginArgs string = "{\"RunAsDaemon\": false, \"PluginLogPath\": \"/var/tmp/snap_plugin.log\"}"
			m := PluginMeta{
				RPCType: NativeRPC,
				Type:    CollectorPluginType,
			}
			sess, err, rc := NewSessionState(mockPluginArgs, &MockPlugin{Meta: m}, &m)
			So(rc, ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(sess, ShouldNotBeNil)
		})
		Convey("heartbeatWatch timeout expired", func() {
			PingTimeoutLimit = 1
			ss.LastPing = now.Truncate(time.Minute)
			killChan := make(chan int)
			ss.heartbeatWatch(killChan)
			rc := <-killChan
			So(rc, ShouldEqual, 0)
		})
		Convey("heatbeatWatch reset", func() {
			PingTimeoutLimit = 2
			killChan := make(chan int)
			ss.heartbeatWatch(killChan)
			rc := <-killChan
			So(rc, ShouldEqual, 0)
		})
	})
}

func TestGetConfigPolicy(t *testing.T) {
	Convey("Get Config Policy", t, func() {
		logger := log.New()
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
		var reply []byte
		c.Session.GetConfigPolicy([]byte{}, &reply)
		var cpr GetConfigPolicyReply
		err := c.Session.Decode(reply, &cpr)
		So(err, ShouldBeNil)
		So(cpr.Policy, ShouldNotBeNil)
	})
	Convey("Get error in Config Policy ", t, func() {
		logger := log.New()
		errSession := &errSessionState{
			&MockSessionState{
				Encoder:             encoding.NewGobEncoder(),
				listenPort:          "0",
				token:               "abcdef",
				logger:              logger,
				PingTimeoutDuration: time.Millisecond * 100,
				killChan:            make(chan int),
			}}
		mockErrorPlugin := &mockErrorPlugin{}
		errC := &collectorPluginProxy{
			Plugin:  mockErrorPlugin,
			Session: errSession,
		}
		var reply []byte
		err := errC.Session.GetConfigPolicy([]byte{}, &reply)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "GetConfigPolicy call error : Error in get config policy")
	})
}
