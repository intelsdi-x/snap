package riemann

import (
	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core/ctypes"

	"github.com/amir/raidman"
)

const (
	PluginName    = "riemann"
	PluginVersion = 1
	PluginType    = plugin.PublisherPluginType
)

// Meta returns the metadata details for the Riemann Publisher Plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(PluginName, PluginVersion, PluginType)
}

// ConfigPolicyNode returns the config policy for the Riemann Publisher Plugin
func ConfigPolicyNode() *config.ConfigPolicyNode {
	config := cpolicy.ConfigPolicyNode()
	// Host metric applies to
	r1, err := cpolicy.NewStringRule("host", true)
	handleErr(err)
	r1.description = "Host the metric was collected from"

	// Metric that is being collected
	r2, err := cpolicy.NewStringRule("service", true)
	handleErr(err)
	r2.description = "Service (metric) being collected"

	// Riemann server to publish event to
	r3, err := cpolicy.NewStringRule("broker", true)
	handleErr(err)
	r3.description = "Broker in the format of broker-ip:port (ex: 192.168.1.1:5555)"

	config.Add(r1, r2, r3)
	return config
}

type Riemann struct{}

// NewRiemannPublisher does something cool
func NewRiemannPublisher() *Riemann {
	var r *Riemann
	return r
}

// Publish serializes the data and calls publish to send events to Riemann
func (r *riemann) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {

	err := r.publish(event, broker)
	return err
}

// publish sends events to riemann
func (r *riemann) publish(event *raidman.Event, broker string) error {
	c, err := raidman.Dial("tcp", broker)
	if err != nil {
		return err
	}
	err = c.Send(event)
	if err != nil {
		return err
	}
	c.Close()
}

func handleErr(e Error) {
	if e != nil {
		panic(e)
	}
}
