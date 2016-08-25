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
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
	"golang.org/x/net/context"
)

type ControlGRPCServer struct {
	control *pluginControl
}

// --------- Scheduler's managesMetrics implementation ----------
func (pc *ControlGRPCServer) PublishMetrics(ctx context.Context, r *rpc.PubProcMetricsRequest) (*rpc.ErrorReply, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	errs := pc.control.PublishMetrics(
		metrics,
		common.ParseConfig(r.Config),
		r.TaskId, r.PluginName,
		int(r.PluginVersion))

	return &rpc.ErrorReply{Errors: errorsToStrings(errs)}, nil
}

func (pc *ControlGRPCServer) ProcessMetrics(ctx context.Context, r *rpc.PubProcMetricsRequest) (*rpc.ProcessMetricsReply, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	mts, errs := pc.control.ProcessMetrics(
		metrics,
		common.ParseConfig(r.Config),
		r.TaskId, r.PluginName,
		int(r.PluginVersion))

	reply := &rpc.ProcessMetricsReply{
		Metrics: common.NewMetrics(mts),
		Errors:  errorsToStrings(errs),
	}
	return reply, nil
}

func (pc *ControlGRPCServer) CollectMetrics(ctx context.Context, r *rpc.CollectMetricsRequest) (*rpc.CollectMetricsResponse, error) {
	var AllTags map[string]map[string]string
	for k, v := range r.AllTags {
		AllTags[k] = make(map[string]string)
		for _, entry := range v.Entries {
			AllTags[k][entry.Key] = entry.Value
		}
	}
	mts, errs := pc.control.CollectMetrics(r.TaskID, AllTags)
	var reply *rpc.CollectMetricsResponse
	if mts == nil {
		reply = &rpc.CollectMetricsResponse{
			Errors: errorsToStrings(errs),
		}
	} else {
		reply = &rpc.CollectMetricsResponse{
			Metrics: common.NewMetrics(mts),
			Errors:  errorsToStrings(errs),
		}
	}
	return reply, nil
}

func (pc *ControlGRPCServer) ValidateDeps(ctx context.Context, r *rpc.ValidateDepsRequest) (*rpc.ValidateDepsReply, error) {
	metrics := common.ToRequestedMetrics(r.Metrics)
	plugins := common.ToSubPlugins(r.Plugins)
	configTree := cdata.NewTree()
	serrors := pc.control.ValidateDeps(metrics, plugins, configTree)
	return &rpc.ValidateDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) SubscribeDeps(ctx context.Context, r *rpc.SubscribeDepsRequest) (*rpc.SubscribeDepsReply, error) {
	plugins := common.ToSubPlugins(r.Plugins)
	configTree := cdata.NewTree()
	requested := common.MetricToRequested(r.Requested)
	serrors := pc.control.SubscribeDeps(r.TaskId, requested, plugins, configTree)
	return &rpc.SubscribeDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) UnsubscribeDeps(ctx context.Context, r *rpc.UnsubscribeDepsRequest) (*rpc.UnsubscribeDepsReply, error) {
	serrors := pc.control.UnsubscribeDeps(r.TaskId)
	return &rpc.UnsubscribeDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) GetAutodiscoverPaths(ctx context.Context, _ *common.Empty) (*rpc.GetAutodiscoverPathsReply, error) {
	paths := pc.control.GetAutodiscoverPaths()
	reply := &rpc.GetAutodiscoverPathsReply{
		Paths: paths,
	}
	return reply, nil
}

//-------- util ---------------

func convertNSS(nss []core.Namespace) []*rpc.ArrString {
	res := make([]*rpc.ArrString, len(nss))
	for i := range nss {
		var tmp rpc.ArrString
		tmp.S = common.ToNamespace(nss[i])
		res[i] = &tmp
	}
	return res
}

func errorsToStrings(in []error) []string {
	if len(in) == 0 {
		return []string{}
	}
	erro := make([]string, len(in))
	for i, e := range in {
		erro[i] = e.Error()
	}
	return erro
}
