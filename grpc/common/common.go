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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	log "github.com/sirupsen/logrus"
)

// Convert a core.Metric to common.Metric protobuf message
func ToMetric(co core.Metric) *Metric {
	cm := &Metric{
		Namespace: ToNamespace(co.Namespace()),
		Version:   int64(co.Version()),
		Tags:      co.Tags(),
		Timestamp: &Time{
			Sec:  co.Timestamp().Unix(),
			Nsec: int64(co.Timestamp().Nanosecond()),
		},
		LastAdvertisedTime: &Time{
			Sec:  time.Now().Unix(),
			Nsec: int64(time.Now().Nanosecond()),
		},
	}
	if co.Config() != nil {
		cm.Config = ConfigToConfigMap(co.Config())
	}
	switch t := co.Data().(type) {
	case string:
		cm.Data = &Metric_StringData{t}
	case float64:
		cm.Data = &Metric_Float64Data{t}
	case float32:
		cm.Data = &Metric_Float32Data{t}
	case int32:
		cm.Data = &Metric_Int32Data{t}
	case int:
		cm.Data = &Metric_Int64Data{int64(t)}
	case int64:
		cm.Data = &Metric_Int64Data{t}
	case uint32:
		cm.Data = &Metric_Uint32Data{t}
	case uint64:
		cm.Data = &Metric_Uint64Data{t}
	case []byte:
		cm.Data = &Metric_BytesData{t}
	case bool:
		cm.Data = &Metric_BoolData{t}
	case nil:
		cm.Data = nil
	default:
		panic(fmt.Sprintf("unsupported type: %s", t))
	}
	return cm
}

// Convert core.Namespace to common.Namespace protobuf message
func ToNamespace(n core.Namespace) []*NamespaceElement {
	elements := make([]*NamespaceElement, 0, len(n))
	for _, value := range n {
		ne := &NamespaceElement{
			Value:       value.Value,
			Description: value.Description,
			Name:        value.Name,
		}
		elements = append(elements, ne)
	}
	return elements
}

func ToTime(t time.Time) *Time {
	return &Time{
		Nsec: t.Unix(),
		Sec:  int64(t.Second()),
	}
}

// Convert a slice of core.Metrics to []*common.Metric protobuf messages
func NewMetrics(ms []core.Metric) []*Metric {
	metrics := make([]*Metric, len(ms))
	for i, m := range ms {
		metrics[i] = ToMetric(m)
	}
	return metrics
}

type metric struct {
	namespace          core.Namespace
	version            int
	config             *cdata.ConfigDataNode
	lastAdvertisedTime time.Time
	timeStamp          time.Time
	data               interface{}
	tags               map[string]string
	description        string
	unit               string
}

func (m *metric) Namespace() core.Namespace     { return m.namespace }
func (m *metric) Config() *cdata.ConfigDataNode { return m.config }
func (m *metric) Version() int                  { return m.version }
func (m *metric) Data() interface{}             { return m.data }
func (m *metric) Tags() map[string]string       { return m.tags }
func (m *metric) LastAdvertisedTime() time.Time { return m.lastAdvertisedTime }
func (m *metric) Timestamp() time.Time          { return m.timeStamp }
func (m *metric) Description() string           { return m.description }
func (m *metric) Unit() string                  { return m.unit }

