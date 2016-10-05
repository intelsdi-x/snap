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

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/plugin/helper"
	"github.com/intelsdi-x/snap/scheduler"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	// Switching this turns on logging for all the REST API calls
	LOG_LEVEL = log.WarnLevel

	SNAP_PATH              = helper.BuildPath
	SNAP_AUTODISCOVER_PATH = os.Getenv("SNAP_AUTODISCOVER_PATH")
	MOCK_PLUGIN_PATH1      = helper.PluginFilePath("snap-plugin-collector-mock1")
	MOCK_PLUGIN_PATH2      = helper.PluginFilePath("snap-plugin-collector-mock2")
	FILE_PLUGIN_PATH       = helper.PluginFilePath("snap-plugin-publisher-mock-file")

	CompressedUpload = true
	TotalUploadSize  = 0
	UploadCount      = 0
)

const (
	MOCK_CONSTRAINTS = `{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "snapd global config schema",
		"type": ["object", "null"],
		"properties": {
			"control": { "$ref": "#/definitions/control" },
			"scheduler": { "$ref": "#/definitions/scheduler"},
			"restapi" : { "$ref": "#/definitions/restapi"},
			"tribe": { "$ref": "#/definitions/tribe"}
		},
		"additionalProperties": true,
		"definitions": { ` +
		`"control": {}, "scheduler": {}, ` + CONFIG_CONSTRAINTS + `, "tribe":{}` +
		`}` +
		`}`
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

func getAPIResponse(resp *http.Response) *rbody.APIResponse {
	r := new(rbody.APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	r.JSONResponse = string(rb)
	return r
}

func getStreamingAPIResponse(resp *http.Response) *rbody.APIResponse {
	r := new(rbody.APIResponse)
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

func watchTask(id string, port int) *watchTaskResult {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks/%s/watch", port, id))
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
					log.Fatal(err)
					r.close()
					return
				}
				switch ste.EventType {
				case rbody.TaskWatchTaskDisabled:
					r.eventChan <- ste.EventType
					r.close()
					return
				case rbody.TaskWatchTaskStopped, rbody.TaskWatchTaskStarted, rbody.TaskWatchMetricEvent:
					log.Info(ste.EventType)
					r.eventChan <- ste.EventType
				}
			}
		}
	}()
	return r
}

