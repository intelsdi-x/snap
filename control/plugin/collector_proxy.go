package plugin

import (
	"errors"
	"fmt"
)

// Arguments passed to CollectMetrics() for a Collector implementation
type CollectMetricsArgs struct {
	MetricTypes []MetricType
}

// Reply assigned by a Collector implementation using CollectMetrics()
type CollectMetricsReply struct {
	Metrics []Metric
}

// GetMetricTypesArgs args passed to GetMetricTypes
type GetMetricTypesArgs struct {
}

// GetMetricTypesReply assigned by GetMetricTypes() implementation
type GetMetricTypesReply struct {
	MetricTypes []MetricType
}

type collectorPluginProxy struct {
	Plugin  CollectorPlugin
	Session Session
}

func (c *collectorPluginProxy) GetMetricTypes(args GetMetricTypesArgs, reply *GetMetricTypesReply) error {
	c.Session.Logger().Println("GetMetricTypes called")
	mts, err := c.Plugin.GetMetricTypes()
	if err != nil {
		return errors.New(fmt.Sprintf("GetMetricTypes call error : %s", err.Error()))
	}
	reply.MetricTypes = mts
	return nil
}

func (c *collectorPluginProxy) CollectMetrics(args CollectMetricsArgs, reply *CollectMetricsReply) error {
	c.Session.Logger().Println("CollectMetrics called")
	ms, err := c.Plugin.CollectMetrics(args.MetricTypes)
	if err != nil {
		return errors.New(fmt.Sprintf("CollectMetrics call error : %s", err.Error()))
	}
	reply.Metrics = ms
	return nil
}
