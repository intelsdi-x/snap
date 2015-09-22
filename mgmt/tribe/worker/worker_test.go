package worker

// import (
// 	"sync"
// 	"testing"
// 	"time"
//
// 	log "github.com/Sirupsen/logrus"
// 	. "github.com/smartystreets/goconvey/convey"
//
// 	"github.com/intelsdi-x/pulse/core"
// 	"github.com/intelsdi-x/pulse/core/perror"
// )
//
// type mockPluginManager struct{}
//
// func (m *mockPluginManager) Load(path string) (core.CatalogedPlugin, perror.PulseError) {
// 	return nil, nil
// }
//
// func (m *mockPluginManager) Unload(plugin core.Plugin) (core.CatalogedPlugin, perror.PulseError) {
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
