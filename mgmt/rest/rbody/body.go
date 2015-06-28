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
		p := &PluginListReturned{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case PluginsLoadedType:
		p := &PluginsLoaded{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case PluginUnloadedType:
		p := &PluginUnloaded{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ScheduledTaskListReturnedType:
		p := &ScheduledTaskListReturned{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ScheduledTaskType:
		p := &ScheduledTask{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case AddScheduledTaskType:
		p := &AddScheduledTask{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ScheduledTaskStartedType:
		p := &ScheduledTaskStarted{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ScheduledTaskStoppedType:
		p := &ScheduledTaskStopped{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ScheduledTaskRemovedType:
		p := &ScheduledTaskRemoved{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case MetricCatalogReturnedType:
		p := &MetricCatalogReturned{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	case ErrorType:
		p := &Error{}
		err := json.Unmarshal(b, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, ErrCannotUnmarshalBody
	}
}
