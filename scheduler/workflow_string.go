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

package scheduler

import (
	"fmt"

	"github.com/intelsdi-x/snap/core"
)

func (s *schedulerWorkflow) String() string {
	var out string
	out += "Workflow\n"
	out += "  (Collector)\n"
	out += fmt.Sprintf("    State: %s\n", s.StateString())
	out += metricString("      ", s.metrics)
	out += fmt.Sprintf("    (Processor)\n")
	for _, p := range s.processNodes {
		out += p.String("      ")
	}
	out += fmt.Sprintf("    (Publishers)\n")
	for _, p := range s.publishNodes {
		out += p.String("      ")
	}
	return out
}

func metricString(pad string, rm []core.RequestedMetric) string {
	var out string
	for _, m := range rm {
		out += fmt.Sprintf("%sMetric: %s\n", pad, m.Namespace().String())
		out += fmt.Sprintf("%s  Version: %d\n", pad, m.Version())
	}
	return out
}

func (p *processNode) String(args ...string) string {
	pad := ""
	var out string
	if len(args) > 0 {
		pad = args[0]
	}
	out += fmt.Sprintf("%sName: %s\n", pad, p.Name())
	out += fmt.Sprintf("%s   Version: %d\n", pad, p.Version())
	out += fmt.Sprintf("%s   Config:\n", pad)
	for k, v := range p.Config().Table() {
		out += fmt.Sprintf("%s      %s=%+v\n", pad, k, v)
	}
	out += fmt.Sprintf("%s   (Processors): \n", pad)
	for _, p2 := range p.ProcessNodes {
		out += p2.String(fmt.Sprintf("%s      ", pad))
	}
	out += fmt.Sprintf("%s   (Publishers): \n", pad)
	for _, p3 := range p.PublishNodes {
		out += p3.String(fmt.Sprintf("%s      ", pad))
	}
	return out
}

func (p *publishNode) String(args ...string) string {
	pad := ""
	var out string
	if len(args) > 0 {
		pad = args[0]
	}
	out += fmt.Sprintf("%sName: %s\n", pad, p.Name())
	out += fmt.Sprintf("%s   Version: %d\n", pad, p.Version())
	out += fmt.Sprintf("%s   Config:\n", pad)
	for k, v := range p.Config().Table() {
		out += fmt.Sprintf("%s      %s=%+v\n", pad, k, v)
	}
	return out
}
