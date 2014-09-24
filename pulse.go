package pulse

import (
	"fmt"
	"github.com/intelsdi/pulse/collection"
	"github.com/intelsdi/pulse/publishing"
	"github.com/intelsdi/pulse/scheduling"
	"os"
	"os/signal"
	"syscall"
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
	Message    string
	StatusCode int
}

func GetMetricList() []collection.Metric {
	// TODO make all collector plugins explicit
	// CollectD collector
	collectdColl := collection.NewCollectorByType("collectd", collection.NewCollectDConfig(UNIX_SOCK))
	metrics := collectdColl.GetMetricList()

	// TODO make all collector plugins explicit
	// Facter collector

	facterColl := collection.NewCollectorByType("facter", collection.NewFacterConfig("facter"))
	metrics = append(metrics, facterColl.GetMetricList()...)

	// TODO make all collector plugins explicit
	// Container collector
	container_coll := collection.NewCollectorByType("container", collection.NewContainerConfig())
	metrics = append(metrics, container_coll.GetMetricList()...)

	//	metrics := []collection.Metric{}
	return metrics
}

// TODO convert string to CollectorConfig interface once implemented
func GetMetricValues(string ...interface{}) []collection.Metric {
	// Our metric slice
	metrics := []collection.Metric{}
	// TODO collect for each collector provided
	// <>
	// Static for now
	metrics = getFromCollectDCollector(metrics)
	metrics = getFromFacterCollector(metrics)
	metrics = getFromLibcontainerCollector(metrics)
	//
	return metrics
}

func StartScheduler(initWorkerCount int) {

	metrics := GetMetricValues()
	t1 := scheduling.MetricTask{
		Label: "Foo",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source":     "code debugging",
			"created_by": "nick",
		},
		CollectorConfigs: map[string]collection.CollectorConfig{},
		Metrics:          metrics,
		Schedule:         scheduling.NewSchedule(time.Millisecond * 500),
		PublisherConfig:  publishing.STDOUTPublishingConfig{},
	}
	t2 := scheduling.MetricTask{
		Label: "Bar",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source":     "code debugging",
			"created_by": "nick",
		},
		CollectorConfigs: map[string]collection.CollectorConfig{},
		Metrics:          metrics,
		Schedule:         scheduling.NewSchedule(time.Second*1, time.Now().Add(time.Second*5), time.Now().Add(time.Second*300)),
		PublisherConfig:  publishing.STDOUTPublishingConfig{},
	}
	t3 := scheduling.MetricTask{
		Label: "Baz",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source":     "code debugging",
			"created_by": "nick",
		},
		CollectorConfigs: map[string]collection.CollectorConfig{},
		Metrics:          metrics,
		Schedule:         scheduling.NewSchedule(time.Second*1, time.Now().Add(time.Second*10), time.Now().Add(time.Second*300)),
		PublisherConfig:  publishing.STDOUTPublishingConfig{},
	}
	t4 := scheduling.MetricTask{
		Label: "Qux",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source":     "code debugging",
			"created_by": "nick",
		},
		CollectorConfigs: map[string]collection.CollectorConfig{},
		Metrics:          metrics,
		Schedule:         scheduling.NewSchedule(time.Second*1, time.Now().Add(time.Second*15), time.Now().Add(time.Second*300)),
		PublisherConfig:  publishing.STDOUTPublishingConfig{},
	}
	t5 := scheduling.MetricTask{
		Label: "Quux",
		Metadata: map[string]string{
			"created_at": time.Now().Format("2006/01/02 15:04:05"),
			"source":     "code debugging",
			"created_by": "nick",
		},
		CollectorConfigs: map[string]collection.CollectorConfig{},
		Metrics:          metrics,
		Schedule:         scheduling.NewSchedule(time.Second*1, time.Now().Add(time.Second*30), time.Now().Add(time.Second*300)),
		PublisherConfig:  publishing.STDOUTPublishingConfig{},
	}

	scheduler := scheduling.NewScheduler(initWorkerCount)
	// Add tasks to scheduler
	scheduler.MetricTasks = []*scheduling.MetricTask{&t1, &t2, &t3, &t4, &t5}

	// Starts scheduler, this is a nonblocking method. So you must provide a way to block until you want to cleanup using Stop()
	err := scheduler.Start()
	// Defer stop
	defer scheduler.Stop()

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	killChannel := make(chan bool)
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			killChannel <- true
		case syscall.SIGTERM:
			killChannel <- true
		}
	}()

	// Blocks and waits for kill
	// TODO move higher than the agent. CLI control preferred.
	<-killChannel
}

func getFromCollectDCollector(metrics []collection.Metric) []collection.Metric {
	collectdColl := collection.NewCollectorByType("collectd", collection.NewCollectDConfig(UNIX_SOCK))
	newMetrics := collectdColl.GetMetricList()
	return append(metrics, newMetrics...)
}

func getFromFacterCollector(metrics []collection.Metric) []collection.Metric {
	facterColl := collection.NewCollectorByType("facter", collection.NewFacterConfig("facter"))
	newMetrics := append(metrics, facterColl.GetMetricList()...)
	return append(metrics, newMetrics...)
}

func getFromLibcontainerCollector(metrics []collection.Metric) []collection.Metric {
	container_coll := collection.NewCollectorByType("container", collection.NewContainerConfig())
	newMetrics := append(metrics, container_coll.GetMetricList()...)
	return append(metrics, newMetrics...)
}
