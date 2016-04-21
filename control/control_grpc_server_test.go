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
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/controlproxy"
	"github.com/intelsdi-x/snap/internal/common"
	. "github.com/smartystreets/goconvey/convey"
)

// Test the functionality of control grpc server
// that is used by the REST interface
func TestControlGRPCServer(t *testing.T) {
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

// This test is meant to cover the grpc implementation of the subset of control
// features that scheduler uses. It is not intended to test the control features
// themselves, only that we are correctly passing data over grpc and correctly
// passing success/errors.
func TestGRPCServerScheduler(t *testing.T) {
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
	// TODO(CDR): Make sure all stuff in controlproxy.go does not  reference reply's
	// Load 3 plugins
	// collector -- mock
	// processor -- passthru
	// publisher -- file
	mock, err := core.NewRequestedPlugin(JSONRPCPluginPath)
	if err != nil {
		log.Fatal(err)
	}
	c.Load(mock)
	passthru, err := core.NewRequestedPlugin(path.Join(SnapPath, "plugin", "snap-processor-passthru"))
	if err != nil {
		log.Fatal(err)
	}
	c.Load(passthru)
	filepub, err := core.NewRequestedPlugin(path.Join(SnapPath, "plugin", "snap-publisher-file"))
	if err != nil {
		log.Fatal(err)
	}
	c.Load(filepub)
	client, err := NewClient(c.Config.ListenAddr, c.Config.ListenPort)
	Convey("Creating an RPC client to control RPC server", t, func() {
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		Convey("And a client should exist", func() {
			So(client, ShouldNotBeNil)
		})
	})
	//GetContentTypes
	Convey("Getting Content Types", t, func() {
		Convey("Should err if invalid plugin given", func() {
			req := &rpc.GetPluginContentTypesRequest{
				Name:       "bogus",
				PluginType: int32(0),
				Version:    int32(0),
			}
			reply, err := client.GetPluginContentTypes(context.Background(), req)
			// We don't expect rpc errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldNotEqual, "")
			So(reply.Error, ShouldResemble, "plugin not found")
		})
		Convey("Should return content types with valid plugin", func() {
			req := &rpc.GetPluginContentTypesRequest{
				Name:       "mock",
				PluginType: int32(0),
				Version:    0,
			}
			reply, err := client.GetPluginContentTypes(context.Background(), req)
			So(err, ShouldBeNil)
			So(reply.Error, ShouldEqual, "")
			So(reply.AcceptedTypes, ShouldContain, "snap.gob")
			So(reply.ReturnedTypes, ShouldContain, "snap.gob")
		})
	})

	validMetric := MockMetricType{
		namespace: []string{"intel", "mock", "foo"},
		cfg:       cdata.NewNode(),
		ver:       0,
	}
	invalidMetric := MockMetricType{
		namespace: []string{"this", "is", "invalid"},
		cfg:       cdata.NewNode(),
		ver:       1000,
	}

	// Verify that validate deps is properly passing through errors
	Convey("Validating Deps", t, func() {
		Convey("Should Fail if given invalid info", func() {
			req := &rpc.ValidateDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{invalidMetric}),
				Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{}),
			}
			reply, err := client.ValidateDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})
		Convey("with valid metrics", func() {
			req := &rpc.ValidateDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{validMetric}),
				Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{}),
			}
			reply, err := client.ValidateDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})
	})
	//Subscribe Deps: valid/invalid
	Convey("SubscribeDeps", t, func() {
		Convey("Should Error with invalid inputs", func() {
			req := &rpc.SubscribeDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{invalidMetric}),
				Plugins: common.ToCorePluginsMsg([]core.Plugin{}),
				TaskId:  "my-snowflake-id",
			}
			reply, err := client.SubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
			So(reply.Errors[0].ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with valid inputs", func() {
			req := &rpc.SubscribeDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{validMetric}),
				Plugins: common.ToCorePluginsMsg([]core.Plugin{}),
				TaskId:  "my-snowflake-id",
			}
			reply, err := client.SubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
		})
	})
	// unsubscribedeps -- valid/invalid
	Convey("UnsubscribeDeps", t, func() {
		Convey("Should Error with invalid inputs", func() {
			req := &rpc.SubscribeDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{invalidMetric}),
				Plugins: common.ToCorePluginsMsg([]core.Plugin{}),
				TaskId:  "my-snowflake-id",
			}
			reply, err := client.UnsubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
			So(reply.Errors[0].ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with valid inputs", func() {
			req := &rpc.SubscribeDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{validMetric}),
				Plugins: common.ToCorePluginsMsg([]core.Plugin{}),
				TaskId:  "my-snowflake-id",
			}
			reply, err := client.UnsubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
		})
	})
	//matchquerytonamespaces -- valid/invalid
	Convey("MatchingQueryToNamespaces", t, func() {
		Convey("Should error with invalid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: invalidMetric.Namespace(),
			}
			reply, err := client.MatchQueryToNamespaces(context.Background(), req)
			// we don't expect rpc.errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldNotBeNil)
			So(reply.Error.ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with invalid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: validMetric.Namespace(),
			}
			reply, err := client.MatchQueryToNamespaces(context.Background(), req)
			// we don't expect rpc.errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldBeNil)
		})
	})
	//expandwildcards -- valid/invalid
	Convey("ExpandWildcards", t, func() {
		Convey("Should error with invalid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: invalidMetric.Namespace(),
			}
			reply, err := client.ExpandWildcards(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldNotBeNil)
			So(reply.Error.ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with valid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: validMetric.Namespace(),
			}
			reply, err := client.ExpandWildcards(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldBeNil)
		})
	})
	// start plugin pools/provide task info so we can do collect/process/publishMetrics
	// errors here indicate problems outside the scope of this test.
	plugins := []string{"collector:mock:1", "processor:passthru:1", "publisher:file:3"}
	lps := make([]*loadedPlugin, len(plugins))
	pools := make([]strategy.Pool, len(plugins))
	for i, v := range plugins {
		lps[i], err = c.pluginManager.get(v)
		if err != nil {
			log.Fatal(err)
		}
		pools[i], err = c.pluginRunner.AvailablePlugins().getOrCreatePool(v)
		if err != nil {
			log.Fatal(err)
		}
		pools[i].Subscribe("my-snowflake-id", strategy.BoundSubscriptionType)
		err = c.pluginRunner.runPlugin(lps[i].Details)
		if err != nil {
			log.Fatal(err)
		}
	}
	//our returned metrics
	var mts []core.Metric
	//collect
	Convey("CollectMetrics", t, func() {
		Convey("Should error with invalid inputs", func() {
			req := &rpc.CollectMetricsRequest{
				Metrics: common.NewMetrics([]core.Metric{invalidMetric}),
				Deadline: &common.Time{
					Sec:  int64(time.Now().Unix()),
					Nsec: int64(time.Now().Nanosecond()),
				},
				TaskID: "my-snowflake-id",
			}
			reply, err := client.CollectMetrics(context.Background(), req)
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})
		Convey("should not error with valid inputs", func() {
			req := &rpc.CollectMetricsRequest{
				Metrics: common.NewMetrics([]core.Metric{validMetric}),
				Deadline: &common.Time{
					Sec:  int64(time.Now().Unix()),
					Nsec: int64(time.Now().Nanosecond()),
				},
				TaskID: "my-snowflake-id",
			}
			reply, err := client.CollectMetrics(context.Background(), req)
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
			So(reply.Metrics[0].Namespace, ShouldResemble, []string{"intel", "mock", "foo"})
			// Used in a later test as metrics to be passed to processor
			mts = common.ToCoreMetrics(reply.Metrics)
		})
	})
	//our content to pass to publish
	var content []byte
	//process
	Convey("ProcessMetrics", t, func() {
		Convey("Should error with invalid inputs", func() {
			req := controlproxy.GetPubProcReq("snap.gob", []byte{}, "bad name", 1, map[string]ctypes.ConfigValue{}, "my-snowflake-id")
			reply, err := client.ProcessMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
			So(reply.Errors[0], ShouldResemble, "bad key")
		})
		Convey("should not error with valid inputs", func() {
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			metrics := make([]plugin.PluginMetricType, len(mts))
			for i, m := range mts {
				metrics[i] = m.(plugin.PluginMetricType)
			}
			enc.Encode(metrics)
			req := controlproxy.GetPubProcReq("snap.gob", buf.Bytes(), "passthru", 1, map[string]ctypes.ConfigValue{}, "my-snowflake-id")
			reply, err := client.ProcessMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
			// content to pass to publisher
			content = reply.Content

		})
	})
	//publishmetrics
	Convey("PublishMetrics", t, func() {
		Convey("Should error with invalid inputs", func() {
			req := controlproxy.GetPubProcReq("snap.gob", []byte{}, "bad name", 1, map[string]ctypes.ConfigValue{}, "my-snowflake-id")
			reply, err := client.PublishMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)

			So(reply.Errors[0], ShouldResemble, "bad key")
		})
		// Publish only returns no errors on success
		Convey("should not error with valid inputs", func() {
			config := make(map[string]ctypes.ConfigValue)
			config["file"] = ctypes.ConfigValueStr{Value: "/tmp/grpcservertest.snap"}
			req := controlproxy.GetPubProcReq("snap.gob", content, "file", 3, config, "my-snowflake-id")
			reply, err := client.PublishMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
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
