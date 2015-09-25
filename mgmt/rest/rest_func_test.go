package rest

// This test runs through basic REST API calls and validates them.

import (
	"bufio"
	"bytes"
	"compress/gzip"
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
	// Switching this turns on logging for all the REST API calls
	LOG_LEVEL = log.FatalLevel

	PULSE_PATH         = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH1 = PULSE_PATH + "/plugin/pulse-collector-dummy1"
	DUMMY_PLUGIN_PATH2 = PULSE_PATH + "/plugin/pulse-collector-dummy2"
	PSUTIL_PLUGIN_PATH = PULSE_PATH + "/plugin/pulse-collector-psutil"
	FILE_PLUGIN_PATH   = PULSE_PATH + "/plugin/pulse-publisher-file"

	NextPort         = 40000
	CompressedUpload = true
	TotalUploadSize  = 0
	UploadCount      = 0
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

func getAPIResponse(resp *http.Response) *APIResponse {
	r := new(APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	r.JSONResponse = string(rb)
	return r
}

func getStreamingAPIResponse(resp *http.Response) *APIResponse {
	r := new(APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	r.JSONResponse = string(rb)
	return r
}

type watchTaskResult struct {
	eventChan chan string
	doneChan  chan struct{}
	killChan  chan struct{}
}

func (w *watchTaskResult) close() {
	close(w.doneChan)
}

func watchTask(id, port int) *watchTaskResult {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks/%d/watch", port, id))
	if err != nil {
		log.Fatal(err)
	}

	r := &watchTaskResult{
		eventChan: make(chan string),
		doneChan:  make(chan struct{}),
		killChan:  make(chan struct{}),
	}
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-r.doneChan:
				resp.Body.Close()
				return
			default:
				line, _ := reader.ReadBytes('\n')
				ste := &rbody.StreamedTaskEvent{}
				err := json.Unmarshal(line, ste)
				if err != nil {
					r.close()
					return
				}
				switch ste.EventType {
				case rbody.TaskWatchTaskDisabled:
					r.eventChan <- ste.EventType
					r.close()
					return
				case rbody.TaskWatchTaskStopped, rbody.TaskWatchTaskStarted, rbody.TaskWatchMetricEvent:
					r.eventChan <- ste.EventType
				}
			}
		}
	}()
	return r
}

func getTasks(port int) *APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getTask(id, port int) *APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks/%d", port, id))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func startTask(id, port int) *APIResponse {
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

