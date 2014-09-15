package collection

import (
	"encoding/json"
	"fmt"
	"github.com/docker/libcontainer"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// TODO move to collector config
var (
	docker_folder string = "/var/lib/docker/execdriver/native/"
)

type containerCollector struct {
	Collector
}

func NewContainerCollector() *containerCollector {
	c := new(containerCollector)
	return c
}

func getState(state_config string) *libcontainer.State {
	f, err := os.Open(state_config)
	if err != nil {
		fmt.Printf("failed to open %s - %s\n", state_config, err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	retConfig := new(libcontainer.State)
	d.Decode(retConfig)
	return retConfig
}

func getConfig(container_config string) *libcontainer.Config {
	f, err := os.Open(container_config)
	if err != nil {
		fmt.Printf("failed to open %s - %s\n", container_config, err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	retConfig := new(libcontainer.Config)
	d.Decode(retConfig)
	return retConfig
}

// TODO libcontainer update removed support for collection of CPU by removing sampling. Need to reimplement locally.
// func getCpuMetrics(stats *libcontainer.ContainerStats, c *libcontainer.Config, s *libcontainer.State) Metric {
// 	m := Metric{}
// 	// Hostname for container
// 	m.Host = GetHostname()
// 	m.Namespace = []string{"container", c.Hostname, "cpu"}
// 	m.Collector = "libcontainer"
// 	m.LastUpdate = time.Now()
// 	m.Values = map[string]metricType{}
// 	m.Values["percent_usage"] = stats.CgroupStats.CpuStats.CpuUsage.PercentUsage
// 	pc := stats.CgroupStats.CpuStats.CpuUsage.PercpuUsage
// 	var sum uint64 = 0
// 	for _, x := range pc {
// 		sum += x
// 	}
// 	for i, x := range pc {
// 		s := fmt.Sprintf("core%d_ratio", i)
// 		m.Values[s] = fmt.Sprintf("%0.3f", float64(x)/float64(sum))
// 	}
// 	//	metrics = append(metrics, m)
// 	return m
// }

func getNetMetrics(stats *libcontainer.ContainerStats, c *libcontainer.Config, s *libcontainer.State) Metric {
	values := map[string]metricType{}
	values["tx_bytes"] = stats.NetworkStats.TxBytes
	values["rx_bytes"] = stats.NetworkStats.RxBytes
	values["tx_packets"] = stats.NetworkStats.TxPackets
	values["rx_packets"] = stats.NetworkStats.RxPackets
	values["tx_dropped"] = stats.NetworkStats.TxDropped
	values["rx_dropped"] = stats.NetworkStats.RxDropped
	values["tx_errors"] = stats.NetworkStats.TxErrors
	values["rx_errors"] = stats.NetworkStats.RxErrors
	return Metric{
		Host:       GetHostname(),
		Namespace:  []string{"container", c.Hostname, "net"},
		LastUpdate: time.Now(),
		Values:     values,
		Collector:  "libcontainer",
		MetricType: Polling,
	}
}

func getBlkMetrics(stats *libcontainer.ContainerStats, c *libcontainer.Config, s *libcontainer.State) Metric {
	return Metric{}
}

func getContainerMetric(container_path string) ([]Metric, error) {
	s := getState(filepath.Join(container_path, "state.json"))
	c := getConfig(filepath.Join(container_path, "container.json"))
	stats, err := libcontainer.GetStats(c, s)

	values := map[string]metricType{}
	values["start_time"] = s.InitStartTime
	values["pid"] = s.InitPid
	values["veth_host"] = s.NetworkState.VethHost
	values["veth_child"] = s.NetworkState.VethChild
	// State metric
	m := Metric{
		Host:       GetHostname(),
		Namespace:  []string{"container", c.Hostname, "state"},
		LastUpdate: time.Now(),
		Values:     values,
		Collector:  "libcontainer",
		MetricType: Polling,
	}
	//
	// cpu_m := getCpuMetrics(stats, c, s)
	net_m := getNetMetrics(stats, c, s)
	metrics := []Metric{m, cpu_m, net_m}
	return metrics, err
}

// TODO switch to static Metric List
func (x *containerCollector) GetMetricList() []Metric {
	folders, _ := ioutil.ReadDir(docker_folder)
	container_count := 0
	container_ids := []string{}
	metrics := []Metric{}
	for _, folder := range folders {
		if folder.IsDir() {
			m, err := getContainerMetric(filepath.Join(docker_folder, folder.Name()))
			if err == nil {
				container_count++
				container_ids = append(container_ids, folder.Name()[0:11])
				metrics = append(metrics, m...)
			}
		}
	}
	// Global container metric
	values := map[string]metricType{}
	values["count"] = container_count
	values["ids"] = container_ids

	m := Metric{
		Host:       GetHostname(),
		Namespace:  []string{"container", "global"},
		LastUpdate: time.Now(),
		Values:     values,
		Collector:  "libcontainer",
		MetricType: Polling,
	}

	// Insert global metrics at the start
	metrics = append(metrics[:0], append([]Metric{m}, metrics[0:]...)...)
	return metrics
}

func (x *containerCollector) GetMetricValues() []Metric {
	return x.GetMetricList()
}
