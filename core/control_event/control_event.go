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

package controlevent

const (
	// AvailablePluginDead represents the state of a plugin is dead
	AvailablePluginDead = "Control.AvailablePluginDead"
	// PluginLoaded represents the state of a plugin is loaded
	PluginLoaded = "Control.PluginLoaded"
	// PluginUnloaded represents the state of a plugin is unloaded
	PluginUnloaded = "Control.PluginUnloaded"
	// PluginsSwapped represents the state of a plugin is swapped
	PluginsSwapped = "Control.PluginsSwapped"
	// PluginSubscribed represents the state of a plugin is subscribed
	PluginSubscribed = "Control.PluginSubscribed"
	// PluginUnsubscribed represents the state of a plugin is unsubscribed
	PluginUnsubscribed = "Control.PluginUnsubscribed"
	// ProcessorSubscribed represents the state of a processor is subscribed
	ProcessorSubscribed = "Control.ProcessorSubscribed"
	// ProcessorUnsubscribed represents the state of a processor is unsubscribed
	ProcessorUnsubscribed = "Control.ProcessorUnsubscribed"
	// MetricSubscribed represents the state of a metric is subscribed
	MetricSubscribed = "Control.MetricSubscribed"
	// MetricUnsubscribed represents the state of a metric is unsubscribed
	MetricUnsubscribed = "Control.MetricUnsubscribed"
	// HealthCheckFailed represents the state of a plugin health check failed
	HealthCheckFailed = "Control.PluginHealthCheckFailed"
	// MoveSubscription represents the state of a plugin subscription moved
	MoveSubscription = "Control.PluginSubscriptionMoved"
)

// LoadPluginEvent struct type
// defining the plugin name, version, type
// and if a plugin is signed
type LoadPluginEvent struct {
	Name    string
	Version int
	Type    int
	Signed  bool
}

// Namespace returns PluginLoaded string message
func (e LoadPluginEvent) Namespace() string {
	return PluginLoaded
}

// UnloadPluginEvent struct describing
// unloaded plugin name, version, and type
type UnloadPluginEvent struct {
	Name    string
	Version int
	Type    int
}

// Namespace returns PluginUnloaded message
func (e UnloadPluginEvent) Namespace() string {
	return PluginUnloaded
}

// DeadAvailablePluginEvent struct type describing
// a dead plugin attributes
type DeadAvailablePluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      uint32
	String  string
}

// Namespace returns AvailablePluginDead string message
func (e *DeadAvailablePluginEvent) Namespace() string {
	return AvailablePluginDead
}

// SwapPluginsEvent struct type describing
// the swapped plugin names, versions and type
type SwapPluginsEvent struct {
	LoadedPluginName      string
	LoadedPluginVersion   int
	UnloadedPluginName    string
	UnloadedPluginVersion int
	PluginType            int
}

// Namespace returns PluginSwapped message
// after a swapping plugin event
func (s SwapPluginsEvent) Namespace() string {
	return PluginsSwapped
}

// PluginSubscriptionEvent struct type describing
// a plugin name, version, and other plugin attributes
type PluginSubscriptionEvent struct {
	PluginName       string
	PluginVersion    int
	PluginType       int
	SubscriptionType int
	TaskId           string
}

// Namespace returns the PluginSubscribed message
// after the plugin subscribe event
func (ps PluginSubscriptionEvent) Namespace() string {
	return PluginSubscribed
}

// PluginUnsubscriptionEvent struct type describing
// plugin unsubscribing attributes
type PluginUnsubscriptionEvent struct {
	TaskId        string
	PluginName    string
	PluginVersion int
	PluginType    int
}

// Namespace returns PluginUnsubscribed message
// after plugin unsubscribed event
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
	TaskId          string
	PluginName      string
	PreviousVersion int
	NewVersion      int
	PluginType      int
}

func (mse MovePluginSubscriptionEvent) Namespace() string {
	return MoveSubscription
}
