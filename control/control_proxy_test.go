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
	"io/ioutil"
	"net"
	"testing"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/internal/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestControlProxy(t *testing.T) {
	l, _ := net.Listen("tcp", ":0")
	l.Close()
	cfg := GetDefaultConfig()
	cfg.ListenPort = l.Addr().(*net.TCPAddr).Port
	c := New(cfg)
	err := c.Start()
	Convey("Starting control_proxy server/client", t, func() {
		Convey("So err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
	client, err := NewClient(c.Config.ListenAddr, c.Config.ListenPort)
	Convey("Creating an RPC client to control RPC server", t, func() {
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		Convey("And a client should exist", func() {
			So(client, ShouldNotBeNil)
		})
	})

	rp, err := core.NewRequestedPlugin(JSONRPCPluginPath)
	if err != nil {
		log.Fatal(err)
	}
	pluginBytes, err := ioutil.ReadFile(rp.Path())
	if err != nil {
		log.Fatal(err)
	}
	checkSum := rp.CheckSum()
	reply, err := client.Load(
		context.Background(),
		&rpc.PluginRequest{
			Name:       "snap-collector-mock1",
			CheckSum:   checkSum[:],
			Signature:  rp.Signature(),
			PluginFile: pluginBytes,
		},
	)
	Convey("Loading a plugin", t, func() {
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		plugin, _ := rpc.ReplyToLoadedPlugin(reply)
		Convey("Plugin name should be mock", func() {
			So(plugin.Name, ShouldEqual, "mock")
		})
		Convey("Plugin version should be 1", func() {
			So(plugin.Version, ShouldEqual, 1)
		})
		Convey("Plugin type should be collector", func() {
			So(plugin.TypeName, ShouldEqual, "collector")
		})
		Convey("Plugin should not be signed", func() {
			So(plugin.IsSigned, ShouldEqual, false)
		})
	})

	reply2, err := client.PluginCatalog(context.Background(), &common.Empty{})
	Convey("Listing all loaded plugins", t, func() {
		Convey("Should not result in an error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Plugins in catalog should equal 1", func() {
			So(len(reply2.Plugins), ShouldEqual, 1)
		})
		Convey("Should contain the plugin that is loaded", func() {
			loadedPlugin, _ := rpc.ReplyToLoadedPlugin(reply2.Plugins[0])
			So(loadedPlugin.Name, ShouldEqual, "mock")
			So(loadedPlugin.Version, ShouldEqual, 1)
			So(loadedPlugin.TypeName, ShouldEqual, "collector")
			So(loadedPlugin.IsSigned, ShouldEqual, false)
		})
	})

	reply3, err := client.GetPlugin(
		context.Background(),
		&rpc.GetPluginRequest{
			Name:     "mock",
			Version:  1,
			Type:     "collector",
			Download: false,
		},
	)
	Convey("Getting the plugin loaded", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		Convey("And details should match the plugin loaded", func() {
			lp, _ := rpc.ReplyToLoadedPlugin(reply3.Plugin)
			So(lp.Name, ShouldEqual, "mock")
			So(lp.Version, ShouldEqual, 1)
			So(lp.TypeName, ShouldEqual, "collector")
		})
	})

	reply4, err := client.GetPlugin(
		context.Background(),
		&rpc.GetPluginRequest{
			Name:     "titanium",
			Version:  1,
			Type:     "collector",
			Download: false,
		},
	)
	Convey("Requesting a plugin that doesn't exist", t, func() {
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("And reply should be nil", func() {
			So(reply4, ShouldBeNil)
		})
	})

	reply5, err := client.GetPlugin(
		context.Background(),
		&rpc.GetPluginRequest{
			Name:     "mock",
			Version:  1,
			Type:     "collector",
			Download: true,
		},
	)
	Convey("Requesting to download a plugin that exists", t, func() {
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		Convey("And the reply PluginBytes should not be nil", func() {
			So(reply5.PluginBytes, ShouldNotBeNil)
		})
	})

	reply6, err := client.MetricCatalog(context.Background(), &common.Empty{})
	Convey("Getting Metric catalog", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		metrics, err := rpc.ReplyToMetrics(reply6.Metrics)
		Convey("Reply should return three metrics", func() {
			So(len(metrics), ShouldEqual, 3)
		})
		Convey("Err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})

	// We expect this to return these metrics:
	// '/intel/mock/foo'
	// '/intel/mock/bar'
	// 'intel/mock/*/baz'
	reply7, err := client.FetchMetrics(
		context.Background(),
		&rpc.FetchMetricsRequest{
			Namespace: []string{"intel", "mock"},
			Version:   1,
		},
	)
	Convey("Fetching metrics with supplied namespace", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		metrics, err := rpc.ReplyToMetrics(reply7.Metrics)
		Convey("Should return 3 metrics starting with that namespace", func() {
			So(len(metrics), ShouldEqual, 3)
		})
		Convey("Err should be nil", func() {
			So(err, ShouldBeNil)
		})
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
		Convey("Verify our 3 metrics returned match what is expected", func() {
			So(found, ShouldEqual, 3)
		})
	})

	// We expect a 404 if metric is not found
	reply8, err := client.FetchMetrics(
		context.Background(),
		&rpc.FetchMetricsRequest{
			Namespace: []string{"titanium", "poodle"},
			Version:   1,
		},
	)
	Convey("Requesting a namespace that is not in the catalog", t, func() {
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("And reply should be nil", func() {
			So(reply8, ShouldBeNil)
		})
	})

	reply9, err := client.GetMetric(
		context.Background(),
		&rpc.FetchMetricsRequest{
			Namespace: []string{"intel", "mock", "foo"},
			Version:   1,
		},
	)
	Convey("Geting a single metric from the catalog", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		metric, err := rpc.ReplyToMetric(reply9)
		Convey("Should return that metric", func() {
			So(metric.Namespace, ShouldResemble, []string{"intel", "mock", "foo"})
		})
		Convey("Reply error should be nil", func() {
			So(err, ShouldBeNil)
		})
	})

	reply10, err := client.GetMetric(
		context.Background(),
		&rpc.FetchMetricsRequest{
			Namespace: []string{"titanium", "poodle"},
			Version:   1,
		},
	)
	Convey("Requesting a metric that doesn't exist from the catalog", t, func() {
		Convey("Should fail if metric doesn't exist", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("And reply should be nil", func() {
			So(reply10, ShouldBeNil)
		})
	})

	rp2, err := core.NewRequestedPlugin(PluginPath)
	if err != nil {
		log.Fatal(err)
	}
	pluginBytes, err = ioutil.ReadFile(rp2.Path())
	if err != nil {
		log.Fatal(err)
	}
	checkSum = rp2.CheckSum()
	_, err = client.Load(
		context.Background(),
		&rpc.PluginRequest{
			Name:       "snap-collector-mock2",
			CheckSum:   checkSum[:],
			Signature:  rp2.Signature(),
			PluginFile: pluginBytes,
		},
	)
	if err != nil {
		panic(err)
	}
	reply11, err := client.GetMetricVersions(
		context.Background(),
		&rpc.GetMetricVersionsRequest{
			Namespace: []string{"intel", "mock", "foo"},
		},
	)
	Convey("Getting all versions of a given metric in the metric catalog", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		metrics, err := rpc.ReplyToMetrics(reply11.Metrics)
		Convey("Should return all versions of that metric", func() {
			So(len(metrics), ShouldEqual, 2)
		})
		Convey("Reply error should be nil", func() {
			So(err, ShouldBeNil)
		})
	})

	reply12, err := client.GetMetricVersions(
		context.Background(),
		&rpc.GetMetricVersionsRequest{
			Namespace: []string{"titanium", "poodle"},
		},
	)
	Convey("Requesting all versions for a metric that doesn't exist", t, func() {

		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("And reply should be nil", func() {
			So(reply12, ShouldBeNil)
		})
	})

	reply13, err := client.AvailablePlugins(context.Background(), &common.Empty{})
	Convey("Listing Available plugins", t, func() {
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should return no plugins when none are available", func() {
			So(len(reply13.Plugins), ShouldEqual, 0)
		})
	})

	arg := &rpc.UnloadPluginRequest{
		Name:       "mock",
		Version:    1,
		PluginType: "collector",
	}
	reply14, err := client.Unload(context.Background(), arg)
	Convey("Unloading a plugin", t, func() {
		Convey("Should not return an error if the plugin is loaded", func() {
			So(err, ShouldBeNil)
		})
		Convey("Reply should match plugin requested to be unloaded", func() {
			So(reply14.Name, ShouldEqual, "mock")
			So(reply14.Version, ShouldEqual, 1)
			So(reply14.TypeName, ShouldEqual, "collector")
		})
	})
	reply15, err := client.Unload(context.Background(), arg)
	Convey("Requesting to unload a plugin not loaded", t, func() {
		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Reply should be nil", func() {
			So(reply15, ShouldBeNil)
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
