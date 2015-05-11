package wmap

func (w *WorkflowMap) String() string {
	var out string
	out += "Workflow\n"
	out += "\tCollect:\n"
	if w.CollectNode != nil {
		out += w.CollectNode.String("\t\t")
	} else {
		out += "\n"
	}

	return out
}

func (c *CollectWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + "Metric Namespaces:\n"
	for _, x := range c.MetricsNamespaces {
		out += pad + "\t\t" + x + "\n"
	}
	out += "\n"
	out += pad + "Process Nodes:\n"
	for _, pr := range c.ProcessNodes {
		out += pr.String(pad)
	}
	out += "\n"
	out += pad + "Publish Nodes:\n"
	for _, pu := range c.PublishNodes {
		out += pu.String(pad)
	}
	out += "\n"
	return out
}

func (p *ProcessWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("Name: %s\n", p.Name)
	out += pad + fmt.Sprintf("Version: %d\n", p.Version)

	out += pad + "Process Nodes:\n"
	for _, pr := range p.ProcessNodes {
		out += pr.String(pad)
	}
	out += pad + "Publish Nodes:\n"
	for _, pu := range p.PublishNodes {
		out += pu.String(pad)
	}
	return out
}

func (p *PublishWorkflowMapNode) String(pad string) string {
	var out string
	out += pad + fmt.Sprintf("\tName: %s\n", p.Name)
	out += pad + fmt.Sprintf("\tVersion: %d\n", p.Version)

	return out
}
