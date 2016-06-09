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

package control

import (
	"bytes"
	"encoding/gob"
	"net"
	"path"
	"testing"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/strategy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
	. "github.com/smartystreets/goconvey/convey"
)

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
	// Load 3 plugins
	// collector -- mock
	// processor -- passthru
	// publisher -- file
	mock, err := core.NewRequestedPlugin(fixtures.JSONRPCPluginPath)
	if err != nil {
		log.Fatal(err)
	}
	c.Load(mock)
	passthru, err := core.NewRequestedPlugin(path.Join(fixtures.SnapPath, "plugin", "snap-processor-passthru"))
	if err != nil {
		log.Fatal(err)
	}
	c.Load(passthru)
	filepub, err := core.NewRequestedPlugin(path.Join(fixtures.SnapPath, "plugin", "snap-publisher-file"))
	if err != nil {
		log.Fatal(err)
	}
	c.Load(filepub)

	conn, err := rpcutil.GetClientConnection(c.Config.ListenAddr, c.Config.ListenPort)

	Convey("Creating an rpc connection", t, func() {
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
	})

	client := rpc.NewMetricManagerClient(conn)

	Convey("Creating an RPC client to control RPC server", t, func() {
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

	// Verify that validate deps is properly passing through errors
	Convey("Validating Deps", t, func() {
		Convey("Should Fail if given invalid info", func() {
			req := &rpc.ValidateDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{fixtures.InvalidMetric}),
				Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{}),
			}
			reply, err := client.ValidateDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})
		Convey("with valid metrics", func() {
			req := &rpc.ValidateDepsRequest{
				Metrics: common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.InvalidMetric}),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.InvalidMetric}),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
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
				Namespace: common.ToNamespace(fixtures.InvalidMetric.Namespace()),
			}
			reply, err := client.MatchQueryToNamespaces(context.Background(), req)
			// we don't expect rpc.errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldNotBeNil)
			So(reply.Error.ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with invalid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()),
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
				Namespace: common.ToNamespace(fixtures.InvalidMetric.Namespace()),
			}
			reply, err := client.ExpandWildcards(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(reply.Error, ShouldNotBeNil)
			So(reply.Error.ErrorString, ShouldResemble, "Metric not found: /this/is/invalid")
		})
		Convey("Should not error with valid inputs", func() {
			req := &rpc.ExpandWildcardsRequest{
				Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.InvalidMetric}),
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
				Metrics: common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
				Deadline: &common.Time{
					Sec:  int64(time.Now().Unix()),
					Nsec: int64(time.Now().Nanosecond()),
				},
				TaskID: "my-snowflake-id",
			}
			reply, err := client.CollectMetrics(context.Background(), req)
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
			So(reply.Metrics[0].Namespace, ShouldResemble, common.ToNamespace(fixtures.ValidMetric.Namespace()))
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
			metrics := make([]plugin.MetricType, len(mts))
			for i, m := range mts {
				metrics[i] = m.(plugin.MetricType)
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
