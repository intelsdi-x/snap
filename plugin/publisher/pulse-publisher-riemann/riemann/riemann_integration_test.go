package riemann

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/amir/raidman"
)

// integration test
func TestPublish(t *testing.T) {
	// This integration test requires a Riemann Server
	broker := os.Getenv("PULSE_TEST_RIEMANN")
	if broker == "" {
		fmt.Println("Skipping integration tests")
		return
	}

	Convey("Publish to Riemann", t, func() {
		Convey("publish and consume", func() {
			r := NewRiemannPublisher()
			cdn := cdata.NewNode()
			cdn.AddItem("broker", ctypes.ConfigValueStr{Value: broker})
			cdn.AddItem("host", ctypes.ConfigValueStr{Value: "bacon-powered"})
			p := r.GetConfigPolicyNode()
			f, cErr := p.Process(cdn.Table())
			So(getProcessErrorStr(cErr), ShouldEqual, "")

			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"intel", "cpu", "temp"}, 100),
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := r.Publish(plugin.PulseGOBContentType, buf.Bytes(), *f, log.New(os.Stdout, "", log.LstdFlags))
			So(err, ShouldBeNil)

			c, _ := raidman.Dial("tcp", broker)
			events, _ := c.Query("host = \"bacon-powered\"")
			So(len(events), ShouldBeGreaterThan, 0)
		})
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
