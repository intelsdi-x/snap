// +build legacy

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
	"bufio"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/fileutils"
	. "github.com/smartystreets/goconvey/convey"
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

			_, err := lp.get("not" + core.Separator + "found" + core.Separator + "1")
			So(err, ShouldResemble, errors.New("plugin not found"))

		})
	})
}

func loadPlugin(p *pluginManager, fileName string, retries ...int) (*loadedPlugin, serror.SnapError) {
	// This is a Travis optimized loading of plugins. From time to time, tests will error in Travis
	// due to a timeout when waiting for a response from a plugin. We are going to attempt loading a plugin
	// 3 times before letting the error through. Hopefully this cuts down on the number of Travis failures
	var retryCount int
	if len(retries) == 0 {
		retryCount = 3
	} else {
		retryCount = retries[0]
	}
	var e serror.SnapError
	var lp *loadedPlugin
	file, err := os.Open(fileName)
	if err != nil {
		return nil, serror.New(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, serror.New(err)
	}
	size := info.Size()
	bytes := make([]byte, size)
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)
	if err != nil {
		return nil, serror.New(err)
	}

	path, err := fileutils.WriteFile(filepath.Base(fileName), GetDefaultConfig().TempDirPath, bytes)
	if err != nil {
		return nil, serror.New(err)
	}

	details := &pluginDetails{
		Path:     path,
		ExecPath: filepath.Dir(path),
		Exec:     []string{filepath.Base(path)},
	}

	for i := 0; i < retryCount; i++ {
		lp, e = p.LoadPlugin(details, nil)
		if e == nil {
			break //success
		}
	}
	return lp, e
}

func loadStandalonePlugin(p *pluginManager, uriName string) (*loadedPlugin, serror.SnapError) {
	// This is a Travis optimized loading of plugins. From time to time, tests will error in Travis
	// due to a timeout when waiting for a response from a plugin. We are going to attempt loading a plugin
	// 3 times before letting the error through. Hopefully this cuts down on the number of Travis failures
	var e serror.SnapError
	var lp *loadedPlugin

	uri, err := url.ParseRequestURI(uriName)
	if err != nil {
		return nil, serror.New(err)
	}
	details := &pluginDetails{
		Uri: uri,
	}

	lp, e = p.LoadPlugin(details, nil)
	if e != nil {
		return nil, e
	}
	return lp, nil
}

