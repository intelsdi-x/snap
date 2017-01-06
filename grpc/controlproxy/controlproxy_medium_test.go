// +build medium

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

package controlproxy

import (
	"errors"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"

	"github.com/intelsdi-x/snap/core/cdata"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	rpcErr = errors.New("RPC ERROR")
)

type mockClient struct {
	RpcErr           bool
	PublishReply     *rpc.ErrorReply
	ProcessReply     *rpc.ProcessMetricsReply
	CollectReply     *rpc.CollectMetricsResponse
	ValidateReply    *rpc.ValidateDepsReply
	SubscribeReply   *rpc.SubscribeDepsReply
	UnsubscribeReply *rpc.UnsubscribeDepsReply
	AutoDiscoReply   *rpc.GetAutodiscoverPathsReply
}

func (mc mockClient) GetAutodiscoverPaths(ctx context.Context, _ *common.Empty, opts ...grpc.CallOption) (*rpc.GetAutodiscoverPathsReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.AutoDiscoReply, nil
}

func (mc mockClient) CollectMetrics(ctx context.Context, in *rpc.CollectMetricsRequest, opts ...grpc.CallOption) (*rpc.CollectMetricsResponse, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.CollectReply, nil
}
func (mc mockClient) PublishMetrics(ctx context.Context, in *rpc.PubProcMetricsRequest, opts ...grpc.CallOption) (*rpc.ErrorReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.PublishReply, nil
}
func (mc mockClient) ProcessMetrics(ctx context.Context, in *rpc.PubProcMetricsRequest, opts ...grpc.CallOption) (*rpc.ProcessMetricsReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.ProcessReply, nil
}
func (mc mockClient) ValidateDeps(ctx context.Context, in *rpc.ValidateDepsRequest, opts ...grpc.CallOption) (*rpc.ValidateDepsReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.ValidateReply, nil
}
func (mc mockClient) SubscribeDeps(ctx context.Context, in *rpc.SubscribeDepsRequest, opts ...grpc.CallOption) (*rpc.SubscribeDepsReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.SubscribeReply, nil
}
func (mc mockClient) UnsubscribeDeps(ctx context.Context, in *rpc.UnsubscribeDepsRequest, opts ...grpc.CallOption) (*rpc.UnsubscribeDepsReply, error) {
	if mc.RpcErr {
		return nil, rpcErr
	}
	return mc.UnsubscribeReply, nil
}

func TestPublishMetrics(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		errs := proxy.PublishMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "fake", 1)

		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.Publish returns an error", t, func() {
		reply := &rpc.ErrorReply{
			Errors: []string{"errors"},
		}

		proxy := ControlProxy{Client: mockClient{PublishReply: reply}}
		errs := proxy.PublishMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "fake", 1)

		Convey("So err should not be nil", func() {
			So(errs, ShouldNotBeNil)
		})

		Convey("So errs should contain 'errors'", func() {
			So(errs[0].Error(), ShouldResemble, "errors")
		})
	})

	Convey("control.Publish does not error", t, func() {
		reply := &rpc.ErrorReply{Errors: []string{}}

		proxy := ControlProxy{Client: mockClient{PublishReply: reply}}
		errs := proxy.PublishMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "fake", 1)

		Convey("So publishing should not error", func() {
			So(len(errs), ShouldEqual, 0)
		})
	})
}

func TestProcessMetrics(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		_, errs := proxy.ProcessMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "fake", 1)

		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.Process returns an error", t, func() {
		reply := &rpc.ProcessMetricsReply{
			Metrics: nil,
			Errors:  []string{"error in control.Process"},
		}

		proxy := ControlProxy{Client: mockClient{ProcessReply: reply}}
		_, errs := proxy.ProcessMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "", 1)

		Convey("So errs should not be nil", func() {
			So(errs, ShouldNotBeNil)
		})

		Convey("So len of errs should be 1", func() {
			So(len(errs), ShouldEqual, 1)
		})

		Convey("So that error should have content 'error in control.Process'", func() {
			So(errs[0].Error(), ShouldResemble, "error in control.Process")
		})
	})

	Convey("Control.Process returns successfully", t, func() {
		reply := &rpc.ProcessMetricsReply{
			Metrics: nil,
			Errors:  []string{},
		}

		proxy := ControlProxy{Client: mockClient{ProcessReply: reply}}
		_, errs := proxy.ProcessMetrics([]core.Metric{}, map[string]ctypes.ConfigValue{}, "", "", 1)

		Convey("So len of errs should be 0", func() {
			So(len(errs), ShouldEqual, 0)
		})

	})
}

