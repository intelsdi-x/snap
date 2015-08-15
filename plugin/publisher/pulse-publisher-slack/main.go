package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-slack/slack"
)

func main() {
	meta := slack.Meta()
	plugin.Start(meta, slack.NewSlackPublisher(), os.Args[1])
}
