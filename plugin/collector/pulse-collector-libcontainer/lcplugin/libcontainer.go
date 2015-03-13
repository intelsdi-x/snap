package lcplugin

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

/*******************
 *  pulse plugin  *
 *******************/

const (
	name   = "Intel libcontainer Plugin (c) 2015 Intel Corporation"
	vendor = "intel"
	prefix = "libcontainer"

	version                     = 1
	pluginType                  = plugin.CollectorPluginType
	DefaultCacheTTL             = 60 * time.Second
	DefaultMetricTypesTTL       = DefaultCacheTTL
	DefautlLibcontainerDeadline = 5 * time.Second

	defaultDockerFolder = "/var/lib/docker/execdriver/native/"
)

// TODO those should go to some plugin helper file maybe?
const nsSeparator string = "/"

const (
	common string = "common"
	state  string = "state"
	config string = "config"
	net    string = "net"
	cpu    string = "cpu"
	blkio  string = "blkio"
	memory string = "memory"
)

/*****************************************
 *  pulse public methods implementation  *
 *****************************************/

// returns PluginMeta
func Meta() *plugin.PluginMeta {
	//	var statsy libcontainer.Stats.CgroupStats
	return plugin.NewPluginMeta(name, version, pluginType)
}

// returns ConfigPolicy
func ConfigPolicy() *plugin.ConfigPolicy {
	//TODO What is plugin policy?

	c := new(plugin.ConfigPolicy)
	return c
}

/*************
 *  facter   *
 *************/

// wrapper to use facter binnary, responsible for running external binnary and caching
type Libcontainer struct {
	metricTypes          []*plugin.MetricType //map[string]interface{}
	metricTypesLastCheck time.Time
	metricTypesTTL       time.Duration

	cacheTTL   time.Duration
	cache      map[string]metric
	cacheMutex sync.Mutex

	dockerFolder   string
	lcExecDeadline time.Duration // libcontainer execution deadline
}

type cacheBucket struct {
	namespace []string
	metrics   map[string]metric
}

func NewLibcontainerPlugin() *Libcontainer {
	l := new(Libcontainer)
	//TODO read from config
	l.dockerFolder = defaultDockerFolder
	l.cacheTTL = DefaultCacheTTL
	//	l.cache = make(map[string]metric)
	l.lcExecDeadline = DefautlLibcontainerDeadline
	return l
}

type metric struct {
	value      interface{}
	lastUpdate time.Time
}

func newMetric(value interface{}, lastUpdate time.Time) metric {
	var m metric
	m.value = value
	m.lastUpdate = lastUpdate
	return m
}

func applyBucketsToCache(cache map[string]metric, buckets []cacheBucket) map[string]metric {
	for _, bucket := range buckets {
		for key, metric := range bucket.metrics {
			ns := strings.Join(append(bucket.namespace, key), nsSeparator)
			cache[ns] = metric
		}
	}

	return cache
}

// TODO this will only get one folder to search (then renamed), and specific buckets to fill \
// 		so we can order refresh only for single bucket of one container
func getAllMetrics(dockerFolder string) (map[string]metric, error) {

	folders, err := ioutil.ReadDir(dockerFolder)
	if err != nil {
		log.Println("Libcontainer: Cannot read container folders in path: " + dockerFolder)
		return nil, err
	}

	contIds := make([]string, 0, 20)
	metricBckts := make([]cacheBucket, 0, 60)

	// TODO this should run in parallel goroutines
	for _, folder := range folders {
		if folder.IsDir() {
			// grab all info about container
			containerName := folder.Name()
			containerPath := filepath.Join(dockerFolder, containerName)
			config, state, stats, err := getContainerInfo(containerPath)
			if err == nil {
				// TODO this should run in parallel goroutines
				netBucket := getNetMetrics(containerName, stats)
				stateBucket := getStateMetrics(containerName, state)
				confBucket := getConfigMetrics(containerName, config)

				metricBckts = append(metricBckts, netBucket, stateBucket, confBucket)
				contIds = append(contIds, containerName)

			} else {
				log.Printf("Libcontainer: Error while obtaining container info: %s | path: %s\n", err.Error(), containerPath)
			}
		}
	}

	// add metrics common to all containers
	commonMetrics := map[string]metric{}
	now := time.Now()

	//TODO gathering common metrics should go to different function
	commonMetrics["count"] = newMetric(len(contIds), now)
	commonMetrics["container_ids"] = newMetric(strings.Join(contIds, ";"), now)
	metricBckts = append(metricBckts, cacheBucket{namespace: []string{common}, metrics: commonMetrics})

	// merge all buckets into cache
	retCache := applyBucketsToCache(map[string]metric{}, metricBckts)

	return retCache, nil
}

//TODO this should be common to all plugins, so maybe it should go to some plugin helper file
func prepareMetricTypes(metrics map[string]metric, prefix []string) ([]*plugin.MetricType, time.Time) {
	metricTypes := make([]*plugin.MetricType, 0, len(metrics))
	var timestamp time.Time
	for key, value := range metrics {
		timestamp = value.lastUpdate
		metricTypes = append(metricTypes,
			plugin.NewMetricType(append(prefix,
				strings.Split(key, nsSeparator)...),
				timestamp.Unix()))
	}

	return metricTypes, timestamp
}

/******************************************
 *  Pulse plugin interface implmentation  *
 ******************************************/

// get available metrics types
func (lc *Libcontainer) GetMetricTypes(_ plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	if time.Since(lc.metricTypesLastCheck) > lc.metricTypesTTL {
		metrics, err := getAllMetrics(lc.dockerFolder)
		if err != nil {
			log.Printf("libcontainer: error while gathering metrics: %s in path %s",
				err.Error(), lc.dockerFolder)
			return err
		}
		lc.cacheMutex.Lock()
		lc.cache = metrics
		lc.cacheMutex.Unlock()

		metricPrefix := []string{vendor, prefix}
		metricTypes, lastCheck := prepareMetricTypes(metrics, metricPrefix)

		lc.metricTypes = metricTypes
		lc.metricTypesLastCheck = lastCheck

		reply.MetricTypes = metricTypes

		return nil
	} else {
		reply.MetricTypes = lc.metricTypes
		return nil
	}
}

// collect metrics
func (l *Libcontainer) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	//	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	//	// waits for lynxbat/SDI-98

	//	// TODO: INPUT: read CollectorArgs structure to extrac requested metrics
	//	requestedNames := []string{"kernel", "uptime"}

	//	// prepare cache to have all we need
	//	err := f.synchronizeCache(requestedNames)
	//	if err != nil {
	//		return err
	//	}

	//	// TODO: OUTPUT: fullfill reply structure with requested metrics
	//	// for _, name := range requestedNames {
	//	// 	// convert it some how if required
	//	// 	reply.metrics[name] = f.cache[name].value
	//	// }

	return nil
}
