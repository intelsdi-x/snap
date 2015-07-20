//
// +build integration

package influx

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInfluxPublish(t *testing.T) {
	config := make(map[string]ctypes.ConfigValue)

	Convey("TestInflux", t, func() {
		config["host"] = ctypes.ConfigValueStr{Value: os.Getenv("PULSE_INFLUXDB_HOST")}
		config["port"] = ctypes.ConfigValueInt{Value: 8086}
		config["user"] = ctypes.ConfigValueStr{Value: "root"}
		config["password"] = ctypes.ConfigValueStr{Value: "root"}
		config["database"] = ctypes.ConfigValueStr{Value: "test"}

		ip := NewInfluxPublisher()
		So(ip, ShouldNotBeNil)

		policy := ip.GetConfigPolicyNode()
		So(policy, ShouldNotBeNil)
		cfg, errs := policy.Process(config)
		So(cfg, ShouldNotBeNil)
		So(errs.HasErrors(), ShouldBeFalse)
		So(cfg, ShouldNotBeNil)

		Convey("Publish", func() {
			var buf bytes.Buffer
			metrics := []plugin.PluginMetric{
				*plugin.NewPluginMetric([]string{"foo"}, 99),
			}
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			Convey("int", func() {
				err := ip.PublishType(plugin.ContentTypes[plugin.PulseGobContentType], buf.Bytes(), *cfg, log.New(os.Stdout, "influx_test", log.LstdFlags))
				So(err, ShouldBeNil)
			})

			Convey("float", func() {
				metrics = []plugin.PluginMetric{
					*plugin.NewPluginMetric([]string{"foo"}, 3.141),
				}
				buf.Reset()
				enc = gob.NewEncoder(&buf)
				enc.Encode(metrics)
				err := ip.PublishType(plugin.ContentTypes[plugin.PulseGobContentType], buf.Bytes(), *cfg, log.New(os.Stdout, "influx_test", log.LstdFlags))
				So(err, ShouldBeNil)
			})

			Convey("Unsupported data value error", func() {
				metrics = []plugin.PluginMetric{
					*plugin.NewPluginMetric([]string{"foo"}, "bar"),
				}
				buf.Reset()
				enc = gob.NewEncoder(&buf)
				enc.Encode(metrics)
				err := ip.PublishType(plugin.ContentTypes[plugin.PulseGobContentType], buf.Bytes(), *cfg, log.New(os.Stdout, "influx_test", log.LstdFlags))
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("Unsupported data type 'string'"))

			})
		})

	})
}
