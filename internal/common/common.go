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

package common

import (
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

func ConfigMapToConfig(cfg *ConfigMap) *cdata.ConfigDataNode {
	config := cdata.FromTable(ParseConfig(cfg))

	return config
}

func ParseConfig(config *ConfigMap) map[string]ctypes.ConfigValue {
	c := make(map[string]ctypes.ConfigValue)
	for k, v := range config.IntMap {
		ival := ctypes.ConfigValueInt{Value: int(v)}
		c[k] = ival
	}
	for k, v := range config.FloatMap {
		fval := ctypes.ConfigValueFloat{Value: v}
		c[k] = fval
	}
	for k, v := range config.StringMap {
		sval := ctypes.ConfigValueStr{Value: v}
		c[k] = sval
	}
	for k, v := range config.BoolMap {
		bval := ctypes.ConfigValueBool{Value: v}
		c[k] = bval
	}
	return c
}

func ConfigToConfigMap(cd *cdata.ConfigDataNode) *ConfigMap {

	return ToConfigMap(cd.Table())
}

func ToConfigMap(cv map[string]ctypes.ConfigValue) *ConfigMap {
	newConfig := &ConfigMap{
		IntMap:    make(map[string]int64),
		FloatMap:  make(map[string]float64),
		StringMap: make(map[string]string),
		BoolMap:   make(map[string]bool),
	}
	for k, v := range cv {
		switch v.Type() {
		case "integer":
			newConfig.IntMap[k] = int64(v.(ctypes.ConfigValueInt).Value)
		case "float":
			newConfig.FloatMap[k] = v.(ctypes.ConfigValueFloat).Value
		case "string":
			newConfig.StringMap[k] = v.(ctypes.ConfigValueStr).Value
		case "bool":
			newConfig.BoolMap[k] = v.(ctypes.ConfigValueBool).Value
		}
	}
	return newConfig
}
