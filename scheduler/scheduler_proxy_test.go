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

package scheduler

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/rpc"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchedulerProxy(t *testing.T) {
	l, _ := net.Listen("tcp", ":0")
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	defer l.Close()
	scheduler := New(ListenPortOption(port))
	c := new(mockMetricManager)
	scheduler.SetMetricManager(c)
	err := scheduler.Start()
	Convey("Scheduler started", t, func() {
		So(err, ShouldBeNil)
	})

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", DefaultListenAddr, port), grpc.WithInsecure())
	Convey("RPC endpoint dialed", t, func() {
		So(err, ShouldBeNil)
	})

	client := rpc.NewTaskManagerClient(conn)
	Convey("RPC client created", t, func() {
		So(client, ShouldNotBeNil)
	})

	Convey("CreateTask", t, func() {
		Convey("provided a schedule and workflow", func() {
			sampleWFMap := wmap.Sample()
			sch := schedule.NewSimpleSchedule(time.Millisecond * 10)
			arg := &rpc.CreateTaskArg{Start: true}
			arg.WmapJson, err = sampleWFMap.ToJson()
			So(err, ShouldBeNil)
			arg.ScheduleJson, err = json.Marshal(sch)
			So(err, ShouldBeNil)
			reply, err := client.CreateTask(context.Background(), arg)
			So(err, ShouldBeNil)
			So(reply, ShouldNotBeNil)
			So(reply.Errors, ShouldNotBeEmpty)
			So(reply.Errors[0].ErrorString, ShouldResemble, "metric validation error")
			So(reply.Task, ShouldNotBeNil)
			So(reply.Task.Id, ShouldNotResemble, "")
			Convey("GetTask - providied the task we just created", func() {
				task, err := client.GetTask(context.Background(), &rpc.GetTaskArg{Id: reply.Task.Id})
				So(err, ShouldBeNil)
				So(task, ShouldNotBeNil)
				So(task.Id, ShouldResemble, reply.Task.Id)
			})
		})
	})

	Convey("GetTask", t, func() {
		Convey("provided an invalid task id", func() {
			task, err := client.GetTask(context.Background(), &rpc.GetTaskArg{Id: "asdf"})
			So(task, ShouldBeNil)
			So(grpc.ErrorDesc(err), ShouldResemble, "Task not found: ID(asdf)")
		})

	})

}
