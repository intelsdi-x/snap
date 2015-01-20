package plugin

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	PluginName = "pulse-collector-dummy"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", "collector", PluginName)
)

type MockController struct {
}

func (p *MockController) GenerateArgs() Arg {
	a := Arg{
		PluginLogPath: "/tmp",
	}
	return a
}

// Mock Executor used to test
type MockPluginExecutor struct {
	Killed          bool
	Response        string
	WaitTime        time.Duration
	WaitError       error
	WaitForResponse func(time.Duration) (Response, error)
}

// Mock
func (m *MockPluginExecutor) WaitForExit() error {
	t := time.Now()

	// Loop until wait time expired
	for time.Now().Sub(t) < m.WaitTime {
		// Return if Killed while waiting
		if m.Killed {
			return m.WaitError
		}
	}
	return m.WaitError
}

// Mock
func (m *MockPluginExecutor) Kill() error {
	m.Killed = true
	return nil
}

// Mock
func (m *MockPluginExecutor) ResponseReader() io.Reader {
	readbuffer := bytes.NewBuffer([]byte(m.Response))
	reader := bufio.NewReader(readbuffer)
	return reader
}

func TestNewExecutablePlugin(t *testing.T) {
	Convey("pluginControl.WaitForResponse", t, func() {
		c := new(MockController)

		ex, err := NewExecutablePlugin(c.GenerateArgs(), "/foo/bar", false)

		Convey("returns ExecutablePlugin", func() {
			So(ex, ShouldNotBeNil)
		})

		Convey("does not return error", func() {
			So(err, ShouldBeNil)
		})

	})

}

func TestWaitForPluginResponse(t *testing.T) {
	Convey(".waitHandling", t, func() {

		Convey("called with PluginExecutor that returns a valid response", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.Response = "{}"
			mockExecutor.WaitTime = time.Millisecond * 1
			Convey("daemon mode off", func() {
				resp, err := waitHandling(mockExecutor, time.Second*3, false)

				So(mockExecutor.Killed, ShouldEqual, false)
				So(resp, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("daemon mode on", func() {
				resp, err := waitHandling(mockExecutor, time.Second*3, true)

				So(mockExecutor.Killed, ShouldEqual, false)
				So(resp, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey("called with PluginExecutor that returns an invalid response", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.Response = "junk"
			mockExecutor.WaitTime = time.Millisecond * 1000

			Convey("daemon mode off", func() {
				resp, err := waitHandling(mockExecutor, time.Millisecond*100, false)
				So(mockExecutor.Killed, ShouldEqual, true)
				So(resp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "JSONError")
			})
			Convey("daemon mode on", func() {
				resp, err := waitHandling(mockExecutor, time.Millisecond*100, true)
				So(mockExecutor.Killed, ShouldEqual, true)
				So(resp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "JSONError")
			})
		})

		Convey("called with PluginExecutor that exits immediately without returning a reponse", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.WaitTime = time.Millisecond * 100
			mockExecutor.WaitError = errors.New("Exit 127")
			resp, err := waitHandling(mockExecutor, time.Millisecond*100, false)

			So(mockExecutor.Killed, ShouldEqual, false)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Exit 127")
		})

		Convey("called with PluginExecutor that will run longer than timeout without responding", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.WaitTime = time.Second * 120
			resp, err := waitHandling(mockExecutor, time.Millisecond*100, false)

			So(mockExecutor.Killed, ShouldEqual, true)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "timeout waiting for response")
		})

		// These tests don't mock and directly use dummy collector plugin
		// They require pulse path being set and a recent build of the plugin
		// WIP
		if PluginPath != "" {
			Convey("dummy", func() {
				m := new(MockController)
				a := m.GenerateArgs()
				a.PluginLogPath = ""
				ex, err := NewExecutablePlugin(a, PluginPath, true)
				if err != nil {
					panic(err)
				}

				ex.Start()
				r, e := ex.WaitForResponse(time.Second * 5)
				if r != nil {
					println("ListenAddress: " + r.ListenAddress)
				}
				if e != nil {
					println(e.Error())
				}

			})

			Convey("dummy2", func() {
				m := new(MockController)
				a := m.GenerateArgs()
				a.PluginLogPath = ""
				ex, err := NewExecutablePlugin(a, PluginPath, false)
				if err != nil {
					panic(err)
				}

				ex.Start()
				r, e := ex.WaitForResponse(time.Second * 5)
				if r != nil {
					println("ListenAddress: " + r.ListenAddress)
				}
				if e != nil {
					println(e.Error())
				}

			})
		}

	})
}
