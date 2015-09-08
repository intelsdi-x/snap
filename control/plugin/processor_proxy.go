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

func (p *processorPluginProxy) GetConfigPolicy(args GetConfigPolicyArgs, reply *GetConfigPolicyReply) error {
	defer catchPluginPanic(p.Session.Logger())

	p.Session.Logger().Println("GetConfigPolicy called")
	p.Session.ResetHeartbeat()

	reply.Policy = p.Plugin.GetConfigPolicy()

	return nil
}

func (p *processorPluginProxy) Process(args ProcessorArgs, reply *ProcessorReply) error {
	defer catchPluginPanic(p.Session.Logger())
	p.Session.ResetHeartbeat()
	var err error
	reply.ContentType, reply.Content, err = p.Plugin.Process(args.ContentType, args.Content, args.Config)
	if err != nil {
		return errors.New(fmt.Sprintf("Processor call error: %v", err.Error()))
	}
	return nil
}
