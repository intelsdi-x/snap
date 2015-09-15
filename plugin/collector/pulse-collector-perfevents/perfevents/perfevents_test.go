package perfevents

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName    = "pulse-collector-perfevents"
	PluginType    = "collector"
	PluginVersion = 1
	PulsePath     = os.Getenv("PULSE_PATH")
	PluginPath    = path.Join(PulsePath, "plugin", PluginName)
)

func TestPluginLoads(t *testing.T) {
	// These tests only work if PULSE_PATH is known.
	// It is the responsibility of the testing framework to
	// build the plugins first into the build dir.
	if PulsePath != "" {
		// Helper plugin trigger build if possible for this plugin
		helper.BuildPlugin(PluginType, PluginName)
		//
		Convey("GetMetricTypes functionality", t, func() {
			p := NewPerfevents()
			Convey("invalid init", func() {
				p.Init = func() error { return errors.New("error") }
				_, err := p.GetMetricTypes()
				So(err, ShouldNotBeNil)
			})
			Convey("set_supported_metrics", func() {
				cg := []string{"cgroup1", "cgroup2", "cgroup3"}
				events := []string{"event1", "event2", "event3"}
				a := set_supported_metrics(ns_subtype, cg, events)
				So(a[len(a)-1].Namespace_, ShouldResemble, []string{ns_vendor, ns_class, ns_type, ns_subtype, "event3", "cgroup3"})
			})
			Convey("flatten cgroup name", func() {
				cg := []string{"cg_root/cg_sub1/cg_sub2"}
				events := []string{"event"}
				a := set_supported_metrics(ns_subtype, cg, events)
				So(a[len(a)-1].Namespace_, ShouldContain, "cg_root_cg_sub1_cg_sub2")
			})
		})
		Convey("CollectMetrics error cases", t, func() {
			p := NewPerfevents()
			Convey("empty list of requested metrics", func() {
				metricTypes := []plugin.PluginMetricType{}
				metrics, err := p.CollectMetrics(metricTypes)
				So(err, ShouldBeNil)
				So(metrics, ShouldBeEmpty)
			})
			Convey("namespace too short", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{"invalid"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "segments")
			})
			Convey("namespace wrong vendor", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{"invalid", ns_class, ns_type, ns_subtype, "cycles", "A"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "1st")
			})
			Convey("namespace wrong class", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{ns_vendor, "invalid", ns_type, ns_subtype, "cycles", "A"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "2nd")
			})
			Convey("namespace wrong type", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{ns_vendor, ns_class, "invalid", ns_subtype, "cycles", "A"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "3rd")
			})
			Convey("namespace wrong subtype", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{ns_vendor, ns_class, ns_type, "invalid", "cycles", "A"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "4th")
			})
			Convey("namespace wrong event", func() {
				_, err := p.CollectMetrics(
					[]plugin.PluginMetricType{
						plugin.PluginMetricType{
							Namespace_: []string{ns_vendor, ns_class, ns_type, ns_subtype, "invalid", "A"},
						},
					},
				)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "5th")
			})

		})
	} else {
		fmt.Printf("PULSE_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
}
