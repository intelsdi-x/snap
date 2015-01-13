package control

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Mock Executor used to test
type MockPluginExecutor struct {
	Killed          bool
	Response        string
	WaitTime        time.Duration
	WaitError       error
	WaitForResponse func(time.Duration) (*plugin.Response, error)
}

var (
	PluginName = "pulse-collector-dummy"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", "collector", "pulse-collector-dummy")
)

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

// Uses the dummy collector plugin to simulate Loading
func TestLoad(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		Convey("pluginControl.Load", t, func() {

			Convey("loads successfully", func() {
				c := Control()
				c.Start()
				loadedPlugin, err := c.Load(PluginPath)

				So(loadedPlugin, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("returns error if not started", func() {
				c := Control()
				loadedPlugin, err := c.Load(PluginPath)

				So(loadedPlugin, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("adds to pluginControl.LoadedPlugins on successful load",
				func() {
					c := Control()
					c.Start()
					loadedPlugin, err := c.Load(PluginPath)

					So(loadedPlugin, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(len(c.LoadedPlugins), ShouldBeGreaterThan, 0)
				})

		})

	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestStop(t *testing.T) {
	Convey("pluginControl.Stop", t, func() {
		c := Control()
		c.Start()
		c.Stop()

		Convey("returns ExecutablePlugin", func() {
			So(c.Started, ShouldBeFalse)
		})

	})

}

func TestNewExecutablePlugin(t *testing.T) {
	Convey("pluginControl.WaitForResponse", t, func() {
		ex, err := newExecutablePlugin(Control(), "/foo/bar", false)

		Convey("returns ExecutablePlugin", func() {
			So(ex, ShouldNotBeNil)
		})

		Convey("does not return error", func() {
			So(err, ShouldBeNil)
		})

	})

}

func TestWaitForPluginResponse(t *testing.T) {
	Convey(".waitForResponse", t, func() {
		Convey("called with PluginExecutor that returns a valid response", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.Response = "{}"
			mockExecutor.WaitTime = time.Millisecond * 1
			resp, err := waitForResponse(mockExecutor, time.Millisecond*100)

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
			resp, err := waitForResponse(mockExecutor, time.Millisecond*100)

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
				resp, err := waitForResponse(mockExecutor, time.Second*10)

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
			resp, err := waitForResponse(mockExecutor, time.Millisecond*10)

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
