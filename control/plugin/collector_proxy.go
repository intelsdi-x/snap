package plugin

import (
	"errors"
	"fmt"
)

// Arguments passed to CollectMetrics() for a Collector implementation
type CollectMetricsArgs struct {
	PluginMetricTypes []PluginMetricType
}

//func (c *CollectMetricsArgs) UnmarshalJSON(data []byte) error {
//	pmt := &[]PluginMetricType{}
//	if err := json.Unmarshal(data, pmt); err != nil {
//		return err
//	}
//	c.PluginMetricTypes = *pmt
//	return nil
//}

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

type collectorPluginProxy struct {
	Plugin  CollectorPlugin
	Session Session
}

func (c *collectorPluginProxy) GetMetricTypes(args []byte, reply *[]byte) error {
	defer catchPluginPanic(c.Session.Logger())

	c.Session.Logger().Println("GetMetricTypes called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()

	mts, err := c.Plugin.GetMetricTypes()
	if err != nil {
		return errors.New(fmt.Sprintf("GetMetricTypes call error : %s", err.Error()))
	}

	r := GetMetricTypesReply{PluginMetricTypes: mts}
	*reply, err = c.Session.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

func (c *collectorPluginProxy) CollectMetrics(args []byte, reply *[]byte) error {
	defer catchPluginPanic(c.Session.Logger())
	c.Session.Logger().Println("CollectMetrics called")
	// Reset heartbeat
	c.Session.ResetHeartbeat()

	dargs := &CollectMetricsArgs{}
	c.Session.Decode(args, dargs)

	ms, err := c.Plugin.CollectMetrics(dargs.PluginMetricTypes)
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
