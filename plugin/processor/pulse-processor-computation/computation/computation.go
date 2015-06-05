package computation

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	name       = "computation"
	version    = 1
	pluginType = plugin.ProcessorPluginType
)

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

func NewComputationPublisher() *computationProcessor {
	return &computationProcessor{}
}

type computationProcessor struct{}

func (p *computationProcessor) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()
	return *config
}

func (p *computationProcessor) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue, logger *log.Logger) (string, []byte, error) {
	logger.Println("Computation Processor started")

	var metrics []plugin.PluginMetricType
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&metrics); err != nil {
		logger.Printf("Error decoding: error=%v content=%v", err, content)
		return "", nil, err
	}

	for i, m := range metrics {
		//Determining the type of data
		switch v := m.Data().(type) {
		default:
			logger.Printf("I am here %T", v)
		case int:
			metrics[i].Data_ = m.Data().(int) * 20
			logger.Printf("I am  int here %v", m.Data())

		case string:
			metrics[i].Data_ = m.Data().(string) + "i am changed"

		}

	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)
	logger.Printf("The change data is : %v", metrics[0].Data())
	return contentType, buf.Bytes(), nil
}
