package riemann

import (
	"log"

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

type Riemann struct{}

// NewRiemannPublisher does something cool
func NewRiemannPublisher() *Riemann {
	var r *Riemann
	return r
}

// GetConfigPolicyNode returns the config policy for the Riemann Publisher Plugin
func (r *Riemann) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()
	// Host metric applies to
	r1, err := cpolicy.NewStringRule("host", true)
	handleErr(err)
	r1.Description = "Host the metric was collected from"

	// Metric that is being collected
	r2, err := cpolicy.NewStringRule("service", true)
	handleErr(err)
	r2.Description = "Service (metric) being collected"

	// Riemann server to publish event to
	r3, err := cpolicy.NewStringRule("broker", true)
	handleErr(err)
	r3.Description = "Broker in the format of broker-ip:port (ex: 192.168.1.1:5555)"

	config.Add(r1, r2, r3)
	return *config
}

// Publish serializes the data and calls publish to send events to Riemann
func (r *Riemann) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue, logger *log.Logger) error {
	//err := r.publish(event, broker)
	//return err
	return nil
}

// publish sends events to riemann
func (r *Riemann) publish(event *raidman.Event, broker string) error {
	c, err := raidman.Dial("tcp", broker)
	defer c.Close()
	if err != nil {
		return err
	}
	return c.Send(event)
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
