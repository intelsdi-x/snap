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

package agreement

import (
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/memberlist"
	"github.com/intelsdi-x/snap/core"
)

const (
	RestPort               = "rest_api_port"
	RestProtocol           = "rest_proto"
	RestInsecureSkipVerify = "rest_insecure"
)

var logger = log.WithFields(log.Fields{
	"_module": "tribe-agreement",
})

type Agreement struct {
	Name            string             `json:"name"`
	PluginAgreement *pluginAgreement   `json:"plugin_agreement,omitempty"`
	TaskAgreement   *taskAgreement     `json:"task_agreement,omitempty"`
	Members         map[string]*Member `json:"members,omitempty"`
}

type plugins []Plugin

type pluginAgreement struct {
	Name    string  `json:"-"`
	Plugins plugins `json:"plugins,omitempty"`
}

type tasks []Task

type taskAgreement struct {
	Name  string `json:"-"`
	Tasks tasks  `json:"tasks,omitempty"`
}

type Task struct {
	ID            string `json:"id"`
	StartOnCreate bool   `json:"start_on_create"`
}

func New(name string) *Agreement {
	return &Agreement{
		Name: name,
		PluginAgreement: &pluginAgreement{
			Name:    name,
			Plugins: plugins{},
		},
		TaskAgreement: &taskAgreement{
			Name:  name,
			Tasks: tasks{},
		},
		Members: map[string]*Member{},
	}
}

type Member struct {
	Tags            map[string]string         `json:"tags,omitempty"`
	Name            string                    `json:"name"`
	Node            *memberlist.Node          `json:"-"`
	PluginAgreement *pluginAgreement          `json:"-"`
	TaskAgreements  map[string]*taskAgreement `json:"-"`
}

func NewMember(node *memberlist.Node) *Member {
	return &Member{
		Name:           node.Name,
		Node:           node,
		TaskAgreements: map[string]*taskAgreement{},
	}
}

func (m *Member) GetRestPort() string {
	return m.Tags[RestPort]
}

func (m *Member) GetRestProto() string {
	return m.Tags[RestProtocol]
}

func (m *Member) GetRestInsecureSkipVerify() bool {
	if m.Tags[RestInsecureSkipVerify] == "true" {
		return true
	}
	return false
}

func (m *Member) GetName() string {
	return m.Name
}

func (m *Member) GetAddr() net.IP {
	return m.Node.Addr
}

type Plugin struct {
	Name_    string          `json:"name"`
	Version_ int             `json:"version"`
	Type_    core.PluginType `json:"type"`
}

func (p Plugin) Name() string {
	return p.Name_
}

func (p Plugin) Version() int {
	return p.Version_
}

func (p Plugin) TypeName() string {
	return p.Type_.String()
}

func newPlugin(n string, v int, t core.PluginType) *Plugin {
	return &Plugin{
		Name_:    n,
		Version_: v,
		Type_:    t,
	}
}

// contains - Returns boolean indicating whether the plugin was found.
// If the plugin is found the index returned as the second return value.
func (p plugins) Contains(item Plugin) (bool, int) {
	for idx, i := range p {
		if i.Name() == item.Name() && i.Version() == item.Version() && i.TypeName() == item.TypeName() {
			return true, idx
		}
	}
	return false, -1
}

// contains - Returns boolean indicating whether the plugin was found.
// If the plugin is found the index returned as the second return value.
func (t tasks) Contains(item Task) (bool, int) {
	for idx, i := range t {
		if i.ID == item.ID {
			return true, idx
		}
	}
	return false, -1
}

func (a *pluginAgreement) Remove(plugin Plugin) bool {
	logger.WithFields(log.Fields{
		"agreement": a.Name,
		"plugin":    plugin.Name(),
		"_block":    "remove",
	}).Debugln("Removing plugin")
	if ok, idx := a.Plugins.Contains(plugin); ok {
		a.Plugins = append(a.Plugins[idx+1:], a.Plugins[:idx]...)
		return true
	}
	return false
}

func (a *pluginAgreement) Add(plugin Plugin) bool {
	logger.WithFields(log.Fields{
		"agreement": a.Name,
		"plugin":    plugin.Name(),
		"_block":    "add",
	}).Debugln("Adding plugin")
	if ok, _ := a.Plugins.Contains(plugin); ok {
		return false
	}
	a.Plugins = append(a.Plugins, plugin)
	return true
}

func (a *taskAgreement) Add(task Task) bool {
	logger.WithFields(log.Fields{
		"agreement": a.Name,
		"task_id":   task.ID,
		"_block":    "add",
	}).Debugln("Adding task")
	if ok, _ := a.Tasks.Contains(task); ok {
		return false
	}
	a.Tasks = append(a.Tasks, task)
	return true
}

func (a *taskAgreement) Remove(task Task) bool {
	logger.WithFields(log.Fields{
		"agreement": a.Name,
		"task_id":   task.ID,
		"_block":    "remove",
	}).Debugln("Removing task")
	if ok, idx := a.Tasks.Contains(task); ok {
		a.Tasks = append(a.Tasks[idx+1:], a.Tasks[:idx]...)
		return true
	}
	return false
}
