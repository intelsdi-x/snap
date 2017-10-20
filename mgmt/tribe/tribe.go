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

package tribe

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/core/tribe_event"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	"github.com/intelsdi-x/snap/mgmt/tribe/worker"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
)

const (
	HandlerRegistrationName = "tribe"
)

var (
	errAgreementDoesNotExist          = errors.New("Agreement does not exist")
	errAgreementAlreadyExists         = errors.New("Agreement already exists")
	errUnknownMember                  = errors.New("Unknown member")
	errAlreadyMemberOfPluginAgreement = errors.New("Already a member of a plugin agreement")
	errNotAMember                     = errors.New("Not a member of agreement")
	errTaskAlreadyExists              = errors.New("Task already exists")
	errTaskDoesNotExist               = errors.New("Task does not exist")
	errCreateMemberlist               = errors.New("Failed to start tribe")
	errMemberlistJoin                 = errors.New("Failed to join tribe")
	errPluginCatalogNotSet            = errors.New("Plugin Catalog not set")
	errTaskManagerNotSet              = errors.New("Task Manager not set")
)

var logger = log.WithFields(log.Fields{
	"_module": "tribe",
})

type tribe struct {
	clock              LClock
	agreements         map[string]*agreement.Agreement
	mutex              sync.RWMutex
	msgBuffer          []msg
	intentBuffer       []msg
	broadcasts         *memberlist.TransmitLimitedQueue
	memberlist         *memberlist.Memberlist
	logger             *log.Entry
	taskStartStopCache *cache
	taskStateResponses map[string]*taskStateQueryResponse
	members            map[string]*agreement.Member
	tags               map[string]string
	EventManager       *gomit.EventController
	config             *Config

	pluginCatalog   worker.ManagesPlugins
	taskManager     worker.ManagesTasks
	pluginWorkQueue chan worker.PluginRequest
	taskWorkQueue   chan worker.TaskRequest

	workerQuitChan  chan struct{}
	workerWaitGroup *sync.WaitGroup
}

func New(cfg *Config) (*tribe, error) {
	cfg.MemberlistConfig.Name = cfg.Name
	cfg.MemberlistConfig.BindAddr = cfg.BindAddr
	cfg.MemberlistConfig.BindPort = cfg.BindPort
	logger := logger.WithFields(log.Fields{
		"port": cfg.MemberlistConfig.BindPort,
		"addr": cfg.MemberlistConfig.BindAddr,
		"name": cfg.MemberlistConfig.Name,
	})

	tribe := &tribe{
		agreements:         map[string]*agreement.Agreement{},
		members:            map[string]*agreement.Member{},
		taskStateResponses: map[string]*taskStateQueryResponse{},
		taskStartStopCache: newCache(),
		msgBuffer:          make([]msg, 512),
		intentBuffer:       []msg{},
		logger:             logger.WithField("_name", cfg.MemberlistConfig.Name),
		tags: map[string]string{
			agreement.RestPort:               strconv.Itoa(cfg.RestAPIPort),
			agreement.RestProtocol:           cfg.RestAPIProto,
			agreement.RestInsecureSkipVerify: cfg.RestAPIInsecureSkipVerify,
		},
		pluginWorkQueue: make(chan worker.PluginRequest, 999),
		taskWorkQueue:   make(chan worker.TaskRequest, 999),
		workerQuitChan:  make(chan struct{}),
		workerWaitGroup: &sync.WaitGroup{},
		config:          cfg,
		EventManager:    gomit.NewEventController(),
	}

	tribe.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return len(tribe.memberlist.Members())
		},
		RetransmitMult: memberlist.DefaultLANConfig().RetransmitMult,
	}

	//configure delegates
	cfg.MemberlistConfig.Delegate = &delegate{tribe: tribe}
	cfg.MemberlistConfig.Events = &memberDelegate{tribe: tribe}

	ml, err := memberlist.Create(cfg.MemberlistConfig)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	tribe.memberlist = ml

	if cfg.Seed != "" {
		_, err := ml.Join([]string{cfg.Seed})
		if err != nil {
			logger.WithFields(log.Fields{
				"seed": cfg.Seed,
			}).Error(errMemberlistJoin)
			return nil, errMemberlistJoin
		}
		logger.WithFields(log.Fields{
			"seed": cfg.Seed,
		}).Infoln("tribe started")
		return tribe, nil
	}
	logger.WithFields(log.Fields{
		"seed": "none",
	}).Infoln("tribe started")
	return tribe, nil
}

type cache struct {
	sync.RWMutex
	table map[string]time.Time
}

func newCache() *cache {
	return &cache{
		table: make(map[string]time.Time),
	}
}

func (c *cache) get(m msg) (time.Time, bool) {
	c.RLock()
	defer c.RUnlock()
	key := fmt.Sprintf("%v:%v", m.GetType(), m.ID())
	v, ok := c.table[key]
	return v, ok
}

