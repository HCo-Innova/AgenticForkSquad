package values

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskPending, TaskInProgress, TaskCompleted, TaskFailed:
		return true
	default:
		return false
	}
}

func (s TaskStatus) String() string {
	return string(s)
}
