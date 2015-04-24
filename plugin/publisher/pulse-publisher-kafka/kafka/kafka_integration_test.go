package kafka

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/Shopify/sarama.v1"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// integration test
func TestPublish(t *testing.T) {
	// This integration test requires KAFKA_BROKERS
	if os.Getenv("PULSE_TEST_KAFKA") == "" {
		fmt.Println("Skipping integration")
		return
	}
	brokers := os.Getenv("PULSE_TEST_KAFKA")

	// Pick a random topic
	topic := fmt.Sprintf("%d", time.Now().Nanosecond())
	fmt.Printf("Topic: %s\n", topic)

	Convey("Publish to Kafka", t, func() {
		Convey("publish and consume", func() {
			k := NewKafkaPublisher()

			// Build some config
			cdn := cdata.NewNode()
			cdn.AddItem("brokers", ctypes.ConfigValueStr{Value: brokers})
			cdn.AddItem("topic", ctypes.ConfigValueStr{Value: topic})

			// Get validated policy
			p := k.GetConfigPolicyNode()
			f, err := p.Process(cdn.Table())
			So(getProcessErrorStr(err), ShouldEqual, "")

			t := time.Now().String()

			// Send data to create topic. There is a weird bug where first message won't be consumed
			// ref: http://mail-archives.apache.org/mod_mbox/kafka-users/201411.mbox/%3CCAHwHRrVmwyJg-1eyULkzwCUOXALuRA6BqcDV-ffSjEQ+tmT7dw@mail.gmail.com%3E
			k.Publish("", []byte(t), *f, log.New(os.Stdout, "kafka-integration-test", log.LstdFlags))
			// Send the same message. This will be consumed.
			k.Publish("", []byte(t), *f, log.New(os.Stdout, "kafka-integration-test", log.LstdFlags))

			// start timer and wait
			m, mErr := consumer(topic, brokers)
			So(mErr, ShouldBeNil)
			So(m, ShouldNotBeNil)
			So(string(m.Value), ShouldEqual, t)
		})

		Convey("error on bad broker l", func() {

		})

	})
}

func consumer(topic, brokers string) (*sarama.ConsumerMessage, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true // Handle errors manually instead of letting Sarama log them.

	master, err := sarama.NewConsumer([]string{brokers}, config)
	if err != nil {
		return nil, err
	}
	defer func() {
		master.Close()
	}()

	consumer, err := master.ConsumePartition(topic, 0, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		consumer.Close()
	}()

	msgCount := 0
	timeout := make(chan bool, 1)

	go func() {
		time.Sleep(time.Millisecond * 500)
		timeout <- true
	}()

	select {
	case err := <-consumer.Errors():
		return nil, err
	case m := <-consumer.Messages():
		msgCount++
		return m, nil
	case <-timeout:
		return nil, errors.New("timed out waiting for produced message")
	}
}

func getProcessErrorStr(err *cpolicy.ProcessingErrors) string {
	errString := ""

	if err.HasErrors() {
		for _, e := range err.Errors() {
			errString += fmt.Sprintln(e.Error())
		}
	}
	return errString
}
