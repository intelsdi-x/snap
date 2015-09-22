package worker

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
	"github.com/intelsdi-x/pulse/mgmt/rest/request"
	"github.com/intelsdi-x/pulse/pkg/schedule"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

const (
	PluginLoadedType = iota
	// pluginAddedToAgreementType
)

const (
	TaskCreatedType = iota
)

var workerLogger = log.WithFields(log.Fields{
	"_module": "worker",
})

type PluginRequest struct {
	Plugin      core.Plugin
	RequestType int
}

type TaskRequest struct {
	Task        Task
	RequestType int
}

type Task struct {
	ID uint64
}

type ManagesPlugins interface {
	Load(path string) (core.CatalogedPlugin, perror.PulseError)
	Unload(plugin core.Plugin) (core.CatalogedPlugin, perror.PulseError)
	PluginCatalog() core.PluginCatalog
}

type ManagesTasks interface {
	GetTask(id uint64) (core.Task, error)
	CreateTask(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors)
}

type getsMembers interface {
	GetPluginAgreementMembers() ([]Member, error)
	GetTaskAgreementMembers() ([]Member, error)
}

type Member interface {
	GetAddr() net.IP
	GetRESTAPIPort() string
	GetName() string
}

// newPluginWorker
func newWorker(id int,
	pluginWorkerQueue chan chan PluginRequest,
	taskWorkerQueue chan chan TaskRequest,
	quitChan chan interface{},
	wg *sync.WaitGroup,
	pm ManagesPlugins,
	tm ManagesTasks,
	mm getsMembers) worker {
	// Create, and return the worker.
	worker := worker{
		pluginManager:     pm,
		taskManager:       tm,
		memberManager:     mm,
		id:                id,
		pluginWork:        make(chan PluginRequest),
		pluginWorkerQueue: pluginWorkerQueue,
		taskWork:          make(chan TaskRequest),
		taskWorkerQueue:   taskWorkerQueue,
		quitChan:          make(chan bool)}

	return worker
}

type worker struct {
	pluginManager     ManagesPlugins
	memberManager     getsMembers
	taskManager       ManagesTasks
	id                int
	pluginWork        chan PluginRequest
	pluginWorkerQueue chan chan PluginRequest
	taskWork          chan TaskRequest
	taskWorkerQueue   chan chan TaskRequest
	quitChan          chan bool
	waitGroup         *sync.WaitGroup
}

func DispatchWorkers(nworkers int, pluginWorkQueue chan PluginRequest, taskWorkQueue chan TaskRequest, quitChan chan interface{}, workerWaitGroup *sync.WaitGroup, cp ManagesPlugins, tm ManagesTasks, mm getsMembers) {
	pluginWorkerQueue := make(chan chan PluginRequest, nworkers)
	taskWorkerQueue := make(chan chan TaskRequest, nworkers)

	for i := 0; i < nworkers; i++ {
		workerLogger.Infof("Starting tribe worker-%d", i+1)
		worker := newWorker(i+1, pluginWorkerQueue, taskWorkerQueue, quitChan, workerWaitGroup, cp, tm, mm)
		worker.start()
	}

	go func() {
		for {
			select {
			case pluginWork := <-pluginWorkQueue:
				workerLogger.Debug("Received plugin work requeust")
				go func() {
					pluginWorker := <-pluginWorkerQueue

					workerLogger.Debug("Dispatching plugin work request")
					pluginWorker <- pluginWork
				}()
			case taskWork := <-taskWorkQueue:
				workerLogger.Debug("Received task work requeust")
				go func() {
					taskWorker := <-taskWorkerQueue

					workerLogger.Debug("Dispatching plugin work request")
					taskWorker <- taskWork
				}()
			case <-quitChan:
				workerLogger.Debug("Stopping plugin work dispatcher")
				return
			}
		}
	}()
}

