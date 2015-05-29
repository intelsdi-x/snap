package psutil

import (
	"runtime"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPsutilCollectMetrics(t *testing.T) {
	Convey("psutil collector", t, func() {
		p := &Psutil{}
		Convey("collect metrics", func() {
			mts := []plugin.PluginMetricType{
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load1"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load5"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "load", "load15"},
				},
				plugin.PluginMetricType{
					Namespace_: []string{"psutil", "vm", "total"},
				},
			}
			if runtime.GOOS != "darwin" {
				mts = append(mts, plugin.PluginMetricType{
					Namespace_: []string{"psutil", "cpu0", "user"},
				})
			}
			metrics, err := p.CollectMetrics(mts)
			//prettyPrint(metrics)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
		})
		Convey("get metric types", func() {
			mts, err := p.GetMetricTypes()
			//prettyPrint(mts)
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
		})

	})
}
