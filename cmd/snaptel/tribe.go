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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	"github.com/urfave/cli"
)

func listMembers(ctx *cli.Context) error {
	resp := pClient.ListMembers()
	if resp.Err != nil {
		return fmt.Errorf("Error getting members:\n%v\n", resp.Err)
	}

	if len(resp.Members) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		defer w.Flush()
		printFields(w, false, 0,
			"Name",
		)
		for _, m := range resp.Members {
			printFields(w, false, 0, m)
		}
	} else {
		fmt.Println("None")
	}

	return nil
}

func showMember(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.GetMember(ctx.Args().First())
	if resp.Err != nil {
		return fmt.Errorf("Error:\n%v\n", resp.Err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()
	fields := []interface{}{"Name", "Plugin Agreement", "Task Agreements"}
	if ctx.Bool("verbose") {
		fields = append(fields, "tags")
	}
	printFields(w, false, 0,
		fields...,
	)
	var tasks bytes.Buffer
	for idx, task := range resp.TaskAgreements {
		tasks.WriteString(task)
		if idx < (len(resp.TaskAgreements) - 1) {
			tasks.WriteString(",")
		}
	}
	tags, err := json.Marshal(resp.Tags)
	if err != nil {
		return fmt.Errorf("Error:\n%v\n", err)
	}

	values := []interface{}{resp.Name, resp.PluginAgreement, tasks.String()}
	if ctx.Bool("verbose") {
		values = append(values, string(tags))
	}

	printFields(w, false, 0, values...)

	return nil
}

func listAgreements(ctx *cli.Context) error {
	resp := pClient.ListAgreements()
	if resp.Err != nil {
		return fmt.Errorf("Error getting members:\n%v\n", resp.Err)
	}
	printAgreements(resp.Agreements)
	return nil
}

func createAgreement(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.AddAgreement(ctx.Args().First())
	if resp.Err != nil {
		return fmt.Errorf("Error creating agreement: %v\n", resp.Err)
	}
	printAgreements(resp.Agreements)
	return nil
}

func deleteAgreement(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.DeleteAgreement(ctx.Args().First())
	if resp.Err != nil {
		return fmt.Errorf("Error: %v\n", resp.Err)
	}
	printAgreements(resp.Agreements)
	return nil
}

func joinAgreement(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.JoinAgreement(ctx.Args().First(), ctx.Args().Get(1))
	if resp.Err != nil {
		return fmt.Errorf("Error: %v\n", resp.Err)
	}
	printAgreements(map[string]*agreement.Agreement{resp.Agreement.Name: resp.Agreement})
	return nil
}

func leaveAgreement(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.LeaveAgreement(ctx.Args().First(), ctx.Args().Get(1))
	if resp.Err != nil {
		return fmt.Errorf("Error: %v\n", resp.Err)
	}
	printAgreements(map[string]*agreement.Agreement{resp.Agreement.Name: resp.Agreement})
	return nil
}

func agreementMembers(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	resp := pClient.GetAgreement(ctx.Args().First())
	if resp.Err != nil {
		return fmt.Errorf("Error: %v\n", resp.Err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()
	printFields(w, false, 0, "Name")
	for _, v := range resp.Agreement.Members {
		printFields(w, false, 0, v.Name)
	}

	return nil
}

func printAgreements(agreements map[string]*agreement.Agreement) {
	if len(agreements) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		defer w.Flush()
		printFields(w, false, 0,
			"Name", "Number of Members", "plugins", "tasks",
		)

		var keys []string
		for k := range agreements {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := agreements[k]
			var plugins interface{}
			var tasks interface{}
			if v.PluginAgreement != nil {
				plugins = len(v.PluginAgreement.Plugins)
			}
			if v.TaskAgreement != nil {
				tasks = len(v.TaskAgreement.Tasks)
			}
			printFields(w, false, 0, v.Name, len(v.Members), plugins, tasks)
		}
	} else {
		fmt.Println("None")
	}
}
