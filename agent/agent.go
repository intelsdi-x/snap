package agent

import (
	"github.com/lynxbat/pulse/agent/collection"
	"github.com/lynxbat/pulse/agent/scheduling"
	"github.com/lynxbat/pulse/agent/publishing"
	"fmt"
	"time"
)

// TODO defined by Collector Config
var caching = true
var caching_ttl float64 = 1

// MAC
// Move out
//const UNIX_SOCK = "/usr/local/Cellar/collectd/5.4.1/var/run/collectd-unixsock"

// VM
// Move out
// TODO defined by Collector Config
const UNIX_SOCK = "/var/run/collectd-unixsock"

type CommandResponse struct {
	Message string
	StatusCode int
}

func GetMetricList() []collection.Metric{
	// TODO make all collector plugins explicit
	// CollectD collector
	collectd_coll := collection.NewCollectDCollector(UNIX_SOCK, caching, caching_ttl)
	metrics := collectd_coll.GetMetricList()

	// TODO make all collector plugins explicit
	// Facter collector
	facter_coll := collection.NewFacterCollector("facter", caching, caching_ttl)
	metrics = append(metrics, facter_coll.GetMetricList()...)

	// TODO make all collector plugins explicit
	// Container collector
	container_coll := collection.NewContainerCollector()
	metrics = append(metrics, container_coll.GetMetricList()...)


//	metrics := []collection.Metric{}
	return metrics
}

// TODO convert string to CollectorConfig interface once implemented
func GetMetricValues(string...interface {}) []collection.Metric{
	// Our metric slice
	metrics := []collection.Metric{}
	// TODO collect for each collector provided
	// <>
	// Static for now
	metrics = getFromCollectDCollector(metrics)
//	metrics = getFromFacterCollector(metrics)
	metrics = getFromLibcontainerCollector(metrics)
	//
	return metrics
}

func StartScheduler() {

	metrics := GetMetricValues()
	// Testing scheduler

	// convert to newMetricTask to add error handling on construction

	start := time.Now()
//	stop := time.Now().Add(time.Hour * 24 * 30)

	t := scheduling.MetricTask{
		Label: "Foo",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source": "code debugging",
			"created_by": "nick",
		},
		Metrics: metrics[len(metrics)-3:],
		// start, stop, interval
		Schedule: scheduling.NewSchedule(time.Second * 10, start),
		PublisherConfig: publishing.STDOUTPublishingConfig{},
	}
	fmt.Println(t)
}

func getFromCollectDCollector(metrics []collection.Metric) []collection.Metric{
	c := collection.NewCollectDCollector(UNIX_SOCK, caching, caching_ttl)
	new_metrics := c.GetMetricValues(c.GetMetricList())
	return append(metrics, new_metrics...)
}

func getFromFacterCollector(metrics []collection.Metric) []collection.Metric{
	c := collection.NewFacterCollector("facter", caching, caching_ttl)
	new_metrics := c.GetMetricValues(c.GetMetricValues([]collection.Metric{}))
	return append(metrics, new_metrics...)
}

func getFromLibcontainerCollector(metrics []collection.Metric) []collection.Metric{
	c := collection.NewContainerCollector()
	new_metrics := c.GetMetricValues()
	return append(metrics, new_metrics...)
}




