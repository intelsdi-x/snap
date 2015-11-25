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

package file

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name       = "file"
	version    = 3
	pluginType = plugin.PublisherPluginType
)

type filePublisher struct {
}

func NewFilePublisher() *filePublisher {
	return &filePublisher{}
}

func (f *filePublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	logger.Println("Publishing started")
	var metrics []plugin.PluginMetricType

	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Printf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Printf("Error unknown content type '%v'", contentType)
		return errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}

	logger.Printf("publishing %v to %v", metrics, config)
	file, err := os.OpenFile(config["file"].(ctypes.ConfigValueStr).Value, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		logger.Printf("Error: %v", err)
		return err
	}
	w := bufio.NewWriter(file)
	for _, m := range metrics {
		source := m.Source()
		if source == "" {
			source = "unknown"
		}
		w.WriteString(fmt.Sprintf("%v|%v|%v|%v\n", m.Timestamp(), m.Namespace(), m.Data(), source))
	}
	w.Flush()

	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func (f *filePublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("file", true)
	handleErr(err)
	r1.Description = "Absolute path to the output file for publishing"

	config.Add(r1)
	cp.Add([]string{""}, config)
	return cp, nil
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
