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
	"fmt"
)

func (w *WorkflowMap) String() string {
	var out string
	out += "Workflow\n"
	out += "   Collect:\n"
	if w.Collect != nil {
		out += w.Collect.String("      ")
	} else {
		out += "\n"
	}

	return out
}

func (c *CollectWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + "Metrics:\n"
	for k, v := range c.Metrics {
		out += pad + fmt.Sprintf("      Namespace: %s\n", k)
		out += pad + fmt.Sprintf("         Version: %d\n", v.Version_)
	}
	out += "\n"
	out += pad + "Config:\n"
	for k, v := range c.Config {
		out += pad + "   " + k + "\n"
		for x, y := range v {
			out += pad + "      " + fmt.Sprintf("%s=%+v\n", x, y)
		}
	}
	out += "\n"
	out += pad + "Tags:\n"
	for k, v := range c.Tags {
		out += pad + "   " + k + "\n"
		for x, y := range v {
			out += pad + "      " + fmt.Sprintf("%s=%+v\n", x, y)
		}
	}
	out += "\n"
	out += pad + "Process Nodes:\n"
	for _, pr := range c.Process {
		out += pr.String(pad)
	}
	out += "\n"
	out += pad + "Publish Nodes:\n"
	for _, pu := range c.Publish {
		out += pu.String(pad) + "\n"
	}
	out += "\n"
	return out
}

func (p *ProcessWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("   Name: %s\n", p.PluginName)
	out += pad + fmt.Sprintf("   Version: %d\n", p.PluginVersion)

	out += pad + "   Config:\n"
	for k, v := range p.Config {
		out += pad + "      " + fmt.Sprintf("%s=%+v\n", k, v)
	}
	out += pad + "   Target:" + p.Target + "\n"

	out += pad + "   Process Nodes:\n"
	for _, pr := range p.Process {
		out += pr.String(pad + "   ")
	}
	out += pad + "   Publish Nodes:\n"
	for _, pu := range p.Publish {
		out += pu.String(pad + "   ")
	}
	return out
}

func (p *PublishWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("   Name: %s\n", p.PluginName)
	out += pad + fmt.Sprintf("   Version: %d\n", p.PluginVersion)

	out += pad + "   Config:\n"
	for k, v := range p.Config {
		out += pad + "      " + fmt.Sprintf("%s=%+v\n", k, v)
	}
	return out
}
