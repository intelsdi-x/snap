package agent

import (
	"testing"
	"github.com/lynxbat/pulse/agent/collection"
	"runtime"
)

const test_caching = true
const test_caching_ttl float64 = 5

func TestMetric(t *testing.T) {}

func BenchmarkMetricCollectDSingleCore(b *testing.B) {	
	count := 1
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricCollectDSingleCore10(b *testing.B) {
	count := 10
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricCollectDSingleCore100(b *testing.B) {
	count := 100
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricCollectDSingleCore1000(b *testing.B) {
	count := 1000
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricCollectDSingleCore100000(b *testing.B) {
	count := 100000
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricCollectDMultiCore(b *testing.B) {
	count := 1
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l, runtime.NumCPU())
	}
}

func BenchmarkMetricCollectDMultiCore10(b *testing.B) {
	count := 10
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l, runtime.NumCPU())
	}
}

func BenchmarkMetricCollectDMultiCore100(b *testing.B) {
	count := 100
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l, runtime.NumCPU())
	}
}

func BenchmarkMetricCollectDMultiCore1000(b *testing.B) {
	count := 1000
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l, runtime.NumCPU())
	}
}

func BenchmarkMetricCollectDMultiCore100000(b *testing.B) {
	count := 100000
	c := collection.NewCollectDCollector(UNIX_SOCK, test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l, runtime.NumCPU())
	}
}

func BenchmarkMetricFacterSingleCore(b *testing.B) {
	count := 1
	c := collection.NewFacterCollector("facter", test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricFacterSingleCore10(b *testing.B) {
	count := 10
	c := collection.NewFacterCollector("facter", test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricFacterSingleCore100(b *testing.B) {
	count := 100
	c := collection.NewFacterCollector("facter", test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricFacterSingleCore1000(b *testing.B) {
	count := 1000
	c := collection.NewFacterCollector("facter", test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}

func BenchmarkMetricFacterSingleCore100000(b *testing.B) {
	count := 100000
	c := collection.NewFacterCollector("facter", test_caching, test_caching_ttl)
	l := c.GetMetricList()
	addition := count - len(l)

	for x := 0; x < addition; x++ {
		l = append(l, l[0])
	}
	l = l[:count]

	//	fmt.Printf("\nTotal metrics: %v\n", len(l))
	for n := 0; n < b.N; n++ {
		c.GetMetricValues(l)
	}
}
