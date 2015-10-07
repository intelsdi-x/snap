package ipmiplugin

import (
	"fmt"
	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-ipmi/ipmi"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func req2Str(req ipmi.IpmiRequest) string {
	return fmt.Sprintf("%v", req)
}

func res2Str(res ipmi.IpmiResponse) string {
	return fmt.Sprintf("%v", res)
}

func bs2Str(bs []byte) string {
	return fmt.Sprintf("%v", bs)
}

func ns2Str(ns []string) string {
	return fmt.Sprintf("%v", ns)
}

type fakeIpmi struct {
	ret_err   error
	ret_res   map[string]ipmi.IpmiResponse
	logged_rq []string
	logged_n  int

	batchResuests int
}

func (al *fakeIpmi) BatchExecRaw(requests []ipmi.IpmiRequest, nSim int) ([]ipmi.IpmiResponse, error) {
	al.batchResuests++

	al.logged_rq = make([]string, len(requests))
	res := make([]ipmi.IpmiResponse, len(requests))
	for i, r := range requests {
		s := req2Str(r)
		al.logged_rq[i] = s
		res[i] = al.ret_res[s]
	}
	al.logged_n = nSim

	return res, al.ret_err
}

type fakeParserSimple struct {
	validMap     map[string]error
	validDefault error
	metrics      []string

	validateCalled map[string]int
	parseCalled    map[string]int

	parseResults map[string]map[string]uint16
}

func NewFakeParserSimple() *fakeParserSimple {
	r := fakeParserSimple{}
	r.validMap = make(map[string]error)
	r.validateCalled = make(map[string]int)
	r.parseCalled = make(map[string]int)
	r.parseResults = make(map[string]map[string]uint16)
	return &r
}

func (p *fakeParserSimple) GetMetrics() []string {
	return p.metrics
}
func (p *fakeParserSimple) Validate(response []byte) error {
	p.validateCalled[bs2Str(response)]++
	v, ok := p.validMap[bs2Str(response)]
	if ok {
		return v
	} else {
		return p.validDefault
	}
}
func (p *fakeParserSimple) Parse(response []byte) map[string]uint16 {
	p.parseCalled[bs2Str(response)]++

	m, ok := p.parseResults[bs2Str(response)]
	if ok {
		return m
	} else {
		return map[string]uint16{}
	}
}

