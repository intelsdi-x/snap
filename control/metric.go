package control

import (
	"sort"
	"strings"

	"github.com/intelsdilabs/gomit"
)

type metrics map[string]*metric

func (m *metrics) Max() int {
	if m.len() > 0 {
		var subs []int
		for _, s := range *m {
			subs = append(subs, s.Subscriptions())
		}
		sort.Ints(subs)
		return subs[len(subs)-1]
	}
	return 0
}

func (m *metrics) len() int {
	return len(*m)
}

type metric struct {
	Namespace []string

	emitter gomit.Emitter
	sub     *subscriptions
}

type metricOpts struct {
	Namespace []string
}

func newMetric(opts *metricOpts) *metric {
	m := &metric{
		Namespace: opts.Namespace,

		emitter: gomit.NewEventController(),
		sub:     new(subscriptions),
	}
	return m
}

func (m *metric) Key() string {
	return strings.Join(m.Namespace, ".")
}
