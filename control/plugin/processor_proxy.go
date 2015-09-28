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

func (p *processorPluginProxy) GetConfigPolicy(args GetConfigPolicyArgs, reply *GetConfigPolicyReply) error {
	defer catchPluginPanic(p.Session.Logger())

	p.Session.Logger().Println("GetConfigPolicy called")
	p.Session.ResetHeartbeat()

	reply.Policy = p.Plugin.GetConfigPolicy()

	return nil
}

func (p *processorPluginProxy) Process(args ProcessorArgs, reply *ProcessorReply) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()
	var err error
	reply.ContentType, reply.Content, err = p.Plugin.Process(args.ContentType, args.Content, args.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Processor call error: %v", err.Error()))
	}
	return nil
}
