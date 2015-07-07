package plugin

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

type MockPublisher struct {
	Meta PluginMeta
}

func (f *MockPublisher) Publish(_ string, _ []byte, _ map[string]ctypes.ConfigValue) error {
	return nil
}

func (f *MockPublisher) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	return cpolicy.ConfigPolicyNode{}
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
	// These setting ensure it exists before test timeout
	PingTimeoutLimit = 1
	logger := log.New(os.Stdout,
		"test: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Convey("Collector", t, func() {
		Convey("start with unknown port", func() {
			s := &MockPublisherSessionState{
				listenPort:          "-1",
				token:               "abcdef",
				logger:              logger,
				PingTimeoutDuration: time.Millisecond * 100,
				killChan:            make(chan int),
			}
			r := new(Response)
			c := new(MockPublisher)
			So(func() { StartPublisher(c, s, r) }, ShouldPanic)
			Convey("start with dynamic port", func() {
				s = &MockPublisherSessionState{
					listenPort:          "0",
					token:               "1234",
					logger:              logger,
					PingTimeoutDuration: time.Millisecond * 100,
					killChan:            make(chan int),
				}
				r := new(Response)
				c := new(MockPublisher)
				err, rc := StartPublisher(c, s, r)
				So(err, ShouldBeNil)
				So(rc, ShouldEqual, 0)

				Convey("RPC service already registered", func() {
					err, _ := StartPublisher(c, s, r)
					So(err, ShouldResemble, errors.New("rpc: service already defined: MockPublisherSessionState"))
				})

			})
		})
	})
}
