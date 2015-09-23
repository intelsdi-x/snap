// Tests for ipmi commands parser

package ipmi

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidator(t *testing.T) {
	Convey("Check validator", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}
		a := &GenericValidator{}
		err := errors.New("Zero length response")
		validator := a.Validate(validResponse)
		So(validator, ShouldEqual, nil)
		validResponse = []byte{}
		validator = a.Validate(validResponse)
		So(validator.Error(), ShouldEqual, err.Error())
		validResponse = []byte{0x88, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}
		err = fmt.Errorf("Unexpected error code : %d", validResponse[0])
		validator = a.Validate(validResponse)
		So(validator.Error(), ShouldEqual, err.Error())
	})
}

func TestCUPSParsing(t *testing.T) {
	Convey("Check CUPS parser", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}
		a := &ParserCUPS{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"cpu_cstate", "memory_bandwith", "io_bandwith"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut["cpu_cstate"], ShouldEqual, 100)
		So(parserOut["memory_bandwith"], ShouldEqual, 80)
		So(parserOut["io_bandwith"], ShouldEqual, 256)
	})
}

func TestNodeManagerParsing(t *testing.T) {
	Convey("Check NodeManager parser", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, 0x69, 0x00, 0x03, 0x00, 0x7d, 0x01, 0x6E, 0x00, 0xC7, 0x3F, 0x05, 0x56, 0xB9, 0xAD, 0x0C, 0x00, 0x50}
		a := &ParserNodeManager{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"", "min", "max", "avg"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut[""], ShouldEqual, 105)
		So(parserOut["min"], ShouldEqual, 3)
		So(parserOut["max"], ShouldEqual, 381)
		So(parserOut["avg"], ShouldEqual, 110)
	})
}

func TestPECIParsing(t *testing.T) {
	Convey("Check PECI parser", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, 0x40, 0x00, 0x0A, 0x59, 0x00}
		a := &ParserPECI{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"", "margin_offset"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut[""], ShouldEqual, 89)
		So(parserOut["margin_offset"], ShouldEqual, 10)
	})
}

func TestPMBusParsing(t *testing.T) {
	Convey("Check PMBus parser", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, 0x25, 0x00, 0x2A, 0x00, 0x1F, 0x00, 0x21, 0x00, 0x20, 0x00, 0x1F, 0x00}
		a := &ParserPMBus{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"VR0", "VR1", "VR2", "VR3", "VR4", "VR5"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut["VR0"], ShouldEqual, 37)
		So(parserOut["VR1"], ShouldEqual, 42)
		So(parserOut["VR2"], ShouldEqual, 31)
		So(parserOut["VR3"], ShouldEqual, 33)
		So(parserOut["VR4"], ShouldEqual, 32)
		So(parserOut["VR5"], ShouldEqual, 31)
	})
}

func TestTemperatureParsing(t *testing.T) {
	Convey("Check Temperature parser", t, func() {
		validResponse := []byte{0x00, 0x57, 0x01, 0x00, //response header
			0x23, 0x25, 0xFF, 0xFF, //CPUS
			0xFF, 0xFF, 0x1E, 0x20, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x1F, 0x23, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x20, 0x22, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x1E, 0x21, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		a := &ParserTempMargin{}
		metrics := a.GetMetrics()
		fmt.Println(metrics)
		parserOut := a.Parse(validResponse)
		expects := []string{"cpu/cpu0", "cpu/cpu1", "cpu/cpu2", "cpu/cpu3",
			"memory/dimm0", "memory/dimm1", "memory/dimm2", "memory/dimm3", "memory/dimm4",
			"memory/dimm5", "memory/dimm6", "memory/dimm7", "memory/dimm8", "memory/dimm9",
			"memory/dimm10", "memory/dimm11", "memory/dimm12", "memory/dimm13", "memory/dimm14",
			"memory/dimm15", "memory/dimm16", "memory/dimm17", "memory/dimm18", "memory/dimm19",
			"memory/dimm20", "memory/dimm21", "memory/dimm22", "memory/dimm23", "memory/dimm24",
			"memory/dimm25", "memory/dimm26", "memory/dimm27", "memory/dimm28", "memory/dimm29",
			"memory/dimm30", "memory/dimm31", "memory/dimm32", "memory/dimm33", "memory/dimm34",
			"memory/dimm35", "memory/dimm36", "memory/dimm37", "memory/dimm38", "memory/dimm39",
			"memory/dimm40", "memory/dimm41", "memory/dimm42", "memory/dimm43", "memory/dimm44",
			"memory/dimm45", "memory/dimm46", "memory/dimm47", "memory/dimm48", "memory/dimm49",
			"memory/dimm50", "memory/dimm51", "memory/dimm52", "memory/dimm53", "memory/dimm54",
			"memory/dimm55", "memory/dimm56", "memory/dimm57", "memory/dimm58", "memory/dimm59",
			"memory/dimm60", "memory/dimm61", "memory/dimm62", "memory/dimm63"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(metrics); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		for i := 0; i < len(metrics); i++ {
			So(parserOut[metrics[i]], ShouldEqual, validResponse[i+4])
		}
	})
}
