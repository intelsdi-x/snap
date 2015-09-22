package tribe

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	pluginLoadedType = iota
	pluginAddedToAgreementType
)

var workerLogger = log.WithFields(log.Fields{
	"_module": "tribe-worker",
})

type pluginRequest struct {
	Plugin      plugin
	requestType int
}

// newPluginWorker
func newPluginWorker(id int, workerQueue chan chan pluginRequest, quitChan chan interface{}, wg *sync.WaitGroup, pm managesPlugins, mm getsMembers) pluginWorker {
	// Create, and return the worker.
	worker := pluginWorker{
		pluginManager: pm,
		memberManager: mm,
		id:            id,
		work:          make(chan pluginRequest),
		workerQueue:   workerQueue,
		quitChan:      quitChan,
		waitGroup:     wg,
	}

	return worker
}

type pluginWorker struct {
	pluginManager managesPlugins
	memberManager getsMembers
	id            int
	work          chan pluginRequest
	workerQueue   chan chan pluginRequest
	quitChan      chan interface{}
	waitGroup     *sync.WaitGroup
}

type getsMembers interface {
	getPluginAgreementMembers() ([]*member, error)
}

// Start "starts" the worker
func (w pluginWorker) Start() {
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		for {
			// Add ourselves into the worker queue.
			w.workerQueue <- w.work

			var done bool
			select {
			case work := <-w.work:
				// Receive a work request.
				wlogger := workerLogger.WithFields(log.Fields{
					"plugin_name":    work.Plugin.Name_,
					"plugin_version": work.Plugin.Version_,
					"plugin_type":    work.Plugin.Type_.String(),
					"worker":         w.id,
				})
				workerLogger.Debug("received work")
				done = false
				for {
					if w.isPluginLoaded(work.Plugin.Name_, work.Plugin.Type_.String(), work.Plugin.Version_) {
						break
					}
					members, err := w.memberManager.getPluginAgreementMembers()
					if err != nil {
						wlogger.Error(err)
						continue
					}
					for _, member := range shuffle(members) {
						url := fmt.Sprintf("http://%s:%s/v1/plugins/%s/%s/%d", member.Node.Addr, member.Tags[RestAPIPort], work.Plugin.Type_.String(), work.Plugin.Name_, work.Plugin.Version_)
						workerLogger.Debugf("worker-%v is trying %v ", w.id, url)
						resp, err := http.Get(url)
						if err != nil {
							wlogger.Error(err)
							continue
						}
						if resp.StatusCode == 200 {
							if resp.Header.Get("Content-Type") != "application/x-gzip" {
								wlogger.WithField("content-type", resp.Header.Get("Content-Type")).Error("Expected application/x-gzip")
							}
							f, err := ioutil.TempFile("", fmt.Sprintf("%s-%s-%d", work.Plugin.Type_.String(), work.Plugin.Name_, work.Plugin.Version_))
							if err != nil {
								wlogger.Error(err)
								continue
							}
							io.Copy(f, resp.Body)
							f.Close()
							err = os.Chmod(f.Name(), 0700)
							if err != nil {
								wlogger.Error(err)
								continue
							}
							_, err = w.pluginManager.Load(f.Name())
							if err != nil {
								wlogger.Error(err)
								continue
							}
							if w.isPluginLoaded(work.Plugin.Name_, work.Plugin.Type_.String(), work.Plugin.Version_) {
								wlogger.WithField("path", f.Name()).Info("loaded plugin")
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

			case <-w.quitChan:
				workerLogger.Debugf("Tribe plugin worker-%d is stopping\n", w.id)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening
func (w pluginWorker) Stop() {
	go func() {
		w.quitChan <- true
	}()
}

func shuffle(m []*member) []*member {
	result := make([]*member, len(m))
	perm := rand.Perm(len(m))
	for i, v := range perm {
		result[v] = m[i]
	}
	return result
}

func (w pluginWorker) isPluginLoaded(n, t string, v int) bool {
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