func (c *cache) put(m msg, duration time.Duration) bool {
	c.Lock()
	defer c.Unlock()
	key := fmt.Sprintf("%v:%v", m.GetType(), m.ID())
	if ti, ok := c.table[key]; ok {
		logger.WithFields(log.Fields{
			"task-id":      m.ID(),
			"time-started": ti,
			"_block":       "task-cache-put",
			"key":          key,
			"cache-size":   len(c.table),
			"ltime":        m.Time(),
		}).Debugln("task cache entry exists")
		return false
	}
	c.table[key] = time.Now()
	time.AfterFunc(duration, func() {
		c.Lock()
		delete(c.table, key)
		c.Unlock()
	})
	return true
}

func (t *tribe) SetPluginCatalog(p worker.ManagesPlugins) {
	t.pluginCatalog = p
}

func (t *tribe) SetTaskManager(m worker.ManagesTasks) {
	t.taskManager = m
}

func (t *tribe) Name() string {
	return "tribe"
}

func (t *tribe) Start() error {
	if t.pluginCatalog == nil {
		return errPluginCatalogNotSet
	}
	if t.taskManager == nil {
		return errTaskManagerNotSet
	}
	worker.DispatchWorkers(
		4,
		t.pluginWorkQueue,
		t.taskWorkQueue,
		t.workerQuitChan,
		t.workerWaitGroup,
		t.pluginCatalog,
		t.taskManager,
		t)
	return nil
}

func (t *tribe) Stop() {
	logger := t.logger.WithFields(log.Fields{
		"_block": "stop",
	})
	err := t.memberlist.Leave(1 * time.Second)
	if err != nil {
		logger.Error(err)
	}
	err = t.memberlist.Shutdown()
	if err != nil {
		logger.Error(err)
	}
	close(t.workerQuitChan)
	t.workerWaitGroup.Wait()
}

func (t *tribe) GetTaskAgreementMembers() ([]worker.Member, error) {
	m, ok := t.members[t.memberlist.LocalNode().Name]
	if !ok || m.TaskAgreements == nil {
		return nil, errNotAMember
	}

	mm := map[*agreement.Member]struct{}{}
	for name := range m.TaskAgreements {
		for _, mem := range t.agreements[name].Members {
			mm[mem] = struct{}{}
		}
	}
	members := make([]worker.Member, 0, len(mm))
	for k := range mm {
		members = append(members, k)
	}
	return members, nil
}

func (t *tribe) GetPluginAgreementMembers() ([]worker.Member, error) {
	m, ok := t.members[t.memberlist.LocalNode().Name]
	if !ok || m.PluginAgreement == nil {
		return nil, errNotAMember
	}
	members := make([]worker.Member, 0, len(t.agreements[m.PluginAgreement.Name].Members))
	for _, v := range t.agreements[m.PluginAgreement.Name].Members {
		members = append(members, v)
	}
	return members, nil
}

// encodeTags
func (t *tribe) encodeTags(tags map[string]string) []byte {
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := enc.Encode(tags); err != nil {
		panic(fmt.Sprintf("Failed to encode tags: %v", err))
	}
	return buf.Bytes()
}

// decodeTags is used to decode a tag map
func (t *tribe) decodeTags(buf []byte) map[string]string {
	tags := make(map[string]string)
	r := bytes.NewReader(buf)
	dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
	if err := dec.Decode(&tags); err != nil {
		t.logger.WithFields(log.Fields{
			"_block": "decode-tags",
			"error":  err,
		}).Error("Failed to decode tags")
	}
	return tags
}

