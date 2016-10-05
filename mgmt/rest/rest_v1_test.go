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

	. "github.com/smartystreets/goconvey/convey"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/mgmt/rest/fixtures"
	"github.com/intelsdi-x/snap/plugin/helper"
)

var (
	LOG_LEVEL         = log.WarnLevel
	MOCK_PLUGIN_PATH1 = helper.PluginFilePath("snap-plugin-collector-mock1")
)

type restAPIInstance struct {
	port   int
	server *Server
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
				responses,
				ShouldContain,
				string(body))
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
				fmt.Sprintf(fixtures.GET_TASK_RESPONSE, r.port),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.ADD_TASK_RESPONSE, r.port),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.STOP_TASK_RESPONSE_ID_STOP),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.ENABLE_TASK_RESPONSE_ID_ENABLE),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.REMOVE_TASK_RESPONSE_ID),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.GET_TRIBE_AGREEMENTS_RESPONSE),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.ADD_TRIBE_AGREEMENT_RESPONSE),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.GET_TRIBE_AGREEMENTS_RESPONSE_NAME),
				ShouldResemble,
				string(body))
		})

		Convey("Get tribe members - v1/tribe/members", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v1/tribe/members", r.port))
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			body, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(
				fmt.Sprintf(fixtures.GET_TRIBE_MEMBERS_RESPONSE),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.GET_TRIBE_MEMBER_NAME),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.DELETE_TRIBE_AGREEMENT_RESPONSE_NAME),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.LEAVE_TRIBE_AGREEMENT_REPSONSE_NAME_LEAVE),
				ShouldResemble,
				string(body))
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
				fmt.Sprintf(fixtures.JOIN_TRIBE_AGREEMENT_RESPONSE_NAME_JOIN),
				ShouldResemble,
				string(body))

		})
	})
}
