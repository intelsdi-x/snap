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