// HandleGomitEvent handles events emitted from control
func (t *tribe) HandleGomitEvent(e gomit.Event) {
	logger := t.logger.WithFields(log.Fields{
		"_block": "handle-gomit-event",
	})
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		logger.WithFields(log.Fields{
			"event":          e.Namespace(),
			"plugin-name":    v.Name,
			"plugin-version": v.Version,
			"plugin-type":    core.PluginType(v.Type).String(),
		}).Debugf("handling load plugin event")
		plugin := agreement.Plugin{
			Name_:    v.Name,
			Version_: v.Version,
			Type_:    core.PluginType(v.Type),
		}
		if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
			if m.PluginAgreement != nil {
				if ok, _ := m.PluginAgreement.Plugins.Contains(plugin); !ok {
					t.AddPlugin(m.PluginAgreement.Name, plugin)
				}
			}
		}
	case *control_event.UnloadPluginEvent:
		logger.WithFields(log.Fields{
			"event":          e.Namespace(),
			"plugin-name":    v.Name,
			"plugin-version": v.Version,
			"plugin-type":    core.PluginType(v.Type).String(),
		}).Debugf("handling unload plugin event")
		plugin := agreement.Plugin{
			Name_:    v.Name,
			Version_: v.Version,
			Type_:    core.PluginType(v.Type),
		}
		if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
			if m.PluginAgreement != nil {
				if ok, _ := m.PluginAgreement.Plugins.Contains(plugin); ok {
					t.RemovePlugin(m.PluginAgreement.Name, plugin)
				}
			}
		}
	case *scheduler_event.TaskCreatedEvent:
		if v.Source != "tribe" {
			logger.WithFields(log.Fields{
				"event":                e.Namespace(),
				"task-id":              v.TaskID,
				"task-start-on-create": v.StartOnCreate,
			}).Debugf("handling task create event")
			task := agreement.Task{
				ID:            v.TaskID,
				StartOnCreate: v.StartOnCreate,
			}
			if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
				if m.TaskAgreements != nil {
					for n, a := range m.TaskAgreements {
						if ok, _ := a.Tasks.Contains(task); !ok {
							t.AddTask(n, task)
						}
					}
				}
			}
		}
	case *scheduler_event.TaskStoppedEvent:
		if v.Source != "tribe" {
			logger.WithFields(log.Fields{
				"event":   e.Namespace(),
				"task-id": v.TaskID,
			}).Debugf("handling task stop event")
			task := agreement.Task{
				ID: v.TaskID,
			}
			if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
				if m.TaskAgreements != nil {
					for n, a := range m.TaskAgreements {
						if ok, _ := a.Tasks.Contains(task); ok {
							t.StopTask(n, task)
						}
					}
				}
			}
		}
	case *scheduler_event.TaskStartedEvent:
		if v.Source != "tribe" {
			logger.WithFields(log.Fields{
				"event":   e.Namespace(),
				"task-id": v.TaskID,
			}).Debugf("handling task start event")
			task := agreement.Task{
				ID: v.TaskID,
			}
			if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
				if m.TaskAgreements != nil {
					for n, a := range m.TaskAgreements {
						if ok, _ := a.Tasks.Contains(task); ok {
							t.StartTask(n, task)
						}
					}
				}
			}
		}
	case *scheduler_event.TaskDeletedEvent:
		if v.Source != "tribe" {
			logger.WithFields(log.Fields{
				"event":   e.Namespace(),
				"task-id": v.TaskID,
			}).Debugf("handling task start event")
			task := agreement.Task{
				ID: v.TaskID,
			}
			if m, ok := t.members[t.memberlist.LocalNode().Name]; ok {
				if m.TaskAgreements != nil {
					for n, a := range m.TaskAgreements {
						if ok, _ := a.Tasks.Contains(task); ok {
							t.RemoveTask(n, task)
						}
					}
				}
			}
		}
	}
}

// broadcast takes a tribe message type, encodes it for the wire, and queues
// the broadcast. If a notify channel is given, this channel will be closed
// when the broadcast is sent.
func (t *tribe) broadcast(mt msgType, msg interface{}, notify chan<- struct{}) error {
	raw, err := encodeMessage(mt, msg)
	if err != nil {
		return err
	}

	t.broadcasts.QueueBroadcast(&broadcast{
		msg:    raw,
		notify: notify,
	})
	return nil
}

func (t *tribe) GetMember(name string) *agreement.Member {
	if m, ok := t.members[name]; ok {
		return m
	}
	return nil
}

func (t *tribe) GetMembers() []string {
	var members []string
	for _, member := range t.memberlist.Members() {
		members = append(members, member.Name)
	}
	return members
}

