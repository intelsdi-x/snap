package metric

type Metric struct {
	Namespace string

	sub *subscriptions
}

type MetricOpts struct {
	Namespace string
}

func NewMetric(opts *MetricOpts) *Metric {
	m := &Metric{
		Namespace: opts.Namespace,

		sub: new(subscriptions),
	}
	return m
}
