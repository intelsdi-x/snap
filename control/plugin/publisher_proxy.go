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

type PublishArgs struct {
	ContentType string
	Content     []byte
	Config      map[string]ctypes.ConfigValue
}

type PublishReply struct {
}

type publisherPluginProxy struct {
	Plugin  PublisherPlugin
	Session Session
}

func (p *publisherPluginProxy) Publish(args []byte, reply *[]byte) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()

	dargs := &PublishArgs{}
	err := p.Session.Decode(args, dargs)
	if err != nil {
		return err
	}

	err = p.Plugin.Publish(dargs.ContentType, dargs.Content, dargs.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Publish call error: %v", err.Error()))
	}
	return nil
}

type gRPCPublisherProxy struct {
	Plugin  PublisherPlugin
	Session Session
	gRPCPluginProxy
}

func (p *gRPCPublisherProxy) Publish(ctx context.Context, arg *rpc.PublishArg) (*common.Empty, error) {
	defer catchPluginPanic(p.Session.Logger())
	err := p.Plugin.Publish(arg.ContentType, arg.Content, common.ParseConfig(arg.Config))
	if err != nil {
		return &common.Empty{}, err
	}
	return &common.Empty{}, nil
}
