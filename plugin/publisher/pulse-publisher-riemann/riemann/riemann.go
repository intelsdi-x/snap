/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package riemann

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/amir/raidman"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	PluginName    = "riemann"
	PluginVersion = 1
	PluginType    = plugin.PublisherPluginType
)

// Meta returns the metadata details for the Riemann Publisher Plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(PluginName, PluginVersion, PluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

type Riemann struct{}

// NewRiemannPublisher does something cool
func NewRiemannPublisher() *Riemann {
	var r *Riemann
	return r
}

// GetConfigPolicy returns the config policy for the Riemann Publisher Plugin
func (r *Riemann) GetConfigPolicy() cpolicy.ConfigPolicy {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()
	// Host metric applies to
	r1, err := cpolicy.NewStringRule("host", true)
	handleErr(err)
	r1.Description = "Host the metric was collected from"

	// Riemann server to publish event to
	r2, err := cpolicy.NewStringRule("broker", true)
	handleErr(err)
	r2.Description = "Broker in the format of broker-ip:port (ex: 192.168.1.1:5555)"

	config.Add(r1, r2)
	cp.Add([]string{""}, config)
	return *cp
}

// Publish serializes the data and calls publish to send events to Riemann
func (r *Riemann) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	//err := r.publish(event, broker)
	//return err
	logger.Println("Riemann Publishing Started")
	var metrics []plugin.PluginMetricType
	switch contentType {
	case plugin.PulseGOBContentType:
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
	for _, m := range metrics {
		e := createEvent(m, config)
		if err := r.publish(e, config["broker"].(ctypes.ConfigValueStr).Value); err != nil {
			logger.Println(err)
			return err
		}
	}
	return nil
}

// publish sends events to riemann
func (r *Riemann) publish(event *raidman.Event, broker string) error {
	c, err := raidman.Dial("tcp", broker)
	defer c.Close()
	if err != nil {
		return err
	}
	return c.Send(event)
}

func createEvent(m plugin.PluginMetricType, config map[string]ctypes.ConfigValue) *raidman.Event {
	return &raidman.Event{
		Host:    config["host"].(ctypes.ConfigValueStr).Value,
		Service: strings.Join(m.Namespace(), "/"),
		Metric:  m.Data(),
	}
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
