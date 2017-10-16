// +build legacy

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

	"github.com/intelsdi-x/snap/pkg/chrono"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorker(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("runs a job sent to the worker", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan queuedJob)
		w := newWorker(rcv)
		go w.start()
		mj := newMockJob()
		rcv <- newQueuedJob(mj)
		mj.Await()
		So(mj.worked, ShouldEqual, true)
	})
	Convey("replies without running job if deadline is exceeded", t, func() {
		// Make sure global clock is restored after test.
		defer chrono.Chrono.Reset()
		defer chrono.Chrono.Continue()

		// Use artificial time: pause to get base time.
		chrono.Chrono.Pause()

		workerKillChan = make(chan struct{})
		rcv := make(chan queuedJob)
		w := newWorker(rcv)
		go w.start()
		mj := newMockJob()
		// Time travel 1.5 seconds.
		chrono.Chrono.Forward(1500 * time.Millisecond)
		qj := newQueuedJob(mj)
		rcv <- qj
		errors := qj.Promise().Await()
		So(errors, ShouldNotBeEmpty)
		So(mj.worked, ShouldBeFalse)
	})
	Convey("stops the worker if kamikaze chan is closed", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan queuedJob)
		w := newWorker(rcv)
		go func() { close(w.kamikaze) }()
		w.start()
		So(0, ShouldEqual, 0)
	})
}
