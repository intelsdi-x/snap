/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package scheduler

import (
	"errors"
	"fmt"
	"sync"
)

type managers struct {
	mutex          *sync.RWMutex
	local          managesMetrics
	remoteManagers map[string]managesMetrics
}

func newManagers(mm managesMetrics) managers {
	return managers{
		mutex:          &sync.RWMutex{},
		remoteManagers: make(map[string]managesMetrics),
		local:          mm,
	}
}

// Adds the key:value to the remoteManagers map to make them accessible
// via Get() calls.
func (m *managers) Add(key string, val managesMetrics) {
	m.mutex.Lock()
	if key == "" {
		m.local = val
	} else {
		m.remoteManagers[key] = val
	}
	m.mutex.Unlock()
}

// Returns the managesMetric instance that maps to given
// string. If an empty string is given, will instead return
// the local instance passed in on initialization.
func (m *managers) Get(key string) (managesMetrics, error) {
	if key == "" {
		return m.local, nil
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if val, ok := m.remoteManagers[key]; ok {
		return val, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Client not found for: %v", key))
	}
}
