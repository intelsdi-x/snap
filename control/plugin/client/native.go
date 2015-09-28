/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package client

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"time"
	"unicode"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// CallsRPC provides an interface for RPC clients
type CallsRPC interface {
	Call(methd string, args interface{}, reply interface{}) error
}

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	connection CallsRPC
	pluginType plugin.PluginType
}

func NewCollectorNativeClient(address string, timeout time.Duration) (PluginCollectorClient, error) {
	return newNativeClient(address, timeout, plugin.CollectorPluginType)
}

func NewPublisherNativeClient(address string, timeout time.Duration) (PluginPublisherClient, error) {
	return newNativeClient(address, timeout, plugin.PublisherPluginType)
}

func NewProcessorNativeClient(address string, timeout time.Duration) (PluginProcessorClient, error) {
	return newNativeClient(address, timeout, plugin.ProcessorPluginType)
}

func (p *PluginNativeClient) Ping() error {
	a := plugin.PingArgs{}
	b := true
	err := p.connection.Call("SessionState.Ping", a, &b)
	return err
}

func (p *PluginNativeClient) Kill(reason string) error {
	a := plugin.KillArgs{Reason: reason}
	var b bool
	err := p.connection.Call("SessionState.Kill", a, &b)
	return err
}

func (p *PluginNativeClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	args := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}
	reply := plugin.PublishReply{}

	err := p.connection.Call("Publisher.Publish", args, &reply)

	return err
}

func (p *PluginNativeClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	args := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}
	reply := plugin.ProcessorReply{}

	err := p.connection.Call("Processor.Process", args, &reply)

	return reply.ContentType, reply.Content, err
}

func (p *PluginNativeClient) CollectMetrics(coreMetricTypes []core.Metric) ([]core.Metric, error) {
	// Convert core.MetricType slice into plugin.PluginMetricType slice as we have
	// to send structs over RPC
	if len(coreMetricTypes) == 0 {
		return nil, errors.New("no metrics to collect")
	}

	var fromCache []core.Metric
	for i, mt := range coreMetricTypes {
		var metric core.Metric
		// Attempt to retreive the metric from the cache. If it is available,
		// nil out that entry in the requested collection.
		if metric = metricCache.get(core.JoinNamespace(mt.Namespace())); metric != nil {
			fromCache = append(fromCache, metric)
			coreMetricTypes[i] = nil
		}
	}
	// If the size of fromCache is equal to the length of the requested metrics,
	// then we retrieved all of the requested metrics and do not need to go the
	// motions of the rpc call.
	if len(fromCache) != len(coreMetricTypes) {
		var pluginMetricTypes []plugin.PluginMetricType
		// Walk through the requested collection. If the entry is not nil,
		// add it to the slice of metrics to collect over rpc.
		for i, mt := range coreMetricTypes {
			if mt != nil {
				pluginMetricTypes = append(pluginMetricTypes, plugin.PluginMetricType{
					Namespace_:          mt.Namespace(),
					LastAdvertisedTime_: mt.LastAdvertisedTime(),
					Version_:            mt.Version(),
				})
				if mt.Config() != nil {
					pluginMetricTypes[i].Config_ = mt.Config()
				}
			}
		}

		args := plugin.CollectMetricsArgs{PluginMetricTypes: pluginMetricTypes}
		reply := plugin.CollectMetricsReply{}

		err := p.connection.Call("Collector.CollectMetrics", args, &reply)

		var offset int
		for i, mt := range fromCache {
			coreMetricTypes[i] = mt
			offset++
		}
		for i, mt := range reply.PluginMetrics {
			metricCache.put(core.JoinNamespace(mt.Namespace_), mt)
			coreMetricTypes[i+offset] = mt
		}
		return coreMetricTypes, err
	}
	return fromCache, nil
}

func (p *PluginNativeClient) GetMetricTypes() ([]core.Metric, error) {
	args := plugin.GetMetricTypesArgs{}
	reply := plugin.GetMetricTypesReply{}

	err := p.connection.Call("Collector.GetMetricTypes", args, &reply)

	retMetricTypes := make([]core.Metric, len(reply.PluginMetricTypes))
	for i, _ := range reply.PluginMetricTypes {
		// Set the advertised time
		reply.PluginMetricTypes[i].LastAdvertisedTime_ = time.Now()
		retMetricTypes[i] = reply.PluginMetricTypes[i]
	}
	return retMetricTypes, err
}

func (p *PluginNativeClient) GetConfigPolicy() (cpolicy.ConfigPolicy, error) {
	args := plugin.GetConfigPolicyArgs{}
	reply := plugin.GetConfigPolicyReply{Policy: *cpolicy.New()}
	err := p.connection.Call(fmt.Sprintf("%s.GetConfigPolicy", p.GetType()), args, &reply)
	if err != nil {
		return cpolicy.ConfigPolicy{}, err
	}

	return reply.Policy, nil
}

// GetType returns the string type of the plugin
// Note: the first letter of the type will be capitalized.
func (p *PluginNativeClient) GetType() string {
	return upcaseInitial(p.pluginType.String())
}

func newNativeClient(address string, timeout time.Duration, t plugin.PluginType) (*PluginNativeClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &PluginNativeClient{connection: r, pluginType: t}
	return p, nil
}

func init() {
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
	gob.Register(&cpolicy.IntRule{})
	gob.Register(&cpolicy.FloatRule{})
}

func upcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
