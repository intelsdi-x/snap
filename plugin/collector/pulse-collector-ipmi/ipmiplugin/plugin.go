package ipmiplugin

import (
	"fmt"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-ipmi/ipmi"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	Name    = "pulse-collector-ipmi"
	Version = 1
	Type    = plugin.CollectorPluginType
)

var namespace_prefix = []string{"intel", "ipmi"}

func makeName(metric string) []string {
	return append(namespace_prefix, strings.Split(metric, "/")...)
}

func parseName(namespace []string) string {
	return strings.Join(namespace[len(namespace_prefix):], "/")
}

func extendPath(path, ext string) string {
	if ext == "" {
		return path
	} else {
		return path + "/" + ext
	}
}

// Ipmi Collector Plugin class.
// IpmiLayer specifies interface to perform ipmi commands.
// NSim is number of requests allowed to be 'in processing' state.
// Vendor is list of request descriptions. Each of them specifies
// RAW request data, root path for metrics
// and format (which also specifies submetrics)
type IpmiCollector struct {
	IpmiLayer ipmi.IpmiAL
	NSim      int
	Vendor    []ipmi.RequestDescription

	requestIndexCache     map[string]int
	requestIndexCacheOnce sync.Once
}

func (ic *IpmiCollector) ensureRequestIdxCache() {
	ic.requestIndexCache = make(map[string]int)

	for i, req := range ic.Vendor {
		for _, metric := range req.Format.GetMetrics() {
			path := extendPath(req.MetricsRoot, metric)
			ic.requestIndexCache[path] = i
		}
	}
}

func (ic *IpmiCollector) validateName(namespace []string) error {
	for i, e := range namespace_prefix {
		if namespace[i] != e {
			return fmt.Errorf("Wrong namespace prefix in namespace %v", namespace)
		}
	}

	name := parseName(namespace)
	_, ok := ic.requestIndexCache[name]

	if !ok {
		return fmt.Errorf("Key %s not in index cache", name)
	}

	return nil
}

// Performs metric collection.
// Ipmi request are never duplicated in order to read multiple metrics.
// Timestamp is set to time when batch processing is complete.
// Source is hostname returned by operating system.
func (ic *IpmiCollector) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {

	ic.requestIndexCacheOnce.Do(func() {
		ic.ensureRequestIdxCache()
	})

	requestSet := map[int]bool{}
	for _, mt := range mts {
		ns := mt.Namespace()
		if err := ic.validateName(ns); err != nil {
			return nil, err
		}
		rid := ic.requestIndexCache[parseName(ns)]
		requestSet[rid] = true
	}

	requestList := make([]ipmi.IpmiRequest, 0)
	requestDescList := make([]*ipmi.RequestDescription, 0)
	responseCache := map[string]uint16{}
	for k, _ := range requestSet {
		desc := &ic.Vendor[k]
		requestList = append(requestList, desc.Request)
		requestDescList = append(requestDescList, desc)
	}

	//TODO: nSim from config
	resp, err := ic.IpmiLayer.BatchExecRaw(requestList, ic.NSim)

	if err != nil {
		return nil, err
	}

	for i, r := range resp {
		format := requestDescList[i].Format
		if err := format.Validate(r.Data); err != nil {
			return nil, err
		}
		submetrics := format.Parse(r.Data)
		for k, v := range submetrics {
			path := extendPath(requestDescList[i].MetricsRoot, k)
			responseCache[path] = v
		}
	}

	results := make([]plugin.PluginMetricType, len(mts))
	t := time.Now()
	host, _ := os.Hostname()

	for i, mt := range mts {
		ns := mt.Namespace()
		key := parseName(ns)
		data := responseCache[key]
		metric := plugin.PluginMetricType{Namespace_: ns, Source_: host,
			Timestamp_: t, Data_: data}

		results[i] = metric
	}

	return results, nil

}

// Returns list of metrics available for current vendor.
func (ic *IpmiCollector) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	mts := make([]plugin.PluginMetricType, 0)
	for _, req := range ic.Vendor {
		for _, metric := range req.Format.GetMetrics() {
			path := extendPath(req.MetricsRoot, metric)
			mts = append(mts, plugin.PluginMetricType{Namespace_: makeName(path)})
		}
	}

	return mts, nil
}

func (ic *IpmiCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}
