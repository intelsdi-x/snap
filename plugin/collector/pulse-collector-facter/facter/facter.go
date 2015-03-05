package facter

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/intelsdilabs/pulse/plugin/helper"
)

var (
	namespace = []string{"intel", "facter"}
	Name      = GetPluginName(&namespace) //preprocessor needed
)

const (
	//	Name    = "facter" //should it be intel/facter ?
	Version = 1
	Type    = plugin.CollectorPluginType
)

type Facter struct {
	avaibleMetrics *[]*plugin.MetricType //map[string]interface{}

	cacheTimestamp time.Time
	cacheTTL       time.Duration

	facterPath string
	shellPath  string
	shellArgs  string
}

func NewFacterPlugin() *Facter {
	f := new(Facter)
	//TODO read from config
	f.cacheTTL = 60 * time.Second
	f.facterPath = "/usr/bin/facter"
	f.shellPath = "/usr/bin/sh"
	f.shellArgs = "-c" //TODO slice of args of course
	return f
}

func (f *Facter) GetMetricTypes(kotens plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {

	if time.Since(f.cacheTimestamp) > f.cacheTTL {

		//TODO create slice, flatten with " " separator
		out, err := exec.Command(f.shellPath, " ", f.shellArgs, " ", f.facterPath+" -j").Output()
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
		reply.MetricTypes = *f.avaibleMetrics
		return nil
	}
}

func (f *Facter) Collect(args plugin.CollectorArgs, reply *plugin.CollectorReply) error {

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
