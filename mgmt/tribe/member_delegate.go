package tribe

import (
	"github.com/hashicorp/memberlist"
)

type memberDelegate struct {
	tribe *tribe
}

func (m *memberDelegate) NotifyJoin(n *memberlist.Node) {
	m.tribe.handleMemberJoin(n)
}

func (m *memberDelegate) NotifyLeave(n *memberlist.Node) {
	m.tribe.handleMemberLeave(n)
}

func (m *memberDelegate) NotifyUpdate(n *memberlist.Node) {
	m.tribe.handleMemberUpdate(n)
}
