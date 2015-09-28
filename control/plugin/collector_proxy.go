/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
)

// Arguments passed to CollectMetrics() for a Collector implementation
type CollectMetricsArgs struct {
	PluginMetricTypes []PluginMetricType
}

func (c *CollectMetricsArgs) UnmarshalJSON(data []byte) error {
	pmt := &[]PluginMetricType{}
	if err := json.Unmarshal(data, pmt); err != nil {
		return err
	}
	c.PluginMetricTypes = *pmt
	return nil
}

// Reply assigned by a Collector implementation using CollectMetrics()
type CollectMetricsReply struct {
	PluginMetrics []PluginMetricType
}

// GetMetricTypesArgs args passed to GetMetricTypes
type GetMetricTypesArgs struct {
}

// GetMetricTypesReply assigned by GetMetricTypes() implementation
type GetMetricTypesReply struct {
	PluginMetricTypes []PluginMetricType
}

type GetConfigPolicyArgs struct{}

type GetConfigPolicyReply struct {
	Policy cpolicy.ConfigPolicy
}

type collectorPluginProxy struct {
	Plugin  CollectorPlugin
	Session Session
}

func (c *collectorPluginProxy) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
	defer catchPluginPanic(c.Session.Logger())

	c.Session.Logger().Println("GetMetricTypes called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()
	mts, err := c.Plugin.GetMetricTypes()
	if err != nil {
		return errors.New(fmt.Sprintf("GetMetricTypes call error : %s", err.Error()))
	}
	reply.PluginMetricTypes = mts
	return nil
}

func (c *collectorPluginProxy) CollectMetrics(args CollectMetricsArgs, reply *CollectMetricsReply) error {
	defer catchPluginPanic(c.Session.Logger())
	c.Session.Logger().Println("CollectMetrics called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()
	ms, err := c.Plugin.CollectMetrics(args.PluginMetricTypes)
	if err != nil {
		return errors.New(fmt.Sprintf("CollectMetrics call error : %s", err.Error()))
	}
	reply.PluginMetrics = ms
	return nil
}

func (c *collectorPluginProxy) GetConfigPolicy(args GetConfigPolicyArgs, reply *GetConfigPolicyReply) error {
	defer catchPluginPanic(c.Session.Logger())

	c.Session.Logger().Println("GetConfigPolicy called")
	policy, err := c.Plugin.GetConfigPolicy()

	if err != nil {
		return errors.New(fmt.Sprintf("GetConfigPolicy call error : %s", err.Error()))
	}
	reply.Policy = policy
	return nil
}
