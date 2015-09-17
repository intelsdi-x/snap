package rmq

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/streadway/amqp"
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

type rmqPublisher struct{}

func NewRmqPublisher() *rmqPublisher {
	return &rmqPublisher{}

}

const (
	name       = "rabbitmq"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

func (rmq *rmqPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
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
		return fmt.Errorf("Unknown content type '%s'", contentType)
	}
	err := publishDataToRmq(
		metrics,
		config["address"].(ctypes.ConfigValueStr).Value,
		config["exchange_name"].(ctypes.ConfigValueStr).Value,
		config["routing_key"].(ctypes.ConfigValueStr).Value,
		config["exchange_type"].(ctypes.ConfigValueStr).Value,
		logger,
	)
	return err
}

func (rmq *rmqPublisher) GetConfigPolicy() cpolicy.ConfigPolicy {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("address", true)
	handleErr(err)
	r1.Description = "RabbitMQ Address (host:port)"
	config.Add(r1)

	r2, err := cpolicy.NewStringRule("exchange_name", true)
	handleErr(err)
	r2.Description = "RabbitMQ Exchange Name"
	config.Add(r2)

	r3, err := cpolicy.NewStringRule("exchange_type", true)
	handleErr(err)
	r3.Description = "RabbitMQ Exchange Type"
	config.Add(r3)

	r4, err := cpolicy.NewStringRule("routing_key", true)
	handleErr(err)
	r4.Description = "RabbitMQ Routing Key"
	config.Add(r4)

	cp.Add([]string{""}, config)
	return *cp
}

func publishDataToRmq(metrics []plugin.PluginMetricType, address string, exName string, rtKey string, exKind string, logger *log.Logger) error {
	conn, err := amqp.Dial("amqp://" + address)
	if err != nil {
		logger.Printf("RMQ Publisher: cannot open connection, %s", err)
		return err
	}
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		logger.Printf("RMQ Publisher: cannot open channel: %s", err)
		return err
	}

	err = c.ExchangeDeclare(exName, exKind, true, false, false, false, nil)
	if err != nil {
		logger.Printf("RMQ Publisher: cannot declare exchange: %v", err)
		return err
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.Printf("RMQ Publisher: cannot marshal metrics: %v", err)
		return err
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         data,
	}

	err = c.Publish(exName, rtKey, false, false, msg)
	if err != nil {
		logger.Printf("RMQ Publisher: cannot publish metric %v", err)
		return err
	}

	return nil
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
