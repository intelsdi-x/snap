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
	"io/ioutil"
	"net/http"
	"os"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/plugin/helper"
	"github.com/intelsdi-x/snap/scheduler"
	log "github.com/sirupsen/logrus"
)

// common resources used for medium tests

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

// Since we do not have a global snap package that could be imported
// we create a mock config struct to mock what is in snapteld.go

type mockConfig struct {
	LogLevel   int    `json:"-"yaml:"-"`
	GoMaxProcs int    `json:"-"yaml:"-"`
	LogPath    string `json:"-"yaml:"-"`
	Control    *control.Config
	Scheduler  *scheduler.Config `json:"-",yaml:"-"`
	RestAPI    *Config           `json:"-",yaml:"-"`
}

func getDefaultMockConfig() *mockConfig {
	return &mockConfig{
		LogLevel:   3,
		GoMaxProcs: 1,
		LogPath:    "",
		Control:    control.GetDefaultConfig(),
		Scheduler:  scheduler.GetDefaultConfig(),
		RestAPI:    GetDefaultConfig(),
	}
}

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