func TestCollectMetrics(t *testing.T) {
	Convey("CollectMetrics", t, func() {
		ipmilayer := &fakeIpmi{}
		format1 := NewFakeParserSimple()
		format1.metrics = []string{"qwe", "rty"}
		format2 := NewFakeParserSimple()
		format2.metrics = []string{"ppp"}
		format3 := NewFakeParserSimple()
		format3.metrics = []string{"x"}
		vendor := []ipmi.RequestDescription{
			{ipmi.IpmiRequest{[]byte{1, 2, 3, 9}, 9, 5}, "a", format1},
			{ipmi.IpmiRequest{[]byte{2, 2}, 8, 6}, "b/c", format2},
			{ipmi.IpmiRequest{[]byte{3, 2, 3}, 6, 7}, "b/d", format3},
		}

		mts := []plugin.PluginMetricType{
			plugin.PluginMetricType{Namespace_: []string{"intel", "ipmi", "a", "qwe"}},
			plugin.PluginMetricType{Namespace_: []string{"intel", "ipmi", "a", "rty"}},
			plugin.PluginMetricType{Namespace_: []string{"intel", "ipmi", "b/c", "ppp"}},
			plugin.PluginMetricType{Namespace_: []string{"intel", "ipmi", "b/d", "x"}},
		}

		ipmilayer.ret_res = map[string]ipmi.IpmiResponse{
			req2Str(vendor[0].Request): ipmi.IpmiResponse{[]byte{0, 1}},
			req2Str(vendor[1].Request): ipmi.IpmiResponse{[]byte{0, 2}},
			req2Str(vendor[2].Request): ipmi.IpmiResponse{[]byte{0, 3}},
		}

		format1.parseResults[bs2Str([]byte{0, 1})] = map[string]uint16{
			"qwe": 1, "rty": 2,
		}
		format2.parseResults[bs2Str([]byte{0, 2})] = map[string]uint16{
			"ppp": 3,
		}
		format3.parseResults[bs2Str([]byte{0, 3})] = map[string]uint16{
			"x": 4,
		}

		sut := &IpmiCollector{IpmiLayer: ipmilayer, Vendor: vendor, NSim: 123}

		Convey("Should do each required call to ipmi to get required metrics", func() {

			sut.CollectMetrics(mts)

			for _, v := range vendor {
				So(ipmilayer.logged_rq, ShouldContain, req2Str(v.Request))
			}

			Convey("Each call to ipmi is done exactly once", func() {

				sut.CollectMetrics(mts)
				cnt := map[string]int{}

				for _, r := range ipmilayer.logged_rq {
					cnt[r]++
				}

				for k, v := range cnt {
					So(v, ShouldEqual, 1)
					if v != 1 {
						fmt.Printf("Actual count for %v is %d", k, v)
					}
				}

			})

			Convey("Souldn't do any unnecessary requests", func() {

				sut.CollectMetrics(mts)

				wanted := make([]string, len(vendor))
				for i, v := range vendor {
					wanted[i] = req2Str(v.Request)
				}

				for _, r := range ipmilayer.logged_rq {
					So(wanted, ShouldContain, r)
				}

			})

			Convey("Passes correct number of simultaneous requests to ipmi layer", func() {

				sut.CollectMetrics(mts)

				So(ipmilayer.logged_n, ShouldEqual, 123)

			})

		})

		Convey("List of returned metrics matches list of requested metrics", func() {

			dut, _ := sut.CollectMetrics(mts)

			expected := [][]string{
				[]string{"intel", "ipmi", "a", "qwe"},
				[]string{"intel", "ipmi", "a", "rty"},
				[]string{"intel", "ipmi", "b/c", "ppp"},
				[]string{"intel", "ipmi", "b/d", "x"},
			}

			got := make([]string, len(dut))

			for i, r := range dut {
				got[i] = ns2Str(r.Namespace())
			}

			So(len(dut), ShouldEqual, len(expected))
			for _, r := range expected {
				So(got, ShouldContain, ns2Str(r))
			}

		})

		Convey("Error should be returned if ipmi layer returned error", func() {

			ipmilayer.ret_err = fmt.Errorf("TEST")

			_, err_dut := sut.CollectMetrics(mts)

			So(err_dut, ShouldNotBeNil)

		})

		Convey("Correct parser is called for each request", func() {

			sut.CollectMetrics(mts)

			So(len(format1.parseCalled), ShouldEqual, 1)
			So(format1.parseCalled[bs2Str([]byte{0, 1})], ShouldEqual, 1)

			So(len(format2.parseCalled), ShouldEqual, 1)
			So(format2.parseCalled[bs2Str([]byte{0, 2})], ShouldEqual, 1)

			So(len(format3.parseCalled), ShouldEqual, 1)
			So(format3.parseCalled[bs2Str([]byte{0, 3})], ShouldEqual, 1)

		})

		Convey("Function returns what parser returned", func() {
			expected := map[string]uint16{
				ns2Str([]string{"intel", "ipmi", "a", "qwe"}):   1,
				ns2Str([]string{"intel", "ipmi", "a", "rty"}):   2,
				ns2Str([]string{"intel", "ipmi", "b/c", "ppp"}): 3,
				ns2Str([]string{"intel", "ipmi", "b/d", "x"}):   4,
			}

			dut, _ := sut.CollectMetrics(mts)

			for _, v := range dut {
				So(v.Data(), ShouldEqual, expected[ns2Str(v.Namespace())])
			}

		})

		Convey("Has correct validation", func() {

			Convey("Each request is validated", func() {

				sut.CollectMetrics(mts)

				So(len(format1.validateCalled), ShouldEqual, 1)
				So(format1.validateCalled[bs2Str([]byte{0, 1})], ShouldEqual, 1)

				So(len(format2.validateCalled), ShouldEqual, 1)
				So(format2.validateCalled[bs2Str([]byte{0, 2})], ShouldEqual, 1)

				So(len(format3.validateCalled), ShouldEqual, 1)
				So(format3.validateCalled[bs2Str([]byte{0, 3})], ShouldEqual, 1)

			})

			Convey("If validation fails error is returned", func() {

				format3.validMap[bs2Str([]byte{0, 3})] = fmt.Errorf("x")

				_, err := sut.CollectMetrics(mts)

				So(err, ShouldNotBeNil)

			})

		})

		Convey("When everything is ok no error is returned", func() {
			_, err := sut.CollectMetrics(mts)
			So(err, ShouldBeNil)
		})

	})
}

func TestGetMetricTypes(t *testing.T) {

	Convey("GetMetricTypes", t, func() {

		format1 := NewFakeParserSimple()
		format1.metrics = []string{"qwe", "rty"}
		format2 := NewFakeParserSimple()
		format2.metrics = []string{"ppp"}
		format3 := NewFakeParserSimple()
		format3.metrics = []string{"x"}
		vendor := []ipmi.RequestDescription{
			{ipmi.IpmiRequest{[]byte{1, 2, 3, 9}, 9, 5}, "a", format1},
			{ipmi.IpmiRequest{[]byte{2, 2}, 8, 6}, "b/c", format2},
			{ipmi.IpmiRequest{[]byte{3, 2, 3}, 6, 7}, "b/d", format3},
		}

		sut := &IpmiCollector{IpmiLayer: nil, Vendor: vendor, NSim: 123}

		dut, _ := sut.GetMetricTypes()

		k1 := []string{}
		k2 := []string{}
		k3 := []string{}

		for _, mt := range dut {
			v := mt.Namespace()
			switch {
			case v[2] == "a":
				k1 = append(k1, ns2Str(v[3:]))

			case v[2] == "b" && v[3] == "c":
				k2 = append(k2, ns2Str(v[4:]))

			case v[2] == "b" && v[3] == "d":
				k3 = append(k3, ns2Str(v[4:]))

			}
		}

		Convey("Should return each root key from vendor", func() {

			So(len(k1), ShouldBeGreaterThan, 0)
			So(len(k2), ShouldBeGreaterThan, 0)
			So(len(k3), ShouldBeGreaterThan, 0)

			Convey("Each root key should have metrics exposed by parser", func() {

				So(k1, ShouldContain, ns2Str([]string{"qwe"}))
				So(k1, ShouldContain, ns2Str([]string{"qwe"}))

				So(k2, ShouldContain, ns2Str([]string{"ppp"}))

				So(k3, ShouldContain, ns2Str([]string{"x"}))

			})

		})

	})

}
