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

package pcm

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "pcm"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

// PCM
type PCM struct {
	keys  []string
	data  map[string]interface{}
	mutex *sync.RWMutex
}

func (p *PCM) Keys() []string {
	return p.keys
}

func (p *PCM) Data() map[string]interface{} {
	return p.data
}

// CollectMetrics returns metrics from pcm
func (p *PCM) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := make([]plugin.PluginMetricType, len(mts))
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	hostname, _ := os.Hostname()
	for i, m := range mts {
		if v, ok := p.data[joinNamespace(m.Namespace())]; ok {
			metrics[i] = plugin.PluginMetricType{
				Namespace_: m.Namespace(),
				Data_:      v,
				Source_:    hostname,
				Timestamp_: time.Now(),
			}
		}
	}
	// fmt.Fprintf(os.Stderr, "collected >>> %+v\n", metrics)
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by pcm
func (p *PCM) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	mts := make([]plugin.PluginMetricType, len(p.keys))
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	for i, k := range p.keys {
		mts[i] = plugin.PluginMetricType{Namespace_: strings.Split(strings.TrimPrefix(k, "/"), "/")}
	}
	return mts, nil
}

//GetConfigPolicy
func (p *PCM) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

func New() (*PCM, error) {
	pcm := &PCM{mutex: &sync.RWMutex{}, data: map[string]interface{}{}}
	var cmd *exec.Cmd
	if path := os.Getenv("PULSE_PCM_PATH"); path != "" {
		cmd = exec.Command(filepath.Join(path, "pcm.x"), "/csv", "-nc", "-r", "1")
	} else {
		c, err := exec.LookPath("pcm.x")
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to find PCM.  Ensure it's in your path or set PULSE_PCM_PATH.")
			panic(err)
		}
		cmd = exec.Command(c, "/csv", "-nc", "-r", "1")
	}

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe", err)
		return nil, err
	}
	// read the data from stdout
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		first := true
		for scanner.Scan() {
			if first {
				first = false
				continue
			}
			if len(pcm.keys) == 0 {
				pcm.mutex.Lock()
				keys := strings.Split(strings.TrimSuffix(scanner.Text(), ";"), ";")
				//skip the date and time fields
				pcm.keys = make([]string, len(keys[2:]))
				for i, k := range keys[2:] {
					pcm.keys[i] = fmt.Sprintf("/intel/pcm/%s", k)
				}
				pcm.mutex.Unlock()
				continue
			}

			pcm.mutex.Lock()
			datal := strings.Split(strings.TrimSuffix(scanner.Text(), ";"), ";")
			for i, d := range datal[2:] {
				v, err := strconv.ParseFloat(strings.TrimSpace(d), 64)
				if err == nil {
					pcm.data[pcm.keys[i]] = v
				} else {
					panic(err)
				}
			}
			pcm.mutex.Unlock()
			// fmt.Fprintf(os.Stderr, "data >>> %+v\n", pcm.data)
			// fmt.Fprintf(os.Stdout, "data >>> %+v\n", pcm.data)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting pcm", err)
		return nil, err
	}

	// we need to wait until we have our metric types
	st := time.Now()
	for {
		pcm.mutex.RLock()
		c := len(pcm.keys)
		pcm.mutex.RUnlock()
		if c > 0 {
			break
		}
		if time.Since(st) > time.Second*2 {
			return nil, fmt.Errorf("Timed out waiting for metrics from pcm")
		}
	}

	// LEAVE the following block for debugging
	// err = cmd.Wait()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error waiting for pcm", err)
	// 	return nil, err
	// }

	return pcm, nil
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}
