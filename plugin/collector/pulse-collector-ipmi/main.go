package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-ipmi/ipmi"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-ipmi/ipmiplugin"
)

func main() {

	ipmilayer := &ipmi.LinuxInband{Device: "/dev/ipmi0"}

	ipmiCollector := &ipmiplugin.IpmiCollector{IpmiLayer: ipmilayer,
		Vendor: ipmi.GenericVendor, NSim: 3}

	plugin.Start(plugin.NewPluginMeta(ipmiplugin.Name, ipmiplugin.Version,
		ipmiplugin.Type, []string{}, []string{plugin.PulseGOBContentType},
		plugin.ConcurrencyCount(1)), ipmiCollector, os.Args[1])
}
