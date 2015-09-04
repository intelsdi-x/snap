package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

var logger = log.WithField("_module", "client-httpjsonrpc")

type httpJSONRPCClient struct {
	url     string
	id      uint64
	timeout time.Duration
}

// NewCollectorHttpJSONRPCClient returns CollectorHttpJSONRPCClient
func NewCollectorHttpJSONRPCClient(u string, timeout time.Duration) PluginCollectorClient {
	return &httpJSONRPCClient{
		url:     u,
		timeout: timeout,
	}
}

func NewProcessorHttpJSONRPCClient(u string, timeout time.Duration) PluginProcessorClient {
	return &httpJSONRPCClient{
		url:     u,
		timeout: timeout,
	}
}

func NewPublisherHttpJSONRPCClient(u string, timeout time.Duration) PluginPublisherClient {
	return &httpJSONRPCClient{
		url:     u,
		timeout: timeout,
	}
}

// Ping
func (h *httpJSONRPCClient) Ping() error {
	a := plugin.PingArgs{}
	_, err := h.call("SessionState.Ping", []interface{}{a})
	return err
}

// kill
func (h *httpJSONRPCClient) Kill(reason string) error {
	k := plugin.KillArgs{Reason: reason}
	_, err := h.call("SessionState.Kill", []interface{}{k})
	return err
}

// CollectMetrics returns collected metrics
func (h *httpJSONRPCClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	// Here we create two slices from the requested metric collection. One which
	// contains the metrics we retreived from the cache, and one from which we had
	// to use the plugin.

	// This is managed by walking through the complete list and hitting the cache for each item.
	// If the metric is found in the cache, we nil out that entry in the complete collection.
	// Then, we walk through the collection once more and create a new slice of metrics which
	// were not found in the cache.
	var fromCache []core.Metric
	for i, m := range mts {
		var metric core.Metric
		if metric = metricCache.get(core.JoinNamespace(m.Namespace())); metric != nil {
			fromCache = append(fromCache, metric)
			mts[i] = nil
		}
	}
	var fromPlugin []core.Metric
	for _, mt := range mts {
		if mt != nil {
			fromPlugin = append(fromPlugin, &plugin.PluginMetricType{
				Namespace_: mt.Namespace(),
				Config_:    mt.Config(),
			})
		}
	}
	// We only need to send a request to the plugin if there are metrics which were not available in the cache.
	if len(fromPlugin) > 0 {
		res, err := h.call("Collector.CollectMetrics", []interface{}{fromPlugin})
		if err != nil {
			return nil, err
		}
		var metrics []core.Metric
		if _, ok := res["result"]; !ok {
			err := errors.New("Invalid response: expected the response map to contain the key 'result'.")
			logger.WithFields(log.Fields{
				"_block":           "CollectMetrics",
				"jsonrpc response": fmt.Sprintf("%+v", res),
			}).Error(err)
			return nil, err
		}
		if resmap, ok := res["result"].(map[string]interface{}); ok {
			if _, ok := resmap["PluginMetrics"]; !ok {
				err := errors.New("Invalid response: expected the result value to be a map that contains key 'PluginMetrics'.")
				logger.WithFields(log.Fields{
					"_block":           "CollectMetrics",
					"jsonrpc response": fmt.Sprintf("%+v", res),
				}).Error(err)
				return nil, err
			}
			if pms, ok := resmap["PluginMetrics"].([]interface{}); ok {
				for _, m := range pms {
					j, err := json.Marshal(m)
					if err != nil {
						return nil, err
					}
					pmt := &plugin.PluginMetricType{}
					if err := json.Unmarshal(j, &pmt); err != nil {
						return nil, err
					}
					metrics = append(metrics, pmt)
				}
			} else {
				err := errors.New("Invalid response: expected 'PluginMetrics' to contain a list of metrics")
				logger.WithFields(log.Fields{
					"_block":           "CollectMetrics",
					"jsonrpc response": fmt.Sprintf("%+v", res),
				}).Error(err)
				return nil, err
			}
		} else {
			err := errors.New("Invalid response: expected 'result' to be a map")
			logger.WithFields(log.Fields{
				"_block":           "CollectMetrics",
				"jsonrpc response": fmt.Sprintf("%+v", res),
			}).Error(err)
			return nil, err
		}
		for _, m := range metrics {
			metricCache.put(core.JoinNamespace(m.Namespace()), m)
		}
		metrics = append(metrics, fromCache...)
		return metrics, err
	}
	return fromCache, nil
}

