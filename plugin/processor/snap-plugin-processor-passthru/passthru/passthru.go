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

package passthru

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name       = "passthru"
	version    = 1
	pluginType = plugin.ProcessorPluginType
	debug      = "debug"
)

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func NewPassthruProcessor() *passthruProcessor {
	return &passthruProcessor{}
}

type passthruProcessor struct{}

func (p *passthruProcessor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewBoolRule(debug, false)
	if err != nil {
		panic(err)
	}
	r1.Description = "Debug mode"
	config.Add(r1)

	cp.Add([]string{""}, config)
	return cp, nil
}

func (p *passthruProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	if val, ok := config[debug]; ok && val.(ctypes.ConfigValueBool).Value {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Debug("processing started")

	// The following block is for testing config see.. control_test.go
	if _, ok := config["test"]; ok {
		log.Debug("test configuration found")
		var metrics []plugin.MetricType
		//Decodes the content into MetricType
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"content": content,
			}).Errorf("error decoding")
			return "", nil, err
		}

		for idx, m := range metrics {
			if m.Namespace()[0].Value == "foo" {
				log.Print("found foo metric")
				metrics[idx].Data_ = 2
			}
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(metrics)
		content = buf.Bytes()
	}

	//just passing through
	return contentType, content, nil
}
