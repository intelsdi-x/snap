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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/fixtures"
	"github.com/intelsdi-x/snap/plugin/helper"
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

func startV2API(cfg *mockConfig, testType string) *restAPIInstance {
	log.SetLevel(LOG_LEVEL)
	r, _ := New(cfg.RestAPI)
	switch testType {
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
				fmt.Sprintf(fixtures.GET_METRICS_RESPONSE, r.port))
		})
	})
}

func TestV2Task(t *testing.T) {
	r := startV2API(getDefaultMockConfig(), "task")
	Convey("Test Task REST API V2", t, func() {
		Convey("Get tasks - v2/tasks", func() {
			resp, err := http.Get(
				fmt.Sprintf("http://localhost:%d/v2/tasks", r.port))
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

	})
}
