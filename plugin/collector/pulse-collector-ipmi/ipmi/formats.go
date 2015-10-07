package ipmi

import (
	"errors"
	"fmt"
)

// Performs basic response validation. Checks response code ensures response
// has non-zero lenght.
type GenericValidator struct {
}

func (gv *GenericValidator) Validate(response []byte) error {
	if len(response) > 0 {
		if response[0] == 0 {
			return nil
		} else {
			return fmt.Errorf("Unexpected error code : %d", response[0])
		}
	} else {
		return errors.New("Zero length response")
	}
}

// Extracts data from CUPS specific response format.
// Data contains info about cpu utilization and memory & io bandwidth.
type ParserCUPS struct {
	*GenericValidator
}

// Instance of ParserCUPS
var FormatCUPS = &ParserCUPS{}

func (p *ParserCUPS) GetMetrics() []string {
	return []string{"cpu_cstate", "memory_bandwith", "io_bandwith"}
}

func (p *ParserCUPS) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m["cpu_cstate"] = uint16(response[4]) + uint16(response[5])*256
	m["memory_bandwith"] = uint16(response[6]) + uint16(response[7])*256
	m["io_bandwith"] = uint16(response[8]) + uint16(response[9])*256
	return m
}

// Extracs data from Node manager response format.
// Data contains current, min, max and average value.
type ParserNodeManager struct {
	*GenericValidator
}

// Instance of ParserNodeManager
var FormatNodeManager = &ParserNodeManager{}

func (p *ParserNodeManager) GetMetrics() []string {
	return []string{"", "min", "max", "avg"}
}

func (p *ParserNodeManager) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m[""] = uint16(response[4]) + uint16(response[5])*256
	m["min"] = uint16(response[6]) + uint16(response[7])*256
	m["max"] = uint16(response[8]) + uint16(response[9])*256
	m["avg"] = uint16(response[10]) + uint16(response[11])*256
	return m
}

// Extracts temperature data.
// Data contains info about temperatures for first 4 cpus
// and 64 dimms.
type ParserTemp struct {
	*GenericValidator
}

// Instance of ParserTempMargin.
var FormatTemp = &ParserTemp{}

func (p *ParserTemp) GetMetrics() []string {
	a := []string{"cpu/cpu0", "cpu/cpu1", "cpu/cpu2", "cpu/cpu3"}
	for i := 8; i < 72; i++ {
		c := fmt.Sprintf("memory/dimm%d", i-8)
		a = append(a, c)
	}
	return a
}

func (p *ParserTemp) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m["cpu/cpu0"] = uint16(response[4])
	m["cpu/cpu1"] = uint16(response[5])
	m["cpu/cpu2"] = uint16(response[6])
	m["cpu/cpu3"] = uint16(response[7])
	for i := 8; i < len(response); i++ {
		a := fmt.Sprintf("memory/dimm%d", i-8)
		m[a] = uint16(response[i])
	}

	return m
}

// Extracts temperature margin datas from PECI response.
// Main metric value is TJ max.
// margin_offset current value of margin offset, which is value
// of TJ max reduction.
type ParserPECI struct {
	*GenericValidator
}

// Instance of ParserPECI.
var FormatPECI = &ParserPECI{}

func (p *ParserPECI) GetMetrics() []string {
	return []string{"", "margin_offset"}
}

func (p *ParserPECI) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m["margin_offset"] = uint16(response[6])
	m[""] = uint16(response[7]) + uint16(response[8])*256
	return m
}

// Extracts temperatures of voltage regulators.
type ParserPMBus struct {
	*GenericValidator
}

// Instance of ParserPMBus.
var FormatPMBus = &ParserPMBus{}

func (p *ParserPMBus) GetMetrics() []string {
	return []string{"VR0", "VR1", "VR2", "VR3", "VR4", "VR5"}
}

func (p *ParserPMBus) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m["VR0"] = uint16(response[4]) + uint16(response[5])*256
	m["VR1"] = uint16(response[6]) + uint16(response[7])*256
	m["VR2"] = uint16(response[8]) + uint16(response[9])*256
	m["VR3"] = uint16(response[10]) + uint16(response[11])*256
	m["VR4"] = uint16(response[12]) + uint16(response[13])*256
	m["VR5"] = uint16(response[14]) + uint16(response[15])*256
	return m
}

// Extracts sensor value from response to Get Sensor Record.
type ParserSR struct {
	*GenericValidator
}

// Instance of ParserSR.
var FormatSR = &ParserSR{}

func (p *ParserSR) GetMetrics() []string {
	return []string{""}
}

func (p *ParserSR) Parse(response []byte) map[string]uint16 {
	m := map[string]uint16{}
	m[""] = uint16(response[1])
	return m
}
