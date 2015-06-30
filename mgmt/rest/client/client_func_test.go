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
	PULSE_PATH          = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH1  = PULSE_PATH + "/plugin/collector/pulse-collector-dummy1"
	DUMMY_PLUGIN_PATH2  = PULSE_PATH + "/plugin/collector/pulse-collector-dummy2"
	RIEMANN_PLUGIN_PATH = PULSE_PATH + "/plugin/publisher/pulse-publisher-riemann"

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
	log.SetLevel(log.FatalLevel)
	r := rest.New()
	c := control.New()
	c.Start()
	s := scheduler.New()
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.Start(":" + fmt.Sprint(port))
	time.Sleep(time.Millisecond * 100)
	return fmt.Sprintf("http://localhost:%d", port)
}

func TestPulseClient(t *testing.T) {
	Convey("REST API functional V1", t, func() {
		Convey("GetPlugins", func() {
			Convey("empty list", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.GetPlugins(false)

				So(p.Err, ShouldBeNil)
				So(len(p.LoadedPlugins), ShouldEqual, 0)
				So(p.AvailablePlugins, ShouldBeEmpty)
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
		})
		Convey("LoadPlugin", func() {
			Convey("single load", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				So(p.Err, ShouldBeNil)
				So(p.LoadedPlugins, ShouldNotBeEmpty)
				So(p.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(p.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(p.LoadedPlugins[0].LoadedTime().Unix(), ShouldBeLessThanOrEqualTo, time.Now().Unix())
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
		})
		Convey("UnloadPlugin", func() {
			Convey("unload unknown plugin", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.UnloadPlugin("foo", 3)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "plugin not found (has it already been unloaded?)")
			})

			Convey("unload only one there is", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				p := c.UnloadPlugin("dummy1", 1)
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
				p1 := c.UnloadPlugin("dummy2", 2)
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
				So(p.Len(), ShouldEqual, 2)
				So(p.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(p.Catalog[0].Versions), ShouldEqual, 2)
				So(p.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(p.Catalog[1].Versions), ShouldEqual, 2)
			})
		})
		Convey("CreateTask", func() {
			Convey("invalid task (missing metric)", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron")
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "metric not found")
			})
			Convey("invalid task (missing publisher)", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron")
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "Loaded plugin not found")
			})
			Convey("valid task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				wf := getWMFromSample("1.json")
				sch := &Schedule{Type: "simple", Interval: "1s"}

				p := c.CreateTask(sch, wf, "baron")
				So(p.Err, ShouldBeNil)
				So(p.Name, ShouldEqual, "baron")
				So(p.State, ShouldEqual, "Stopped")
			})
		})
		Convey("StartTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				p := c.StartTask(9999999)
				So(p.Err, ShouldNotBeNil)
				So(p.Err.Error(), ShouldEqual, "No task found with id '9999999'")
			})
			Convey("existing task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron")

				p2 := c.StartTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)
			})
		})
		Convey("StopTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron")

				p2 := c.StopTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)
			})
		})
		Convey("RemoveTask", func() {
			Convey("unknown task", func() {
				port := getPort()
				uri := startAPI(port)
				c := New(uri, "v1")

				c.LoadPlugin(DUMMY_PLUGIN_PATH1)
				c.LoadPlugin(RIEMANN_PLUGIN_PATH)

				p1 := c.CreateTask(&Schedule{Type: "simple", Interval: "1s"}, getWMFromSample("1.json"), "baron")

				p2 := c.RemoveTask(p1.ID)
				So(p2.Err, ShouldBeNil)
				So(p2.ID, ShouldEqual, p1.ID)
			})
		})
	})
}
