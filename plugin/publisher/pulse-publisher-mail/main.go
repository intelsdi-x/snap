package main

import (
	"os"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/publisher/pulse-publisher-mail/mail"
)

func main() {
	meta := mail.Meta()
	plugin.Start(meta, mail.NewMailPublisher(), os.Args[1])
}
