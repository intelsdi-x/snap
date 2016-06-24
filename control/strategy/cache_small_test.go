// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package strategy

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewCache(t *testing.T) {
	c := NewCache(1 * time.Second)
	Convey("Calling NewCache(1 * time.Second)", t, func() {
		Convey("Should return a cache", func() {
			So(c, ShouldNotBeNil)
		})
		Convey("Cache should have a TTL of 1s", func() {
			So(c.ttl, ShouldResemble, time.Duration(1*time.Second))
		})
	})
}

// TestUpdateCache places static and dynamic entries into the cache and verify those entries
// are inserted into the cache correctly.
func TestUpdateCache(t *testing.T) {
	scache := NewCache(300 * time.Millisecond)
	// Static metrics
	staticMetrics := []core.Metric{
		fixtures.MockMetricType{
			Namespace_: core.NewNamespace("foo", "bar"),
			Ver:        0,
		},
		fixtures.MockMetricType{
			Namespace_: core.NewNamespace("foo", "baz"),
			Ver:        0,
		},
	}
	scache.updateCache(staticMetrics)
	Convey("Updating cache with two static metrics", t, func() {
		Convey("Should result in a cache with two entries", func() {
			So(len(scache.table), ShouldEqual, 2)
		})
		Convey("Should have an entry for '/foo/bar:0'", func() {
			_, ok := scache.table["/foo/bar:0"]
			So(ok, ShouldBeTrue)
		})
		Convey("Should have an entry for '/foo/baz:0'", func() {
			_, ok := scache.table["/foo/baz:0"]
			So(ok, ShouldBeTrue)
		})
	})
	dcache := NewCache(300 * time.Millisecond)
	// Dynamic Metrics
	dynNS1 := core.NewNamespace("foo", "bar").AddDynamicElement("host", "Mock host").AddStaticElement("qux")
	dynNS2 := core.NewNamespace("foo", "baz").AddDynamicElement("host", "Mock host").AddStaticElement("avg")
	dynNS3 := core.NewNamespace("foo", "bar").AddDynamicElement("host", "Mock host").AddDynamicElement("rack", "Mock rack").AddStaticElement("temp")
	mockNS1 := make([]core.NamespaceElement, len(dynNS1))
	mockNS2 := make([]core.NamespaceElement, len(dynNS1))
	mockNS3 := make([]core.NamespaceElement, len(dynNS2))
	mockNS4 := make([]core.NamespaceElement, len(dynNS3))
	copy(mockNS1, dynNS1)
	copy(mockNS2, dynNS1)
	copy(mockNS3, dynNS2)
	copy(mockNS4, dynNS3)
	mockNS1[2].Value = "hosta"
	mockNS2[2].Value = "hostb"
	mockNS3[2].Value = "hosta"
	mockNS4[2].Value = "hostc"
	mockNS4[3].Value = "rack1"

	dynamicMetrics := []core.Metric{
		fixtures.MockMetricType{
			Namespace_: mockNS1,
			Ver:        0,
		},
		fixtures.MockMetricType{
			Namespace_: mockNS2,
			Ver:        0,
		},
		fixtures.MockMetricType{
			Namespace_: mockNS3,
			Ver:        0,
		},
		fixtures.MockMetricType{
			Namespace_: mockNS4,
			Ver:        0,
		},
	}
	dcache.updateCache(dynamicMetrics)
	Convey("Updating cache with four metrics on three dynamic namespaces", t, func() {
		Convey("Should result in a cache with two entries", func() {
			So(len(dcache.table), ShouldEqual, 3)
		})
		Convey("Should have an entry for '/foo/bar/*/qux:0'", func() {
			_, ok := dcache.table["/foo/bar/*/qux:0"]
			So(ok, ShouldBeTrue)
		})
		Convey("Should have an entry for '/foo/baz/*/avg:0'", func() {
			_, ok := dcache.table["/foo/baz/*/avg:0"]
			So(ok, ShouldBeTrue)
		})
		Convey("Should have an entry for '/foo/bar/*/*/temp:0'", func() {
			_, ok := dcache.table["/foo/bar/*/*/temp:0"]
			So(ok, ShouldBeTrue)
		})
	})
}
