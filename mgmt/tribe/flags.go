package tribe

import (
	"math/rand"
	"os"

	"github.com/codegangsta/cli"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	flTribeNodeName = cli.StringFlag{
		Name:   "tribe-node-name",
		Usage:  "Name of this node in tribe cluster (default: hostname)",
		EnvVar: "PULSE_TRIBE_NODE_NAME",
		Value:  getHostname(),
	}

	flTribeSeed = cli.StringFlag{
		Name: "tribe",
		Usage: `IP or resolvable hostname of a seed node to join.
	The default empty value assumes this is the first node in a cluster.`,
		EnvVar: "PULSE_TRIBE_SEED",
		Value:  "",
	}

	flTribeAdvertisePort = cli.IntFlag{
		Name:   "tribe-port",
		Usage:  "Port tribe gossips over to maintain membership",
		EnvVar: "PULSE_TRIBE_PORT",
		Value:  6000,
	}

	Flags = []cli.Flag{flTribeNodeName, flTribeSeed, flTribeAdvertisePort}
)

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return randHostname(8)
	}
	return hostname
}

func randHostname(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
