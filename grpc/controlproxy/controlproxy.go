/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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
	"context"
	"errors"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/grpc/controlproxy/rpc"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
)

var (
	MAX_CONNECTION_TIMEOUT = 10 * time.Second
)

func getContext() context.Context {
	cd, _ := context.WithTimeout(context.Background(), MAX_CONNECTION_TIMEOUT)
	return cd
}

// Implements managesMetrics interface provided by scheduler and
// proxies those calls to the grpc client.
type ControlProxy struct {
	Client rpc.MetricManagerClient
}

func New(addr string, port int) (ControlProxy, error) {
	conn, err := rpcutil.GetClientConnection(context.Background(), addr, port)
	if err != nil {
		return ControlProxy{}, err
	}
	c := rpc.NewMetricManagerClient(conn)
	return ControlProxy{Client: c}, nil
}

func (c ControlProxy) PublishMetrics(metrics []core.Metric,
	config map[string]ctypes.ConfigValue,
	taskId string,
	pluginName string,
	pluginVersion int) []error {

	req := &rpc.PubProcMetricsRequest{
		Metrics:       common.NewMetrics(metrics),
		PluginName:    pluginName,
		PluginVersion: int64(pluginVersion),
		TaskId:        taskId,
		Config:        common.ToConfigMap(config),
	}
	reply, err := c.Client.PublishMetrics(getContext(), req)
	var errs []error
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	rerrs := replyErrorsToErrors(reply.Errors)
	errs = append(errs, rerrs...)
	return errs
}

func (c ControlProxy) ProcessMetrics(metrics []core.Metric,
	config map[string]ctypes.ConfigValue,
	taskId string,
	pluginName string,
	pluginVersion int) ([]core.Metric, []error) {
	req := &rpc.PubProcMetricsRequest{
		Metrics:       common.NewMetrics(metrics),
		PluginName:    pluginName,
		PluginVersion: int64(pluginVersion),
		TaskId:        taskId,
		Config:        common.ToConfigMap(config),
	}
	reply, err := c.Client.ProcessMetrics(getContext(), req)
	var errs []error
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	rerrs := replyErrorsToErrors(reply.Errors)
	errs = append(errs, rerrs...)
	return common.ToCoreMetrics(reply.Metrics), errs
}

func (c ControlProxy) CollectMetrics(taskID string, AllTags map[string]map[string]string) ([]core.Metric, []error) {
	var allTags map[string]*rpc.Map
	for k, v := range AllTags {
		tags := &rpc.Map{}
		for kn, vn := range v {
			entry := &rpc.MapEntry{
				Key:   kn,
				Value: vn,
			}
			tags.Entries = append(tags.Entries, entry)
		}
		allTags[k] = tags
	}
	req := &rpc.CollectMetricsRequest{
		TaskID:  taskID,
		AllTags: allTags,
	}
	reply, err := c.Client.CollectMetrics(getContext(), req)
	var errs []error
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	rerrs := replyErrorsToErrors(reply.Errors)
	if len(rerrs) > 0 {
		errs = append(errs, rerrs...)
		return nil, errs
	}
	metrics := common.ToCoreMetrics(reply.Metrics)
	return metrics, nil
}

func (c ControlProxy) StreamMetrics(
	_ string,
	_ map[string]map[string]string,
	_ time.Duration,
	_ int64) (chan []core.Metric, chan error, []error) {
	return nil, nil, []error{
		errors.New("Streaming not supported in distributed workflows"),
	}
}

func (c ControlProxy) ValidateDeps(mts []core.RequestedMetric, plugins []core.SubscribedPlugin, _ *cdata.ConfigDataTree, _ ...core.SubscribedPluginAssert) []serror.SnapError {
	// The configDataTree is kept so that we can match the interface provided by control
	// we do not need it here though since the configDataTree is only used for metrics
	// and we do not allow remote collection.
	req := &rpc.ValidateDepsRequest{
		Metrics: common.RequestedToMetric(mts),
		Plugins: common.ToSubPluginsMsg(plugins),
	}
	reply, err := c.Client.ValidateDeps(getContext(), req)
	if err != nil {
		return []serror.SnapError{serror.New(err)}
	}
	serrs := common.ConvertSnapErrors(reply.Errors)
	return serrs
}

func (c ControlProxy) SubscribeDeps(taskID string, requested []core.RequestedMetric, plugins []core.SubscribedPlugin, configTree *cdata.ConfigDataTree) []serror.SnapError {
	req := depsRequest(taskID, requested, plugins, configTree)
	reply, err := c.Client.SubscribeDeps(getContext(), req)
	if err != nil {
		return []serror.SnapError{serror.New(err)}
	}
	serrs := common.ConvertSnapErrors(reply.Errors)
	return serrs
}

func (c ControlProxy) UnsubscribeDeps(taskID string) []serror.SnapError {
	req := &rpc.UnsubscribeDepsRequest{TaskId: taskID}
	reply, err := c.Client.UnsubscribeDeps(getContext(), req)
	if err != nil {
		return []serror.SnapError{serror.New(err)}
	}
	serrs := common.ConvertSnapErrors(reply.Errors)
	return serrs
}

func (c ControlProxy) GetAutodiscoverPaths() []string {
	req := &common.Empty{}
	reply, err := c.Client.GetAutodiscoverPaths(getContext(), req)
	if err != nil {
		return nil
	}
	return reply.Paths
}

///---------Util-------------------------------------------------------------------------
func getPluginType(t core.PluginType) int32 {
	val := int32(-1)
	switch t {
	case core.CollectorPluginType:
		val = 0
	case core.ProcessorPluginType:
		val = 1
	case core.PublisherPluginType:
		val = 2
	}
	return val
}

func depsRequest(taskID string, requested []core.RequestedMetric, plugins []core.SubscribedPlugin, configTree *cdata.ConfigDataTree) *rpc.SubscribeDepsRequest {
	req := &rpc.SubscribeDepsRequest{
		Requested: common.RequestedToMetric(requested),
		Plugins:   common.ToSubPluginsMsg(plugins),
		TaskId:    taskID,
	}
	return req
}

func toNSS(arr []*rpc.ArrString) []core.Namespace {
	nss := make([]core.Namespace, len(arr))
	for i, v := range arr {
		nss[i] = common.ToCoreNamespace(v.S)
	}
	return nss
}

func replyErrorsToErrors(errs []string) []error {
	if len(errs) == 0 {
		return []error{}
	}
	var erro []error
	for _, e := range errs {
		erro = append(erro, errors.New(e))
	}
	return erro
}
