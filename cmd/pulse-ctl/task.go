package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

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
	Name     string
}

func createTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	file, e := ioutil.ReadFile(ctx.Args().First())
	if e != nil {
		fmt.Printf("File error - %v\n", e)
		os.Exit(1)
	}

	t := task{}
	e = json.Unmarshal(file, &t)
	if e != nil {
		fmt.Printf("json error - %v\n", e)
	}

	t.Name = ctx.String("name")

	if t.Version != 1 {
		fmt.Println("Invalid version provided")
		os.Exit(1)

	}

	ct := client.NewTask(t.Schedule, t.Workflow, t.Name)

	e = client.CreateTask(ct)
	if e != nil {
		fmt.Printf("Error creating task - %v\n", e)
		os.Exit(1)
	}
}

func listTask(ctx *cli.Context) {
	tasks, err := client.GetTasks()
	if err != nil {
		fmt.Printf("Error getting tasks - %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0,
		"ID",
		"STATE",
		"HIT COUNT",
		"MISS COUNT",
		"FAILURE COUNT",
		"LAST FAILURE MSG",
		"CREATION TIME",
	)
	for _, task := range tasks {
		printFields(w, false, 0,
			task.ID,
			task.Name,
			task.State,
			task.HitCount,
			task.MissCount,
			task.FailedCount,
			task.LastFailureMessage,
			task.CreationTime.Format(time.RFC1123),
		)
	}
	w.Flush()
}

func startTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	id, err := strconv.ParseUint(ctx.Args().First(), 0, 64)
	if err != nil {
		fmt.Printf("Incorrect usage - %v\n", err.Error())
		os.Exit(1)
	}
	err = client.StartTask(id)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func stopTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	id, err := strconv.ParseUint(ctx.Args().First(), 0, 64)
	if err != nil {
		fmt.Printf("Incorrect usage - %v\n", err.Error())
		os.Exit(1)
	}
	err = client.StopTask(id)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func removeTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	id, err := strconv.ParseUint(ctx.Args().First(), 0, 64)
	if err != nil {
		fmt.Printf("Incorrect usage - %v\n", err.Error())
		os.Exit(1)
	}
	err = client.RemoveTask(id)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
