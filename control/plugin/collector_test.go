package plugin

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type MockSessionState struct {
	runAsDaemon   bool
	listenAddress string
	listenPort    string
	token         string
	logger        *log.Logger
	killChan      chan int
}

// func (s *MockSessionState) NewMockSesssionState(pluginArgsMsg string) (*MockSessionState, error, int) {
// 	return &SessionState{
// 		runAsDaemon: true,
// 	}, nil, 0
// }

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

func (s *MockSessionState) RunAsDaemon() bool {
	return s.runAsDaemon
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

func (s *MockSessionState) KillChan() chan int {
	return s.killChan
}

func (s *MockSessionState) generateResponse(r *Response) []byte {
	return []byte("mockResponse")
}

func (s *MockSessionState) heartbeatWatch(killChan chan int) {
	time.Sleep(time.Millisecond * 200)
	killChan <- 0
}

type MockPlugin struct {
	Meta   PluginMeta
	Policy ConfigPolicy
}

func (f *MockPlugin) Collect(args CollectorArgs, reply *CollectorReply) error {
	return nil
}

func (c *MockPlugin) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
	reply.MetricTypes = []*MetricType{
		NewMetricType([]string{"org", "some_metric"}, time.Now().Unix()),
	}
	return nil
}

func TestStartCollector(t *testing.T) {
	// These setting ensure it exists before test timeout
	PingTimeoutDuration = time.Millisecond * 100
	PingTimeoutLimit = 1
	logger := log.New(os.Stdout,
		"test: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Convey("Collector", t, func() {
		Convey("start with unknown port", func() {
			s := &MockSessionState{
				listenPort: "-1",
				token:      "abcdef",
				logger:     logger,
				killChan:   make(chan int),
			}
			r := new(Response)
			c := new(MockPlugin)
			So(func() { StartCollector(c, s, r) }, ShouldPanic)
			Convey("start with dynamic port", func() {
				s = &MockSessionState{
					listenPort: "0",
					token:      "abcdef",
					logger:     logger,
					killChan:   make(chan int),
				}
				r := new(Response)
				c := new(MockPlugin)
				err, rc := StartCollector(c, s, r)
				So(err, ShouldBeNil)
				So(rc, ShouldEqual, 0)

				Convey("RPC service already registered", func() {
					err, _ := StartCollector(c, s, r)
					So(err, ShouldResemble, errors.New("rpc: service already defined: MockSessionState"))
				})

			})
		})
	})
}
