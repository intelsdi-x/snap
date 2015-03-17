/*
# testing
go test -v github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	. "github.com/smartystreets/goconvey/convey"
)

// allows to use fake facter within tests
func withFakeFacter(facter *Facter, mockFacts facts, f func()) func() {

	// getFactsMock
	getFactsMock := func(names []string, _ time.Duration, _ *cmdConfig) (*facts, *time.Time, error) {
		now := time.Now()
		return &mockFacts, &now, nil
	}

	return func() {
		// set mock
		facter.getFacts = getFactsMock
		// set reset function to restore original version of getFacts
		Reset(func() {
			facter.getFacts = getFacts
		})
		f()
	}
}

func TestFacterCollect(t *testing.T) {
	Convey("TestFacterCollect tests", t, func() {

		Convey("Collect executes without error", func() {
			f := NewFacter()
			// ok. even for emtyp request ?
			metricTypes := []plugin.PluginMetricType{}
			metrics, err := f.CollectMetrics(metricTypes)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

		})
	})
}

func TestFacterPluginMeta(t *testing.T) {
	Convey("PluginMeta tests", t, func() {
		meta := Meta()
		Convey("Meta is not nil", func() {
			So(meta, ShouldNotBeNil)
		})
		Convey("Name should be right", func() {
			So(meta.Name, ShouldEqual, "Intel Fact Gathering Plugin")
		})
		Convey("Version should be 1", func() {
			So(meta.Version, ShouldEqual, 1)
		})
		Convey("Type should be plugin.CollectorPluginType", func() {
			So(meta.Type, ShouldEqual, plugin.CollectorPluginType)
		})
	})
}

func TestFacterConfigPolicy(t *testing.T) {
	Convey("config policy has right type", t, func() {
		foo := ConfigPolicyTree()
		bar := cpolicy.NewTree()
		So(foo, ShouldHaveSameTypeAs, bar)
	})
}
