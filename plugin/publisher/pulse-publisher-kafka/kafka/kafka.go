package kafka

import (
	"fmt"
	"strings"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core/ctypes"

	"gopkg.in/Shopify/sarama.v1"
)

const (
	PluginName    = "kafka"
	PluginVersion = 1
	PluginType    = plugin.PublisherPluginType
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(PluginName, PluginVersion, PluginType)
}

func ConfigPolicyNode() *cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("topic", true)
	handleErr(err)
	r1.Description = "Kafka topic for publishing"

	r2, _ := cpolicy.NewStringRule("brokers", true)
	handleErr(err)
	r2.Description = "List of brokers in the format: broker-ip:port;broker-ip:port (ex: 192.168.1.1:9092;172.16.9.99:9092"

	config.Add(r1, r2)
	return config
}

// func (rmq *rmqPublisher) Publish(data []byte) error {
// 	err := publishDataToRmq(data, rmq.rmqAddress, rmq.rmqExName, rmq.rmqRtKey, rmq.rmqExKind)
// 	return err
// }

type Kafka struct{}

func NewKafkaPublisher() *Kafka {
	var k *Kafka
	//TODO get data from config
	// rmqpub := new(rmqPublisher)
	// rmqpub.rmqAddress = defaultRmqAddress
	// rmqpub.rmqExName = defaultExchangeName
	// rmqpub.rmqExKind = defaultExchangeKind
	// rmqpub.rmqRtKey = defaultRoutingKey
	return k

}

// Publish sends data to a Kafka server
func (k *Kafka) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) {
	//
	topic := config["topic"].(ctypes.ConfigValueStr).Value
	brokers := parseBrokerString(config["brokers"].(ctypes.ConfigValueStr).Value)
	//
	k.publish(topic, brokers, content)
}

// Internal method after data has been converted to serialized bytes to send
func (k *Kafka) publish(topic string, brokers []string, content []byte) {
	// handle config errors
	// ! these should not happen if the config rule is created correctly

	fmt.Println(brokers)

	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		fmt.Println("ERROR:", err)
		panic(err)
	}

	producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(content),
	})
}

func parseBrokerString(brokerStr string) []string {
	return strings.Split(brokerStr, ";")
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

// type rmqPublisher struct {
// rmqAddress string
// rmqExName  string
// rmqExKind  string
// rmqRtKey   string
// }

// const (
// 	name       = "rabbitmq"
// 	version    = 1
// 	pluginType = plugin.PublisherPluginType

// 	defaultRmqAddress   = "127.0.0.1:5672"
// 	defaultExchangeName = "pulse"
// 	defaultExchangeKind = "fanout"
// 	defaultRoutingKey   = "metrics"
// )

// func publishDataToRmq(data []byte, address string, exName string, rtKey string, exKind string) error {
// 	conn, err := amqp.Dial("amqp://" + address)
// 	if err != nil {
// 		log.Printf("RMQ Publisher: cannot open connection, %s", err)
// 		return err
// 	}
// 	defer conn.Close()

// 	c, err := conn.Channel()
// 	if err != nil {
// 		log.Printf("RMQ Publisher: cannot open channel: %s", err)
// 		return err
// 	}

// 	err = c.ExchangeDeclare(exName, exKind, true, false, false, false, nil)
// 	if err != nil {
// 		log.Printf("RMQ Publisher: cannot declare exchange: %v", err)
// 	}

// 	msg := amqp.Publishing{
// 		DeliveryMode: amqp.Persistent,
// 		Timestamp:    time.Now(),
// 		ContentType:  "text/plain",
// 		Body:         data,
// 	}

// 	err = c.Publish(exName, rtKey, false, false, msg)
// 	if err != nil {
// 		log.Printf("RMQ Publisher: cannot publish metric %v", err)
// 		return err
// 	}

// 	return nil
// }
