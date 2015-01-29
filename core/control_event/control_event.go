package control_event

const (
	PluginLoaded       = "Control.PluginLoaded"
	PluginUnloaded     = "Control.PluginUnloaded"
	PluginsSwapped     = "Control.PluginsSwapped"
	MetricSubscribed   = "Control.MetricSubscribed"
	MetricUnsubscribed = "Control.MetricUnsubscribed"
)

type LoadPluginEvent struct{}

func (e *LoadPluginEvent) Namespace() string {
	return PluginLoaded
}

type UnloadPluginEvent struct {
}

func (e *UnloadPluginEvent) Namespace() string {
	return PluginUnloaded
}

type SwapPluginsEvent struct{}

func (s *SwapPluginsEvent) Namespace() string {
	return PluginsSwapped
}

type MetricSubscriptionEvent struct {
	MetricNamespace []string
}

func (se *MetricSubscriptionEvent) Namespace() string {
	return MetricSubscribed
}

type MetricUnsubscriptionEvent struct {
	MetricNamespace []string
}

func (ue *MetricUnsubscriptionEvent) Namespace() string {
	return MetricUnsubscribed
}
