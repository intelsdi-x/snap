package rmq

import (
	"log"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/streadway/amqp"
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType)
}

func ConfigPolicyTree() *cpolicy.ConfigPolicyTree {
	c := cpolicy.NewTree()
	return c
}

func (rmq *rmqPublisher) Publish(data []byte) error {
	err := publishDataToRmq(data, rmq.rmqAddress, rmq.rmqExName, rmq.rmqRtKey, rmq.rmqExKind)
	return err
}

func NewRmqPublisher() *rmqPublisher {
	//TODO get data from config
	rmqpub := new(rmqPublisher)
	rmqpub.rmqAddress = defaultRmqAddress
	rmqpub.rmqExName = defaultExchangeName
	rmqpub.rmqExKind = defaultExchangeKind
	rmqpub.rmqRtKey = defaultRoutingKey
	return rmqpub

}

type rmqPublisher struct {
	rmqAddress string
	rmqExName  string
	rmqExKind  string
	rmqRtKey   string
}

const (
	name       = "rabbitmq"
	version    = 1
	pluginType = plugin.PublisherPluginType

	defaultRmqAddress   = "127.0.0.1:5672"
	defaultExchangeName = "pulse"
	defaultExchangeKind = "fanout"
	defaultRoutingKey   = "metrics"
)

func publishDataToRmq(data []byte, address string, exName string, rtKey string, exKind string) error {
	conn, err := amqp.Dial("amqp://" + address)
	if err != nil {
		log.Printf("RMQ Publisher: cannot open connection, %s", err)
		return err
	}
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Printf("RMQ Publisher: cannot open channel: %s", err)
		return err
	}

	err = c.ExchangeDeclare(exName, exKind, true, false, false, false, nil)
	if err != nil {
		log.Printf("RMQ Publisher: cannot declare exchange: %v", err)
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         data,
	}

	err = c.Publish(exName, rtKey, false, false, msg)
	if err != nil {
		log.Printf("RMQ Publisher: cannot publish metric %v", err)
		return err
	}

	return nil
}
