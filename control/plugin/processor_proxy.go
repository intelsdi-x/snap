package plugin

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/core/ctypes"
)

type ProcessorArgs struct {
	//PluginMetrics []PluginMetric
	ContentType string
	Content     []byte
	Config      map[string]ctypes.ConfigValue
}

type ProcessorReply struct {
	ContentType string
	Content     []byte
}

type processorPluginProxy struct {
	Plugin  ProcessorPlugin
	Session Session
}

func (p *processorPluginProxy) Process(args []byte, reply *[]byte) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()

	dargs := &ProcessorArgs{}
	err := p.Session.Decode(args, dargs)
	if err != nil {
		return err
	}

	r := ProcessorReply{}
	r.ContentType, r.Content, err = p.Plugin.Process(dargs.ContentType, dargs.Content, dargs.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Processor call error: %v", err.Error()))
	}

	*reply, err = p.Session.Encode(r)
	if err != nil {
		return err
	}

	return nil
}
