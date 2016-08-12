/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
)

type pluginClient interface {
	SetKey(ctx context.Context, in *rpc.SetKeyArg, opts ...grpc.CallOption) (*rpc.SetKeyReply, error)
	Ping(ctx context.Context, in *common.Empty, opts ...grpc.CallOption) (*rpc.PingReply, error)
	Kill(ctx context.Context, in *rpc.KillRequest, opts ...grpc.CallOption) (*rpc.KillReply, error)
	GetConfigPolicy(ctx context.Context, in *common.Empty, opts ...grpc.CallOption) (*rpc.GetConfigPolicyReply, error)
}

type grpcClient struct {
	collector rpc.CollectorClient
	processor rpc.ProcessorClient
	publisher rpc.PublisherClient
	plugin    pluginClient

	pluginType plugin.PluginType
	timeout    time.Duration
	conn       *grpc.ClientConn
	encrypter  *encrypter.Encrypter
}

// NewCollectorGrpcClient returns a collector gRPC Client.
func NewCollectorGrpcClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginCollectorClient, error) {
	address, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	p, err := newGrpcClient(address, int(port), timeout, plugin.CollectorPluginType)
	if err != nil {
		return nil, err
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		encrypter := encrypter.New(pub, nil)
		encrypter.Key = key
		p.encrypter = encrypter
	}

	return p, nil
}

// NewProcessorGrpcClient returns a processor gRPC Client.
func NewProcessorGrpcClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginProcessorClient, error) {
	address, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	p, err := newGrpcClient(address, int(port), timeout, plugin.ProcessorPluginType)
	if err != nil {
		return nil, err
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		encrypter := encrypter.New(pub, nil)
		encrypter.Key = key
		p.encrypter = encrypter
	}

	return p, nil
}

// NewPublisherGrpcClient returns a publisher gRPC Client.
func NewPublisherGrpcClient(address string, timeout time.Duration, pub *rsa.PublicKey, secure bool) (PluginPublisherClient, error) {
	address, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	p, err := newGrpcClient(address, int(port), timeout, plugin.PublisherPluginType)
	if err != nil {
		return nil, err
	}
	if secure {
		key, err := encrypter.GenerateKey()
		if err != nil {
			return nil, err
		}
		encrypter := encrypter.New(pub, nil)
		encrypter.Key = key
		p.encrypter = encrypter
	}

	return p, nil
}

func parseAddress(address string) (string, int64, error) {
	addr := strings.Split(address, ":")
	if len(addr) != 2 {
		return "", 0, fmt.Errorf("bad address")
	}
	address = addr[0]
	port, err := strconv.ParseInt(addr[1], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return address, port, nil
}

func newGrpcClient(addr string, port int, timeout time.Duration, typ plugin.PluginType) (*grpcClient, error) {
	conn, err := rpcutil.GetClientConnection(addr, port)
	if err != nil {
		return nil, err
	}
	p := &grpcClient{
		timeout: timeout,
		conn:    conn,
	}

	switch typ {
	case plugin.CollectorPluginType:
		p.collector = rpc.NewCollectorClient(conn)
		p.plugin = p.collector
	case plugin.ProcessorPluginType:
		p.processor = rpc.NewProcessorClient(conn)
		p.plugin = p.processor
	case plugin.PublisherPluginType:
		p.publisher = rpc.NewPublisherClient(conn)
		p.plugin = p.publisher
	default:
		return nil, errors.New(fmt.Sprintf("Invalid plugin type provided %v", typ))
	}

	return p, nil
}

func getContext(timeout time.Duration) context.Context {
	ctxTimeout, _ := context.WithTimeout(context.Background(), timeout)
	return ctxTimeout
}

func (g *grpcClient) Ping() error {
	_, err := g.plugin.Ping(getContext(g.timeout), &common.Empty{})
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) SetKey() error {
	out, err := g.encrypter.EncryptKey()
	if err != nil {
		return err
	}
	reply, err := g.plugin.SetKey(getContext(g.timeout), &rpc.SetKeyArg{Key: out})
	if err != nil {
		return err
	}

	if reply.Error != "" {
		return errors.New(reply.Error)
	}

	return nil
}

func (g *grpcClient) Kill(reason string) error {
	_, err := g.plugin.Kill(getContext(g.timeout), &rpc.KillRequest{Reason: reason})
	g.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	arg := &rpc.PublishArg{
		ContentType: contentType,
		Content:     content,
		Config:      common.ToConfigMap(config),
	}
	// return is empty so we don't need it
	_, err := g.publisher.Publish(getContext(g.timeout), arg)
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	arg := &rpc.ProcessArg{
		ContentType: contentType,
		Content:     content,
		Config:      common.ToConfigMap(config),
	}
	reply, err := g.processor.Process(getContext(g.timeout), arg)
	if err != nil {
		return "", nil, err
	}
	if reply.Error != "" {
		return "", nil, errors.New(reply.Error)
	}
	return reply.ContentType, reply.Content, nil
}

func (g *grpcClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	arg := &rpc.CollectMetricsArg{
		Metrics: common.NewMetrics(mts),
	}
	reply, err := g.collector.CollectMetrics(getContext(g.timeout), arg)

	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	metrics := common.ToCoreMetrics(reply.Metrics)
	var results []core.Metric
	// Convert it to plugin.MetricType because scheduler/job.go checks that is the type before encoding
	// and sending to the plugin.
	//TODO(CDR): Decide what to do here if/when content-type handling is refactored
	for _, metric := range metrics {
		mt := plugin.MetricType{
			Namespace_:          metric.Namespace(),
			LastAdvertisedTime_: metric.LastAdvertisedTime(),
			Version_:            metric.Version(),
			Config_:             metric.Config(),
			Data_:               metric.Data(),
			Tags_:               metric.Tags(),
			Unit_:               metric.Unit(),
			Description_:        metric.Description(),
			Timestamp_:          metric.Timestamp(),
		}
		results = append(results, mt)
	}
	return results, nil
}

func (g *grpcClient) GetMetricTypes(config plugin.ConfigType) ([]core.Metric, error) {
	arg := &rpc.GetMetricTypesArg{
		Config: common.ToConfigMap(config.Table()),
	}
	reply, err := g.collector.GetMetricTypes(getContext(g.timeout), arg)

	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	for _, metric := range reply.Metrics {
		metric.LastAdvertisedTime = common.ToTime(time.Now())
	}

	results := common.ToCoreMetrics(reply.Metrics)
	return results, nil
}

func (g *grpcClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	reply, err := g.plugin.GetConfigPolicy(getContext(g.timeout), &common.Empty{})

	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return rpc.ToConfigPolicy(reply), nil
}
