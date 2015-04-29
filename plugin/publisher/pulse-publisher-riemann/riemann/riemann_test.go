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

func TestConfigPolicyNode(t *testing.T) {
	Convey("GetConfigPolicyNode returns non nil object", t, func() {
		r := NewRiemannPublisher()
		c := r.GetConfigPolicyNode()
		So(c, ShouldNotBeNil)
	})
}
