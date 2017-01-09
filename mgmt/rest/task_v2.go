/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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

package rest

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/v2/rbody"
	"github.com/julienschmidt/httprouter"
)

func (s *Server) getTasksV2(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// get tasks from the task manager
	sts := s.mt.GetTasks()

	// create the task list response
	tasks := make(rbody.Tasks, len(sts))
	i := 0
	for _, t := range sts {
		tasks[i] = *rbody.SchedulerTaskFromTask(t)
		tasks[i].Href = taskURI(r.Host, "v2", t)
		i++
	}
	sort.Sort(tasks)

	respondV2(200, tasks, w)
}

func taskURI(host, version string, t core.Task) string {
	return fmt.Sprintf("%s://%s/%s/tasks/%s", protocolPrefix, host, version, t.ID())
}