// Start "starts" the worker
func (w worker) start() {
	go func() {
		defer w.waitGroup.Done()
		for {
			// Add ourselves into the worker queue.
			w.pluginWorkerQueue <- w.pluginWork
			w.taskWorkerQueue <- w.taskWork

			select {
			case work := <-w.pluginWork:
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"plugin_name":    work.Plugin.Name(),
					"plugin_version": work.Plugin.Version(),
					"plugin_type":    work.Plugin.TypeName(),
					"worker":         w.id,
					"_block":         "start",
				})
				wLogger.Debug("received work")
				done := false
				for {
					if w.isPluginLoaded(work.Plugin.Name(), work.Plugin.TypeName(), work.Plugin.Version()) {
						break
					}
					members, err := w.memberManager.GetPluginAgreementMembers()
					if err != nil {
						wLogger.Error(err)
						continue
					}
					for _, member := range shuffle(members) {
						url := fmt.Sprintf("http://%s:%s/v1/plugins/%s/%s/%d", member.GetAddr(), member.GetRESTAPIPort(), work.Plugin.TypeName(), work.Plugin.Name(), work.Plugin.Version())
						wLogger.Debugf("worker-%v is trying %v ", w.id, url)
						resp, err := http.Get(url)
						if err != nil {
							wLogger.Error(err)
							continue
						}
						if resp.StatusCode == 200 {
							if resp.Header.Get("Content-Type") != "application/x-gzip" {
								wLogger.WithField("content-type", resp.Header.Get("Content-Type")).Error("Expected application/x-gzip")
							}
							f, err := ioutil.TempFile("", fmt.Sprintf("%s-%s-%d", work.Plugin.TypeName(), work.Plugin.Name(), work.Plugin.Version()))
							if err != nil {
								wLogger.Error(err)
								continue
							}
							io.Copy(f, resp.Body)
							f.Close()
							err = os.Chmod(f.Name(), 0700)
							if err != nil {
								wLogger.Error(err)
								continue
							}
							_, err = w.pluginManager.Load(f.Name())
							if err != nil {
								wLogger.Error(err)
								continue
							}
							if w.isPluginLoaded(work.Plugin.Name(), work.Plugin.TypeName(), work.Plugin.Version()) {
								wLogger.WithField("path", f.Name()).Info("loaded plugin")
								done = true
								break
							}
						}
					}
					if done {
						break
					}
					time.Sleep(200 * time.Millisecond)
				}
			case work := <-w.taskWork:
				done := false
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"task":   work.Task.ID,
					"worker": w.id,
					"_block": "start",
				})
				wLogger.Debug("received work")
				if work.RequestType == TaskCreatedType {
					task, _ := w.taskManager.GetTask(work.Task.ID)
					if task != nil {
						wLogger.Warn("we already have a task with this Id")
					} else {
						for {
							members, err := w.memberManager.GetTaskAgreementMembers()
							if err != nil {
								wLogger.Error(err)
								continue
							}
							for _, member := range shuffle(members) {
								// workerLogger.Error(member)
								c := client.New(fmt.Sprintf("http://%s:%s", member.GetAddr(), member.GetRESTAPIPort()), "v1")
								taskResult := c.GetTask(uint(work.Task.ID))
								if taskResult.Err != nil {
									wLogger.Debug(err)
									continue
								}
								_, err := w.taskManager.CreateTask(getSchedule(taskResult.ScheduledTaskReturned.Schedule), taskResult.Workflow, false)
								if err != nil {
									wLogger.Error(err)
									continue
								}
								done = true
								break
							}
							if done {
								break
							}
							time.Sleep(200 * time.Millisecond)
						}
					}
				}

			case <-w.quitChan:
				workerLogger.Debugf("Tribe plugin worker-%d is stopping\n", w.id)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening
func (w worker) Stop() {
	go func() {
		w.quitChan <- true
	}()
}

func shuffle(m []Member) []Member {
	result := make([]Member, len(m))
	perm := rand.Perm(len(m))
	for i, v := range perm {
		result[v] = m[i]
	}
	return result
}

func (w worker) isPluginLoaded(n, t string, v int) bool {
	catalog := w.pluginManager.PluginCatalog()
	for _, item := range catalog {
		if item.TypeName() == t &&
			item.Name() == n &&
			item.Version() == v {
			workerLogger.WithField("_block", "isPluginLoaded").Info("Plugin already loaded")
			return true
		}
	}
	return false
}

func getSchedule(s *request.Schedule) schedule.Schedule {
	switch s.Type {
	case "simple":
		d, e := time.ParseDuration(s.Interval)
		if e != nil {
			workerLogger.WithField("_block", "getSchedule").Error(e)
			return nil
		}
		return schedule.NewSimpleSchedule(d)
	}
	return nil
}