func (t *tribe) LeaveAgreement(agreementName, memberName string) serror.SnapError {
	if err := t.canLeaveAgreement(agreementName, memberName); err != nil {
		return err
	}

	msg := &agreementMsg{
		LTime:         t.clock.Increment(),
		UUID:          uuid.New(),
		AgreementName: agreementName,
		MemberName:    memberName,
		Type:          leaveAgreementMsgType,
	}
	if t.handleLeaveAgreement(msg) {
		t.broadcast(leaveAgreementMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) JoinAgreement(agreementName, memberName string) serror.SnapError {
	if err := t.canJoinAgreement(agreementName, memberName); err != nil {
		return err
	}

	msg := &agreementMsg{
		LTime:         t.clock.Increment(),
		UUID:          uuid.New(),
		AgreementName: agreementName,
		MemberName:    memberName,
		Type:          joinAgreementMsgType,
	}
	if t.handleJoinAgreement(msg) {
		t.broadcast(joinAgreementMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) AddPlugin(agreementName string, p agreement.Plugin) error {
	if _, ok := t.agreements[agreementName]; !ok {
		return errAgreementDoesNotExist
	}
	msg := &pluginMsg{
		LTime:         t.clock.Increment(),
		Plugin:        p,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          addPluginMsgType,
	}
	defer t.EventManager.Emit(&tribe_event.AddPluginEvent{
		Agreement: struct{ Name string }{agreementName},
		Plugin: struct {
			Name    string
			Type    core.PluginType
			Version int
		}{Name: p.Name(), Type: p.Type_, Version: p.Version_},
	})
	if t.handleAddPlugin(msg) {
		t.broadcast(addPluginMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) RemovePlugin(agreementName string, p agreement.Plugin) error {
	if _, ok := t.agreements[agreementName]; !ok {
		return errAgreementDoesNotExist
	}
	msg := &pluginMsg{
		LTime:         t.clock.Increment(),
		Plugin:        p,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          removePluginMsgType,
	}
	if t.handleRemovePlugin(msg) {
		t.broadcast(removePluginMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) GetAgreement(name string) (*agreement.Agreement, serror.SnapError) {
	a, ok := t.agreements[name]
	if !ok {
		return nil, serror.New(errAgreementDoesNotExist, map[string]interface{}{"agreement_name": name})
	}
	return a, nil
}

func (t *tribe) GetAgreements() map[string]*agreement.Agreement {
	return t.agreements
}

func (t *tribe) AddTask(agreementName string, task agreement.Task) serror.SnapError {
	if err := t.canAddTask(task, agreementName); err != nil {
		return err
	}
	msg := &taskMsg{
		LTime:         t.clock.Increment(),
		TaskID:        task.ID,
		StartOnCreate: task.StartOnCreate,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          addTaskMsgType,
	}
	if t.handleAddTask(msg) {
		t.broadcast(addTaskMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) RemoveTask(agreementName string, task agreement.Task) serror.SnapError {
	if err := t.canStartStopRemoveTask(task, agreementName); err != nil {
		return err
	}
	msg := &taskMsg{
		LTime:         t.clock.Increment(),
		TaskID:        task.ID,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          removeTaskMsgType,
	}
	if t.handleRemoveTask(msg) {
		t.broadcast(removeTaskMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) StopTask(agreementName string, task agreement.Task) serror.SnapError {
	if err := t.canStartStopRemoveTask(task, agreementName); err != nil {
		return err
	}
	msg := &taskMsg{
		LTime:         t.clock.Increment(),
		TaskID:        task.ID,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          stopTaskMsgType,
	}
	if t.handleStopTask(msg) {
		t.broadcast(stopTaskMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) StartTask(agreementName string, task agreement.Task) serror.SnapError {
	if err := t.canStartStopRemoveTask(task, agreementName); err != nil {
		return err
	}

	msg := &taskMsg{
		LTime:         t.clock.Increment(),
		TaskID:        task.ID,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          startTaskMsgType,
	}
	if t.handleStartTask(msg) {
		t.broadcast(startTaskMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) AddAgreement(name string) serror.SnapError {
	if _, ok := t.agreements[name]; ok {
		fields := log.Fields{
			"agreement": name,
		}
		return serror.New(errAgreementAlreadyExists, fields)
	}
	msg := &agreementMsg{
		LTime:         t.clock.Increment(),
		AgreementName: name,
		UUID:          uuid.New(),
		Type:          addAgreementMsgType,
	}
	if t.handleAddAgreement(msg) {
		t.broadcast(addAgreementMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) RemoveAgreement(name string) serror.SnapError {
	if _, ok := t.agreements[name]; !ok {
		fields := log.Fields{
			"Agreement": name,
		}
		return serror.New(errAgreementDoesNotExist, fields)
	}
	msg := &agreementMsg{
		LTime:         t.clock.Increment(),
		AgreementName: name,
		UUID:          uuid.New(),
		Type:          removeAgreementMsgType,
	}
	if t.handleRemoveAgreement(msg) {
		t.broadcast(removeAgreementMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) TaskStateQuery(agreementName string, taskId string) core.TaskState {
	resp := t.taskStateQuery(agreementName, taskId)

	responses := taskStateResponses{}
	for r := range resp.resp {
		responses = append(responses, r)
	}

	return responses.State()
}

func (t *tribe) taskStateQuery(agreementName string, taskId string) *taskStateQueryResponse {
	timeout := t.getTimeout()
	msg := &taskStateQueryMsg{
		LTime:         t.clock.Increment(),
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          getTaskStateMsgType,
		Addr:          t.memberlist.LocalNode().Addr,
		Port:          t.memberlist.LocalNode().Port,
		TaskID:        taskId,
		Deadline:      time.Now().Add(timeout),
	}

	resp := newStateQueryResponse(len(t.memberlist.Members()), msg)
	t.registerQueryResponse(timeout, resp)
	t.broadcast(msg.Type, msg, nil)

	return resp
}

func (t *tribe) processIntents() {
	for {
		if t.processAddPluginIntents() &&
			t.processRemovePluginIntents() &&
			t.processAddAgreementIntents() &&
			t.processRemoveAgreementIntents() &&
			t.processJoinAgreementIntents() &&
			t.processLeaveAgreementIntents() &&
			t.processAddTaskIntents() &&
			t.processRemoveTaskIntents() {
			return
		}
	}
}

func (t *tribe) processAddPluginIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == addPluginMsgType {
			intent := v.(*pluginMsg)
			if _, ok := t.agreements[intent.AgreementName]; ok {
				if ok, _ := t.agreements[intent.AgreementName].PluginAgreement.Plugins.Contains(intent.Plugin); !ok {
					t.agreements[intent.AgreementName].PluginAgreement.Plugins = append(t.agreements[intent.AgreementName].PluginAgreement.Plugins, intent.Plugin)
					t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)

					ptype, _ := core.ToPluginType(intent.Plugin.TypeName())
					work := worker.PluginRequest{
						Plugin: agreement.Plugin{
							Name_:    intent.Plugin.Name(),
							Version_: intent.Plugin.Version(),
							Type_:    ptype,
						},
						RequestType: worker.PluginLoadedType,
					}
					t.pluginWorkQueue <- work

					return false
				}
			}
		}
	}
	return true
}

func (t *tribe) processRemovePluginIntents() bool {
	for k, v := range t.intentBuffer {
		if v.GetType() == removePluginMsgType {
			intent := v.(*pluginMsg)
			if a, ok := t.agreements[intent.AgreementName]; ok {
				if ok, idx := a.PluginAgreement.Plugins.Contains(intent.Plugin); ok {
					a.PluginAgreement.Plugins = append(a.PluginAgreement.Plugins[:idx], a.PluginAgreement.Plugins[idx+1:]...)
					t.intentBuffer = append(t.intentBuffer[:k], t.intentBuffer[k+1:]...)
					return false
				}
			}
		}
	}
	return true
}

func (t *tribe) processAddTaskIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == addTaskMsgType {
			intent := v.(*taskMsg)
			if a, ok := t.agreements[intent.AgreementName]; ok {
				if ok, _ := a.TaskAgreement.Tasks.Contains(agreement.Task{ID: intent.TaskID}); !ok {
					a.TaskAgreement.Tasks = append(a.TaskAgreement.Tasks, agreement.Task{ID: intent.TaskID})
					t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)

					work := worker.TaskRequest{
						Task: worker.Task{
							ID:            intent.TaskID,
							StartOnCreate: intent.StartOnCreate,
						},
						RequestType: worker.TaskCreatedType,
					}
					t.taskWorkQueue <- work

					return false
				}
			}
		}
	}
	return true
}

func (t *tribe) processRemoveTaskIntents() bool {
	for k, v := range t.intentBuffer {
		if v.GetType() == removeTaskMsgType {
			intent := v.(*taskMsg)
			if _, ok := t.agreements[intent.AgreementName]; ok {
				if ok, idx := t.agreements[intent.AgreementName].TaskAgreement.Tasks.Contains(agreement.Task{ID: intent.TaskID}); ok {
					t.agreements[intent.AgreementName].TaskAgreement.Tasks = append(t.agreements[intent.AgreementName].TaskAgreement.Tasks[:idx], t.agreements[intent.AgreementName].TaskAgreement.Tasks[idx+1:]...)
					t.intentBuffer = append(t.intentBuffer[:k], t.intentBuffer[k+1:]...)
					return false
				}
			}
		}
	}
	return true
}

func (t *tribe) processAddAgreementIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == addAgreementMsgType {
			intent := v.(*agreementMsg)
			if _, ok := t.agreements[intent.AgreementName]; !ok {
				t.agreements[intent.AgreementName] = agreement.New(intent.AgreementName)
				t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)
				return false
			}
		}
	}
	return true
}

func (t *tribe) processRemoveAgreementIntents() bool {
	for k, v := range t.intentBuffer {
		if v.GetType() == removeAgreementMsgType {
			intent := v.(*agreementMsg)
			if _, ok := t.agreements[intent.Agreement()]; ok {
				delete(t.agreements, intent.Agreement())
				t.intentBuffer = append(t.intentBuffer[:k], t.intentBuffer[k+1:]...)
				return false
			}
		}
	}
	return true
}

func (t *tribe) processJoinAgreementIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == joinAgreementMsgType {
			intent := v.(*agreementMsg)
			if _, ok := t.members[intent.MemberName]; ok {
				if _, ok := t.agreements[intent.AgreementName]; ok {
					err := t.joinAgreement(intent)
					if err == nil {
						t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)
					}
					return false
				}
			}
		}
	}
	return true
}

func (t *tribe) processLeaveAgreementIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == joinAgreementMsgType {
			intent := v.(*agreementMsg)
			if _, ok := t.members[intent.MemberName]; ok {
				if _, ok := t.agreements[intent.AgreementName]; ok {
					if _, ok := t.agreements[intent.AgreementName].Members[intent.MemberName]; ok {
						err := t.leaveAgreement(intent)
						if err == nil {
							t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)
						}
						return false
					}
				}
			}
		}
	}
	return true
}

func (t *tribe) handleRemovePlugin(msg *pluginMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.Agreement()]; ok {
		if t.agreements[msg.AgreementName].PluginAgreement.Remove(msg.Plugin) {
			t.processIntents()
			if t.pluginCatalog != nil {
				_, err := t.pluginCatalog.Unload(msg.Plugin)
				if err != nil {
					t.logger.WithFields(log.Fields{
						"_block":         "handle-remove-plugin",
						"plugin-name":    msg.Plugin.Name(),
						"plugin-type":    msg.Plugin.TypeName(),
						"plugin-version": msg.Plugin.Version(),
					}).Error(err)
				}
			}
			return true
		}
	}

	t.addPluginIntent(msg)
	return true
}

func (t *tribe) handleAddPlugin(msg *pluginMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.AgreementName]; ok {
		if t.agreements[msg.AgreementName].PluginAgreement.Add(msg.Plugin) {

			ptype, _ := core.ToPluginType(msg.Plugin.TypeName())
			work := worker.PluginRequest{
				Plugin: agreement.Plugin{
					Name_:    msg.Plugin.Name(),
					Version_: msg.Plugin.Version(),
					Type_:    ptype,
				},
				RequestType: worker.PluginLoadedType,
			}
			t.pluginWorkQueue <- work

			t.processIntents()
			return true
		}
	}

	t.addPluginIntent(msg)
	return true
}

func (t *tribe) handleAddTask(msg *taskMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.AgreementName]; ok {
		if t.agreements[msg.AgreementName].TaskAgreement.Add(agreement.Task{ID: msg.TaskID}) {

			work := worker.TaskRequest{
				Task: worker.Task{
					ID:            msg.TaskID,
					StartOnCreate: msg.StartOnCreate,
				},
				RequestType: worker.TaskCreatedType,
			}
			t.taskWorkQueue <- work

			t.processIntents()
			return true
		}
	}

	t.addTaskIntent(msg)
	return true
}

func (t *tribe) handleRemoveTask(msg *taskMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.Agreement()]; ok {
		if t.agreements[msg.AgreementName].TaskAgreement.Remove(agreement.Task{ID: msg.TaskID}) {

			work := worker.TaskRequest{
				Task: worker.Task{
					ID: msg.TaskID,
				},
				RequestType: worker.TaskRemovedType,
			}
			t.taskWorkQueue <- work

			t.processIntents()
			return true
		}
	}

	t.addTaskIntent(msg)
	return true
}

func (t *tribe) handleStartTask(msg *taskMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.Agreement()]; ok {

		if ok := t.taskStartStopCache.put(msg, t.getTimeout()); !ok {
			// A cache entry exists; return and do not broadcast event again
			return false
		}

		work := worker.TaskRequest{
			Task: worker.Task{
				ID: msg.TaskID,
			},
			RequestType: worker.TaskStartedType,
		}
		t.taskWorkQueue <- work

		return true
	}

	return true
}

