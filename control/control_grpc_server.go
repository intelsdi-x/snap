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
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
	"golang.org/x/net/context"
)

type ControlGRPCServer struct {
	control *pluginControl
}

// --------- Scheduler's managesMetrics implementation ----------

func (pc *ControlGRPCServer) GetPluginContentTypes(ctx context.Context, r *rpc.GetPluginContentTypesRequest) (*rpc.GetPluginContentTypesReply, error) {
	accepted, returned, err := pc.control.GetPluginContentTypes(r.Name, core.PluginType(int(r.PluginType)), int(r.Version))
	reply := &rpc.GetPluginContentTypesReply{
		AcceptedTypes: accepted,
		ReturnedTypes: returned,
	}
	if err != nil {
		reply.Error = err.Error()
	}
	return reply, nil
}

func (pc *ControlGRPCServer) PublishMetrics(ctx context.Context, r *rpc.PubProcMetricsRequest) (*rpc.ErrorReply, error) {
	errs := pc.control.PublishMetrics(r.ContentType, r.Content, r.PluginName, int(r.PluginVersion), common.ParseConfig(r.Config), r.TaskId)
	erro := make([]string, len(errs))
	for i, v := range errs {
		erro[i] = v.Error()
	}
	return &rpc.ErrorReply{Errors: erro}, nil
}

func (pc *ControlGRPCServer) ProcessMetrics(ctx context.Context, r *rpc.PubProcMetricsRequest) (*rpc.ProcessMetricsReply, error) {
	contentType, content, errs := pc.control.ProcessMetrics(r.ContentType, r.Content, r.PluginName, int(r.PluginVersion), common.ParseConfig(r.Config), r.TaskId)
	erro := make([]string, len(errs))
	for i, v := range errs {
		erro[i] = v.Error()
	}
	reply := &rpc.ProcessMetricsReply{
		ContentType: contentType,
		Content:     content,
		Errors:      erro,
	}
	return reply, nil
}

func (pc *ControlGRPCServer) CollectMetrics(ctx context.Context, r *rpc.CollectMetricsRequest) (*rpc.CollectMetricsResponse, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	deadline := time.Unix(r.Deadline.Sec, r.Deadline.Nsec)
	var AllTags map[string]map[string]string
	for k, v := range r.AllTags {
		AllTags[k] = make(map[string]string)
		for _, entry := range v.Entries {
			AllTags[k][entry.Key] = entry.Value
		}
	}
	mts, errs := pc.control.CollectMetrics(metrics, deadline, r.TaskID, AllTags)
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

func (pc *ControlGRPCServer) ExpandWildcards(ctx context.Context, r *rpc.ExpandWildcardsRequest) (*rpc.ExpandWildcardsReply, error) {
	nss, serr := pc.control.ExpandWildcards(common.ToCoreNamespace(r.Namespace))
	reply := &rpc.ExpandWildcardsReply{}
	if nss != nil {
		reply.NSS = convertNSS(nss)
	}
	if serr != nil {
		reply.Error = common.NewErrors([]serror.SnapError{serr})[0]
	}
	return reply, nil
}

func (pc *ControlGRPCServer) ValidateDeps(ctx context.Context, r *rpc.ValidateDepsRequest) (*rpc.ValidateDepsReply, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	plugins := common.ToSubPlugins(r.Plugins)
	serrors := pc.control.ValidateDeps(metrics, plugins)
	return &rpc.ValidateDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) SubscribeDeps(ctx context.Context, r *rpc.SubscribeDepsRequest) (*rpc.SubscribeDepsReply, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	plugins := common.MsgToCorePlugins(r.Plugins)
	serrors := pc.control.SubscribeDeps(r.TaskId, metrics, plugins)
	return &rpc.SubscribeDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) UnsubscribeDeps(ctx context.Context, r *rpc.SubscribeDepsRequest) (*rpc.SubscribeDepsReply, error) {
	metrics := common.ToCoreMetrics(r.Metrics)
	plugins := common.MsgToCorePlugins(r.Plugins)
	serrors := pc.control.UnsubscribeDeps(r.TaskId, metrics, plugins)
	return &rpc.SubscribeDepsReply{Errors: common.NewErrors(serrors)}, nil
}

func (pc *ControlGRPCServer) MatchQueryToNamespaces(ctx context.Context, r *rpc.ExpandWildcardsRequest) (*rpc.ExpandWildcardsReply, error) {
	nss, serr := pc.control.MatchQueryToNamespaces(common.ToCoreNamespace(r.Namespace))
	reply := &rpc.ExpandWildcardsReply{}
	if nss != nil {
		reply.NSS = convertNSS(nss)
	}
	if serr != nil {
		reply.Error = common.NewErrors([]serror.SnapError{serr})[0]
	}
	return reply, nil
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
