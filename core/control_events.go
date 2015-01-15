package core

const (
	Unloaded     = "Control.PluginUnloaded"
	Subscribed   = "Control.MetricSubscribed"
	Unsubscribed = "Control.MetricUnsubscribed"
)

type UnloadPluginEvent struct {
}

func (e *UnloadPluginEvent) Namespace() string {
	return Unloaded
}

type SubscriptionEvent struct {
	Count           int
	MetricNamespace []string
}

func (se *SubscriptionEvent) Namespace() string {
	return Subscribed
}

type UnsubscriptionEvent struct {
	Count           int
	MetricNamespace []string
}

func (ue *UnsubscriptionEvent) Namespace() string {
	return Unsubscribed
}
