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

package mock

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	// Name of plugin
	Name = "mock"
	// Version of plugin
	Version = 2
	// Type of plugin
	Type = plugin.CollectorPluginType
)

// Mock collector implementation used for testing
type Mock struct {
}

// list of available hosts
var availableHosts = getAllHostnames()

// CollectMetrics collects metrics for testing
func (f *Mock) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	for _, p := range mts {
		log.Printf("collecting %+v\n", p)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	metrics := []plugin.MetricType{}
	for i := range mts {
		if c, ok := mts[i].Config().Table()["long_stdout_log"]; ok && c.(ctypes.ConfigValueBool).Value {
			letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			longLine := []byte{}
			for i := 0; i < bufio.MaxScanTokenSize; i++ {
				longLine = append(longLine, letterBytes[rand.Intn(len(letterBytes))])
			}
			fmt.Fprintln(os.Stdout, string(longLine))
		}

		if c, ok := mts[i].Config().Table()["long_stderr_log"]; ok && c.(ctypes.ConfigValueBool).Value {
			letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			longLine := []byte{}
			for i := 0; i < bufio.MaxScanTokenSize; i++ {
				longLine = append(longLine, letterBytes[rand.Intn(len(letterBytes))])
			}
			fmt.Fprintln(os.Stderr, string(longLine))
		}

		if c, ok := mts[i].Config().Table()["panic"]; ok && c.(ctypes.ConfigValueBool).Value {
			panic("Oops!")
		}

		if c, ok := mts[i].Config().Table()["time_to_wait"]; ok {
			dur, err := time.ParseDuration(c.(ctypes.ConfigValueStr).Value)
			if err != nil {
				log.Println(err.Error())
			} else {
				time.Sleep(dur)
			}
		}

		if isDynamic, _ := mts[i].Namespace().IsDynamic(); isDynamic {
			requestedHosts := []string{}

			if mts[i].Namespace()[2].Value == "*" {
				// when dynamic element is not specified (equals an asterisk)
				// then consider all available hosts as requested hosts
				requestedHosts = append(requestedHosts, availableHosts...)
			} else {
				// when the dynamic element is specified
				// then consider this specified host as requested hosts
				host := mts[i].Namespace()[2].Value

				// check if specified host is available in system
				if contains(availableHosts, host) {
					requestedHosts = append(requestedHosts, host)
				} else {
					return nil, fmt.Errorf("requested hostname `%s` is not available (list of available hosts: %s)", host, availableHosts)
				}

			}
			// collect data for each of requested hosts
			for _, host := range requestedHosts {
				//generate random data
				data := randInt(65, 90) + 1000
				// prepare namespace as a copy of incoming dynamic namespace,
				// but with the set value of dynamic element
				ns := make([]core.NamespaceElement, len(mts[i].Namespace()))
				copy(ns, mts[i].Namespace())
				ns[2].Value = host

				// metric with set data, ns, timestamp and the version of the plugin
				mt := plugin.MetricType{
					Data_:      data,
					Namespace_: ns,
					Timestamp_: time.Now(),
					Unit_:      mts[i].Unit(),
					Version_:   mts[i].Version(),
				}
				metrics = append(metrics, mt)
			}
		} else if strings.Contains(mts[i].Namespace().String(), "all") {
			mts[i].Data_ = 1001
			mts[i].Timestamp_ = time.Now()
			metrics = append(metrics, mts[i])
		} else {
			data := randInt(65, 90) + 1000
			mts[i].Data_ = data
			mts[i].Timestamp_ = time.Now()
			metrics = append(metrics, mts[i])
		}
	}
	return metrics, nil
}

// GetMetricTypes returns metric types for testing
func (f *Mock) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}
	if _, ok := cfg.Table()["test-fail"]; ok {
		return mts, fmt.Errorf("testing")
	}
	if _, ok := cfg.Table()["test"]; ok {
		mts = append(mts, plugin.MetricType{
			Namespace_:   core.NewNamespace("intel", "mock", "test"),
			Description_: "mock description",
			Unit_:        "mock unit",
		})
	}
	if _, ok := cfg.Table()["test-less"]; !ok {
		mts = append(mts, plugin.MetricType{
			Namespace_:   core.NewNamespace("intel", "mock", "foo"),
			Description_: "mock description",
			Unit_:        "mock unit",
		})
	}
	mts = append(mts, plugin.MetricType{
		Namespace_:   core.NewNamespace("intel", "mock", "bar"),
		Description_: "mock description",
		Unit_:        "mock unit",
	})
	mts = append(mts, plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "mock").
			AddDynamicElement("host", "name of the host").
			AddStaticElement("baz"),
		Description_: "mock description",
		Unit_:        "mock unit",
	})
	mts = append(mts, plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "mock").
			AddStaticElement("all").
			AddStaticElement("baz"),
		Description_: "mock description",
		Unit_:        "mock unit",
	})

	return mts, nil
}

// GetConfigPolicy returns a ConfigPolicy for testing
func (f *Mock) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	rule, _ := cpolicy.NewStringRule("name", false, "bob")
	rule2, _ := cpolicy.NewStringRule("password", true)
	p := cpolicy.NewPolicyNode()
	p.Add(rule)
	p.Add(rule2)
	c.Add([]string{"intel", "mock", "foo"}, p)
	return c, nil
}

// Meta returns meta data for testing
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		Name,
		Version,
		Type,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.CacheTTL(100*time.Millisecond),
		plugin.RoutingStrategy(plugin.StickyRouting),
	)
}

// contains reports whether a given item is found in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getAllHostnames returns all available hostnames ('host0', 'host1', ..., 'host9')
func getAllHostnames() []string {
	res := []string{}
	for j := 0; j < 10; j++ {
		res = append(res, fmt.Sprintf("host%d", j))
	}
	return res
}

// random number generator
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
