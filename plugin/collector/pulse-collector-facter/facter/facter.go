package facter

import (
	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	Name    = "facter"
	Version = 1
)

// Facter collector
//
// Attempts to call facter and convert metrics into plugins
type Facter struct {
}

func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	return nil
}

func ConfigPolicy() *plugin.ConfigPolicy {
	c := new(plugin.ConfigPolicy)
	return c
}
