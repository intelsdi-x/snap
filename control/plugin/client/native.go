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

package client

import (
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"net"
	"net/rpc"
	"time"
	"unicode"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

// CallsRPC provides an interface for RPC clients
type CallsRPC interface {
	Call(methd string, args interface{}, reply interface{}) error
}

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	PluginCacheClient
	connection CallsRPC
	pluginType plugin.PluginType
	encoder    encoding.Encoder
	encrypter  *encrypter.Encrypter
}

func NewCollectorNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginCollectorClient, error) {
	return newNativeClient(address, timeout, plugin.CollectorPluginType, pub, secure)
}

func NewPublisherNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginPublisherClient, error) {
	return newNativeClient(address, timeout, plugin.PublisherPluginType, pub, secure)
}

func NewProcessorNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginProcessorClient, error) {
	return newNativeClient(address, timeout, plugin.ProcessorPluginType, pub, secure)
}

func (p *PluginNativeClient) Ping() error {
	var reply []byte
	err := p.connection.Call("SessionState.Ping", []byte{}, &reply)
	return err
}

func (p *PluginNativeClient) SetKey() error {
	out, err := p.encrypter.EncryptKey()
	if err != nil {
		return err
	}
	return p.connection.Call("SessionState.SetKey", plugin.SetKeyArgs{
		Key: out,
	}, &[]byte{})
}

func (p *PluginNativeClient) Kill(reason string) error {
	args := plugin.KillArgs{Reason: reason}
	out, err := p.encoder.Encode(args)
	if err != nil {
		return err
	}

	var reply []byte
	err = p.connection.Call("SessionState.Kill", out, &reply)
	return err
}

func (p *PluginNativeClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	args := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return err
	}

	var reply []byte
	err = p.connection.Call("Publisher.Publish", out, &reply)
	return err
}

func (p *PluginNativeClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	args := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return "", nil, err
	}

	var reply []byte
	err = p.connection.Call("Processor.Process", out, &reply)
	if err != nil {
		return "", nil, err
	}

	r := plugin.ProcessorReply{}
	err = p.encoder.Decode(reply, &r)
	if err != nil {
		return "", nil, err
	}

	return r.ContentType, r.Content, nil
}

func (p *PluginNativeClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	// Convert core.MetricType slice into plugin.PluginMetricType slice as we have
	// to send structs over RPC
	var results []core.Metric
	if len(mts) == 0 {
		return nil, errors.New("no metrics to collect")
	}

	metricsToCollect, metricsFromCache := checkCache(mts)

	if len(metricsToCollect) > 0 {
		args := plugin.CollectMetricsArgs{PluginMetricTypes: metricsToCollect}

		out, err := p.encoder.Encode(args)
		if err != nil {
			return nil, err
		}

		var reply []byte
		err = p.connection.Call("Collector.CollectMetrics", out, &reply)
		if err != nil {
			return nil, err
		}

		r := &plugin.CollectMetricsReply{}
		err = p.encoder.Decode(reply, r)
		if err != nil {
			return nil, err
		}

		updateCache(r.PluginMetrics)

		results = make([]core.Metric, len(metricsFromCache)+len(r.PluginMetrics))
		idx := 0
		for _, m := range r.PluginMetrics {
			results[idx] = m
			idx++
		}
		for _, m := range metricsFromCache {
			results[idx] = m
			idx++
		}
		return results, nil
	} else {
		return metricsFromCache, nil
	}

}

func (p *PluginNativeClient) GetMetricTypes(config plugin.PluginConfigType) ([]core.Metric, error) {
	var reply []byte

	args := plugin.GetMetricTypesArgs{PluginConfig: config}

	out, err := p.encoder.Encode(args)
	if err != nil {
		log.Error("error while encoding args for getmetrictypes :(")
		return nil, err
	}

	err = p.connection.Call("Collector.GetMetricTypes", out, &reply)
	if err != nil {
		return nil, err
	}

	r := &plugin.GetMetricTypesReply{}
	err = p.encoder.Decode(reply, r)
	if err != nil {
		return nil, err
	}

	retMetricTypes := make([]core.Metric, len(r.PluginMetricTypes))
	for i, mt := range r.PluginMetricTypes {
		// Set the advertised time
		mt.LastAdvertisedTime_ = time.Now()
		retMetricTypes[i] = mt
	}
	return retMetricTypes, nil
}

func (p *PluginNativeClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	var reply []byte
	err := p.connection.Call("SessionState.GetConfigPolicy", []byte{}, &reply)
	if err != nil {
		return nil, err
	}

	r := &plugin.GetConfigPolicyReply{}
	err = p.encoder.Decode(reply, r)
	if err != nil {
		return nil, err
	}

	return r.Policy, nil
}

// GetType returns the string type of the plugin
// Note: the first letter of the type will be capitalized.
func (p *PluginNativeClient) GetType() string {
	return upcaseInitial(p.pluginType.String())
}

func newNativeClient(address string, timeout time.Duration, t plugin.PluginType, pub *rsa.PublicKey, secure bool) (*PluginNativeClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &PluginNativeClient{
		PluginCacheClient: &pluginCacheClient{},
		connection:        r,
		pluginType:        t,
	}

	p.encoder = encoding.NewGobEncoder()

	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		encrypter := encrypter.New(pub, nil)
		encrypter.Key = key
		p.encrypter = encrypter
		p.encoder.SetEncrypter(encrypter)
	}

	return p, nil
}

func init() {
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))
	gob.Register(*(&ctypes.ConfigValueBool{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cdata.ConfigDataNode{})
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
