package control_event

const (
	PluginLoaded        = "Control.PluginLoaded"
	PluginDisabled      = "Control.PluginDisabled"
	PluginUnloaded      = "Control.PluginUnloaded"
	PluginsSwapped      = "Control.PluginsSwapped"
	PublisherSubscribed = "Control.PublisherSubscribed"
	ProcessorSubscribed = "Control.ProcessorSubscribed"
	MetricSubscribed    = "Control.MetricSubscribed"
	MetricUnsubscribed  = "Control.MetricUnsubscribed"
	HealthCheckFailed   = "Control.PluginHealthCheckFailed"
)

type LoadPluginEvent struct{}

func (e LoadPluginEvent) Namespace() string {
	return PluginLoaded
}

type UnloadPluginEvent struct {
}

func (e UnloadPluginEvent) Namespace() string {
	return PluginUnloaded
}

type DisabledPluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Index   int
}

func (e *DisabledPluginEvent) Namespace() string {
	return PluginDisabled
}

type SwapPluginsEvent struct{}

func (s SwapPluginsEvent) Namespace() string {
	return PluginsSwapped
}

type PublisherSubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (se PublisherSubscriptionEvent) Namespace() string {
	return PublisherSubscribed
}

type ProcessorSubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (se ProcessorSubscriptionEvent) Namespace() string {
	return ProcessorSubscribed
}

type MetricSubscriptionEvent struct {
	MetricNamespace []string
	Version         int
}

func (se MetricSubscriptionEvent) Namespace() string {
	return MetricSubscribed
}

type MetricUnsubscriptionEvent struct {
	MetricNamespace []string
}

func (ue MetricUnsubscriptionEvent) Namespace() string {
	return MetricUnsubscribed
}

type HealthCheckFailedEvent struct {
	Name    string
	Version int
	Type    int
}

func (hfe HealthCheckFailedEvent) Namespace() string {
	return HealthCheckFailed
}
