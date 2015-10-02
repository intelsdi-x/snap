package rest

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/core/perror"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/julienschmidt/httprouter"
)

var (
	tribeLogger = restLogger.WithFields(log.Fields{
		"_module": "rest-tribe",
	})

	ErrInvalidJSON           = errors.New("Invalid JSON")
	ErrAgreementDoesNotExist = errors.New("Agreement not found")
	ErrMemberNotFound        = errors.New("Member not found")
)

func (s *Server) getAgreements(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	res := &rbody.TribeAgreementList{}
	res.Agreements = s.tr.GetAgreements()
	respond(200, res, w)
}

func (s *Server) getAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "joinAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromPulseError(perror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}
	a := &rbody.TribeGetAgreement{}
	var perr perror.PulseError
	a.Agreement, perr = s.tr.GetAgreement(name)
	if perr != nil {
		tribeLogger.Error(perr)
		respond(400, rbody.FromPulseError(perr), w)
		return
	}
	respond(200, a, w)
}

func (s *Server) joinAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "joinAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromPulseError(perror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		tribeLogger.Error(err)
		respond(500, rbody.FromError(err), w)
		return
	}

	m := struct {
		MemberName string `json:"member_name"`
	}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"hint":  `The body of the request should be of the form '{"member_name": "some_value"}'`,
		}
		pe := perror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}

	perr := s.tr.JoinAgreement(name, m.MemberName)
	if perr != nil {
		tribeLogger.Error(perr)
		respond(400, rbody.FromPulseError(perr), w)
		return
	}
	respond(200, &rbody.TribeJoinAgreement{}, w)

}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	members := s.tr.GetMembers()
	respond(200, &rbody.TribeMemberList{Members: members}, w)
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "getMember")
	name := p.ByName("name")
	member := s.tr.GetMember(name)
	if member == nil {
		fields := map[string]interface{}{
			"name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrMemberNotFound)
		respond(404, rbody.FromPulseError(perror.New(ErrMemberNotFound, fields)), w)
		return
	}
	resp := &rbody.TribeMemberShow{
		Name: member.Name,
		Tags: member.Tags,
	}
	if member.PluginAgreement != nil {
		resp.PluginAgreement = member.PluginAgreement.Name
	}
	for k, _ := range member.TaskAgreements {
		resp.TaskAgreements = append(resp.TaskAgreements, k)
	}
	respond(200, resp, w)
}

func (s *Server) addAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "addAgreement")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		tribeLogger.Error(err)
		respond(500, rbody.FromError(err), w)
		return
	}

	a := struct{ Name string }{}
	err = json.Unmarshal(b, &a)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"hint":  `The body of the request should be of the form '{"name": "some_value"}'`,
		}
		pe := perror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}

	if a.Name == "" {
		fields := map[string]interface{}{
			"hint": `The body of the request should be of the form '{"name": "some_value"}'`,
		}
		pe := perror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromPulseError(pe), w)
		return
	}

	err = s.tr.AddAgreement(a.Name)
	if err != nil {
		tribeLogger.WithField("agreement-name", a.Name).Error(err)
		respond(400, rbody.FromError(err), w)
		return
	}

	respond(200, &rbody.TribeAddAgreement{Name: a.Name}, w)
}
