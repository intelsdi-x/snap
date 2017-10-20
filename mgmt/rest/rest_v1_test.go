// +build medium

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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/fixtures"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/intelsdi-x/snap/pkg/cfgfile"
	"github.com/intelsdi-x/snap/plugin/helper"
	"github.com/intelsdi-x/snap/scheduler"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

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
				case rbody.TaskWatchTaskStopped, rbody.TaskWatchTaskEnded, rbody.TaskWatchTaskStarted, rbody.TaskWatchMetricEvent:
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

func startV1API(cfg *mockConfig, testType string) *restAPIInstance {
	log.SetLevel(LOG_LEVEL)
	r, _ := New(cfg.RestAPI)
	switch testType {
	case "tribe":
		mockTribeManager := &fixtures.MockTribeManager{}
		r.BindTribeManager(mockTribeManager)
	case "plugin":
		mockMetricManager := &fixtures.MockManagesMetrics{}
		mockConfigManager := &fixtures.MockConfigManager{}
		r.BindMetricManager(mockMetricManager)
		r.BindConfigManager(mockConfigManager)
	case "metric":
		mockMetricManager := &fixtures.MockManagesMetrics{}
		r.BindMetricManager(mockMetricManager)
	case "task":
		mockTaskManager := &fixtures.MockTaskManager{}
		r.BindTaskManager(mockTaskManager)
	}
	go func(ch <-chan error) {
		// Block on the error channel. Will return exit status 1 for an error or
		// just return if the channel closes.
		err, ok := <-ch
		if !ok {
			return
		}
		log.Fatal(err)
	}(r.Err())
	r.SetAddress("127.0.0.1:0")
	r.Start()
	return &restAPIInstance{
		port:   r.Port(),
		server: r,
	}
}

func TestV1Plugin(t *testing.T) {
	r := startV1API(getDefaultMockConfig(), "plugin")
	Convey("Test Plugin REST API V1", t, func() {
		Convey("Get plugins - v1/plugins", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/plugins", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_PLUGINS_RESPONSE, r.port, r.port,
					r.port, r.port, r.port, r.port),
				ShouldResemble,
				string(body))
		})
		Convey("Get plugins - v1/plugins/:type", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/plugins/collector", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_PLUGINS_RESPONSE_TYPE, r.port, r.port),
				ShouldResemble,
				string(body))
		})
		Convey("Get plugins - v1/plugins/:type:name", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/plugins/publisher/bar", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_PLUGINS_RESPONSE_TYPE_NAME, r.port),
				ShouldResemble,
				string(body))
		})
		Convey("Get plugins - v1/plugins/:type:name:version", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/plugins/publisher/bar/3", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_PLUGINS_RESPONSE_TYPE_NAME_VERSION, r.port),
				ShouldResemble,
				string(body))
		})

		Convey("Post plugins - v1/plugins/:type:name", func(c C) {
			f, err := os.Open(MOCK_PLUGIN_PATH1)
			defer f.Close()
			So(err, ShouldBeNil)

			// We create a pipe so that we can write the file in multipart
			// format and read it in to the body of the http request
			reader, writer := io.Pipe()
			mwriter := multipart.NewWriter(writer)
			bufin := bufio.NewReader(f)

			// A go routine is needed since we must write the multipart file
			// to the pipe so we can read from it in the http call
			go func() {
				part, err := mwriter.CreateFormFile("snap-plugins", "mock")
				c.So(err, ShouldBeNil)
				bufin.WriteTo(part)
				mwriter.Close()
				writer.Close()
			}()

			resp1, err1 := http.Post(
				fmt.Sprintf("http://localhost:%d/v1/plugins", r.port),
				mwriter.FormDataContentType(), reader)
			So(err1, ShouldBeNil)
			So(resp1.StatusCode, ShouldEqual, 201)
		})

		Convey("Delete plugins - v1/plugins/:type:name:version", func() {
			c := &http.Client{}
			pluginName := "foo"
			pluginType := "collector"
			pluginVersion := 2
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%d",
					r.port,
					pluginType,
					pluginName,
					pluginVersion),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.UNLOAD_PLUGIN_RESPONSE),
				ShouldResemble,
				string(body))
		})

		Convey("Get plugin config items - v1/plugins/:type/:name/:version/config", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/plugins/publisher/bar/3/config", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_PLUGIN_CONFIG_ITEM),
				ShouldResemble,
				string(body))
		})

		Convey("Set plugin config item- v1/plugins/:type/:name/:version/config", func() {
			c := &http.Client{}
			pluginName := "foo"
			pluginType := "collector"
			pluginVersion := 2
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Jane"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%d/config",
					r.port,
					pluginType,
					pluginName,
					pluginVersion),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.SET_PLUGIN_CONFIG_ITEM),
				ShouldResemble,
				string(body))

		})

		Convey("Delete plugin config item - /v1/plugins/:type/:name/:version/config", func() {
			c := &http.Client{}
			pluginName := "foo"
			pluginType := "collector"
			pluginVersion := 2
			cd := []string{"foo"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v1/plugins/%s/%s/%d/config",
					r.port,
					pluginType,
					pluginName,
					pluginVersion),
				bytes.NewReader(body))

			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.DELETE_PLUGIN_CONFIG_ITEM),
				ShouldResemble,
				string(body))

		})
	})
}

