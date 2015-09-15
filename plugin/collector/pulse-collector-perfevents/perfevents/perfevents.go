package perfevents

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "perfevents"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
	// Namespace definition
	ns_vendor  = "intel"
	ns_class   = "linux"
	ns_type    = "perfevents"
	ns_subtype = "cgroup"
)

type event struct {
	id    string
	etype string
	value uint64
}

type Perfevents struct {
	cgroup_events []event
	Init          func() error
}

var CGROUP_EVENTS = []string{"cycles", "instructions", "cache-references", "cache-misses",
	"branch-instructions", "branch-misses", "stalled-cycles-frontend",
	"stalled-cycles-backend", "ref-cycles"}

// CollectMetrics returns HW metrics from perf events subsystem
// for Cgroups present on the host.
func (p *Perfevents) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	if len(mts) == 0 {
		return nil, nil
	}
	events := []string{}
	cgroups := []string{}

	// Get list of events and cgroups from Namespace
	// Replace "_" with "/" in cgroup name
	for _, m := range mts {
		err := validateNamespace(m.Namespace())
		if err != nil {
			return nil, err
		}
		events = append(events, m.Namespace()[4])
		cgroups = append(cgroups, strings.Replace(m.Namespace()[5], "_", "/", -1))
	}

	// Prepare events (-e) and Cgroups (-G) switches for "perf stat"
	cgroups_switch := "-G" + strings.Join(cgroups, ",")
	events_switch := "-e" + strings.Join(events, ",")

	// Prepare "perf stat" command
	cmd := exec.Command("perf", "stat", "--log-fd", "1", `-x;`, "-a", events_switch, cgroups_switch, "--", "sleep", "1")

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe", err)
		return nil, err
	}

	// Parse "perf stat" output
	p.cgroup_events = make([]event, len(mts))
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for i := 0; scanner.Scan(); i++ {
			line := strings.Split(scanner.Text(), ";")
			value, err := strconv.ParseUint(line[0], 10, 64)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Invalid metric value", err)
			}
			etype := line[2]
			id := line[3]
			e := event{id: id, etype: etype, value: value}
			p.cgroup_events[i] = e
		}
	}()

	// Run command and wait (up to 2 secs) for completion
	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting perf stat", err)
		return nil, err
	}

	st := time.Now()
	for {
		if len(p.cgroup_events) == cap(p.cgroup_events) {
			break
		}
		if time.Since(st) > time.Second*2 {
			return nil, fmt.Errorf("Timed out waiting for metrics from perf stat")
		}
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for perf stat", err)
		return nil, err
	}

	// Populate metrics
	metrics := make([]plugin.PluginMetricType, len(mts))
	i := 0
	for _, m := range mts {
		metric, err := populate_metric(m.Namespace(), p.cgroup_events[i])
		if err != nil {
			return nil, err
		}
		metrics[i] = *metric
		metrics[i].Source_, _ = os.Hostname()
		i++
	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by perf events subsystem
func (p *Perfevents) GetMetricTypes() ([]plugin.PluginMetricType, error) {
	err := p.Init()
	if err != nil {
		return nil, err
	}
	cgroups, err := list_cgroups()
	if err != nil {
		return nil, err
	}
	if len(cgroups) == 0 {
		return nil, nil
	}
	mts := []plugin.PluginMetricType{}
	mts = append(mts, set_supported_metrics(ns_subtype, cgroups, CGROUP_EVENTS)...)

	return mts, nil
}

// GetConfigPolicy returns a ConfigPolicy
func (p *Perfevents) GetConfigPolicy() (cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return *c, nil
}

// New initializes Perfevents plugin
func NewPerfevents() *Perfevents {
	return &Perfevents{Init: initialize}
}

func initialize() error {
	file, err := os.Open("/proc/sys/kernel/perf_event_paranoid")
	if err != nil {
		if os.IsExist(err) {
			return errors.New("perf_event_paranoid file exists but couldn't be opened")
		}
		return errors.New("perf event system not enabled")
	}

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return errors.New("cannot read from perf_event_paranoid")
	}

	i, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		return errors.New("invalid value in perf_event_paranoid file")
	}

	if i >= 1 {
		return errors.New("insufficient perf event subsystem capabilities")
	}
	return nil
}

func set_supported_metrics(source string, cgroups []string, events []string) []plugin.PluginMetricType {
	mts := make([]plugin.PluginMetricType, len(events)*len(cgroups))
	for _, e := range events {
		for _, c := range flatten_cg_name(cgroups) {
			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{ns_vendor, ns_class, ns_type, source, e, c}})
		}
	}
	return mts
}
func flatten_cg_name(cg []string) []string {
	flat_cg := []string{}
	for _, c := range cg {
		flat_cg = append(flat_cg, strings.Replace(c, "/", "_", -1))
	}
	return flat_cg
}

func populate_metric(ns []string, e event) (*plugin.PluginMetricType, error) {
	return &plugin.PluginMetricType{
		Namespace_: ns,
		Data_:      e.value,
		Timestamp_: time.Now(),
	}, nil
}

func list_cgroups() ([]string, error) {
	cgroups := []string{}
	base_path := "/sys/fs/cgroup/perf_event/"
	err := filepath.Walk(base_path, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			cgroup_name := strings.TrimPrefix(path, base_path)
			if len(cgroup_name) > 0 {
				cgroups = append(cgroups, cgroup_name)
			}
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	return cgroups, nil
}

func validateNamespace(namespace []string) error {
	if len(namespace) != 6 {
		return errors.New(fmt.Sprintf("unknown metricType %s (should containt exactly 6 segments)", namespace))
	}
	if namespace[0] != ns_vendor {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected 1st segment %s)", namespace, ns_vendor))
	}

	if namespace[1] != ns_class {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected 2nd segment %s)", namespace, ns_class))
	}
	if namespace[2] != ns_type {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected 3rd segment %s)", namespace, ns_type))
	}
	if namespace[3] != ns_subtype {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected 4th segment %s)", namespace, ns_subtype))
	}
	if !namespaceContains(namespace[4], CGROUP_EVENTS) {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected 5th segment %v)", namespace, CGROUP_EVENTS))
	}
	return nil
}

func namespaceContains(element string, slice []string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}