// GetMetricTypes returns metric types that can be collected
func (h *httpJSONRPCClient) GetMetricTypes() ([]core.Metric, error) {
	res, err := h.call("Collector.GetMetricTypes", []interface{}{})
	if err != nil {
		return nil, err
	}
	var metrics []core.Metric
	for _, m := range res["result"].(map[string]interface{})["PluginMetricTypes"].([]interface{}) {
		j, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		pmt := &plugin.PluginMetricType{}
		if err := json.Unmarshal(j, &pmt); err != nil {
			return nil, err
		}
		metrics = append(metrics, pmt)
	}
	return metrics, nil
}

// GetConfigPolicyTree returns a config policy tree
func (h *httpJSONRPCClient) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	res, err := h.call("Collector.GetConfigPolicyTree", []interface{}{})
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "GetConfigPolicyTree",
			"result": fmt.Sprintf("%+v", res),
			"error":  err,
		}).Error("error getting config policy tree")
		return cpolicy.ConfigPolicyTree{}, err
	}
	bres, err := json.Marshal(res["result"].(map[string]interface{}))
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "GetConfigPolicyTree",
			"result": fmt.Sprintf("%+v", res),
			"error":  err,
		}).Error("error marshalling result into json")
		return cpolicy.ConfigPolicyTree{}, err
	}
	cpt := cpolicy.NewTree()
	if err := json.Unmarshal(bres, cpt); err != nil {
		logger.WithFields(log.Fields{
			"_block": "GetConfigPolicyTree",
			"result": string(bres),
			"error":  err,
		}).Error("error unmarshalling result into cpolicy tree")
		return cpolicy.ConfigPolicyTree{}, err
	}
	return *cpt, nil
}

func (h *httpJSONRPCClient) GetConfigPolicyNode() (cpolicy.ConfigPolicyNode, error) {
	res, err := h.call("Processor.GetConfigPolicyNode", []interface{}{})
	if err != nil {
		return cpolicy.ConfigPolicyNode{}, err
	}
	bres, err := json.Marshal(res["result"].(map[string]interface{}))
	if err != nil {
		return cpolicy.ConfigPolicyNode{}, err
	}
	cpn := cpolicy.NewPolicyNode()
	if err := json.Unmarshal(bres, cpn); err != nil {
		return cpolicy.ConfigPolicyNode{}, err
	}
	return *cpn, nil
}

func (h *httpJSONRPCClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	publisherArgs := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}
	_, err := h.call("Publisher.Publish", []interface{}{publisherArgs})
	if err != nil {
		return err
	}
	return nil
}

func (h *httpJSONRPCClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	processorArgs := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}
	res, err := h.call("Processor.Process", []interface{}{processorArgs})
	if err != nil {
		return "", nil, err
	}
	bres, err := json.Marshal(res["result"].(map[string]interface{}))
	if err != nil {
		return "", nil, err
	}
	processorReply := &plugin.ProcessorReply{}
	if err := json.Unmarshal(bres, processorReply); err != nil {
		return "", nil, err
	}
	return processorReply.ContentType, processorReply.Content, nil
}

func (h *httpJSONRPCClient) call(method string, args []interface{}) (map[string]interface{}, error) {
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
	resp, err := client.Post(h.url,
		"application/json", strings.NewReader(string(data)))
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
	result := make(map[string]interface{})
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	// log.Debugf("result: %v \n", result)
	atomic.AddUint64(&h.id, 1)
	return result, nil
}
