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

package movingaverage

import (
	"bytes"
	"encoding/gob"
	"errors"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	name       = "movingaverage"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

func NewMovingaverageProcessor() *movingAverageProcessor {
	a := make(map[string]*average)
	return &movingAverageProcessor{
		movingAverageMap: a,
		movingBufLength:  10,
	}
}

//The default buffer length is assumed to be 10

func newmovingAverage(length int) *average {
	initCounter := 0
	return &average{
		movingAverageBuf: make([]interface{}, length),
		counter:          initCounter,
	}
}

//movingAverageProcessor is a struct which has a map that acts like a buffer for storage of values for different namespace
// key is a namespace (type: string)
//value is a pointer average struct which stores values of the namespace key
type movingAverageProcessor struct {
	movingAverageMap map[string]*average
	movingBufLength  int
}

//Each Namespace would have its own buffer-length and counter . Counter is used for the purpose of
//replacing the oldest (when buffer is full) with the new value using mod operation
type average struct {
	movingAverageBuf []interface{}
	counter          int
}

//Gets the current counter for the particular namespace
func (p *movingAverageProcessor) getCounter(namespace string) (int, error) {
	if _, ok := p.movingAverageMap[namespace]; ok {
		return p.movingAverageMap[namespace].counter, nil
	} else {
		return -1, errors.New("Namespace is not present in the map")
	}
}

//Sets the counter for the particular namespace
func (p *movingAverageProcessor) setCounter(namespace string, counter int) error {
	if _, ok := p.movingAverageMap[namespace]; ok {
		p.movingAverageMap[namespace].counter = counter
		return nil
	} else {
		return errors.New("Namespace is not present in the map")
	}
}

func (p *movingAverageProcessor) getBufferLength() int {
	return p.movingBufLength
}

func (p *movingAverageProcessor) setBufferLength(length int) error {
	p.movingBufLength = length
	return nil
}

//Adds data in the buffer for a particular namespace
func (p *movingAverageProcessor) addBufferData(index int, data interface{}, namespace string) error {

	if _, ok := p.movingAverageMap[namespace]; ok {
		counter, _ := p.getCounter(namespace)
		p.movingAverageMap[namespace].movingAverageBuf[counter] = data
		return nil
	} else {
		return errors.New("Namespace is not present in the map")
	}
}

//Retrieves the buffer data for a particular namespace
func (p *movingAverageProcessor) getBufferData(index int, namespace string) interface{} {

	return p.movingAverageMap[namespace].movingAverageBuf[index]
}

//Since namespace is an array of string. Its required to concatenate to make it a primary unique key
func concatNameSpace(namespace []string) string {
	completeNamespace := ""
	for i := 0; i < len(namespace); i++ {
		completeNamespace += namespace[i]
	}
	return completeNamespace
}

func (p *movingAverageProcessor) calculateMovingAverage(m plugin.PluginMetricType, logger *log.Logger) (float64, error) {

	namespace := concatNameSpace(m.Namespace())
	switch v := m.Data().(type) {
	default:
		logger.Printf("Unknown data received: Type %T", v)
		return 0.0, errors.New("Unknown data received: Type")
	case int:
		if _, ok := p.movingAverageMap[namespace]; ok {
			counter, err := p.getCounter(namespace)
			counterCurrent := counter % p.movingBufLength
			p.addBufferData(counterCurrent, m.Data(), namespace)
			sum := int(0)
			//Initial Counter is used to give correct average for initial iterations ie when the buffer is not full
			initialCounter := 0
			for i := 0; i < p.movingBufLength; i++ {
				if p.getBufferData(i, namespace) != nil {
					initialCounter++
					sum += p.getBufferData(i, namespace).(int)
				}
			}
			movingAvg := float64(sum) / float64(initialCounter)
			counterCurrent++
			p.setCounter(namespace, counterCurrent)
			return movingAvg, err

		} else {

			//Since map doesnot have an entry of this namespace, its creating an entry for the namespace.
			//Also m.data value is inserted into 0th position of the buffer because we know that this buffer is being used for the first time
			p.movingAverageMap[namespace] = newmovingAverage(p.getBufferLength())
			p.addBufferData(0, m.Data(), namespace)
			sum := p.getBufferData(0, namespace).(int)
			p.setCounter(namespace, 1)
			return float64(sum), nil
		}

	case float64:

		if _, ok := p.movingAverageMap[namespace]; ok {
			counter, err := p.getCounter(namespace)
			counterCurrent := counter % p.movingBufLength
			p.addBufferData(counterCurrent, m.Data(), namespace)
			sum := float64(0)
			initialCounter := 0
			for i := 0; i < p.movingBufLength; i++ {
				if p.getBufferData(i, namespace) != nil {
					initialCounter++
					sum += p.getBufferData(i, namespace).(float64)
				}
			}
			movingAvg := float64(sum) / float64(initialCounter)
			counterCurrent++
			p.setCounter(namespace, counterCurrent)
			return movingAvg, err

		} else {
			p.movingAverageMap[namespace] = newmovingAverage(p.getBufferLength())
			p.addBufferData(0, m.Data(), namespace)
			sum := p.getBufferData(0, namespace).(float64)
			p.setCounter(namespace, 1)
			return float64(sum), nil
		}

	case float32:
		if _, ok := p.movingAverageMap[namespace]; ok {
			counter, err := p.getCounter(namespace)
			counterCurrent := counter % p.movingBufLength
			p.addBufferData(counterCurrent, m.Data(), namespace)
			sum := float32(0)

			initialCounter := 0
			for i := 0; i < p.movingBufLength; i++ {
				if p.getBufferData(i, namespace) != nil {
					initialCounter++
					sum += p.getBufferData(i, namespace).(float32)
				}
			}
			movingAvg := float64(sum) / float64(initialCounter)
			p.setCounter(namespace, counterCurrent)
			return movingAvg, err

		} else {
			p.movingAverageMap[namespace] = newmovingAverage(p.getBufferLength())
			p.addBufferData(0, m.Data(), namespace)
			sum := p.getBufferData(0, namespace).(float32)
			p.setCounter(namespace, 1)
			return float64(sum), nil
		}

	case uint32:
		if _, ok := p.movingAverageMap[namespace]; ok {
			counter, err := p.getCounter(namespace)
			counterCurrent := counter % p.movingBufLength
			p.addBufferData(counterCurrent, m.Data(), namespace)
			sum := uint32(0)
			initialCounter := 0
			for i := 0; i < p.movingBufLength; i++ {
				if p.getBufferData(i, namespace) != nil {
					initialCounter++
					sum += p.getBufferData(i, namespace).(uint32)
				}
			}
			movingAvg := float64(sum) / float64(initialCounter)
			counterCurrent++
			p.setCounter(namespace, counterCurrent)
			return movingAvg, err

		} else {
			p.movingAverageMap[namespace] = newmovingAverage(p.getBufferLength())
			p.addBufferData(0, m.Data(), namespace)
			sum := p.getBufferData(0, namespace).(uint32)
			p.setCounter(namespace, 1)
			return float64(sum), nil
		}

	case uint64:
		if _, ok := p.movingAverageMap[namespace]; ok {
			counter, err := p.getCounter(namespace)
			counterCurrent := counter % p.movingBufLength
			p.addBufferData(counterCurrent, m.Data(), namespace)
			sum := uint64(0)
			initialCounter := 0
			for i := 0; i < p.movingBufLength; i++ {
				if p.getBufferData(i, namespace) != nil {
					initialCounter++
					sum += p.getBufferData(i, namespace).(uint64)
				}
			}
			movingAvg := float64(sum) / float64(initialCounter)
			counterCurrent++
			p.setCounter(namespace, counterCurrent)
			return movingAvg, err

		} else {
			p.movingAverageMap[namespace] = newmovingAverage(p.getBufferLength())
			p.addBufferData(0, m.Data(), namespace)
			sum := p.getBufferData(0, namespace).(uint64)
			p.setCounter(namespace, 1)
			return float64(sum), nil
		}

	}
}
func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

func (p *movingAverageProcessor) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()
	r1, err := cpolicy.NewIntegerRule("MovingAvgBufLength", true)
	handleErr(err)
	r1.Description = "Buffer Length for moving average "
	config.Add(r1)
	cp.Add([]string{""}, config)
	return cp, nil
}

func (p *movingAverageProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	logger := log.New()
	logger.Println("movingAverage Processor started")

	var metrics []plugin.PluginMetricType
	//if the MovingAvgBufLength is set to number less than or equal to 0 then the MovingAvgBufferLength is set to 10
	if config != nil {
		if config["MovingAvgBufLength"].(ctypes.ConfigValueInt).Value > 0 {
			p.setBufferLength(config["MovingAvgBufLength"].(ctypes.ConfigValueInt).Value)

		} else {
			p.setBufferLength(10)
		}

	} else {
		p.setBufferLength(10)
	}

	//Decodes the content into pluginMetricType
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for i, m := range metrics {
		//Determining the type of data
		logger.Printf("Data received %v", metrics[i].Data())

		metrics[i].Data_, _ = p.calculateMovingAverage(m, logger)
		logger.Printf("Moving Average %v", metrics[i].Data())

	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)
	return contentType, buf.Bytes(), nil
}
