package movingaverage

import (
	"bytes"
	"encoding/gob"
	"log"

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
	}
}

//The default buffer length is assumed to be 10

func newmovingAverage() *average {
	initCounter := 0
	return &average{
		movingAverageBuf: make([]int, 10),
		movingBufLength:  10,
		counter:          &initCounter,
	}
}

//movingAverageProcessor is a struct which has a map that acts like a buffer for storage of values for different namespace
// key is a namespace (type: string)
//value is a pointer average struct which stores values of the namespace key
type movingAverageProcessor struct {
	movingAverageMap map[string]*average
}

//Each Namespace would have its own buffer-length and counter . Counter is used for the purpose of
//replacing the oldest (when buffer is full) with the new value using mod operation
type average struct {
	movingAverageBuf []int
	movingBufLength  int
	counter          *int
}

//Gets the current counter for the particular namespace
func (p *movingAverageProcessor) getCounter(namespace string) int {
	return *p.movingAverageMap[namespace].counter
}

//Sets the counter for the particular namespace
func (p *movingAverageProcessor) setCounter(namespace string, counter int) error {
	*p.movingAverageMap[namespace].counter = counter
	return nil
}

//Adds data in the buffer for a particular namespace
func (p *movingAverageProcessor) addBufferData(index int, data int, namespace string) error {
	p.movingAverageMap[namespace].movingAverageBuf[p.getCounter(namespace)] = data
	return nil

}

//Retrieves the buffer data for a particular namespace
func (p *movingAverageProcessor) getBufferData(index int, namespace string) int {
	return p.movingAverageMap[namespace].movingAverageBuf[index]
}

//Since namespace is an array of string. Its required to concatenate to make it a primary unique key
func concatNameSpace(namespace []string) string {
	completeNamespace := ""
	for i := 0; i < len(namespace); i++ {
		completeNamespace += namespace[0]
	}
	return completeNamespace
}

func (p *movingAverageProcessor) calculateMovingAverage(m plugin.PluginMetricType) (float64, error) {
	sum := 0
	namespace := concatNameSpace(m.Namespace())

	if movingAverage, ok := p.movingAverageMap[namespace]; ok {
		counterCurrent := p.getCounter(namespace) % movingAverage.movingBufLength
		p.addBufferData(counterCurrent, m.Data().(int), namespace)
		for i := 0; i < movingAverage.movingBufLength; i++ {
			sum += p.getBufferData(i, namespace)
		}
		counterCurrent++
		p.setCounter(namespace, counterCurrent)

	} else {
		//Since map doesnot have an entry of this namespace, its creating an entry for the namespace.
		//Also m.data value is inserted into 0th position of the buffer because we know that this buffer is being used for the first time
		p.movingAverageMap[namespace] = newmovingAverage()
		p.movingAverageMap[namespace].movingAverageBuf[0] = m.Data().(int)
		sum = m.Data().(int)
		*p.movingAverageMap[namespace].counter++
	}

	movingAverage := float64(sum) / float64(p.movingAverageMap[namespace].movingBufLength)

	return movingAverage, nil

}

func (p *movingAverageProcessor) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()
	return *config
}

func (p *movingAverageProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue, logger *log.Logger) (string, []byte, error) {
	logger.Println("movingAverage Processor started")

	var metrics []plugin.PluginMetricType

	//Decodes the content into pluginMetricType
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for i, m := range metrics {
		//Determining the type of data
		switch v := m.Data().(type) {
		default:
			logger.Printf("Unknown Type %T", v)
		case int:
			logger.Printf("Data received %v", metrics[i].Data())

			metrics[i].Data_, _ = p.calculateMovingAverage(m)
			logger.Printf("Moving Average %v", metrics[i].Data())
		}

	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)
	return contentType, buf.Bytes(), nil
}
