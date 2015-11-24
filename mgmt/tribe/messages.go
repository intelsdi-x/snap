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
	"fmt"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
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
	stopTaskMsgType
	startTaskMsgType
	getTaskStateMsgType
	taskStateQueryResponseMsgType
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
	"Stop task",
	"Start task",
	"Get task state",
	"Get task state response",
}

func (m msgType) String() string {
	return msgTypes[int(m)]
}

type msg interface {
	ID() string
	Time() LTime
	GetType() msgType
	Agreement() string
	String() string
}

type pluginMsg struct {
	LTime         LTime
	Plugin        agreement.Plugin
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
	APIPort       int
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
	StartOnCreate bool
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

type taskStateQueryMsg struct {
	LTime         LTime
	UUID          string
	Deadline      time.Time
	AgreementName string
	TaskID        string
	Addr          []byte
	Port          uint16
	Type          msgType
}

func (t *taskStateQueryMsg) Agreement() string {
	return t.AgreementName
}

func (t *taskStateQueryMsg) GetType() msgType {
	return t.Type
}

func (t *taskStateQueryMsg) ID() string {
	return t.UUID
}

func (t *taskStateQueryMsg) Time() LTime {
	return t.LTime
}

func (t *taskStateQueryMsg) String() string {
	return fmt.Sprintf("msg type='%v' agreementName='%v' uuid='%v' task='%v'",
		t.GetType(), t.Agreement(), t.ID(), t.TaskID)
}

type taskStateQueryResponseMsg struct {
	LTime LTime
	UUID  string
	From  string
	State core.TaskState
}

type fullStateMsg struct {
	LTime               LTime
	PluginMsgs          []*pluginMsg
	AgreementMsgs       []*agreementMsg
	TaskMsgs            []*taskMsg
	PluginIntentMsgs    []*pluginMsg
	AgreementIntentMsgs []*agreementMsg
	TaskIntentMsgs      []*taskMsg

	Agreements map[string]*agreement.Agreement
	Members    map[string]*agreement.Member
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
