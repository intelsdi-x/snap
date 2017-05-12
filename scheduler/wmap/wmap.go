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

package wmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/pkg/stringutils"
)

var (
	InvalidPayload = errors.New("Payload to convert must be string or []byte")
)

func FromYaml(payload interface{}) (*WorkflowMap, error) {
	p, err := inStringBytes(payload)
	if err != nil {
		return nil, err
	}

	wmap := new(WorkflowMap)
	err = yaml.Unmarshal(p, wmap)
	if err != nil {
		return nil, err
	}
	return wmap, nil
}

func FromJson(payload interface{}) (*WorkflowMap, error) {
	p, err := inStringBytes(payload)
	if err != nil {
		return nil, err
	}

	wmap := new(WorkflowMap)
	err = json.Unmarshal(p, wmap)
	if err != nil {
		return nil, err
	}
	return wmap, nil
}

func inStringBytes(payload interface{}) ([]byte, error) {
	var p []byte
	switch tp := payload.(type) {
	case string:
		p = []byte(tp)
	case []byte:
		p = tp
	default:
		return p, InvalidPayload
	}
	return p, nil
}

func SampleWorkflowMapJson() string {
	wf := Sample()
	b, e := wf.ToJson()
	if e != nil {
		panic(e)
	}
	return string(b)
}

func SampleWorkflowMapYaml() string {
	wf := Sample()
	b, e := wf.ToYaml()
	if e != nil {
		panic(e)
	}
	return string(b)
}

func Sample() *WorkflowMap {
	wf := new(WorkflowMap)

	c1 := &CollectWorkflowMapNode{
		Metrics: make(map[string]metricInfo),
		Config:  make(map[string]map[string]interface{}),
	}
	c1.Config["/foo/bar"] = make(map[string]interface{})
	c1.Config["/foo/bar"]["user"] = "root"

	// pr1 := &ProcessWorkflowMapNode{Name: "learn", Version: 3}
	pu1 := &PublishWorkflowMapNode{
		PluginName:    "rabbitmq",
		PluginVersion: 5,
		Config:        make(map[string]interface{}),
	}

	pu1.Config["user"] = "root"
	var e error
	// e = pr1.Add(pu1)
	// handleErr(e)
	e = c1.Add(pu1)
	if e != nil {
		panic(e)
	}
	e = c1.AddMetric("/foo/bar", 1)
	if e != nil {
		panic(e)
	}
	wf.Collect = c1
	return wf
}

// WorkflowMap represents a map of a desired workflow that is used to create a scheduleWorkflow
type WorkflowMap struct {
	// required: true
	Collect *CollectWorkflowMapNode `json:"collect"yaml:"collect"`
}

func (w *WorkflowMap) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "collect":
			if err := json.Unmarshal(v, &w.Collect); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in workflow of task.", k)
		}
	}
	return nil
}

func NewWorkflowMap() *WorkflowMap {
	w := &WorkflowMap{}
	c := &CollectWorkflowMapNode{
		Metrics: make(map[string]metricInfo),
		Config:  make(map[string]map[string]interface{}),
	}
	w.Collect = c
	return w
}

func (w *WorkflowMap) ToJson() ([]byte, error) {
	return json.Marshal(w)
}

func (w *WorkflowMap) ToYaml() ([]byte, error) {
	return yaml.Marshal(w)
}

// CollectWorkflowMapNode represents Snap workflow data model.
type CollectWorkflowMapNode struct {
	// required: true
	Metrics map[string]metricInfo             `json:"metrics"yaml:"metrics"`
	Config  map[string]map[string]interface{} `json:"config,omitempty"yaml:"config"`
	Tags    map[string]map[string]string      `json:"tags,omitempty"yaml:"tags"`
	Process []ProcessWorkflowMapNode          `json:"process,omitempty"yaml:"process"`
	Publish []PublishWorkflowMapNode          `json:"publish,omitempty"yaml:"publish"`
}

func (cw *CollectWorkflowMapNode) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "metrics":
			if err := json.Unmarshal(v, &cw.Metrics); err != nil {
				return err
			}
		case "config":
			if err := json.Unmarshal(v, &cw.Config); err != nil {
				return fmt.Errorf("%v (while parsing 'config')", err)
			}
		case "tags":
			if err := json.Unmarshal(v, &cw.Tags); err != nil {
				return fmt.Errorf("%v (while parsing 'tags')", err)
			}
		case "process":
			if err := json.Unmarshal(v, &cw.Process); err != nil {
				return err
			}
		case "publish":
			if err := json.Unmarshal(v, &cw.Publish); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in collect workflow of task.", k)
		}
	}
	return nil
}

