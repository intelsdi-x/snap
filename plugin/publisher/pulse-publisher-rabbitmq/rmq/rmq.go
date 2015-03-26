package rmq

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/streadway/amqp"
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType)
}

func ConfigPolicyTree() *cpolicy.ConfigPolicyTree {
	c := cpolicy.NewTree()
	return c
}

func (rmq *RmqPublisher) PublishMetrics(metrics []plugin.PluginMetric) error {

	marshalledData, err := json.Marshal(metrics)
	if err != nil {
		errString := "RMQ Publisher: Error while marshalling data"
		log.Println(errString)
		return errors.New(errString)
	}

	data := []byte(marshalledData)
	err = publishDataToRmq(data, rmq.rmqAddress, rmq.rmqExName, rmq.rmqRtKey, rmq.rmqExKind)
	return err
}

func NewRmqPublisher() *RmqPublisher {
	//TODO get data from config
	rmqpub := new(RmqPublisher)
	rmqpub.rmqAddress = defaultRmqAddress
	rmqpub.rmqExName = defaultExchangeName
	rmqpub.rmqExKind = defaultExchangeKind
	rmqpub.rmqRtKey = defaultRoutingKey
	return rmqpub

}

type RmqPublisher struct {
	rmqAddress string
	rmqExName  string
	rmqExKind  string
	rmqRtKey   string
}

var _ plugin.PublisherPlugin = (*RmqPublisher)(nil)

const (
	name       = "Intel RMQ Publisher Plugin (c) 2015 Intel Corporation"
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
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		log.Printf("RMQ Publisher: cannot publish metric %v", err)
		return err
	}

	return nil
}
