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

package rbody

import "github.com/intelsdi-x/snap/mgmt/tribe/agreement"

// const A list of tribe constants
const (
	TribeListAgreementType   = "tribe_agreement_list_returned"
	TribeGetAgreementType    = "tribe_agreement_returned"
	TribeAddAgreementType    = "tribe_agreement_created"
	TribeDeleteAgreementType = "tribe_agreement_deleted"
	TribeAddMemberType       = "tribe_member_added"
	TribeJoinAgreementType   = "tribe_agreement_joined"
	TribeLeaveAgreementType  = "tribe_agreement_left"
	TribeMemberListType      = "tribe_member_list_returned"
	TribeMemberShowType      = "tribe_member_details_returned"
)

// TribeAddAgreement a map of agreement struct type
type TribeAddAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

// ResponseBodyMessage The tribe agreement response
func (t *TribeAddAgreement) ResponseBodyMessage() string {
	return "Tribe agreement created"
}

// ResponseBodyType The tribe response body type
func (t *TribeAddAgreement) ResponseBodyType() string {
	return TribeAddAgreementType
}

// TribeGetAgreement struct type
type TribeGetAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

// ResponseBodyMessage The tribe agreement response
func (t *TribeGetAgreement) ResponseBodyMessage() string {
	return "Tribe agreement returned"
}

// ResponseBodyType returns  a string representation of
// tribe get agreement response body type
func (t *TribeGetAgreement) ResponseBodyType() string {
	return TribeGetAgreementType
}

// TribeDeleteAgreement struct type
type TribeDeleteAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

// ResponseBodyMessage returns a string response
func (t *TribeDeleteAgreement) ResponseBodyMessage() string {
	return "Tribe agreement deleted"
}

// ResponseBodyType returns a string representation of tribe
// delete agreement type
func (t *TribeDeleteAgreement) ResponseBodyType() string {
	return TribeDeleteAgreementType
}

// TribeListAgreement struct type
type TribeListAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

// ResponseBodyMessage returns a string response
func (t *TribeListAgreement) ResponseBodyMessage() string {
	return "Tribe agreements retrieved"
}

// ResponseBodyType returns a string of TribeListAgreementType
func (t *TribeListAgreement) ResponseBodyType() string {
	return TribeListAgreementType
}

// TribeJoinAgreement struct type
type TribeJoinAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

// ResponseBodyMessage returns a string response
func (t *TribeJoinAgreement) ResponseBodyMessage() string {
	return "Tribe agreement joined"
}

// ResponseBodyType returns a string represetation of TribeJoinAgreementType
func (t *TribeJoinAgreement) ResponseBodyType() string {
	return TribeJoinAgreementType
}

// TribeLeaveAgreement struct type
type TribeLeaveAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

// ResponseBodyMessage returns  a string response
func (t *TribeLeaveAgreement) ResponseBodyMessage() string {
	return "Tribe agreement left"
}

// ResponseBodyType returns a string response of TribeLeaveAgreementType
func (t *TribeLeaveAgreement) ResponseBodyType() string {
	return TribeLeaveAgreementType
}

// TribeMemberList struct type
type TribeMemberList struct {
	Members []string `json:"members"`
}

// ResponseBodyMessage returns a string response
func (t *TribeMemberList) ResponseBodyMessage() string {
	return "Tribe members retrieved"
}

// ResponseBodyType returns a string response of TribeMemberListType
func (t *TribeMemberList) ResponseBodyType() string {
	return TribeMemberListType
}

// TribeMemberShow struct type
type TribeMemberShow struct {
	Name            string            `json:"name"`
	PluginAgreement string            `json:"plugin_agreement"`
	Tags            map[string]string `json:"tags"`
	TaskAgreements  []string          `json:"task_agreements"`
}

// ResponseBodyMessage returns a string response
func (t *TribeMemberShow) ResponseBodyMessage() string {
	return "Tribe member details retrieved"
}

// ResponseBodyType returns a string response of TribeMemberShowType
func (t *TribeMemberShow) ResponseBodyType() string {
	return TribeMemberShowType
}
