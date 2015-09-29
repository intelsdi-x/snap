package plugin

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core/ctypes"
)

type PublishArgs struct {
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

func (p *publisherPluginProxy) Publish(args []byte, reply *[]byte) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()

	dargs := &PublishArgs{}
	err := p.Session.Decode(args, dargs)
	if err != nil {
		return err
	}

	err = p.Plugin.Publish(dargs.ContentType, dargs.Content, dargs.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Publish call error: %v", err.Error()))
	}
	return nil
}
