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

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func startV2API(cfg *mockConfig, testType string) *restAPIInstance {
	log.SetLevel(LOG_LEVEL)
	r, _ := New(cfg.RestAPI)
	switch testType {
	case "plugin":
		mockMetricManager := &mock.MockManagesMetrics{}
		mockConfigManager := &mock.MockConfigManager{}
		r.BindMetricManager(mockMetricManager)
		r.BindConfigManager(mockConfigManager)
	case "metric":
		mockMetricManager := &mock.MockManagesMetrics{}
		r.BindMetricManager(mockMetricManager)
	case "task":
		mockTaskManager := &mock.MockTaskManager{}
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

func TestV2Plugin(t *testing.T) {
	r := startV2API(getDefaultMockConfig(), "plugin")
	Convey("Test Plugin REST API V2", t, func() {

		Convey("Post plugins - v2/plugins/:type:name", func(c C) {
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
				fmt.Sprintf("http://localhost:%d/v2/plugins", r.port),
				mwriter.FormDataContentType(), reader)
			So(err1, ShouldBeNil)
			So(resp1.StatusCode, ShouldEqual, 201)
		})

		Convey("Get plugins - v2/plugins", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/plugins", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_PLUGINS_RESPONSE, r.port, r.port,
					r.port, r.port, r.port, r.port))
		})
		Convey("Get plugins - v2/plugins/:type", func() {
			c := &http.Client{}
			req, err := http.NewRequest("GET",
				fmt.Sprintf("http://localhost:%d/v2/plugins", r.port),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			q := req.URL.Query()
			q.Add("type", "collector")
			req.URL.RawQuery = q.Encode()
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_PLUGINS_RESPONSE_TYPE, r.port, r.port))
		})
		Convey("Get plugins - v2/plugins/:type:name", func() {
			c := &http.Client{}
			req, err := http.NewRequest("GET",
				fmt.Sprintf("http://localhost:%d/v2/plugins", r.port),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			q := req.URL.Query()
			q.Add("type", "publisher")
			q.Add("name", "bar")
			req.URL.RawQuery = q.Encode()
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_PLUGINS_RESPONSE_TYPE_NAME, r.port))
		})
		Convey("Get plugin - v2/plugins/:type:name:version", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/plugins/publisher/bar/3", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_PLUGINS_RESPONSE_TYPE_NAME_VERSION, r.port))
		})

		Convey("Delete plugins - v2/plugins/:type:name:version", func() {
			c := &http.Client{}
			pluginName := "foo"
			pluginType := "collector"
			pluginVersion := 2
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v2/plugins/%s/%s/%d",
					r.port,
					pluginType,
					pluginName,
					pluginVersion),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.UNLOAD_PLUGIN_RESPONSE))
		})

		Convey("Get plugin config items - v2/plugins/:type/:name/:version/config", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/plugins/publisher/bar/3/config", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_PLUGIN_CONFIG_ITEM))
		})

		Convey("Set plugin config item- v2/plugins/:type/:name/:version/config", func() {
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
				fmt.Sprintf("http://localhost:%d/v2/plugins/%s/%s/%d/config",
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
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.SET_PLUGIN_CONFIG_ITEM))

		})

		Convey("Delete plugin config item - /v2/plugins/:type/:name/:version/config", func() {
			c := &http.Client{}
			pluginName := "foo"
			pluginType := "collector"
			pluginVersion := 2
			cd := []string{"foo"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v2/plugins/%s/%s/%d/config",
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
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.DELETE_PLUGIN_CONFIG_ITEM))
		})
	})
}

func TestV2Task(t *testing.T) {
	r := startV2API(getDefaultMockConfig(), "task")
	Convey("Test Task REST API V2", t, func() {

		Convey("Add tasks - v2/tasks", func() {
			reader := strings.NewReader(mock.TASK)
			resp, err := http.Post(
				fmt.Sprintf("http://localhost:%d/v2/tasks", r.port),
				http.DetectContentType([]byte(mock.TASK)),
				reader)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeEmpty)
			So(resp.StatusCode, ShouldEqual, 201)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.ADD_TASK_RESPONSE, r.port),
			)
		})

		Convey("Get tasks - v2/tasks", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/tasks", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			responses := []string{
				fmt.Sprintf(mock.GET_TASKS_RESPONSE, r.port, r.port),
				fmt.Sprintf(mock.GET_TASKS_RESPONSE2, r.port, r.port),
			}
			// GetTasks returns an unordered map,
			// thus there is more than one possible response
			So(
				string(body),
				ShouldBeIn,
				responses,
			)
		})

		Convey("Get task - v2/tasks/:id", func() {
			taskID := "1234"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/tasks/:%s", r.port, taskID))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.GET_TASK_RESPONSE, r.port),
			)
		})

		Convey("Watch tasks - v2/tasks/:id/watch", func() {
			taskID := "1234"
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/tasks/:%s/watch", r.port, taskID))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Start tasks - v2/tasks/:id", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v2/tasks/%s", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			q := req.URL.Query()
			q.Add("action", "start")
			req.URL.RawQuery = q.Encode()
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.START_TASK_RESPONSE_ID_START))
		})

		Convey("Stop tasks - v2/tasks/:id", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v2/tasks/%s", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			q := req.URL.Query()
			q.Add("action", "stop")
			req.URL.RawQuery = q.Encode()
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.STOP_TASK_RESPONSE_ID_STOP))
		})

		Convey("Enable tasks - v2/tasks/:id", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := cdata.NewNode()
			cd.AddItem("user", ctypes.ConfigValueStr{Value: "Kelly"})
			body, err := cd.MarshalJSON()
			So(err, ShouldBeNil)

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("http://localhost:%d/v2/tasks/%s", r.port, taskID),
				bytes.NewReader(body))
			So(err, ShouldBeNil)
			q := req.URL.Query()
			q.Add("action", "enable")
			req.URL.RawQuery = q.Encode()
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.ENABLE_TASK_RESPONSE_ID_ENABLE))
		})

		Convey("Remove tasks - v2/tasks/:id", func() {
			c := &http.Client{}
			taskID := "MockTask1234"
			cd := []string{"foo"}
			body, err := json.Marshal(cd)
			So(err, ShouldBeNil)
			req, err := http.NewRequest(
				"DELETE",
				fmt.Sprintf("http://localhost:%d/v2/tasks/%s",
					r.port,
					taskID),
				bytes.NewReader([]byte{}))
			So(err, ShouldBeNil)
			resp, err := c.Do(req)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			body, err = ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				string(body),
				ShouldResemble,
				fmt.Sprintf(mock.REMOVE_TASK_RESPONSE_ID))
		})
	})
}

func TestV2Metric(t *testing.T) {
	r := startV2API(getDefaultMockConfig(), "metric")
	Convey("Test Metric REST API V2", t, func() {

		Convey("Get metrics - v2/metrics", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/metrics", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			resp1, err := url.QueryUnescape(string(body))
			So(err, ShouldBeNil)
			So(
				resp1,
				ShouldResemble,
				fmt.Sprintf(mock.GET_METRICS_RESPONSE, r.port))
		})
	})
}
