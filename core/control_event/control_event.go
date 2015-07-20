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
