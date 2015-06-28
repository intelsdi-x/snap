package rest

// This test runs through basic REST API calls and validates them.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/scheduler"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PULSE_PATH          = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH1  = PULSE_PATH + "/plugin/collector/pulse-collector-dummy1"
	DUMMY_PLUGIN_PATH2  = PULSE_PATH + "/plugin/collector/pulse-collector-dummy2"
	RIEMANN_PLUGIN_PATH = PULSE_PATH + "/plugin/publisher/pulse-publisher-riemann"

	NextPort = 8000
)

type restAPIInstance struct {
	port   int
	server *Server
}

func getPort() int {
	defer incrPort()
	return NextPort
}

func incrPort() {
	NextPort += 10
}

func command() string {
	return "curl"
}

func readBody(r *http.Response) []byte {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()
	return b
}

func getAPIResponse(resp *http.Response) (*APIResponse, string) {
	r := new(APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	return r, string(rb)
}

func getTasks(port int) (*APIResponse, string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func startTask(id, port int) (*APIResponse, string) {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%d/start", port, id)
	client := &http.Client{}
	b := bytes.NewReader([]byte{})
	req, err := http.NewRequest("PUT", uri, b)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func stopTask(id, port int) (*APIResponse, string) {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%d/stop", port, id)
	client := &http.Client{}
	b := bytes.NewReader([]byte{})
	req, err := http.NewRequest("PUT", uri, b)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func removeTask(id, port int) (*APIResponse, string) {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%d", port, id)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func createTask(sample, name, interval string, port int) (*APIResponse, string) {
	jsonP, err := ioutil.ReadFile("./wmap_sample/" + sample)
	if err != nil {
		log.Fatal(err)
	}
	wf, err := wmap.FromJson(jsonP)
	if err != nil {
		log.Fatal(err)
	}

	uri := fmt.Sprintf("http://localhost:%d/v1/tasks", port)

	t := request.TaskCreationRequest{
		Schedule: request.Schedule{Type: "simple", Interval: "1s"},
		Workflow: wf,
		Name:     name,
	}
	// Marshal to JSON for request body
	j, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	b := bytes.NewReader(j)
	req, err := http.NewRequest("POST", uri, b)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func uploadPlugin(pluginPath string, port int) (*APIResponse, string) {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins", port)

	client := &http.Client{}
	file, err := os.Open(pluginPath)
	if err != nil {
		log.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("pulse-plugins", filepath.Base(pluginPath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	file.Close()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func unloadPlugin(port int, name string, version int) (*APIResponse, string) {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%d", port, name, version)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

func getPluginList(port int) (*APIResponse, string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/plugins", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getMetricCatalog(port int) (*APIResponse, string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/metrics", port))
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI(port int) *restAPIInstance {
	// Start a REST API to talk to
	log.SetLevel(log.FatalLevel)
	r := New()
	c := control.New()
	c.Start()
	s := scheduler.New()
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.Start(":" + fmt.Sprint(port))
	time.Sleep(time.Millisecond * 100)
	return &restAPIInstance{
		port:   port,
		server: r,
	}
}

func TestPluginRestCalls(t *testing.T) {
	Convey("REST API functional V1", t, func() {
		Convey("Load Plugin - POST - /v1/plugins", func() {
			Convey("a single plugin loads", func() {
				port := getPort()
				startAPI(port) // Make this unique for each Convey hierarchy

				// The second argument here is a string from the HTTP response body
				// Useful to println if you want to see what the return looks like.
				r, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				plr := r.Body.(*rbody.PluginsLoaded)

				// We should have gotten out loaded plugin back
				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// Should only be one in the list
				r2, _ := getPluginList(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r2.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr2.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr2.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr2.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("load attempt to load same plugin", func() {
				port := getPort()
				startAPI(port)

				r, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				plr := r.Body.(*rbody.PluginsLoaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				r2, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr2 := r2.Body.(*rbody.Error)

				So(plr2.ResponseBodyType(), ShouldEqual, rbody.ErrorType)
				So(plr2.ResponseBodyMessage(), ShouldEqual, "plugin is already loaded")

				// Should only be one in the list
				r3, _ := getPluginList(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr3 := r3.Body.(*rbody.PluginListReturned)

				So(len(plr3.LoadedPlugins), ShouldEqual, 1)
				So(plr3.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr3.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr3.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr3.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr3.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("load two plugins", func() {
				port := getPort()
				startAPI(port)

				r, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				plr := r.Body.(*rbody.PluginsLoaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				r2, _ := uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				plr2 := r2.Body.(*rbody.PluginsLoaded)

				So(plr2.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr2.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy2(collector v2)")
				So(len(plr2.LoadedPlugins), ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Name, ShouldEqual, "dummy2")
				So(plr2.LoadedPlugins[0].Version, ShouldEqual, 2)
				So(plr2.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr2.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr2.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// Should be two in the list
				r3, _ := getPluginList(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr3 := r3.Body.(*rbody.PluginListReturned)

				So(len(plr3.LoadedPlugins), ShouldEqual, 2)
				So(plr3.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr3.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr3.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr3.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr3.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr3.LoadedPlugins[1].Name, ShouldEqual, "dummy2")
				So(plr3.LoadedPlugins[1].Version, ShouldEqual, 2)
				So(plr3.LoadedPlugins[1].Status, ShouldEqual, "loaded")
				So(plr3.LoadedPlugins[1].Type, ShouldEqual, "collector")
				So(plr3.LoadedPlugins[1].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
		})

		Convey("Unload Plugin - DELETE - /v1/plugins/:name/:version", func() {
			Convey("error in unload of unknown plugin", func() {
				port := getPort()
				startAPI(port)

				r, _ := unloadPlugin(port, "dummy1", 1)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr := r.Body.(*rbody.Error)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.ErrorType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "plugin not found (has it already been unloaded?)")
			})

			Convey("unload single plugin", func() {
				port := getPort()
				startAPI(port)
				// Load one
				r1, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))

				// Unload it now
				r, _ := unloadPlugin(port, "dummy1", 1)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
				plr := r.Body.(*rbody.PluginUnloaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy1v1)")
				So(plr.Name, ShouldEqual, "dummy1")
				So(plr.Version, ShouldEqual, 1)
				So(plr.Type, ShouldEqual, "collector")

				// Plugin should NOT be in the list
				r2, _ := getPluginList(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r2.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 0)
			})

			Convey("unload one of two plugins", func() {
				port := getPort()
				startAPI(port)
				// Load first
				r1, _ := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				// Load second
				r2, _ := uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))

				// Unload second
				r, _ := unloadPlugin(port, "dummy2", 2)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
				plr := r.Body.(*rbody.PluginUnloaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy2v2)")
				So(plr.Name, ShouldEqual, "dummy2")
				So(plr.Version, ShouldEqual, 2)
				So(plr.Type, ShouldEqual, "collector")

				r, _ = getPluginList(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Name, ShouldNotEqual, "dummy2")
				So(plr2.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr2.LoadedPlugins[0].Type, ShouldEqual, "collector")
			})
		})

		Convey("Plugin List - GET - /v1/plugins", func() {
			Convey("no plugins", func() {
				port := getPort()
				startAPI(port)

				r, _ := getPluginList(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr := r.Body.(*rbody.PluginListReturned)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list returned")
				So(len(plr.LoadedPlugins), ShouldEqual, 0)
				So(len(plr.AvailablePlugins), ShouldEqual, 0)
			})

			Convey("one plugin in list", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH1, port)

				r, _ := getPluginList(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr := r.Body.(*rbody.PluginListReturned)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list returned")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(len(plr.AvailablePlugins), ShouldEqual, 0)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("multiple plugins in list", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				uploadPlugin(DUMMY_PLUGIN_PATH2, port)

				r, _ := getPluginList(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr := r.Body.(*rbody.PluginListReturned)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list returned")
				So(len(plr.LoadedPlugins), ShouldEqual, 2)
				So(len(plr.AvailablePlugins), ShouldEqual, 0)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				//
				So(plr.LoadedPlugins[1].Name, ShouldEqual, "dummy2")
				So(plr.LoadedPlugins[1].Version, ShouldEqual, 2)
				So(plr.LoadedPlugins[1].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[1].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[1].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
		})

		Convey("Metric Catalog - GET - /v1/metrics", func() {
			Convey("empty catalog", func() {
				port := getPort()
				startAPI(port)

				r, _ := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr := r.Body.(*rbody.MetricCatalogReturned)

				So(len(plr.Catalog), ShouldEqual, 0)
			})

			Convey("plugin metrics show up in the catalog", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r, _ := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr := r.Body.(*rbody.MetricCatalogReturned)

				So(len(plr.Catalog), ShouldEqual, 2)
				So(plr.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr.Catalog[0].Versions), ShouldEqual, 1)
				So(plr.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(len(plr.Catalog[1].Versions), ShouldEqual, 1)
				So(plr.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(plr.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("newer plugin upgrades the metrics", func() {
				port := getPort()
				startAPI(port)

				// upload v1
				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r, _ := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr := r.Body.(*rbody.MetricCatalogReturned)

				So(len(plr.Catalog), ShouldEqual, 2)
				So(plr.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr.Catalog[0].Versions), ShouldEqual, 1)
				So(plr.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(plr.Catalog[1].Versions), ShouldEqual, 1)
				So(plr.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// upload v2
				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				r2, _ := getMetricCatalog(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr2 := r2.Body.(*rbody.MetricCatalogReturned)

				So(len(plr2.Catalog), ShouldEqual, 2)
				So(plr2.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr2.Catalog[0].Versions), ShouldEqual, 2)
				So(plr2.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr2.Catalog[1].Versions["2"], ShouldNotBeNil)
				So(plr2.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr2.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(plr2.Catalog[1].Versions), ShouldEqual, 2)
				So(plr2.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr2.Catalog[1].Versions["2"], ShouldNotBeNil)
				So(plr2.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("removing a newer plugin downgrades the metrics", func() {
				port := getPort()
				startAPI(port)

				// upload v1
				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r, _ := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr := r.Body.(*rbody.MetricCatalogReturned)

				So(plr.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr.Catalog[0].Versions), ShouldEqual, 1)
				So(plr.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(plr.Catalog[1].Versions), ShouldEqual, 1)
				So(plr.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// upload v2
				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				r2, _ := getMetricCatalog(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr2 := r2.Body.(*rbody.MetricCatalogReturned)

				So(plr2.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr2.Catalog[0].Versions), ShouldEqual, 2)
				So(plr2.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr2.Catalog[0].Versions["2"], ShouldNotBeNil)
				So(plr2.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr2.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(plr2.Catalog[1].Versions), ShouldEqual, 2)
				So(plr2.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr2.Catalog[1].Versions["2"], ShouldNotBeNil)
				So(plr2.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// remove v2
				unloadPlugin(port, "dummy2", 2)
				r3, _ := getMetricCatalog(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.MetricCatalogReturned))
				plr3 := r3.Body.(*rbody.MetricCatalogReturned)

				So(plr3.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So(len(plr3.Catalog[0].Versions), ShouldEqual, 1)
				So(plr3.Catalog[0].Versions["1"], ShouldNotBeNil)
				So(plr3.Catalog[0].Versions["2"], ShouldBeNil)
				So(plr3.Catalog[0].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr3.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So(len(plr3.Catalog[1].Versions), ShouldEqual, 1)
				So(plr3.Catalog[1].Versions["1"], ShouldNotBeNil)
				So(plr3.Catalog[1].Versions["2"], ShouldBeNil)
				So(plr3.Catalog[1].Versions["1"].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
		})

		Convey("Create Task - POST - /v1/tasks", func() {
			Convey("creating task with missing metric errors", func() {
				port := getPort()
				startAPI(port)

				r, _ := createTask("1.json", "foo", "1s", port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr := r.Body.(*rbody.Error)
				So(plr.ErrorMessage, ShouldEqual, "metric not found")
			})

			Convey("create task works when plugins are loaded", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)
				r, _ := createTask("1.json", "foo", "1s", port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr := r.Body.(*rbody.AddScheduledTask)
				So(plr.CreationTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So(plr.Name, ShouldEqual, "foo")
				So(plr.HitCount, ShouldEqual, 0)
				So(plr.FailedCount, ShouldEqual, 0)
				So(plr.MissCount, ShouldEqual, 0)
				So(plr.State, ShouldEqual, "Stopped")
				So(plr.Deadline, ShouldEqual, "5s")
			})

		})

		Convey("Get Tasks - GET - /v1/tasks", func() {
			Convey("get tasks after single task added", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)
				r, _ := createTask("1.json", "bar", "1s", port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r2, _ := getTasks(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr2 := r2.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr2.ScheduledTasks), ShouldEqual, 1)
				So(plr2.ScheduledTasks[0].Name, ShouldEqual, "bar")
			})

			Convey("get tasks after multiple tasks added", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)

				r1, _ := createTask("1.json", "alpha", "1s", port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r2, _ := createTask("1.json", "beta", "1s", port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r3, _ := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 2)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "alpha")
				So(plr3.ScheduledTasks[1].Name, ShouldEqual, "beta")
			})
		})

		Convey("Start Task - PUT - /v1/tasks/:id/start", func() {
			Convey("starts after being created", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)

				r1, _ := createTask("1.json", "xenu", "1s", port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2, _ := startTask(id, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStarted))
				plr2 := r2.Body.(*rbody.ScheduledTaskStarted)
				So(plr2.ID, ShouldEqual, id)

				r3, _ := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 1)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "xenu")
				So(plr3.ScheduledTasks[0].State, ShouldEqual, "Spinning")
			})
		})

		Convey("Stop Task - PUT - /v1/tasks/:id/stop", func() {
			Convey("stops after being started", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)

				r1, _ := createTask("1.json", "yeti", "1s", port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2, _ := startTask(id, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStarted))
				plr2 := r2.Body.(*rbody.ScheduledTaskStarted)
				So(plr2.ID, ShouldEqual, id)

				r3, _ := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 1)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr3.ScheduledTasks[0].State, ShouldEqual, "Spinning")

				r4, _ := stopTask(id, port)
				So(r4.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStopped))
				plr4 := r4.Body.(*rbody.ScheduledTaskStopped)
				So(plr4.ID, ShouldEqual, id)

				r5, _ := getTasks(port)
				So(r5.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr5 := r5.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr5.ScheduledTasks), ShouldEqual, 1)
				So(plr5.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr5.ScheduledTasks[0].State, ShouldEqual, "Stopped")
			})
		})

		Convey("Remove Task - DELETE - /v1/tasks/:id", func() {
			Convey("error on trying to remove unknown plugin", func() {
				port := getPort()
				startAPI(port)

				r1, _ := removeTask(99999, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr1 := r1.Body.(*rbody.Error)
				So(plr1.ErrorMessage, ShouldEqual, "Task not found")
			})
			Convey("removes a task", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(RIEMANN_PLUGIN_PATH, port)

				r1, _ := createTask("1.json", "yeti", "1s", port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2, _ := getTasks(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr2 := r2.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr2.ScheduledTasks), ShouldEqual, 1)
				So(plr2.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr2.ScheduledTasks[0].State, ShouldEqual, "Stopped")

				r3, _ := removeTask(id, port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskRemoved))
				plr3 := r3.Body.(*rbody.ScheduledTaskRemoved)
				So(plr3.ID, ShouldEqual, id)

				r4, _ := getTasks(port)
				So(r4.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr4 := r4.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr4.ScheduledTasks), ShouldEqual, 0)
			})
		})
	})
}
