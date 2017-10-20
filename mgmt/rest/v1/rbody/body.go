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

import (
	"encoding/json"
	"errors"

	"bytes"
	"fmt"
	"net/http"

	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type Body interface {
	// These function names are rather verbose to avoid field vs function namespace collisions
	// with varied object types that use them.
	ResponseBodyMessage() string
	ResponseBodyType() string
}

func Write(code int, b Body, w http.ResponseWriter) {
	w.Header().Set("Deprecated", "true")
	resp := &APIResponse{
		Meta: &APIResponseMeta{
			Code:    code,
			Message: b.ResponseBodyMessage(),
			Type:    b.ResponseBodyType(),
			Version: 1,
		},
		Body: b,
	}
	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

	j, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		logrus.Fatalln(err)
	}
	j = bytes.Replace(j, []byte("\\u0026"), []byte("&"), -1)
	fmt.Fprint(w, string(j))
}

var (
	ErrCannotUnmarshalBody = errors.New("Cannot unmarshal body: invalid type")
)

type APIResponse struct {
	Meta         *APIResponseMeta `json:"meta"`
	Body         Body             `json:"body"`
	JSONResponse string           `json:"-"`
}

type apiResponseJSON struct {
	Meta *APIResponseMeta `json:"meta"`
	Body json.RawMessage  `json:"body"`
}

type APIResponseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Version int    `json:"version"`
}

func (a *APIResponse) UnmarshalJSON(b []byte) error {
	ar := &apiResponseJSON{}
	err := json.Unmarshal(b, ar)
	if err != nil {
		panic(err)
	}
	if ar.Meta == nil {
		return errors.New("Unable to parse JSON response")
	}
	body, err := UnmarshalBody(ar.Meta.Type, ar.Body)
	if err != nil {
		return err
	}
	// Assign
	a.Meta = ar.Meta
	a.Body = body
	return nil
}

func UnmarshalBody(t string, b []byte) (Body, error) {
	switch t {
	case PluginListType:
		return unmarshalAndHandleError(b, &PluginList{})
	case PluginsLoadedType:
		return unmarshalAndHandleError(b, &PluginsLoaded{})
	case PluginUnloadedType:
		return unmarshalAndHandleError(b, &PluginUnloaded{})
	case PluginReturnedType:
		return unmarshalAndHandleError(b, &PluginReturned{})
	case ScheduledTaskListReturnedType:
		return unmarshalAndHandleError(b, &ScheduledTaskListReturned{})
	case ScheduledTaskReturnedType:
		return unmarshalAndHandleError(b, &ScheduledTaskReturned{})
	case ScheduledTaskType:
		return unmarshalAndHandleError(b, &ScheduledTask{})
	case AddScheduledTaskType:
		return unmarshalAndHandleError(b, &AddScheduledTask{})
	case ScheduledTaskStartedType:
		return unmarshalAndHandleError(b, &ScheduledTaskStarted{})
	case ScheduledTaskStoppedType:
		return unmarshalAndHandleError(b, &ScheduledTaskStopped{})
	case ScheduledTaskRemovedType:
		return unmarshalAndHandleError(b, &ScheduledTaskRemoved{})
	case ScheduledTaskEnabledType:
		return unmarshalAndHandleError(b, &ScheduledTaskEnabled{})
	case MetricReturnedType:
		return unmarshalAndHandleError(b, &MetricReturned{})
	case MetricsReturnedType:
		return unmarshalAndHandleError(b, &MetricsReturned{})
	case ScheduledTaskWatchingEndedType:
		return unmarshalAndHandleError(b, &ScheduledTaskWatchingEnded{})
	case TribeMemberListType:
		return unmarshalAndHandleError(b, &TribeMemberList{})
	case TribeListAgreementType:
		return unmarshalAndHandleError(b, &TribeListAgreement{})
	case TribeAddAgreementType:
		return unmarshalAndHandleError(b, &TribeAddAgreement{})
	case TribeDeleteAgreementType:
		return unmarshalAndHandleError(b, &TribeDeleteAgreement{})
	case TribeMemberShowType:
		return unmarshalAndHandleError(b, &TribeMemberShow{})
	case TribeJoinAgreementType:
		return unmarshalAndHandleError(b, &TribeJoinAgreement{})
	case TribeLeaveAgreementType:
		return unmarshalAndHandleError(b, &TribeLeaveAgreement{})
	case TribeGetAgreementType:
		return unmarshalAndHandleError(b, &TribeGetAgreement{})
	case PluginConfigItemType:
		return unmarshalAndHandleError(b, &PluginConfigItem{*cdata.NewNode()})
	case SetPluginConfigItemType:
		return unmarshalAndHandleError(b, &SetPluginConfigItem{*cdata.NewNode()})
	case DeletePluginConfigItemType:
		return unmarshalAndHandleError(b, &DeletePluginConfigItem{*cdata.NewNode()})
	case ErrorType:
		return unmarshalAndHandleError(b, &Error{})
	default:
		return nil, ErrCannotUnmarshalBody
	}
}

func unmarshalAndHandleError(b []byte, body Body) (Body, error) {
	err := json.Unmarshal(b, body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