func (c *CollectWorkflowMapNode) GetMetrics() []Metric {
	metrics := make([]Metric, len(c.Metrics))
	i := 0
	for k, v := range c.Metrics {
		// Identify the character to split on by peaking
		// at the first character of each metric.
		firstChar := stringutils.GetFirstChar(k)
		ns := strings.Trim(k, firstChar)
		metrics[i] = Metric{
			namespace: strings.Split(ns, firstChar),
			version:   v.Version_,
		}
		i++
	}
	return metrics
}

func (c *CollectWorkflowMapNode) GetTags() map[string]map[string]string {
	return c.Tags
}

func NewCollectWorkflowMapNode() *CollectWorkflowMapNode {
	return &CollectWorkflowMapNode{
		Metrics: make(map[string]metricInfo),
		Config:  make(map[string]map[string]interface{}),
	}
}

// GetConfigTree converts config data for collection node in wmap into a proper cdata.ConfigDataTree
func (c *CollectWorkflowMapNode) GetConfigTree() (*cdata.ConfigDataTree, error) {
	cdt := cdata.NewTree()
	// Iterate over config and attempt to convert into data nodes in the tree
	for ns_, cmap := range c.Config {

		ns := strings.Split(ns_, "/")[1:]
		cdn, err := configtoConfigDataNode(cmap, ns_)
		if err != nil {
			return nil, err
		}
		cdt.Add(ns, cdn)
	}
	return cdt, nil
}

func (c *CollectWorkflowMapNode) Add(node interface{}) error {
	switch x := node.(type) {
	case *ProcessWorkflowMapNode:
		c.Process = append(c.Process, *x)
	case *PublishWorkflowMapNode:
		c.Publish = append(c.Publish, *x)
	default:
		return errors.New(fmt.Sprintf("cannot add workflow node type (%v) to collect node as child", x))
	}
	return nil
}

func (c *CollectWorkflowMapNode) AddMetric(ns string, v int) error {
	// TODO regex validation here that this matches /one/two/three format
	// c.MetricsNamespaces = append(c.MetricsNamespaces, ns)
	c.Metrics[ns] = metricInfo{Version_: v}
	return nil
}

func (c *CollectWorkflowMapNode) AddConfigItem(ns, key string, value interface{}) {
	if c.Config[ns] == nil {
		c.Config[ns] = make(map[string]interface{})
	}
	c.Config[ns][key] = value
}

type ProcessWorkflowMapNode struct {
	// required: true
	PluginName    string                   `json:"plugin_name"yaml:"plugin_name"`
	PluginVersion int                      `json:"plugin_version"yaml:"plugin_version"`
	Process       []ProcessWorkflowMapNode `json:"process,omitempty"yaml:"process"`
	Publish       []PublishWorkflowMapNode `json:"publish,omitempty"yaml:"publish"`
	// Config the configuration of a processor.
	Config map[string]interface{} `json:"config,omitempty"yaml:"config"`
	Target string                 `json:"target"yaml:"target"`
}

func (pw *ProcessWorkflowMapNode) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "plugin_name":
			if err := json.Unmarshal(v, &pw.PluginName); err != nil {
				return fmt.Errorf("%v (while parsing 'plugin_name')", err)
			}
		case "plugin_version":
			if err := json.Unmarshal(v, &pw.PluginVersion); err != nil {
				return fmt.Errorf("%v (while parsing 'plugin_version')", err)
			}
		case "process":
			if err := json.Unmarshal(v, &pw.Process); err != nil {
				return err
			}
		case "publish":
			if err := json.Unmarshal(v, &pw.Publish); err != nil {
				return err
			}
		case "config":
			if err := json.Unmarshal(v, &pw.Config); err != nil {
				return fmt.Errorf("%v (while parsing 'config')", err)
			}
		case "target":
			if err := json.Unmarshal(v, &pw.Target); err != nil {
				return fmt.Errorf("%v (while parsing 'target')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in process workflow of task.", k)
		}
	}
	return nil

}

func NewProcessNode(name string, version int) *ProcessWorkflowMapNode {
	p := &ProcessWorkflowMapNode{
		PluginName:    name,
		PluginVersion: version,
	}
	return p
}

