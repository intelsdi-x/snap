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

package control

import (
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/internal/common"
)

type configProxy struct {
	cfg *Config
}

func (cfg *configProxy) GetPluginConfigDataNode(ctx context.Context, r *rpc.ConfigDataNodeRequest) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	Type := int(r.PluginType)
	Name := r.Name
	Version := int(r.Version)
	result_node := cfg.cfg.GetPluginConfigDataNode(core.PluginType(Type), Name, Version)
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

func (cfg *configProxy) GetPluginConfigDataNodeAll(ctx context.Context, _ *common.Empty) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	result_node := cfg.cfg.GetPluginConfigDataNodeAll()
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

func (cfg *configProxy) MergePluginConfigDataNode(ctx context.Context, r *rpc.MergeConfigDataNodeRequest) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	Type := core.PluginType(int(r.Request.PluginType))
	Name := r.Request.Name
	Version := int(r.Request.Version)
	node := cdata.NewNode()
	err := json.Unmarshal(r.DataNode.Node, &node)
	if err != nil {
		return nil, err
	}
	result_node := cfg.cfg.MergePluginConfigDataNode(Type, Name, Version, node)
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

func (cfg *configProxy) MergePluginConfigDataNodeAll(ctx context.Context, r *rpc.ConfigDataNode) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	node := cdata.NewNode()
	err := json.Unmarshal(r.Node, &node)
	if err != nil {
		return nil, err
	}
	result_node := cfg.cfg.MergePluginConfigDataNodeAll(node)
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

func (cfg *configProxy) DeletePluginConfigDataNodeField(ctx context.Context, r *rpc.DeleteConfigDataNodeFieldRequest) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	Fields := r.Fields
	Type := core.PluginType(int(r.Request.PluginType))
	Name := r.Request.Name
	Version := int(r.Request.Version)
	result_node := cfg.cfg.DeletePluginConfigDataNodeField(Type, Name, Version, Fields...)
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

func (cfg *configProxy) DeletePluginConfigDataNodeFieldAll(ctx context.Context, r *rpc.DeleteConfigDataNodeFieldAllRequest) (*rpc.ConfigDataNode, error) {
	reply := &rpc.ConfigDataNode{}
	Fields := r.Fields
	result_node := cfg.cfg.DeletePluginConfigDataNodeFieldAll(Fields...)
	reply.Node, _ = nodeToJSON(&result_node)
	return reply, nil
}

//--------Utility functions--------------------------------

func nodeToJSON(in *cdata.ConfigDataNode) ([]byte, error) {
	return json.Marshal(in)
}
