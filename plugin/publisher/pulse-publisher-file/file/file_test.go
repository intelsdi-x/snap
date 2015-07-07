package file

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublish(t *testing.T) {
	var buf bytes.Buffer
	metrics := []plugin.PluginMetricType{
		*plugin.NewPluginMetricType([]string{"foo"}, 99),
	}
	config := make(map[string]ctypes.ConfigValue)
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)

	Convey("TestFilePublish", t, func() {
		config["file"] = ctypes.ConfigValueStr{Value: "/tmp/pub.out"}
		fp := NewFilePublisher()
		So(fp, ShouldNotBeNil)
		err := fp.Publish("", buf.Bytes(), config)
		So(err, ShouldResemble, errors.New("Unknown content type ''"))
		err = fp.Publish(plugin.PulseGOBContentType, buf.Bytes(), config)
		So(err, ShouldBeNil)
		_, err = os.Stat(config["file"].(ctypes.ConfigValueStr).Value)
		So(err, ShouldBeNil)
		meta := Meta()
		So(meta, ShouldNotBeNil)
	})
}
