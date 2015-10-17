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

package plugin

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/core/cdata"
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

type PluginConfigType struct {
	*cdata.ConfigDataNode
}

func (p *PluginConfigType) UnmarshalJSON(data []byte) error {
	cdn := cdata.NewNode()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(cdn); err != nil {
		return err
	}
	p.ConfigDataNode = cdn
	return nil
}

func (p PluginConfigType) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(p.ConfigDataNode); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (p *PluginConfigType) GobDecode(data []byte) error {
	cdn := cdata.NewNode()
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(cdn); err != nil {
		return err
	}
	p.ConfigDataNode = cdn

	return nil
}

func NewPluginConfigType() PluginConfigType {
	return PluginConfigType{
		ConfigDataNode: cdata.NewNode(),
	}
}

// Represents a metric type. Only used within plugins and across plugin calls.
// Converted to core.MetricType before being used within modules.
type PluginMetricType struct {
	// Namespace is the identifier for a metric.
	Namespace_ []string `json:"namespace"`

	// Last advertised time is the last time the Pulse agent was told about
	// a metric.
	LastAdvertisedTime_ time.Time `json:"last_advertised_time"`

	// The metric version. It is bound to the Plugin version.
	Version_ int `json:"version"`

	// The config data needed to collect a metric.
	Config_ *cdata.ConfigDataNode `json:"config"`

	Data_ interface{} `json:"data"`

	// The source of the metric (host, IP, etc).
	Source_ string `json:"source"`

	// The timestamp from when the metric was created.
	Timestamp_ time.Time `json:"timestamp"`
}

// // PluginMetricType Constructor
func NewPluginMetricType(namespace []string, timestamp time.Time, source string, data interface{}) *PluginMetricType {
	return &PluginMetricType{
		Namespace_: namespace,
		Data_:      data,
		Timestamp_: timestamp,
		Source_:    source,
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

// returns the timestamp of when the metric was collected
func (p PluginMetricType) Timestamp() time.Time {
	return p.Timestamp_
}

// returns the source of the metric
func (p PluginMetricType) Source() string {
	return p.Source_
}

func (p PluginMetricType) Data() interface{} {
	return p.Data_
}

func (p *PluginMetricType) AddData(data interface{}) {
	p.Data_ = data
}

// MarshalMetricTypes returns a []byte containing a serialized version of []PluginMetricType using the content type provided.
func MarshalPluginMetricTypes(contentType string, metrics []PluginMetricType) ([]byte, string, error) {
	// If we have an empty slice we return an error
	if len(metrics) == 0 {
		es := fmt.Sprintf("attempt to marshall empty slice of metrics: %s", contentType)
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "marshal-content-type",
			"error":   es,
		}).Error("error while marshalling")
		return nil, "", errors.New(es)
	}
	// Switch on content type
	switch contentType {
	case PulseAllContentType, PulseGOBContentType:
		// NOTE: A Pulse All wildcard will result in GOB
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(metrics)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control-plugin",
				"block":   "marshal-content-type",
				"error":   err.Error(),
			}).Error("error while marshalling")
			return nil, "", err
		}
		// contentType := PulseGOBContentType
		return buf.Bytes(), PulseGOBContentType, nil
	case PulseJSONContentType:
		// Serialize into JSON
		b, err := json.Marshal(metrics)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control-plugin",
				"block":   "marshal-content-type",
				"error":   err.Error(),
			}).Error("error while marshalling")
			return nil, "", err
		}
		return b, PulseJSONContentType, nil
	default:
		// We don't recognize this content type. Log and return error.
		es := fmt.Sprintf("invalid pulse content type: %s", contentType)
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "marshal-content-type",
			"error":   es,
		}).Error("error while marshalling")
		return nil, "", errors.New(es)
	}
}

// UnmarshallPluginMetricTypes takes a content type and []byte payload and returns a []PluginMetricType
func UnmarshallPluginMetricTypes(contentType string, payload []byte) ([]PluginMetricType, error) {
	switch contentType {
	case PulseGOBContentType:
		var metrics []PluginMetricType
		r := bytes.NewBuffer(payload)
		err := gob.NewDecoder(r).Decode(&metrics)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control-plugin",
				"block":   "unmarshal-content-type",
				"error":   err.Error(),
			}).Error("error while unmarshalling")
			return nil, err
		}
		return metrics, nil
	case PulseJSONContentType:
		var metrics []PluginMetricType
		err := json.Unmarshal(payload, &metrics)
		if err != nil {
			log.WithFields(log.Fields{
				"_module": "control-plugin",
				"block":   "unmarshal-content-type",
				"error":   err.Error(),
			}).Error("error while unmarshalling")
			return nil, err
		}
		return metrics, nil
	default:
		// We don't recognize this content type as one we can unmarshal. Log and return error.
		es := fmt.Sprintf("invalid pulse content type for unmarshalling: %s", contentType)
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "unmarshal-content-type",
			"error":   es,
		}).Error("error while unmarshalling")
		return nil, errors.New(es)
	}
}

// SwapPluginMetricContentType swaps a payload with one content type to another one.
func SwapPluginMetricContentType(contentType, requestedContentType string, payload []byte) ([]byte, string, error) {
	metrics, err1 := UnmarshallPluginMetricTypes(contentType, payload)
	if err1 != nil {
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "swap-content-type",
			"error":   err1.Error(),
		}).Error("error while swaping")
		return nil, "", err1
	}
	newPayload, newContentType, err2 := MarshalPluginMetricTypes(requestedContentType, metrics)
	if err2 != nil {
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "swap-content-type",
			"error":   err2.Error(),
		}).Error("error while swaping")
		return nil, "", err2
	}
	return newPayload, newContentType, nil
}
