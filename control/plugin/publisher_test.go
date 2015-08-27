package plugin

import (
	"log"
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

func (f *MockPublisher) GetConfigPolicy() cpolicy.ConfigPolicy {
	return cpolicy.ConfigPolicy{}
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
			c := new(MockProcessor)
			m := &PluginMeta{
				RPCType: JSONRPC,
				Type:    PublisherPluginType,
			}
			// we will panic since rpc.HandleHttp has already
			// been called during TestStartCollector
			Convey("RPC service already registered", func() {
				So(func() { Start(m, c, "{}") }, ShouldPanic)
			})

		})
	})
}
