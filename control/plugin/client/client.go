package client

import (
	"net"
	"net/rpc"
	"time"
)

// A client providing common plugin method calls.
type PluginClient interface {
	Ping() error
}

// A client providing collector specific plugin method calls.
type PluginCollectorClient interface {
	PluginClient
	// CollectMetricData()
	// ListMetrics()
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
	return nil
}

func NewCollectorClient(address string, timeout time.Duration) (PluginCollectorClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &PluginNativeClient{connection: r}
	return p, nil
}
