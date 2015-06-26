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
	"github.com/intelsdi-x/pulse/scheduler"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PULSE_PATH         = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH1 = PULSE_PATH + "/plugin/collector/pulse-collector-dummy1"
	DUMMY_PLUGIN_PATH2 = PULSE_PATH + "/plugin/collector/pulse-collector-dummy2"
)

type restAPIInstance struct {
	port   int
	server *Server
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

func uploadPlugin(pluginPath string, port int) *http.Response {
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
	return resp
}

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI(port int) *restAPIInstance {
	// Start a REST API to talk to
	log.SetLevel(log.WarnLevel)
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
			Convey("load one plugin", func() {
				startAPI(8000) // Make this unique for each Convey hierarchy

				resp := uploadPlugin(DUMMY_PLUGIN_PATH1, 8000)

				r, _ := getAPIResponse(resp)
				plr := r.Body.(*rbody.PluginsLoaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

			Convey("load two plugins", func() {
				startAPI(8001)

				resp := uploadPlugin(DUMMY_PLUGIN_PATH1, 8001)

				r, _ := getAPIResponse(resp)
				plr := r.Body.(*rbody.PluginsLoaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
				So(len(plr.LoadedPlugins), ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
				So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
				So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

				resp2 := uploadPlugin(DUMMY_PLUGIN_PATH2, 8001)

				r2, _ := getAPIResponse(resp2)
				plr2 := r2.Body.(*rbody.PluginsLoaded)

				So(plr2.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
				So(plr2.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy2(collector v2)")
				So(len(plr2.LoadedPlugins), ShouldEqual, 1)
				So(plr2.LoadedPlugins[0].Name, ShouldEqual, "dummy2")
				So(plr2.LoadedPlugins[0].Version, ShouldEqual, 2)
				So(plr2.LoadedPlugins[0].Status, ShouldEqual, "loaded")
				So(plr2.LoadedPlugins[0].Type, ShouldEqual, "collector")
				So(plr2.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
			})

		})

		Convey("Unload Plugin - DELETE - /v1/plugins/:name/:version", func() {
			Convey("Attempt unload on unknown plugin", func() {
				startAPI(8002)

				client := &http.Client{}
				req, err := http.NewRequest("DELETE", "http://localhost:8002/v1/plugins/dummy2/2", nil)
				if err != nil {
					log.Fatal(err)
				}
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}

				r, _ := getAPIResponse(resp)
				// print(rs)

				So(r.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr := r.Body.(*rbody.Error)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.ErrorType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "plugin not found (has it already been unloaded?)")
			})

			Convey("unload plugin when it is the only one", func() {
				startAPI(8003)
				// Load one
				resp1 := uploadPlugin(DUMMY_PLUGIN_PATH1, 8003)
				r1, _ := getAPIResponse(resp1)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.PluginsLoaded))

				// Unload it now
				client := &http.Client{}
				req, err := http.NewRequest("DELETE", "http://localhost:8003/v1/plugins/dummy1/1", nil)
				if err != nil {
					log.Fatal(err)
				}
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}

				r, _ := getAPIResponse(resp)
				// print(rs)

				So(r.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
				plr := r.Body.(*rbody.PluginUnloaded)

				So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
				So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy1v1)")
				So(plr.Name, ShouldEqual, "dummy1")
				So(plr.Version, ShouldEqual, 1)
				So(plr.Type, ShouldEqual, "collector")

				// Plugin should NOT be in the list
				resp, err = http.Get("http://localhost:8003/v1/plugins")
				if err != nil {
					log.Fatal(err)
				}

				r2, _ := getAPIResponse(resp)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.PluginListReturned))
				plr2 := r2.Body.(*rbody.PluginListReturned)

				So(len(plr2.LoadedPlugins), ShouldEqual, 0)
			})

			// Convey("unload plugin when there are two", func() {
			// 	startAPI(8002)

			// 	client := &http.Client{}
			// 	req, err := http.NewRequest("DELETE", "http://localhost:8002/v1/plugins/dummy2/2", nil)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	resp, err := client.Do(req)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	r, _ := getAPIResponse(resp)
			// 	// print(rs)
			// 	plr := r.Body.(*rbody.PluginUnloaded)

			// 	So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginUnloadedType)
			// 	So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin successfuly unloaded (dummy2v2)")
			// 	So(plr.Name, ShouldEqual, "dummy2")
			// 	So(plr.Version, ShouldEqual, 2)
			// 	So(plr.Type, ShouldEqual, "collector")

			// 	// Plugin should NOT be in the list
			// 	resp, err = http.Get("http://localhost:8181/v1/plugins")
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	r, _ = getAPIResponse(resp)
			// 	// print("\n" + rs)
			// 	plr2 := r.Body.(*rbody.PluginListReturned)

			// 	So(len(plr2.LoadedPlugins), ShouldEqual, 1)
			// 	So(plr2.LoadedPlugins[0].Name, ShouldNotEqual, "dummy2")
			// })

		})

		// Convey("Plugin List - GET - /v1/plugins", func() {
		// 	startAPI(8003)

		// 	resp, err := http.Get("http://localhost:8003/v1/plugins")
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}

		// 	r, rs := getAPIResponse(resp)
		// 	print("\n" + rs)
		// 	plr := r.Body.(*rbody.PluginListReturned)

		// 	So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
		// 	So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list returned")
		// 	So(len(plr.LoadedPlugins), ShouldEqual, 2)
		// 	So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
		// 	So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
		// 	So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
		// 	So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
		// 	So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
		// 	So(plr.LoadedPlugins[1].Name, ShouldEqual, "dummy2")
		// 	So(plr.LoadedPlugins[1].Version, ShouldEqual, 2)
		// 	So(plr.LoadedPlugins[1].Status, ShouldEqual, "loaded")
		// 	So(plr.LoadedPlugins[1].Type, ShouldEqual, "collector")
		// 	So(plr.LoadedPlugins[1].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
		// })

		// Convey("Metric Catalog - GET - /v1/metrics", func() {
		// 	startAPI(8004)

		// 	resp, err := http.Get("http://localhost:8004/v1/metrics")
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}

		// 	r, _ := getAPIResponse(resp)
		// 	// print("\n" + rs)
		// 	plr := r.Body.(*rbody.MetricCatalogReturned)

		// 	So(plr.ResponseBodyType(), ShouldEqual, rbody.MetricCatalogReturnedType)
		// 	So(plr.ResponseBodyMessage(), ShouldEqual, "Metric catalog returned")
		// 	So(len(plr.Catalog), ShouldEqual, 2)
		// 	So(plr.Catalog[0].Namespace, ShouldEqual, "/intel/dummy/foo")
		// 	So(plr.Catalog[0].Version, ShouldEqual, 2)
		// 	So(plr.Catalog[1].Namespace, ShouldEqual, "/intel/dummy/bar")
		// 	So(plr.Catalog[1].Version, ShouldEqual, 2)
		// })

	})
}
