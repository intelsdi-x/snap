package control

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/intelsdilabs/gomit"
	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

type MockController struct {
}

func (p *MockController) GenerateArgs(daemon bool) plugin.Arg {
	a := plugin.Arg{
		PluginLogPath: "/tmp/pulse-test-plugin.log",
		RunAsDaemon:   daemon,
	}
	return a
}

type MockExecutablePlugin struct {
	Timeout     bool
	NilResponse bool
	NoPing      bool
}

func (m *MockExecutablePlugin) ResponseReader() io.Reader {
	return nil
}

func (m *MockExecutablePlugin) Start() error {
	return nil
}

func (m *MockExecutablePlugin) Kill() error {
	return nil
}

func (m *MockExecutablePlugin) Wait() error {
	return nil
}

func (m *MockExecutablePlugin) WaitForResponse(t time.Duration) (*plugin.Response, error) {
	if m.Timeout {
		return nil, errors.New("timeout")
	}

	if m.NilResponse {
		return nil, nil
	}

	resp := new(plugin.Response)
	resp.Type = plugin.CollectorPluginType
	return resp, nil
}

type MockHandlerDelegate struct {
	ErrorMode       bool
	WasRegistered   bool
	WasUnregistered bool
	StopError       error
}

func (m *MockHandlerDelegate) RegisterHandler(s string, h gomit.Handler) error {
	if m.ErrorMode {
		return errors.New("fake delegate error")
	}
	m.WasRegistered = true
	return nil
}

func (m *MockHandlerDelegate) UnregisterHandler(s string) error {
	if m.StopError != nil {
		return m.StopError
	}
	m.WasUnregistered = true
	return nil
}

func (m *MockHandlerDelegate) HandlerCount() int {
	return 0
}

func (m *MockHandlerDelegate) IsHandlerRegistered(s string) bool {
	return false
}

type MockHealthyPluginCollectorClient struct{}

func (mpcc *MockHealthyPluginCollectorClient) Ping() error {
	return nil
}

type MockUnhealthyPluginCollectorClient struct{}

func (mucc *MockUnhealthyPluginCollectorClient) Ping() error {
	return errors.New("Fail")
}

