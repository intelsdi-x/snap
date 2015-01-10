package control

import (
	"errors"
	"sync"
)

type subscriptions struct {
	currentIter int
	table       *map[string]int
	keys        *[]string
	mutex       *sync.Mutex
}

func (s *subscriptions) Values() (string, int) {
	key := (*s.keys)[s.currentIter-1]
	return key, (*s.table)[key]
}

func (s *subscriptions) Next() bool {
	s.currentIter++
	if s.currentIter > len(*s.table) {
		s.currentIter = 0
		return false
	}
	return true
}

func (s *subscriptions) Init() {
	tabe := make(map[string]int)
	s.table = &tabe
	s.mutex = new(sync.Mutex)
	s.keys = new([]string)
}

func (s *subscriptions) Subscribe(key string) {
	s.Lock()
	if _, ok := (*s.table)[key]; !ok {
		*s.keys = append(*s.keys, key)
	}
	(*s.table)[key]++
	s.Unlock()
}

func (s *subscriptions) Unsubscribe(key string) error {
	s.Lock()
	defer s.Unlock()
	if (*s.table)[key] == 0 {
		return errors.New("subscription count cannot be less than 0 for key " + key)
	}
	(*s.table)[key]--
	return nil
}

func (s *subscriptions) Count(key string) int {
	return (*s.table)[key]
}

func (s *subscriptions) Lock() {
	s.mutex.Lock()
}

func (s *subscriptions) Unlock() {
	s.mutex.Unlock()
}
