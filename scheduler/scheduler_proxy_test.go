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
	"net"
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
	cfg := GetDefaultConfig()
	l, _ := net.Listen("tcp", ":0")
	l.Close()
	cfg.ListenPort = l.Addr().(*net.TCPAddr).Port
	scheduler := New(cfg)
	c := new(mockMetricManager)
	scheduler.SetMetricManager(c)
	err := scheduler.Start()
	Convey("Scheduler started", t, func() {
		Convey("So err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})

	client, err := NewClient(scheduler.Config().ListenAddr, scheduler.Config().ListenPort)
	Convey("RPC Client to scheduler rpc server should be created ", t, func() {
		Convey("So err should be nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("So client should be created", func() {
			So(client, ShouldNotBeNil)
		})
	})

	Convey("CreateTask", t, func() {
		Convey("provided a simple schedule and workflow", func() {
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
			Convey("GetTask - provided the task we just created", func() {
				task, err := client.GetTask(context.Background(), &rpc.GetTaskArg{Id: reply.Task.Id})
				So(err, ShouldBeNil)
				So(task, ShouldNotBeNil)
				So(task.Id, ShouldResemble, reply.Task.Id)
			})
		})
		Convey("provided a windowed schedule and workflow", func() {
			sampleWFMap := wmap.Sample()
			start := time.Now()
			stop := time.Now().Add(3 * time.Second)
			sch := schedule.NewWindowedSchedule(time.Millisecond*10, &start, &stop)
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
			Convey("GetTask - provided the task we just created", func() {
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
