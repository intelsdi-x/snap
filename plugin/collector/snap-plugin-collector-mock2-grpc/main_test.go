// +build legacy

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

package main

import (
	"testing"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "snap-plugin-collector-mock2-grpc"
	PluginType = "collector"
	PluginPath = helper.PluginFilePath(PluginName)
)

func TestMockPluginLoad(t *testing.T) {
	// These tests only work if SNAP_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	Convey("make sure plugin has been built", t, func() {
		err := helper.PluginFileCheck(PluginName)
		So(err, ShouldBeNil)

		Convey("ensure plugin loads and responds", func() {
			c := control.New(control.GetDefaultConfig())
			c.Start()
			rp, err := core.NewRequestedPlugin(PluginPath, c.GetTempDir(), nil)
			Convey("Should not return an error when requested for a plugin", func() {
				So(err, ShouldBeNil)
			})

			_, err = c.Load(rp)
			Convey("should not return an error when loading a plugin", func() {
				So(err, ShouldBeNil)
			})
		})

	})
}
