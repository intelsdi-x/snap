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

package mock

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

// Mock collector implementation used for testing
type Mock struct {
}

// list of available hosts
var availableHosts = getAllHostnames()

// CollectMetrics collects metrics for testing
func (f *Mock) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	for _, p := range mts {
		log.Printf("collecting %+v\n", p)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	metrics := []plugin.Metric{}
	for i := range mts {
		if _, err := mts[i].Config.GetBool("long_print"); err == nil {
			letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
			longLine := []byte{}
			for i := 0; i < 8193; i++ {
				longLine = append(longLine, letterBytes[rand.Intn(len(letterBytes))])
			}
			fmt.Println(string(longLine))
		}
		if _, err := mts[i].Config.GetBool("panic"); err == nil {
			panic("Oops!")
		}

		if isDynamic, _ := mts[i].Namespace.IsDynamic(); isDynamic {
			requestedHosts := []string{}

			if mts[i].Namespace[2].Value == "*" {
				// when dynamic element is not specified (equals an asterisk)
				// then consider all available hosts as requested hosts
				requestedHosts = append(requestedHosts, availableHosts...)
			} else {
				// when the dynamic element is specified
				// then consider this specified host as requested hosts
				host := mts[i].Namespace[2].Value

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
				ns := plugin.CopyNamespace(mts[i].Namespace)
				ns[2].Value = host
				// metric with set data, ns, timestamp and the version of the plugin
				mt := plugin.Metric{
					Data:      data,
					Namespace: ns,
					Timestamp: time.Now(),
					Unit:      mts[i].Unit,
					Version:   mts[i].Version,
				}
				metrics = append(metrics, mt)
			}

		} else {
			data := randInt(65, 90) + 1000
			mts[i].Data = data
			mts[i].Timestamp = time.Now()
			metrics = append(metrics, mts[i])
		}
	}
	return metrics, nil
}

// GetMetricTypes returns metric types for testing
func (f *Mock) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	mts := []plugin.Metric{}
	if _, err := cfg.GetBool("test-fail"); err == nil {
		return mts, fmt.Errorf("testing")
	}
	if _, err := cfg.GetBool("test"); err == nil {
		mts = append(mts, plugin.Metric{
			Namespace:   plugin.NewNamespace("intel", "mock", "test%>"),
			Description: "mock description",
			Unit:        "mock unit",
		})
	}
	if _, err := cfg.GetBool("test-less"); err != nil {
		mts = append(mts, plugin.Metric{
			Namespace:   plugin.NewNamespace("intel", "mock", "/foo=㊽"),
			Description: "mock description",
			Unit:        "mock unit",
		})
	}
	mts = append(mts, plugin.Metric{
		Namespace:   plugin.NewNamespace("intel", "mock", "/bar⽔"),
		Description: "mock description",
		Unit:        "mock unit",
	})
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("intel", "mock").
			AddDynamicElement("host", "name of the host").
			AddStaticElement("/baz⽔"),
		Description: "mock description",
		Unit:        "mock unit",
	})
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("intel", "mock").
			AddDynamicElement("host", "name of the host").
			AddStaticElements("baz㊽", "/bar⽔"),
		Description: "mock description",
		Unit:        "mock unit",
	})
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("intel", "mock").
			AddDynamicElement("host", "name of the host").
			AddStaticElements("baz㊽", "|barᵹÄ☍"),
		Description: "mock description",
		Unit:        "mock unit",
	})

	return mts, nil
}

// GetConfigPolicy returns a ConfigPolicy for testing
func (f *Mock) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	p := plugin.NewConfigPolicy()

	err := p.AddNewStringRule([]string{"intel", "mock", "test%>"}, "name", false, plugin.SetDefaultString("bob"))
	if err != nil {
		return *p, err
	}

	err = p.AddNewStringRule([]string{"intel", "mock", "/foo=㊽"}, "password", true)

	return *p, err
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
