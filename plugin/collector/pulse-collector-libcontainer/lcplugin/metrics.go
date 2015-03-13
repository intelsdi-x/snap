package lcplugin

import (
	"time"

	"github.com/docker/libcontainer"
)

func getNetMetrics(container string, stats *libcontainer.ContainerStats) cacheBucket {

	ns := []string{container, net}
	cache := make(map[string]metric)
	now := time.Now()

	cache["tx_bytes"] = newMetric(stats.NetworkStats.TxBytes, now)
	cache["rx_bytes"] = newMetric(stats.NetworkStats.RxBytes, now)
	cache["tx_packets"] = newMetric(stats.NetworkStats.TxPackets, now)
	cache["rx_packets"] = newMetric(stats.NetworkStats.RxPackets, now)
	cache["tx_dropped"] = newMetric(stats.NetworkStats.TxDropped, now)
	cache["rx_dropped"] = newMetric(stats.NetworkStats.RxDropped, now)
	cache["tx_errors"] = newMetric(stats.NetworkStats.TxErrors, now)
	cache["rx_errors"] = newMetric(stats.NetworkStats.RxErrors, now)

	return cacheBucket{namespace: ns, metrics: cache}
}

func getStateMetrics(container string, st *libcontainer.State) cacheBucket {

	ns := []string{container, state}
	cache := make(map[string]metric)
	now := time.Now()

	cache["start_time"] = newMetric(st.InitStartTime, now)
	cache["pid"] = newMetric(st.InitPid, now)
	cache["veth_host"] = newMetric(st.NetworkState.VethHost, now)
	cache["veth_child"] = newMetric(st.NetworkState.VethChild, now)

	return cacheBucket{namespace: ns, metrics: cache}
}

func getConfigMetrics(container string, conf *libcontainer.Config) cacheBucket {

	ns := []string{container, config}
	cache := make(map[string]metric)
	now := time.Now()

	cache["hostname"] = newMetric(conf.Hostname, now)

	return cacheBucket{namespace: ns, metrics: cache}
}
