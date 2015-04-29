package plugin

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/pkg/logger"
)

const (
	// These are our built-in content types for plugins

	// PulseAll the wildcard for accepting all pulse content types
	PulseAllContentType = "pulse.*"
	// PulseGOB pulse metrics serialized into go binary format
	PulseGOBContentType = "pulse.gob"
	// PulseJSON pulse metrics serialized into json
	PulseJSONContentType = "pulse.json"
	// PulseProtoBuff pulse metrics serialized into protocol buffers
	// PulseProtoBuff = "pulse.pb" // TO BE IMPLEMENTED
)

// Represents a metric type. Only used within plugins and across plugin calls.
// Converted to core.MetricType before being used within modules.
type PluginMetricType struct {
	Namespace_          []string              `json:"namespace"`
	LastAdvertisedTime_ time.Time             `json:"last_advertised_time"`
	Version_            int                   `json:"version"`
	Config_             *cdata.ConfigDataNode `json:"config"`
	Data_               interface{}           `json:"data"`
}

// // PluginMetricType Constructor
func NewPluginMetricType(namespace []string, data interface{}) *PluginMetricType {
	return &PluginMetricType{
		Namespace_: namespace,
		Data_:      data,
	}
}

// Returns the namespace.
func (p PluginMetricType) Namespace() []string {
	return p.Namespace_
}

// Returns the last time this metric type was received from the plugin.
func (p PluginMetricType) LastAdvertisedTime() time.Time {
	return p.LastAdvertisedTime_
}

// Returns the namespace.
func (p PluginMetricType) Version() int {
	return p.Version_
}

// Config returns the map of config data for this metric
func (p PluginMetricType) Config() *cdata.ConfigDataNode {
	return p.Config_
}

func (p PluginMetricType) Data() interface{} {
	return p.Data_
}

func (p *PluginMetricType) AddData(data interface{}) {
	p.Data_ = data
}

// MarshallMetricTypes returns a []byte containing a serialized version of []PluginMetricType using the content type provided.
func MarshallMetricTypes(contentType string, metrics []PluginMetricType) ([]byte, string, error) {
	if len(metrics) == 0 {
		es := fmt.Sprintf("attempt to marshall empty slice of metrics: %s", contentType)
		logger.Errorf("marshal-metric-types", es)
		return nil, "", errors.New(es)
	}

	switch contentType {
	case PulseAllContentType, PulseGOBContentType:
		// NOTE: A Pulse All wildcard will result in GOB
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(metrics)
		if err != nil {
			logger.Errorf("marshal-metric-types", err.Error())
			return nil, "", err
		}
		// contentType := PulseGOBContentType
		return buf.Bytes(), PulseGOBContentType, nil
	case PulseJSONContentType:
		// Serialize into JSON
		b, err := json.Marshal(metrics)
		if err != nil {
			logger.Errorf("marshal-metric-types", err.Error())
			return nil, "", err
		}
		return b, PulseJSONContentType, nil
	default:
		// We don't recognize this content type. Log and return error.
		es := fmt.Sprintf("invlaid pulse content type: %s", contentType)
		logger.Errorf("marshal-metric-types", es)
		return nil, "", errors.New(es)
	}

	// remove
	return nil, "", nil
}

/*

core.Metric(i) (used by pulse modules)
core.MetricType(i) (used by pulse modules)

PluginMetric (created by plugins and returned, goes over RPC)
PLuginMetricType (created by plugins and returned, goes over RPC)

Get

agent - call Get
client - call Get
plugin - builds array of PluginMetricTypes
plugin - return array of PluginMetricTypes
client - returns array of PluginMetricTypes through MetricType interface
agent - receives MetricTypes

Collect

agent - call Collect pass MetricTypes
client - call Collect, convert MetricTypes into new (plugin) PluginMetricTypes struct
plugin - build array of PluginMetric based on (plugin) MetricTypes
plugin - return array of PluginMetrics
client - return array of PluginMetrics through core.Metrics interface
agent - receive array of core.Metrics


*/
