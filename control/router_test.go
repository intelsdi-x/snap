package control

import (
	"fmt"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	. "github.com/smartystreets/goconvey/convey"
)

type MockMetricType struct {
	namespace []string
}

func (m MockMetricType) Namespace() []string {
	return m.namespace
}

func (m MockMetricType) LastAdvertisedTime() time.Time {
	return time.Now()
}

func (m MockMetricType) Version() int {
	return 1
}

func TestRouter(t *testing.T) {
	Convey("given a new router", t, func() {
		c := New()
		c.Start()
		e := c.Load(PluginPath)
		fmt.Println(c.PluginCatalog())
		fmt.Println(c.MetricCatalog(), e)

		mc := newMetricCatalog()
		r := newRouter(mc)

		m := MockMetricType{namespace: []string{"foo", "bar"}}
		cd := cdata.NewNode()

		r.Collect([]core.MetricType{m}, cd, time.Now().Add(time.Second*60))

	})
}
