package smart

import (
	"errors"
	"fmt"
	"strings"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

const (
	Name    = "pulse-collector-smart"
	Version = 1
	Type    = plugin.CollectorPluginType
)

var sysUtilProvider SysutilProvider = NewSysutilProvider()

var namespace_prefix = []string{"intel", "disk"}
var namespace_suffix = []string{"smart"}

func makeName(device, metric string) []string {
	splited := strings.Split(metric, "/")

	name := []string{}
	name = append(name, namespace_prefix...)
	name = append(name, device)
	name = append(name, namespace_suffix...)
	name = append(name, splited...)

	return name
}

func parseName(namespace []string) (disk, attribute string) {
	disk = namespace[len(namespace_prefix)]
	smart_namespace := namespace[len(namespace_prefix)+len(namespace_suffix)+1:]
	attribute = strings.Join(smart_namespace, "/")
	return
}

func validateName(namespace []string) bool {
	for i, v := range namespace_prefix {
		if namespace[i] != v {
			return false
		}
	}

	offset := len(namespace_prefix) + 1
	for i, v := range namespace_suffix {
		if namespace[offset+i] != v {
			return false
		}
	}

	return true
}

type SmartCollector struct {
}

type smartResults map[string]interface{}

// CollectMetrics returns metrics from smart
func (sc *SmartCollector) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	buffered_results := map[string]smartResults{}

	results := make([]plugin.PluginMetricType, len(mts))

	for i, mt := range mts {
		if !validateName(mt.Namespace()) {
			return nil, errors.New(fmt.Sprintf("%v is not valid metric", mt.Namespace()))
		}
		disk, attribute_path := parseName(mt.Namespace())
		buffered, ok := buffered_results[disk]
		if !ok {
			values, err := ReadSmartData(disk, sysUtilProvider)
			if err != nil {
				return nil, err
			}
			buffered = values.GetAttributes()
			buffered_results[disk] = buffered
		}

		attribute, ok := buffered[attribute_path]

		if !ok {
			return nil, errors.New("Unknown attribute " + attribute_path)
		}

		results[i] = plugin.PluginMetricType{
			Namespace_: mt.Namespace(),
			Data_:      attribute,
		}
	}

	return results, nil
}

// GetMetricTypes returns the metric types exposed by smart
func (sc *SmartCollector) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	smart_metrics := ListAllKeys()
	devices, err := sysUtilProvider.ListDevices()
	if err != nil {
		return nil, err
	}
	mts := make([]plugin.PluginMetricType, 0, len(smart_metrics)*len(devices))

	for _, device := range devices {
		for _, metric := range smart_metrics {
			path := makeName(device, metric)
			mts = append(mts, plugin.PluginMetricType{Namespace_: path})
		}
	}

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
func (p *SmartCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}
