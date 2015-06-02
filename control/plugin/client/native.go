package client

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/rpc"
	"time"
	"unicode"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	connection *rpc.Client
	pluginType plugin.PluginType
}

func NewCollectorNativeClient(address string, timeout time.Duration) (PluginCollectorClient, error) {
	p, err := newNativeClient(address, timeout, plugin.CollectorPluginType)

	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewPublisherNativeClient(address string, timeout time.Duration) (PluginPublisherClient, error) {
	p, err := newNativeClient(address, timeout, plugin.PublisherPluginType)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewProcessorNativeClient(address string, timeout time.Duration) (PluginProcessorClient, error) {
	p, err := newNativeClient(address, timeout, plugin.ProcessorPluginType)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PluginNativeClient) Ping() error {
	a := plugin.PingArgs{}
	b := true
	err := p.connection.Call("SessionState.Ping", a, &b)
	return err
}

func (p *PluginNativeClient) Kill(reason string) error {
	a := plugin.KillArgs{Reason: reason}
	var b bool
	err := p.connection.Call("SessionState.Kill", a, &b)
	return err
}

func (p *PluginNativeClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	args := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}
	reply := plugin.PublishReply{}

	err := p.connection.Call("Publisher.Publish", args, &reply)

	return err
}

func (p *PluginNativeClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	args := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}
	reply := plugin.ProcessorReply{}

	err := p.connection.Call("Processor.Process", args, &reply)

	return reply.ContentType, reply.Content, err
}

func (p *PluginNativeClient) CollectMetrics(coreMetricTypes []core.Metric) ([]core.Metric, error) {
	// Convert core.MetricType slice into plugin.PluginMetricType slice as we have
	// to send structs over RPC
	pluginMetricTypes := make([]plugin.PluginMetricType, len(coreMetricTypes))
	for i, _ := range coreMetricTypes {
		pluginMetricTypes[i] = plugin.PluginMetricType{
			Namespace_:          coreMetricTypes[i].Namespace(),
			LastAdvertisedTime_: coreMetricTypes[i].LastAdvertisedTime(),
			Version_:            coreMetricTypes[i].Version(),
		}
		if coreMetricTypes[i].Config() != nil {
			///pluginMetricTypes[i].Config_ = coreMetricTypes[i].Config().Table()
			pluginMetricTypes[i].Config_ = coreMetricTypes[i].Config()
		}
	}

	// TODO return err if mts is empty
	args := plugin.CollectMetricsArgs{PluginMetricTypes: pluginMetricTypes}
	reply := plugin.CollectMetricsReply{}

	err := p.connection.Call("Collector.CollectMetrics", args, &reply)

	retMetrics := make([]core.Metric, len(reply.PluginMetrics))
	for i, _ := range reply.PluginMetrics {
		retMetrics[i] = reply.PluginMetrics[i]
	}
	return retMetrics, err
}

func (p *PluginNativeClient) GetMetricTypes() ([]core.Metric, error) {
	args := plugin.GetMetricTypesArgs{}
	reply := plugin.GetMetricTypesReply{}

	err := p.connection.Call("Collector.GetMetricTypes", args, &reply)

	retMetricTypes := make([]core.Metric, len(reply.PluginMetricTypes))
	for i, _ := range reply.PluginMetricTypes {
		retMetricTypes[i] = reply.PluginMetricTypes[i]
	}
	return retMetricTypes, err
}

func (p *PluginNativeClient) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	args := plugin.GetConfigPolicyTreeArgs{}
	reply := plugin.GetConfigPolicyTreeReply{PolicyTree: *cpolicy.NewTree()}
	err := p.connection.Call("Collector.GetConfigPolicyTree", args, &reply)
	if err != nil {
		return cpolicy.ConfigPolicyTree{}, err
	}

	return reply.PolicyTree, nil
}

func (p *PluginNativeClient) GetConfigPolicyNode() (cpolicy.ConfigPolicyNode, error) {
	args := plugin.GetConfigPolicyNodeArgs{}
	reply := plugin.GetConfigPolicyNodeReply{PolicyNode: *cpolicy.NewPolicyNode()}
	err := p.connection.Call(fmt.Sprintf("%s.GetConfigPolicyNode", p.GetType()), args, &reply)
	if err != nil {
		return cpolicy.ConfigPolicyNode{}, err
	}

	return reply.PolicyNode, nil
}

// GetType returns the string type of the plugin
// Note: the first letter of the type will be capitalized.
func (p *PluginNativeClient) GetType() string {
	return upcaseInitial(p.pluginType.String())
}

func newNativeClient(address string, timeout time.Duration, t plugin.PluginType) (*PluginNativeClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &PluginNativeClient{connection: r, pluginType: t}
	return p, nil
}

func init() {
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
}

func upcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
