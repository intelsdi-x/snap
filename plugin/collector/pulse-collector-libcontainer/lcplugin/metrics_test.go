package lcplugin

import (
	"testing"
	"time"

	"github.com/docker/libcontainer"
	"github.com/docker/libcontainer/network"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetNetMetrics(t *testing.T) {

	Convey("TestGetNetMetrics with container.json fixture", t, func() {

		containerId := "1234abc"
		var stats libcontainer.ContainerStats
		stats.NetworkStats = new(network.NetworkStats)

		stats.NetworkStats.TxBytes = 1
		stats.NetworkStats.RxBytes = 2
		stats.NetworkStats.TxPackets = 3
		stats.NetworkStats.RxPackets = 4
		stats.NetworkStats.TxDropped = 5
		stats.NetworkStats.RxDropped = 6
		stats.NetworkStats.TxErrors = 7
		stats.NetworkStats.RxErrors = 8

		timestamp := time.Now()
		cb := getNetMetrics(containerId, &stats, timestamp)

		So(cb.namespace, ShouldResemble, []string{containerId, net})
		So(cb.metrics["tx_packets"].value, ShouldEqual, 3)
		So(cb.metrics["tx_packets"].lastUpdate, ShouldResemble, timestamp)

	})
}

func TestGetStateMetrics(t *testing.T) {

	Convey("TestGetState with state.json fixture", t, func() {

		containerId := "1234abc"
		var s libcontainer.State

		s.InitStartTime = "12323"
		s.InitPid = 2

		timestamp := time.Now()
		cb := getStateMetrics(containerId, &s, timestamp)

		So(cb.namespace, ShouldResemble, []string{containerId, state})
		So(cb.metrics["start_time"].value, ShouldEqual, "12323")
		So(cb.metrics["start_time"].lastUpdate, ShouldResemble, timestamp)
		So(cb.metrics["pid"].value, ShouldEqual, 2)
		So(cb.metrics["pid"].lastUpdate, ShouldResemble, timestamp)

	})
}

func TestConfigMetrics(t *testing.T) {

	Convey("TestGetState with state.json fixture", t, func() {

		containerId := "1234abc"
		var c libcontainer.Config

		c.Hostname = "hostz"

		timestamp := time.Now()
		cb := getConfigMetrics(containerId, &c, timestamp)

		So(cb.namespace, ShouldResemble, []string{containerId, config})
		So(cb.metrics["hostname"].value, ShouldEqual, "hostz")
		So(cb.metrics["hostname"].lastUpdate, ShouldResemble, timestamp)

	})
}
