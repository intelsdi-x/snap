package tribe

import (
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/pborman/uuid"

	"github.com/hashicorp/memberlist"
)

var (
	errAgreementDoesNotExist  = errors.New("Agreement does not exist")
	errAgreementAlreadyExists = errors.New("Agreement already exists")
	errUnknownMember          = errors.New("Unknown member")
	errAlreadyMember          = errors.New("Already a member of agreement")
	errNotAMember             = errors.New("Not a member of agreement")
)

var logger = log.WithFields(log.Fields{
	"_module": "tribe",
})

type agreements struct {
	PluginAgreement *pluginAgreement //TODO unexport
	members         map[string]*member
}

type pluginAgreement struct {
	Plugins []plugin //TODO unexport
	mutex   sync.RWMutex
}

type taskAgreement struct {
	Tasks []task
	mutex sync.RWMutex
}

type task struct {
	Name string
	ID   uint64
}

type plugin struct {
	Name    string
	Version int
}

type tribe struct {
	clock        LClock
	agreements   map[string]*agreements
	mutex        sync.RWMutex
	msgBuffer    []msg
	intentBuffer []msg
	broadcasts   *memberlist.TransmitLimitedQueue
	memberlist   *memberlist.Memberlist
	logger       *log.Entry
	members      map[string]*member
}

type member struct {
	name            string
	node            *memberlist.Node
	pluginAgreement *pluginAgreement
}

func newMember(node *memberlist.Node) *member {
	return &member{name: node.Name, node: node}
}

// TODO refactor this so a config struct is passed in

// New returns a tribe instance
func New(name, advertiseAddr string, advertisePort int, seed string) (*tribe, error) {
	tribe := &tribe{
		agreements:   map[string]*agreements{},
		members:      map[string]*member{},
		msgBuffer:    make([]msg, 512),
		intentBuffer: []msg{},
		logger:       logger.WithField("_name", name),
	}

	tribe.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return len(tribe.memberlist.Members())
		},
		RetransmitMult: memberlist.DefaultLANConfig().RetransmitMult,
	}

	// mlConf := memberlist.DefaultLocalConfig()
	mlConf := memberlist.DefaultLANConfig()
	mlConf.PushPullInterval = 300 * time.Second
	mlConf.Name = name
	mlConf.BindAddr = advertiseAddr
	mlConf.BindPort = advertisePort
	mlConf.Delegate = &delegate{tribe: tribe}
	mlConf.Events = &memberDelegate{tribe: tribe}

	ml, err := memberlist.Create(mlConf)
	if err != nil {
		//todo remove fatal
		log.Fatal("Failed to create memberlist: " + err.Error())
	}
	tribe.memberlist = ml

	if seed != "" {
		// Join an existing cluster by specifying at least one known member.
		_, err := ml.Join([]string{seed})
		if err != nil {
			//todo remove fatal
			log.Fatal("Failed to join cluster: " + err.Error())
		}

	}
	return tribe, nil
}

