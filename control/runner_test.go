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

package control

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/control/routing"
	. "github.com/smartystreets/goconvey/convey"
)

type MockController struct {
}

func (p *MockController) GenerateArgs(daemon bool) plugin.Arg {
	a := plugin.Arg{
		PluginLogPath: "/tmp/pulse-test-plugin.log",
		NoDaemon:      daemon,
	}
	return a
}

type MockExecutablePlugin struct {
	Timeout       bool
	NilResponse   bool
	NoPing        bool
	StartError    bool
	PluginFailure bool
}

func (m *MockExecutablePlugin) ResponseReader() io.Reader {
	return nil
}

func (m *MockExecutablePlugin) Start() error {
	if m.StartError {
		return errors.New("start error")
	}
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
	if m.PluginFailure {
		resp.State = plugin.PluginFailure
		resp.ErrorMessage = "plugin start error"
	}
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

func (mucc *MockHealthyPluginCollectorClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return nil, errors.New("Fail")
}

func (mucc *MockHealthyPluginCollectorClient) SetKey() error {
	return nil
}

type MockUnhealthyPluginCollectorClient struct{}

func (mucc *MockUnhealthyPluginCollectorClient) Ping() error {
	return errors.New("Fail")
}

func (mucc *MockUnhealthyPluginCollectorClient) Kill(string) error {
	return errors.New("Fail")
}

func (mucc *MockUnhealthyPluginCollectorClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return nil, errors.New("Fail")
}

func (mucc *MockUnhealthyPluginCollectorClient) SetKey() error {
	return nil
}

type MockEmitter struct{}

func (me *MockEmitter) Emit(gomit.EventBody) (int, error) { return 0, nil }

func TestRunnerState(t *testing.T) {
	// Enabled log output in test
	// log.SetFormatter(&log.TextFormatter{ForceColors: true, DisableTimestamp: false})
	// log.SetLevel(log.DebugLevel)
	// log.SetOutput(os.Stdout)

	Convey("pulse/control", t, func() {

		Convey("Runner", func() {

			Convey(".AddDelegates", func() {

				Convey("adds a handler delegate", func() {
					r := newRunner(&routing.RoundRobinStrategy{})

					r.AddDelegates(new(MockHandlerDelegate))
					r.SetEmitter(new(MockEmitter))
					So(len(r.delegates), ShouldEqual, 1)
				})

				Convey("adds multiple delegates", func() {
					r := newRunner(&routing.RoundRobinStrategy{})

					r.AddDelegates(new(MockHandlerDelegate))
					r.AddDelegates(new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

				Convey("adds multiple delegates (batch)", func() {
					r := newRunner(&routing.RoundRobinStrategy{})

					r.AddDelegates(new(MockHandlerDelegate), new(MockHandlerDelegate))
					So(len(r.delegates), ShouldEqual, 2)
				})

			})

			Convey(".Start", func() {

				Convey("returns error without adding delegates", func() {
					r := newRunner(&routing.RoundRobinStrategy{})
					e := r.Start()

					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, "No delegates added before called Start()")
				})

				Convey("starts after adding one delegates", func() {
					r := newRunner(&routing.RoundRobinStrategy{})
					m1 := new(MockHandlerDelegate)
					r.AddDelegates(m1)
					e := r.Start()

					So(e, ShouldBeNil)
					So(m1.WasRegistered, ShouldBeTrue)
				})

				Convey("starts after  after adding multiple delegates", func() {
					r := newRunner(&routing.RoundRobinStrategy{})
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
					r := newRunner(&routing.RoundRobinStrategy{})
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
					r := newRunner(&routing.RoundRobinStrategy{})
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
					r := newRunner(&routing.RoundRobinStrategy{})
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
	// log.SetLevel(log.DebugLevel)
	Convey("pulse/control", t, func() {
		Convey("Runner", func() {
			Convey("startPlugin", func() {

				// These tests only work if Pulse Path is known to discover mock plugin used for testing
				if PulsePath != "" {
					Convey("should return an AvailablePlugin", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
							// Daemon:        true,
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)

						// exPlugin := new(MockExecutablePlugin)
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldBeNil)
						So(ap, ShouldNotBeNil)

						err = r.stopPlugin("testing", ap)

						So(err, ShouldBeNil)
					})

					Convey("availablePlugins should include returned availablePlugin", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						colCount := len(r.availablePlugins.all())
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						So(ap, ShouldNotBeNil)
						So(len(r.availablePlugins.all()), ShouldEqual, colCount+1)
						So(ap, ShouldBeIn, r.availablePlugins.all())
					})

					Convey("healthcheck on healthy plugin does not increment failedHealthChecks", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockHealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("healthcheck on unhealthy plugin increments failedHealthChecks", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 1)
					})

					Convey("successful healthcheck resets failedHealthChecks", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin-foo.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 2)
						ap.client = new(MockHealthyPluginCollectorClient)
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 0)
					})

					Convey("three consecutive failedHealthChecks disables the plugin", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						a := plugin.Arg{
							PluginLogPath: "/tmp/pulse-test-plugin.log",
						}
						exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
						if err != nil {
							panic(err)
						}

						So(err, ShouldBeNil)
						ap, e := r.startPlugin(exPlugin)
						So(e, ShouldBeNil)
						ap.client = new(MockUnhealthyPluginCollectorClient)
						ap.CheckHealth()
						ap.CheckHealth()
						ap.CheckHealth()
						So(ap.failedHealthChecks, ShouldEqual, 3)
					})

					Convey("should return error for WaitForResponse error", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						r.SetEmitter(new(MockEmitter))
						exPlugin := new(MockExecutablePlugin)
						exPlugin.Timeout = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(ap, ShouldBeNil)
						So(e, ShouldResemble, errors.New("error while waiting for response: timeout"))
					})

					Convey("should return error for nil availablePlugin", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						exPlugin := new(MockExecutablePlugin)
						exPlugin.NilResponse = true // set to not response
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("no reponse object returned from plugin"))
						So(ap, ShouldBeNil)
					})

					Convey("should return error if plugin fails while starting", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						exPlugin := &MockExecutablePlugin{
							StartError: true,
						}
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("error while starting plugin: start error"))
						So(ap, ShouldBeNil)
					})

					Convey("should return error if plugin fails to start", func() {
						r := newRunner(&routing.RoundRobinStrategy{})
						exPlugin := &MockExecutablePlugin{
							PluginFailure: true,
						}
						ap, e := r.startPlugin(exPlugin)

						So(e, ShouldResemble, errors.New("plugin could not start error: plugin start error"))
						So(ap, ShouldBeNil)
					})
				}

			})

			Convey("stopPlugin", func() {
				Convey("should return an AvailablePlugin in a Running state", func() {
					r := newRunner(&routing.RoundRobinStrategy{})
					a := plugin.Arg{
						PluginLogPath: "/tmp/pulse-test-plugin-stop.log",
					}
					exPlugin, err := plugin.NewExecutablePlugin(a, PluginPath)
					if err != nil {
						panic(err)
					}

					So(err, ShouldBeNil)

					// exPlugin := new(MockExecutablePlugin)
					ap, e := r.startPlugin(exPlugin)

					So(e, ShouldBeNil)
					So(ap, ShouldNotBeNil)

					e = r.stopPlugin("testing", ap)
					So(e, ShouldBeNil)
				})
			})
		})
	})
}
