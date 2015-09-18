package rbody

import "github.com/intelsdi-x/pulse/mgmt/tribe"

const (
	TribeAgreementListType = "tribe_agreement_list_returned"
	TribeGetAgreementType  = "tribe_agreement_returned"
	TribeAddAgreementType  = "tribe_agreement_created"
	TribeAddMemberType     = "tribe_member_added"
	TribeJoinAgreementType = "tribe_agreement_joined"
	TribeMemberListType    = "tribe_member_list_returned"
	TribeMemberShowType    = "tribe_member_details_returned"
)

type TribeAddAgreement struct {
	Name string
}

func (t *TribeAddAgreement) ResponseBodyMessage() string {
	return "Tribe agreement created"
}

func (t *TribeAddAgreement) ResponseBodyType() string {
	return TribeAddAgreementType
}

type TribeGetAgreement struct {
	Agreement *tribe.Agreement `json:"agreement"`
}

func (t *TribeGetAgreement) ResponseBodyMessage() string {
	return "Tribe agreement returned"
}

func (t *TribeGetAgreement) ResponseBodyType() string {
	return TribeGetAgreementType
}

type TribeAgreementList struct {
	Agreements map[string]*tribe.Agreement `json:"agreements"`
}

func (t *TribeAgreementList) ResponseBodyMessage() string {
	return "Tribe agreements retrieved"
}

func (t *TribeAgreementList) ResponseBodyType() string {
	return TribeAgreementListType
}

type TribeJoinAgreement struct {
	Agreement *tribe.Agreement `json:"agreement"`
}

func (t *TribeJoinAgreement) ResponseBodyMessage() string {
	return "Tribe agreement joined"
}

func (t *TribeJoinAgreement) ResponseBodyType() string {
	return TribeJoinAgreementType
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
