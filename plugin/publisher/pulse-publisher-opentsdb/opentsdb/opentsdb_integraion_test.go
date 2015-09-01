package opentsdb

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOpentsdbPublish(t *testing.T) {
	config := make(map[string]ctypes.ConfigValue)

	Convey("TestOpentsdb", t, func() {
		config["host"] = ctypes.ConfigValueStr{Value: os.Getenv("PULSE_OPENTSDB_HOST")}
		config["port"] = ctypes.ConfigValueInt{Value: 4242}

		ip := NewOpentsdbPublisher()
		So(ip, ShouldNotBeNil)

		policy := ip.GetConfigPolicy()
		So(policy, ShouldNotBeNil)

		Convey("Publish", func() {
			var buf bytes.Buffer
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"/psutil/load/load15"}, time.Now(), "mac1", 23.1),
				*plugin.NewPluginMetricType([]string{"/psutil/vm/available"}, time.Now().Add(2*time.Second), "mac2", 23.2),
				*plugin.NewPluginMetricType([]string{"/psutil/load/load1"}, time.Now().Add(3*time.Second), "linux3", 23.3),
			}
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			Convey("float", func() {
				err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), config)
				So(err, ShouldBeNil)
			})

			Convey("int", func() {
				metrics = []plugin.PluginMetricType{
					*plugin.NewPluginMetricType([]string{"/psutil/vm/free"}, time.Now().Add(5*time.Second), "linux7", 23),
				}
				buf.Reset()
				enc = gob.NewEncoder(&buf)
				enc.Encode(metrics)

				err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), config)
				So(err, ShouldBeNil)
			})
		})

	})
}