func (t *tribe) handleStopTask(msg *taskMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.Agreement()]; ok {

		if ok := t.taskStartStopCache.put(msg, t.getTimeout()); !ok {
			// A cache entry exists; return and do not broadcast event again
			return false
		}

		work := worker.TaskRequest{
			Task: worker.Task{
				ID: msg.TaskID,
			},
			RequestType: worker.TaskStoppedType,
		}
		t.taskWorkQueue <- work

		return true
	}

	return true
}

func (t *tribe) handleMemberJoin(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; !ok {
		t.members[n.Name] = agreement.NewMember(n)
		t.members[n.Name].Tags = t.decodeTags(n.Meta)
		t.members[n.Name].Tags["host"] = n.Addr.String()
	}
	t.processIntents()
}

func (t *tribe) handleMemberLeave(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if m, ok := t.members[n.Name]; ok {
		if m.PluginAgreement != nil {
			delete(t.agreements[m.PluginAgreement.Name].Members, n.Name)
		}
		for k := range m.TaskAgreements {
			delete(t.agreements[k].Members, n.Name)
		}
		delete(t.members, n.Name)
	}
}

func (t *tribe) handleMemberUpdate(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; ok {
		t.members[n.Name].Tags = t.decodeTags(n.Meta)
	}
}

func (t *tribe) handleAddAgreement(msg *agreementMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	// add msg to seen buffer
	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	// add agreement
	if _, ok := t.agreements[msg.AgreementName]; !ok {
		t.agreements[msg.AgreementName] = agreement.New(msg.AgreementName)
		t.processIntents()
		return true
	}
	t.addAgreementIntent(msg)
	return true
}

