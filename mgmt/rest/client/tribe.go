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

	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
)

// ListMembers retrieves a list of tribe members through an HTTP GET call.
// A list of tribe member returns if it succeeds. Otherwise, an error is returned.
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

// GetMember retrieves the tribe member given a member name.
// The request is an HTTP GET call.  The corresponding tribe member object returns
// if it succeeds. Otherwise, an error is returned.
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

// ListAgreements retrieves a list of a tribe agreements through an HTTP GET call.
// A list of tribe agreement map returns if it succeeds. Otherwise, an error is returned.
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

// AddAgreement adds a tribe agreement giving an agreement name into tribe agreement list
// through an HTTP POST call. A map of tribe agreements with the newly added named agreement
// returns if it succeeds. Otherwise, an error is returned. Note that the newly added agreement
// has no effect unless members join the agreement.
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

// DeleteAgreement removes a tribe agreement giving an agreement name from the tribe agreement list
// through an HTTP DELETE call. A map of tribe agreements with the specified agreement removed returns
// if it succeeds. Otherwise, an error is returned. Note deleting an agreement removes the agreement
// from the tribe entirely for all the members of the agreement.
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

// GetAgreement retrieves a tribe agreement given an agreement name through an HTTP GET call.
// A tribe agreement returns if it succeeded. Otherwise, an error is returned.
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

// JoinAgreement adds a tribe member into the agreement given the agreement name and the member name.
// It is an HTTP PUT request. The agreement with the newly added member returns if it succeeds.
// Otherwise, an error is returned. Note that dual directional agreement replication happens automatically
// through the gossip protocol between a newly joined member and existing members within the same agreement.
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

// LeaveAgreement removes a member from the agreement given the agreement and member names through
// an HTTP DELETE call. The agreement with the removed member returns if it succeeds.
// Otherwise, an error is returned. For example, it is useful to leave an agreement for a member node repair.
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

// ListMembersResult is the response from snap/client on a ListMembers call.
type ListMembersResult struct {
	*rbody.TribeMemberList
	Err error
}

// GetMemberResult is the response from snap/client on a GetMember call.
type GetMemberResult struct {
	*rbody.TribeMemberShow
	Err error
}

// AddAgreementResult is the response from snap/client on a AddAgreement call.
type AddAgreementResult struct {
	*rbody.TribeAddAgreement
	Err error
}

// ListAgreementResult is the response from snap/client on a ListAgreements call.
type ListAgreementResult struct {
	*rbody.TribeListAgreement
	Err error
}

// DeleteAgreementResult is the response from snap/client on a DeleteAgreement call.
type DeleteAgreementResult struct {
	*rbody.TribeDeleteAgreement
	Err error
}

// GetAgreementResult is the response from snap/client on a GetAgreement call.
type GetAgreementResult struct {
	*rbody.TribeGetAgreement
	Err error
}

// JoinAgreementResult is the response from snap/client on a JoinAgreement call.
type JoinAgreementResult struct {
	*rbody.TribeJoinAgreement
	Err error
}

// LeaveAgreementResult is the response from snap/client on a LeaveAgreement call.
type LeaveAgreementResult struct {
	*rbody.TribeLeaveAgreement
	Err error
}
