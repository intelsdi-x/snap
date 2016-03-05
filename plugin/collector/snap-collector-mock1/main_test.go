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

package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "snap-collector-mock1"
	PluginType = "collector"
	SnapPath   = os.Getenv("SNAP_PATH")
	PluginPath = path.Join(SnapPath, "plugin", PluginName)
)

func TestMockPluginLoad(t *testing.T) {
	// These tests only work if SNAP_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if SnapPath != "" {
		// Helper plugin trigger build if possible for this plugin
		helper.BuildPlugin(PluginType, PluginName)
		//
		Convey("ensure plugin loads and responds", t, func() {
			c := control.New(control.GetDefaultConfig())
			c.Start()
			rp, _ := core.NewRequestedPlugin(PluginPath)
			_, err := c.Load(rp)

			So(err, ShouldBeNil)
		})
	} else {
		fmt.Printf("SNAP_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}

func TestMain(t *testing.T) {
	Convey("ensure plugin loads and responds", t, func() {
		os.Args = []string{"", "{\"NoDaemon\": true}"}
		So(func() { main() }, ShouldNotPanic)
	})
}
