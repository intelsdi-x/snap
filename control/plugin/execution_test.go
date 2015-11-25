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
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"time"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "snap-collector-mock2"
	SnapPath   = os.Getenv("SNAP_PATH")
	PluginPath = path.Join(SnapPath, "plugin", PluginName)
)

type MockController struct {
}

func (p *MockController) GenerateArgs() Arg {
	a := Arg{
		PluginLogPath: "/tmp/plugin.log",
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

func (m *MockPluginExecutor) ErrorResponseReader() io.Reader {
	readbuffer := bytes.NewBuffer([]byte(m.Response))
	reader := bufio.NewReader(readbuffer)
	return reader
}

func TestNewExecutablePlugin(t *testing.T) {
	Convey("pluginControl.WaitForResponse", t, func() {
		c := new(MockController)

		ex, err := NewExecutablePlugin(c.GenerateArgs(), "/foo/bar")

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
				resp, err := waitHandling(mockExecutor, time.Second*3, "/tmp/some.log")

				So(mockExecutor.Killed, ShouldEqual, false)
				So(resp, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("daemon mode on", func() {
				resp, err := waitHandling(mockExecutor, time.Second*3, "/tmp/some.log")

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
				resp, err := waitHandling(mockExecutor, time.Millisecond*100, "/tmp/some.log")
				So(mockExecutor.Killed, ShouldEqual, true)
				So(resp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "JSONError")
			})
			Convey("daemon mode on", func() {
				resp, err := waitHandling(mockExecutor, time.Millisecond*100, "/tmp/some.log")
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
			resp, err := waitHandling(mockExecutor, time.Millisecond*500, "/tmp/some.log")

			So(mockExecutor.Killed, ShouldEqual, false)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Exit 127")
		})

		Convey("called with PluginExecutor that will run longer than timeout without responding", func() {
			mockExecutor := new(MockPluginExecutor)
			mockExecutor.WaitTime = time.Second * 120
			resp, err := waitHandling(mockExecutor, time.Millisecond*100, "/tmp/some.log")

			So(mockExecutor.Killed, ShouldEqual, true)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "timeout waiting for response")
		})

		// These tests don't mock and directly use mock collector plugin
		// They require snap path being set and a recent build of the plugin
		// WIP
		if PluginPath != "" {
			Convey("mock", func() {
				m := new(MockController)
				a := m.GenerateArgs()
				a.PluginLogPath = "/tmp/snap-mock.log"
				ex, err := NewExecutablePlugin(a, PluginPath)
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

			Convey("mock2", func() {
				m := new(MockController)
				a := m.GenerateArgs()
				a.PluginLogPath = "/tmp/snap-mock.log"
				ex, err := NewExecutablePlugin(a, PluginPath)
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
