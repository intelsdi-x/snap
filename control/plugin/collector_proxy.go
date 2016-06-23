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
	"time"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/grpc/common"
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

	c.Session.Logger().Println("GetMetricTypes called")
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
	c.Session.Logger().Println("CollectMetrics called")
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

// collectorPluginGrpc server
type gRPCCollectorProxy struct {
	Plugin  CollectorPlugin
	Session Session
	gRPCPluginProxy
}

func (g *gRPCCollectorProxy) CollectMetrics(ctx context.Context, arg *rpc.CollectMetricsArg) (*rpc.CollectMetricsReply, error) {
	defer catchPluginPanic(g.Session.Logger())

	metrics, err := g.Plugin.CollectMetrics(toPluginMetricTypes(arg.Metrics))
	if err != nil {
		return &rpc.CollectMetricsReply{
			Error: err.Error(),
		}, nil
	}

	coreMetrics := make([]core.Metric, len(metrics))
	idx := 0
	for _, m := range metrics {
		coreMetrics[idx] = m
		idx++
	}

	reply := &rpc.CollectMetricsReply{
		Metrics: common.NewMetrics(coreMetrics),
	}

	return reply, nil
}

func (g *gRPCCollectorProxy) GetMetricTypes(ctx context.Context, arg *rpc.GetMetricTypesArg) (*rpc.GetMetricTypesReply, error) {
	defer catchPluginPanic(g.Session.Logger())

	metricTypes, err := g.Plugin.GetMetricTypes(
		ConfigType{common.ConfigMapToConfig(arg.Config)},
	)
	if err != nil {
		return &rpc.GetMetricTypesReply{
			Error: err.Error(),
		}, nil
	}

	coreMetrics := make([]core.Metric, len(metricTypes))
	idx := 0
	for _, m := range metricTypes {
		coreMetrics[idx] = m
		idx++
	}

	reply := &rpc.GetMetricTypesReply{
		Metrics: common.NewMetrics(coreMetrics),
	}

	return reply, nil
}

// Convert common.Metric to plugin.MetricType
func toPluginMetricTypes(mts []*common.Metric) []MetricType {
	ret := make([]MetricType, len(mts))
	for i, metric := range mts {
		ret[i] = MetricType{
			Namespace_:          common.ToCoreNamespace(metric.Namespace),
			Version_:            int(metric.Version),
			Tags_:               metric.Tags,
			LastAdvertisedTime_: time.Unix(metric.LastAdvertisedTime.Sec, metric.LastAdvertisedTime.Nsec),
			Config_:             common.ConfigMapToConfig(metric.Config),
			Description_:        metric.Description,
			Unit_:               metric.Unit,
		}
		if metric.Timestamp != nil {
			ret[i].Timestamp_ = time.Unix(metric.Timestamp.Sec, metric.Timestamp.Nsec)
		}
		if metric.LastAdvertisedTime != nil {
			ret[i].LastAdvertisedTime_ = time.Unix(metric.LastAdvertisedTime.Sec, metric.LastAdvertisedTime.Nsec)
		}
		if metric.Data != nil {
			switch t := metric.Data.(type) {
			case *common.Metric_StringData:
				ret[i].Data_ = t.StringData
			case *common.Metric_Float32Data:
				ret[i].Data_ = t.Float32Data
			case *common.Metric_Float64Data:
				ret[i].Data_ = t.Float64Data
			case *common.Metric_Int32Data:
				ret[i].Data_ = t.Int32Data
			case *common.Metric_Int64Data:
				ret[i].Data_ = t.Int64Data
			case *common.Metric_BytesData:
				ret[i].Data_ = t.BytesData
			default:
				panic(fmt.Sprintf("unsupported type: %s", t))
			}
		}
	}
	return ret
}
