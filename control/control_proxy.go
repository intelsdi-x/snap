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

package control

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/internal/common"
	"golang.org/x/net/context"
)

type controlProxy struct {
	control *pluginControl
}

func (pc *controlProxy) Load(ctx context.Context, arg *rpc.PluginRequest) (*rpc.PluginReply, error) {
	//Write the file to local disk
	var localPath string
	var err error
	if localPath, err = writeFile(arg.Name, arg.PluginFile); err != nil {
		return nil, err
	}
	rp, err := core.NewRequestedPlugin(localPath)
	if err != nil {
		return nil, err
	}
	//Verify checksum
	var checkSum [32]byte
	copy(checkSum[:], arg.CheckSum)
	if rp.CheckSum() != checkSum {
		return nil, errors.New("Checksum mismatch on requested plugin when loading")
	}
	rp.SetSignature(arg.Signature)
	pl, err := pc.control.Load(rp)
	if err != nil {
		err2 := os.RemoveAll(filepath.Dir(rp.Path()))
		if err2 != nil {
			controlLogger.Error("Unable to remove: ", filepath.Dir(rp.Path()))
		}
		return nil, err
	}
	return catalogPluginToPluginReply(pl)
}

func (pc *controlProxy) MetricCatalog(ctx context.Context, _ *common.Empty) (*rpc.MetricCatalogReply, error) {
	mets, err := pc.control.MetricCatalog()
	if err != nil {
		return nil, err
	}
	return catalogMetricsToReply(mets)
}

func (pc *controlProxy) FetchMetrics(ctx context.Context, arg *rpc.FetchMetricsRequest) (*rpc.MetricCatalogReply, error) {
	mets, err := pc.control.FetchMetrics(arg.Namespace, int(arg.Version))
	if err != nil {
		return nil, err
	}
	return catalogMetricsToReply(mets)
}

func (pc *controlProxy) GetMetricVersions(ctx context.Context, arg *rpc.GetMetricVersionsRequest) (*rpc.MetricCatalogReply, error) {
	mets, err := pc.control.GetMetricVersions(arg.Namespace)
	if err != nil {
		return nil, err
	}
	return catalogMetricsToReply(mets)
}

func (pc *controlProxy) GetMetric(ctx context.Context, arg *rpc.FetchMetricsRequest) (*rpc.MetricReply, error) {
	mets, err := pc.control.GetMetric(arg.Namespace, int(arg.Version))
	if err != nil {
		return nil, err
	}
	return catalogMetricToMetricReply(mets)
}

func (pc *controlProxy) Unload(ctx context.Context, arg *rpc.UnloadPluginRequest) (*rpc.PluginReply, error) {
	pl, err := pc.control.Unload(rpc.NewCatalogedPlugin(arg.Name, int(arg.Version), arg.PluginType))
	if err != nil {
		return nil, err
	}
	return catalogPluginToPluginReply(pl)
}

func (pc *controlProxy) PluginCatalog(ctx context.Context, _ *common.Empty) (*rpc.PluginCatalogReply, error) {
	plugins := pc.control.PluginCatalog()
	return catalogPluginsToReply(plugins)
}

func (pc *controlProxy) GetPlugin(ctx context.Context, arg *rpc.GetPluginRequest) (*rpc.GetPluginReply, error) {
	lp, err := pc.control.pluginManager.Get(arg.Name, int(arg.Version), arg.Type)
	if err != nil {
		return nil, err
	}
	reply := &rpc.GetPluginReply{}
	if arg.Download {
		b, err := ioutil.ReadFile(lp.PluginPath())
		if err != nil {
			return nil, err
		}
		reply.PluginBytes = b
	}
	reply.Plugin, err = catalogPluginToPluginReply(lp)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (pc *controlProxy) AvailablePlugins(ctx context.Context, _ *common.Empty) (*rpc.AvailablePluginsReply, error) {
	aPlugins := pc.control.AvailablePlugins()
	reply := availablePluginToReply(aPlugins)
	return reply, nil
}

//--------Utility functions--------------------------------

func catalogMetricsToReply(mets []core.CatalogedMetric) (*rpc.MetricCatalogReply, error) {
	result := make([]*rpc.MetricReply, 0, len(mets))
	for _, met := range mets {
		m, err := catalogMetricToMetricReply(met)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return &rpc.MetricCatalogReply{Metrics: result}, nil
}

func catalogMetricToMetricReply(met core.CatalogedMetric) (*rpc.MetricReply, error) {
	var err error
	metric := &rpc.MetricReply{
		Namespace:          met.Namespace(),
		Version:            int64(met.Version()),
		LastAdvertisedTime: timeToTimeReply(met.LastAdvertisedTime()),
	}
	metric.ConfigPolicy, err = json.Marshal(met.Policy())
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func catalogPluginsToReply(plugins []core.CatalogedPlugin) (*rpc.PluginCatalogReply, error) {
	result := make([]*rpc.PluginReply, 0, len(plugins))
	for _, pl := range plugins {
		p, err := catalogPluginToPluginReply(pl)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return &rpc.PluginCatalogReply{Plugins: result}, nil
}

func catalogPluginToPluginReply(pl core.CatalogedPlugin) (*rpc.PluginReply, error) {
	cp, err := json.Marshal(pl.Policy())
	if err != nil {
		return nil, err
	}
	return &rpc.PluginReply{
		Name:            pl.Name(),
		Version:         int64(pl.Version()),
		TypeName:        pl.TypeName(),
		IsSigned:        pl.IsSigned(),
		Status:          pl.Status(),
		LoadedTimestamp: timeToTimeReply(*pl.LoadedTimestamp()),
		ConfigPolicy:    cp,
	}, nil
}

func timeToTimeReply(t time.Time) *common.Time {
	return &common.Time{Sec: t.Unix(), Nsec: int64(t.Nanosecond())}
}

func availablePluginToReply(plugins []core.AvailablePlugin) *rpc.AvailablePluginsReply {
	result := make([]*rpc.AvailablePluginReply, 0, len(plugins))
	for _, pl := range plugins {
		p := availablePluginToPluginReply(pl)
		result = append(result, p)
	}
	return &rpc.AvailablePluginsReply{Plugins: result}
}

func availablePluginToPluginReply(pl core.AvailablePlugin) *rpc.AvailablePluginReply {
	return &rpc.AvailablePluginReply{
		Name:             pl.Name(),
		Version:          int64(pl.Version()),
		TypeName:         pl.TypeName(),
		HitCount:         int64(pl.HitCount()),
		ID:               pl.ID(),
		LastHitTimestamp: timeToTimeReply(pl.LastHit()),
	}
}

func writeFile(filename string, b []byte) (string, error) {
	// Create temporary directory
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	f, err := os.Create(path.Join(dir, filename))
	if err != nil {
		return "", err
	}
	n, err := f.Write(b)
	if err != nil {
		return "", err
	}
	log.Debugf("wrote %v to %v", n, f.Name())
	if runtime.GOOS != "windows" {
		err = f.Chmod(0700)
		if err != nil {
			return "", err
		}
	}
	// Close before load
	f.Close()
	return f.Name(), nil
}
