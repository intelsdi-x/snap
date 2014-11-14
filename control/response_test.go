package control

import (
	"errors"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockPluginExecutor struct {
	Killed bool

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

// to test
// if plugin does not respond within timeout return error
// if plugin responds with invalid data return error
// if plugin stops before returning data return error
// if plugin responds in time return response

func TestWaitForPluginResponse(t *testing.T) {

	Convey("Given a PluginExector that exits immediately", t, func() {
		mockExecutor := new(MockPluginExecutor)
		mockExecutor.WaitTime = time.Millisecond * 100
		mockExecutor.WaitError = errors.New("Exit 127")

		Convey("when control.WaitForPluginResponse is passed the PluginExecutor", func() {
			resp, err := WaitForPluginResponse(mockExecutor, time.Second*10)

			Convey("The PluginExecutor.Kill() should be called", func() {
				So(mockExecutor.Killed, ShouldEqual, false)
			})

			Convey("Returns nil response", func() {
				So(resp, ShouldBeNil)
			})

			Convey("Returns error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Returns error indicating timeout occured", func() {
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
			})

			Convey("Returns error indicating timeout occured", func() {
				So(err.Error(), ShouldEqual, "Timeout waiting for response")
			})
		})
	})
}
