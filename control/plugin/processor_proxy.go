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

package plugin

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/grpc/common"
)

type ProcessorArgs struct {
	//PluginMetrics []PluginMetric
	ContentType string
	Content     []byte
	Config      map[string]ctypes.ConfigValue
}

type ProcessorReply struct {
	ContentType string
	Content     []byte
}

type processorPluginProxy struct {
	Plugin  ProcessorPlugin
	Session Session
}

func (p *processorPluginProxy) Process(args []byte, reply *[]byte) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()

	dargs := &ProcessorArgs{}
	err := p.Session.Decode(args, dargs)
	if err != nil {
		return err
	}

	r := ProcessorReply{}
	r.ContentType, r.Content, err = p.Plugin.Process(dargs.ContentType, dargs.Content, dargs.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Processor call error: %v", err.Error()))
	}

	*reply, err = p.Session.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

type gRPCProcessorProxy struct {
	Plugin  ProcessorPlugin
	Session Session
	gRPCPluginProxy
}

func (p *gRPCProcessorProxy) Process(ctx context.Context, arg *rpc.ProcessArg) (*rpc.ProcessReply, error) {
	defer catchPluginPanic(p.Session.Logger())
	ct, content, err := p.Plugin.Process(arg.ContentType, arg.Content, common.ParseConfig(arg.Config))
	reply := &rpc.ProcessReply{
		ContentType: ct,
		Content:     content,
	}
	if err != nil {
		reply.Error = err.Error()
	}
	return reply, nil
}
