package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/intelsdilabs/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublish(t *testing.T) {
	metrics := []byte("does not matter")
	config := make(map[string]ctypes.ConfigValue)

	Convey("TestFilePublish", t, func() {
		fp := NewFilePublisher()
		So(fp, ShouldNotBeNil)
		err := fp.Publish("", metrics, config)
		So(err, ShouldBeNil)
		_, err = os.Stat(filepath.Join(fp.path, fp.name))
		So(err, ShouldBeNil)
		meta := Meta()
		So(meta, ShouldNotBeNil)
		defaultPath = "/does/not/exist"
		fp2 := NewFilePublisher()
		err = fp2.Publish("", metrics, config)
		So(err, ShouldNotBeNil)
	})
}