// Convert common.Metric to core.Metric
func ToCoreMetric(mt *Metric) core.Metric {
	var lastAdvertisedTime time.Time
	// if the lastAdvertisedTime is not set we handle.  -62135596800 represents the
	// number of seconds from 0001-1970 and is the default value for time.Unix.
	if mt.LastAdvertisedTime.Sec == int64(-62135596800) {
		lastAdvertisedTime = time.Unix(time.Now().Unix(), int64(time.Now().Nanosecond()))
	} else {
		lastAdvertisedTime = time.Unix(mt.LastAdvertisedTime.Sec, mt.LastAdvertisedTime.Nsec)
	}
	ret := &metric{
		namespace:          ToCoreNamespace(mt.Namespace),
		version:            int(mt.Version),
		tags:               mt.Tags,
		timeStamp:          time.Unix(mt.Timestamp.Sec, mt.Timestamp.Nsec),
		lastAdvertisedTime: lastAdvertisedTime,
		config:             ConfigMapToConfig(mt.Config),
		description:        mt.Description,
		unit:               mt.Unit,
	}

	switch mt.Data.(type) {
	case *Metric_BytesData:
		ret.data = mt.GetBytesData()
	case *Metric_StringData:
		ret.data = mt.GetStringData()
	case *Metric_Float32Data:
		ret.data = mt.GetFloat32Data()
	case *Metric_Float64Data:
		ret.data = mt.GetFloat64Data()
	case *Metric_Int32Data:
		ret.data = mt.GetInt32Data()
	case *Metric_Int64Data:
		ret.data = mt.GetInt64Data()
	case *Metric_Uint32Data:
		ret.data = mt.GetUint32Data()
	case *Metric_Uint64Data:
		ret.data = mt.GetUint64Data()
	case *Metric_BoolData:
		ret.data = mt.GetBoolData()
	}
	return ret
}

func MetricToRequested(mts []*Metric) []core.RequestedMetric {
	ret := make([]core.RequestedMetric, len(mts))
	for i, mt := range mts {
		met := &metric{
			namespace: ToCoreNamespace(mt.Namespace),
			version:   int(mt.Version),
		}
		ret[i] = met
	}
	return ret
}

// Convert common.Namespace protobuf message to core.Namespace
func ToCoreNamespace(n []*NamespaceElement) core.Namespace {
	var namespace core.Namespace
	for _, val := range n {
		ele := core.NamespaceElement{
			Value:       val.Value,
			Description: val.Description,
			Name:        val.Name,
		}
		namespace = append(namespace, ele)
	}
	return namespace
}

// Convert slice of common.Metric to []core.Metric
func ToCoreMetrics(mts []*Metric) []core.Metric {
	metrics := make([]core.Metric, len(mts))
	for i, mt := range mts {
		metrics[i] = ToCoreMetric(mt)
	}
	return metrics
}

// Convert slice of common.Metric to []core.RequestedMetric
func ToRequestedMetrics(mts []*Metric) []core.RequestedMetric {
	metrics := make([]core.RequestedMetric, len(mts))
	for i, mt := range mts {
		metrics[i] = ToCoreMetric(mt)
	}
	return metrics
}

