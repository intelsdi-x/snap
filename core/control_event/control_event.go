/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package control_event

const (
	AvailablePluginDead   = "Control.AvailablePluginDead"
	PluginLoaded          = "Control.PluginLoaded"
	PluginUnloaded        = "Control.PluginUnloaded"
	PluginsSwapped        = "Control.PluginsSwapped"
	PluginSubscribed      = "Control.PluginSubscribed"
	PluginUnsubscribed    = "Control.PluginUnsubscribed"
	ProcessorSubscribed   = "Control.ProcessorSubscribed"
	ProcessorUnsubscribed = "Control.ProcessorUnsubscribed"
	MetricSubscribed      = "Control.MetricSubscribed"
	MetricUnsubscribed    = "Control.MetricUnsubscribed"
	HealthCheckFailed     = "Control.PluginHealthCheckFailed"
	MoveSubscription      = "Control.PluginSubscriptionMoved"
)

type LoadPluginEvent struct {
	Name    string
	Version int
	Type    int
	Signed  bool
}

func (e LoadPluginEvent) Namespace() string {
	return PluginLoaded
}

type UnloadPluginEvent struct {
	Name    string
	Version int
	Type    int
}

func (e UnloadPluginEvent) Namespace() string {
	return PluginUnloaded
}

type DeadAvailablePluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      uint32
	String  string
}

func (e *DeadAvailablePluginEvent) Namespace() string {
	return AvailablePluginDead
}

type SwapPluginsEvent struct {
	LoadedPluginName      string
	LoadedPluginVersion   int
	UnloadedPluginName    string
	UnloadedPluginVersion int
	PluginType            int
}

func (s SwapPluginsEvent) Namespace() string {
	return PluginsSwapped
}

type PluginSubscriptionEvent struct {
	PluginName       string
	PluginVersion    int
	PluginType       int
	SubscriptionType int
	TaskId           uint64
}

func (ps PluginSubscriptionEvent) Namespace() string {
	return PluginSubscribed
}

type PluginUnsubscriptionEvent struct {
	TaskId        uint64
	PluginName    string
	PluginVersion int
	PluginType    int
}

func (pu PluginUnsubscriptionEvent) Namespace() string {
	return PluginUnsubscribed
}

type HealthCheckFailedEvent struct {
	Name    string
	Version int
	Type    int
}

func (hfe HealthCheckFailedEvent) Namespace() string {
	return HealthCheckFailed
}

type MovePluginSubscriptionEvent struct {
	TaskId          uint64
	PluginName      string
	PreviousVersion int
	NewVersion      int
	PluginType      int
}

func (mse MovePluginSubscriptionEvent) Namespace() string {
	return MoveSubscription
}
