package wmap

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
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
	wf := sample()
	b, e := wf.ToJson()
	if e != nil {
		panic(e)
	}
	return string(b)
}

func SampleWorkflowMapYaml() string {
	wf := sample()
	b, e := wf.ToYaml()
	if e != nil {
		panic(e)
	}
	return string(b)
}

func sample() *WorkflowMap {
	wf := new(WorkflowMap)

	c1 := &CollectWorkflowMapNode{
		Config: make(map[string]map[string]interface{}),
	}
	c1.Config["/foo/bar"] = make(map[string]interface{})
	c1.Config["/foo/bar"]["user"] = "root"

	// pr1 := &ProcessWorkflowMapNode{Name: "learn", Version: 3}
	pu1 := &PublishWorkflowMapNode{
		Name:    "rabbitmq",
		Version: 5,
		Config:  make(map[string]interface{}),
	}

	pu1.Config["user"] = "root"
	var e error
	// e = pr1.Add(pu1)
	// handleErr(e)
	e = c1.Add(pu1)
	if e != nil {
		panic(e)
	}
	e = c1.AddMetricNamespace("/foo/bar")
	if e != nil {
		panic(e)
	}
	wf.CollectNode = c1
	return wf
}

// A map of a desired workflow that is used to create a scheduleWorkflow
type WorkflowMap struct {
	CollectNode *CollectWorkflowMapNode `json:"collect"yaml:"collect"`
}

func (w *WorkflowMap) ToJson() ([]byte, error) {
	return json.Marshal(w)
}

func (w *WorkflowMap) ToYaml() ([]byte, error) {
	return yaml.Marshal(w)
}

type CollectWorkflowMapNode struct {
	MetricsNamespaces []string                          `json:"metric_namespaces"yaml:"metric_namespaces"`
	Config            map[string]map[string]interface{} `json:"config"yaml:"config"`
	ProcessNodes      []ProcessWorkflowMapNode          `json:"process"yaml:"process"`
	PublishNodes      []PublishWorkflowMapNode          `json:"publish"yaml:"publish"`
}

func (c *CollectWorkflowMapNode) Add(node interface{}) error {
	switch x := node.(type) {
	case *ProcessWorkflowMapNode:
		c.ProcessNodes = append(c.ProcessNodes, *x)
	case *PublishWorkflowMapNode:
		c.PublishNodes = append(c.PublishNodes, *x)
	default:
		return errors.New(fmt.Sprintf("cannot add workflow node type (%v) to collect node as child", x))
	}
	return nil
}

func (c *CollectWorkflowMapNode) AddMetricNamespace(ns string) error {
	// TODO regex validation here that this matches /one/two/three format
	c.MetricsNamespaces = append(c.MetricsNamespaces, ns)
	return nil
}

type ProcessWorkflowMapNode struct {
	Name         string                   `json:"plugin_name"yaml:"plugin_name"`
	Version      int                      `json:"plugin_version"yaml:"plugin_version"`
	ProcessNodes []ProcessWorkflowMapNode `json:"process"yaml:"process"`
	PublishNodes []PublishWorkflowMapNode `json:"publish"yaml:"publish"`
	// TODO processor config
	Config map[string]interface{} `json:"config"yaml:"config"`
}

func (p *ProcessWorkflowMapNode) Add(node interface{}) error {
	switch x := node.(type) {
	case *ProcessWorkflowMapNode:
		p.ProcessNodes = append(p.ProcessNodes, *x)
	case *PublishWorkflowMapNode:
		p.PublishNodes = append(p.PublishNodes, *x)
	default:
		return errors.New(fmt.Sprintf("cannot add workflow node type (%v) to process node as child", x))
	}
	return nil
}

type PublishWorkflowMapNode struct {
	Name    string `json:"plugin_name"yaml:"plugin_name"`
	Version int    `json:"plugin_version"yaml:"plugin_version"`
	// TODO publisher config
	Config map[string]interface{} `json:"config"yaml:"config"`
}
