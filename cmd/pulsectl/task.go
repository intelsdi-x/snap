package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
	"github.com/intelsdi-x/pulse/scheduler/wmap"

	"github.com/ghodss/yaml"
)

type task struct {
	Version  int
	Schedule *client.Schedule
	Workflow *wmap.WorkflowMap
	Name     string
}

func createTask(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		os.Exit(1)
	}

	path := ctx.Args().First()
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)
	if e != nil {
		fmt.Printf("File error - %v\n", e)
		os.Exit(1)
	}

	t := task{}
	switch ext {
	case ".yaml", ".yml":
		e = yaml.Unmarshal(file, &t)
		if e != nil {
			fmt.Printf("Error parsing YAML file input - %v\n", e)
			os.Exit(1)
		}
	case ".json":
		e = json.Unmarshal(file, &t)
		if e != nil {
			fmt.Printf("Error parsing JSON file input - %v\n", e)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unsupported file type %s\n", ext)
		os.Exit(1)
	}

	t.Name = ctx.String("name")
	if t.Version != 1 {
		fmt.Println("Invalid version provided")
		os.Exit(1)
	}

	r := pClient.CreateTask(t.Schedule, t.Workflow, t.Name)

	if r.Err != nil {
		fmt.Printf("Error creating task:\n%v\n", r.Err)
		os.Exit(1)
	}
	fmt.Println("Task created")
	fmt.Printf("ID: %d\n", r.ID)
	fmt.Printf("Name: %s\n", r.Name)
	fmt.Printf("State: %s\n", r.State)
}

func listTask(ctx *cli.Context) {
	tasks := pClient.GetTasks()
	if tasks.Err != nil {
		fmt.Printf("Error getting tasks:\n%v\n", tasks.Err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0,
		"ID",
		"NAME",
		"STATE",
		"HIT COUNT",
		"MISS COUNT",
		"FAILURE COUNT",
		"CREATION TIME",
		"LAST FAILURE MSG",
	)
	for _, task := range tasks.ScheduledTasks {
		printFields(w, false, 0,
			task.ID,
			task.Name,
			task.State,
			task.HitCount,
			task.MissCount,
			task.FailedCount,
			task.CreationTime().Format(timeFormat),
			task.LastFailureMessage,
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
	r := pClient.StartTask(int(id))
	if r.Err != nil {
		fmt.Printf("Error starting task:\n%v\n", r.Err)
		os.Exit(1)
	}
	fmt.Println("Task started:")
	fmt.Printf("ID: %d\n", r.ID)
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
	r := pClient.StopTask(int(id))
	if r.Err != nil {
		fmt.Printf("Error stopping task:\n%v\n", r.Err)
		os.Exit(1)
	}
	fmt.Println("Task stopped:")
	fmt.Printf("ID: %d\n", r.ID)
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
	r := pClient.RemoveTask(int(id))
	if r.Err != nil {
		fmt.Printf("Error stopping task:\n%v\n", r.Err)
		os.Exit(1)
	}
	fmt.Println("Task removed:")
	fmt.Printf("ID: %d\n", r.ID)
}
