package rbody

import (
	"encoding/json"
	"errors"
)

type Body interface {
	// These function names are rather verbose to avoid field vs function namespace collisions
	// with varied object types that use them.
	ResponseBodyMessage() string
	ResponseBodyType() string
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
	case TribeAgreementListType:
		return unmarshalAndHandleError(b, &TribeAgreementList{})
	case TribeAddAgreementType:
		return unmarshalAndHandleError(b, &TribeAddAgreement{})
	case TribeMemberShowType:
		return unmarshalAndHandleError(b, &TribeMemberShow{})
	case TribeJoinAgreementType:
		return unmarshalAndHandleError(b, &TribeJoinAgreement{})
	case TribeGetAgreementType:
		return unmarshalAndHandleError(b, &TribeGetAgreement{})
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
