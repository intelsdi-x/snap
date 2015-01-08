package availability

import "errors"

type Subscriptions int

func (s *Subscriptions) Add() {
	*s = *s + 1
}

func (s *Subscriptions) Remove() error {
	if int(*s) > 0 {
		*s = *s - 1
		return nil
	}
	return errors.New("count is at zero")
}

func (s *Subscriptions) Count() int {
	return int(*s)
}
