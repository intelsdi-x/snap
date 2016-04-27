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

package rpc

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/internal/common"
)

type Metric struct {
	Namespace          []string                  `json:"namespace"`
	Version            int                       `json:"version"`
	LastAdvertisedTime time.Time                 `json:"last_advertised_time"`
	ConfigPolicy       *cpolicy.ConfigPolicyNode `json:"policy"`
}

type LoadedPlugin struct {
	Name            string                `json:"name"`
	Version         int                   `json:"version"`
	TypeName        string                `json:"type_name"`
	IsSigned        bool                  `json:"signed"`
	Status          string                `json:"status"`
	LoadedTimestamp time.Time             `json:"timestamp"`
	ConfigPolicy    *cpolicy.ConfigPolicy `json:"policy,omitempty"`
}

type AvailablePlugin struct {
	Name     string    `json:"name"`
	Version  int       `json:"version"`
	TypeName string    `json:"type_name"`
	Signed   bool      `json:"signed"`
	HitCount int       `json:"hit_count"`
	ID       uint32    `json:"id"`
	LastHit  time.Time `json:"last_hit"`
}

type catalogedPlugin struct {
	name       string
	version    int
	pluginType string
}

func (p *catalogedPlugin) Name() string {
	return p.name
}

func (p *catalogedPlugin) Version() int {
	return p.version
}

func (p *catalogedPlugin) TypeName() string {
	return p.pluginType
}

func NewCatalogedPlugin(n string, v int, pt string) *catalogedPlugin {
	return &catalogedPlugin{
		name:       n,
		version:    v,
		pluginType: pt,
	}
}

func ReplyToLoadedPlugin(lp *PluginReply) (*LoadedPlugin, error) {
	loadedPlugin := &LoadedPlugin{
		Name:            lp.Name,
		Version:         int(lp.Version),
		TypeName:        lp.TypeName,
		IsSigned:        lp.IsSigned,
		Status:          lp.Status,
		LoadedTimestamp: time.Unix(lp.LoadedTimestamp.Sec, lp.LoadedTimestamp.Nsec),
		ConfigPolicy:    cpolicy.New(),
	}
	err := json.Unmarshal(lp.ConfigPolicy, &loadedPlugin.ConfigPolicy)
	if err != nil {
		return nil, err
	}
	return loadedPlugin, nil
}

func ReplyToMetric(m *MetricReply) (*Metric, error) {
	metric := &Metric{
		Namespace:          m.Namespace,
		Version:            int(m.Version),
		LastAdvertisedTime: time.Unix(m.LastAdvertisedTime.Sec, m.LastAdvertisedTime.Nsec),
		ConfigPolicy:       cpolicy.NewPolicyNode(),
	}
	err := json.Unmarshal(m.ConfigPolicy, &metric.ConfigPolicy)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func ReplyToMetrics(reply []*MetricReply) ([]*Metric, error) {
	metrics := make([]*Metric, 0, len(reply))
	for _, met := range reply {
		m, err := ReplyToMetric(met)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func ReplyToAvailablePlugins(reply []*AvailablePluginReply) []*AvailablePlugin {
	plugins := make([]*AvailablePlugin, 0, len(reply))
	for _, p := range reply {
		plugin := ReplyToAvailablePlugin(p)
		plugins = append(plugins, plugin)
	}
	return plugins
}

func ReplyToAvailablePlugin(reply *AvailablePluginReply) *AvailablePlugin {
	return &AvailablePlugin{
		Name:     reply.Name,
		Version:  int(reply.Version),
		TypeName: reply.TypeName,
		Signed:   reply.IsSigned,
		HitCount: int(reply.HitCount),
		ID:       reply.ID,
		LastHit:  time.Unix(reply.LastHitTimestamp.Sec, reply.LastHitTimestamp.Nsec),
	}
}

func ConvertSnapErrors(s []*common.SnapError) []serror.SnapError {
	rerrs := make([]serror.SnapError, len(s))
	for i, err := range s {
		rerrs[i] = serror.New(errors.New(err.ErrorString), getFields(err))
	}
	return rerrs
}

func NewErrors(errs []serror.SnapError) []*common.SnapError {
	errors := make([]*common.SnapError, len(errs))
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
		errors[i] = &common.SnapError{ErrorFields: fields, ErrorString: err.Error()}
	}
	return errors
}

func getError(s *common.SnapError) string {
	return s.ErrorString
}

func getFields(s *common.SnapError) map[string]interface{} {
	fields := make(map[string]interface{}, len(s.ErrorFields))
	for key, value := range s.ErrorFields {
		fields[key] = value
	}
	return fields
}
