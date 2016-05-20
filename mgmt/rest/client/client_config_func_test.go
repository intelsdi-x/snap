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

package client

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

func TestSnapClientConfig(t *testing.T) {
	uri := startAPI()
	CompressUpload = false
	c, err := New(uri, "v1", true)
	Convey("Client should be created", t, func() {
		So(err, ShouldBeNil)
		Convey("When no config is set", func() {
			Convey("Get global config", func() {
				res := c.GetPluginConfig("", "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 0)
			})
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 0)
			})
			Convey("Get config for all processors", func() {
				res := c.GetPluginConfig(core.ProcessorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 0)
			})
			Convey("Get config for all publishers", func() {
				res := c.GetPluginConfig(core.PublisherPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 0)
			})
		})
		Convey("A global plugin config is set", func() {
			res := c.SetPluginConfig("", "", "", "password", ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			So(res.Err, ShouldBeNil)
			So(res.SetPluginConfigItem, ShouldNotBeNil)
			So(len(res.SetPluginConfigItem.Table()), ShouldEqual, 1)
			So(res.SetPluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			Convey("Get global config", func() {
				res := c.GetPluginConfig("", "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Get config for all processors", func() {
				res := c.GetPluginConfig(core.ProcessorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Get config for all publishers", func() {
				res := c.GetPluginConfig(core.PublisherPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
		})

		Convey("A plugin config item for a specific type of plugin is added", func() {
			res := c.SetPluginConfig(core.CollectorPluginType.String(), "", "", "user", ctypes.ConfigValueStr{Value: "john"})
			So(res.Err, ShouldBeNil)
			So(res.SetPluginConfigItem, ShouldNotBeNil)
			So(len(res.SetPluginConfigItem.Table()), ShouldEqual, 2)
			So(res.SetPluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
			So(res.SetPluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			Convey("Get global config", func() {
				res := c.GetPluginConfig("", "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
			})
			Convey("Get config for all processors", func() {
				res := c.GetPluginConfig(core.ProcessorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all publishers", func() {
				res := c.GetPluginConfig(core.PublisherPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
		})

		Convey("A plugin config item for a specific type and name of a plugin is added", func() {
			res := c.SetPluginConfig(core.CollectorPluginType.String(), "test", "", "foo", ctypes.ConfigValueStr{Value: "bar"})
			So(res.Err, ShouldBeNil)
			So(res.SetPluginConfigItem, ShouldNotBeNil)
			So(len(res.SetPluginConfigItem.Table()), ShouldEqual, 3)
			Convey("Get global config", func() {
				res := c.GetPluginConfig("", "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 2)
			})
			Convey("Get config for all processors", func() {
				res := c.GetPluginConfig(core.ProcessorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all publishers", func() {
				res := c.GetPluginConfig(core.PublisherPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for the 'test' collector", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "test", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				So(res.PluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 3)
			})
		})

		Convey("A plugin config item for a specific type, name and version of a plugin is added", func() {
			res := c.SetPluginConfig(core.CollectorPluginType.String(), "test", "1", "go", ctypes.ConfigValueStr{Value: "pher"})
			So(res.Err, ShouldBeNil)
			So(res.SetPluginConfigItem, ShouldNotBeNil)
			So(len(res.SetPluginConfigItem.Table()), ShouldEqual, 4)
			Convey("Get global config", func() {
				res := c.GetPluginConfig("", "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 2)
			})
			Convey("Get config for all processors", func() {
				res := c.GetPluginConfig(core.ProcessorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for all publishers", func() {
				res := c.GetPluginConfig(core.PublisherPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for the version 1 'test' collector plugin", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "test", "1")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(res.PluginConfigItem.Table()["password"], ShouldResemble, ctypes.ConfigValueStr{Value: "p@ssw0rd"})
				So(res.PluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
				So(res.PluginConfigItem.Table()["go"], ShouldResemble, ctypes.ConfigValueStr{Value: "pher"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 4)
			})
		})
		Convey("A global plugin config item is removed", func() {
			res := c.DeletePluginConfig("", "", "", "password")
			So(res.Err, ShouldBeNil)
			So(res.DeletePluginConfigItem, ShouldNotBeNil)
			So(len(res.DeletePluginConfigItem.Table()), ShouldEqual, 0)
			Convey("Get config for all collectors", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "", "")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
			Convey("Get config for the version 1 'test' collector plugin", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "test", "1")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["user"], ShouldResemble, ctypes.ConfigValueStr{Value: "john"})
				So(res.PluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
				So(res.PluginConfigItem.Table()["go"], ShouldResemble, ctypes.ConfigValueStr{Value: "pher"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 3)
			})
		})
		Convey("A plugin config item for a specific type of a plugin is removed", func() {
			res := c.DeletePluginConfig(core.CollectorPluginType.String(), "", "", "user")
			So(res.Err, ShouldBeNil)
			So(res.DeletePluginConfigItem, ShouldNotBeNil)
			So(len(res.DeletePluginConfigItem.Table()), ShouldEqual, 0)
			Convey("Get config for the version 1 'test' collector plugin", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "test", "1")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
				So(res.PluginConfigItem.Table()["go"], ShouldResemble, ctypes.ConfigValueStr{Value: "pher"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 2)
			})
		})
		Convey("A plugin config item for a specific type and version of a plugin is removed", func() {
			res := c.DeletePluginConfig(core.CollectorPluginType.String(), "test", "1", "go")
			So(res.Err, ShouldBeNil)
			So(res.DeletePluginConfigItem, ShouldNotBeNil)
			So(len(res.DeletePluginConfigItem.Table()), ShouldEqual, 1)
			So(res.DeletePluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
			Convey("Get config for the version 1 'test' collector plugin", func() {
				res := c.GetPluginConfig(core.CollectorPluginType.String(), "test", "1")
				So(res.Err, ShouldBeNil)
				So(res.PluginConfigItem, ShouldNotBeNil)
				So(res.PluginConfigItem.Table()["foo"], ShouldResemble, ctypes.ConfigValueStr{Value: "bar"})
				So(len(res.PluginConfigItem.Table()), ShouldEqual, 1)
			})
		})
	})
}
