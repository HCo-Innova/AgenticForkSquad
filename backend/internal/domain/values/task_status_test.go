package values

import "testing"

func TestTaskStatusValidation(t *testing.T) {
	valid := []TaskStatus{TaskPending, TaskInProgress, TaskCompleted, TaskFailed}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if (TaskStatus("invalid").IsValid()) {
		t.Error("expected invalid status to be false")
	}
}
