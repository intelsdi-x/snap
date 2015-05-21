package wmap

import (
	"fmt"
)

func (w *WorkflowMap) String() string {
	var out string
	out += "Workflow\n"
	out += "   Collect:\n"
	if w.CollectNode != nil {
		out += w.CollectNode.String("      ")
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
	out += pad + "Process Nodes:\n"
	for _, pr := range c.ProcessNodes {
		out += pr.String(pad)
	}
	out += "\n"
	out += pad + "Publish Nodes:\n"
	for _, pu := range c.PublishNodes {
		out += pu.String(pad) + "\n"
	}
	out += "\n"
	return out
}

func (p *ProcessWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("   Name: %s\n", p.Name)
	out += pad + fmt.Sprintf("   Version: %d\n", p.Version)

	out += pad + "   Process Nodes:\n"
	for _, pr := range p.ProcessNodes {
		out += pr.String(pad + "   ")
	}
	out += pad + "   Publish Nodes:\n"
	for _, pu := range p.PublishNodes {
		out += pu.String(pad + "   ")
	}
	return out
}

func (p *PublishWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("   Name: %s\n", p.Name)
	out += pad + fmt.Sprintf("   Version: %d\n", p.Version)

	out += pad + "   Config:\n"
	for k, v := range p.Config {
		out += pad + "      " + fmt.Sprintf("%s=%+v\n", k, v)
	}
	return out
}
