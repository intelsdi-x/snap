/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package rand

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

var (
	strs = []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}
)

func init() {
	rand.Seed(42)
}

// Rand collector implementation used as an example of streaming.
type RandCollector struct {
	metrics []plugin.Metric
}

// StreamMetrics takes both an in and out channel of []plugin.Metric
//
// The metrics_in channel is used to set/update the metrics that Snap is
// currently requesting to be collected by the plugin.
//
// The metrics_out channel is used by the plugin to send the collected metrics
// to Snap.
func (r *RandCollector) StreamMetrics(
	ctx context.Context,
	metrics_in chan []plugin.Metric,
	metrics_out chan []plugin.Metric,
	err chan string) error {

	go r.streamIt(metrics_out, err)
	r.drainMetrics(metrics_in)
	return nil
}

func (r *RandCollector) drainMetrics(in chan []plugin.Metric) {
	for {
		var mts []plugin.Metric
		mts = <-in
		r.metrics = mts
	}
}

func (r *RandCollector) streamIt(ch chan []plugin.Metric, err chan string) {
	for {
		if r.metrics == nil {
			time.Sleep(time.Second)
			continue
		}
		metrics := []plugin.Metric{}
		for idx, mt := range r.metrics {
			r.metrics[idx].Timestamp = time.Now()
			if val, err := mt.Config.GetBool("testbool"); err == nil && val {
				continue
			}
			if mt.Namespace[len(mt.Namespace)-1].Value == "integer" {
				if val, err := mt.Config.GetInt("testint"); err == nil {
					r.metrics[idx].Data = val
				} else {
					r.metrics[idx].Data = rand.Int31()
				}
				metrics = append(metrics, r.metrics[idx])
			} else if mt.Namespace[len(mt.Namespace)-1].Value == "float" {
				if val, err := mt.Config.GetFloat("testfloat"); err == nil {
					r.metrics[idx].Data = val
				} else {
					r.metrics[idx].Data = rand.Float64()
				}
				metrics = append(metrics, r.metrics[idx])
			} else if mt.Namespace[len(mt.Namespace)-1].Value == "string" {
				if val, err := mt.Config.GetString("teststring"); err == nil {
					r.metrics[idx].Data = val
				} else {
					r.metrics[idx].Data = strs[rand.Intn(len(strs)-1)]
				}
				metrics = append(metrics, r.metrics[idx])
			} else {
				err <- fmt.Sprintf("Invalid namespace: %v", mt.Namespace.Strings())
			}
		}
		ch <- metrics
		time.Sleep(time.Second * time.Duration(rand.Int63n(10)))
	}
}

/*
	GetMetricTypes returns metric types for testing.
	GetMetricTypes() will be called when your plugin is loaded in order to populate the metric catalog(where snaps stores all
	available metrics).

	Config info is passed in. This config information would come from global config snap settings.

	The metrics returned will be advertised to users who list all the metrics and will become targetable by tasks.
*/
func (RandCollector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	vals := []string{"integer", "float", "string"}
	for _, val := range vals {
		metric := plugin.Metric{
			Namespace: plugin.NewNamespace("random", val),
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.

	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (RandCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewIntRule([]string{"random", "integer"},
		"testint",
		false,
		plugin.SetMaxInt(1000),
		plugin.SetMinInt(0))

	policy.AddNewFloatRule([]string{"random", "float"},
		"testfloat",
		false,
		plugin.SetMaxFloat(1000.0),
		plugin.SetMinFloat(0.0))

	policy.AddNewStringRule([]string{"random", "string"},
		"teststring",
		false)

	policy.AddNewBoolRule([]string{"random"},
		"testbool",
		false)
	return *policy, nil
}
