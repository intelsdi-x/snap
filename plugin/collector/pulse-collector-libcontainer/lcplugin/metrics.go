package lcplugin

import (
	"reflect"
	"time"

	"github.com/docker/libcontainer"
)

func getNetMetrics(container string, stats *libcontainer.ContainerStats, timestamp time.Time) cacheBucket {

	ns := []string{vendor, prefix, container, net}
	cache := make(map[string]metric)

	reflected := reflect.Indirect(reflect.ValueOf(stats.NetworkStats))

	for idx := 0; idx < reflected.NumField(); idx++ {
		val := reflected.Field(idx).Interface()
		tag := reflected.Type().Field(idx).Tag.Get("json")
		cache[tag] = newMetric(val, timestamp, append(ns, tag))
	}

	return cacheBucket{namespace: ns, metrics: cache}
}

func getStateMetrics(container string, st *libcontainer.State, timestamp time.Time) cacheBucket {

	ns := []string{vendor, prefix, container, state}
	cache := make(map[string]metric)

	cache["start_time"] = newMetric(st.InitStartTime, timestamp, append(ns, "start_time"))
	cache["pid"] = newMetric(st.InitPid, timestamp, append(ns, "pid"))
	cache["veth_host"] = newMetric(st.NetworkState.VethHost, timestamp, append(ns, "veth_host"))
	cache["veth_child"] = newMetric(st.NetworkState.VethChild, timestamp, append(ns, "veth_child"))

	return cacheBucket{namespace: ns, metrics: cache}
}

func getConfigMetrics(container string, conf *libcontainer.Config, timestamp time.Time) cacheBucket {

	ns := []string{vendor, prefix, container, config}
	cache := make(map[string]metric)

	cache["hostname"] = newMetric(conf.Hostname, timestamp, append(ns, "hostname"))

	return cacheBucket{namespace: ns, metrics: cache}
}
