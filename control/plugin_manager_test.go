/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"os"
	"path"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/pulse/control/plugin"
)

var (
	PluginName = "pulse-collector-dummy1"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", PluginName)

	JSONRPC_PluginName = "pulse-collector-dummy2"
	JSONRPC_PluginPath = path.Join(PulsePath, "plugin", JSONRPC_PluginName)
)

func TestLoadedPlugins(t *testing.T) {
	Convey("Append", t, func() {
		Convey("returns an error when loading duplicate plugins", func() {
			lp := newLoadedPlugins()
			lp.add(&loadedPlugin{
				Meta: plugin.PluginMeta{
					Name: "test1",
				},
			})
			err := lp.add(&loadedPlugin{
				Meta: plugin.PluginMeta{
					Name: "test1",
				},
			})
			So(err.Error(), ShouldResemble, "plugin is already loaded")

		})
	})
	Convey("get", t, func() {
		Convey("returns an error when index is out of range", func() {
			lp := newLoadedPlugins()

			_, err := lp.get("not:found:1")
			So(err, ShouldResemble, errors.New("plugin not found"))

		})
	})
}

// Uses the dummy collector plugin to simulate loading
func TestLoadPlugin(t *testing.T) {
	// These tests only work if PULSE_PATH is known
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir

	if PulsePath != "" {
		Convey("PluginManager.LoadPlugin", t, func() {

			Convey("loads plugin successfully", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := p.LoadPlugin(PluginPath, nil)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.all(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.all()), ShouldBeGreaterThan, 0)
			})

			Convey("loads json-rpc plugin successfully", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := p.LoadPlugin(JSONRPC_PluginPath, nil)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.loadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.loadedPlugins.table), ShouldBeGreaterThan, 0)
			})

			Convey("loads plugin with cache TTL set", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := p.LoadPlugin(PluginPath, nil)

				So(err, ShouldBeNil)
				So(lp.Meta.CacheTTL, ShouldNotBeNil)
				So(lp.Meta.CacheTTL, ShouldResemble, time.Duration(time.Millisecond*100))
			})

			// Convey("error is returned on a bad PluginPath", func() {
			// 	p := newPluginManager()
			// 	lp, err := p.LoadPlugin("", nil)

			// 	So(lp, ShouldBeNil)
			// 	So(err, ShouldNotBeNil)
			// })

		})

	}
}

func TestUnloadPlugin(t *testing.T) {
	if PulsePath != "" {
		Convey("pluginManager.UnloadPlugin", t, func() {

			Convey("when a loaded plugin is unloaded", func() {
				Convey("then it is removed from the loadedPlugins", func() {
					p := newPluginManager()
					p.SetMetricCatalog(newMetricCatalog())
					_, err := p.LoadPlugin(PluginPath, nil)
					So(err, ShouldBeNil)

					numPluginsLoaded := len(p.all())
					So(numPluginsLoaded, ShouldEqual, 1)
					lp, _ := p.get("collector:dummy1:1")
					_, err = p.UnloadPlugin(lp)

					So(err, ShouldBeNil)
					So(len(p.all()), ShouldEqual, numPluginsLoaded-1)
				})
			})

			Convey("when a loaded plugin is not in a PluginLoaded state", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.SetMetricCatalog(newMetricCatalog())
					lp, err := p.LoadPlugin(PluginPath, nil)
					glp, err2 := p.get("collector:dummy1:1")
					So(err2, ShouldBeNil)
					glp.State = DetectedState
					_, err = p.UnloadPlugin(lp)
					So(err.Error(), ShouldResemble, "Plugin must be in a LoadedState")
				})
			})

			Convey("when a plugin is already unloaded", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.SetMetricCatalog(newMetricCatalog())
					_, err := p.LoadPlugin(PluginPath, nil)

					lp, err2 := p.get("collector:dummy1:1")
					So(err2, ShouldBeNil)
					_, err = p.UnloadPlugin(lp)

					_, err = p.UnloadPlugin(lp)
					So(err.Error(), ShouldResemble, "plugin not found")

				})
			})
		})
	}
}

func TestLoadedPlugin(t *testing.T) {
	lp := new(loadedPlugin)
	lp.Meta = plugin.PluginMeta{Name: "test", Version: 1}
	Convey(".Name()", t, func() {
		Convey("it returns the name from the plugin metadata", func() {
			So(lp.Name(), ShouldEqual, "test")
		})
	})
	Convey(".Version()", t, func() {
		Convey("it returns the version from the plugin metadata", func() {
			So(lp.Version(), ShouldEqual, 1)
		})
	})
	Convey(".TypeName()", t, func() {
		lp.Type = 0
		Convey("it returns the string representation of the plugin type", func() {
			So(lp.TypeName(), ShouldEqual, "collector")
		})
	})
	Convey(".Status()", t, func() {
		lp.State = LoadedState
		Convey("it returns a string of the current plugin state", func() {
			So(lp.Status(), ShouldEqual, "loaded")
		})
	})
	Convey(".LoadedTimestamp()", t, func() {
		ts := time.Now()
		lp.LoadedTime = ts
		Convey("it returns the timestamp of the LoadedTime", func() {
			So(lp.LoadedTimestamp(), ShouldResemble, &ts)
		})
	})
}
