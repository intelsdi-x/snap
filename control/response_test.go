package control

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockPluginExecutor struct {
	Killed   bool
	Response string

	WaitTime  time.Duration
	WaitError error
}

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

func (m *MockPluginExecutor) Kill() error {
	m.Killed = true
	return nil
}

func (m *MockPluginExecutor) StdoutPipe() io.Reader {
	readbuffer := bytes.NewBuffer([]byte(m.Response))
	reader := bufio.NewReader(readbuffer)
	return reader
}

// to test
// if plugin does not respond within timeout return error (x)
// if plugin responds with invalid data return error
// if plugin stops before returning data return error (x)
// if plugin responds in time return response

func TestWaitForPluginResponse(t *testing.T) {

	Convey("Given a PluginExector returns a response with invalid data", t, func() {
		mockExecutor := new(MockPluginExecutor)
		mockExecutor.Response = "junk"
		mockExecutor.WaitTime = time.Millisecond * 1000

		Convey("when control.WaitForPluginResponse is passed the PluginExecutor", func() {
			resp, err := WaitForPluginResponse(mockExecutor, time.Millisecond*100)

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

	})

	Convey("Given a PluginExector that exits immediately", t, func() {
		mockExecutor := new(MockPluginExecutor)
		mockExecutor.WaitTime = time.Millisecond * 100
		mockExecutor.WaitError = errors.New("Exit 127")

		Convey("when control.WaitForPluginResponse is passed the PluginExecutor", func() {
			resp, err := WaitForPluginResponse(mockExecutor, time.Second*10)

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

	Convey("Given a PluginExector that will run longer than timeout without responding", t, func() {
		mockExecutor := new(MockPluginExecutor)
		mockExecutor.WaitTime = time.Second * 120

		Convey("when control.WaitForPluginResponse is passed the PluginExecutor", func() {
			resp, err := WaitForPluginResponse(mockExecutor, time.Millisecond*10)

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
