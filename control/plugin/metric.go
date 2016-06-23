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

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
)

const (
	// These are our built-in content types for plugins

	// SnapAll the wildcard for accepting all snap content types
	SnapAllContentType = "snap.*"
	// SnapGOB snap metrics serialized into go binary format
	SnapGOBContentType = "snap.gob"
	// SnapJSON snap metrics serialized into json
	SnapJSONContentType = "snap.json"
	// SnapProtoBuff snap metrics serialized into protocol buffers
	// SnapProtoBuff = "snap.pb" // TO BE IMPLEMENTED
)

type ConfigType struct {
	*cdata.ConfigDataNode
}

func (p *ConfigType) UnmarshalJSON(data []byte) error {
	cdn := cdata.NewNode()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(cdn); err != nil {
		return err
	}
	p.ConfigDataNode = cdn
	return nil
}

func (p ConfigType) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(p.ConfigDataNode); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (p *ConfigType) GobDecode(data []byte) error {
	cdn := cdata.NewNode()
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(cdn); err != nil {
		return err
	}
	p.ConfigDataNode = cdn

	return nil
}

func NewPluginConfigType() ConfigType {
	return ConfigType{
		ConfigDataNode: cdata.NewNode(),
	}
}

// Represents a metric type. Only used within plugins and across plugin calls.
// Converted to core.MetricType before being used within modules.
type MetricType struct {
	// Namespace is the identifier for a metric.
	Namespace_ []core.NamespaceElement `json:"namespace"`

	// Last advertised time is the last time the snap agent was told about
	// a metric.
	LastAdvertisedTime_ time.Time `json:"last_advertised_time"`

	// The metric version. It is bound to the Plugin version.
	Version_ int `json:"version"`

	// The config data needed to collect a metric.
	Config_ *cdata.ConfigDataNode `json:"config"`

	Data_ interface{} `json:"data"`

	// Tags are key value pairs that can be added by the framework or any
	// plugin along the collect -> process -> publish pipeline.
	Tags_ map[string]string `json:"tags"`

	// Unit represents the unit of magnitude of the measured quantity.
	// See http://metrics20.org/spec/#units as a guideline for this
	// field.
	Unit_ string

	// A (long) description for the metric.  The description is stored on the
	// metric catalog and not sent through  collect -> process -> publish.
	Description_ string `json:"description"`

	// The timestamp from when the metric was created.
	Timestamp_ time.Time `json:"timestamp"`
}

// NewMetricType returns a Constructor
func NewMetricType(namespace core.Namespace, timestamp time.Time, tags map[string]string, unit string, data interface{}) *MetricType {
	return &MetricType{
		Namespace_:          namespace,
		Tags_:               tags,
		Data_:               data,
		Timestamp_:          timestamp,
		LastAdvertisedTime_: timestamp,
		Unit_:               unit,
	}
}

// Returns the namespace.
func (p MetricType) Namespace() core.Namespace {
	return p.Namespace_
}

// Returns the last time this metric type was received from the plugin.
func (p MetricType) LastAdvertisedTime() time.Time {
	return p.LastAdvertisedTime_
}

// Returns the namespace.
func (p MetricType) Version() int {
	return p.Version_
}

// Config returns the map of config data for this metric
func (p MetricType) Config() *cdata.ConfigDataNode {
	return p.Config_
}

// Tags returns the map of  tags for this metric
func (p MetricType) Tags() map[string]string {
	return p.Tags_
}

// returns the timestamp of when the metric was collected
func (p MetricType) Timestamp() time.Time {
	return p.Timestamp_
}

// returns the data for the metric
func (p MetricType) Data() interface{} {
	return p.Data_
}

// returns the description of the metric
func (p MetricType) Description() string {
	return p.Description_
}

// returns the metrics unit
func (p MetricType) Unit() string {
	return p.Unit_
}

func (p *MetricType) AddData(data interface{}) {
	p.Data_ = data
}

// MarshalMetricTypes returns a []byte containing a serialized version of []MetricType using the content type provided.
func MarshalMetricTypes(contentType string, metrics []MetricType) ([]byte, string, error) {
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
	case SnapAllContentType, SnapGOBContentType:
		// NOTE: A snap All wildcard will result in GOB
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
		// contentType := SnapGOBContentType
		return buf.Bytes(), SnapGOBContentType, nil
	case SnapJSONContentType:
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
		return b, SnapJSONContentType, nil
	default:
		// We don't recognize this content type. Log and return error.
		es := fmt.Sprintf("invalid snap content type: %s", contentType)
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "marshal-content-type",
			"error":   es,
		}).Error("error while marshalling")
		return nil, "", errors.New(es)
	}
}

// UnmarshallMetricTypes takes a content type and []byte payload and returns a []MetricType
func UnmarshallMetricTypes(contentType string, payload []byte) ([]MetricType, error) {
	switch contentType {
	case SnapGOBContentType:
		var metrics []MetricType
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
	case SnapJSONContentType:
		var metrics []MetricType
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
		es := fmt.Sprintf("invalid snap content type for unmarshalling: %s", contentType)
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "unmarshal-content-type",
			"error":   es,
		}).Error("error while unmarshalling")
		return nil, errors.New(es)
	}
}

// SwapMetricContentType swaps a payload with one content type to another one.
func SwapMetricContentType(contentType, requestedContentType string, payload []byte) ([]byte, string, error) {
	metrics, err1 := UnmarshallMetricTypes(contentType, payload)
	if err1 != nil {
		log.WithFields(log.Fields{
			"_module": "control-plugin",
			"block":   "swap-content-type",
			"error":   err1.Error(),
		}).Error("error while swaping")
		return nil, "", err1
	}
	newPayload, newContentType, err2 := MarshalMetricTypes(requestedContentType, metrics)
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
