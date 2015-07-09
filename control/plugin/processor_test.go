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

type MockProcessor struct {
	Meta PluginMeta
}

func (f *MockProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	return "", nil, nil
}

func (f *MockProcessor) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	return cpolicy.ConfigPolicyNode{}
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
	// These setting ensure it exists before test timeout
	PingTimeoutLimit = 1
	logger := log.New(os.Stdout,
		"test: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Convey("Processor", t, func() {
		Convey("start with unknown port", func() {
			s := &MockProcessorSessionState{
				listenPort:          "-1",
				token:               "abcdef",
				logger:              logger,
				PingTimeoutDuration: time.Millisecond * 100,
				killChan:            make(chan int),
			}
			r := new(Response)
			c := new(MockProcessor)
			So(func() { StartProcessor(c, s, r) }, ShouldPanic)
			Convey("start with dynamic port", func() {
				s = &MockProcessorSessionState{
					listenPort:          "0",
					token:               "1234",
					logger:              logger,
					PingTimeoutDuration: time.Millisecond * 100,
					killChan:            make(chan int),
				}
				r := new(Response)
				c := new(MockProcessor)
				err, rc := StartProcessor(c, s, r)
				So(err, ShouldBeNil)
				So(rc, ShouldEqual, 0)

				Convey("RPC service already registered", func() {
					err, _ := StartProcessor(c, s, r)
					So(err, ShouldResemble, errors.New("rpc: service already defined: MockProcessorSessionState"))
				})

			})
		})
	})
}
