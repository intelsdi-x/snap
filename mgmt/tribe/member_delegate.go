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
