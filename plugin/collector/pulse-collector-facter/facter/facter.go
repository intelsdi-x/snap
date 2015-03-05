/*
# testing
go test github.com/intelsdilabs/pulse/plugin/collector/pulse-collector-facter/facter
*/
package facter

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/intelsdilabs/pulse/plugin/helper"
)

/*******************
 *  pulse plugin  *
 *******************/

var (
	namespace = []string{"intel", "facter"}
	Name      = GetPluginName(&namespace) //preprocessor needed
)

const (
	//	Name    = "facter" //should it be intel/facter ?
	Version         = 1
	Type            = plugin.CollectorPluginType
	DefaultCacheTTL = 60 * time.Second
)

type Facter struct {
	availableMetricTypes *[]*plugin.MetricType //map[string]interface{}

	cacheTimestamp time.Time
	cacheTTL       time.Duration
}

// fullfill the availableMetricTypes with data from facter
func (f *Facter) loadAvailableMetricTypes() error {

	facterMap, err := getFacts()
	if err != nil {
		log.Fatalln("getting facts fatal error:", err)
		return err
	}

	avaibleMetrics := make([]*plugin.MetricType, 0, len(*facterMap))
	for key := range *facterMap {
		avaibleMetrics = append(
			avaibleMetrics,
			plugin.NewMetricType(
				append(namespace, key),
				f.cacheTimestamp))
	}

	f.availableMetricTypes = &avaibleMetrics
	f.cacheTimestamp = time.Now()
	return nil
}

func NewFacterPlugin() *Facter {
	f := new(Facter)
	//TODO read from config
	f.cacheTTL = DefaultCacheTTL
	return f
}

func (f *Facter) GetMetricTypes(kotens plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	if time.Since(f.cacheTimestamp) > f.cacheTTL {

		//TODO args: create slice, flatten with " " separator
		out, err := exec.Command(f.shellPath, f.shellArgs, f.facterPath+" "+f.facterArgs).Output()
		if err != nil {
			log.Println("exec returned " + err.Error())
			return err
		}
		f.cacheTimestamp = time.Now()

		var facterMap map[string]interface{}
		err = json.Unmarshal(out, &facterMap)
		if err != nil {
			log.Println("Unmarshal failed " + err.Error())
			return err
		}

		avaibleMetrics := make([]*plugin.MetricType, 0, len(facterMap))
		for key := range facterMap {
			avaibleMetrics = append(
				avaibleMetrics,
				plugin.NewMetricType(
					append(namespace, key),
					f.cacheTimestamp.Unix()))
		}

		f.avaibleMetrics = &avaibleMetrics
		reply.MetricTypes = *f.avaibleMetrics

		return nil
	} else {
		reply.MetricTypes = *f.availableMetricTypes
		return nil
	}
}

func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {
	// it would be: CollectMetrics([]plugin.MetricType) ([]plugin.Metric, error)
	// waits for lynxbat/SDI-98

	//	out, err := exec.Command("sh", "-c", f.facterPath+" -j").Output()
	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type)
}

func ConfigPolicy() *plugin.ConfigPolicy {
	//TODO What is plugin policy?

	c := new(plugin.ConfigPolicy)
	return c
}

/**********************
 *  facter interface  *
 **********************/

// get facts from facter (external command)
func getFacts() (*map[string]interface{}, error) {

	//TODO args: create slice, flatten with " " separator
	out, err := exec.Command("facter", "--json").Output()
	if err != nil {
		log.Println("exec returned " + err.Error())
		return nil, err
	}

	log.Println("OUT:")
	log.Println(out)
	var facterMap map[string]interface{}
	err = json.Unmarshal(out, &facterMap)
	if err != nil {
		log.Println("Unmarshal failed " + err.Error())
		return nil, err
	}
	return &facterMap, nil
}
