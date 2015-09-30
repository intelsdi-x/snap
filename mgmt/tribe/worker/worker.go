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
	ID            string
	StartOnCreate bool
}

type ManagesPlugins interface {
	Load(path string) (core.CatalogedPlugin, perror.PulseError)
	Unload(plugin core.Plugin) (core.CatalogedPlugin, perror.PulseError)
	PluginCatalog() core.PluginCatalog
}

type ManagesTasks interface {
	GetTask(id string) (core.Task, error)
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
		quitChan:          make(chan bool),
	}

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
				workerLogger.Infof("Received plugin work request")
				go func() {
					pluginWorker := <-pluginWorkerQueue

					workerLogger.Infof("Dispatching plugin work request")
					pluginWorker <- pluginWork
				}()
			case taskWork := <-taskWorkQueue:
				workerLogger.Infof("Received task work request")
				go func() {
					workerLogger.Infof("Waiting for free worker")
					taskWorker := <-taskWorkerQueue

					workerLogger.Infof("Dispatching task work request")
					taskWorker <- taskWork
				}()
			case <-quitChan:
				workerLogger.Infof("Stopping plugin work dispatcher")
				return
			}
		}
	}()
}

// Start "starts" the workers
func (w worker) start() {
	// task worker
	go func() {
		for {
			defer w.waitGroup.Done()
			w.taskWorkerQueue <- w.taskWork

			select {
			case work := <-w.taskWork:
				done := false
				// Receive a work request.
				wLogger := workerLogger.WithFields(log.Fields{
					"task":   work.Task.ID,
					"worker": w.id,
					"_block": "start",
				})
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
								uri := fmt.Sprintf("http://%s:%s", member.GetAddr(), member.GetRESTAPIPort())
								wLogger.Debugf("getting task %v from %v", work.Task.ID, uri)
								c := client.New(uri, "v1")
								taskResult := c.GetTask(work.Task.ID)
								if taskResult.Err != nil {
									wLogger.WithField("err", taskResult.Err.Error()).Debug("error getting task")
									continue
								}
								wLogger.Debug("creating task")
								opt := core.SetTaskID(work.Task.ID)
								t, errs := w.taskManager.CreateTask(
									getSchedule(taskResult.ScheduledTaskReturned.Schedule),
									taskResult.Workflow,
									work.Task.StartOnCreate,
									opt)
								if errs != nil && len(errs.Errors()) > 0 {
									fields := log.Fields{
										"task":   work.Task.ID,
										"worker": w.id,
										"_block": "start",
									}
									for idx, e := range errs.Errors() {
										fields[fmt.Sprintf("err-%d", idx)] = e
									}
									wLogger.WithFields(fields).Debug("error creating task")
									continue
								}
								wLogger.WithField("id", t.ID()).Debugf("task created")
								done = true
								break
							}
							if done {
								break
							}
							time.Sleep(500 * time.Millisecond)
						}
					}
				}

			case <-w.quitChan:
				workerLogger.Infof("Tribe plugin worker-%d is stopping\n", w.id)
				return

			}
		}
	}()

	// plugin worker
	go func() {
		defer w.waitGroup.Done()
		for {
			// Add ourselves into the worker queue.
			w.pluginWorkerQueue <- w.pluginWork

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
								f.Close()
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
								done = true
								break
							}
						}
					}
					if done {
						break
					}
					time.Sleep(500 * time.Millisecond)
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
		workerLogger.WithFields(log.Fields{
			"name":    n,
			"version": v,
			"type":    t,
		}).Errorf("loaded plugin.. looking for %v %v %v", item.Name(), item.Version(), item.TypeName())
		if item.TypeName() == t &&
			item.Name() == n &&
			item.Version() == v {
			workerLogger.WithField("_block", "isPluginLoaded").Error("Plugin already loaded")
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
