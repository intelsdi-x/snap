package plugin

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core/ctypes"
)

type PublishArgs struct {
	//PluginMetrics []PluginMetric
	ContentType string
	Content     []byte
	Config      map[string]ctypes.ConfigValue
}

type PublishReply struct {
}

type publisherPluginProxy struct {
	Plugin  PublisherPlugin
	Session Session
}

func (p *publisherPluginProxy) GetConfigPolicy(args GetConfigPolicyArgs, reply *GetConfigPolicyReply) error {
	defer catchPluginPanic(p.Session.Logger())

	p.Session.Logger().Println("GetConfigPolicy called")
	p.Session.ResetHeartbeat()

	reply.Policy = p.Plugin.GetConfigPolicy()

	return nil
}

func (p *publisherPluginProxy) Publish(args PublishArgs, reply *PublishReply) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()
	err := p.Plugin.Publish(args.ContentType, args.Content, args.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Publish call error: %v", err.Error()))
	}
	return nil
}
