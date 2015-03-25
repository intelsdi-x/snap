package plugin

import (
	"errors"
	"fmt"
)

// Arguments passed to PublishMetricsArgs() for a Publisher implementation
type PublishMetricsArgs struct {
	PluginMetrics []PluginMetric
}

type publisherPluginProxy struct {
	Plugin  PublisherPlugin
	Session Session
}

func (p *publisherPluginProxy) PublishMetrics(args PublishMetricsArgs) error {
	p.Session.Logger().Println("Publish called")
	// Reset heartbeat
	p.Session.ResetHeartbeat()
	err := p.Plugin.PublishMetrics(args.PluginMetrics)
	if err != nil {
		return errors.New(fmt.Sprintf("PublishMetrics call error : %s", err.Error()))
	}
	return nil
}
