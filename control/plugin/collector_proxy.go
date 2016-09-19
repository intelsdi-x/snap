/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package plugin

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/snap/core/cdata"
)

// Arguments passed to CollectMetrics() for a Collector implementation
type CollectMetricsArgs struct {
	MetricTypes []MetricType
}

// Reply assigned by a Collector implementation using CollectMetrics()
type CollectMetricsReply struct {
	PluginMetrics []MetricType
}

// GetMetricTypesArgs args passed to GetMetricTypes
type GetMetricTypesArgs struct {
	PluginConfig ConfigType
}

// GetMetricTypesReply assigned by GetMetricTypes() implementation
type GetMetricTypesReply struct {
	MetricTypes []MetricType
}

type collectorPluginProxy struct {
	Plugin  CollectorPlugin
	Session Session
}

func (c *collectorPluginProxy) GetMetricTypes(args []byte, reply *[]byte) error {
	defer catchPluginPanic(c.Session.Logger())

	c.Session.Logger().Debugln("GetMetricTypes called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()

	dargs := &GetMetricTypesArgs{PluginConfig: ConfigType{ConfigDataNode: cdata.NewNode()}}
	c.Session.Decode(args, dargs)

	mts, err := c.Plugin.GetMetricTypes(dargs.PluginConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("GetMetricTypes call error : %s", err.Error()))
	}

	r := GetMetricTypesReply{MetricTypes: mts}
	*reply, err = c.Session.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

func (c *collectorPluginProxy) CollectMetrics(args []byte, reply *[]byte) error {
	defer catchPluginPanic(c.Session.Logger())
	c.Session.Logger().Debugln("CollectMetrics called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()

	dargs := &CollectMetricsArgs{}
	c.Session.Decode(args, dargs)

	ms, err := c.Plugin.CollectMetrics(dargs.MetricTypes)
	if err != nil {
		return errors.New(fmt.Sprintf("CollectMetrics call error : %s", err.Error()))
	}

	r := CollectMetricsReply{PluginMetrics: ms}
	*reply, err = c.Session.Encode(r)
	if err != nil {
		return err
	}
	return nil
}
