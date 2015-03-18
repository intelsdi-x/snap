package lcplugin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
)

/******************************************
 *  Pulse plugin interface implmentation  *
 ******************************************/

// returns PluginMeta
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType)
}

func ConfigPolicyTree() *cpolicy.ConfigPolicyTree {
	c := cpolicy.NewTree()
	p := cpolicy.NewPolicyNode()
	c.Add([]string{"intel", "libcontainer"}, p)
	return c
}

// get available metrics types
func (lc *libcntr) GetMetricTypes() ([]plugin.PluginMetricType, error) {

	if time.Since(lc.metricTypesLastCheck) > lc.metricTypesTTL {
		metrics, err := getAllMetrics(lc.dockerFolder)
		if err != nil {
			log.Printf("libcontainer: error while gathering metrics: %s in path %s",
				err.Error(), lc.dockerFolder)
			return nil, err
		}
		lc.cacheMutex.Lock()
		lc.cache = metrics
		lc.cacheMutex.Unlock()

		metricPrefix := []string{vendor, prefix}
		metricTypes, lastCheck := prepareMetricTypes(metrics, metricPrefix)

		lc.metricTypes = metricTypes
		lc.metricTypesLastCheck = lastCheck

		return lc.metricTypes, nil
	} else {
		return lc.metricTypes, nil
	}
}

// collect metrics
func (lc *libcntr) CollectMetrics(reqMetrics []plugin.PluginMetricType) ([]plugin.PluginMetric, error) {

	retVals := make([]plugin.PluginMetric, 0, len(reqMetrics))
	cacheKeys := make([]string, 0, len(reqMetrics))     //- list of metrics that will be read from cache
	staleKeys := make([]([]string), 0, len(reqMetrics)) //- list of metric that will be refreshed
	now := time.Now()
	for _, mt := range reqMetrics {

		//- validation
		if mt.Version() != version {
			errs := fmt.Sprintf("Libcontainer: invalid metric version: %d, LC version is %d\n",
				mt.Version(), version)
			log.Printf(errs)
			return nil, errors.New(errs)
		}

		ns := mt.Namespace()
		if ns[0] != vendor || ns[1] != prefix {
			errs := fmt.Sprintf("Libcontainer: invalid metric signature: %s/%s, needs %s/%s\n",
				ns[0], ns[1], vendor, prefix)
			log.Printf(errs)
			return nil, errors.New(errs)
		}

		//- check for stale entries
		key := strings.Join(ns, nsSeparator)
		cacheKeys = append(cacheKeys, key)
		metric, exists := lc.cache[key]
		if exists {
			cacheDeadline := metric.lastUpdate.Add(lc.cacheTTL)
			if cacheDeadline.Before(now) {
				//- cache timeout, add metric to be refreshed
				staleKeys = append(staleKeys, ns)
			}
		} else {
			//- no entry in cache (it should be), add to staleKeys
			staleKeys = append(staleKeys, ns)
		}
	}

	//- Refresh proper cache entries
	if len(staleKeys) > 0 {
		//TODO refresh only buckets that need refreshing
		//TODO parallel and with barrier
		_, err := lc.GetMetricTypes()
		if err != nil {
			log.Printf("libcontainer: could not refresh metrics")
			return nil, err
		}
	}

	for _, key := range cacheKeys {
		metric, exists := lc.cache[key]
		if exists {
			retVals = append(retVals,
				plugin.PluginMetric{Namespace_: metric.namespace,
					Data_: metric.value})
		} else {
			// append nil metric to cache for caching purposes
			ns := strings.Split(key, nsSeparator)
			lc.cache[key] = newMetric(nil, time.Now(), ns)
			retVals = append(retVals,
				plugin.PluginMetric{Namespace_: ns,
					Data_: nil})

		}
	}

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

	return retVals, nil
}

func NewLibCntr() *libcntr {
	l := new(libcntr)
	//TODO read from config
	l.dockerFolder = defaultDockerFolder
	l.cacheTTL = defaultCacheTTL
	l.cache = make(map[string]metric)
	l.lcExecDeadline = defautlLibcontainerDeadline
	return l
}

/****************/
/* Private part */
/****************/

const (
	name   = "Intel libcontainer Plugin (c) 2015 Intel Corporation"
	vendor = "intel"
	prefix = "libcontainer"

	version                     = 1
	pluginType                  = plugin.CollectorPluginType
	defaultCacheTTL             = 60 * time.Second
	defaultMetricTypesTTL       = defaultCacheTTL
	defautlLibcontainerDeadline = 5 * time.Second

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

/********************************************
 *  libcntr private methods implementation  *
 ********************************************/

type libcntr struct {
	metricTypes          []plugin.PluginMetricType //map[string]interface{}
	metricTypesLastCheck time.Time
	metricTypesTTL       time.Duration

	cacheTTL   time.Duration
	cache      map[string]metric
	cacheMutex sync.Mutex

	dockerFolder   string
	lcExecDeadline time.Duration // libcontainer execution deadline
}

type metric struct {
	namespace  []string // full namespace, with vendor + prefix + ... + metric_name
	value      interface{}
	lastUpdate time.Time
}

func newMetric(value interface{}, lastUpdate time.Time, namespace []string) metric {
	var m metric
	m.value = value
	m.lastUpdate = lastUpdate
	m.namespace = namespace
	return m
}

type cacheBucket struct {
	namespace []string
	metrics   map[string]metric
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
	var timestamp time.Time
	for _, folder := range folders {
		if folder.IsDir() {
			// grab all info about container
			containerName := folder.Name()
			containerPath := filepath.Join(dockerFolder, containerName)
			config, state, stats, err := getContainerInfo(containerPath)
			timestamp = time.Now()
			if err == nil {
				// TODO this should run in parallel goroutines
				netBucket := getNetMetrics(containerName, stats, timestamp)
				stateBucket := getStateMetrics(containerName, state, timestamp)
				confBucket := getConfigMetrics(containerName, config, timestamp)

				metricBckts = append(metricBckts, netBucket, stateBucket, confBucket)
				contIds = append(contIds, containerName)

			} else {
				log.Printf("Libcontainer: Error while obtaining container info: %s | path: %s\n", err.Error(), containerPath)
			}
		}
	}

	// add metrics common to all containers
	ns := []string{vendor, prefix, common}
	commonMetrics := map[string]metric{}

	//TODO gathering common metrics should go to different function
	commonMetrics["count"] = newMetric(len(contIds), timestamp, append(ns, "count"))
	commonMetrics["container_ids"] = newMetric(strings.Join(contIds, ";"), timestamp, append(ns, "container_ids"))
	metricBckts = append(metricBckts, cacheBucket{namespace: ns, metrics: commonMetrics})

	// merge all buckets into cache
	retCache := applyBucketsToCache(map[string]metric{}, metricBckts)

	return retCache, nil
}

//TODO this should be common to all plugins, so maybe it should go to some plugin helper file
func prepareMetricTypes(metrics map[string]metric, prefix []string) ([]plugin.PluginMetricType, time.Time) {
	metricTypes := make([]plugin.PluginMetricType, 0, len(metrics))
	var timestamp time.Time
	for _, value := range metrics {
		timestamp = value.lastUpdate
		metricTypes = append(metricTypes,
			*plugin.NewPluginMetricType(value.namespace))
	}

	return metricTypes, timestamp
}
