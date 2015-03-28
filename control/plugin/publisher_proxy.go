package plugin

import (
	"errors"
	"fmt"
)

// Arguments passed to PublishMetricsArgs() for a Publisher implementation
type PublishArgs struct {
	Data []byte
}

type PublishReply struct {
}

type publisherPluginProxy struct {
	Plugin  PublisherPlugin
	Session Session
}

func (p *publisherPluginProxy) Publish(args PublishArgs, reply *PublishReply) error {
	p.Session.Logger().Println("Publish called")
	// Reset heartbeat
	p.Session.ResetHeartbeat()
	err := p.Plugin.Publish(args.Data)
	if err != nil {
		return errors.New(fmt.Sprintf("PublishMetrics call error : %s", err.Error()))
	}
	return nil
}
