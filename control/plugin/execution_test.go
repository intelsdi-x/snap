// +build small

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
	"io"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type mockCmd struct{}

func (mc *mockCmd) Path() string { return "" }
func (mc *mockCmd) Kill() error  { return nil }
func (mc *mockCmd) Start() error { return nil }

func setupMockExec(resp []byte, timeout bool) *ExecutablePlugin {
	stdout, stdoutw := io.Pipe()
	stderr, stderrw := io.Pipe()
	go func() {
		if timeout {
			time.Sleep(time.Second)
		}
		stdoutw.Write(resp)
		stdoutw.Write([]byte("\n"))
		stdoutw.Write([]byte("some log message on stdout\n"))
		stderrw.Write([]byte("some log message on stderr\n"))
		stdoutw.Close()
		stderrw.Close()
	}()
	return &ExecutablePlugin{
		cmd:    &mockCmd{},
		stdout: stdout,
		stderr: stderr,
	}
}

func TestExecutablePlugin(t *testing.T) {
	Convey("NewExecutablePlugin returns a pointer to the correct type", t, func() {
		e, err := NewExecutablePlugin(Arg{}, "")
		So(err, ShouldBeNil)
		So(e, ShouldHaveSameTypeAs, &ExecutablePlugin{})
	})
	Convey("Run()", t, func() {
		Convey("returns a valid response when a valid response is given", func() {
			e := setupMockExec([]byte(`{"Token": "a token"}`), false)
			resp, err := e.Run(time.Millisecond * 100)
			So(err, ShouldBeNil)
			So(resp.Token, ShouldEqual, "a token")
		})
		Convey("returns an error if an invalid response is given", func() {
			e := setupMockExec([]byte(`this is bad`), false)
			_, err := e.Run(time.Millisecond * 100)
			So(err, ShouldNotBeNil)
		})
		Convey("returns an error if the timeout expires", func() {
			e := setupMockExec([]byte(`{"Token": "a token"}`), true)
			_, err := e.Run(time.Millisecond * 100)
			So(err, ShouldNotBeNil)
		})
	})
}
