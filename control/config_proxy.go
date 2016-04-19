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
	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/internal/common"
)

type configProxy struct {
	cfg *Config
}

func (cfg *configProxy) GetPluginConfigDataNode(ctx context.Context, r *rpc.ConfigDataNodeRequest) (*common.ConfigMap, error) {
	Type := int(r.PluginType)
	Name := r.Name
	Version := int(r.Version)
	resultNode := cfg.cfg.GetPluginConfigDataNode(core.PluginType(Type), Name, Version)
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}

func (cfg *configProxy) GetPluginConfigDataNodeAll(ctx context.Context, _ *common.Empty) (*common.ConfigMap, error) {
	resultNode := cfg.cfg.GetPluginConfigDataNodeAll()
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}

func (cfg *configProxy) MergePluginConfigDataNode(ctx context.Context, r *rpc.MergeConfigDataNodeRequest) (*common.ConfigMap, error) {
	Type := core.PluginType(int(r.Request.PluginType))
	Name := r.Request.Name
	Version := int(r.Request.Version)
	node := common.ConfigMapToConfig(r.Config)
	resultNode := cfg.cfg.MergePluginConfigDataNode(Type, Name, Version, node)
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}

func (cfg *configProxy) MergePluginConfigDataNodeAll(ctx context.Context, r *common.ConfigMap) (*common.ConfigMap, error) {
	node := common.ConfigMapToConfig(r)
	resultNode := cfg.cfg.MergePluginConfigDataNodeAll(node)
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}

func (cfg *configProxy) DeletePluginConfigDataNodeField(ctx context.Context, r *rpc.DeleteConfigDataNodeFieldRequest) (*common.ConfigMap, error) {
	Fields := r.Fields
	Type := core.PluginType(int(r.Request.PluginType))
	Name := r.Request.Name
	Version := int(r.Request.Version)
	resultNode := cfg.cfg.DeletePluginConfigDataNodeField(Type, Name, Version, Fields...)
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}

func (cfg *configProxy) DeletePluginConfigDataNodeFieldAll(ctx context.Context, r *rpc.DeleteConfigDataNodeFieldAllRequest) (*common.ConfigMap, error) {
	Fields := r.Fields
	resultNode := cfg.cfg.DeletePluginConfigDataNodeFieldAll(Fields...)
	reply := common.ConfigToConfigMap(&resultNode)
	return reply, nil
}
