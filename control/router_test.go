package control

import (
	"fmt"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
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

func (m MockMetricType) Config() *cdata.ConfigDataNode {
	return nil
}

func TestRouter(t *testing.T) {
	Convey("given a new router", t, func() {
		// adjust HB timeouts for test
		plugin.PingTimeoutLimit = 1
		plugin.PingTimeoutDuration = time.Second * 1

		// Create controller
		c := New()
		c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		// Load plugin
		c.Load(PluginPath)

		m := []core.MetricType{}
		m1 := MockMetricType{namespace: []string{"intel", "dummy", "foo"}}
		m2 := MockMetricType{namespace: []string{"intel", "dummy", "bar"}}
		// m3 := MockMetricType{namespace: []string{"intel", "dummy", "baz"}}
		m = append(m, m1)
		m = append(m, m2)
		// m = append(m, m3)
		cd := cdata.NewNode()
		fmt.Println(cd.Table())

		fmt.Println(m1.Namespace(), m1.Version(), cd)
		// Subscribe
		a, b := c.SubscribeMetricType(m1, cd)
		fmt.Println(a, b)
		time.Sleep(time.Millisecond * 100)
		c.SubscribeMetricType(m2, cd)
		time.Sleep(time.Millisecond * 200)

		// Call collect on router

		for x := 0; x < 5; x++ {
			fmt.Println("\n *  Calling Collect")
			cr, err := c.pluginRouter.Collect(m, cd, time.Now().Add(time.Second*60))
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf(" *  Collect Response: %+v\n", cr)
		}
		time.Sleep(time.Millisecond * 200)
	})
}
