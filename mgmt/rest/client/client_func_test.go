package client

// Functional tests through client to REST API

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest"
	"github.com/intelsdi-x/pulse/scheduler"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	// Change to set the REST API logging to debug
	LOG_LEVEL = log.FatalLevel

	PULSE_PATH          = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH1  = []string{PULSE_PATH + "/plugin/pulse-collector-dummy1"}
	DUMMY_PLUGIN_PATH2  = []string{PULSE_PATH + "/plugin/pulse-collector-dummy2"}
	RIEMANN_PLUGIN_PATH = []string{PULSE_PATH + "/plugin/pulse-publisher-riemann"}
	DIRECTORY_PATH      = []string{PULSE_PATH + "/plugin/"}

	NextPort = 9000
)

func getPort() int {
	defer incrPort()
	return NextPort
}

func incrPort() {
	NextPort += 10
}

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
func startAPI(port int) string {
	// Start a REST API to talk to
	log.SetLevel(LOG_LEVEL)
	r := rest.New()
	c := control.New()
	c.Start()
	s := scheduler.New()
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.Start(":" + fmt.Sprint(port))
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://localhost:%d", port)
}

func TestPulseClient(t *testing.T) {
	CompressUpload = false
	Convey("REST API functional V1", t, func() {
		Convey("GetPlugins", func() {
			Convey("empty version", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "")
				So(c.Version, ShouldEqual, "v1")
			})
			Convey("empty list", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")
				p := c.GetPlugins(false)
				p2 := c.GetPlugins(true)

				So(p.Err, ShouldBeNil)
				So(p2.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 0)
				So(p.AvailablePlugins, ShouldBeEmpty)
				So(len(p2.LoadedPlugins), ShouldEqual, 0)
				So(p2.AvailablePlugins, ShouldBeEmpty)

				_, err := c.pluginUploadRequest([]string{""})
				So(err.Error(), ShouldEqual, "stat : no such file or directory")
			})
			Convey("single item", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				p := c.GetPlugins(false)

				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 1)
				So(p.AvailablePlugins, ShouldBeEmpty)
			})
			Convey("multiple items", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p := c.GetPlugins(false)

				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 2)
				So(p.AvailablePlugins, ShouldBeEmpty)
			})
			Convey("empty list, err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.GetPlugins(false)
				p2 := c.GetPlugins(true)

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)

				So(p.Err, ShouldNotBeNil)
				So(p2.Err, ShouldNotBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 0)
				So(p.AvailablePlugins, ShouldBeEmpty)
				So(len(p2.LoadedPlugins), ShouldEqual, 0)
				So(p2.AvailablePlugins, ShouldBeEmpty)
			})
		})
		Convey("LoadPlugin", func() {
			Convey("single load", func() {
				CompressUpload = true
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				So(p.Err, ShouldBeNil)
				So(p.LoadedPlugins, ShouldNotBeEmpty)
				So(p.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(p.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(p.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
				CompressUpload = false
			})
			Convey("multiple load", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p1 := c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				So(p1.Err, ShouldBeNil)
				So(p1.LoadedPlugins, ShouldNotBeEmpty)
				So(p1.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(p1.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(p1.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())

				p2 := c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				So(p2.Err, ShouldBeNil)
				So(p2.LoadedPlugins, ShouldNotBeEmpty)
				So(p2.LoadedPlugins[0].Name, ShouldEqual, "dummy2")
				So(p2.LoadedPlugins[0].Version, ShouldEqual, 2)
				So(p2.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("already loaded", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p1 := c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				So(p1.Err, ShouldBeNil)
				So(p1.LoadedPlugins, ShouldNotBeEmpty)
				So(p1.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(p1.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(p1.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())

				p2 := c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				So(p2.Err, ShouldNotBeNil)
				So(p2.Err.Error(), ShouldEqual, "plugin is already loaded")
			})

			Convey("directory error", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p1 := c.LoadPlugin(DIRECTORY_PATH)
				So(p1.Err, ShouldNotBeNil)
				So(p1.LoadedPlugins, ShouldBeEmpty)
				So(p1.Err.Error(), ShouldEqual, "Provided plugin path is a directory not file")
			})
		})
		Convey("UnloadPlugin", func() {
			Convey("unload unknown plugin", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.UnloadPlugin("not a type", "foo", 3)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "plugin not found")
			})

			Convey("unload only one there is", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				p := c.UnloadPlugin("collector", "dummy1", 1)
				So(p.Err, ShouldBeNil)
				So(p.Name, ShouldEqual, "dummy1")
				So(p.Version, ShouldEqual, 1)
				So(p.Type, ShouldEqual, "collector")
			})

			Convey("unload one of multiple", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p1 := c.UnloadPlugin("collector", "dummy2", 2)
				So(p1.Err, ShouldBeNil)
				So(p1.Name, ShouldEqual, "dummy2")
				So(p1.Version, ShouldEqual, 2)
				So(p1.Type, ShouldEqual, "collector")

				p2 := c.GetPlugins(false)
				So(p2.Err, ShouldBeNil)
				So(len(p2.LoadedPlugins), ShouldEqual, 1)
				So(p2.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
			})
		})
		Convey("GetMetricCatalog", func() {
			Convey("empty catalog", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.GetMetricCatalog()
				So(p.Err, ShouldBeNil)
				So(p.Len(), ShouldEqual, 0)
			})
			Convey("items in catalog", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p := c.GetMetricCatalog()
				So(p.Err, ShouldBeNil)
				So(p.Len(), ShouldEqual, 4)
				So(p.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[0].Version, ShouldEqual, 1)
				So(p.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[1].Version, ShouldEqual, 2)
				So(p.Catalog[2].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(p.Catalog[2].Version, ShouldEqual, 1)
				So(p.Catalog[3].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(p.Catalog[3].Version, ShouldEqual, 2)
			})
		})
		Convey("FetchMetrics", func() {
			Convey("leaf metric all versions", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p := c.FetchMetrics("/intel/dummy/bar/*", 0)
				So(p.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[0].Version, ShouldEqual, 1)
				So(p.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[1].Version, ShouldEqual, 2)
			})
			Convey("version 2 leaf metric", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p := c.FetchMetrics("/intel/dummy/bar/*", 2)

				So(p.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[0].Version, ShouldEqual, 2)

			})
			Convey("version 2 non-leaf metrics", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				p := c.FetchMetrics("/intel/dummy/*", 2)

				So(p.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(p.Catalog[0].Version, ShouldEqual, 2)
				So(p.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(p.Catalog[1].Version, ShouldEqual, 2)

			})
		})
		Convey("CreateTask", func() {
			Convey("invalid task (missing metric)", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron", true)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldContainSubstring, "Metric not found: /intel/dummy/foo")
			})
			Convey("invalid task (missing publisher)", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron", false)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldContainSubstring, "Plugin not found: type(publisher) name(riemann) version(1)")
			})
			Convey("valid task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron", false)
				So(p.Err, ShouldBeNil)
				So(p.Name, ShouldEqual, "baron")
				So(p.State, ShouldEqual, "Stopped")

				// method not allowed
				rsp, err := c.do("POST", fmt.Sprintf("/tasks/%v", p.ID), ContentTypeJSON) //case len(body) == 0
				So(rsp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				b := make([]byte, 5)
				rsp2, err2 := c.do("POST", fmt.Sprintf("/tasks/%v", p.ID), ContentTypeJSON, b) //case len(body) != 0
				So(rsp2, ShouldBeNil)
				So(err2, ShouldNotBeNil)
			})
			Convey("valid task started on creation", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron", true)
				So(p.Err, ShouldBeNil)
				So(p.Name, ShouldEqual, "baron")
				So(p.State, ShouldEqual, "Running")

				// method not allowed
				rsp, err := c.do("POST", fmt.Sprintf("/tasks/%v", p.ID), ContentTypeJSON) //case len(body) == 0
				So(rsp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				b := make([]byte, 5)
				rsp2, err2 := c.do("POST", fmt.Sprintf("/tasks/%v", p.ID), ContentTypeJSON, b) //case len(body) != 0
				So(rsp2, ShouldBeNil)
				So(err2, ShouldNotBeNil)
			})
			Convey("do returns err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron", false)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "Post http://localhost:-1/v1/tasks: dial tcp: unknown port tcp/-1")
			})
		})
		Convey("StartTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.StartTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "error 0: No task found with id '9999999' ")
			})
			Convey("existing task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron", false)

				p2 := c.StartTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)
			})
			Convey("do returns err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.StartTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "Put http://localhost:-1/v1/tasks/9999999/start: dial tcp: unknown port tcp/-1")
			})
		})
		Convey("StopTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.StopTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "error 0: No task found with id '9999999' ")
			})
			Convey("existing task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron", false)
				p2 := c.StopTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)

				b := make([]byte, 5)
				rsp, err := c.do("PUT", fmt.Sprintf("/tasks/%v/stop", p1.ID), ContentTypeJSON, b)
				So(rsp, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("do returns err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.StopTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "Put http://localhost:-1/v1/tasks/9999999/stop: dial tcp: unknown port tcp/-1")
			})
		})
		Convey("RemoveTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.RemoveTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "No task found with id '9999999'")
			})
			Convey("existing task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron", false)

				p2 := c.RemoveTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)

				b := make([]byte, 5)
				rsp, err := c.do("DELETE", fmt.Sprintf("/tasks/%v", p1.ID), ContentTypeJSON, b) //case len(body) != 0
				So(rsp, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("do returns err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.RemoveTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "dial tcp: unknown port tcp/-1")
			})
		})

		Convey("GetTasks", func() {
			Convey("valid task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}
				p := c.CreateTask(sch, wf, "baron", false)
				So(p.Err, ShouldBeNil)
				So(p.Name, ShouldEqual, "baron")
				So(p.State, ShouldEqual, "Stopped")

				p2 := c.GetTasks()
				So(p2.Err, ShouldBeNil)
				p3 := c.GetTask(uint(p.ID))
				So(p3.Err, ShouldBeNil)
				p4 := c.GetTask(0)
				So(p4.Err, ShouldNotBeNil)
				So(p4.Err.Error(), ShouldEqual, "No task with Id '0'")
				So(p4.ScheduledTaskReturned, ShouldBeNil)
			})
			Convey("do returns err!=nil", func() {
				port := -1
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.GetTask(0)
				p2 := c.GetTasks()

				So(p.Err, ShouldNotBeNil)
				So(p2.Err, ShouldNotBeNil)
			})

		})
		Convey("WatchTasks", func() {
			Convey("event stream", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH2)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "10ms"}
				p := c.CreateTask(sch, wf, "baron", false)

				a := make([]string, 0)
				r := c.WatchTask(uint(p.ID))
				wait := make(chan struct{})
				go func() {
					for {
						select {
						case e := <-r.EventChan:
							a = append(a, e.EventType)
						case <-r.DoneChan:
							close(wait)
							return
						}
					}
				}()
				c.StopTask(p.ID)
				c.StartTask(p.ID)
				<-wait
				So(len(a), ShouldBeGreaterThanOrEqualTo, 10)
				So(a[0], ShouldEqual, "task-stopped")
				So(a[1], ShouldEqual, "task-started")
				for x := 2; x <= 10; x++ {
					So(a[x], ShouldEqual, "metric-event")
				}
				// Signal we are done
				r.Close()
			})

		})
	})
}
