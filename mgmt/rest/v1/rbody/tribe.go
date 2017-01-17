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

type TribeAddAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

func (t *TribeAddAgreement) ResponseBodyMessage() string {
	return "Tribe agreement created"
}

func (t *TribeAddAgreement) ResponseBodyType() string {
	return TribeAddAgreementType
}

type TribeGetAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

func (t *TribeGetAgreement) ResponseBodyMessage() string {
	return "Tribe agreement returned"
}

func (t *TribeGetAgreement) ResponseBodyType() string {
	return TribeGetAgreementType
}

type TribeDeleteAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

func (t *TribeDeleteAgreement) ResponseBodyMessage() string {
	return "Tribe agreement deleted"
}

func (t *TribeDeleteAgreement) ResponseBodyType() string {
	return TribeDeleteAgreementType
}

type TribeListAgreement struct {
	Agreements map[string]*agreement.Agreement `json:"agreements"`
}

func (t *TribeListAgreement) ResponseBodyMessage() string {
	return "Tribe agreements retrieved"
}

func (t *TribeListAgreement) ResponseBodyType() string {
	return TribeListAgreementType
}

type TribeJoinAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

func (t *TribeJoinAgreement) ResponseBodyMessage() string {
	return "Tribe agreement joined"
}

func (t *TribeJoinAgreement) ResponseBodyType() string {
	return TribeJoinAgreementType
}

type TribeLeaveAgreement struct {
	Agreement *agreement.Agreement `json:"agreement"`
}

func (t *TribeLeaveAgreement) ResponseBodyMessage() string {
	return "Tribe agreement left"
}

func (t *TribeLeaveAgreement) ResponseBodyType() string {
	return TribeLeaveAgreementType
}

type TribeMemberList struct {
	Members []string `json:"members"`
}

func (t *TribeMemberList) ResponseBodyMessage() string {
	return "Tribe members retrieved"
}

func (t *TribeMemberList) ResponseBodyType() string {
	return TribeMemberListType
}

type TribeMemberShow struct {
	Name            string            `json:"name"`
	PluginAgreement string            `json:"plugin_agreement"`
	Tags            map[string]string `json:"tags"`
	TaskAgreements  []string          `json:"task_agreements"`
}

func (t *TribeMemberShow) ResponseBodyMessage() string {
	return "Tribe member details retrieved"
}

func (t *TribeMemberShow) ResponseBodyType() string {
	return TribeMemberShowType
}
