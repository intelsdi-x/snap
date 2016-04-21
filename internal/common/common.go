/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package common

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
)

func ToMetric(co core.Metric) *Metric {
	cm := &Metric{
		Namespace: co.Namespace(),
		Version:   int64(co.Version()),
		Source:    co.Source(),
		Tags:      co.Tags(),
		Timestamp: &Time{
			Sec:  co.Timestamp().Unix(),
			Nsec: int64(co.Timestamp().Nanosecond()),
		},
		LastAdvertisedTime: &Time{
			Sec:  co.LastAdvertisedTime().Unix(),
			Nsec: int64(co.Timestamp().Nanosecond()),
		},
	}
	if co.Config() != nil {
		cm.Config = ConfigToConfigMap(co.Config())
	}
	cm.Labels = make([]*Label, len(co.Labels()))
	for y, label := range co.Labels() {
		cm.Labels[y] = &Label{
			Index: uint64(label.Index),
			Name:  label.Name,
		}
	}
	cm.Data, cm.DataType = encodeData(co)
	return cm
}

func encodeData(mt core.Metric) ([]byte, string) {

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	var Data []byte
	var DataType string
	switch t := mt.Data().(type) {
	case string:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "string"
	case float64:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "float64"
	case float32:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "float32"
	case int32:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "int32"
	case int:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "int"
	case int64:
		enc.Encode(t)
		Data = b.Bytes()
		DataType = "int64"
	case nil:
		Data = nil
		DataType = "nil"
	default:
		panic(t)
	}
	return Data, DataType
}

func decodeData(b []byte, t string) interface{} {
	var Data interface{}
	switch t {
	case "int":
		var val int
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	case "int32":
		var val int32
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	case "int64":
		var val int64
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	case "float32":
		var val float32
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	case "float64":
		var val float64
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	case "string":
		var val string
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		decoder.Decode(&val)
		Data = val
	}
	return Data
}

func NewMetrics(ms []core.Metric) []*Metric {
	metrics := make([]*Metric, len(ms))
	for i, m := range ms {
		metrics[i] = ToMetric(m)
	}
	return metrics
}

func ToCoreMetric(mt *Metric) core.Metric {
	ret := plugin.PluginMetricType{
		Namespace_:          mt.Namespace,
		Version_:            int(mt.Version),
		Source_:             mt.Source,
		Tags_:               mt.Tags,
		Timestamp_:          time.Unix(mt.Timestamp.Sec, mt.Timestamp.Nsec),
		LastAdvertisedTime_: time.Unix(mt.LastAdvertisedTime.Sec, mt.LastAdvertisedTime.Nsec),
	}
	ret.Config_ = ConfigMapToConfig(mt.Config)
	ret.Data_ = decodeData(mt.Data, mt.DataType)

	return ret
}

func ToCoreMetrics(mts []*Metric) []core.Metric {
	metrics := make([]core.Metric, len(mts))
	for i, mt := range mts {
		metrics[i] = ToCoreMetric(mt)
	}
	return metrics
}

// implements core.SubscribedPlugin
type SubPlugin struct {
	typeName string
	name     string
	version  int
	config   *cdata.ConfigDataNode
}

func (sp SubPlugin) TypeName() string {
	return sp.typeName
}

func (sp SubPlugin) Name() string {
	return sp.name
}

func (sp SubPlugin) Version() int {
	return sp.version
}

func (sp SubPlugin) Config() *cdata.ConfigDataNode {
	return sp.config
}

func ToCorePlugin(pl core.SubscribedPlugin) core.Plugin {
	return core.Plugin(pl)
}

func ToCorePlugins(pl []core.SubscribedPlugin) []core.Plugin {
	plugins := make([]core.Plugin, len(pl))
	for i, v := range pl {
		plugins[i] = v
	}
	return plugins
}

func ToSubPluginMsg(pl core.SubscribedPlugin) *SubscribedPlugin {
	return &SubscribedPlugin{
		TypeName: pl.TypeName(),
		Name:     pl.Name(),
		Version:  int64(pl.Version()),
		Config:   ConfigToConfigMap(pl.Config()),
	}
}

func ToSubPlugin(msg *SubscribedPlugin) core.SubscribedPlugin {
	return SubPlugin{
		typeName: msg.TypeName,
		name:     msg.Name,
		version:  int(msg.Version),
		config:   ConfigMapToConfig(msg.Config),
	}
}

func ToCorePluginMsg(pl core.Plugin) *Plugin {
	return &Plugin{
		TypeName: pl.TypeName(),
		Name:     pl.Name(),
		Version:  int64(pl.Version()),
	}
}