func TestCollectMetrics(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		_, errs := proxy.CollectMetrics("", map[string]map[string]string{})

		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.CollectMetrics returns an error", t, func() {
		reply := &rpc.CollectMetricsResponse{
			Metrics: nil,
			Errors:  []string{"error in collect"},
		}

		proxy := ControlProxy{Client: mockClient{CollectReply: reply}}
		_, errs := proxy.CollectMetrics("", map[string]map[string]string{})

		Convey("So len of errs should be 1", func() {
			So(len(errs), ShouldEqual, 1)
		})

		Convey("So error should contain the string 'error in collect'", func() {
			So(errs[0].Error(), ShouldResemble, "error in collect")
		})
	})

	Convey("Control.CollectMetrics returns successfully", t, func() {
		reply := &rpc.CollectMetricsResponse{
			Metrics: []*common.Metric{&common.Metric{
				Namespace:          common.ToNamespace(core.NewNamespace("testing", "this")),
				Version:            6,
				Tags:               map[string]string{},
				Timestamp:          &common.Time{Sec: time.Now().Unix(), Nsec: int64(time.Now().Nanosecond())},
				LastAdvertisedTime: &common.Time{Sec: time.Now().Unix(), Nsec: int64(time.Now().Nanosecond())},
			}},
			Errors: nil,
		}

		proxy := ControlProxy{Client: mockClient{CollectReply: reply}}
		mts, errs := proxy.CollectMetrics("", map[string]map[string]string{})

		Convey("So len of errs should be 0", func() {
			So(len(errs), ShouldEqual, 0)
		})

		Convey("So mts should not be nil", func() {
			So(mts, ShouldNotBeNil)
		})

		Convey("So len of metrics returned should be 1", func() {
			So(len(mts), ShouldEqual, 1)
		})
	})
}

func TestValidateDeps(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		errs := proxy.ValidateDeps([]core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldBeGreaterThan, 0)
		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.ValidateDeps returns an error", t, func() {
		reply := &rpc.ValidateDepsReply{
			Errors: []*common.SnapError{&common.SnapError{ErrorFields: map[string]string{}, ErrorString: "test"}},
		}

		proxy := ControlProxy{Client: mockClient{ValidateReply: reply}}
		errs := proxy.ValidateDeps([]core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldEqual, 1)
		Convey("So the error should resemble 'test'", func() {
			So(errs[0].Error(), ShouldResemble, "test")
		})

	})

	Convey("Control.ValidateDeps returns successfully", t, func() {
		reply := &rpc.ValidateDepsReply{}

		proxy := ControlProxy{Client: mockClient{ValidateReply: reply}}
		errs := proxy.ValidateDeps([]core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		Convey("So the there should be no errors", func() {
			So(len(errs), ShouldEqual, 0)
		})

	})
}

func TestSubscribeDeps(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		errs := proxy.SubscribeDeps("", []core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldBeGreaterThan, 0)
		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.SubscribeDeps returns an error", t, func() {
		reply := &rpc.SubscribeDepsReply{
			Errors: []*common.SnapError{&common.SnapError{ErrorFields: map[string]string{}, ErrorString: "test"}},
		}

		proxy := ControlProxy{Client: mockClient{SubscribeReply: reply}}
		errs := proxy.SubscribeDeps("", []core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldEqual, 1)
		Convey("So the error should resemble 'test'", func() {
			So(errs[0].Error(), ShouldResemble, "test")
		})

	})

	Convey("Control.SubscribeDeps returns successfully", t, func() {
		reply := &rpc.SubscribeDepsReply{}

		proxy := ControlProxy{Client: mockClient{SubscribeReply: reply}}
		errs := proxy.SubscribeDeps("", []core.RequestedMetric{}, []core.SubscribedPlugin{}, cdata.NewTree())
		Convey("So the there should be no errors", func() {
			So(len(errs), ShouldEqual, 0)
		})

	})
}

func TestUnsubscribeDeps(t *testing.T) {
	Convey("RPC client errors", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		errs := proxy.UnsubscribeDeps("")
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldBeGreaterThan, 0)
		Convey("So the error should be passed through", func() {
			So(errs[0].Error(), ShouldResemble, rpcErr.Error())
		})
	})

	Convey("Control.UnsubscribeDeps returns an error", t, func() {
		reply := &rpc.UnsubscribeDepsReply{
			Errors: []*common.SnapError{&common.SnapError{ErrorFields: map[string]string{}, ErrorString: "test"}},
		}

		proxy := ControlProxy{Client: mockClient{UnsubscribeReply: reply}}
		errs := proxy.UnsubscribeDeps("")
		So(errs, ShouldNotBeNil)
		So(len(errs), ShouldEqual, 1)
		Convey("So the error should resemble 'test'", func() {
			So(errs[0].Error(), ShouldResemble, "test")
		})

	})

	Convey("Control.UnsubscribeDeps returns successfully", t, func() {
		reply := &rpc.UnsubscribeDepsReply{}

		proxy := ControlProxy{Client: mockClient{UnsubscribeReply: reply}}
		errs := proxy.UnsubscribeDeps("")
		Convey("So the there should be no errors", func() {
			So(len(errs), ShouldEqual, 0)
		})

	})
}

func TestGetAutoDiscoverPaths(t *testing.T) {
	Convey("Able to call successfully", t, func() {
		reply := &rpc.GetAutodiscoverPathsReply{
			Paths: []string{"a", "titanium", "poodle"},
		}
		proxy := ControlProxy{Client: mockClient{AutoDiscoReply: reply}}
		val := proxy.GetAutodiscoverPaths()
		So(val, ShouldNotBeNil)
		So(val, ShouldResemble, []string{"a", "titanium", "poodle"})
	})

	Convey("returns nil on rpc error", t, func() {
		proxy := ControlProxy{Client: mockClient{RpcErr: true}}
		val := proxy.GetAutodiscoverPaths()
		So(val, ShouldBeNil)
	})
}
