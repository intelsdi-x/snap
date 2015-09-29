package riemann

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginMeta(t *testing.T) {
	Convey("Meta returns proper metadata", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, PluginName)
		So(meta.Version, ShouldResemble, PluginVersion)
		So(meta.Type, ShouldResemble, PluginType)
	})
}

func TestConfigPolicy(t *testing.T) {
	Convey("GetConfigPolicy returns non nil object", t, func() {
		r := NewRiemannPublisher()
		c, err := r.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
	})
}