func stopTask(id, port int) *APIResponse {
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

func removeTask(id, port int) *APIResponse {
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

func createTask(sample, name, interval string, noStart bool, port int) *APIResponse {
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
		Schedule: request.Schedule{Type: "simple", Interval: interval},
		Workflow: wf,
		Name:     name,
		Start:    !noStart,
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

func uploadPlugin(pluginPath string, port int) *APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins", port)

	client := &http.Client{}
	file, err := os.Open(pluginPath)
	if err != nil {
		log.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	var part io.Writer
	part, err = writer.CreateFormFile("pulse-plugins", filepath.Base(pluginPath))
	if err != nil {
		log.Fatal(err)
	}
	if CompressedUpload {
		cpart := gzip.NewWriter(part)
		_, err = io.Copy(cpart, file)
		if err != nil {
			log.Fatal(err)
		}
		err = cpart.Close()
	} else {
		_, err = io.Copy(part, file)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}
	TotalUploadSize += body.Len()
	UploadCount += 1
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	if CompressedUpload {
		req.Header.Add("Plugin-Compression", "gzip")
	}
	file.Close()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func unloadPlugin(port int, pluginType string, name string, version int) *APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%d", port, pluginType, name, version)
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

func getPluginList(port int) *APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/plugins", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getMetricCatalog(port int) *APIResponse {
	return fetchMetrics(port, "")
}

func fetchMetrics(port int, ns string) *APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/metrics%s", port, ns))
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

func fetchMetricsWithVersion(port int, ns string, ver int) *APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/metrics%s?ver=%d", port, ns, ver))
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI(port int) *restAPIInstance {
	// Start a REST API to talk to
	log.SetLevel(LOG_LEVEL)
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
	CompressedUpload = false
	Convey("REST API functional V1", t, func() {
		Convey("Load Plugin - POST - /v1/plugins", func() {
			Convey("a single plugin loads", func() {
				// This test alone tests gzip. Saves on test time.
				CompressedUpload = true
				port := getPort()
				startAPI(port) // Make this unique for each Convey hierarchy

				// The second argument here is a string from the HTTP response body
				// Useful to println if you want to see what the return looks like.
				r := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
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
				r2 := getPluginList(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r2.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr2.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr2.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr2.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				CompressedUpload = false
			})

			Convey("load attempt to load same plugin", func() {
				port := getPort()
				startAPI(port)

				r := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
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

				r2 := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr2 := r2.Body.(*rbody.Error)

				So(plr2.ResponseBodyType(), ShouldEqual, rbody.ErrorType)
				So(plr2.ResponseBodyMessage(), ShouldEqual, "plugin is already loaded")

				// Should only be one in the list
				r3 := getPluginList(port)
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

				r := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
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

				r2 := uploadPlugin(DUMMY_PLUGIN_PATH2, port)
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
				r3 := getPluginList(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr3 := r3.Body.(*rbody.PluginListReturned)

				So(len(plr3.LoadedPlugins), ShouldEqual, 2)
				So(plr3.LoadedPlugins[0].Name, ShouldContainSubstring, "dummy")
				So(plr3.LoadedPlugins[0].Version, ShouldBeIn, []int{1, 2})
				So(plr3.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr3.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr3.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				if plr3.LoadedPlugins[0].Name == "dummy1" {
					So(plr3.LoadedPlugins[1].Name, ShouldEqual, "dummy2")
					So(plr3.LoadedPlugins[1].Version, ShouldEqual, 2)
				} else {
					So(plr3.LoadedPlugins[1].Name, ShouldEqual, "dummy1")
					So(plr3.LoadedPlugins[1].Version, ShouldEqual, 1)
				}
				So(plr3.LoadedPlugins[1].Status, ShouldEqual, "loaded")
				So(plr3.LoadedPlugins[1].Type, ShouldEqual, "collector")
				So(plr3.LoadedPlugins[1].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
		})

		Convey("Unload Plugin - DELETE - /v1/plugins/:name/:version", func() {
			Convey("error in unload of unknown plugin", func() {
				port := getPort()
				startAPI(port)

				r := unloadPlugin(port, "collector", "dummy1", 1)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr := r.Body.(*rbody.Error)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.ErrorType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "plugin not found")
			})

			Convey("unload single plugin", func() {
				port := getPort()
				startAPI(port)
				// Load one
				r1 := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))

				// Unload it now
				r := unloadPlugin(port, "collector", "dummy1", 1)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
				plr := r.Body.(*rbody.PluginUnloaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy1v1)")
				So(plr.Name, ShouldEqual, "dummy1")
				So(plr.Version, ShouldEqual, 1)
				So(plr.Type, ShouldEqual, "collector")

				// Plugin should NOT be in the list
				r2 := getPluginList(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r2.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 0)
			})

			Convey("unload one of two plugins", func() {
				port := getPort()
				startAPI(port)
				// Load first
				r1 := uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))
				// Load second
				r2 := uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))

				// Unload second
				r := unloadPlugin(port, "collector", "dummy2", 2)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
				plr := r.Body.(*rbody.PluginUnloaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy2v2)")
				So(plr.Name, ShouldEqual, "dummy2")
				So(plr.Version, ShouldEqual, 2)
				So(plr.Type, ShouldEqual, "collector")

				r = getPluginList(port)
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

				r := getPluginList(port)
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

				r := getPluginList(port)
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

				r := getPluginList(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr := r.Body.(*rbody.PluginListReturned)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list returned")
				So(len(plr.LoadedPlugins), ShouldEqual, 2)
				So(len(plr.AvailablePlugins), ShouldEqual, 0)
				var (
					x, y int
				)
				if plr.LoadedPlugins[0].Name == "dummy1" {
					y = 1
				} else {
					x = 1
				}
				So(plr.LoadedPlugins[x].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[x].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[x].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[x].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[x].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				//
				So(plr.LoadedPlugins[y].Name, ShouldEqual, "dummy2")
				So(plr.LoadedPlugins[y].Version, ShouldEqual, 2)
				So(plr.LoadedPlugins[y].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[y].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[y].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})
		})

		Convey("Metric Catalog - GET - /v1/metrics", func() {
			Convey("empty catalog", func() {
				port := getPort()
				startAPI(port)

				r := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr := r.Body.(*rbody.MetricsReturned)

				So(len(*plr), ShouldEqual, 0)
			})

			Convey("plugin metrics show up in the catalog", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr := r.Body.(*rbody.MetricsReturned)

				So(len(*plr), ShouldEqual, 2)
				So((*plr)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("newer plugin upgrades the metrics", func() {
				port := getPort()
				startAPI(port)

				// upload v1
				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr := r.Body.(*rbody.MetricsReturned)

				So(len(*plr), ShouldEqual, 2)
				So((*plr)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// upload v2
				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				r2 := getMetricCatalog(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr2 := r2.Body.(*rbody.MetricsReturned)

				So(len(*plr2), ShouldEqual, 4)
				So((*plr2)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr2)[0].Version, ShouldEqual, 1)
				So((*plr2)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[1].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr2)[1].Version, ShouldEqual, 2)
				So((*plr2)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[2].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr2)[2].Version, ShouldEqual, 1)
				So((*plr2)[2].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[3].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr2)[3].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[3].Version, ShouldEqual, 2)

			})

			Convey("removing a newer plugin downgrades the metrics", func() {
				port := getPort()
				startAPI(port)

				// upload v1
				uploadPlugin(DUMMY_PLUGIN_PATH1, port)
				r := getMetricCatalog(port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr := r.Body.(*rbody.MetricsReturned)

				So(len(*plr), ShouldEqual, 2)
				So((*plr)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				// upload v2
				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				r2 := getMetricCatalog(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr2 := r2.Body.(*rbody.MetricsReturned)

				So(len(*plr2), ShouldEqual, 4)
				So((*plr2)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr2)[0].Version, ShouldEqual, 1)
				So((*plr2)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[1].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr2)[1].Version, ShouldEqual, 2)
				So((*plr2)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[2].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr2)[2].Version, ShouldEqual, 1)
				So((*plr2)[2].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[3].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr2)[3].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr2)[3].Version, ShouldEqual, 2)

				// remove v2
				unloadPlugin(port, "collector", "dummy2", 2)
				r3 := getMetricCatalog(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
				plr3 := r3.Body.(*rbody.MetricsReturned)

				So(len(*plr3), ShouldEqual, 2)
				So((*plr3)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
				So((*plr3)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
				So((*plr3)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
				So((*plr3)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

			})
		})
		Convey("metrics accessible via tree-like lookup", func() {
			port := getPort()
			startAPI(port)

			uploadPlugin(DUMMY_PLUGIN_PATH1, port)
			r := fetchMetrics(port, "/intel/dummy/*")
			So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
			plr := r.Body.(*rbody.MetricsReturned)

			So(len(*plr), ShouldEqual, 2)
			So((*plr)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
			So((*plr)[0].Version, ShouldEqual, 1)
			So((*plr)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			So((*plr)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
			So((*plr)[1].Version, ShouldEqual, 1)
			So((*plr)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

		})

		Convey("metrics with version accessible via tree-like lookup", func() {
			port := getPort()
			startAPI(port)

			uploadPlugin(DUMMY_PLUGIN_PATH1, port)
			uploadPlugin(DUMMY_PLUGIN_PATH2, port)
			r := fetchMetricsWithVersion(port, "/intel/dummy/*", 2)
			So(r.Body, ShouldHaveSameTypeAs, new(rbody.MetricsReturned))
			plr := r.Body.(*rbody.MetricsReturned)

			So(len(*plr), ShouldEqual, 2)
			So((*plr)[0].Namespace, ShouldEqual, "/intel/dummy/bar")
			So((*plr)[0].Version, ShouldEqual, 2)
			So((*plr)[0].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			So((*plr)[1].Namespace, ShouldEqual, "/intel/dummy/foo")
			So((*plr)[1].Version, ShouldEqual, 2)
			So((*plr)[1].LastAdvertisedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

		})

		Convey("Create Task - POST - /v1/tasks", func() {
			Convey("creating task with missing metric errors", func() {
				port := getPort()
				startAPI(port)

				r := createTask("1.json", "foo", "1s", true, port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr := r.Body.(*rbody.Error)
				So(plr.ErrorMessage, ShouldContainSubstring, "Metric not found: /intel/dummy/foo")
			})

			Convey("create task works when plugins are loaded", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)
				r := createTask("1.json", "foo", "1s", true, port)
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
				uploadPlugin(FILE_PLUGIN_PATH, port)
				r := createTask("1.json", "bar", "1s", true, port)
				So(r.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r2 := getTasks(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr2 := r2.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr2.ScheduledTasks), ShouldEqual, 1)
				So(plr2.ScheduledTasks[0].Name, ShouldEqual, "bar")
			})

			Convey("get tasks after multiple tasks added", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "alpha", "1s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r2 := createTask("1.json", "beta", "1s", true, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

				r3 := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 2)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "alpha")
				So(plr3.ScheduledTasks[1].Name, ShouldEqual, "beta")
			})
		})

		Convey("Get Task By ID - GET - /v1/tasks/:id", func() {
			Convey("get task after task added", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)
				r1 := createTask("1.json", "foo", "3s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				t1 := r1.Body.(*rbody.AddScheduledTask)
				r2 := getTask(t1.ID, port)
				t2 := r2.Body.(*rbody.ScheduledTaskReturned)
				So(t2.AddScheduledTask.Name, ShouldEqual, "foo")
			})
		})

		Convey("Start Task - PUT - /v1/tasks/:id/start", func() {
			Convey("starts after being created", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "xenu", "1s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2 := startTask(id, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStarted))
				plr2 := r2.Body.(*rbody.ScheduledTaskStarted)
				So(plr2.ID, ShouldEqual, id)

				r3 := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 1)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "xenu")
				So(plr3.ScheduledTasks[0].State, ShouldEqual, "Running")

				// cleanup for test perf reasons
				removeTask(id, port)
			})
			Convey("starts when created", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "xenu", "1s", false, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r3 := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 1)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "xenu")
				So(plr3.ScheduledTasks[0].State, ShouldEqual, "Running")

				// cleanup for test perf reasons
				removeTask(id, port)
			})
		})

		Convey("Stop Task - PUT - /v1/tasks/:id/stop", func() {
			Convey("stops after being started", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "yeti", "1s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2 := startTask(id, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStarted))
				plr2 := r2.Body.(*rbody.ScheduledTaskStarted)
				So(plr2.ID, ShouldEqual, id)

				r3 := getTasks(port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr3 := r3.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr3.ScheduledTasks), ShouldEqual, 1)
				So(plr3.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr3.ScheduledTasks[0].State, ShouldEqual, "Running")

				r4 := stopTask(id, port)
				So(r4.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStopped))
				plr4 := r4.Body.(*rbody.ScheduledTaskStopped)
				So(plr4.ID, ShouldEqual, id)

				time.Sleep(1 * time.Second)

				r5 := getTasks(port)
				So(r5.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr5 := r5.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr5.ScheduledTasks), ShouldEqual, 1)
				So(plr5.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr5.ScheduledTasks[0].State, ShouldEqual, "Stopped")
			})
		})

		Convey("Remove Task - DELETE - /v1/tasks/:id", func() {
			Convey("error on trying to remove unknown task", func() {
				port := getPort()
				startAPI(port)

				r1 := removeTask(99999, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr1 := r1.Body.(*rbody.Error)
				So(plr1.ErrorMessage, ShouldEqual, "No task found with id '99999'")
			})
			Convey("removes a task", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "yeti", "1s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2 := getTasks(port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr2 := r2.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr2.ScheduledTasks), ShouldEqual, 1)
				So(plr2.ScheduledTasks[0].Name, ShouldEqual, "yeti")
				So(plr2.ScheduledTasks[0].State, ShouldEqual, "Stopped")

				r3 := removeTask(id, port)
				So(r3.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskRemoved))
				plr3 := r3.Body.(*rbody.ScheduledTaskRemoved)
				So(plr3.ID, ShouldEqual, id)

				r4 := getTasks(port)
				So(r4.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskListReturned))
				plr4 := r4.Body.(*rbody.ScheduledTaskListReturned)
				So(len(plr4.ScheduledTasks), ShouldEqual, 0)
			})
		})
		Convey("Watch task - get - /v1/tasks/:id/watch", func() {
			Convey("---", func() {
				port := getPort()
				startAPI(port)

				uploadPlugin(DUMMY_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)
				uploadPlugin(PSUTIL_PLUGIN_PATH, port)

				r1 := createTask("1.json", "xenu", "10ms", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				// Change buffer window to 10ms (do not do this IRL)
				StreamingBufferWindow = 0.01
				r := watchTask(id, port)
				time.Sleep(time.Millisecond * 100)
				startTask(id, port)
				var events []string
				wait := make(chan struct{})
				go func() {
					for {
						select {
						case e := <-r.eventChan:
							events = append(events, e)
							if len(events) == 10 {
								r.close()
							}
						case <-r.doneChan:
							close(wait)
							return
						}
					}
				}()
				<-wait
				stopTask(id, port)
				So(len(events), ShouldBeGreaterThanOrEqualTo, 0)
				So(events[0], ShouldEqual, "task-started")
				for x := 1; x <= 9; x++ {
					So(events[x], ShouldEqual, "metric-event")
				}
			})
		})
	})
}
