package plugin

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockController struct {
}

func (p *MockController) GenerateArgs(daemon bool) Arg {
	a := Arg{
		PluginLogPath: "/tmp",
		RunAsDaemon:   daemon,
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
func (m *MockPluginExecutor) Wait() error {
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
		ex, err := NewExecutablePlugin(c, "/foo/bar", false)

		Convey("returns ExecutablePlugin", func() {
			So(ex, ShouldNotBeNil)
		})

		Convey("does not return error", func() {
			So(err, ShouldBeNil)
		})

	})

}

func TestWaitForPluginResponse(t *testing.T) {
	Convey(".WaitForResponse", t, func() {
		Convey("called with PluginExecutor that returns a valid response", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.Response = "{}"
			mockExecutor.WaitTime = time.Millisecond * 1
			resp, err := WaitForResponse(mockExecutor, time.Millisecond*100)

			Convey("The PluginExecutor.Kill() should not be called", func() {
				So(mockExecutor.Killed, ShouldEqual, false)
			})

			Convey("Returns a response", func() {
				So(resp, ShouldNotBeNil)
			})

			Convey("Returns nil instead of error", func() {
				So(err, ShouldBeNil)
			})

		})

		Convey("called with PluginExecutor that returns an invalid response", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.Response = "junk"
			mockExecutor.WaitTime = time.Millisecond * 1000
			resp, err := WaitForResponse(mockExecutor, time.Millisecond*100)

			Convey("The PluginExecutor.Kill() should be called", func() {
				So(mockExecutor.Killed, ShouldEqual, true)
			})

			Convey("Returns nil response", func() {
				So(resp, ShouldBeNil)
			})

			Convey("Returns error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "JSONError")
			})

		})

		Convey("called with PluginExecutor that exits immediately without returning a reponse", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.WaitTime = time.Millisecond * 100
			mockExecutor.WaitError = errors.New("Exit 127")
			Convey("when control.WaitForPluginResponse is passed the PluginExecutor", func() {
				resp, err := WaitForResponse(mockExecutor, time.Second*10)

				Convey("The PluginExecutor.Kill() should not be called", func() {
					So(mockExecutor.Killed, ShouldEqual, false)
				})

				Convey("Returns nil response", func() {
					So(resp, ShouldBeNil)
				})

				Convey("Returns error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Exit 127")
				})

			})
		})

		Convey("called with PluginExecutor that will run longer than timeout without responding", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.WaitTime = time.Second * 120
			resp, err := WaitForResponse(mockExecutor, time.Millisecond*10)

			Convey("The PluginExecutor.Kill() should be called", func() {
				So(mockExecutor.Killed, ShouldEqual, true)
			})

			Convey("Returns nil response", func() {
				So(resp, ShouldBeNil)
			})

			Convey("Returns error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Timeout waiting for response")
			})
		})

	})
}
