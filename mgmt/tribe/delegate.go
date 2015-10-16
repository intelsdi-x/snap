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

import "fmt"

type delegate struct {
	tribe *tribe
}

func (t *delegate) NodeMeta(limit int) []byte {
	logger.WithField("_block", "NodeMeta").Debugln("NodeMeta called")
	tags := t.tribe.encodeTags(t.tribe.tags)
	if len(tags) > limit {
		panic(fmt.Errorf("Node tags '%v' exceeds length limit of %d bytes", t.tribe.tags, limit))
	}
	return tags
}

func (t *delegate) NotifyMsg(buf []byte) {
	if len(buf) == 0 {
		return
	}

	var rebroadcast = true

	switch msgType(buf[0]) {
	case addPluginMsgType:
		msg := &pluginMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleAddPlugin(msg)
	case removePluginMsgType:
		msg := &pluginMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleRemovePlugin(msg)
	case addAgreementMsgType:
		msg := &agreementMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleAddAgreement(msg)
	case removeAgreementMsgType:
		msg := &agreementMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleRemoveAgreement(msg)
	case joinAgreementMsgType:
		msg := &agreementMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleJoinAgreement(msg)
	case leaveAgreementMsgType:
		msg := &agreementMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleLeaveAgreement(msg)
	case addTaskMsgType:
		msg := &taskMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleAddTask(msg)
	case removeTaskMsgType:
		msg := &taskMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleRemoveTask(msg)
	default:
		logger.WithField("_block", "NotifyMsg").Errorln("NodeMeta called")
		return
	}

	if rebroadcast {
		newBuf := make([]byte, len(buf))
		copy(newBuf, buf)
		t.tribe.broadcasts.QueueBroadcast(&broadcast{
			msg:    newBuf,
			notify: nil,
		})
	}
}

func (t *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return t.tribe.broadcasts.GetBroadcasts(overhead, limit)
}

func (t *delegate) LocalState(join bool) []byte {
	// t.tribe.mutex.RLock()
	// defer t.tribe.mutex.RUnlock()

	// TODO the sizes here need to be set with a flag that is also ref in tribe.go
	pluginMsgs := make([]*pluginMsg, 512)
	agreementMsgs := make([]*agreementMsg, 512)
	taskMsgs := make([]*taskMsg, 512)
	pluginIntentMsgs := make([]*pluginMsg, 512)
	agreementIntentMsgs := make([]*agreementMsg, 512)
	taskIntentMsgs := make([]*taskMsg, 512)

	for idx, msg := range t.tribe.msgBuffer {
		if msg == nil {
			continue
		}
		switch msg.GetType() {
		case addPluginMsgType:
			pluginMsgs[idx] = msg.(*pluginMsg)
		case removePluginMsgType:
			pluginMsgs[idx] = msg.(*pluginMsg)
		case addAgreementMsgType:
			agreementMsgs[idx] = msg.(*agreementMsg)
		case removeAgreementMsgType:
			agreementMsgs[idx] = msg.(*agreementMsg)
		case joinAgreementMsgType:
			agreementMsgs[idx] = msg.(*agreementMsg)
		case leaveAgreementMsgType:
			agreementMsgs[idx] = msg.(*agreementMsg)
		case addTaskMsgType:
			taskMsgs[idx] = msg.(*taskMsg)
		case removeTaskMsgType:
			taskMsgs[idx] = msg.(*taskMsg)
		}
	}

	for idx, msg := range t.tribe.intentBuffer {
		if msg == nil {
			continue
		}
		switch msg.GetType() {
		case addPluginMsgType:
			pluginIntentMsgs[idx] = msg.(*pluginMsg)
		case removePluginMsgType:
			pluginIntentMsgs[idx] = msg.(*pluginMsg)
		case addAgreementMsgType:
			agreementIntentMsgs[idx] = msg.(*agreementMsg)
		case removeAgreementMsgType:
			agreementIntentMsgs[idx] = msg.(*agreementMsg)
		case joinAgreementMsgType:
			agreementIntentMsgs[idx] = msg.(*agreementMsg)
		case leaveAgreementMsgType:
			agreementIntentMsgs[idx] = msg.(*agreementMsg)
		case addTaskMsgType:
			taskIntentMsgs[idx] = msg.(*taskMsg)
		case removeTaskMsgType:
			taskIntentMsgs[idx] = msg.(*taskMsg)
		}
	}

	fs := fullStateMsg{
		LTime:               t.tribe.clock.Time(),
		PluginMsgs:          pluginMsgs,
		AgreementMsgs:       agreementMsgs,
		TaskMsgs:            taskMsgs,
		PluginIntentMsgs:    pluginIntentMsgs,
		AgreementIntentMsgs: agreementIntentMsgs,
		TaskIntentMsgs:      taskIntentMsgs,
		Agreements:          t.tribe.agreements,
		Members:             t.tribe.members,
	}

	buf, err := encodeMessage(fullStateMsgType, fs)
	if err != nil {
		panic(err)
	}

	return buf
}

