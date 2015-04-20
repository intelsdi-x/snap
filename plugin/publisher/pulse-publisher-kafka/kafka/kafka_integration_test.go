//
// +bbbuild integration

package kafka

import (
	"fmt"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

// integration test
func TestPublish(t *testing.T) {
	Convey("Publish to Kafka", t, func() {
		k := NewKafkaPublisher()

		// Build some config
		cdn := cdata.NewNode()
		cdn.AddItem("brokers", ctypes.ConfigValueStr{"172.16.125.132:9092"})
		cdn.AddItem("topic", ctypes.ConfigValueStr{"test"})

		//
		p := ConfigPolicyNode()
		f, err := p.Process(cdn.Table())
		So(getProcessErrorStr(err), ShouldEqual, "")

		k.Publish("", []byte(time.Now().String()), *f)
	})
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
