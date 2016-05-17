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

package client

// Functional tests through client to REST API

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/mgmt/rest"
	"github.com/intelsdi-x/snap/scheduler"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	// Change to set the REST API logging to debug
	LOG_LEVEL = log.FatalLevel

	SNAP_PATH         = os.Getenv("SNAP_PATH")
	MOCK_PLUGIN_PATH1 = []string{SNAP_PATH + "/plugin/snap-collector-mock1"}
	MOCK_PLUGIN_PATH2 = []string{SNAP_PATH + "/plugin/snap-collector-mock2"}
	FILE_PLUGIN_PATH  = []string{SNAP_PATH + "/plugin/snap-publisher-file"}
	DIRECTORY_PATH    = []string{SNAP_PATH + "/plugin/"}

	NextPort = 45000

	p1 *LoadPluginResult
	p2 *LoadPluginResult
	p3 *LoadPluginResult
)

func getWMFromSample(sample string) *wmap.WorkflowMap {
	jsonP, err := ioutil.ReadFile("../wmap_sample/" + sample)
	if err != nil {
		log.Fatal(err)
	}
	wf, err := wmap.FromJson(jsonP)
	if err != nil {
		log.Fatal(err)
	}
	return wf
}

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI() string {
	// Start a REST API to talk to
	rest.StreamingBufferWindow = 0.01
	log.SetLevel(LOG_LEVEL)
	r, _ := rest.New(rest.GetDefaultConfig())
	c := control.New(control.GetDefaultConfig())
	c.Start()
	s := scheduler.New(scheduler.GetDefaultConfig())
	s.SetMetricManager(c)
	s.Start()
	r.BindConfigManager(c.Config)
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	go func(ch <-chan error) {
		// Block on the error channel. Will return exit status 1 for an error or just return if the channel closes.
		err, ok := <-ch
		if !ok {
			return
		}
		log.Fatal(err)
	}(r.Err())
	r.SetAddress("127.0.0.1:0")
	r.Start()
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://localhost:%d", r.Port())
}

