/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package plugin

import (
	"time"

	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/grpc/common"
	"golang.org/x/net/context"
)

type gRPCPluginProxy struct {
	plugin  Plugin
	session Session
}

func (g *gRPCPluginProxy) SetKey(ctx context.Context, arg *rpc.SetKeyArg) (*rpc.SetKeyReply, error) {
	out, err := g.session.DecryptKey(arg.Key)
	if err != nil {
		return &rpc.SetKeyReply{Error: err.Error()}, nil
	}
	g.session.setKey(out)
	return &rpc.SetKeyReply{}, nil
}

func (g *gRPCPluginProxy) Ping(ctx context.Context, arg *common.Empty) (*rpc.PingReply, error) {
	g.session.ResetHeartbeat()
	return &rpc.PingReply{}, nil
}

func (g *gRPCPluginProxy) Kill(ctx context.Context, arg *rpc.KillRequest) (*rpc.KillReply, error) {
	killChan := g.session.KillChan()
	g.session.Logger().Printf("Kill called by agent, reason: %s\n", arg.Reason)
	go func() {
		time.Sleep(time.Second * 2)
		killChan <- 0
	}()
	return &rpc.KillReply{}, nil
}

func (g *gRPCPluginProxy) GetConfigPolicy(ctx context.Context, arg *common.Empty) (*rpc.GetConfigPolicyReply, error) {
	defer catchPluginPanic(g.session.Logger())

	g.session.Logger().Println("GetConfigPolicy called")

	policy, err := g.plugin.GetConfigPolicy()
	if err != nil {
		return &rpc.GetConfigPolicyReply{
			Error: err.Error(),
		}, nil
	}

	reply, err := rpc.NewGetConfigPolicyReply(policy)
	if err != nil {
		return nil, err
	}

	return reply, nil
}