func (t *delegate) MergeRemoteState(buf []byte, join bool) {
	logger.WithField("_block", "MergeRemoteState").Debugln("calling merge")
	logger.Debugln("Updating full state")

	if msgType(buf[0]) != fullStateMsgType {
		logger.Errorln("Unknown message type")
		return
	}

	fs := &fullStateMsg{}
	if err := decodeMessage(buf[1:], fs); err != nil {
		panic(err)
	}

	if t.tribe.clock.Time() > fs.LTime {
		return
	}

	t.tribe.clock.Update(fs.LTime - 1)

	if join {
		t.tribe.agreements = fs.Agreements
		for k, v := range fs.Members {
			t.tribe.members[k] = v
		}
		// t.tribe.members = fs.Members
		for idx, pluginMsg := range fs.PluginMsgs {
			if pluginMsg == nil {
				continue
			}
			t.tribe.msgBuffer[idx] = pluginMsg
		}
		for idx, agreementMsg := range fs.AgreementMsgs {
			if agreementMsg == nil {
				continue
			}
			t.tribe.msgBuffer[idx] = agreementMsg
		}
		for idx, taskMsg := range fs.TaskMsgs {
			if taskMsg == nil {
				continue
			}
			t.tribe.msgBuffer[idx] = taskMsg
		}
		for idx, pluginMsg := range fs.PluginIntentMsgs {
			if pluginMsg == nil {
				continue
			}
			t.tribe.intentBuffer[idx] = pluginMsg
		}
		for idx, agreementMsg := range fs.AgreementIntentMsgs {
			if agreementMsg == nil {
				continue
			}
			t.tribe.intentBuffer[idx] = agreementMsg
		}
		for idx, taskMsg := range fs.TaskIntentMsgs {
			if taskMsg == nil {
				continue
			}
			t.tribe.intentBuffer[idx] = taskMsg
		}
	} else {
		for _, m := range fs.PluginMsgs {
			if m == nil {
				continue
			}
			if m.GetType() == addPluginMsgType {
				t.tribe.handleAddPlugin(m)
			}
			if m.GetType() == removePluginMsgType {
				t.tribe.handleRemovePlugin(m)
			}
		}
		for _, m := range fs.AgreementMsgs {
			if m == nil {
				continue
			}
			if m.GetType() == addAgreementMsgType {
				t.tribe.handleAddAgreement(m)
			}
			if m.GetType() == removeAgreementMsgType {
				t.tribe.handleRemoveAgreement(m)
			}
			if m.GetType() == joinAgreementMsgType {
				t.tribe.handleJoinAgreement(m)
			}
			if m.GetType() == leaveAgreementMsgType {
				t.tribe.handleLeaveAgreement(m)
			}
		}
		for _, m := range fs.TaskMsgs {
			if m == nil {
				continue
			}
			if m.GetType() == addTaskMsgType {
				t.tribe.handleAddTask(m)
			}
			if m.GetType() == removeTaskMsgType {
				t.tribe.handleRemoveTask(m)
			}
		}
	}

}