func RequestedToMetric(requested []core.RequestedMetric) []*Metric {
	reqMets := make([]*Metric, len(requested))
	for i, r := range requested {
		rm := &Metric{
			Namespace: ToNamespace(r.Namespace()),
			Version:   int64(r.Version()),
			Config:    &ConfigMap{},
			Tags:      map[string]string{},
			Timestamp: &Time{
				Sec:  time.Now().Unix(),
				Nsec: int64(time.Now().Nanosecond()),
			},
			LastAdvertisedTime: &Time{
				Sec:  time.Now().Unix(),
				Nsec: int64(time.Now().Nanosecond()),
			},
		}
		reqMets[i] = rm
	}
	return reqMets
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

// Convert from core.SubscribedPlugin to core.Plugin
func ToCorePlugin(pl core.SubscribedPlugin) core.Plugin {
	return core.Plugin(pl)
}

// Convert from []core.SubscribedPlugin to []core.Plugin
func ToCorePlugins(pl []core.SubscribedPlugin) []core.Plugin {
	plugins := make([]core.Plugin, len(pl))
	for i, v := range pl {
		plugins[i] = v
	}
	return plugins
}

// Convert core.SubscribedPlugin to SubscribedPlugin protobuf message
func ToSubPluginMsg(pl core.SubscribedPlugin) *SubscribedPlugin {
	return &SubscribedPlugin{
		TypeName: pl.TypeName(),
		Name:     pl.Name(),
		Version:  int64(pl.Version()),
		Config:   ConfigToConfigMap(pl.Config()),
	}
}

// Convert from a SubscribedPlugin protobuf message to core.SubscribedPlugin
func ToSubPlugin(msg *SubscribedPlugin) core.SubscribedPlugin {
	return SubPlugin{
		typeName: msg.TypeName,
		name:     msg.Name,
		version:  int(msg.Version),
		config:   ConfigMapToConfig(msg.Config),
	}
}

// Convert from core.Plugin to Plugin protobuf message
func ToCorePluginMsg(pl core.Plugin) *Plugin {
	return &Plugin{
		TypeName: pl.TypeName(),
		Name:     pl.Name(),
		Version:  int64(pl.Version()),
	}
}

// Convert from Plugin protobuf message to core.Plugin
func ToCorePluginsMsg(pls []core.Plugin) []*Plugin {
	plugins := make([]*Plugin, len(pls))
	for i, v := range pls {
		plugins[i] = ToCorePluginMsg(v)
	}
	return plugins
}

// Converts Plugin protobuf message to core.Plugin
func MsgToCorePlugin(msg *Plugin) core.Plugin {
	pl := &SubPlugin{
		typeName: msg.TypeName,
		name:     msg.Name,
		version:  int(msg.Version),
	}
	return core.Plugin(pl)
}

// Converts slice of plugin protobuf messages to core.Plugins
func MsgToCorePlugins(msg []*Plugin) []core.Plugin {
	plugins := make([]core.Plugin, len(msg))
	for i, v := range msg {
		plugins[i] = MsgToCorePlugin(v)
	}
	return plugins
}

// Converts slice of SubscribedPlugin Messages to core.SubscribedPlugins
func ToSubPlugins(msg []*SubscribedPlugin) []core.SubscribedPlugin {
	plugins := make([]core.SubscribedPlugin, len(msg))
	for i, v := range msg {
		plugins[i] = ToSubPlugin(v)
	}
	return plugins
}

// Converts core.SubscribedPlugins to protobuf messages
func ToSubPluginsMsg(sp []core.SubscribedPlugin) []*SubscribedPlugin {
	plugins := make([]*SubscribedPlugin, len(sp))
	for i, v := range sp {
		plugins[i] = ToSubPluginMsg(v)
	}
	return plugins
}

// Converts configMaps to ConfigDataNode
func ConfigMapToConfig(cfg *ConfigMap) *cdata.ConfigDataNode {
	if cfg == nil {
		return nil
	}
	config := cdata.FromTable(ParseConfig(cfg))
	return config
}

// Parses a configMap to a map[string]ctypes.ConfigValue
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

// Converts ConfigDataNode to ConfigMap protobuf message
func ConfigToConfigMap(cd *cdata.ConfigDataNode) *ConfigMap {
	if cd == nil {
		return nil
	}
	return ToConfigMap(cd.Table())
}

// Converts ConfigDataNode to ConfigMap protobuf message
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

// Converts SnapError protobuf messages to serror.Snaperrors
func ConvertSnapErrors(s []*SnapError) []serror.SnapError {
	rerrs := make([]serror.SnapError, len(s))
	for i, err := range s {
		rerrs[i] = serror.New(errors.New(err.ErrorString), GetFields(err))
	}
	return rerrs
}

// Converts a single SnapError protobuf message to SnapError
func ToSnapError(e *SnapError) serror.SnapError {
	if e == nil {
		return nil
	}
	return serror.New(errors.New(e.ErrorString), GetFields(e))
}

// Converts a group of SnapErrors to SnapError protobuf messages
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

// Returns a SnapError protobuf messages errorString
func GetError(s *SnapError) string {
	return s.ErrorString
}

// Returns the fields from a SnapError protobuf message
func GetFields(s *SnapError) map[string]interface{} {
	fields := make(map[string]interface{}, len(s.ErrorFields))
	for key, value := range s.ErrorFields {
		fields[key] = value
	}
	return fields
}
