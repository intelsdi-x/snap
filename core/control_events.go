package core

type UnloadPluginEvent struct {
}

func (e *UnloadPluginEvent) Namespace() string {
	return "Control.Plugin"
}
