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
	"errors"
	"time"

	"golang.org/x/net/context"

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
	conn, err := rpcutil.GetClientConnection(addr, port)
	if err != nil {
		return ControlProxy{}, err
	}
	c := rpc.NewMetricManagerClient(conn)
	return ControlProxy{Client: c}, nil
}

func (c ControlProxy) PublishMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) []error {
	req := GetPubProcReq(contentType, content, pluginName, pluginVersion, config, taskID)
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

func (c ControlProxy) ProcessMetrics(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) (string, []byte, []error) {
	req := GetPubProcReq(contentType, content, pluginName, pluginVersion, config, taskID)
	reply, err := c.Client.ProcessMetrics(getContext(), req)
	var errs []error
	if err != nil {
		errs = append(errs, err)
		return "", nil, errs
	}
	rerrs := replyErrorsToErrors(reply.Errors)
	errs = append(errs, rerrs...)
	return reply.ContentType, reply.Content, errs
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

func (c ControlProxy) GetPluginContentTypes(n string, t core.PluginType, v int) ([]string, []string, error) {
	req := &rpc.GetPluginContentTypesRequest{
		Name:       n,
		PluginType: getPluginType(t),
		Version:    int32(v),
	}
	reply, err := c.Client.GetPluginContentTypes(getContext(), req)
	if err != nil {
		return nil, nil, err
	}
	if reply.Error != "" {
		return nil, nil, errors.New(reply.Error)
	}
	return reply.AcceptedTypes, reply.ReturnedTypes, nil
}

func (c ControlProxy) ValidateDeps(mts []core.RequestedMetric, plugins []core.SubscribedPlugin, ctree *cdata.ConfigDataTree) []serror.SnapError {
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

// Constructs a protobuf message for publish/process given the relevant information
func GetPubProcReq(contentType string, content []byte, pluginName string, pluginVersion int, config map[string]ctypes.ConfigValue, taskID string) *rpc.PubProcMetricsRequest {
	newConfig := common.ToConfigMap(config)
	request := &rpc.PubProcMetricsRequest{
		ContentType:   contentType,
		Content:       content,
		PluginName:    pluginName,
		PluginVersion: int64(pluginVersion),
		Config:        newConfig,
		TaskId:        taskID,
	}
	return request
}
