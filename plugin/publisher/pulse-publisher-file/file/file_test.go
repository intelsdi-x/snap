package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublish(t *testing.T) {
	metrics := []plugin.PluginMetric{
		plugin.PluginMetric{
			Namespace_: []string{"foo", "bar"},
			Data_:      "val0",
		},
	}
	Convey("TestFilePublish", t, func() {
		fp := NewFilePublisher()
		So(fp, ShouldNotBeNil)
		err := fp.Publish(metrics)
		So(err, ShouldBeNil)
		_, err = os.Stat(filepath.Join(fp.path, fp.name))
		So(err, ShouldBeNil)
		meta := Meta()
		So(meta, ShouldNotBeNil)
		defaultPath = "/does/not/exist"
		fp2 := NewFilePublisher()
		err = fp2.Publish(metrics)
		So(err, ShouldNotBeNil)
	})
}
