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
	"net"
	"testing"

	"github.com/intelsdi-x/gomit"
	"golang.org/x/net/context"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
	"github.com/intelsdi-x/snap/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

type listenPluginLoadEvent struct {
	plugin *mockPluginEvent
	done   chan struct{}
}

func newFoo() *listenPluginLoadEvent {
	return &listenPluginLoadEvent{
		done: make(chan struct{}),
	}
}

func (l *listenPluginLoadEvent) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		l.done <- struct{}{}
	default:
		controlLogger.WithFields(log.Fields{
			"event:": v.Namespace(),
			"_block": "HandleGomit",
		}).Info("Unhandled Event")
	}
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

	lpe := newFoo()
	c.eventManager.RegisterHandler("Control.PluginLoaded", lpe)

	Convey("Starting control_proxy server/client", t, func() {
		Convey("So err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
	// Load 3 plugins
	// collector -- mock
	// processor -- passthru
	// publisher -- file
	mock, err := core.NewRequestedPlugin(fixtures.PluginPathMock1, GetDefaultConfig().TempDirPath, nil)
	if err != nil {
		log.Fatal(err)
	}
	_, serr := c.Load(mock)
	Convey("Loading mock collector", t, func() {
		Convey("should not error", func() {
			So(serr, ShouldBeNil)
		})
	})
	<-lpe.done
	passthru, err := core.NewRequestedPlugin(helper.PluginFilePath("snap-plugin-processor-passthru"), GetDefaultConfig().TempDirPath, nil)
	if err != nil {
		log.Fatal(err)
	}
	catalogedPassthru, serr := c.Load(passthru)
	Convey("Loading passthru processor", t, func() {
		Convey("should not error", func() {
			So(serr, ShouldBeNil)
		})
	})
	subscribedPassThruPlugin := subscribedPlugin{
		name:     catalogedPassthru.Name(),
		version:  catalogedPassthru.Version(),
		typeName: catalogedPassthru.TypeName(),
	}
	<-lpe.done
	filepub, err := core.NewRequestedPlugin(helper.PluginFilePath("snap-plugin-publisher-mock-file"), GetDefaultConfig().TempDirPath, nil)
	if err != nil {
		log.Fatal(err)
	}
	catalogedFile, serr := c.Load(filepub)
	Convey("Loading file publisher", t, func() {
		Convey("should not error", func() {
			So(serr, ShouldBeNil)
		})
	})
	subscribedFilePlugin := subscribedPlugin{
		name:     catalogedFile.Name(),
		version:  catalogedFile.Version(),
		typeName: catalogedFile.TypeName(),
	}

	<-lpe.done
	conn, err := rpcutil.GetClientConnection(context.Background(), c.Config.ListenAddr, c.Config.ListenPort)

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
				Requested: []*common.Metric{&common.Metric{Namespace: common.ToNamespace(fixtures.InvalidMetric.Namespace()), Version: int64(fixtures.InvalidMetric.Version())}},
				Plugins:   common.ToSubPluginsMsg([]core.SubscribedPlugin{subscribedFilePlugin, subscribedPassThruPlugin}),
				TaskId:    "my-snowflake-id",
			}
			reply, err := client.SubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
			So(reply.Errors[0].ErrorString, ShouldResemble, "Metric not found: /this/is/invalid (version: 1000)")
		})
		Convey("Should not error with valid inputs", func() {
			req := &rpc.SubscribeDepsRequest{
				Requested: []*common.Metric{&common.Metric{Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()), Version: int64(fixtures.ValidMetric.Version())}},
				Plugins:   common.ToSubPluginsMsg([]core.SubscribedPlugin{}),
				TaskId:    "my-snowflake-valid",
			}
			reply, err := client.SubscribeDeps(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
		})
	})
	//our returned metrics
	var mts []core.Metric
	//collect
	Convey("CollectMetrics", t, func() {
		req := &rpc.SubscribeDepsRequest{
			Requested: []*common.Metric{&common.Metric{Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()), Version: int64(fixtures.ValidMetric.Version())}},
			Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{
				subscribedPassThruPlugin,
				subscribedFilePlugin,
			},
			),
			TaskId: "my-snowflake-id",
		}
		_, err := client.SubscribeDeps(context.Background(), req)
		So(err, ShouldBeNil)

		Convey("should error with invalid inputs", func() {
			req := &rpc.CollectMetricsRequest{
				TaskID: "my-fake-snowflake-id",
			}
			reply, err := client.CollectMetrics(context.Background(), req)
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})

		Convey("should not error with valid inputs", func() {
			req := &rpc.CollectMetricsRequest{
				TaskID: "my-snowflake-valid",
			}
			reply, err := client.CollectMetrics(context.Background(), req)
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)
			So(reply.Metrics[0].Namespace, ShouldResemble, common.ToNamespace(fixtures.ValidMetric.Namespace()))
			// Used in a later test as metrics to be passed to processor
			mts = common.ToCoreMetrics(reply.Metrics)
		})
	})

	//process
	Convey("ProcessMetrics", t, func() {
		req := &rpc.SubscribeDepsRequest{
			Requested: []*common.Metric{&common.Metric{Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()), Version: int64(fixtures.ValidMetric.Version())}},
			Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{
				subscribedPassThruPlugin,
				subscribedFilePlugin,
			},
			),
			TaskId: "my-snowflake-id",
		}
		_, err := client.SubscribeDeps(context.Background(), req)
		So(err, ShouldBeNil)
		Convey("should error with invalid inputs", func() {
			req := &rpc.PubProcMetricsRequest{
				Metrics:       common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
				PluginName:    "passthru-invalid",
				PluginVersion: 1,
				TaskId:        "my-snowflake-id",
				Config:        common.ToConfigMap(map[string]ctypes.ConfigValue{}),
			}
			reply, err := client.ProcessMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
			// content to pass to publisher
		})
		Convey("should not error with valid inputs", func() {
			req := &rpc.PubProcMetricsRequest{
				Metrics:       common.NewMetrics(mts),
				PluginName:    "passthru",
				PluginVersion: 1,
				TaskId:        "my-snowflake-id",
				Config:        common.ToConfigMap(map[string]ctypes.ConfigValue{}),
			}
			reply, err := client.ProcessMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldEqual, 0)

		})
	})
	//publishmetrics
	Convey("PublishMetrics", t, func() {
		req := &rpc.SubscribeDepsRequest{
			Requested: []*common.Metric{&common.Metric{Namespace: common.ToNamespace(fixtures.ValidMetric.Namespace()), Version: int64(fixtures.ValidMetric.Version())}},
			Plugins: common.ToSubPluginsMsg([]core.SubscribedPlugin{
				subscribedPassThruPlugin,
				subscribedFilePlugin,
			},
			),
			TaskId: "my-snowflake-id",
		}
		_, err := client.SubscribeDeps(context.Background(), req)
		So(err, ShouldBeNil)

		Convey("Should error with invalid inputs", func() {
			req := &rpc.PubProcMetricsRequest{
				Metrics:       common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
				PluginName:    "mock-file-invalid",
				PluginVersion: 3,
				TaskId:        "my-snowflake-id",
				Config:        common.ToConfigMap(map[string]ctypes.ConfigValue{}),
			}
			reply, err := client.PublishMetrics(context.Background(), req)
			// we don't expect rpc errors
			So(err, ShouldBeNil)
			So(len(reply.Errors), ShouldNotEqual, 0)
		})

		// Publish only returns no errors on success
		Convey("should not error with valid inputs", func() {
			config := make(map[string]ctypes.ConfigValue)
			config["file"] = ctypes.ConfigValueStr{Value: "/tmp/grpcservertest.snap"}
			req := &rpc.PubProcMetricsRequest{
				Metrics:       common.NewMetrics([]core.Metric{fixtures.ValidMetric}),
				PluginName:    "mock-file",
				PluginVersion: 3,
				TaskId:        "my-snowflake-id",
				Config:        common.ToConfigMap(config),
			}
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
