package slack

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"

	slackApi "github.com/bluele/slack"
)

const (
	name       = "slack"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

type slackPublisher struct {
}

func NewSlackPublisher() *slackPublisher {
	return &slackPublisher{}
}

func (m *slackPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	logger.Println("Publishing started")
	var metrics []plugin.PluginMetricType

	switch contentType {
	case plugin.PulseGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Printf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Printf("Error unknown content type '%v'", contentType)
		return errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}

	logger.Printf("publishing %v to %v", metrics, config)

	name := config["name"].(ctypes.ConfigValueStr).Value
	token := config["token"].(ctypes.ConfigValueStr).Value
	channel := config["token"].(ctypes.ConfigValueStr).Value

	nowTime := time.Now()
	var buffer bytes.Buffer
	for _, m := range metrics {
		buffer.WriteString(fmt.Sprintf("%v|%v|%v\n", nowTime, m.Namespace(), m.Data()))
	}
	err := sendToSlack(name, token, channel, buffer.String())
	if err != nil {
		logger.Printf("Error: %v", err)
		return err
	}

	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

func (m *slackPublisher) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()

	name, err := cpolicy.NewStringRule("name", true)
	handleErr(err)
	name.Description = "Message originator name"
	config.Add(name)

	token, err := cpolicy.NewStringRule("token", true)
	handleErr(err)
	token.Description = "Token used in connection"
	config.Add(token)

	channel, err := cpolicy.NewStringRule("channel", true)
	handleErr(err)
	channel.Description = "Channel name"
	config.Add(channel)

	return *config
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

func sendToSlack(name, token, channel string, body string) error {
	slackClient := slackApi.New(token)
	return slackClient.ChatPostMessage(channel, body, nil)
}
