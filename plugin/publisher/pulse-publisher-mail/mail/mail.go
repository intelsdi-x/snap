package mail

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net/smtp"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	name       = "mail"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

type mailPublisher struct {
}

func NewMailPublisher() *mailPublisher {
	return &mailPublisher{}
}

func (m *mailPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
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

	username := config["username"].(ctypes.ConfigValueStr).Value
	password := config["password"].(ctypes.ConfigValueStr).Value
	sender := config["sender address"].(ctypes.ConfigValueStr).Value
	to := config["recipient address"].(ctypes.ConfigValueStr).Value
	host := config["server address"].(ctypes.ConfigValueStr).Value
	port := config["server port"].(ctypes.ConfigValueInt).Value
	subject := config["subject"].(ctypes.ConfigValueStr).Value

	nowTime := time.Now()
	var buffer bytes.Buffer
	for _, m := range metrics {
		buffer.WriteString(fmt.Sprintf("%v|%v|%v\n", nowTime, m.Namespace(), m.Data()))
	}
	err := sendMail(username, password, sender, to, host, port, subject, buffer.String())
	if err != nil {
		logger.Printf("Error: %v", err)
		return err
	}

	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

func (m *mailPublisher) GetConfigPolicyNode() cpolicy.ConfigPolicyNode {
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("username", true)
	handleErr(err)
	r1.Description = "User name used in smtp connection"
	config.Add(r1)

	r2, err := cpolicy.NewStringRule("password", true)
	handleErr(err)
	r2.Description = "Password used in smtp connection"
	config.Add(r2)

	r3, err := cpolicy.NewStringRule("sender address", true)
	handleErr(err)
	r3.Description = "Mail address to set as sender address"
	config.Add(r3)

	r4, err := cpolicy.NewStringRule("recipient address", true)
	handleErr(err)
	r4.Description = "Recipient address"
	config.Add(r4)

	r5, err := cpolicy.NewStringRule("server address", true, "smtp.gmail.com")
	handleErr(err)
	r5.Description = "SMTP server address to use (defualt: smtp.gmail.com)"
	config.Add(r5)

	r6, err := cpolicy.NewIntegerRule("server port", true, 587)
	handleErr(err)
	r6.Description = "SMTP server port to use (defualt: 587)"
	config.Add(r6)

	r7, err := cpolicy.NewStringRule("subject", true, "[Pulse metrics]")
	handleErr(err)
	r7.Description = "Mail subject (default: [Pulse metrics])"
	config.Add(r7)

	return *config
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

func sendMail(username, password, sender, to, host string, port int, subject, body string) error {
	auth := smtp.PlainAuth(
		"",
		username,
		password,
		host,
	)

	address := fmt.Sprintf("%v:%v", host, port)

	data := []byte("Subject: " + subject + "\r\n\r\n" + body)

	return smtp.SendMail(
		address,
		auth,
		sender,
		[]string{to},
		data,
	)
}