func newAgreements() *agreements {
	return &agreements{
		PluginAgreement: &pluginAgreement{
			Plugins: []plugin{},
		},
		members: map[string]*member{},
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

func (t *tribe) AddPlugin(agreentName, pluginName string, ver int) error {
	if _, ok := t.agreements[agreentName]; !ok {
		return errAgreementDoesNotExist
	}
	msg := &tribeMsg{
		LTime:         t.clock.Increment(),
		Plugin:        plugin{Name: pluginName, Version: ver},
		AgreementName: agreentName,
		UUID:          uuid.New(),
		Type:          addPluginMsgType,
	}
	if t.handleAddPlugin(msg) {
		t.broadcast(addPluginMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) RemovePlugin(clan, name string, ver int) error {
	if _, ok := t.agreements[clan]; !ok {
		return errAgreementDoesNotExist
	}
	msg := &tribeMsg{
		LTime:         t.clock.Increment(),
		Plugin:        plugin{Name: name, Version: ver},
		AgreementName: clan,
		UUID:          uuid.New(),
		Type:          removePluginMsgType,
	}
	if t.handleRemovePlugin(msg) {
		t.broadcast(removePluginMsgType, msg, nil)
	}
	return nil
}

func (t *tribe) AddAgreement(name string) perror.PulseError {
	if _, ok := t.agreements[name]; ok {
		fields := log.Fields{
			"Agreement": name,
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
			t.processJoinAgreementIntents() {
			return
		}
	}
}

func (t *tribe) processAddPluginIntents() bool {
	for idx, v := range t.intentBuffer {
		if v.GetType() == addPluginMsgType {
			intent := v.(*tribeMsg)
			if _, ok := t.agreements[intent.AgreementName]; ok {
				if ok, _ := containsPlugin(t.agreements[intent.AgreementName].PluginAgreement.Plugins, intent.Plugin); !ok {
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
			intent := v.(*tribeMsg)
			if _, ok := t.agreements[intent.AgreementName]; ok {
				if ok, idx := containsPlugin(t.agreements[intent.AgreementName].PluginAgreement.Plugins, intent.Plugin); ok {
					t.agreements[intent.AgreementName].PluginAgreement.Plugins = append(t.agreements[intent.AgreementName].PluginAgreement.Plugins[:idx], t.agreements[intent.AgreementName].PluginAgreement.Plugins[idx+1:]...)
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
				t.agreements[intent.AgreementName] = newAgreements()
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
			if member, ok := t.members[intent.MemberName]; ok {
				if _, ok := t.agreements[intent.AgreementName]; ok {
					if member.pluginAgreement == nil {
						if _, ok := t.agreements[intent.AgreementName].members[intent.MemberName]; !ok {
							t.agreements[intent.AgreementName].members[intent.MemberName] = member
							t.members[intent.MemberName].pluginAgreement = t.agreements[intent.Agreement()].PluginAgreement
							t.intentBuffer = append(t.intentBuffer[:idx], t.intentBuffer[idx+1:]...)
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func (t *tribe) handleRemovePlugin(msg *tribeMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.Agreement()]; ok {
		if t.agreements[msg.AgreementName].PluginAgreement.remove(msg, t.logger) {
			t.processIntents()
			return true
		}
	}

	t.addPluginIntent(msg)
	return true
}

func (t *tribe) handleAddPlugin(msg *tribeMsg) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// update the clock if newer
	t.clock.Update(msg.LTime)

	if t.isDuplicate(msg) {
		return false
	}

	t.msgBuffer[msg.LTime%LTime(len(t.msgBuffer))] = msg

	if _, ok := t.agreements[msg.AgreementName]; ok {
		if t.agreements[msg.AgreementName].PluginAgreement.add(msg, t.logger) {
			t.processIntents()
			return true
		}
	}

	t.addPluginIntent(msg)
	return true
}

func (t *tribe) handleMemberJoin(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; !ok {
		t.members[n.Name] = newMember(n)
	}
}

func (t *tribe) handleMemberLeave(n *memberlist.Node) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.members[n.Name]; ok {
		delete(t.members, n.Name)
	}
}

func (t *tribe) handleMemberUpdate(n *memberlist.Node) {

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
		t.agreements[msg.AgreementName] = newAgreements()
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
	t.members[msg.MemberName].pluginAgreement = t.agreements[msg.Agreement()].PluginAgreement
	// update the agreements membership
	t.agreements[msg.Agreement()].members[msg.MemberName] = t.members[msg.MemberName]
	return nil
}

func (t *tribe) leaveAgreement(msg *agreementMsg) perror.PulseError {
	if err := t.canLeaveAgreement(msg.Agreement(), msg.MemberName); err != nil {
		return err
	}

	delete(t.agreements[msg.AgreementName].members, msg.MemberName)
	t.members[msg.MemberName].pluginAgreement = nil

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
	if m.pluginAgreement == nil {
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
	if m.pluginAgreement != nil {
		fields := log.Fields{
			"MemberName": t.memberlist.LocalNode().Name,
			"Agreement":  agreementName,
		}
		t.logger.WithFields(fields).Debugln(errAlreadyMember)
		return perror.New(errAlreadyMember, fields)
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

// contains - Returns boolean indicating whether the plugin was found.
// If the plugin is found the index returned as the second return value.
func containsPlugin(items []plugin, item plugin) (bool, int) {
	for idx, i := range items {
		if i.Name == item.Name && i.Version == item.Version {
			return true, idx
		}
	}
	return false, -1
}

// remove - removes a plugin from the agreed plugins
func (a *pluginAgreement) remove(msg *tribeMsg, tlogger *log.Entry) bool {
	tlogger.WithFields(log.Fields{
		"event_clock": msg.LTime,
		"event":       msg.Type.String(),
		"agreement":   msg.AgreementName,
		"plugin":      msg.Plugin,
	}).Debugln("Removing plugin")
	if ok, idx := containsPlugin(a.Plugins, msg.Plugin); ok {
		a.Plugins = append(a.Plugins[idx+1:], a.Plugins[:idx]...)
		return true
	}
	return false
}

func (a *pluginAgreement) add(msg *tribeMsg, tlogger *log.Entry) bool {
	tlogger.WithFields(log.Fields{
		"event_clock": msg.LTime,
		"agreement":   msg.AgreementName,
		"plugin":      msg.Plugin,
		"_block":      "add",
	}).Debugln("Adding plugin")
	if ok, _ := containsPlugin(a.Plugins, msg.Plugin); ok {
		return false
	}
	a.Plugins = append(a.Plugins, msg.Plugin)
	return true
}

func (t *tribe) addPluginIntent(msg *tribeMsg) bool {
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
