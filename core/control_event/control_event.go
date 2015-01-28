package control_event

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	PluginLoaded       = "Control.PluginLoaded"
	PluginDisabled     = "Control.PluginDisabled"
	PluginUnloaded     = "Control.PluginUnloaded"
	PluginsSwapped     = "Control.PluginsSwapped"
	MetricSubscribed   = "Control.MetricSubscribed"
	MetricUnsubscribed = "Control.MetricUnsubscribed"
	HealthCheckFailed  = "Control.PluginHealthCheckFailed"
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

type DisabledPluginEvent struct {
	Type plugin.PluginType
	Meta plugin.PluginMeta
}

func (e *DisabledPluginEvent) Namespace() string {
	return PluginDisabled
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

type HealthCheckFailedEvent struct {
	Type plugin.PluginType
	Meta plugin.PluginMeta
}

func (hfe *HealthCheckFailedEvent) Namespace() string {
	return HealthCheckFailed
}
