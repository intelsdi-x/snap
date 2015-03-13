package container

import (
	"testing"

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

		cb := getNetMetrics(containerId, &stats)

		So(cb.namespace, ShouldResemble, []string{containerId, net})
		So(cb.metrics["tx_packets"].value, ShouldEqual, 3)

	})
}

func TestGetStateMetrics(t *testing.T) {

	Convey("TestGetState with state.json fixture", t, func() {

		containerId := "1234abc"
		var s libcontainer.State

		s.InitStartTime = "12323"
		s.InitPid = 2

		cb := getStateMetrics(containerId, &s)

		So(cb.namespace, ShouldResemble, []string{containerId, state})
		So(cb.metrics["start_time"].value, ShouldEqual, "12323")
		So(cb.metrics["pid"].value, ShouldEqual, 2)

	})
}

func TestConfigMetrics(t *testing.T) {

	Convey("TestGetState with state.json fixture", t, func() {

		containerId := "1234abc"
		var c libcontainer.Config

		c.Hostname = "hostz"

		cb := getConfigMetrics(containerId, &c)

		So(cb.namespace, ShouldResemble, []string{containerId, config})
		So(cb.metrics["hostname"].value, ShouldEqual, "hostz")

	})
}
