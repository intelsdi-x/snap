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

package response


type Plugin struct {
	Name            string        `json:"name"`
	Version         int           `json:"version"`
	Type            string        `json:"type"`
	Signed          bool          `json:"signed"`
	Status          string        `json:"status"`
	LoadedTimestamp int64         `json:"loaded_timestamp"`
	Href            string        `json:"href"`
	ConfigPolicy    []PolicyTable `json:"policy,omitempty"`
}

type RunningPlugin struct {
	Name             string `json:"name"`
	Version          int    `json:"version"`
	Type             string `json:"type"`
	HitCount         int    `json:"hitcount"`
	LastHitTimestamp int64  `json:"last_hit_timestamp"`
	ID               uint32 `json:"id"`
	Href             string `json:"href"`
	PprofPort        string `json:"pprof_port"`
}