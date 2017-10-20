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

package passthru

import (
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	debug = "debug"
)

func NewPassthruProcessor() *passthruProcessor {
	return &passthruProcessor{}
}

type passthruProcessor struct{}

func (p *passthruProcessor) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, policy.AddNewBoolRule([]string{""}, debug, false)
}

func (p *passthruProcessor) Process(metrics []plugin.Metric, config plugin.Config) ([]plugin.Metric, error) {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	if _, err := config.GetBool(debug); err == nil {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Debug("processing started")

	if _, err := config.GetBool("test"); err == nil {
		log.Debug("test configuration found")
		for idx, m := range metrics {
			if m.Namespace.Strings()[0] == "foo" {
				log.Print("found foo metric")
				metrics[idx].Data = 2
			}
		}
	}

	return metrics, nil
}
