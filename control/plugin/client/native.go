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
	"bytes"
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"fmt"
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
	connection CallsRPC
	pluginType plugin.PluginType
	encoder    encoding.Encoder
	encrypter  *encrypter.Encrypter
	timeout    time.Duration
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

func (p *PluginNativeClient) Close() error {
	// Added to conform to interface, but not needed by native
	return nil
}

// Used to catch zero values for times and overwrite with current time
// the 0 value for time.Time is year 1 which isn't a valid value for metric
// collection (until we get a time machine).
func checkTime(in time.Time) time.Time {
	if in.Year() < 1970 {
		return time.Now()
	}
	return in
}

func encodeMetrics(metrics []core.Metric) []byte {
	mts := make([]plugin.MetricType, len(metrics))
	for i, m := range metrics {
		mts[i] = plugin.MetricType{
			Namespace_:          m.Namespace(),
			Tags_:               m.Tags(),
			Timestamp_:          checkTime(m.Timestamp()),
			Version_:            m.Version(),
			Config_:             m.Config(),
			LastAdvertisedTime_: checkTime(m.LastAdvertisedTime()),
			Unit_:               m.Unit(),
			Description_:        m.Description(),
			Data_:               m.Data(),
		}
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(mts)
	return buf.Bytes()
}

func decodeMetrics(bts []byte) ([]core.Metric, error) {
	var mts []plugin.MetricType
	dec := gob.NewDecoder(bytes.NewBuffer(bts))
	if err := dec.Decode(&mts); err != nil {
		return nil, fmt.Errorf("Error decoding metrics: %v", err)
	}
	var cmetrics []core.Metric
	for _, mt := range mts {
		mt.Timestamp_ = checkTime(mt.Timestamp())
		mt.LastAdvertisedTime_ = checkTime(mt.LastAdvertisedTime())
		cmetrics = append(cmetrics, mt)
	}
	return cmetrics, nil
}

func enforceTimeout(p *PluginNativeClient, dl time.Duration, done chan int) {
	select {
	case <-time.After(dl):
		p.Kill("Passed deadline")
	case <-done:
		return
	}
}

func (p *PluginNativeClient) Publish(metrics []core.Metric, config map[string]ctypes.ConfigValue) error {

	args := plugin.PublishArgs{
		ContentType: plugin.SnapGOBContentType,
		Content:     encodeMetrics(metrics),
		Config:      config,
	}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return err
	}
	var reply []byte
	done := make(chan int)
	go enforceTimeout(p, p.timeout, done)
	err = p.connection.Call("Publisher.Publish", out, &reply)
	close(done)
	return err
}

func (p *PluginNativeClient) Process(metrics []core.Metric, config map[string]ctypes.ConfigValue) ([]core.Metric, error) {

	args := plugin.ProcessorArgs{
		ContentType: plugin.SnapGOBContentType,
		Content:     encodeMetrics(metrics),
		Config:      config,
	}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return nil, err
	}

	var reply []byte
	done := make(chan int)
	go enforceTimeout(p, p.timeout, done)
	err = p.connection.Call("Processor.Process", out, &reply)
	close(done)
	if err != nil {
		return nil, err
	}

	r := plugin.ProcessorReply{}
	err = p.encoder.Decode(reply, &r)
	if err != nil {
		return nil, err
	}
	mts, err := decodeMetrics(r.Content)
	if err != nil {
		return nil, err
	}
	return mts, nil

}

func (p *PluginNativeClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	// Convert core.MetricType slice into plugin.nMetricType slice as we have
	// to send structs over RPC
	var results []core.Metric
	if len(mts) == 0 {
		return nil, errors.New("no metrics to collect")
	}

	metricsToCollect := make([]plugin.MetricType, len(mts))
	for idx, mt := range mts {
		metricsToCollect[idx] = plugin.MetricType{
			Namespace_:          mt.Namespace(),
			LastAdvertisedTime_: mt.LastAdvertisedTime(),
			Version_:            mt.Version(),
			Tags_:               mt.Tags(),
			Config_:             mt.Config(),
			Unit_:               mt.Unit(),
		}
	}

	args := plugin.CollectMetricsArgs{MetricTypes: metricsToCollect}
	out, err := p.encoder.Encode(args)
	if err != nil {
		return nil, err
	}

	var reply []byte
	done := make(chan int)
	go enforceTimeout(p, p.timeout, done)
	err = p.connection.Call("Collector.CollectMetrics", out, &reply)
	close(done)
	if err != nil {
		return nil, err
	}

	r := &plugin.CollectMetricsReply{}
	err = p.encoder.Decode(reply, r)
	if err != nil {
		return nil, err
	}

	results = make([]core.Metric, len(r.PluginMetrics))
	idx := 0
	for _, m := range r.PluginMetrics {
		results[idx] = m
		idx++
	}

	return results, nil
}

func (p *PluginNativeClient) GetMetricTypes(config plugin.ConfigType) ([]core.Metric, error) {
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

	retMetricTypes := make([]core.Metric, len(r.MetricTypes))
	for i, mt := range r.MetricTypes {
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
		connection: r,
		pluginType: t,
		timeout:    timeout,
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
	gob.RegisterName("conf_value_string", *(&ctypes.ConfigValueStr{}))
	gob.RegisterName("conf_value_int", *(&ctypes.ConfigValueInt{}))
	gob.RegisterName("conf_value_float", *(&ctypes.ConfigValueFloat{}))
	gob.RegisterName("conf_value_bool", *(&ctypes.ConfigValueBool{}))

	gob.RegisterName("conf_policy_node", cpolicy.NewPolicyNode())
	gob.RegisterName("conf_data_node", &cdata.ConfigDataNode{})
	gob.RegisterName("conf_policy_string", &cpolicy.StringRule{})
	gob.RegisterName("conf_policy_int", &cpolicy.IntRule{})
	gob.RegisterName("conf_policy_float", &cpolicy.FloatRule{})
	gob.RegisterName("conf_policy_bool", &cpolicy.BoolRule{})
}

func upcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
