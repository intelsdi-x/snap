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

func UnmarshalBody(t string, b []byte) (Body, error) {
	switch t {
	case PluginListReturnedType:
		return unmarshalAndHandleError(b, &PluginListReturned{})
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
	case MetricReturnedType:
		return unmarshalAndHandleError(b, &MetricReturned{})
	case MetricCatalogReturnedType:
		return unmarshalAndHandleError(b, &MetricCatalogReturned{})
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
