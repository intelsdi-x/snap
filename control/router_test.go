package control

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName = "pulse-collector-dummy"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", "collector", PluginName)
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
		// Create controller
		c := New()
		c.Start()
		// Load plugin
		c.Load(PluginPath)
		// fmt.Println("\nPlugin Catalog\n*****")
		// for _, p := range c.PluginCatalog() {
		// fmt.Printf("%s %d\n", p.Name(), p.Version())
		// }

		// Create router
		// r := newPluginRouter()
		// r.metricCatalog = c.metricCatalog

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

		// Call collect on router
		cr, err := c.pluginRouter.Collect(m, cd, time.Now().Add(time.Second*60))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("\nresponse: %+v\n", cr)

	})
}
