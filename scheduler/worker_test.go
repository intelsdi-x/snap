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
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorker(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("Replies on the Job's reply chan", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go w.start()
		mj := &mockJob{
			replchan:  make(chan struct{}),
			starttime: time.Now(),
			deadline:  time.Now().Add(1 * time.Second),
		}
		rcv <- mj
		<-mj.ReplChan()
		So(mj.worked, ShouldEqual, true)
	})
	Convey("replies without running job if deadline is exceeded", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go w.start()
		mj := &mockJob{
			replchan:  make(chan struct{}),
			starttime: time.Now(),
			deadline:  time.Now().Add(1 * time.Second),
		}
		time.Sleep(time.Millisecond * 1500)
		rcv <- mj
		<-mj.replchan
		So(mj.worked, ShouldEqual, false)
	})
	Convey("stops the worker if kamikaze chan is closed", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go func() { close(w.kamikaze) }()
		w.start()
		So(0, ShouldEqual, 0)
	})
}
