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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

var logger = log.WithField("_module", "client-httpjsonrpc")

type httpJSONRPCClient struct {
	url        string
	id         uint64
	timeout    time.Duration
	pluginType plugin.PluginType
	encrypter  *encrypter.Encrypter
	encoder    encoding.Encoder
}

// NewCollectorHttpJSONRPCClient returns CollectorHttpJSONRPCClient
func NewCollectorHttpJSONRPCClient(u string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginCollectorClient, error) {
	hjr := &httpJSONRPCClient{
		url:        u,
		timeout:    timeout,
		pluginType: plugin.CollectorPluginType,
		encoder:    encoding.NewJsonEncoder(),
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		e := encrypter.New(pub, nil)
		e.Key = key
		hjr.encoder.SetEncrypter(e)
		hjr.encrypter = e
	}
	return hjr, nil
}

func NewProcessorHttpJSONRPCClient(u string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginProcessorClient, error) {
	hjr := &httpJSONRPCClient{
		url:        u,
		timeout:    timeout,
		pluginType: plugin.ProcessorPluginType,
		encoder:    encoding.NewJsonEncoder(),
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		e := encrypter.New(pub, nil)
		e.Key = key
		hjr.encoder.SetEncrypter(e)
		hjr.encrypter = e
	}
	return hjr, nil
}

func NewPublisherHttpJSONRPCClient(u string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginPublisherClient, error) {
	hjr := &httpJSONRPCClient{
		url:        u,
		timeout:    timeout,
		pluginType: plugin.PublisherPluginType,
		encoder:    encoding.NewJsonEncoder(),
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		e := encrypter.New(pub, nil)
		e.Key = key
		hjr.encoder.SetEncrypter(e)
		hjr.encrypter = e
	}
	return hjr, nil
}

// Ping
func (h *httpJSONRPCClient) Ping() error {
	_, err := h.call("SessionState.Ping", []interface{}{})
	return err
}

func (h *httpJSONRPCClient) SetKey() error {
	key, err := h.encrypter.EncryptKey()
	if err != nil {
		return err
	}
	a := plugin.SetKeyArgs{Key: key}
	_, err = h.call("SessionState.SetKey", []interface{}{a})
	return err
}

// kill
func (h *httpJSONRPCClient) Kill(reason string) error {
	args := plugin.KillArgs{Reason: reason}
	out, err := h.encoder.Encode(args)
	if err != nil {
		return err
	}

	_, err = h.call("SessionState.Kill", []interface{}{out})
	return err
}

// CollectMetrics returns collected metrics
func (h *httpJSONRPCClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	var results []core.Metric
	if len(mts) == 0 {
		return nil, errors.New("no metrics to collect")
	}

	metricsToCollect := make([]plugin.PluginMetricType, len(mts))
	for idx, mt := range mts {
		metricsToCollect[idx] = plugin.PluginMetricType{
			Namespace_:          mt.Namespace(),
			LastAdvertisedTime_: mt.LastAdvertisedTime(),
			Version_:            mt.Version(),
			Tags_:               mt.Tags(),
			Labels_:             mt.Labels(),
			Config_:             mt.Config(),
		}
	}

	args := &plugin.CollectMetricsArgs{PluginMetricTypes: metricsToCollect}

	out, err := h.encoder.Encode(args)
	if err != nil {
		return nil, err
	}

	res, err := h.call("Collector.CollectMetrics", []interface{}{out})
	if err != nil {
		return nil, err
	}
	if len(res.Result) == 0 {
		err := errors.New("Invalid response: result is 0")
		logger.WithFields(log.Fields{
			"_block":           "CollectMetrics",
			"jsonrpc response": fmt.Sprintf("%+v", res),
		}).Error(err)
		return nil, err
	}
	r := &plugin.CollectMetricsReply{}
	err = h.encoder.Decode(res.Result, r)
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

// GetMetricTypes returns metric types that can be collected
func (h *httpJSONRPCClient) GetMetricTypes(config plugin.PluginConfigType) ([]core.Metric, error) {
	args := plugin.GetMetricTypesArgs{PluginConfig: config}

	out, err := h.encoder.Encode(args)
	if err != nil {
		return nil, err
	}

	res, err := h.call("Collector.GetMetricTypes", []interface{}{out})
	if err != nil {
		return nil, err
	}
	var mtr plugin.GetMetricTypesReply
	err = h.encoder.Decode(res.Result, &mtr)
	if err != nil {
		return nil, err
	}
	metrics := make([]core.Metric, len(mtr.PluginMetricTypes))
	for i, mt := range mtr.PluginMetricTypes {
		mt.LastAdvertisedTime_ = time.Now()
		metrics[i] = mt
	}
	return metrics, nil
}

// GetConfigPolicy returns a config policy
func (h *httpJSONRPCClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	res, err := h.call("SessionState.GetConfigPolicy", []interface{}{})
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "GetConfigPolicy",
			"result": fmt.Sprintf("%+v", res),
			"error":  err,
		}).Error("error getting config policy")
		return nil, err
	}
	if len(res.Result) == 0 {
		return nil, errors.New(res.Error)
	}
	var cpr plugin.GetConfigPolicyReply
	err = h.encoder.Decode(res.Result, &cpr)
	if err != nil {
		return nil, err
	}
	return cpr.Policy, nil
}

func (h *httpJSONRPCClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	args := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}
	out, err := h.encoder.Encode(args)
	if err != nil {
		return nil
	}
	_, err = h.call("Publisher.Publish", []interface{}{out})
	if err != nil {
		return err
	}
	return nil
}

func (h *httpJSONRPCClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	args := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}
	out, err := h.encoder.Encode(args)
	if err != nil {
		return "", nil, err
	}
	res, err := h.call("Processor.Process", []interface{}{out})
	if err != nil {
		return "", nil, err
	}
	processorReply := &plugin.ProcessorReply{}
	if err := h.encoder.Decode(res.Result, processorReply); err != nil {
		return "", nil, err
	}
	return processorReply.ContentType, processorReply.Content, nil
}

func (h *httpJSONRPCClient) GetType() string {
	return upcaseInitial(h.pluginType.String())
}

type jsonRpcResp struct {
	Id     int    `json:"id"`
	Result []byte `json:"result"`
	Error  string `json:"error"`
}

func (h *httpJSONRPCClient) call(method string, args []interface{}) (*jsonRpcResp, error) {
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     h.id,
		"params": args,
	})
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "call",
			"url":    h.url,
			"args":   fmt.Sprintf("%+v", args),
			"method": method,
			"id":     h.id,
			"error":  err,
		}).Error("error encoding request to json")
		return nil, err
	}
	client := http.Client{Timeout: h.timeout}
	resp, err := client.Post(h.url, "application/json", bytes.NewReader(data))
	if err != nil {
		logger.WithFields(log.Fields{
			"_block":  "call",
			"url":     h.url,
			"request": string(data),
			"error":   err,
		}).Error("error posting request to plugin")
		return nil, err
	}
	defer resp.Body.Close()
	result := &jsonRpcResp{}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		bs, _ := ioutil.ReadAll(resp.Body)
		logger.WithFields(log.Fields{
			"_block":      "call",
			"url":         h.url,
			"request":     string(data),
			"status code": resp.StatusCode,
			"response":    string(bs),
			"error":       err,
		}).Error("error decoding result")
		return nil, err
	}
	atomic.AddUint64(&h.id, 1)
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}
