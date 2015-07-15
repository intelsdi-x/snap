package client

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
)

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
	res, err := h.call("Collector.CollectMetrics", []interface{}{mts})
	if err != nil {
		return nil, err
	}
	var metrics []core.Metric
	for _, m := range res["result"].(map[string]interface{})["PluginMetrics"].([]interface{}) {
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
	return metrics, err
}

// GetMetricTypes returns metric types that can be collected
func (h *httpJSONRPCClient) GetMetricTypes() ([]core.Metric, error) {
	res, _ := h.call("Collector.GetMetricTypes", []interface{}{})
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
	res, _ := h.call("Collector.GetConfigPolicyTree", []interface{}{})
	log.Debugf("!!! GetConfigPolicyTree result: %v \n", res)
	bres, err := json.Marshal(res["result"].(map[string]interface{}))
	if err != nil {
		return cpolicy.ConfigPolicyTree{}, err
	}
	cpt := cpolicy.NewTree()
	if err := json.Unmarshal(bres, cpt); err != nil {
		return cpolicy.ConfigPolicyTree{}, err
	}
	return *cpt, nil
}

func (h *httpJSONRPCClient) call(method string, args []interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     h.id,
		"params": args,
	})
	if err != nil {
		log.Debugf("Marshal: %v", err)
		return nil, err
	}
	client := http.Client{Timeout: h.timeout}
	resp, err := client.Post(h.url,
		"application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Debugf("Post: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	result := make(map[string]interface{})
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Debugf("result: %v\n", result)
		log.Debugf("Unmarshal: %v\n", err)
		return nil, err
	}
	//log.Debugf("result: %v \n", result)
	atomic.AddUint64(&h.id, 1)
	return result, nil
}
