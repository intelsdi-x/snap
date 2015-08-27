package tribe

import (
	"bytes"

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
)

var msgTypes = []string{
	"Add plugin",
	"Remove plugin",
	"Add agreement",
	"Remove agreement",
	"Full state",
	"Join agreement",
	"Leave agreement",
}

func (m msgType) String() string {
	return msgTypes[int(m)]
}

type msg interface {
	ID() string
	Time() LTime
	GetType() msgType //TODO rename to Type
	Agreement() string
}

type tribeMsg struct {
	LTime         LTime
	Plugin        plugin
	UUID          string
	AgreementName string
	Type          msgType
}

func (t *tribeMsg) ID() string {
	return t.UUID
}

func (t *tribeMsg) Time() LTime {
	return t.LTime
}

func (t *tribeMsg) GetType() msgType {
	return t.Type
}

func (t *tribeMsg) Agreement() string {
	return t.AgreementName
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

type fullStateMsg struct {
	LTime      LTime
	PluginMsgs []msg //TODO rename Msgs
	Agreements map[string]*agreements
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
