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

package v1

import (
	"testing"

	"github.com/intelsdi-x/snap/mgmt/rest/v1/fixtures"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetPlugins(t *testing.T) {
	mm := fixtures.MockManagesMetrics{}
	host := "localhost"
	Convey("Test getPlugns method", t, func() {
		Convey("Without details", func() {
			detail := false
			Convey("Get All plugins", func() {
				plName := ""
				plType := ""
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 0)
			})
		})
		Convey("With details", func() {
			detail := true
			Convey("Get All plugins", func() {
				plName := ""
				plType := ""
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 6)
				So(len(plugins.AvailablePlugins), ShouldEqual, 6)
			})
			Convey("Filter plugins by Type", func() {
				plName := ""
				plType := "publisher"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
			Convey("Filter plugins by Type and Name", func() {
				plName := "foo"
				plType := "processor"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 1)
				So(len(plugins.AvailablePlugins), ShouldEqual, 1)
			})
			Convey("Filter plugins by Type and Name expect duplicates", func() {
				plName := "foo"
				plType := "collector"
				plugins := getPlugins(mm, detail, host, plName, plType)
				So(len(plugins.LoadedPlugins), ShouldEqual, 2)
				So(len(plugins.AvailablePlugins), ShouldEqual, 2)
			})
		})
	})

}