func (t *tribe) handleRemoveAgreement(msg *agreementMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	// add msg to seen buffer
	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.AgreementName]; ok {
		delete(t.agreements, msg.AgreementName)
		t.processIntents()
		// TODO consider removing any intents that involve this agreement
		return true
	}

	return true
}

func (t *tribe) handleJoinAgreement(msg *agreementMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if err := t.joinAgreement(msg); err == nil {
		t.processIntents()
		return true
	}

	t.addAgreementIntent(msg)
	return true
}

func (t *tribe) handleLeaveAgreement(msg *agreementMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if err := t.leaveAgreement(msg); err == nil {
		t.processIntents()
		return true
	}

	t.addAgreementIntent(msg)

	return true
}

func (t *tribe) handleTaskStateQuery(msg *taskStateQueryMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if time.Now().After(msg.Deadline) {
		t.logger.WithFields(log.Fields{
			"_block":   "handleStateQuery",
			"deadline": msg.Deadline,
			"ltime":    msg.LTime,
		}).Warn("deadline passed for task state query")
		return false
	}

	if !t.isMemberOfAgreement(msg.Agreement()) {
		// we are not a member of the agreement
		return true
	}

	resp := taskStateQueryResponseMsg{
		LTime: msg.LTime,
		UUID:  msg.UUID,
		From:  t.memberlist.LocalNode().Name,
	}

	tsk, err := t.taskManager.GetTask(msg.TaskID)
	if err != nil {
		t.logger.WithFields(log.Fields{
			"_block":  "handleStateQuery",
			"err":     err,
			"task-id": msg.TaskID,
			"msg-id":  msg.UUID,
		}).Error("failed to get task state")
		return true
	}

	resp.State = tsk.State()

	// Format the response
	raw, err := encodeMessage(taskStateQueryResponseMsgType, &resp)
	if err != nil {
		t.logger.WithFields(log.Fields{
			"_block": "handleStateQuery",
			"err":    err,
		}).Error("failed to encode message")
		return true
	}

	// Check the size limit
	if len(raw) > TaskStateQueryResponseSizeLimit {
		t.logger.WithFields(log.Fields{
			"_block":     "handleStateQuery",
			"err":        err,
			"size-limit": TaskStateQueryResponseSizeLimit,
			"msg-size":   len(raw),
		}).Error("msg exceeds size limit", TaskStateQueryResponseSizeLimit)
		return true
	}

	// Send the response
	addr := net.UDPAddr{IP: msg.Addr, Port: int(msg.Port)}
	if err := t.memberlist.SendTo(&addr, raw); err != nil {
		t.logger.WithFields(log.Fields{
			"_block":      "handleStateQuery",
			"remote-addr": msg.Addr,
			"remote-port": msg.Port,
			"err":         err,
		}).Error("failed to send task state reply")
	}

	return true
}

