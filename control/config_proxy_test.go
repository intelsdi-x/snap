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

package control

import (
	"encoding/json"
	"net"
	"testing"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/internal/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigProxy(t *testing.T) {
	Convey("Starting config_proxy server/client", t, func() {
		l, _ := net.Listen("tcp", ":0")
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()
		cfg := GetDefaultConfig()
		cfg.ListenPort = port

		control := New(cfg)
		err := control.Start()
		So(err, ShouldBeNil)

		configClient, err := NewConfigClient(cfg.ListenAddr, cfg.ListenPort)
		So(err, ShouldBeNil)
		So(configClient, ShouldNotBeNil)
		Convey("Getting a PluginConfig", func() {
			// Load a plugin so we have something to manage configs on
			rp, err := core.NewRequestedPlugin(JSONRPCPluginPath)
			So(err, ShouldBeNil)
			_, err = control.Load(rp)
			So(err, ShouldBeNil)
			control.Config.Plugins.All.AddItem("password", ctypes.ConfigValueStr{Value: "testval1"})
			arg2 := &rpc.ConfigDataNodeRequest{
				PluginType: 0,
				Name:       "snap-test-plugin",
				Version:    1,
			}
			cfg, err := configClient.GetPluginConfigDataNode(context.Background(), arg2)
			So(err, ShouldBeNil)
			c := cdata.NewNode()
			c.UnmarshalJSON(cfg.Node)
			val, ok := c.Table()["password"]
			So(ok, ShouldEqual, true)
			So(val.(ctypes.ConfigValueStr).Value, ShouldEqual, "testval1")

			Convey("Getting All plugin Configs", func() {
				cfg, err := configClient.GetPluginConfigDataNodeAll(context.Background(), &common.Empty{})
				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
				c := cdata.NewNode()
				err = c.UnmarshalJSON(cfg.Node)
				So(err, ShouldBeNil)
				val, ok := c.Table()["password"]
				So(ok, ShouldEqual, true)
				So(val.(ctypes.ConfigValueStr).Value, ShouldEqual, "testval1")
			})
		})
		Convey("Adding a config to a specific plugin", func() {
			node := cdata.NewNode()
			node.AddItem("test", ctypes.ConfigValueStr{Value: "this is a test"})
			nbytes, err := json.Marshal(node)
			So(err, ShouldBeNil)

			request := &rpc.MergeConfigDataNodeRequest{
				DataNode: &rpc.ConfigDataNode{
					Node: nbytes,
				},
				Request: &rpc.ConfigDataNodeRequest{
					PluginType: 0,
					Name:       "snap-test-plugin",
					Version:    1,
				},
			}

			cfg, err := configClient.MergePluginConfigDataNode(context.Background(), request)
			So(err, ShouldBeNil)
			So(cfg, ShouldNotBeNil)
			// Get the item we just set
			config := control.Config.GetPluginConfigDataNode(0, "snap-test-plugin", 1)
			val, ok := config.Table()["test"]
			So(ok, ShouldEqual, true)
			So(val.(ctypes.ConfigValueStr).Value, ShouldEqual, "this is a test")
			Convey("Removing Config Item", func() {
				req := &rpc.DeleteConfigDataNodeFieldRequest{
					Request: &rpc.ConfigDataNodeRequest{
						PluginType: 0,
						Name:       "snap-test-plugin",
						Version:    1,
					},
					Fields: []string{"test"},
				}
				cfg, err := configClient.DeletePluginConfigDataNodeField(context.Background(), req)
				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
				config := control.Config.GetPluginConfigDataNode(core.PluginType(0), "snap-test-plugin", 1)
				_, ok := config.Table()["test"]
				So(ok, ShouldEqual, false)
			})
		})
		Convey("Adding a config to all plugins", func() {
			node := cdata.NewNode()
			node.AddItem("test", ctypes.ConfigValueStr{Value: "this is a test"})
			nbytes, err := json.Marshal(node)
			So(err, ShouldBeNil)
			req := &rpc.ConfigDataNode{
				Node: nbytes,
			}
			_, err = configClient.MergePluginConfigDataNodeAll(context.Background(), req)
			So(err, ShouldBeNil)
			config := control.Config.GetPluginConfigDataNodeAll()
			So(config, ShouldNotBeNil)
			val, ok := config.Table()["test"]
			So(ok, ShouldEqual, true)
			So(val.(ctypes.ConfigValueStr).Value, ShouldEqual, "this is a test")
			Convey("Removing a config from all plugins", func() {
				req := &rpc.DeleteConfigDataNodeFieldAllRequest{
					Fields: []string{"test"},
				}
				cfg, err := configClient.DeletePluginConfigDataNodeFieldAll(context.Background(), req)
				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
				config := control.Config.GetPluginConfigDataNodeAll()
				So(config, ShouldNotBeNil)
				_, ok := config.Table()["test"]
				So(ok, ShouldEqual, false)
			})
		})

	})

}
