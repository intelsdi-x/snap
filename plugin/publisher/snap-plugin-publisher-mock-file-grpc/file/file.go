/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	debug = "debug"
)

type filePublisher struct {
}

func NewFilePublisher() *filePublisher {
	return &filePublisher{}
}

func (f *filePublisher) Publish(metrics []plugin.Metric, config plugin.Config) error {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	if _, err := config.GetBool(debug); err == nil {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Debug("publishing started")

	filename, err := config.GetString("file")
	if err != nil {
		log.Error(err)
		return err
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		log.Error(err)
		return err
	}
	log.WithFields(log.Fields{
		"file": filename,
		"metrics-published-count": len(metrics),
	}).Debug("metrics published")
	w := bufio.NewWriter(file)
	for _, m := range metrics {
		formattedTags := formatMetricTagsAsString(m.Tags)
		w.WriteString(fmt.Sprintf("%v|%v|%v|%v\n", m.Timestamp, m.Namespace, m.Data, formattedTags))
	}
	w.Flush()

	return nil
}

func (f *filePublisher) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	err := policy.AddNewStringRule([]string{""}, "file", true)
	if err != nil {
		return *policy, err
	}

	err = policy.AddNewBoolRule([]string{}, debug, false)

	return *policy, err
}

// formatMetricTagsAsString returns metric's tags as a string in the following format tagKey:tagValue where the next tags are separated by semicolon
func formatMetricTagsAsString(metricTags map[string]string) string {
	var tags string
	for tag, value := range metricTags {
		tags += fmt.Sprintf("%s:%s; ", tag, value)
	}
	// trim the last semicolon
	tags = strings.TrimSuffix(tags, "; ")

	return "tags[" + tags + "]"
}
