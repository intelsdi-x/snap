package tribe

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/control_event"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/core/scheduler_event"
	"github.com/intelsdi-x/pulse/mgmt/tribe/agreement"
	"github.com/intelsdi-x/pulse/mgmt/tribe/worker"
	"github.com/pborman/uuid"

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
	clock        LClock
	agreements   map[string]*agreement.Agreement
	mutex        sync.RWMutex
	msgBuffer    []msg
	intentBuffer []msg
	broadcasts   *memberlist.TransmitLimitedQueue
	memberlist   *memberlist.Memberlist
	logger       *log.Entry
	members      map[string]*agreement.Member
	tags         map[string]string

	pluginCatalog   worker.ManagesPlugins
	taskManager     worker.ManagesTasks
	pluginWorkQueue chan worker.PluginRequest
	taskWorkQueue   chan worker.TaskRequest

	workerQuitChan  chan interface{}
	workerWaitGroup *sync.WaitGroup
}

type config struct {
	seed             string
	restAPIPort      int
	memberlistConfig *memberlist.Config
}

func DefaultConfig(name, advertiseAddr string, advertisePort int, seed string, restAPIPort int) *config {
	c := &config{seed: seed, restAPIPort: restAPIPort}
	c.memberlistConfig = memberlist.DefaultLANConfig()
	c.memberlistConfig.PushPullInterval = 300 * time.Second
	c.memberlistConfig.Name = name
	c.memberlistConfig.BindAddr = advertiseAddr
	c.memberlistConfig.BindPort = advertisePort
	c.memberlistConfig.GossipNodes = c.memberlistConfig.GossipNodes * 2
	return c
}

func New(c *config) (*tribe, error) {
	logger := logger.WithFields(log.Fields{
		"_block": "New",
		"port":   c.memberlistConfig.BindPort,
		"addr":   c.memberlistConfig.BindAddr,
		"name":   c.memberlistConfig.Name,
	})

	tribe := &tribe{
		agreements:      map[string]*agreement.Agreement{},
		members:         map[string]*agreement.Member{},
		msgBuffer:       make([]msg, 512),
		intentBuffer:    []msg{},
		logger:          logger.WithField("_name", c.memberlistConfig.Name),
		tags:            map[string]string{agreement.RestAPIPort: strconv.Itoa(c.restAPIPort)},
		pluginWorkQueue: make(chan worker.PluginRequest, 999),
		taskWorkQueue:   make(chan worker.TaskRequest, 999),
	}

	tribe.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return len(tribe.memberlist.Members())
		},
		RetransmitMult: memberlist.DefaultLANConfig().RetransmitMult,
	}

	//configure delegates
	c.memberlistConfig.Delegate = &delegate{tribe: tribe}
	c.memberlistConfig.Events = &memberDelegate{tribe: tribe}

	ml, err := memberlist.Create(c.memberlistConfig)
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "New",
		}).Error(err)
		return nil, err
	}
	tribe.memberlist = ml

	if c.seed != "" {
		_, err := ml.Join([]string{c.seed})
		if err != nil {
			logger.WithFields(log.Fields{
				"seed": c.seed,
			}).Error(errMemberlistJoin)
			return nil, errMemberlistJoin
		}
		logger.WithFields(log.Fields{
			"seed": c.seed,
		}).Infoln("tribe started")
		return tribe, nil
	}
	logger.WithFields(log.Fields{
		"seed": "none",
	}).Infoln("Tribe started")
	return tribe, nil
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
	err := t.memberlist.Leave(1 * time.Second)
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "Stop",
		}).Error(err)
	}
	err = t.memberlist.Shutdown()
	if err != nil {
		logger.WithFields(log.Fields{
			"_block": "Stop",
		}).Error(err)
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
	logger.Debugf("agreement %s has %d members", m.PluginAgreement.Name, len(t.agreements[m.PluginAgreement.Name].Members))
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
		logger.WithFields(log.Fields{
			"_block": "decodeTags",
			"error":  err,
		}).Error("Failed to decode tags")
	}
	return tags
}

