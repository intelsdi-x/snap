package core

type UnloadPluginEvent struct {
}

func (e *UnloadPluginEvent) Namespace() string {
	return "Control.Plugin"
}

type SubscriptionEvent struct {
	Count           int
	MetricNamespace []string
}

func (se *SubscriptionEvent) Namespace() string {
	return "Control.MetricSubscribed"
}

type UnsubscriptionEvent struct {
	Count           int
	MetricNamespace []string
}

func (ue *UnsubscriptionEvent) Namespace() string {
	return "Control.MetricUnsubscribed"
}
