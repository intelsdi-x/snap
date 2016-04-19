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

package rest

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/internal/common"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/julienschmidt/httprouter"
)

func (s *Server) getPluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	styp := p.ByName("type")
	if styp == "" {
		cfg, err := s.mc.GetPluginConfigDataNodeAll(context.Background(), &common.Empty{})
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		cdn := common.ConfigMapToConfig(cfg)
		item := &rbody.PluginConfigItem{ConfigDataNode: *cdn}
		respond(200, item, w)
		return
	}

	typ, err := getPluginType(styp)
	if err != nil {
		respond(400, rbody.FromError(err), w)
		return
	}

	name := p.ByName("name")
	sver := p.ByName("version")
	var iver int
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	} else {
		iver = -2
	}
	arg := &rpc.ConfigDataNodeRequest{
		PluginType: int32(typ),
		Name:       name,
		Version:    int64(iver),
	}
	cfg, err := s.mc.GetPluginConfigDataNode(context.Background(), arg)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}

	cdn := common.ConfigMapToConfig(cfg)
	item := &rbody.PluginConfigItem{ConfigDataNode: *cdn}
	respond(200, item, w)
}

func (s *Server) deletePluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	var typ core.PluginType
	styp := p.ByName("type")
	if styp != "" {
		typ, err = getPluginType(styp)
		if err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	}

	name := p.ByName("name")
	sver := p.ByName("version")
	var iver int
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	} else {
		iver = -2
	}

	src := []string{}
	errCode, err := marshalBody(&src, r.Body)
	if errCode != 0 && err != nil {
		respond(400, rbody.FromError(err), w)
		return
	}
	var res *cdata.ConfigDataNode
	if styp == "" {
		arg := &rpc.DeleteConfigDataNodeFieldAllRequest{
			Fields: src,
		}
		cfg, err := s.mc.DeletePluginConfigDataNodeFieldAll(context.Background(), arg)
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		res = common.ConfigMapToConfig(cfg)
	} else {
		arg := &rpc.DeleteConfigDataNodeFieldRequest{
			Request: &rpc.ConfigDataNodeRequest{
				PluginType: int32(typ),
				Name:       name,
				Version:    int64(iver),
			},
			Fields: src,
		}
		cfg, err := s.mc.DeletePluginConfigDataNodeField(context.Background(), arg)
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		res = common.ConfigMapToConfig(cfg)
	}

	item := &rbody.DeletePluginConfigItem{ConfigDataNode: *res}
	respond(200, item, w)
}

func (s *Server) setPluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	var typ core.PluginType
	styp := p.ByName("type")
	if styp != "" {
		typ, err = getPluginType(styp)
		if err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	}

	name := p.ByName("name")
	sver := p.ByName("version")
	var iver int
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			respond(400, rbody.FromError(err), w)
			return
		}
	} else {
		iver = -2
	}

	src := cdata.NewNode()
	errCode, err := marshalBody(src, r.Body)
	if errCode != 0 && err != nil {
		respond(400, rbody.FromError(err), w)
		return
	}
	var res *cdata.ConfigDataNode
	if styp == "" {
		arg := common.ConfigToConfigMap(src)
		cfg, err := s.mc.MergePluginConfigDataNodeAll(context.Background(), arg)
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		res = common.ConfigMapToConfig(cfg)
	} else {
		arg := &rpc.MergeConfigDataNodeRequest{
			Config: common.ConfigToConfigMap(src),
			Request: &rpc.ConfigDataNodeRequest{
				PluginType: int32(typ),
				Name:       name,
				Version:    int64(iver),
			},
		}
		cfg, err := s.mc.MergePluginConfigDataNode(context.Background(), arg)
		if err != nil {
			respond(500, rbody.FromError(err), w)
			return
		}
		res = common.ConfigMapToConfig(cfg)
	}

	item := &rbody.SetPluginConfigItem{ConfigDataNode: *res}
	respond(200, item, w)
}

func getPluginType(t string) (core.PluginType, error) {
	if ityp, err := strconv.Atoi(t); err == nil {
		return core.PluginType(ityp), nil
	}
	ityp, err := core.ToPluginType(t)
	if err != nil {
		return core.PluginType(-1), err
	}
	return ityp, nil
}
