package mysql

import (
	"bytes"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMySQLPublish(t *testing.T) {
	var buf bytes.Buffer
	metrics := []plugin.PluginMetricType{
		*plugin.NewPluginMetricType([]string{"test_string"}, time.Now(), "", "example_string"),
		*plugin.NewPluginMetricType([]string{"test_int"}, time.Now(), "", 1),
		*plugin.NewPluginMetricType([]string{"test_string_slice"}, time.Now(), "", []string{"str1", "str2"}),
		*plugin.NewPluginMetricType([]string{"test_string_slice"}, time.Now(), "", []int{1, 2}),
		*plugin.NewPluginMetricType([]string{"test_uint8"}, time.Now(), "", uint8(1)),
	}
	config := make(map[string]ctypes.ConfigValue)
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)

	Convey("TestMySQLPublish", t, func() {
		config["username"] = ctypes.ConfigValueStr{Value: "root"}
		config["password"] = ctypes.ConfigValueStr{Value: "root"}
		config["database"] = ctypes.ConfigValueStr{Value: "PULSE_TEST"}
		config["table name"] = ctypes.ConfigValueStr{Value: "info"}
		sp := NewMySQLPublisher()
		So(sp, ShouldNotBeNil)
		err := sp.Publish("", buf.Bytes(), config)
		So(err, ShouldResemble, errors.New("Unknown content type ''"))
		err = sp.Publish(plugin.PulseGOBContentType, buf.Bytes(), config)
		meta := Meta()
		So(meta, ShouldNotBeNil)
	})
}
