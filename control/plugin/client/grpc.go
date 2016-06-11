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

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
)

type callsGrpcCollector interface {
	SetKey(context.Context, *rpc.SetKeyArg, ...grpc.CallOption) (*rpc.SetKeyReply, error)
	Ping(context.Context, *common.Empty, ...grpc.CallOption) (*rpc.PingReply, error)
	Kill(context.Context, *rpc.KillRequest, ...grpc.CallOption) (*rpc.KillReply, error)
	GetConfigPolicy(context.Context, *common.Empty, ...grpc.CallOption) (*rpc.GetConfigPolicyReply, error)
	CollectMetrics(context.Context, *rpc.CollectMetricsArg, ...grpc.CallOption) (*rpc.CollectMetricsReply, error)
	GetMetricTypes(context.Context, *rpc.GetMetricTypesArg, ...grpc.CallOption) (*rpc.GetMetricTypesReply, error)
}

// Native clients use golang net/rpc for communication to a native rpc server.
type grpcClient struct {
	connection callsGrpcCollector //CallsRPC
	pluginType plugin.PluginType
	// encoder    encoding.Encoder
	// encrypter  *encrypter.Encrypter
}

// NewCollectorGrpcClient returns a collector gRPC Client.
func NewCollectorGrpcClient(address string, port int, pub *rsa.PublicKey, secure bool) (PluginCollectorClient, error) {
	return newGrpcClient(address, port)
}

// func newGrpcClient(addr string, port int, pub *rsa.PublicKey, secure bool) (*grpcClient, error) {
func newGrpcClient(addr string, port int) (*grpcClient, error) {
	conn, err := rpcutil.GetClientConnection(addr, port)
	if err != nil {
		return nil, err
	}
	p := &grpcClient{
		connection: rpc.NewCollectorClient(conn),
	}
	// p.encoder = encoding.NewGobEncoder()

	// if secure {
	// 	key, err := encrypter.GenerateKey()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	encrypter := encrypter.New(pub, nil)
	// 	encrypter.Key = key
	// 	p.encrypter = encrypter
	// 	p.encoder.SetEncrypter(encrypter)
	// }

	return p, nil
}

func (g *grpcClient) Ping() error {
	return nil
}

func (g *grpcClient) SetKey() error {
	return nil
}

func (g *grpcClient) Kill(reason string) error {
	return nil
}

func (g *grpcClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	return nil
}

func (g *grpcClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	return "", nil, nil
}

func (g *grpcClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	var results []core.Metric
	return results, nil
}

func (g *grpcClient) GetMetricTypes(config plugin.ConfigType) ([]core.Metric, error) {
	var results []core.Metric
	return results, nil
}

func (g *grpcClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return nil, nil
}
