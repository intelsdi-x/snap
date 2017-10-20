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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/intelsdi-x/snap/core/ctypes"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type config struct {
	RestAPI restAPIConfig `json:"rest"`
}
type restAPIConfig struct {
	Password *string `json:"rest-auth-pwd"`
}

func (c *config) loadConfig(path string) error {
	log.WithFields(log.Fields{
		"_module":     "snaptel-config",
		"_block":      "loadConfig",
		"config_path": path,
	}).Warning("The snaptel configuration file will be deprecated. Find more information here: https://github.com/intelsdi-x/snap/issues/1539")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Unable to read config. File might not exist.")
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return fmt.Errorf("Invalid config")
	}
	return nil
}

func getConfig(ctx *cli.Context) error {
	pDetails := filepath.SplitList(ctx.Args().First())
	var ptyp string
	var pname string
	var pver int
	var err error

	if len(pDetails) == 3 {
		ptyp = pDetails[0]
		pname = pDetails[1]
		pver, err = strconv.Atoi(pDetails[2])
		if err != nil {
			return newUsageError("Can't convert version string to integer", ctx)
		}
	} else {
		ptyp = ctx.String("plugin-type")
		pname = ctx.String("plugin-name")
		pver = ctx.Int("plugin-version")
	}

	if ptyp == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pname == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if pver < 1 {
		return newUsageError("Plugin version must be greater than zero", ctx)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()
	r := pClient.GetPluginConfig(ptyp, pname, strconv.Itoa(pver))
	if r.Err != nil {
		return fmt.Errorf("Error requesting info: %v", r.Err)
	}
	printFields(w, false, 0,
		"NAME",
		"VALUE",
		"TYPE",
	)
	for k, v := range r.Table() {
		switch t := v.(type) {
		case ctypes.ConfigValueInt:
			printFields(w, false, 0, k, t.Value, t.Type())
		case ctypes.ConfigValueBool:
			printFields(w, false, 0, k, t.Value, t.Type())
		case ctypes.ConfigValueFloat:
			printFields(w, false, 0, k, t.Value, t.Type())
		case ctypes.ConfigValueStr:
			printFields(w, false, 0, k, t.Value, t.Type())
		}
	}

	return nil
}