func getTasks(port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getTask(id string, port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/tasks/%s", port, id))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func startTask(id string, port int) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%s/start", port, id)
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

func stopTask(id string, port int) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%s/stop", port, id)
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

func removeTask(id string, port int) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%s", port, id)
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

func createTask(sample, name, interval string, noStart bool, port int) *rbody.APIResponse {
	jsonP, err := ioutil.ReadFile("./wmap_sample/" + sample)
	if err != nil {
		log.Fatal(err)
	}
	wf, err := wmap.FromJson(jsonP)
	if err != nil {
		log.Fatal(err)
	}

	uri := fmt.Sprintf("http://localhost:%d/v1/tasks", port)

	t := core.TaskCreationRequest{
		Schedule: &core.Schedule{Type: "simple", Interval: interval},
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

func enableTask(id string, port int) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/tasks/%s/enable", port, id)
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

func uploadPlugin(pluginPath string, port int) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins", port)

	client := &http.Client{}
	file, err := os.Open(pluginPath)
	if err != nil {
		log.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	var part io.Writer
	part, err = writer.CreateFormFile("snap-plugins", filepath.Base(pluginPath))
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

func unloadPlugin(port int, pluginType string, name string, version int) *rbody.APIResponse {
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

func getPluginList(port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/plugins", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getMetricCatalog(port int) *rbody.APIResponse {
	return fetchMetrics(port, "")
}

func fetchMetrics(port int, ns string) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/metrics%s", port, ns))
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

func fetchMetricsWithVersion(port int, ns string, ver int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/metrics%s?ver=%d", port, ns, ver))
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

func getPluginConfigItem(port int, typ *core.PluginType, name, ver string) *rbody.APIResponse {
	var uri string
	if typ != nil {
		uri = fmt.Sprintf("http://localhost:%d/v1/plugins/%d/%s/%s/config", port, *typ, name, ver)
	} else {
		uri = fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%s/config", port, "", name, ver)
	}
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}

	return getAPIResponse(resp)
}

func setPluginConfigItem(port int, typ string, name, ver string, cdn *cdata.ConfigDataNode) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%s/config", port, typ, name, ver)

	client := &http.Client{}
	b, err := json.Marshal(cdn)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("PUT", uri, bytes.NewReader(b))
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

func deletePluginConfigItem(port int, typ string, name, ver string, fields []string) *rbody.APIResponse {
	uri := fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%s/config", port, typ, name, ver)

	client := &http.Client{}
	b, err := json.Marshal(fields)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("DELETE", uri, bytes.NewReader(b))
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

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI(cfg *mockConfig) *restAPIInstance {
	// Start a REST API to talk to
	log.SetLevel(LOG_LEVEL)
	r, _ := New(cfg.RestAPI)
	c := control.New(cfg.Control)
	c.Start()
	s := scheduler.New(cfg.Scheduler)
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.BindConfigManager(c.Config)
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
	time.Sleep(time.Millisecond * 100)
	return &restAPIInstance{
		port:   r.Port(),
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
				r := startAPI(getDefaultMockConfig())
				port := r.port
				col := core.CollectorPluginType
				pub := core.PublisherPluginType
				Convey("A global plugin config is added for all plugins", func() {
					cdn := cdata.NewNode()
					cdn.AddItem("password", ctypes.ConfigValueStr{Value: "p@ssw0rd"})
					r := setPluginConfigItem(port, "", "", "", cdn)
					So(r.Body, ShouldHaveSameTypeAs, &rbody.SetPluginConfigItem{})
					r1 := r.Body.(*rbody.SetPluginConfigItem)
					So(r1.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})

					r2 := getPluginConfigItem(port, &col, "", "")
					So(r2.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
					r3 := r2.Body.(*rbody.PluginConfigItem)
					So(len(r3.Table()), ShouldEqual, 1)
					So(r3.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})

					Convey("A plugin config is added for all publishers", func() {
						cdn := cdata.NewNode()
						cdn.AddItem("user", ctypes.ConfigValueStr{Value: "john"})
						r := setPluginConfigItem(port, core.PublisherPluginType.String(), "", "", cdn)
						So(r.Body, ShouldHaveSameTypeAs, &rbody.SetPluginConfigItem{})
						r1 := r.Body.(*rbody.SetPluginConfigItem)
						So(r1.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
						So(len(r1.Table()), ShouldEqual, 2)

						Convey("A plugin config is added for all versions of a publisher", func() {
							cdn := cdata.NewNode()
							cdn.AddItem("path", ctypes.ConfigValueStr{Value: "/usr/local/influxdb/bin"})
							r := setPluginConfigItem(port, "2", "influxdb", "", cdn)
							So(r.Body, ShouldHaveSameTypeAs, &rbody.SetPluginConfigItem{})
							r1 := r.Body.(*rbody.SetPluginConfigItem)
							So(r1.Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/influxdb/bin"})
							So(len(r1.Table()), ShouldEqual, 3)

							Convey("A plugin config is added for a specific version of a publisher", func() {
								cdn := cdata.NewNode()
								cdn.AddItem("rate", ctypes.ConfigValueFloat{Value: .8})
								r := setPluginConfigItem(port, core.PublisherPluginType.String(), "influxdb", "1", cdn)
								So(r.Body, ShouldHaveSameTypeAs, &rbody.SetPluginConfigItem{})
								r1 := r.Body.(*rbody.SetPluginConfigItem)
								So(r1.Table()["rate"], ShouldResemble, ctypes.ConfigValueFloat{Value: .8})
								So(len(r1.Table()), ShouldEqual, 4)

								r2 := getPluginConfigItem(port, &pub, "", "")
								So(r2.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
								r3 := r2.Body.(*rbody.PluginConfigItem)
								So(len(r3.Table()), ShouldEqual, 2)

								r4 := getPluginConfigItem(port, &pub, "influxdb", "1")
								So(r4.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
								r5 := r4.Body.(*rbody.PluginConfigItem)
								So(len(r5.Table()), ShouldEqual, 4)

								Convey("A global plugin config field is deleted", func() {
									r := deletePluginConfigItem(port, "", "", "", []string{"password"})
									So(r.Body, ShouldHaveSameTypeAs, &rbody.DeletePluginConfigItem{})
									r1 := r.Body.(*rbody.DeletePluginConfigItem)
									So(len(r1.Table()), ShouldEqual, 0)

									r2 := setPluginConfigItem(port, core.PublisherPluginType.String(), "influxdb", "", cdn)
									So(r2.Body, ShouldHaveSameTypeAs, &rbody.SetPluginConfigItem{})
									r3 := r2.Body.(*rbody.SetPluginConfigItem)
									So(len(r3.Table()), ShouldEqual, 3)
								})
							})
						})
					})
				})

			})
			Convey("Plugin config is set at startup", func() {
				cfg := getDefaultMockConfig()
				err := cfgfile.Read("../../examples/configs/snap-config-sample.json", &cfg, MOCK_CONSTRAINTS)
				So(err, ShouldBeNil)
				if len(SNAP_AUTODISCOVER_PATH) == 0 {
					if len(SNAP_PATH) != 0 {

						SNAP_AUTODISCOVER_PATH = helper.PluginPath()
						log.Warning(fmt.Sprintf("SNAP_AUTODISCOVER_PATH has been set to plugin build path (%s). This might cause test failures", SNAP_AUTODISCOVER_PATH))
					}
				} else {
					log.Warning(fmt.Sprintf("SNAP_AUTODISCOVER_PATH is set to %s. This might cause test failures", SNAP_AUTODISCOVER_PATH))
				}
				cfg.Control.AutoDiscoverPath = SNAP_AUTODISCOVER_PATH
				r := startAPI(cfg)
				port := r.port
				col := core.CollectorPluginType

				Convey("Gets the collector config by name and version", func() {
					r := getPluginConfigItem(port, &col, "pcm", "1")
					So(r.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
					r1 := r.Body.(*rbody.PluginConfigItem)
					So(r1.Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
					So(r1.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
					So(len(r1.Table()), ShouldEqual, 6)
				})
				Convey("Gets the config for a collector by name", func() {
					r := getPluginConfigItem(port, &col, "pcm", "")
					So(r.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
					r1 := r.Body.(*rbody.PluginConfigItem)
					So(r1.Table()["path"], ShouldResemble, ctypes.ConfigValueStr{Value: "/usr/local/pcm/bin"})
					So(r1.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
					So(len(r1.Table()), ShouldEqual, 3)
				})
				Convey("Gets the config for all collectors", func() {
					r := getPluginConfigItem(port, &col, "", "")
					So(r.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
					r1 := r.Body.(*rbody.PluginConfigItem)
					So(r1.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
					So(r1.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
					So(len(r1.Table()), ShouldEqual, 2)
				})
				Convey("Gets the config for all plugins", func() {
					r := getPluginConfigItem(port, nil, "", "")
					So(r.Body, ShouldHaveSameTypeAs, &rbody.PluginConfigItem{})
					r1 := r.Body.(*rbody.PluginConfigItem)
					So(r1.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
					So(len(r1.Table()), ShouldEqual, 1)
				})
			})
		})

		Convey("Enable task - put - /v1/tasks/:id/enable", func() {
			Convey("Enable a running task", func(c C) {
				r := startAPI(getDefaultMockConfig())
				port := r.port

				uploadPlugin(MOCK_PLUGIN_PATH2, port)
				uploadPlugin(FILE_PLUGIN_PATH, port)

				r1 := createTask("1.json", "yeti", "1s", true, port)
				So(r1.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
				plr1 := r1.Body.(*rbody.AddScheduledTask)

				id := plr1.ID

				r2 := startTask(id, port)
				So(r2.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStarted))
				plr2 := r2.Body.(*rbody.ScheduledTaskStarted)
				So(plr2.ID, ShouldEqual, id)

				r4 := enableTask(id, port)
				So(r4.Body, ShouldHaveSameTypeAs, new(rbody.Error))
				plr4 := r4.Body.(*rbody.Error)
				So(plr4.ErrorMessage, ShouldEqual, "Task must be disabled")
			})
		})
	})
}
