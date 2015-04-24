package plugin

import (
	"errors"
	"fmt"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

// Arguments passed to CollectMetrics() for a Collector implementation
type CollectMetricsArgs struct {
	PluginMetricTypes []PluginMetricType
}

// Reply assigned by a Collector implementation using CollectMetrics()
type CollectMetricsReply struct {
	PluginMetrics []PluginMetric
}

// GetMetricTypesArgs args passed to GetMetricTypes
type GetMetricTypesArgs struct {
}

// GetMetricTypesReply assigned by GetMetricTypes() implementation
type GetMetricTypesReply struct {
	PluginMetricTypes []PluginMetricType
}

type GetConfigPolicyTreeArgs struct{}

type GetConfigPolicyTreeReply struct {
	PolicyTree cpolicy.ConfigPolicyTree
}

type collectorPluginProxy struct {
	Plugin  CollectorPlugin
	Session Session
}

func (c *collectorPluginProxy) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
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

func (c *collectorPluginProxy) GetConfigPolicyTree(args GetConfigPolicyTreeArgs, reply *GetConfigPolicyTreeReply) error {
	c.Session.Logger().Println("GetConfigPolicyTree called")
	policy, err := c.Plugin.GetConfigPolicyTree()

	if err != nil {
		return errors.New(fmt.Sprintf("ConfigPolicyTree call error : %s", err.Error()))
	}
	reply.PolicyTree = policy
	return nil
}
