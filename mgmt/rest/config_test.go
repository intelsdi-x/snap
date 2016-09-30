// +build legacy small medium large

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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
	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/scheduler"
)

// Since we do not have a global snap package that could be imported
// we create a mock config struct to mock what is in snapd.go

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