func ToCorePluginsMsg(pls []core.Plugin) []*Plugin {
	plugins := make([]*Plugin, len(pls))
	for i, v := range pls {
		plugins[i] = ToCorePluginMsg(v)
	}
	return plugins
}

func MsgToCorePlugin(msg *Plugin) core.Plugin {
	pl := &SubPlugin{
		typeName: msg.TypeName,
		name:     msg.Name,
		version:  int(msg.Version),
	}
	return core.Plugin(pl)
}

func MsgToCorePlugins(msg []*Plugin) []core.Plugin {
	plugins := make([]core.Plugin, len(msg))
	for i, v := range msg {
		plugins[i] = MsgToCorePlugin(v)
	}
	return plugins
}

func ToSubPlugins(msg []*SubscribedPlugin) []core.SubscribedPlugin {
	plugins := make([]core.SubscribedPlugin, len(msg))
	for i, v := range msg {
		plugins[i] = ToSubPlugin(v)
	}
	return plugins
}

func ToSubPluginsMsg(sp []core.SubscribedPlugin) []*SubscribedPlugin {
	plugins := make([]*SubscribedPlugin, len(sp))
	for i, v := range sp {
		plugins[i] = ToSubPluginMsg(v)
	}
	return plugins
}

func ConfigMapToConfig(cfg *ConfigMap) *cdata.ConfigDataNode {
	if cfg == nil {
		return nil
	}
	config := cdata.FromTable(ParseConfig(cfg))
	return config
}

func ParseConfig(config *ConfigMap) map[string]ctypes.ConfigValue {
	c := make(map[string]ctypes.ConfigValue)
	for k, v := range config.IntMap {
		ival := ctypes.ConfigValueInt{Value: int(v)}
		c[k] = ival
	}
	for k, v := range config.FloatMap {
		fval := ctypes.ConfigValueFloat{Value: v}
		c[k] = fval
	}
	for k, v := range config.StringMap {
		sval := ctypes.ConfigValueStr{Value: v}
		c[k] = sval
	}
	for k, v := range config.BoolMap {
		bval := ctypes.ConfigValueBool{Value: v}
		c[k] = bval
	}
	return c
}

func ConfigToConfigMap(cd *cdata.ConfigDataNode) *ConfigMap {

	return ToConfigMap(cd.Table())
}

func ToConfigMap(cv map[string]ctypes.ConfigValue) *ConfigMap {
	newConfig := &ConfigMap{
		IntMap:    make(map[string]int64),
		FloatMap:  make(map[string]float64),
		StringMap: make(map[string]string),
		BoolMap:   make(map[string]bool),
	}
	for k, v := range cv {
		switch v.Type() {
		case "integer":
			newConfig.IntMap[k] = int64(v.(ctypes.ConfigValueInt).Value)
		case "float":
			newConfig.FloatMap[k] = v.(ctypes.ConfigValueFloat).Value
		case "string":
			newConfig.StringMap[k] = v.(ctypes.ConfigValueStr).Value
		case "bool":
			newConfig.BoolMap[k] = v.(ctypes.ConfigValueBool).Value
		}
	}
	return newConfig
}

func ConvertSnapErrors(s []*SnapError) []serror.SnapError {
	rerrs := make([]serror.SnapError, len(s))
	for i, err := range s {
		rerrs[i] = serror.New(errors.New(err.ErrorString), GetFields(err))
	}
	return rerrs
}

func ToSnapError(e *SnapError) serror.SnapError {
	if e == nil {
		return nil
	}
	return serror.New(errors.New(e.ErrorString), GetFields(e))
}

func NewErrors(errs []serror.SnapError) []*SnapError {
	errors := make([]*SnapError, len(errs))
	for i, err := range errs {
		fields := make(map[string]string)
		for k, v := range err.Fields() {
			switch t := v.(type) {
			case string:
				fields[k] = t
			case int:
				fields[k] = strconv.Itoa(t)
			case float64:
				fields[k] = strconv.FormatFloat(t, 'f', -1, 64)
			default:
				log.Errorf("Unexpected type %v\n", t)
			}
		}
		errors[i] = &SnapError{ErrorFields: fields, ErrorString: err.Error()}
	}
	return errors
}

func GetError(s *SnapError) string {
	return s.ErrorString
}

func GetFields(s *SnapError) map[string]interface{} {
	fields := make(map[string]interface{}, len(s.ErrorFields))
	for key, value := range s.ErrorFields {
		fields[key] = value
	}
	return fields
}
