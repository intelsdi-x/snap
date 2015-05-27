package core

type PluginType int

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcessorPluginType
)

type Plugin interface {
	Name() string
	Version() int
}
