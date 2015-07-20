package sql

import (
	"bytes"
	"encoding/gob"
	"errors"
	//"os"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSQLPublish(t *testing.T) {
	var buf bytes.Buffer
	metrics := []plugin.PluginMetricType{
		*plugin.NewPluginMetricType([]string{"test_string"}, "example_string"),
		*plugin.NewPluginMetricType([]string{"test_int"}, 1),
		*plugin.NewPluginMetricType([]string{"test_string_slice"}, []string{"str1", "str2"}),
		*plugin.NewPluginMetricType([]string{"test_string_slice"}, []int{1, 2}),
		*plugin.NewPluginMetricType([]string{"test_uint8"}, uint8(1)),
	}
	config := make(map[string]ctypes.ConfigValue)
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)

	Convey("TestSQLPublish", t, func() {
		config["username"] = ctypes.ConfigValueStr{Value: "root"}
		config["password"] = ctypes.ConfigValueStr{Value: "root"}
		config["database"] = ctypes.ConfigValueStr{Value: "PULSE_TEST"}
		config["table name"] = ctypes.ConfigValueStr{Value: "info"}
		sp := NewSQLPublisher()
		So(sp, ShouldNotBeNil)
		err := sp.Publish("", buf.Bytes(), config)
		So(err, ShouldResemble, errors.New("Unknown content type ''"))
		err = sp.Publish(plugin.PulseGOBContentType, buf.Bytes(), config)
		meta := Meta()
		So(meta, ShouldNotBeNil)
	})
}
