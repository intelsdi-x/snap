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
	"net"
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAvailablePlugin(t *testing.T) {
	Convey("newAvailablePlugin()", t, func() {
		Convey("returns an availablePlugin", func() {
			ln, _ := net.Listen("tcp", ":4000")
			defer ln.Close()
			resp := &plugin.Response{
				Meta: plugin.PluginMeta{
					Name:    "testPlugin",
					Version: 1,
				},
				Type:          plugin.CollectorPluginType,
				ListenAddress: "127.0.0.1:4000",
			}
			ap, err := newAvailablePlugin(resp, nil, nil)
			So(ap, ShouldHaveSameTypeAs, new(availablePlugin))
			So(err, ShouldBeNil)
		})
	})

	Convey("Stop()", t, func() {
		Convey("returns nil if plugin successfully stopped", func() {
			r := newRunner()
			a := plugin.Arg{
				PluginLogPath: "/tmp/snap-test-plugin-stop.log",
			}

			exPlugin, _ := plugin.NewExecutablePlugin(a, PluginPath)
			ap, err := r.startPlugin(exPlugin)
			So(err, ShouldBeNil)

			err = ap.Stop("testing")
			So(err, ShouldBeNil)
		})
	})
}

func TestAvailablePlugins(t *testing.T) {
	Convey("newAvailablePlugins()", t, func() {
		Convey("returns a pointer to an availablePlugins struct", func() {
			aps := newAvailablePlugins()
			So(aps, ShouldHaveSameTypeAs, new(availablePlugins))
		})
	})
	Convey("insert()", t, func() {
		Convey("adds a collector into the collectors collection", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				pluginType: plugin.CollectorPluginType,
				name:       "test",
				version:    1,
			}
			err := aps.insert(ap)
			So(err, ShouldBeNil)

			pool, err := aps.getPool("collector:test:1")
			So(err, ShouldBeNil)

			status := aps.hasPool("collector:test:1")
			So(status, ShouldBeTrue)

			nap, ok := pool.Plugins()[ap.id]
			So(ok, ShouldBeTrue)
			So(nap, ShouldEqual, ap)
		})
		Convey("returns an error if an unknown plugin type is given", func() {
			aps := newAvailablePlugins()
			ap := &availablePlugin{
				pluginType: 99,
				name:       "test",
				version:    1,
			}
			err := aps.insert(ap)

			So(err, ShouldResemble, errors.New("bad plugin type"))

			status := aps.hasPool("collector:test:1")
			So(status, ShouldBeFalse)
		})
	})
	Convey("it returns an error if client cannot be created", t, func() {
		resp := &plugin.Response{
			Meta: plugin.PluginMeta{
				Name:    "test",
				Version: 1,
			},
			Type:          plugin.CollectorPluginType,
			ListenAddress: "localhost:",
		}
		ap, err := newAvailablePlugin(resp, nil, nil)
		So(ap, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}
