package kafka

import (
	"strings"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"

	"gopkg.in/Shopify/sarama.v1"
)

const (
	PluginName    = "kafka"
	PluginVersion = 1
	PluginType    = plugin.PublisherPluginType
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(PluginName, PluginVersion, PluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

type Kafka struct{}

func NewKafkaPublisher() *Kafka {
	var k *Kafka
	return k
}

// Publish sends data to a Kafka server
func (k *Kafka) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	topic := config["topic"].(ctypes.ConfigValueStr).Value
	brokers := parseBrokerString(config["brokers"].(ctypes.ConfigValueStr).Value)
	//
	err := k.publish(topic, brokers, content)
	return err
}

func (k *Kafka) GetConfigPolicy() cpolicy.ConfigPolicy {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("topic", true)
	handleErr(err)
	r1.Description = "Kafka topic for publishing"

	r2, _ := cpolicy.NewStringRule("brokers", true)
	handleErr(err)
	r2.Description = "List of brokers in the format: broker-ip:port;broker-ip:port (ex: 192.168.1.1:9092;172.16.9.99:9092"

	config.Add(r1, r2)
	cp.Add([]string{""}, config)
	return *cp
}

// Internal method after data has been converted to serialized bytes to send
func (k *Kafka) publish(topic string, brokers []string, content []byte) error {
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		return err
	}

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(content),
	})
	return err
}

func parseBrokerString(brokerStr string) []string {
	return strings.Split(brokerStr, ";")
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
