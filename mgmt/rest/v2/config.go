/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/julienschmidt/httprouter"
)

type PolicyTable cpolicy.RuleTable

type PolicyTableSlice []cpolicy.RuleTable

// PluginConfigItem represents cdata.ConfigDataNode which implements it's own UnmarshalJSON.
//
// swagger:response PluginConfigResponse
type PluginConfigItem struct {
	//in: body
	Config cdata.ConfigDataNode `json:"config"`
}

// PluginConfigParam type
//
//swagger:parameters setPluginConfigItem
type PluginConfigParam struct {
	// in: formData
	Config string `json:"config"`
}

// PluginConfigDeleteParams defines parameters for deleting a config.
//
// swagger:parameters deletePluginConfigItem
type PluginConfigDeleteParams struct {
	// required: true
	// in: path
	PName string `json:"pname"`
	// required: true
	// in: path
	PVersion int `json:"pversion"`
	// required: true
	// in: path
	// enum: collector, processor, publisher
	PType string `json:"ptype"`
	// in: formData
	// required: true
	Config []string `json:"config"`
}

func (s *apiV2) getPluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	styp := p.ByName("type")
	if styp == "" {
		cdn := s.configManager.GetPluginConfigDataNodeAll()
		item := &PluginConfigItem{cdn}
		Write(200, item, w)
		return
	}

	typ, err := getPluginType(styp)
	if err != nil {
		Write(400, FromError(err), w)
		return
	}

	name := p.ByName("name")
	sver := p.ByName("version")
	iver := -2
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	cdn := s.configManager.GetPluginConfigDataNode(typ, name, iver)
	item := &PluginConfigItem{cdn}
	Write(200, item, w)
}

func (s *apiV2) deletePluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	var typ core.PluginType
	styp := p.ByName("type")

	if styp != "" {
		typ, err = getPluginType(styp)
		if err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	name := p.ByName("name")
	sver := p.ByName("version")

	src, err := deletePluginConfigItemHelper(r)
	if err != nil {
		Write(400, FromError(err), w)
		return
	}

	iver := -2
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	if len(src) == 0 {
		src = []string{}
		errCode, err := core.UnmarshalBody(&src, r.Body)
		if errCode != 0 && err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	var res cdata.ConfigDataNode
	if styp == "" {
		res = s.configManager.DeletePluginConfigDataNodeFieldAll(src...)
	} else {
		res = s.configManager.DeletePluginConfigDataNodeField(typ, name, iver, src...)
	}

	item := &PluginConfigItem{res}
	Write(200, item, w)
}

func (s *apiV2) setPluginConfigItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	var typ core.PluginType
	styp := p.ByName("type")

	if styp != "" {
		typ, err = getPluginType(styp)
		if err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	name := p.ByName("name")
	sver := p.ByName("version")

	err = setPluginConfigItemHelper(r)
	if err != nil {
		Write(400, FromError(err), w)
		return
	}

	iver := -2
	if sver != "" {
		if iver, err = strconv.Atoi(sver); err != nil {
			Write(400, FromError(err), w)
			return
		}
	}

	src := cdata.NewNode()
	errCode, err := core.UnmarshalBody(src, r.Body)
	if errCode != 0 && err != nil {
		Write(400, FromError(err), w)
		return
	}

	var res cdata.ConfigDataNode
	if styp == "" {
		res = s.configManager.MergePluginConfigDataNodeAll(src)
	} else {
		res = s.configManager.MergePluginConfigDataNode(typ, name, iver, src)
	}

	item := &PluginConfigItem{res}
	Write(200, item, w)
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

// deletePluginConfigItemHelper builds different forms of request data into the way method deletePluginConfigItem accepts.
// currently it accepts go-swagger client, swagger-ui and SanpCLI.
func deletePluginConfigItemHelper(r *http.Request) ([]string, error) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	dm := map[string][]string{}
	err = json.Unmarshal(buf, &dm)

	sw := true
	// No error needs to be returned here.
	// As it explores different request formats.
	// There is no way to detect a string is URLEncoded.
	if err != nil {
		sw = false

		// go-swagger sends url-encoded form data.
		// Unescaping is necessary.
		data, _ := url.QueryUnescape(string(buf))
		tokens := strings.Split(data, "=")
		if len(tokens) == 2 {
			itokens := strings.Split(tokens[1], ",")
			dm["config"] = itokens
			sw = true
		}
	}

	var src []string
	if sw {
		src = dm["config"]
	}
	return src, nil
}

// setPluginConfigItemHelper builds different forms of request data into the way method setPluginConfigItem accepts.
// currently it accepts go-swagger client, swagger-ui and SanpCLI.
func setPluginConfigItemHelper(r *http.Request) error {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	dm := map[string]string{}
	err = json.Unmarshal(buf, &dm)

	sw := true
	// No error needs to be returned here.
	// As it explores different request formats.
	// There is no way to detect a string is URLEncoded.
	if err != nil {
		sw = false

		data, _ := url.QueryUnescape(string(buf))
		tokens := strings.Split(data, "=")
		if len(tokens) == 2 {
			dm["config"] = tokens[1]
			sw = true
		}
	}

	if sw {
		cfg := dm["config"]
		r.Body = ioutil.NopCloser(strings.NewReader(cfg))
		r.ContentLength = int64(len(cfg))
	}
	return nil
}
