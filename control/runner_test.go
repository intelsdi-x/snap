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

func (mpcc *MockHealthyPluginCollectorClient) Kill(string) error {
	return nil
}

type MockUnhealthyPluginCollectorClient struct{}

func (mucc *MockUnhealthyPluginCollectorClient) Ping() error {
	return errors.New("Fail")
}

func (mucc *MockUnhealthyPluginCollectorClient) Kill(string) error {
	return errors.New("Fail")
}

func TestRunnerState(t *testing.T) {
	Convey("pulse/control", t, func() {

		Convey("Runner", func() {

			Convey(".AddDelegates", func() {

				Convey("adds a handler delegate", func() {
					r := newRunner()

					r.AddDelegates(new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 1)
				})

				Convey("adds multiple delegates", func() {
					r := newRunner()

					r.AddDelegates(new(MockHandlerDelegate))
					r.AddDelegates(new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

				Convey("adds multiple delegates (batch)", func() {
					r := newRunner()

					r.AddDelegates(new(MockHandlerDelegate), new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

			})

			Convey(".Start", func() {

				Convey("returns error without adding delegates", func() {
					r := newRunner()
					e := r.Start()

					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, "No delegates added before called Start()")
				})

				Convey("starts after adding one delegates", func() {
					r := newRunner()
					m1 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					e := r.Start()

					So(e, ShouldBeNil)
					So(m1.WasRegistered, ShouldBeTrue)
				})

				Convey("starts after  after adding multiple delegates", func() {
					r := newRunner()
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
					r := newRunner()
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
					r := newRunner()
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
					r := newRunner()
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
					Convey("should return an AvailablePlugin", func() {
						r := newRunner()
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
					})

					Convey("availablePlugins should include returned availablePlugin", func() {
						r := newRunner()
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath, true)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						t := r.availablePlugins.Collectors.Table()
						colCount := len(t["/dummy/dumb:1"])
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						So(ap, ShouldNotBeNil)
						t = r.availablePlugins.Collectors.Table()
						So(len(t["/dummy/dumb:1"]), ShouldEqual, colCount+1)
						So(ap, ShouldBeIn, t["/dummy/dumb:1"])
					})

					Convey("healthcheck on healthy plugin does not increment failedHealthChecks", func() {
						r := newRunner()
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
						ap.Client = new(MockHealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("healthcheck on unhealthy plugin increments failedHealthChecks", func() {
						r := newRunner()
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
						ap.Client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 1)
					})

					Convey("successful healthcheck resets failedHealthChecks", func() {
						r := newRunner()
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
						ap.Client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 2)
						ap.Client = new(MockHealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("three consecutive failedHealthChecks disables the plugin", func() {
						r := newRunner()
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
						ap.Client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						ap.CheckHealth()
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 3)
					})

					Convey("should return error for WaitForResponse error", func() {
						r := newRunner()
						exPlugin := new(MockExecutablePlugin)
						exPlugin.Timeout = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(ap, ShouldBeNil)
						So(e, ShouldResemble, errors.New("error while waiting for response: timeout"))
					})

					Convey("should return error for nil availablePlugin", func() {
						r := newRunner()
						exPlugin := new(MockExecutablePlugin)
						exPlugin.NilResponse = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("no reponse object returned from plugin"))
						So(ap, ShouldBeNil)
					})

					Convey("should return error for executable plugin not in daemon mode", func() {
						r := newRunner()
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
				Convey("should return an AvailablePlugin in a Running state", func() {
					r := newRunner()
					a := plugin.Arg{
						PluginLogPath: "/tmp/pulse-test-plugin-stop.log",
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

					e = ap.Stop("testing")
					So(e, ShouldBeNil)
				})
			})
		})
	})
}

func TestAvailablePlugins(t *testing.T) {
	Convey("Insert", t, func() {
		Convey("inserts a collector into the collectors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.CollectorPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			tabe := aps.Collectors.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("inserts a publisher into the publishers collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.PublisherPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			tabe := aps.Publishers.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("inserts a processor into the processors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			aps.Insert(ap)

			tabe := aps.Processors.Table()
			So(ap, ShouldBeIn, tabe["test:1"])
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    99,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldResemble, errors.New("cannot insert into available plugins, unknown plugin type"))
		})
	})
	Convey("Remove", t, func() {
		Convey("it removes a collector from the collector collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.CollectorPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldEqual, nil)
			aps.Remove(ap)
			tabe := aps.Collectors.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("it removes a publisher from the publisher collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.PublisherPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldEqual, nil)
			aps.Remove(ap)
			tabe := aps.Publishers.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("it removes a processor from the processor collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    plugin.ProcessorPluginType,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Insert(ap)
			So(err, ShouldEqual, nil)
			aps.Remove(ap)
			tabe := aps.Processors.Table()
			So(ap, ShouldNotBeIn, tabe["test:1"])
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				Type:    99,
				Name:    "test",
				Version: 1,
				Index:   0,
			}
			ap.makeKey()
			err := aps.Remove(ap)
			So(err, ShouldResemble, errors.New("cannot remove from available plugins, unknown plugin type"))
		})
	})
}

func TestAPCollection(t *testing.T) {
	Convey("Add", t, func() {
		Convey("it returns an error if the pointer already exists in the table", func() {
			apc := newAPCollection()
			ap := &availablePlugin{}
			apc.Add(ap)
			err := apc.Add(ap)
			So(err, ShouldResemble, errors.New("plugin instance already available at index 0"))
		})
	})
}
