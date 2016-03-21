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

package control

import (
	"crypto/sha256"
	"io/ioutil"
	"net"
	"testing"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestControlProxy(t *testing.T) {
	Convey("Starting control_proxy server/client", t, func() {
		l, _ := net.Listen("tcp", ":0")
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()
		control := New(ListenPort(port))
		err := control.Start()
		So(err, ShouldBeNil)
		connection, err := rpcutil.GetClientConnection(DefaultListenAddress, port, "", "")
		So(err, ShouldBeNil)
		client := rpc.NewMetricManagerClient(connection)
		So(client, ShouldNotBeNil)

		Convey("Loading a plugin", func() {
			//Loading a plugin should result in 201 - Created
			mock1Path := "../build/plugin/snap-collector-mock1"
			pluginBytes, _ := ioutil.ReadFile(mock1Path)
			name := "snap-collector-mock1"
			checkSum := sha256.Sum256(pluginBytes)
			var signature []byte
			arg := &rpc.PluginRequest{
				Name:       name,
				CheckSum:   checkSum[:],
				Signature:  signature,
				PluginFile: pluginBytes,
			}
			reply, err := client.Load(context.Background(), arg)
			So(err, ShouldBeNil)
			plugin := rpc.ReplyToLoadedPlugin(reply)
			So(plugin.Name, ShouldEqual, "mock")
			So(plugin.Version, ShouldEqual, 1)
			So(plugin.TypeName, ShouldEqual, "collector")
			So(plugin.IsSigned, ShouldEqual, false)
			Convey("Listing all loaded plugins", func() {
				r, err := client.PluginCatalog(context.Background(), &rpc.EmptyRequest{})
				So(err, ShouldBeNil)
				Convey("Should contain the plugin that is loaded", func() {
					loadedPlugin := rpc.ReplyToLoadedPlugin(r.Plugins[0])
					So(loadedPlugin.Name, ShouldEqual, "mock")
					So(loadedPlugin.Version, ShouldEqual, 1)
					So(loadedPlugin.TypeName, ShouldEqual, "collector")
					So(loadedPlugin.IsSigned, ShouldEqual, false)
				})
			})
			Convey("Getting a specific plugin", func() {
				Convey("Should return that plugin if it exists", func() {
					arg := &rpc.GetPluginRequest{
						Name:     "mock",
						Version:  1,
						Type:     "collector",
						Download: false,
					}
					r, err := client.GetPlugin(context.Background(), arg)
					So(err, ShouldBeNil)
					lp := rpc.ReplyToLoadedPlugin(r.Plugin)
					So(lp.Name, ShouldEqual, "mock")
					So(lp.Version, ShouldEqual, 1)
					So(lp.TypeName, ShouldEqual, "collector")
				})
				Convey("Should error if it doesn't exist", func() {
					arg := &rpc.GetPluginRequest{
						Name:     "titanium",
						Version:  1,
						Type:     "collector",
						Download: false,
					}
					r, err := client.GetPlugin(context.Background(), arg)
					So(err, ShouldNotBeNil)
					So(r, ShouldBeNil)
				})
				Convey("Should return plugin bytes if download set to true", func() {
					arg := &rpc.GetPluginRequest{
						Name:     "mock",
						Version:  1,
						Type:     "collector",
						Download: true,
					}
					r, err := client.GetPlugin(context.Background(), arg)
					So(err, ShouldBeNil)
					So(r.PluginBytes, ShouldNotBeNil)
				})
			})
			Convey("Getting Metric catalog", func() {
				r, err := client.MetricCatalog(context.Background(), &rpc.EmptyRequest{})
				So(err, ShouldBeNil)
				metrics := rpc.ReplyToMetrics(r.Metrics)
				Convey("Should show metrics", func() {
					So(len(metrics), ShouldEqual, 3)
					// Verify the metrics returned our what we expect
					var intelmockfoo rpc.Metric
					for _, metric := range metrics {
						if metric.Namespace[2] == "foo" {
							intelmockfoo = metric
						}
					}
					So(intelmockfoo.Version, ShouldEqual, 1)
					So(intelmockfoo.Namespace, ShouldResemble, []string{"intel", "mock", "foo"})
				})
			})
			Convey("Fetching metrics with supplied namespace", func() {
				Convey("Should return metrics starting with that namespace", func() {
					// We expect this to return these metrics:
					// '/intel/mock/foo'
					// '/intel/mock/bar'
					// 'intel/mock/*/baz'
					arg := &rpc.FetchMetricsRequest{
						Namespace: []string{"intel", "mock"},
						Version:   1,
					}
					r, err := client.FetchMetrics(context.Background(), arg)
					So(err, ShouldBeNil)
					metrics := rpc.ReplyToMetrics(r.Metrics)
					So(len(metrics), ShouldEqual, 3)
					intelmockfoo := []string{"intel", "mock", "foo"}
					intelmockbar := []string{"intel", "mock", "bar"}
					intelmockstarbaz := []string{"intel", "mock", "*", "baz"}
					possibleResults := [][]string{intelmockfoo, intelmockbar, intelmockstarbaz}
					found := 0
					// Because order is not guaranteed we just make sure we have all the
					// metrics we are expecting.
					for _, metric := range metrics {
						for _, result := range possibleResults {
							if equal(metric.Namespace, result) {
								found++
								break
							}
						}
					}
					So(found, ShouldEqual, 3)
				})
				Convey("Should fail if no metrics exist at that namespace", func() {
					// We expect a 404 if metric is not found
					arg := &rpc.FetchMetricsRequest{
						Namespace: []string{"titanium", "poodle"},
						Version:   1,
					}
					r, err := client.FetchMetrics(context.Background(), arg)
					So(err, ShouldNotBeNil)
					So(r, ShouldBeNil)
				})
			})
			Convey("Geting a single metric", func() {
				Convey("Should return that metric", func() {
					arg := &rpc.FetchMetricsRequest{
						Namespace: []string{"intel", "mock", "foo"},
						Version:   1,
					}
					r, err := client.GetMetric(context.Background(), arg)
					So(err, ShouldBeNil)
					metric := rpc.ReplyToMetric(r)
					So(metric.Namespace, ShouldResemble, []string{"intel", "mock", "foo"})
				})
				Convey("Should fail if metric doesn't exist", func() {
					arg := &rpc.FetchMetricsRequest{
						Namespace: []string{"titanium", "poodle"},
						Version:   1,
					}
					r, err := client.GetMetric(context.Background(), arg)
					So(err, ShouldNotBeNil)
					So(r, ShouldBeNil)
				})
			})
			Convey("Getting all versions of a given metric", func() {
				//load mock2 for multiple versions of metric
				//Loading a plugin should result in 201 - Created
				mock1Path := "../build/plugin/snap-collector-mock2"
				pluginBytes, _ := ioutil.ReadFile(mock1Path)
				name := "snap-collector-mock1"
				checkSum := sha256.Sum256(pluginBytes)
				var signature []byte
				arg := &rpc.PluginRequest{
					Name:       name,
					CheckSum:   checkSum[:],
					Signature:  signature,
					PluginFile: pluginBytes,
				}
				_, err := client.Load(context.Background(), arg)
				So(err, ShouldBeNil)
				Convey("Should return all versions of that metric", func() {
					arg := &rpc.GetMetricVersionsRequest{
						Namespace: []string{"intel", "mock", "foo"},
					}
					r, err := client.GetMetricVersions(context.Background(), arg)
					So(err, ShouldBeNil)
					metrics := rpc.ReplyToMetrics(r.Metrics)
					So(len(metrics), ShouldEqual, 2)
				})
				Convey("Should fail if metric doesn't exist", func() {
					arg := &rpc.GetMetricVersionsRequest{
						Namespace: []string{"titanium", "poodle"},
					}
					r, err := client.GetMetricVersions(context.Background(), arg)
					So(err, ShouldNotBeNil)
					So(r, ShouldBeNil)
				})
			})
			Convey("Listing Available plugins", func() {
				Convey("Should return no plugins when none are available", func() {
					arg := &rpc.EmptyRequest{}
					r, err := client.AvailablePlugins(context.Background(), arg)
					So(err, ShouldBeNil)
					So(len(r.Plugins), ShouldEqual, 0)
				})
			})
			Convey("Unloading a plugin", func() {
				Convey("Should be successful if plugin is loaded", func() {
					arg := &rpc.UnloadPluginRequest{
						Name:       "mock",
						Version:    1,
						PluginType: "collector",
					}
					r, err := client.Unload(context.Background(), arg)
					So(err, ShouldBeNil)
					So(r.Name, ShouldEqual, "mock")
					So(r.Version, ShouldEqual, 1)
					So(r.TypeName, ShouldEqual, "collector")
					Convey("Should fail for unloaded plugin", func() {
						r, err := client.Unload(context.Background(), arg)
						So(err, ShouldNotBeNil)
						So(r, ShouldBeNil)
					})
				})
			})
		})
	})
}

func equal(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}
