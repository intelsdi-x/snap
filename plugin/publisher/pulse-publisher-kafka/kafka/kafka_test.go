//

package kafka

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

	Convey("ConfigPolicyNode returns non nil object", t, func() {
		k := NewKafkaPublisher()
		c := k.GetConfigPolicyNode()
		So(c, ShouldNotBeNil)
	})
}