func (t *tribe) registerQueryResponse(timeout time.Duration, resp *taskStateQueryResponse) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, ok := t.taskStateResponses[resp.uuid]; ok {
		panic("not ok")
	}
	t.taskStateResponses[resp.uuid] = resp

	time.AfterFunc(timeout, func() {
		t.mutex.Lock()
		delete(t.taskStateResponses, resp.uuid)
		resp.Close()
		t.mutex.Unlock()
	})
}

func (t *tribe) joinAgreement(msg *agreementMsg) serror.SnapError {
	if err := t.canJoinAgreement(msg.Agreement(), msg.MemberName); err != nil {
		return err
	}
	// add plugin agreement to the member
	if t.agreements[msg.Agreement()].PluginAgreement != nil {
		t.members[msg.MemberName].PluginAgreement = t.agreements[msg.Agreement()].PluginAgreement
	}
	t.members[msg.MemberName].TaskAgreements[msg.Agreement()] = t.agreements[msg.Agreement()].TaskAgreement

	// update the agreements membership
	t.agreements[msg.Agreement()].Members[msg.MemberName] = t.members[msg.MemberName]

	// get plugins and tasks if this is the node joining
	if msg.MemberName == t.memberlist.LocalNode().Name {
		go func(a *agreement.Agreement) {
			for _, p := range a.PluginAgreement.Plugins {
				ptype, _ := core.ToPluginType(p.TypeName())
				work := worker.PluginRequest{
					Plugin: agreement.Plugin{
						Name_:    p.Name(),
						Version_: p.Version(),
						Type_:    ptype,
					},
					RequestType: worker.PluginLoadedType,
				}
				t.pluginWorkQueue <- work
			}

			for _, tsk := range a.TaskAgreement.Tasks {
				state := t.TaskStateQuery(msg.Agreement(), tsk.ID)
				startOnCreate := false
				if state == core.TaskSpinning || state == core.TaskFiring {
					startOnCreate = true
				}
				work := worker.TaskRequest{
					Task: worker.Task{
						ID:            tsk.ID,
						StartOnCreate: startOnCreate,
					},
					RequestType: worker.TaskCreatedType,
				}
				t.taskWorkQueue <- work
			}
		}(t.agreements[msg.Agreement()])
	}
	return nil
}

func (t *tribe) leaveAgreement(msg *agreementMsg) serror.SnapError {
	if err := t.canLeaveAgreement(msg.Agreement(), msg.MemberName); err != nil {
		return err
	}

	delete(t.agreements[msg.AgreementName].Members, msg.MemberName)
	t.members[msg.MemberName].PluginAgreement = nil
	if _, ok := t.members[msg.MemberName].TaskAgreements[msg.Agreement()]; ok {
		delete(t.members[msg.MemberName].TaskAgreements, msg.Agreement())
	}

	return nil
}

func (t *tribe) canLeaveAgreement(agreementName, memberName string) serror.SnapError {
	fields := log.Fields{
		"member-name": memberName,
		"agreement":   agreementName,
	}
	if _, ok := t.agreements[agreementName]; !ok {
		t.logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return serror.New(errAgreementDoesNotExist, fields)
	}
	m, ok := t.members[memberName]
	if !ok {
		t.logger.WithFields(fields).Debugln(errUnknownMember)
		return serror.New(errUnknownMember, fields)
	}
	if m.PluginAgreement == nil {
		t.logger.WithFields(fields).Debugln(errNotAMember)
		return serror.New(errNotAMember, fields)
	}
	return nil
}

