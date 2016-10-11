package v2

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody/v2"
)

func GetTasks(w http.ResponseWriter, r *http.Request, sts map[string]core.Task) {
	tasks := &v2.ScheduledTaskListReturned{}
	tasks.ScheduledTasks = make([]rbody.ScheduledTask, len(sts))

	i := 0
	for _, t := range sts {
		tasks.ScheduledTasks[i] = *rbody.SchedulerTaskFromTask(t)
		tasks.ScheduledTasks[i].Href = taskURI(r.Host, t)
		i++
	}
	sort.Sort(tasks)
	respond(200, tasks, w)
}

func taskURI(host string, t core.Task) string {
	return fmt.Sprintf("%s://%s/v1/tasks/%s", "http", host, t.ID())
}
