package control

import (
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockPluginExecutor struct {
	Killed bool

	WaitMethod func() error
}

func (m *MockPluginExecutor) Wait() error {
	return m.WaitMethod()
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

	Convey("Given a PluginExector that will run longer than timeout without responding", t, func() {
		mockExecutor := new(MockPluginExecutor)
		// Set Wait to return after 500 milliseconds
		mockExecutor.WaitMethod = func() error {
			time.Sleep(time.Millisecond * 100)
			return nil
		}

		Convey("when control.WaitForPluginResponse is passed a PluginExecutor that will not respond", func() {
			resp, err := WaitForPluginResponse(mockExecutor, time.Millisecond*10)

			Convey("The PluginExecutor.Kill() should be called ", func() {
				So(mockExecutor.Killed, ShouldEqual, true)
			})

			Convey("Returns nil response", func() {
				So(resp, ShouldBeNil)
			})

			Convey("Returns error indicating timeout", func() {
				So(err.Error(), ShouldEqual, "Timeout waiting for response")
			})
		})
	})
}
