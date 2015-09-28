/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

	"github.com/intelsdi-x/pulse/core/ctypes"
)

type PublishArgs struct {
	//PluginMetrics []PluginMetric
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

func (p *publisherPluginProxy) GetConfigPolicy(args GetConfigPolicyArgs, reply *GetConfigPolicyReply) error {
	defer catchPluginPanic(p.Session.Logger())

	p.Session.Logger().Println("GetConfigPolicy called")
	p.Session.ResetHeartbeat()

	reply.Policy = p.Plugin.GetConfigPolicy()

	return nil
}

func (p *publisherPluginProxy) Publish(args PublishArgs, reply *PublishReply) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()
	err := p.Plugin.Publish(args.ContentType, args.Content, args.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Publish call error: %v", err.Error()))
	}
	return nil
}