// Uses the mock collector plugin to simulate loading
func TestLoadPlugin(t *testing.T) {
	// These tests only work if SNAP_PATH is known
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir
	if fixtures.SnapPath != "" {
		Convey("PluginManager.LoadPlugin", t, func() {
			Convey("loads plugin successfully", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := loadPlugin(p, fixtures.PluginPathMock2)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.all(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.all()), ShouldBeGreaterThan, 0)
			})

			Convey("load plugin unsuccessful due to timeout", func() {
				cfg := GetDefaultConfig()
				cfg.Plugins.Collector.Plugins["mock"] = newPluginConfigItem(optAddPluginConfigItem("test-sleep-duration", ctypes.ConfigValueStr{Value: "2s"}))
				pm := newPluginManager(OptSetPluginConfig(cfg.Plugins))
				pm.SetPluginLoadTimeout(1)
				pm.SetMetricCatalog(newMetricCatalog())
				lplugin, lerr := loadPlugin(pm, fixtures.PluginPathMock2, 1)
				So(lplugin, ShouldBeNil)
				So(lerr, ShouldNotBeNil)
				So(lerr.Error(), ShouldContainSubstring, "timed out")
			})

			Convey("with a plugin config a plugin loads successfully", func() {
				cfg := GetDefaultConfig()
				cfg.Plugins.Collector.Plugins["mock"] = newPluginConfigItem(optAddPluginConfigItem("test", ctypes.ConfigValueBool{Value: true}))
				tags := newPluginTags()
				tags["/intel/mock"] = make(map[string]string)
				tags["/intel/mock"]["context"] = "plugin_manager_test"
				p := newPluginManager(OptSetPluginConfig(cfg.Plugins), OptSetPluginTags(tags))
				p.SetMetricCatalog(newMetricCatalog())
				lp, serr := loadPlugin(p, fixtures.PluginPathMock2)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.all(), ShouldNotBeEmpty)
				So(serr, ShouldBeNil)
				So(len(p.all()), ShouldBeGreaterThan, 0)
				mts, err := p.metricCatalog.Fetch(core.NewNamespace())
				So(err, ShouldBeNil)
				So(len(mts), ShouldBeGreaterThan, 2)
				So(mts[0].Description(), ShouldResemble, "mock description")
				So(mts[0].Unit(), ShouldResemble, "mock unit")
				So(mts[0].Tags(), ShouldContainKey, "plugin_running_on")
				So(mts[0].Tags(), ShouldContainKey, "context")
				So(mts[0].Tags()["plugin_running_on"], ShouldNotResemble, "")
				So(mts[0].Tags()["context"], ShouldResemble, "plugin_manager_test")
			})

			Convey("for a plugin requiring a config an incomplete config will result in a load failure", func() {
				cfg := GetDefaultConfig()
				cfg.Plugins.Collector.Plugins["mock"] = newPluginConfigItem(optAddPluginConfigItem("test-fail", ctypes.ConfigValueBool{Value: true}))
				p := newPluginManager(OptSetPluginConfig(cfg.Plugins))
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := loadPlugin(p, fixtures.PluginPathMock2)
				So(lp, ShouldBeNil)
				So(p.all(), ShouldBeEmpty)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "testing")
				So(len(p.all()), ShouldEqual, 0)
			})

			Convey("loads native-rpc plugin successfully", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := loadPlugin(p, fixtures.PluginPathMock1)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.loadedPlugins, ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.loadedPlugins.table), ShouldBeGreaterThan, 0)
			})

			Convey("loads plugin with cache TTL set", func() {
				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := loadPlugin(p, fixtures.PluginPathMock1)

				So(err, ShouldBeNil)
				So(lp.Meta.CacheTTL, ShouldNotBeNil)
				So(lp.Meta.CacheTTL, ShouldResemble, time.Duration(time.Millisecond*1100))
			})
			Convey("loads standalone plugin successfully", func() {
				cmd := exec.Command(fixtures.PluginPathMock2Grpc, "--stand-alone", "--stand-alone-port", "8183")
				err := cmd.Start()
				So(err, ShouldBeNil)
				time.Sleep(100 * time.Millisecond)

				p := newPluginManager()
				p.SetMetricCatalog(newMetricCatalog())
				lp, err := loadStandalonePlugin(p, fixtures.PluginUriMock2Grpc)

				So(lp, ShouldHaveSameTypeAs, new(loadedPlugin))
				So(p.all(), ShouldNotBeEmpty)
				So(err, ShouldBeNil)
				So(len(p.all()), ShouldBeGreaterThan, 0)
				So(lp.Details.Uri.String(), ShouldContainSubstring, fixtures.PluginUriMock2Grpc)
				So(len(lp.Details.Uri.String()), ShouldEqual, len(fixtures.PluginUriMock2Grpc))

				cmd.Process.Kill()
			})
		})

	}
}

func TestUnloadPlugin(t *testing.T) {
	if fixtures.SnapPath != "" {
		Convey("pluginManager.UnloadPlugin", t, func() {

			Convey("when a loaded plugin is unloaded", func() {
				Convey("then it is removed from the loadedPlugins", func() {
					p := newPluginManager()
					p.SetMetricCatalog(newMetricCatalog())
					_, err := loadPlugin(p, fixtures.PluginPathMock2)
					So(err, ShouldBeNil)

					numPluginsLoaded := len(p.all())
					So(numPluginsLoaded, ShouldEqual, 1)
					lp, _ := p.get("collector" + core.Separator + "mock" + core.Separator + "2")
					_, err = p.UnloadPlugin(lp)

					So(err, ShouldBeNil)
					So(len(p.all()), ShouldEqual, numPluginsLoaded-1)
				})
			})

			Convey("when a loaded plugin is not in a PluginLoaded state", func() {
				Convey("then an error is thrown", func() {
					p := newPluginManager()
					p.SetMetricCatalog(newMetricCatalog())
					lp, err := loadPlugin(p, fixtures.PluginPathMock2)
					glp, err2 := p.get("collector" + core.Separator + "mock" + core.Separator + "2")
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
					_, err := loadPlugin(p, fixtures.PluginPathMock2)

					lp, err2 := p.get("collector" + core.Separator + "mock" + core.Separator + "2")
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