func (p *ProcessWorkflowMapNode) Add(node interface{}) error {
	switch x := node.(type) {
	case *ProcessWorkflowMapNode:
		p.Process = append(p.Process, *x)
	case *PublishWorkflowMapNode:
		p.Publish = append(p.Publish, *x)
	default:
		return errors.New(fmt.Sprintf("cannot add workflow node type (%v) to process node as child", x))
	}
	return nil
}

func (p *ProcessWorkflowMapNode) AddConfigItem(key string, value interface{}) {
	if p.Config == nil {
		p.Config = make(map[string]interface{})
	}
	p.Config[key] = value
}

func (p *ProcessWorkflowMapNode) GetConfigNode() (*cdata.ConfigDataNode, error) {
	if p.Config == nil {
		return cdata.NewNode(), nil
	}
	return configtoConfigDataNode(p.Config, "")
}

type PublishWorkflowMapNode struct {
	// required: true
	PluginName    string `json:"plugin_name"yaml:"plugin_name"`
	PluginVersion int    `json:"plugin_version"yaml:"plugin_version"`
	// required: true
	// Config the config of a publisher
	Config map[string]interface{} `json:"config,omitempty"yaml:"config"`
	Target string                 `json:"target"yaml:"target"`
}

func (pw *PublishWorkflowMapNode) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "plugin_name":
			if err := json.Unmarshal(v, &pw.PluginName); err != nil {
				return fmt.Errorf("%v (while parsing 'plugin_name')", err)
			}
		case "plugin_version":
			if err := json.Unmarshal(v, &pw.PluginVersion); err != nil {
				return fmt.Errorf("%v (while parsing 'plugin_version')", err)
			}
		case "config":
			if err := json.Unmarshal(v, &pw.Config); err != nil {
				return fmt.Errorf("%v (while parsing 'config')", err)
			}
		case "target":
			if err := json.Unmarshal(v, &pw.Target); err != nil {
				return fmt.Errorf("%v (while parsing 'target')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in publish workflow of task.", k)
		}
	}
	return nil
}

func NewPublishNode(name string, version int) *PublishWorkflowMapNode {
	p := &PublishWorkflowMapNode{
		PluginName:    name,
		PluginVersion: version,
	}
	return p
}

func (p *PublishWorkflowMapNode) AddConfigItem(key string, value interface{}) {
	if p.Config == nil {
		p.Config = make(map[string]interface{})
	}
	p.Config[key] = value
}

func (p *PublishWorkflowMapNode) GetConfigNode() (*cdata.ConfigDataNode, error) {
	if p.Config == nil {
		return cdata.NewNode(), nil
	}
	return configtoConfigDataNode(p.Config, "")
}

type metricInfo struct {
	Version_ int `json:"version"yaml:"version"`
}

func (m *metricInfo) UnmarshalJSON(data []byte) error {
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	for k, v := range t {
		switch k {
		case "version":
			if err := json.Unmarshal(v, &m.Version_); err != nil {
				return fmt.Errorf("%v (while parsing 'version')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in metrics in collect workflow of task", k)
		}
	}
	return nil
}

type Metric struct {
	namespace []string
	version   int
}

func (m Metric) Namespace() []string {
	return m.namespace
}

func (m Metric) Version() int {
	return m.version
}

func configtoConfigDataNode(cmap map[string]interface{}, ns string) (*cdata.ConfigDataNode, error) {
	cdn := cdata.NewNode()
	for ck, cv := range cmap {
		switch v := cv.(type) {
		case string:
			cdn.AddItem(ck, ctypes.ConfigValueStr{Value: v})
		case int:
			cdn.AddItem(ck, ctypes.ConfigValueInt{Value: v})
		case float64:
			//working around the fact that json decodes numbers to floats
			//if we can convert the number to an int without loss it will be an int
			if v == float64(int(v)) {
				cdn.AddItem(ck, ctypes.ConfigValueInt{Value: int(v)})
			} else {
				cdn.AddItem(ck, ctypes.ConfigValueFloat{Value: v})
			}
		case bool:
			cdn.AddItem(ck, ctypes.ConfigValueBool{Value: v})
		default:
			// TODO make sure this is covered in tests!!!
			return nil, errors.New(fmt.Sprintf("Cannot convert config value to config data node: %s=>%+v", ns, v))
		}
	}
	return cdn, nil
}