func TestRunnerState(t *testing.T) {
	Convey("pulse/control", t, func() {

		Convey("Runner", func() {

			Convey(".AddDelegates", func() {

				Convey("adds a handler delegate", func() {
					r := NewRunner()

					r.AddDelegates(new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 1)
				})

				Convey("adds multiple delegates", func() {
					r := NewRunner()

					r.AddDelegates(new(MockHandlerDelegate))
					r.AddDelegates(new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

				Convey("adds multiple delegates (batch)", func() {
					r := NewRunner()

					r.AddDelegates(new(MockHandlerDelegate), new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

			})

			Convey(".Start", func() {

				Convey("returns error without adding delegates", func() {
					r := NewRunner()
					e := r.Start()

					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, "No delegates added before called Start()")
				})

				Convey("starts after adding one delegates", func() {
					r := NewRunner()
					m1 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					e := r.Start()

					So(e, ShouldBeNil)
					So(m1.WasRegistered, ShouldBeTrue)
				})

				Convey("starts after  after adding multiple delegates", func() {
					r := NewRunner()
					m1 := new(MockHandlerDelegate)
					m2 := new(MockHandlerDelegate)
					m3 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					r.AddDelegates(m2, m3)
					e := r.Start()

					So(e, ShouldBeNil)
					So(m1.WasRegistered, ShouldBeTrue)
					So(m2.WasRegistered, ShouldBeTrue)
					So(m3.WasRegistered, ShouldBeTrue)
				})

				Convey("error if delegate cannot RegisterHandler", func() {
					r := NewRunner()
					me := new(MockHandlerDelegate)
					me.ErrorMode = true
					r.AddDelegates(me)
					e := r.Start()

					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, "fake delegate error")
				})

			})

			Convey(".Stop", func() {

				Convey("removes handlers from delegates", func() {
					r := NewRunner()
					m1 := new(MockHandlerDelegate)
					m2 := new(MockHandlerDelegate)
					m3 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					r.AddDelegates(m2, m3)
					r.Start()
					e := r.Stop()

					So(m1.WasUnregistered, ShouldBeTrue)
					So(m2.WasUnregistered, ShouldBeTrue)
					So(m3.WasUnregistered, ShouldBeTrue)
					// No errors
					So(len(e), ShouldEqual, 0)
				})

				Convey("returns errors for handlers errors on stop", func() {
					r := NewRunner()
					m1 := new(MockHandlerDelegate)
					m1.StopError = errors.New("0")
					m2 := new(MockHandlerDelegate)
					m2.StopError = errors.New("1")
					m3 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					r.AddDelegates(m2, m3)
					r.Start()
					e := r.Stop()

					So(m1.WasUnregistered, ShouldBeFalse)
					So(m2.WasUnregistered, ShouldBeFalse)
					So(m3.WasUnregistered, ShouldBeTrue)
					// No errors
					So(len(e), ShouldEqual, 2)
					So(e[0].Error(), ShouldEqual, "0")
					So(e[1].Error(), ShouldEqual, "1")
				})

			})

		})
	})
}

func TestRunnerPluginRunning(t *testing.T) {
	Convey("pulse/control", t, func() {

		Convey("Runner", func() {
			Convey("startPlugin", func() {

				// These tests only work if Pulse Path is known to discover dummy plugin used for testing
				if PulsePath != "" {
					Convey("should return an AvailablePlugin in a Running state", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)

						// exPlugin := new(MockExecutablePlugin)
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldBeNil)
						So(ap, ShouldNotBeNil)
						println(ap.State)
						So(ap.State, ShouldEqual, PluginRunning)
					})

					Convey("availablePlugins should include returned availablePlugin", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						apCount := len(r.availablePlugins.Table())
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						So(ap, ShouldNotBeNil)
						So(len(r.availablePlugins.Table()), ShouldEqual, apCount+1)
						So(ap, ShouldBeIn, r.availablePlugins.Table())
					})

					Convey("healthcheck on healthy plugin does not increment failedHealthChecks", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockHealthyPluginCollectorClient)
						ap.checkHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("healthcheck on unhealthy plugin increments failedHealthChecks", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.checkHealth()
						So(ap.failedHealthChecks, ShouldEqual, 1)
					})

					Convey("successful healthcheck resets failedHealthChecks", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin-foo.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.checkHealth()
						ap.checkHealth()
						So(ap.failedHealthChecks, ShouldEqual, 2)
						ap.client = new(MockHealthyPluginCollectorClient)
						ap.checkHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("three consecutive failedHealthChecks disables the plugin", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.checkHealth()
						ap.checkHealth()
						ap.checkHealth()
						So(ap.failedHealthChecks, ShouldEqual, 3)
						So(ap.State, ShouldEqual, PluginDisabled)
					})

					Convey("should return error for WaitForResponse error", func() {
						r := NewRunner()
						exPlugin := new(MockExecutablePlugin)
						exPlugin.Timeout = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(ap, ShouldBeNil)
						So(e, ShouldResemble, errors.New("error while waiting for response: timeout"))
					})

					Convey("should return error for nil availablePlugin", func() {
						r := NewRunner()
						exPlugin := new(MockExecutablePlugin)
						exPlugin.NilResponse = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("no reponse object returned from plugin"))
						So(ap, ShouldBeNil)
					})

					Convey("should return error for executable plugin not in daemon mode", func() {
						r := NewRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, false)
						if err != nil {
							panic(err)
						}

						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("error while creating client connection: dial tcp: missing address"))
						So(ap, ShouldBeNil)
					})
				}

			})

			Convey("stopPlugin", func() {
				// TODO
			})
		})
	})
}

func TestAvailablePlugins(t *testing.T) {
	Convey("Append", t, func() {
		Convey("returns an error when loading duplicate plugins", func() {
			ap := newAvailablePlugins()
			ap.Append(new(availablePlugin))

			p, _ := ap.Get(0)
			err := ap.Append(p)
			So(err, ShouldResemble, errors.New("plugin instance already available at index 0"))

		})
	})
	Convey("Get", t, func() {
		Convey("returns an error when index is out of range", func() {
			ap := newAvailablePlugins()
			ap.Append(new(availablePlugin))

			_, err := ap.Get(1)
			So(err, ShouldResemble, errors.New("index out of range"))

		})
	})
	Convey("Splice", t, func() {
		Convey("splices an item out of the table", func() {
			ap := newAvailablePlugins()
			ap.Append(new(availablePlugin))
			ap.Append(new(availablePlugin))
			ap.Append(new(availablePlugin))
			ap.Splice(1)
			So(len(ap.Table()), ShouldResemble, 2)

		})
	})
}
