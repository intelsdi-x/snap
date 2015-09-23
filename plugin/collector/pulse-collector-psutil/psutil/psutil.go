package psutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "psutil"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

type Psutil struct {
}

// CollectMetrics returns metrics from gopsutil
func (p *Psutil) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	hostname, _ := os.Hostname()
	metrics := make([]plugin.PluginMetricType, len(mts))
	loadre := regexp.MustCompile(`^/psutil/load/load[1,5,15]`)
	cpure := regexp.MustCompile(`^/psutil/cpu.*/.*`)
	memre := regexp.MustCompile(`^/psutil/vm/.*`)
	netre := regexp.MustCompile(`^/psutil/net/.*`)

	for i, p := range mts {
		ns := joinNamespace(p.Namespace())
		switch {
		case loadre.MatchString(ns):
			metric, err := loadAvg(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		case cpure.MatchString(ns):
			metric, err := cpuTimes(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		case memre.MatchString(ns):
			metric, err := virtualMemory(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		case netre.MatchString(ns):
			metric, err := netIOCounters(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		}
		metrics[i].Source_ = hostname
		metrics[i].Timestamp_ = time.Now()

	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by gopsutil
func (p *Psutil) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	mts := []plugin.PluginMetricType{}

	mts = append(mts, getLoadAvgMetricTypes()...)
	mts_, err := getCPUTimesMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mts_...)
	mts = append(mts, getVirtualMemoryMetricTypes()...)
	mts_, err = getNetIOCounterMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mts_...)

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
func (p *Psutil) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}

func prettyPrint(mts []plugin.PluginMetricType) error {
	var out bytes.Buffer
	mtsb, _, _ := plugin.MarshalPluginMetricTypes(plugin.PulseJSONContentType, mts)
	if err := json.Indent(&out, mtsb, "", "  "); err != nil {
		return err
	}
	fmt.Println(out.String())
	return nil
}
