package tribe

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
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
func newPluginWorker(id int, workerQueue chan chan pluginRequest, pm managesPlugins, mm getsMembers) pluginWorker {
	// Create, and return the worker.
	worker := pluginWorker{
		pluginManager: pm,
		memberManager: mm,
		ID:            id,
		Work:          make(chan pluginRequest),
		WorkerQueue:   workerQueue,
		QuitChan:      make(chan bool)}

	return worker
}

type pluginWorker struct {
	pluginManager managesPlugins
	memberManager getsMembers
	ID            int
	Work          chan pluginRequest
	WorkerQueue   chan chan pluginRequest
	QuitChan      chan bool
}

type getsMembers interface {
	getPluginAgreementMembers() ([]*member, error)
}

// Start "starts" the worker
func (w pluginWorker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			var done bool
			select {
			case work := <-w.Work:
				// Receive a work request.
				wlogger := workerLogger.WithFields(log.Fields{
					"plugin_name":    work.Plugin.Name_,
					"plugin_version": work.Plugin.Version_,
					"plugin_type":    work.Plugin.Type_.String(),
					"worker":         w.ID,
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
						workerLogger.Debugf("worker-%v is trying %v ", w.ID, url)
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

			case <-w.QuitChan:
				workerLogger.Debugf("Tribe plugin worker-%d is stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening
func (w pluginWorker) Stop() {
	go func() {
		w.QuitChan <- true
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