func TestV1Metric(t *testing.T) {
	r := startV1API(getDefaultMockConfig(), "metric")
	Convey("Test Metric REST API V1", t, func() {
		Convey("Get metrics - v1/metrics", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/metrics", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			resp1, err := url.QueryUnescape(string(body))
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_METRICS_RESPONSE, r.port),
				ShouldResemble,
				resp1)
		})

		Convey("Get metrics from tree - v1/metrics/*namespace", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/metrics/*namespace", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			resp1, err := url.QueryUnescape(string(body))
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_METRICS_RESPONSE, r.port),
				ShouldResemble,
				resp1)
		})
	})
}

func TestV1Task(t *testing.T) {
	r := startV1API(getDefaultMockConfig(), "task")
	Convey("Test Task REST API V1", t, func() {
		Convey("Get tasks - v1/tasks", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tasks", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			responses := []string{
				fmt.Sprintf(fixtures.GET_TASKS_RESPONSE, r.port, r.port),
				fmt.Sprintf(fixtures.GET_TASKS_RESPONSE2, r.port, r.port),
			}
			// GetTasks returns an unordered map,
			// thus there is more than one possible response
			So(
				string(body),
				ShouldBeIn,
				responses,
			)
		})

		Convey("Get task - v1/tasks/:id", func() {
			taskID := "1234"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tasks/:%s", r.port, taskID))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.GET_TASK_RESPONSE, r.port),
			)
		})

		Convey("Watch tasks - v1/tasks/:id/watch", func() {
			taskID := "1234"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tasks/:%s/watch", r.port, taskID))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
		})

		Convey("Add tasks - v1/tasks", func() {
			reader := strings.NewReader(fixtures.TASK)
			resp, err := http.Post(
				fmt.Sprintf("http://localhost:%d/v1/tasks", r.port),
				http.DetectContentType([]byte(fixtures.TASK)),
				reader)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeEmpty)
			So(resp.StatusCode, ShouldEqual, 201)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.ADD_TASK_RESPONSE, r.port),
			)
		})

		Convey("Start tasks - v1/tasks/:id/start", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v1/tasks/%s/start", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.START_TASK_RESPONSE_ID_START),
				ShouldResemble,
				string(body))
		})

		Convey("Stop tasks - v1/tasks/:id/stop", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v1/tasks/%s/stop", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.STOP_TASK_RESPONSE_ID_STOP),
			)
		})

		Convey("Enable tasks - v1/tasks/:id/enable", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v1/tasks/%s/enable", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.ENABLE_TASK_RESPONSE_ID_ENABLE),
			)
		})

		Convey("Remove tasks - V1/tasks/:id", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := []string{"foo"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v1/tasks/%s",
					r.port,
					taskID),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.REMOVE_TASK_RESPONSE_ID),
			)
		})
	})
}

func TestV1Tribe(t *testing.T) {
	r := startV1API(getDefaultMockConfig(), "tribe")
	Convey("Test Tribe REST API V1", t, func() {
		Convey("Get tribe agreements - v1/tribe/agreements", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tribe/agreements", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.GET_TRIBE_AGREEMENTS_RESPONSE),
			)
		})

		Convey("Add tribe agreements - /v1/tribe/agreements", func() {
			agreement := "{\"Name\": \"Agree2\"}"
			reader := strings.NewReader(agreement)
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/v1/tribe/agreements", r.port),
				http.DetectContentType([]byte(agreement)), reader)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeEmpty)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.ADD_TRIBE_AGREEMENT_RESPONSE),
			)
		})

		Convey("Get tribe agreements - v1/tribe/agreements/:name", func() {
			tribeName := "Agree1"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tribe/agreements/%s", r.port, tribeName))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.GET_TRIBE_AGREEMENTS_RESPONSE_NAME),
			)
		})

		Convey("Get tribe members - v1/tribe/members", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tribe/members", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.GET_TRIBE_MEMBERS_RESPONSE),
			)
		})

		Convey("Get tribe member - v1/tribe/member/:name", func() {
			tribeName := "Imma_Mock"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tribe/member/%s", r.port, tribeName))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.GET_TRIBE_MEMBER_NAME),
			)
		})

		Convey("Delete tribe agreement - v1/tribe/agreements/:name", func() {
			c := &http.Client{}
			tribeName := "Agree1"
			cd := []string{"foo"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v1/tribe/agreements/%s",
					r.port,
					tribeName),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.DELETE_TRIBE_AGREEMENT_RESPONSE_NAME),
			)
		})

		Convey("Leave tribe agreement - v1/tribe/agreements/:name/leave", func() {
			c := &http.Client{}
			tribeName := "Agree1"
			cd := map[string]string{"Apple": "a", "Ball": "b", "Cat": "c"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v1/tribe/agreements/%s/leave",
					r.port,
					tribeName),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.LEAVE_TRIBE_AGREEMENT_RESPONSE_NAME_LEAVE),
			)
		})

		Convey("Join tribe agreement - v1/tribe/agreements/:name/join", func() {
			c := &http.Client{}
			tribeName := "Agree1"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v1/tribe/agreements/%s/join", r.port, tribeName),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(fixtures.JOIN_TRIBE_AGREEMENT_RESPONSE_NAME_JOIN),
			)

		})
	})
}
