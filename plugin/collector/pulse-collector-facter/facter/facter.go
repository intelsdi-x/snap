/* This modules converts implements Pulse API to become an plugin.

legend:
- metric - represents value of metric from Pulse side
- fact - represents a value about a system gathered from Facter
- name - is string identifier that refers to metric from the Pulse side, so name points to metric

 Implementation details:

 GetMetricTypes()
      +
      |                 +------------------+
 +----v---+  getFacts() |                  |
 | Facter +-------------> ./facter --json  |
 +----^---+             |    (goroutine)   |
      |                 +------------------+
      +
 CollectMetrics()


*/

package facter

import (
	"errors"
	"fmt"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

const (
	// parts of returned namescape
	vendor = "intel"
	prefix = "facter"
	// how long we are caching the date from external binary to prevent overuse of resources
	defaultCacheTTL = 60 * time.Second
	// how long are we going to cache available types of metrics
	defaultMetricTypesTTL = defaultCacheTTL
	// timeout we are ready to wait for external binary to gather the data
	defaultFacterTimeout = 5 * time.Second
)

/**********
 * Facter *
 **********/

// Facter implements API to communicate with Pulse
type Facter struct {
	ttl time.Duration
	// injects implementation for getting facts - defaults to use getFacts from cmd.go
	// but allows to replace with fake during tests
	getFacts func(
		names []string,
		facterTimeout time.Duration,
		cmdConfig *cmdConfig,
	) (facts, error)
	// how much time we are ready to wait for getting result from cmd
	// is going to be passed to facterTimeout parameter in getFacts
	facterTimeout time.Duration
}

// make sure that we actually satisify requierd interface
var _ plugin.CollectorPlugin = (*Facter)(nil)

// NewFacter constructs new Facter with default values
func NewFacter() *Facter {
	return &Facter{
		// injection of default implementation for gathering facts from Facter
		getFacts:      getFacts,
		facterTimeout: defaultFacterTimeout,
	}
}

// ------------ Pulse plugin interface implementation --------------

// GetMetricTypes returns available metrics types
func (f *Facter) GetMetricTypes() ([]plugin.PluginMetricType, error) {

	// facts composed of entries
	facts, err := f.getFacts(
		nil, // ask for everything
		f.facterTimeout,
		nil, //default cmd configuration
	)
	if err != nil {
		return nil, err
	}

	// capacity - we are going to return all the facts
	metricTypes := make([]plugin.PluginMetricType, 0, len(facts))

	// create types withing given namespace
	for name, _ := range facts {
		namespace := createNamespace(name)
		metricType := *plugin.NewPluginMetricType(namespace)
		metricTypes = append(metricTypes, metricType)
	}

	return metricTypes, nil
}

// Collect collects metrics from external binary a returns them in form
// acceptable by Pulse, only returns collects that were asked for and return nothing when asked for none
// the order of requested and received metrics isn't guaranted
func (f *Facter) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {

	// parse and check requested names of metrics
	names := []string{}
	for _, metricType := range metricTypes {
		namespace := metricType.Namespace()

		err := validateNamespace(namespace)
		if err != nil {
			return nil, err
		}

		// name of fact - last part of namespace
		name := namespace[2]
		names = append(names, name)
	}

	if len(names) == 0 {
		// nothing request, none returned
		// !because returned by value, it would return nil slice
		return nil, nil
	}

	// facts composed of entries
	facts, err := f.getFacts(names, f.facterTimeout, nil)
	if err != nil {
		return nil, err
	}

	// make sure that recevied len of names equals asked
	if len(facts) != len(names) {
		return nil, errors.New("assertion: getFacts returns more/less than asked!")
	}

	// convert facts into PluginMetrics
	metrics := make([]plugin.PluginMetric, 0, len(facts))
	for name, value := range facts {
		namespace := createNamespace(name)
		metric := *plugin.NewPluginMetric(namespace, value)
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (f *Facter) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	c := cpolicy.NewTree()
	rule, _ := cpolicy.NewStringRule("name", false, "bob")
	rule2, _ := cpolicy.NewStringRule("password", true)
	p := cpolicy.NewPolicyNode()
	p.Add(rule)
	p.Add(rule2)
	c.Add([]string{"intel", "facter", "foo"}, p)
	return *c, nil
}

// ------------ helper functions --------------

// validateNamespace checks namespace intel(vendor)/facter(prefix)/FACTNAME
func validateNamespace(namespace []string) error {
	if len(namespace) != 3 {
		return errors.New(fmt.Sprintf("unknown metricType %s (should containt just 3 segments)", namespace))
	}
	if namespace[0] != vendor {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected vendor %s)", namespace, vendor))
	}

	if namespace[1] != prefix {
		return errors.New(fmt.Sprintf("unknown metricType %s (expected prefix %s)", namespace, prefix))
	}
	return nil
}

// namspace returns namespace slice of strings
// composed from: vendor, prefix and fact name
func createNamespace(name string) []string {
	return []string{vendor, prefix, name}

}

// helper type to deal with json values which additionally stores last update moment
type entry struct {
	value      interface{}
	lastUpdate time.Time
}
