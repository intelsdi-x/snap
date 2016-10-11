package v2

import "github.com/intelsdi-x/snap/mgmt/rest/rbody"

type ScheduledTaskListReturned struct {
	ScheduledTasks []rbody.ScheduledTask `json:"scheduled_task,omitempty"`
}

func (s *ScheduledTaskListReturned) Len() int {
	return len(s.ScheduledTasks)
}

func (s *ScheduledTaskListReturned) Less(i, j int) bool {
	return s.ScheduledTasks[j].CreationTime().After(s.ScheduledTasks[i].CreationTime())
}

func (s *ScheduledTaskListReturned) Swap(i, j int) {
	s.ScheduledTasks[i], s.ScheduledTasks[j] = s.ScheduledTasks[j], s.ScheduledTasks[i]
}

func (s *ScheduledTaskListReturned) ResponseBodyMessage() string {
	return "Scheduled tasks retrieved"
}

func (s *ScheduledTaskListReturned) ResponseBodyType() string {
	return "scheduled_task_list_returned"
}
