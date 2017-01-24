// +build legacy small medium large

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

package fixtures

import (
	"net"

	"github.com/hashicorp/memberlist"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
)

var (
	mockTribeAgreement *agreement.Agreement
	mockTribeMember    *agreement.Member
)

func init() {
	mockTribeMember = agreement.NewMember(&memberlist.Node{
		Name: "Imma_Mock",
		Addr: net.ParseIP("193.11.22.11"),
		Port: uint16(0),
		Meta: []byte("meta"), // Metadata from the delegate for this node.
		PMin: uint8(0),       // Minimum protocol version this understands
		PMax: uint8(0),       // Maximum protocol version this understands
		PCur: uint8(0),       // Current version node is speaking
		DMin: uint8(0),       // Min protocol version for the delegate to understand
		DMax: uint8(0),       // Max protocol version for the delegate to understand
		DCur: uint8(0),       // Current version delegate is speaking
	})

	mockTribeAgreement = agreement.New("Agree1")
	mockTribeAgreement.PluginAgreement.Add(
		agreement.Plugin{Name_: "mockVersion", Version_: 1, Type_: core.CollectorPluginType})
	mockTribeAgreement.TaskAgreement.Add(
		agreement.Task{ID: "mockTask", StartOnCreate: true})
	mockTribeAgreement.Members["member1"] = agreement.NewMember(&memberlist.Node{
		Name: "mockName",
		Addr: net.ParseIP("193.34.23.11"),
		Port: uint16(0),
		Meta: []byte("meta"), // Metadata from the delegate for this node.
		PMin: uint8(0),       // Minimum protocol version this understands
		PMax: uint8(0),       // Maximum protocol version this understands
		PCur: uint8(0),       // Current version node is speaking
		DMin: uint8(0),       // Min protocol version for the delegate to understand
		DMax: uint8(0),       // Max protocol version for the delegate to understand
		DCur: uint8(0),       // Current version delegate is speaking
	})
}

type MockTribeManager struct{}

func (m *MockTribeManager) GetAgreement(name string) (*agreement.Agreement, serror.SnapError) {
	return mockTribeAgreement, nil
}
func (m *MockTribeManager) GetAgreements() map[string]*agreement.Agreement {
	return map[string]*agreement.Agreement{
		"Agree1": mockTribeAgreement,
		"Agree2": mockTribeAgreement}
}
func (m *MockTribeManager) AddAgreement(name string) serror.SnapError {
	return nil
}
func (m *MockTribeManager) RemoveAgreement(name string) serror.SnapError {
	return nil
}
func (m *MockTribeManager) JoinAgreement(agreementName, memberName string) serror.SnapError {
	return nil
}
func (m *MockTribeManager) LeaveAgreement(agreementName, memberName string) serror.SnapError {
	return nil
}
func (m *MockTribeManager) GetMembers() []string {
	return []string{"one", "two", "three"}
}
func (m *MockTribeManager) GetMember(name string) *agreement.Member {
	return mockTribeMember
}

// These constants are the expected tribe responses from running
// rest_v1_test.go on the tribe routes found in mgmt/rest/server.go
const (
	GET_TRIBE_AGREEMENTS_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreements retrieved",
    "type": "tribe_agreement_list_returned",
    "version": 1
  },
  "body": {
    "agreements": {
      "Agree1": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      },
      "Agree2": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      }
    }
  }
}`

	GET_TRIBE_AGREEMENTS_RESPONSE_NAME = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreement returned",
    "type": "tribe_agreement_returned",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "Agree1",
      "plugin_agreement": {
        "plugins": [
          {
            "name": "mockVersion",
            "version": 1,
            "type": 0
          }
        ]
      },
      "task_agreement": {
        "tasks": [
          {
            "id": "mockTask",
            "start_on_create": true
          }
        ]
      },
      "members": {
        "member1": {
          "name": "mockName"
        }
      }
    }
  }
}`

	GET_TRIBE_MEMBERS_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Tribe members retrieved",
    "type": "tribe_member_list_returned",
    "version": 1
  },
  "body": {
    "members": [
      "one",
      "two",
      "three"
    ]
  }
}`

	GET_TRIBE_MEMBER_NAME = `{
  "meta": {
    "code": 200,
    "message": "Tribe member details retrieved",
    "type": "tribe_member_details_returned",
    "version": 1
  },
  "body": {
    "name": "Imma_Mock",
    "plugin_agreement": "",
    "tags": null,
    "task_agreements": null
  }
}`

	DELETE_TRIBE_AGREEMENT_RESPONSE_NAME = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreement deleted",
    "type": "tribe_agreement_deleted",
    "version": 1
  },
  "body": {
    "agreements": {
      "Agree1": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      },
      "Agree2": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      }
    }
  }
}`

	JOIN_TRIBE_AGREEMENT_RESPONSE_NAME_JOIN = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreement joined",
    "type": "tribe_agreement_joined",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "Agree1",
      "plugin_agreement": {
        "plugins": [
          {
            "name": "mockVersion",
            "version": 1,
            "type": 0
          }
        ]
      },
      "task_agreement": {
        "tasks": [
          {
            "id": "mockTask",
            "start_on_create": true
          }
        ]
      },
      "members": {
        "member1": {
          "name": "mockName"
        }
      }
    }
  }
}`

	LEAVE_TRIBE_AGREEMENT_RESPONSE_NAME_LEAVE = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreement left",
    "type": "tribe_agreement_left",
    "version": 1
  },
  "body": {
    "agreement": {
      "name": "Agree1",
      "plugin_agreement": {
        "plugins": [
          {
            "name": "mockVersion",
            "version": 1,
            "type": 0
          }
        ]
      },
      "task_agreement": {
        "tasks": [
          {
            "id": "mockTask",
            "start_on_create": true
          }
        ]
      },
      "members": {
        "member1": {
          "name": "mockName"
        }
      }
    }
  }
}`

	ADD_TRIBE_AGREEMENT_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Tribe agreement created",
    "type": "tribe_agreement_created",
    "version": 1
  },
  "body": {
    "agreements": {
      "Agree1": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      },
      "Agree2": {
        "name": "Agree1",
        "plugin_agreement": {
          "plugins": [
            {
              "name": "mockVersion",
              "version": 1,
              "type": 0
            }
          ]
        },
        "task_agreement": {
          "tasks": [
            {
              "id": "mockTask",
              "start_on_create": true
            }
          ]
        },
        "members": {
          "member1": {
            "name": "mockName"
          }
        }
      }
    }
  }
}`
)
