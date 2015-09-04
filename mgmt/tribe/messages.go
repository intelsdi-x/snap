package tribe

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-msgpack/codec"
)

type msgType uint8

const (
	addPluginMsgType msgType = iota
	removePluginMsgType
	addAgreementMsgType
	removeAgreementMsgType
	fullStateMsgType
	joinAgreementMsgType
	leaveAgreementMsgType
	addTaskMsgType
	removeTaskMsgType
)

var msgTypes = []string{
	"Add plugin",
	"Remove plugin",
	"Add agreement",
	"Remove agreement",
	"Full state",
	"Join agreement",
	"Leave agreement",
	"Add task",
	"Remove task",
}

func (m msgType) String() string {
	return msgTypes[int(m)]
}

type msg interface {
	ID() string
	Time() LTime
	GetType() msgType //TODO rename to Type
	Agreement() string
	String() string
}

type pluginMsg struct {
	LTime         LTime
	Plugin        plugin
	UUID          string
	AgreementName string
	Type          msgType
}

func (t *pluginMsg) ID() string {
	return t.UUID
}

func (t *pluginMsg) Time() LTime {
	return t.LTime
}

func (t *pluginMsg) GetType() msgType {
	return t.Type
}

func (t *pluginMsg) Agreement() string {
	return t.AgreementName
}

func (t *pluginMsg) String() string {
	return fmt.Sprintf("msg type='%v' agreementName='%v' uuid='%v' plugin='%v'",
		t.GetType(), t.Agreement(), t.ID(), t.Plugin)
}

type agreementMsg struct {
	LTime         LTime
	UUID          string
	AgreementName string
	MemberName    string
	Type          msgType
}

func (a *agreementMsg) ID() string {
	return a.UUID
}

func (a *agreementMsg) Time() LTime {
	return a.LTime
}

func (a *agreementMsg) GetType() msgType {
	return a.Type
}

func (a *agreementMsg) Agreement() string {
	return a.AgreementName
}

func (a *agreementMsg) String() string {
	return fmt.Sprintf("msg type='%v' agreementName='%v' uuid='%v' member='%v'",
		a.GetType(), a.Agreement(), a.ID(), a.MemberName)
}

type taskMsg struct {
	LTime         LTime
	UUID          string
	TaskID        string
	AgreementName string
	Type          msgType
}

func (t *taskMsg) ID() string {
	return t.UUID
}

func (t *taskMsg) Time() LTime {
	return t.LTime
}

func (t *taskMsg) GetType() msgType {
	return t.Type
}

func (t *taskMsg) Agreement() string {
	return t.AgreementName
}

func (t *taskMsg) String() string {
	return fmt.Sprintf("msg type='%v' agreementName='%v' uuid='%v' task='%v'",
		t.GetType(), t.Agreement(), t.ID(), t.TaskID)
}

type fullStateMsg struct {
	LTime               LTime
	PluginMsgs          []*pluginMsg
	AgreementMsgs       []*agreementMsg
	TaskMsgs            []*taskMsg
	PluginIntentMsgs    []*pluginMsg
	AgreementIntentMsgs []*agreementMsg
	TaskIntentMsgs      []*taskMsg

	Agreements map[string]*agreements
	Members    map[string]*member
}

func decodeMessage(buf []byte, out interface{}) error {
	var handle codec.MsgpackHandle
	return codec.NewDecoder(bytes.NewReader(buf), &handle).Decode(out)
}

func encodeMessage(t msgType, msg interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(t))

	handle := codec.MsgpackHandle{}
	encoder := codec.NewEncoder(buf, &handle)
	err := encoder.Encode(msg)
	return buf.Bytes(), err
}
