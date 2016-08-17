/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	AvailablePluginDead      = "Control.AvailablePluginDead"
	AvailablePluginRestarted = "Control.RestartedAvailablePlugin"
	PluginRestartsExceeded   = "Control.PluginRestartsExceeded"
	PluginStarted            = "Control.PluginStarted"
	PluginLoaded             = "Control.PluginLoaded"
	PluginUnloaded           = "Control.PluginUnloaded"
	PluginsSwapped           = "Control.PluginsSwapped"
	PluginSubscribed         = "Control.PluginSubscribed"
	PluginUnsubscribed       = "Control.PluginUnsubscribed"
	ProcessorSubscribed      = "Control.ProcessorSubscribed"
	ProcessorUnsubscribed    = "Control.ProcessorUnsubscribed"
	MetricSubscribed         = "Control.MetricSubscribed"
	MetricUnsubscribed       = "Control.MetricUnsubscribed"
	HealthCheckFailed        = "Control.PluginHealthCheckFailed"
	MoveSubscription         = "Control.PluginSubscriptionMoved"
)

type StartPluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      uint32
}

func (e StartPluginEvent) Namespace() string {
	return PluginStarted
}

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

type RestartedAvailablePluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      uint32
}

func (e *MaxPluginRestartsExceededEvent) Namespace() string {
	return PluginRestartsExceeded
}

type MaxPluginRestartsExceededEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      uint32
}

func (e *RestartedAvailablePluginEvent) Namespace() string {
	return AvailablePluginRestarted
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
	PluginName    string
	PluginVersion int
	PluginType    int
	TaskId        string
}

func (ps PluginSubscriptionEvent) Namespace() string {
	return PluginSubscribed
}

type PluginUnsubscriptionEvent struct {
	TaskId        string
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