// HandleGomitEvent handles events emitted from control
func (t *tribe) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.LoadPluginEvent:
		logger.WithFields(log.Fields{
			"_block":         "HandleGomitEvent",
			"event":          e.Namespace(),
			"plugin_name":    v.Name,
			"plugin_version": v.Version,
			"plugin_type":    core.PluginType(v.Type).String(),
		}).Debugf("Handling load plugin event")
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
			"_block":         "HandleGomitEvent",
			"event":          e.Namespace(),
			"plugin_name":    v.Name,
			"plugin_version": v.Version,
			"plugin_type":    core.PluginType(v.Type).String(),
		}).Debugf("Handling unload plugin event")
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
		logger.WithFields(log.Fields{
			"_block":               "HandleGomitEvent",
			"event":                e.Namespace(),
			"task_id":              v.TaskID,
			"task_start_on_create": v.StartOnCreate,
		}).Debugf("Handling task create event")
		task := agreement.Task{
			ID: v.TaskID,
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

func (t *tribe) LeaveAgreement(agreementName, memberName string) perror.PulseError {
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

func (t *tribe) JoinAgreement(agreementName, memberName string) perror.PulseError {
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

func (t *tribe) GetAgreement(name string) (*agreement.Agreement, perror.PulseError) {
	a, ok := t.agreements[name]
	if !ok {
		return nil, perror.New(errAgreementDoesNotExist, map[string]interface{}{"agreement_name": name})
	}
	return a, nil
}

func (t *tribe) GetAgreements() map[string]*agreement.Agreement {
	return t.agreements
}

func (t *tribe) AddTask(agreementName string, task agreement.Task) perror.PulseError {
	if err := t.canAddTask(task, agreementName); err != nil {
		return err
	}
	msg := &taskMsg{
		LTime:         t.clock.Increment(),
		TaskID:        task.ID,
		AgreementName: agreementName,
		UUID:          uuid.New(),
		Type:          addTaskMsgType,
	}
	if t.handleAddTask(msg) {
		t.broadcast(addTaskMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) RemoveTask(agreementName string, task agreement.Task) perror.PulseError {
	if err := t.canRemoveTask(agreementName, task); err != nil {
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

func (t *tribe) AddAgreement(name string) perror.PulseError {
	if _, ok := t.agreements[name]; ok {
		fields := log.Fields{
			"agreement": name,
		}
		return perror.New(errAgreementAlreadyExists, fields)
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

func (t *tribe) RemoveAgreement(name string) perror.PulseError {
	if _, ok := t.agreements[name]; !ok {
		fields := log.Fields{
			"Agreement": name,
		}
		return perror.New(errAgreementDoesNotExist, fields)
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
					logger.WithFields(log.Fields{
						"_block":         "handleRemovePlugin",
						"plugin_name":    msg.Plugin.Name(),
						"plugin_type":    msg.Plugin.TypeName(),
						"plugin_version": msg.Plugin.Version(),
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
					ID: msg.TaskID,
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
			t.processIntents()
			return true
		}
	}

	t.addTaskIntent(msg)
	return true
}

func (t *tribe) handleMemberJoin(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; !ok {
		t.members[n.Name] = agreement.NewMember(n)
		t.members[n.Name].Tags = t.decodeTags(n.Meta)
	}
	t.processIntents()
}

func (t *tribe) handleMemberLeave(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; ok {
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

func (t *tribe) joinAgreement(msg *agreementMsg) perror.PulseError {
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
	return nil
}

func (t *tribe) leaveAgreement(msg *agreementMsg) perror.PulseError {
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

func (t *tribe) canLeaveAgreement(agreementName, memberName string) perror.PulseError {
	if _, ok := t.agreements[agreementName]; !ok {
		fields := log.Fields{
			"Agreement": agreementName,
		}
		logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return perror.New(errAgreementDoesNotExist, fields)
	}

	m, ok := t.members[memberName]
	if !ok {
		fields := log.Fields{
			"MemberName": memberName,
		}
		t.logger.WithFields(fields).Debugln(errUnknownMember)
		return perror.New(errUnknownMember, fields)
	}
	if m.PluginAgreement == nil {
		fields := log.Fields{
			"MemberName": t.memberlist.LocalNode().Name,
			"Agreement":  agreementName,
		}
		t.logger.WithFields(fields).Debugln(errNotAMember)
		return perror.New(errNotAMember, fields)
	}
	return nil
}

func (t *tribe) canJoinAgreement(agreementName, memberName string) perror.PulseError {
	if _, ok := t.agreements[agreementName]; !ok {
		fields := log.Fields{
			"Agreement": agreementName,
		}
		logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return perror.New(errAgreementDoesNotExist, fields)
	}
	m, ok := t.members[memberName]
	if !ok {
		fields := log.Fields{
			"MemberName": memberName,
		}
		t.logger.WithFields(fields).Debugln(errUnknownMember)
		return perror.New(errUnknownMember, fields)

	}
	if m.PluginAgreement != nil && len(m.PluginAgreement.Plugins) > 0 {
		fields := log.Fields{
			"MemberName": t.memberlist.LocalNode().Name,
			"Agreement":  agreementName,
		}
		t.logger.WithFields(fields).Debugln(errAlreadyMemberOfPluginAgreement)
		return perror.New(errAlreadyMemberOfPluginAgreement, fields)
	}
	return nil
}

func (t *tribe) canAddTask(task agreement.Task, agreementName string) perror.PulseError {
	fields := log.Fields{
		"Agreement": agreementName,
	}
	a, ok := t.agreements[agreementName]
	if !ok {
		logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return perror.New(errAgreementDoesNotExist, fields)
	}
	if ok, _ := a.TaskAgreement.Tasks.Contains(task); ok {
		logger.WithFields(fields).Debugln(errTaskAlreadyExists)
		return perror.New(errTaskAlreadyExists, fields)
	}
	return nil
}

func (t *tribe) canRemoveTask(agreementName string, task agreement.Task) perror.PulseError {
	fields := log.Fields{
		"Agreement": agreementName,
	}
	a, ok := t.agreements[agreementName]
	if !ok {
		logger.WithFields(fields).Debugln(errAgreementDoesNotExist)
		return perror.New(errAgreementDoesNotExist, fields)
	}
	if ok, _ := a.TaskAgreement.Tasks.Contains(task); !ok {
		logger.WithFields(fields).Debugln(errTaskDoesNotExist)
		return perror.New(errTaskDoesNotExist, fields)
	}
	return nil
}

func (t *tribe) isDuplicate(msg msg) bool {
	// is the message old
	if t.clock.Time() > LTime(len(t.msgBuffer)) &&
		msg.Time() < t.clock.Time()-LTime(len(t.msgBuffer)) {
		t.logger.WithFields(log.Fields{
			"event_clock": msg.Time(),
			"event":       msg.GetType().String(),
			"event_uuid":  msg.ID(),
			"clock":       t.clock.Time(),
			"agreement":   msg.Agreement(),
			// "plugin":      msg.Plugin,
		}).Debugln("This message is old")
		return true
	}

	// have we seen it
	idx := msg.Time() % LTime(len(t.msgBuffer))
	seen := t.msgBuffer[idx]
	if seen != nil && seen.ID() == msg.ID() {
		t.logger.WithFields(log.Fields{
			"event_clock": msg.Time(),
			"event":       msg.GetType().String(),
			"event_uuid":  msg.ID(),
			"clock":       t.clock.Time(),
			"agreement":   msg.Agreement(),
			// "plugin":      msg.Plugin,
		}).Debugln("duplicate message")

		return true
	}
	return false
}

func (t *tribe) addPluginIntent(msg *pluginMsg) bool {
	t.logger.WithFields(log.Fields{
		"event_clock": msg.LTime,
		"agreement":   msg.AgreementName,
		"plugin":      msg.Plugin,
		"type":        msg.Type.String(),
	}).Debugln("Out of order msg")
	t.intentBuffer = append(t.intentBuffer, msg)
	return true
}

func (t *tribe) addAgreementIntent(m msg) bool {
	t.logger.WithFields(log.Fields{
		"event_clock": m.Time(),
		"agreement":   m.Agreement(),
		"type":        m.GetType().String(),
	}).Debugln("Out of order msg")
	t.intentBuffer = append(t.intentBuffer, m)
	return true
}

func (t *tribe) addTaskIntent(m *taskMsg) bool {
	t.logger.WithFields(log.Fields{
		"event_clock": m.Time(),
		"agreement":   m.Agreement(),
		"type":        m.GetType().String(),
		"task_id":     m.TaskID,
	}).Debugln("Out of order msg")
	t.intentBuffer = append(t.intentBuffer, m)
	return true
}
