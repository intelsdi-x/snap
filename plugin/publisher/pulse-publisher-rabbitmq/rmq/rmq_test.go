//
// +build integration

package rmq

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// integration test
func TestPublish(t *testing.T) {
	Convey("Publish data in RabbitMQ", t, func() {
		mt := plugin.PluginMetricType{
			Namespace_:          []string{"foo", "bar"},
			LastAdvertisedTime_: time.Now(),
			Version_:            1,
			Data_:               1,
		}
		data, _, err := plugin.MarshallPluginMetricTypes(plugin.PulseGOBContentType, []plugin.PluginMetricType{mt})
		So(err, ShouldBeNil)
		rmqPub := NewRmqPublisher()
		config := map[string]ctypes.ConfigValue{
			"address":       ctypes.ConfigValueStr{Value: "127.0.0.1:5672"},
			"exchange_name": ctypes.ConfigValueStr{Value: "pulse"},
			"routing_key":   ctypes.ConfigValueStr{Value: "metrics"},
			"exchange_type": ctypes.ConfigValueStr{Value: "fanout"},
		}
		logger := log.New(os.Stdout, "", log.LstdFlags)
		err = rmqPub.Publish(plugin.PulseGOBContentType, data, config, logger)
		So(err, ShouldBeNil)
		Convey("We can receive posted message", func() {
			cKill := make(chan struct{})
			cMetrics, err := connectToAmqp(cKill)
			So(err, ShouldBeNil)
			timeout := time.After(time.Second * 2)
			if err == nil {
				select {
				case metric := <-cMetrics:
					var metrix []plugin.PluginMetricType
					err := json.Unmarshal(metric, &metrix)
					So(err, ShouldBeNil)
					So(metrix[1].Version, ShouldEqual, 1)
					cKill <- struct{}{}
				case <-timeout:
					t.Fatal("Timeout when waiting for AMQP message")
				}
			}
		})
	})
}

func connectToAmqp(cKill <-chan struct{}) (chan []byte, error) {
	conn, err := amqp.Dial("amqp://127.0.0.1:5672")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	//	FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,    // queue name
		"metrics", // routing key
		"pulse",
		false,
		nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	cMetrics := make(chan []byte)
	go func() {
		for {
			select {
			case msg := <-msgs:
				cMetrics <- msg.Body

			case _ = <-cKill:
				conn.Close()
				ch.Close()
				return
			}
		}
	}()
	return cMetrics, nil

}

func TestPluginMeta(t *testing.T) {

	Convey("Meta returns proper metadata", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, name)
		So(meta.Version, ShouldResemble, version)
		So(meta.Type, ShouldResemble, plugin.PublisherPluginType)
	})
}
