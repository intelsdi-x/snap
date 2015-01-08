package metric

import "errors"

func (m *Metric) Subscribe() {
	m.sub.Add()
}

func (m *Metric) Unsubscribe() error {
	err := m.sub.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (m *Metric) Subscriptions() int {
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
