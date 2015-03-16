package lcplugin

import (
	"time"

	"github.com/docker/libcontainer"
)

func getNetMetrics(container string, stats *libcontainer.ContainerStats, timestamp time.Time) cacheBucket {

	ns := []string{container, net}
	cache := make(map[string]metric)

	cache["tx_bytes"] = newMetric(stats.NetworkStats.TxBytes, timestamp)
	cache["rx_bytes"] = newMetric(stats.NetworkStats.RxBytes, timestamp)
	cache["tx_packets"] = newMetric(stats.NetworkStats.TxPackets, timestamp)
	cache["rx_packets"] = newMetric(stats.NetworkStats.RxPackets, timestamp)
	cache["tx_dropped"] = newMetric(stats.NetworkStats.TxDropped, timestamp)
	cache["rx_dropped"] = newMetric(stats.NetworkStats.RxDropped, timestamp)
	cache["tx_errors"] = newMetric(stats.NetworkStats.TxErrors, timestamp)
	cache["rx_errors"] = newMetric(stats.NetworkStats.RxErrors, timestamp)

	return cacheBucket{namespace: ns, metrics: cache}
}

func getStateMetrics(container string, st *libcontainer.State, timestamp time.Time) cacheBucket {

	ns := []string{container, state}
	cache := make(map[string]metric)

	cache["start_time"] = newMetric(st.InitStartTime, timestamp)
	cache["pid"] = newMetric(st.InitPid, timestamp)
	cache["veth_host"] = newMetric(st.NetworkState.VethHost, timestamp)
	cache["veth_child"] = newMetric(st.NetworkState.VethChild, timestamp)

	return cacheBucket{namespace: ns, metrics: cache}
}

func getConfigMetrics(container string, conf *libcontainer.Config, timestamp time.Time) cacheBucket {

	ns := []string{container, config}
	cache := make(map[string]metric)

	cache["hostname"] = newMetric(conf.Hostname, timestamp)

	return cacheBucket{namespace: ns, metrics: cache}
}