func TestSnapClient(t *testing.T) {
	CompressUpload = false

	uri := startAPI()
	c, cerr := New(uri, "v1", true)
	wf := getWMFromSample("1.json")
	sch := &Schedule{Type: "simple", Interval: "1s"}
	uuid := uuid.New()

	Convey("Client should exist", t, func() {
		So(cerr, ShouldBeNil)
		Convey("Testing API after startup", func() {
			Convey("empty version", func() {
				c, err := New(uri, "", true)
				So(err, ShouldBeNil)
				So(c.Version, ShouldEqual, "v1")
			})
			Convey("no loaded plugins", func() {
				p := c.GetPlugins(false)
				p2 := c.GetPlugins(true)

				So(p.Err, ShouldBeNil)
				So(p2.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 0)
				So(p.AvailablePlugins, ShouldBeEmpty)
				So(len(p2.LoadedPlugins), ShouldEqual, 0)
				So(p2.AvailablePlugins, ShouldBeEmpty)

				_, err := c.pluginUploadRequest([]string{""})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "stat : no such file or directory")
			})
			Convey("empty catalog", func() {
				m := c.GetMetricCatalog()
				So(m.Err, ShouldNotBeNil)
				So(m.Len(), ShouldEqual, 0)
				So(m.Err.Error(), ShouldEqual, "Metric catalog is empty (no plugin loaded)")
			})
			Convey("load directory error", func() {
				p := c.LoadPlugin(DIRECTORY_PATH)
				So(p.Err, ShouldNotBeNil)
				So(p.LoadedPlugins, ShouldBeEmpty)
				So(p.Err.Error(), ShouldEqual, "Provided plugin path is a directory not file")
			})
			Convey("unknown task", func() {
				Convey("GetTask/GetTasks", func() {
					t1 := c.GetTask(uuid)
					t2 := c.GetTasks()
					So(t1.Err, ShouldNotBeNil)
					So(t2.Err, ShouldBeNil)
				})
				Convey("StopTask", func() {
					t1 := c.StopTask(uuid)
					So(t1.Err, ShouldNotBeNil)
					So(t1.Err.Error(), ShouldEqual, fmt.Sprintf("error 0: Task not found: ID(%s) ", uuid))
				})
				Convey("RemoveTask", func() {
					t1 := c.RemoveTask(uuid)
					So(t1.Err, ShouldNotBeNil)
					So(t1.Err.Error(), ShouldEqual, fmt.Sprintf("Task not found: ID(%s)", uuid))
				})
				Convey("invalid task (missing metric)", func() {
					tt := c.CreateTask(sch, wf, "baron", "", true)
					So(tt.Err, ShouldNotBeNil)
					So(tt.Err.Error(), ShouldContainSubstring, "Metric not found: /intel/mock/foo")
				})
			})
		})
	})

	CompressUpload = true
	if cerr == nil {
		p1 = c.LoadPlugin(MOCK_PLUGIN_PATH1)
	}
	CompressUpload = false
	Convey("Client should exist", t, func() {
		So(cerr, ShouldBeNil)
		Convey("single plugin loaded", func() {
			Convey("an error should not be received loading a plugin", func() {
				So(c.Version, ShouldEqual, "v1")

				So(p1, ShouldNotBeNil)
				So(p1.Err, ShouldBeNil)
				So(p1.LoadedPlugins, ShouldNotBeEmpty)
				So(p1.LoadedPlugins[0].Name, ShouldEqual, "mock")
				So(p1.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(p1.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
			Convey("there should be one loaded plugin", func() {
				p := c.GetPlugins(false)
				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 1)
				So(p.AvailablePlugins, ShouldBeEmpty)
			})
			Convey("invalid task (missing publisher)", func() {
				tf := c.CreateTask(sch, wf, "baron", "", false)
				So(tf.Err, ShouldNotBeNil)
				So(tf.Err.Error(), ShouldContainSubstring, "Plugin not found: type(publisher) name(file)")
			})
			Convey("plugin already loaded", func() {
				p1 := c.LoadPlugin(MOCK_PLUGIN_PATH1)
				So(p1.Err, ShouldNotBeNil)
				So(p1.Err.Error(), ShouldEqual, "plugin is already loaded")
			})
		})
	})

	if cerr == nil {
		p2 = c.LoadPlugin(MOCK_PLUGIN_PATH2)
	}
	Convey("Client should exist", t, func() {
		So(cerr, ShouldBeNil)
		Convey("loading second plugin", func() {
			Convey("an error should not be received loading second plugin", func() {
				So(p2, ShouldNotBeNil)
				So(p2.Err, ShouldBeNil)
				So(p2.LoadedPlugins, ShouldNotBeEmpty)
				So(p2.LoadedPlugins[0].Name, ShouldEqual, "mock")
				So(p2.LoadedPlugins[0].Version, ShouldEqual, 2)
				So(p2.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
			Convey("there should be two loaded plugins", func() {
				p := c.GetPlugins(false)
				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 2)
				So(p.AvailablePlugins, ShouldBeEmpty)
			})
		})
		Convey("Metrics", func() {
			Convey("MetricCatalog", func() {
				m := c.GetMetricCatalog()
				So(m.Err, ShouldBeNil)
				So(m.Len(), ShouldEqual, 6)
				So(m.Catalog[0].Namespace, ShouldEqual, "/intel/mock/*/baz")
				So(m.Catalog[0].Version, ShouldEqual, 1)
				So(m.Catalog[1].Namespace, ShouldEqual, "/intel/mock/*/baz")
				So(m.Catalog[1].Version, ShouldEqual, 2)
				So(m.Catalog[2].Namespace, ShouldEqual, "/intel/mock/bar")
				So(m.Catalog[2].Version, ShouldEqual, 1)
				So(m.Catalog[3].Namespace, ShouldEqual, "/intel/mock/bar")
				So(m.Catalog[3].Version, ShouldEqual, 2)
				So(m.Catalog[4].Namespace, ShouldEqual, "/intel/mock/foo")
				So(m.Catalog[4].Version, ShouldEqual, 1)
				So(m.Catalog[5].Namespace, ShouldEqual, "/intel/mock/foo")
				So(m.Catalog[5].Version, ShouldEqual, 2)
			})
			Convey("FetchMetrics", func() {
				Convey("leaf metric all versions", func() {
					m := c.FetchMetrics("/intel/mock/bar/*", 0)
					So(m.Catalog[0].Namespace, ShouldEqual, "/intel/mock/bar")
					So(m.Catalog[0].Version, ShouldEqual, 1)
					So(m.Catalog[1].Namespace, ShouldEqual, "/intel/mock/bar")
					So(m.Catalog[1].Version, ShouldEqual, 2)
				})
				Convey("version 2 leaf metric", func() {
					m := c.FetchMetrics("/intel/mock/bar/*", 2)
					So(m.Catalog[0].Namespace, ShouldEqual, "/intel/mock/bar")
					So(m.Catalog[0].Version, ShouldEqual, 2)
				})
				Convey("version 2 non-leaf metrics", func() {
					m := c.FetchMetrics("/intel/mock/*", 2)
					So(m.Catalog[0].Namespace, ShouldEqual, "/intel/mock/*/baz")
					So(m.Catalog[0].Version, ShouldEqual, 2)
					So(m.Catalog[1].Namespace, ShouldEqual, "/intel/mock/bar")
					So(m.Catalog[1].Version, ShouldEqual, 2)
					So(m.Catalog[2].Namespace, ShouldEqual, "/intel/mock/foo")
					So(m.Catalog[2].Version, ShouldEqual, 2)
				})
			})
		})
	})

	if cerr == nil {
		p3 = c.LoadPlugin(FILE_PLUGIN_PATH)
	}
	Convey("Client should exist", t, func() {
		So(cerr, ShouldBeNil)
		Convey("publisher plugin loaded", func() {
			Convey("an error should not be received loading publisher plugin", func() {
				So(p3, ShouldNotBeNil)
				So(p3.Err, ShouldBeNil)
				So(p3.LoadedPlugins, ShouldNotBeEmpty)
				So(p3.LoadedPlugins[0].Name, ShouldEqual, "file")
				So(p3.LoadedPlugins[0].Version, ShouldEqual, 3)
				So(p3.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
			Convey("there should be three loaded plugins", func() {
				p := c.GetPlugins(false)
				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 3)
				So(p.AvailablePlugins, ShouldBeEmpty)
			})
		})

		Convey("Tasks", func() {
			Convey("Passing a bad task manifest", func() {
				wfb := getWMFromSample("bad.json")
				ttb := c.CreateTask(sch, wfb, "bad", "", true)
				So(ttb.Err, ShouldNotBeNil)
			})

			tf := c.CreateTask(sch, wf, "baron", "", false)
			Convey("valid task not started on creation", func() {
				So(tf.Err, ShouldBeNil)
				So(tf.Name, ShouldEqual, "baron")
				So(tf.State, ShouldEqual, "Stopped")

				// method not allowed
				rsp, err := c.do("POST", fmt.Sprintf("/tasks/%v", tf.ID), ContentTypeJSON) //case len(body) == 0
				So(rsp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				b := make([]byte, 5)
				rsp2, err2 := c.do("POST", fmt.Sprintf("/tasks/%v", tf.ID), ContentTypeJSON, b) //case len(body) != 0
				So(rsp2, ShouldBeNil)
				So(err2, ShouldNotBeNil)

				Convey("GetTasks", func() {
					t1 := c.GetTasks()
					So(t1.Err, ShouldBeNil)
					t2 := c.GetTask(tf.ID)
					So(t2.Err, ShouldBeNil)
				})
				Convey("StopTask", func() {
					t1 := c.StopTask(tf.ID)
					So(t1.Err, ShouldNotBeNil)
					So(t1.Err.Error(), ShouldEqual, "error 0: Task is already stopped. ")
				})
				Convey("StartTask", func() {
					t1 := c.StartTask(tf.ID)
					So(t1.Err, ShouldBeNil)
					So(t1.ID, ShouldEqual, tf.ID)
				})
				Convey("RemoveTask", func() {
					t1 := c.RemoveTask(tf.ID)
					So(t1.Err, ShouldBeNil)
					So(t1.ID, ShouldEqual, tf.ID)

					b := make([]byte, 5)
					rsp, err := c.do("DELETE", fmt.Sprintf("/tasks/%v", tf.ID), ContentTypeJSON, b) //case len(body) != 0
					So(rsp, ShouldNotBeNil)
					So(err, ShouldBeNil)
				})
			})

			tt := c.CreateTask(sch, wf, "baron", "", true)
			Convey("valid task started on creation", func() {
				So(tt.Err, ShouldBeNil)
				So(tt.Name, ShouldEqual, "baron")
				So(tt.State, ShouldEqual, "Running")

				// method not allowed
				rsp, err := c.do("POST", fmt.Sprintf("/tasks/%v", tt.ID), ContentTypeJSON) //case len(body) == 0
				So(rsp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				b := make([]byte, 5)
				rsp2, err2 := c.do("POST", fmt.Sprintf("/tasks/%v", tt.ID), ContentTypeJSON, b) //case len(body) != 0
				So(rsp2, ShouldBeNil)
				So(err2, ShouldNotBeNil)

				Convey("GetTasks", func() {
					t1 := c.GetTasks()
					So(t1.Err, ShouldBeNil)
					t2 := c.GetTask(tt.ID)
					So(t2.Err, ShouldBeNil)
				})
				Convey("StartTask", func() {
					t1 := c.StartTask(tt.ID)
					So(t1.Err, ShouldNotBeNil)
					So(t1.Err.Error(), ShouldEqual, "error 0: Task is already running. ")
					t2 := c.StartTask(tt.ID)
					So(t2.Err, ShouldNotBeNil)
					So(t2.Err.Error(), ShouldEqual, "error 0: Task is already running. ")
				})
				Convey("RemoveTask", func() {
					t1 := c.RemoveTask(tt.ID)
					So(t1.Err, ShouldNotBeNil)
					So(t1.Err.Error(), ShouldEqual, "Task must be stopped")
				})
				Convey("StopTask", func() {
					t1 := c.StopTask(tt.ID)
					So(t1.Err, ShouldBeNil)
					So(t1.ID, ShouldEqual, tt.ID)
					//try stopping again to make sure channel is closed
					t2 := c.StopTask(tt.ID)
					So(t2.Err, ShouldNotBeNil)
					So(t2.Err.Error(), ShouldEqual, "error 0: Task is already stopped. ")

					b := make([]byte, 5)
					rsp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", tt.ID), ContentTypeJSON, b)
					So(rsp, ShouldNotBeNil)
					So(err, ShouldBeNil)
				})
				Convey("enable a stopped task", func() {
					et := c.EnableTask(tt.ID)
					So(et.Err, ShouldNotBeNil)
					So(et.Err.Error(), ShouldEqual, "Task must be disabled")
				})
				Convey("WatchTasks", func() {
					Convey("invalid task ID", func() {
						rest.StreamingBufferWindow = 0.01

						type ea struct {
							events []string
							sync.Mutex
						}

						a := new(ea)
						r := c.WatchTask("1")

						wait := make(chan struct{})
						go func() {
							for {
								select {
								case e := <-r.EventChan:
									a.Lock()
									a.events = append(a.events, e.EventType)
									if len(a.events) == 5 {
										r.Close()
									}
									a.Unlock()
								case <-r.DoneChan:
									close(wait)
									return
								}
							}
						}()
						<-wait
						So(r.Err.Error(), ShouldEqual, "Task not found: ID(1)")
					})
					Convey("event stream", func() {
						rest.StreamingBufferWindow = 0.01
						sch := &Schedule{Type: "simple", Interval: "100ms"}
						tf := c.CreateTask(sch, wf, "baron", "", false)

						type ea struct {
							events []string
							sync.Mutex
						}

						a := new(ea)
						r := c.WatchTask(tf.ID)
						wait := make(chan struct{})
						go func() {
							for {
								select {
								case e := <-r.EventChan:
									a.Lock()
									a.events = append(a.events, e.EventType)
									if len(a.events) == 5 {
										r.Close()
									}
									a.Unlock()
								case <-r.DoneChan:
									close(wait)
									return
								}
							}
						}()
						startResp := c.StartTask(tf.ID)
						So(startResp.Err, ShouldBeNil)
						<-wait
						a.Lock()
						So(len(a.events), ShouldEqual, 5)
						a.Unlock()
						So(a.events[0], ShouldEqual, "task-started")
						for x := 2; x <= 4; x++ {
							So(a.events[x], ShouldEqual, "metric-event")
						}
					})
				})
			})
		})
		Convey("UnloadPlugin", func() {
			Convey("unload unknown plugin", func() {
				p := c.UnloadPlugin("not a type", "foo", 3)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "plugin not found")
			})
			Convey("unload one of multiple", func() {
				p1 := c.GetPlugins(false)
				So(p1.Err, ShouldBeNil)
				So(len(p1.LoadedPlugins), ShouldEqual, 3)

				p2 := c.UnloadPlugin("collector", "mock", 2)
				So(p2.Err, ShouldBeNil)
				So(p2.Name, ShouldEqual, "mock")
				So(p2.Version, ShouldEqual, 2)
				So(p2.Type, ShouldEqual, "collector")

				p3 := c.UnloadPlugin("publisher", "file", 3)
				So(p3.Err, ShouldBeNil)
				So(p3.Name, ShouldEqual, "file")
				So(p3.Version, ShouldEqual, 3)
				So(p3.Type, ShouldEqual, "publisher")
			})
			Convey("unload when only one plugin loaded", func() {
				p1 := c.GetPlugins(false)
				So(p1.Err, ShouldBeNil)
				So(len(p1.LoadedPlugins), ShouldEqual, 1)
				So(p1.LoadedPlugins[0].Name, ShouldEqual, "mock")

				p2 := c.UnloadPlugin("collector", "mock", 1)
				So(p2.Err, ShouldBeNil)
				So(p2.Name, ShouldEqual, "mock")
				So(p2.Version, ShouldEqual, 1)
				So(p2.Type, ShouldEqual, "collector")

				p3 := c.GetPlugins(false)
				So(p3.Err, ShouldBeNil)
				So(len(p3.LoadedPlugins), ShouldEqual, 0)
			})
		})
	})

	c, err := New("http://localhost:-1", "v1", true)
	Convey("API with invalid port", t, func() {
		So(err, ShouldNotBeNil)
		So(c, ShouldBeNil)
	})

	c, err = New("test", "", true)
	Convey("API with invalid url - no scheme", t, func() {
		So(err, ShouldNotBeNil)
		So(c, ShouldBeNil)
	})

	c, err = New("ftp://127.0.0.1:1", "", true)
	Convey("API with invalid url - ftp", t, func() {
		So(err, ShouldNotBeNil)
		So(c, ShouldBeNil)
	})

	c, err = New("htp://127.0.0.1:1", "", true)
	Convey("API with invalid url - typo", t, func() {
		So(err, ShouldNotBeNil)
		So(c, ShouldBeNil)
	})
}
