package client

import (
	"net"
	"net/rpc"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	connection *rpc.Client
}

func NewCollectorNativeClient(address string, timeout time.Duration) (PluginCollectorClient, error) {
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

func (p *PluginNativeClient) CollectMetrics(mts []plugin.MetricType) ([]plugin.Metric, error) {
	// TODO return err if mts is empty
	args := plugin.CollectMetricsArgs{MetricTypes: mts}
	reply := plugin.CollectMetricsReply{}

	err := p.connection.Call("Collector.CollectMetrics", args, &reply)
	return reply.Metrics, err
}

func (p *PluginNativeClient) GetMetricTypes() ([]plugin.MetricType, error) {
	args := plugin.GetMetricTypesArgs{}
	reply := plugin.GetMetricTypesReply{}

	err := p.connection.Call("Collector.GetMetricTypes", args, &reply)
	return reply.MetricTypes, err
}
