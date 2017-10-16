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

package worker

// import (
// 	"sync"
// 	"testing"
// 	"time"
//
// 	log "github.com/sirupsen/logrus"
// 	. "github.com/smartystreets/goconvey/convey"
//
// 	"github.com/intelsdi-x/snap/core"
// 	"github.com/intelsdi-x/snap/core/serror"
// )
//
// type mockPluginManager struct{}
//
// func (m *mockPluginManager) Load(path string) (core.CatalogedPlugin, serror.SnapError) {
// 	return nil, nil
// }
//
// func (m *mockPluginManager) Unload(plugin core.Plugin) (core.CatalogedPlugin, serror.SnapError) {
// 	return nil, nil
// }
//
// func (m *mockPluginManager) PluginCatalog() core.PluginCatalog {
// 	return nil
// }
//
// type mockMemberManager struct{}
//
// func (m *mockMemberManager) getPluginAgreementMembers() ([]*member, error) {
// 	return nil, nil
// }
//
// func (m *mockMemberManager) getTaskAgreementMembers() ([]*member, error) {
// 	return nil, nil
// }
//
// func TestTaskWorker(t *testing.T) {
// 	log.SetLevel(log.DebugLevel)
// 	Convey("Create a worker", t, func() {
// 		pwq := make(chan chan pluginRequest, 999)
// 		twq := make(chan chan taskRequest, 999)
// 		qc := make(chan interface{})
// 		wg := &sync.WaitGroup{}
// 		pm := &mockPluginManager{}
// 		mm := &mockMemberManager{}
// 		worker := newWorker(1, pwq, twq, qc, wg, pm, mm)
// 		So(worker, ShouldNotBeNil)
// 		worker.start()
//
// 		Convey("Dispatch work (for a task)", func() {
// 			taskRequest := taskRequest{
// 				task: task{ID: 1234},
// 			}
// 			worker := <-twq
// 			worker <- taskRequest
// 			time.Sleep(1 * time.Second)
// 		})
// 	})
//
// }
