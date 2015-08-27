package tribe

type delegate struct {
	tribe *tribe
}

func (t *delegate) NodeMeta(limit int) []byte {
	logger.WithField("_block", "NodeMeta").Debugln("NodeMeta called")
	//consider using NodeMeta to store declared, inferred and derived facts
	return []byte{}
}

func (t *delegate) NotifyMsg(buf []byte) {
	if len(buf) == 0 {
		return
	}

	var rebroadcast = true

	switch msgType(buf[0]) {
	case addPluginMsgType:
		msg := &tribeMsg{}
		if err := decodeMessage(buf[1:], msg); err != nil {
			panic(err)
		}
		rebroadcast = t.tribe.handleAddPlugin(msg)
	case removePluginMsgType:
		msg := &tribeMsg{}
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
	t.tribe.mutex.Lock()
	defer t.tribe.mutex.Unlock()

	fs := fullStateMsg{
		LTime:      t.tribe.clock.Time(),
		PluginMsgs: t.tribe.msgBuffer,
		Agreements: map[string]*agreements{},
	}

	for name, agreements := range t.tribe.agreements {
		agreements.PluginAgreement.mutex.Lock()
		fs.Agreements[name] = agreements
		agreements.PluginAgreement.mutex.Unlock()
	}

	buf, err := encodeMessage(fullStateMsgType, fs)
	if err != nil {
		panic(err)
	}

	return buf
}

func (t *delegate) MergeRemoteState(buf []byte, join bool) {
	logger = logger.WithField("_block", "MergeRemoteState")
	logger.Debugln("calling merge")

	if msgType(buf[0]) != fullStateMsgType {
		logger.Errorln("NodeMeta called")
		return
	}

	fs := &fullStateMsg{}
	if err := decodeMessage(buf[1:], fs); err != nil {
		panic(err)
	}

	if t.tribe.clock.Time() > fs.LTime {
		//we are ahead return now
		return
	}

	logger.Debugln("Updating full state")
	if join {
		t.tribe.mutex.Lock()
		t.tribe.agreements = fs.Agreements
		t.tribe.msgBuffer = fs.PluginMsgs
		t.tribe.clock.Update(fs.LTime - 1)
		t.tribe.mutex.Unlock()
		//todo what about the intents???
	} else {
		//Process Plugin adds
		for _, m := range fs.PluginMsgs {
			// if m == nil {
			// 	continue
			// }
			if m.GetType() == addPluginMsgType {
				t.tribe.handleAddPlugin(m.(*tribeMsg))
			}
		}
	}

}