func (t *tribe) canJoinAgreement(agreementName, memberName string) serror.SnapError {
	fields := log.Fields{
		"member-name": memberName,
		"agreement":   agreementName,
	}
	if _, ok := t.agreements[agreementName]; !ok {
		t.logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return serror.New(errAgreementDoesNotExist, fields)
	}
	m, ok := t.members[memberName]
	if !ok {
		t.logger.WithFields(fields).Debugln(errUnknownMember)
		return serror.New(errUnknownMember, fields)

	}
	if m.PluginAgreement != nil && len(m.PluginAgreement.Plugins) > 0 {
		// This log line creates an extremely large amount of logging
		// under debug. This was tested at 18GB for a 50 node tribe on
		// one node that had debug turned on.
		//
		// Uncomment this line if debugging tribe.
		// t.logger.WithFields(fields).Debugln(errAlreadyMemberOfPluginAgreement)
		return serror.New(errAlreadyMemberOfPluginAgreement, fields)
	}
	return nil
}

func (t *tribe) canAddTask(task agreement.Task, agreementName string) serror.SnapError {
	fields := log.Fields{
		"agreement": agreementName,
		"task-id":   task.ID,
	}
	a, ok := t.agreements[agreementName]
	if !ok {
		t.logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return serror.New(errAgreementDoesNotExist, fields)
	}
	if ok, _ := a.TaskAgreement.Tasks.Contains(task); ok {
		t.logger.WithFields(fields).Debugln(errTaskAlreadyExists)
		return serror.New(errTaskAlreadyExists, fields)
	}
	return nil
}

func (t *tribe) canStartStopRemoveTask(task agreement.Task, agreementName string) serror.SnapError {
	fields := log.Fields{
		"agreement": agreementName,
		"task-id":   task.ID,
	}
	a, ok := t.agreements[agreementName]
	if !ok {
		t.logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return serror.New(errAgreementDoesNotExist, fields)
	}
	if ok, _ := a.TaskAgreement.Tasks.Contains(task); !ok {
		t.logger.WithFields(fields).Debugln(errTaskDoesNotExist)
		return serror.New(errTaskDoesNotExist, fields)
	}
	return nil
}

func (t *tribe) isMemberOfAgreement(name string) bool {
	fields := log.Fields{
		"agreement": name,
		"_block":    "isMemberOfAgreement",
	}
	a, ok := t.agreements[name]
	if !ok {
		t.logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return false
	}
	if _, ok := a.Members[t.memberlist.LocalNode().Name]; !ok {
		t.logger.WithFields(fields).Debugln(errNotAMember)
		return false
	}
	return true
}

func (t *tribe) isDuplicate(msg msg) bool {
	logger := t.logger.WithFields(log.Fields{
		"event-clock": msg.Time(),
		"event":       msg.GetType().String(),
		"event-uuid":  msg.ID(),
		"clock":       t.clock.Time(),
		"agreement":   msg.Agreement(),
	})
	// is the message old
	if t.clock.Time() > LTime(len(t.msgBuffer)) &&
		msg.Time() < t.clock.Time()-LTime(len(t.msgBuffer)) {
		logger.Debugln("old message")
		return true
	}

	// have we seen it
	idx := msg.Time() % LTime(len(t.msgBuffer))
	seen := t.msgBuffer[idx]
	if seen != nil && seen.ID() == msg.ID() {
		logger.Debugln("duplicate message")
		return true
	}
	return false
}

func (t *tribe) addPluginIntent(msg *pluginMsg) bool {
	t.logger.WithFields(log.Fields{
		"event-clock": msg.LTime,
		"agreement":   msg.AgreementName,
		"type":        msg.Type.String(),
		"plugin": fmt.Sprintf("%v:%v:%v",
			msg.Plugin.TypeName(),
			msg.Plugin.Name(),
			msg.Plugin.Version()),
	}).Debugln("out of order message")
	t.intentBuffer = append(t.intentBuffer, msg)
	return true
}

func (t *tribe) addAgreementIntent(m msg) bool {
	t.logger.WithFields(log.Fields{
		"event-clock": m.Time(),
		"agreement":   m.Agreement(),
		"type":        m.GetType().String(),
	}).Debugln("out of order message")
	t.intentBuffer = append(t.intentBuffer, m)
	return true
}

func (t *tribe) addTaskIntent(m *taskMsg) bool {
	t.logger.WithFields(log.Fields{
		"event-clock": m.Time(),
		"agreement":   m.Agreement(),
		"type":        m.GetType().String(),
		"task-id":     m.TaskID,
	}).Debugln("Out of order msg")
	t.intentBuffer = append(t.intentBuffer, m)
	return true
}

func (t *tribe) getTimeout() time.Duration {
	// query duration - gossip interval * timeout mult * log(n+1)
	return time.Duration(t.config.MemberlistConfig.GossipInterval * 5 * time.Duration(math.Ceil(math.Log10(float64(len(t.memberlist.Members())+1)))))
}

func (t *tribe) GetRequestPassword() string {
	return t.config.RestAPIPassword
}
