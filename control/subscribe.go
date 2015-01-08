package control

import (
	"errors"
)

type SubscriptionEvent struct {
	Count           int
	MetricNamespace []string

	namespace *string
}

func (se *SubscriptionEvent) Namespace() string {
	return *se.namespace
}

func (ap *availablePlugin) Subscriptions() int {
	return ap.Metrics.Max()
}

func (m *metric) Subscribe() {
	m.sub.Add()
	evns := "Control.MetricSubscribed"
	m.emitter.Emit(&SubscriptionEvent{
		Count:           m.Subscriptions(),
		MetricNamespace: m.Namespace,

		namespace: &evns,
	})
}

func (m *metric) Unsubscribe() error {
	err := m.sub.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (m *metric) Subscriptions() int {
	return m.sub.Count()
}

type subscriptions int

func (s *subscriptions) Add() {
	*s = *s + 1
}

func (s *subscriptions) Remove() error {
	if int(*s) > 0 {
		*s = *s - 1
		return nil
	}
	return errors.New("count is at zero")
}

func (s *subscriptions) Count() int {
	return int(*s)
}
