//
// +build integration

package rmq

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"

	"github.com/intelsdi-x/pulse/control/plugin"
)

// integration test
func TestPublish(t *testing.T) {
	Convey("Publish []data in RabbitMQ", t, func() {
		data := []byte("RabbitMQ test string")
		rmqPub := NewRmqPublisher()
		err := rmqPub.Publish(data)
		Convey("No errors are returned from Publish function", func() {
			So(err, ShouldBeNil)
		})
		Convey("We can receive posted message", func() {
			cKill := make(chan struct{})
			cMetrics, err := connectToAmqp(rmqPub, cKill)
			So(err, ShouldBeNil)
			if err != nil {
				t.Fatal("Error while executing tests: cannot connect to AMQP ", err)
			}
			err = rmqPub.Publish(data)
			timeout := time.After(time.Second * 2)
			if err == nil {
				select {
				case metric := <-cMetrics:
					So(data, ShouldResemble, []byte(metric))
					cKill <- struct{}{}
				case _ = <-timeout:
					t.Fatal("Timeout when waiting for AMQP message")
				}
			}
		})

	})
}

func connectToAmqp(rmqpub *rmqPublisher, cKill <-chan struct{}) (chan []byte, error) {
	conn, err := amqp.Dial("amqp://" + rmqpub.rmqAddress)
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
		q.Name,           // queue name
		rmqpub.rmqRtKey,  // routing key
		rmqpub.rmqExName, // exchange
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

func TestConfigPolicyTree(t *testing.T) {

	Convey("ConfigPolicyTree returns non nil object", t, func() {
		ct := ConfigPolicyTree()
		So(ct, ShouldNotBeNil)
	})
}
