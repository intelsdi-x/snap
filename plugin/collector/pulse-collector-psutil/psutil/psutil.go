package psutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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

func (p *Psutil) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := make([]plugin.PluginMetricType, len(mts))
	loadre := regexp.MustCompile(`^/psutil/load/load[1,5,15]`)
	cpure := regexp.MustCompile(`^/psutil/cpu.*/.*`)
	memre := regexp.MustCompile(`^/psutil/vm/.*`)

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
		}
	}
	return metrics, nil
}

func (p *Psutil) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	mts := []plugin.PluginMetricType{}

	mts = getLoadAvgMetricTypes(mts)
	mts, err := getCPUTimesMetricTypes(mts)
	if err != nil {
		return nil, err
	}
	mts = getVirtualMemoryMetricTypes(mts)

	return mts, nil
}

//GetConfigPolicyTree returns a ConfigPolicyTree for testing
func (p *Psutil) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	c := cpolicy.NewTree()
	return *c, nil
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}

func prettyPrint(mts []plugin.PluginMetricType) error {
	var out bytes.Buffer
	mtsb, _, _ := plugin.MarshallPluginMetricTypes(plugin.PulseJSONContentType, mts)
	if err := json.Indent(&out, mtsb, "", "  "); err != nil {
		return err
	}
	fmt.Println(out.String())
	return nil
}
