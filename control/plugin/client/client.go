package client

import (
	"net"
	"net/rpc"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

// A client providing common plugin method calls.
type PluginClient interface {
	Ping() error
	Kill(string) error
}

// A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	Collect(plugin.CollectorArgs, *plugin.CollectorReply) error
	GetMetricTypes(plugin.GetMetricTypesArgs, *plugin.GetMetricTypesReply) error
}

// A client providing processor specific plugin method calls.
type PluginProcessorClient interface {
	PluginClient
	ProcessMetricData()
}

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	connection *rpc.Client
}

func (p *PluginNativeClient) Ping() error {
	a := plugin.PingArgs{}
	b := true
	err := p.connection.Call("SessionState.Ping", a, &b)
	return err
}

func (p *PluginNativeClient) Kill(reason string) error {
	a := plugin.KillArgs{Reason: reason}
	b := true
	err := p.connection.Call("SessionState.Kill", a, &b)
	return err
}

type CollectorClient struct {
	//PluginClient
	PluginNativeClient
	//connection *rpc.Client
}

func (c *CollectorClient) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	err := c.connection.Call("Collector.Collect", args, reply)
	return err
}

func (c *CollectorClient) GetMetricTypes(args plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {
	err := c.connection.Call("Collector.GetMetricTypes", args, reply)
	return err
}

func NewCollectorClient(address string, timeout time.Duration) (PluginCollectorClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &CollectorClient{PluginNativeClient{connection: r}}
	return p, nil
}
