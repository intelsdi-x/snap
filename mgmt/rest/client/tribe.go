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

package client

import (
	"encoding/json"
	"fmt"

	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
)

func (c *Client) ListMembers() *ListMembersResult {
	resp, err := c.do("GET", "/tribe/members", ContentTypeJSON, nil)
	if err != nil {
		return &ListMembersResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeMemberListType:
		// Success
		return &ListMembersResult{resp.Body.(*rbody.TribeMemberList), nil}
	case rbody.ErrorType:
		return &ListMembersResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &ListMembersResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) GetMember(name string) *GetMemberResult {
	resp, err := c.do("GET", fmt.Sprintf("/tribe/member/%s", name), ContentTypeJSON, nil)
	if err != nil {
		return &GetMemberResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeMemberShowType:
		// Success
		return &GetMemberResult{resp.Body.(*rbody.TribeMemberShow), nil}
	case rbody.ErrorType:
		return &GetMemberResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &GetMemberResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) ListAgreements() *ListAgreementResult {
	resp, err := c.do("GET", "/tribe/agreements", ContentTypeJSON, nil)
	if err != nil {
		return &ListAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeListAgreementType:
		return &ListAgreementResult{resp.Body.(*rbody.TribeListAgreement), nil}
	case rbody.ErrorType:
		return &ListAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &ListAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) AddAgreement(name string) *AddAgreementResult {
	b, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: name})
	if err != nil {
		return &AddAgreementResult{Err: err}
	}
	resp, err := c.do("POST", "/tribe/agreements", ContentTypeJSON, b)
	if err != nil {
		return &AddAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeAddAgreementType:
		return &AddAgreementResult{resp.Body.(*rbody.TribeAddAgreement), nil}
	case rbody.ErrorType:
		return &AddAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &AddAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) DeleteAgreement(name string) *DeleteAgreementResult {
	resp, err := c.do("DELETE", fmt.Sprintf("/tribe/agreements/%s", name), ContentTypeJSON, nil)
	if err != nil {
		return &DeleteAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeDeleteAgreementType:
		return &DeleteAgreementResult{resp.Body.(*rbody.TribeDeleteAgreement), nil}
	case rbody.ErrorType:
		return &DeleteAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &DeleteAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) GetAgreement(name string) *GetAgreementResult {
	resp, err := c.do("GET", fmt.Sprintf("/tribe/agreements/%s", name), ContentTypeJSON, nil)
	if err != nil {
		return &GetAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeGetAgreementType:
		return &GetAgreementResult{resp.Body.(*rbody.TribeGetAgreement), nil}
	case rbody.ErrorType:
		return &GetAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &GetAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) JoinAgreement(agreementName, memberName string) *JoinAgreementResult {
	b, err := json.Marshal(struct {
		MemberName string `json:"member_name"`
	}{MemberName: memberName})
	if err != nil {
		return &JoinAgreementResult{Err: err}
	}
	resp, err := c.do("PUT", fmt.Sprintf("/tribe/agreements/%s/join", agreementName), ContentTypeJSON, b)
	if err != nil {
		return &JoinAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeJoinAgreementType:
		return &JoinAgreementResult{resp.Body.(*rbody.TribeJoinAgreement), nil}
	case rbody.ErrorType:
		return &JoinAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &JoinAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

func (c *Client) LeaveAgreement(agreementName, memberName string) *LeaveAgreementResult {
	b, err := json.Marshal(struct {
		MemberName string `json:"member_name"`
	}{MemberName: memberName})
	if err != nil {
		return &LeaveAgreementResult{Err: err}
	}
	resp, err := c.do("DELETE", fmt.Sprintf("/tribe/agreements/%s/leave", agreementName), ContentTypeJSON, b)
	if err != nil {
		return &LeaveAgreementResult{Err: err}
	}
	switch resp.Meta.Type {
	case rbody.TribeLeaveAgreementType:
		return &LeaveAgreementResult{resp.Body.(*rbody.TribeLeaveAgreement), nil}
	case rbody.ErrorType:
		return &LeaveAgreementResult{Err: resp.Body.(*rbody.Error)}
	default:
		return &LeaveAgreementResult{Err: ErrAPIResponseMetaType}
	}
}

type ListMembersResult struct {
	*rbody.TribeMemberList
	Err error
}

type GetMemberResult struct {
	*rbody.TribeMemberShow
	Err error
}

type AddAgreementResult struct {
	*rbody.TribeAddAgreement
	Err error
}

type ListAgreementResult struct {
	*rbody.TribeListAgreement
	Err error
}

type DeleteAgreementResult struct {
	*rbody.TribeDeleteAgreement
	Err error
}

type GetAgreementResult struct {
	*rbody.TribeGetAgreement
	Err error
}

type JoinAgreementResult struct {
	*rbody.TribeJoinAgreement
	Err error
}

type LeaveAgreementResult struct {
	*rbody.TribeLeaveAgreement
	Err error
}
