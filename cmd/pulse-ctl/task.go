package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/client"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

// type schedule struct {
// 	Type     string `json:"type"`
// 	Interval string `json:"interval"`
// }

type task struct {
	Version  int
	Schedule *pulse.Schedule
	Workflow *wmap.WorkflowMap
}

func createTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	file, e := ioutil.ReadFile(ctx.Args().First())
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	t := task{}
	e = json.Unmarshal(file, &t)
	if e != nil {
		fmt.Printf("json error: %v\n", e)
	}

	if t.Version != 1 {
		fmt.Println("Invalid version provided")
		os.Exit(1)

	}

	ct := client.NewTask(t.Schedule, t.Workflow)

	e = client.CreateTask(ct)
	if e != nil {
		fmt.Printf("error creating task: %v\n", e)
		os.Exit(1)
	}
}
