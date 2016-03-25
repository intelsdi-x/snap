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

package rpc

import (
	"encoding/json"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
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

func NewcatalogedPlugin(n string, v int, pt string) *catalogedPlugin {
	return &catalogedPlugin{
		name:       n,
		version:    v,
		pluginType: pt,
	}
}

func ReplyToLoadedPlugin(lp *PluginReply) LoadedPlugin {
	loadedPlugin := LoadedPlugin{
		Name:            lp.Name,
		Version:         int(lp.Version),
		TypeName:        lp.TypeName,
		IsSigned:        lp.IsSigned,
		Status:          lp.Status,
		LoadedTimestamp: time.Unix(lp.LoadedTimestamp.Sec, lp.LoadedTimestamp.Nsec),
		ConfigPolicy:    cpolicy.New(),
	}
	json.Unmarshal(lp.ConfigPolicy, &loadedPlugin)
	return loadedPlugin
}

func ReplyToMetric(m *MetricReply) Metric {
	metric := Metric{
		Namespace:          m.Namespace,
		Version:            int(m.Version),
		LastAdvertisedTime: time.Unix(m.LastAdvertisedTime.Sec, m.LastAdvertisedTime.Nsec),
		ConfigPolicy:       cpolicy.NewPolicyNode(),
	}
	json.Unmarshal(m.ConfigPolicy, &metric.ConfigPolicy)
	return metric
}

func ReplyToMetrics(reply []*MetricReply) []Metric {
	metrics := make([]Metric, 0, len(reply))
	for _, met := range reply {
		metrics = append(metrics, ReplyToMetric(met))
	}
	return metrics
}

func ReplyToAvailablePlugins(reply []*AvailablePluginReply) []AvailablePlugin {
	plugins := make([]AvailablePlugin, 0, len(reply))
	for _, p := range reply {
		plugin := ReplyToAvailablePlugin(p)
		plugins = append(plugins, plugin)
	}
	return plugins
}

func ReplyToAvailablePlugin(reply *AvailablePluginReply) AvailablePlugin {
	return AvailablePlugin{
		Name:     reply.Name,
		Version:  int(reply.Version),
		TypeName: reply.TypeName,
		Signed:   reply.IsSigned,
		HitCount: int(reply.HitCount),
		ID:       reply.ID,
		LastHit:  time.Unix(reply.LastHitTimestamp.Sec, reply.LastHitTimestamp.Nsec),
	}
}
