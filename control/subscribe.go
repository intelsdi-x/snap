package control

import (
	"errors"
	"strings"
	"sync"
)

// the struct holding the table of subscriptions
type subscriptions struct {

	// currentIter is used when iterating over the table
	currentIter int

	// the map holding the subscription data
	table *map[string]int

	// keys is used when iterating, the index from currentIter is used to
	// retrieve a key via an index, then this key is used to retrieve
	// an item in the table (map)
	keys *[]string

	// used to make atomic changes to the table
	mutex *sync.Mutex
}

// returns the key and value of a certain index in the table.
// to be used when iterating over the table
func (s *subscriptions) Item() (string, int) {
	key := (*s.keys)[s.currentIter-1]
	return key, (*s.table)[key]
}

// Returns true until the "end" of the table is reached.
// used to iterate over the table:
/*
for sub.Next() {
	key, val := sub.Item()
	// do things with key / val
}
*/
func (s *subscriptions) Next() bool {
	s.currentIter++
	if s.currentIter > len(*s.table) {
		s.currentIter = 0
		return false
	}
	return true
}

// Since subscriptions should be a singleton inside a controller,
// rather than using the NewSubscription() constructor pattern,
// we use Init to initialize an existing subscrition table:
/*
//...
control.sub := new(subscriptions)
sub.Init()
*/
func (s *subscriptions) Init() {
	tabe := make(map[string]int)
	s.table = &tabe
	s.mutex = new(sync.Mutex)
	s.keys = new([]string)
}

// atomically increments a metric's subscription count in the table
// if the key does not exist, it is added with a count of 1
func (s *subscriptions) Subscribe(key string) {
	s.Lock()
	if _, ok := (*s.table)[key]; !ok {
		*s.keys = append(*s.keys, key)
	}
	(*s.table)[key]++
	s.Unlock()
}

// atomically decrements a metric's count in the table
// if the key does not exist, or the count is already zero,
// an error is returned
func (s *subscriptions) Unsubscribe(key string) error {
	s.Lock()
	defer s.Unlock()
	if (*s.table)[key] == 0 {
		return errors.New("subscription count cannot be less than 0 for key " + key)
	}
	(*s.table)[key]--
	return nil
}

// returns the current subscription count of a key
func (s *subscriptions) Count(key string) int {
	s.Lock()
	defer s.Unlock()
	return (*s.table)[key]
}

// an exposure of the mutex of the table, likely needed when iterating over the table:
/*
sub.Lock()
for sub.Next() {
	// do work
}
sub.Unlock()
*/
func (s *subscriptions) Lock() {
	s.mutex.Lock()
}

// the counterpart to Lock, releases the lock.
func (s *subscriptions) Unlock() {
	s.mutex.Unlock()
}

func getMetricKey(metric []string) string {
	return strings.Join(metric, ".")
}
