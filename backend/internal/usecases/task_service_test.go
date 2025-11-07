package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
)

type mockTaskRepo struct{
	created []*entities.Task
	byID map[int]*entities.Task
	list []*entities.Task
	updateErr error
}

func (m *mockTaskRepo) Create(ctx context.Context, task *entities.Task) error {
	if task.ID == 0 { task.ID = int64(len(m.created)+1) }
	m.created = append(m.created, task)
	if m.byID == nil { m.byID = map[int]*entities.Task{} }
	m.byID[int(task.ID)] = task
	return nil
}
func (m *mockTaskRepo) GetByID(ctx context.Context, id int) (*entities.Task, error) {
	if m.byID == nil { return nil, errors.New("not found") }
	t, ok := m.byID[id]
	if !ok { return nil, errors.New("not found") }
	return t, nil
}
func (m *mockTaskRepo) List(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error) {
	return m.list, nil
}
func (m *mockTaskRepo) Update(ctx context.Context, task *entities.Task) error {
	if m.updateErr != nil { return m.updateErr }
	m.byID[int(task.ID)] = task
	return nil
}

func TestTaskService(t *testing.T) {
	repo := &mockTaskRepo{}
	svc := NewTaskService(repo)
	ctx := context.Background()

	// CreateTask
	task := &entities.Task{Type: entities.TaskTypeQueryOptimization, TargetQuery: "SELECT 1"}
	created, err := svc.CreateTask(ctx, task)
	if err != nil { t.Fatalf("CreateTask err: %v", err) }
	if created.ID == 0 { t.Fatalf("expected ID assigned") }

	// GetTask
	got, err := svc.GetTask(ctx, int(created.ID))
	if err != nil { t.Fatalf("GetTask err: %v", err) }
	if got.TargetQuery != "SELECT 1" { t.Fatalf("unexpected task") }

	// ListTasks
	repo.list = []*entities.Task{created}
	lst, err := svc.ListTasks(ctx, entities.TaskFilters{})
	if err != nil || len(lst) == 0 { t.Fatalf("ListTasks err=%v n=%d", err, len(lst)) }

	// UpdateTaskStatus valid transition Pending -> InProgress
	if err := svc.UpdateTaskStatus(ctx, int(created.ID), entities.TaskStatusInProgress); err != nil {
		t.Fatalf("UpdateTaskStatus err: %v", err)
	}
	// Invalid transition Completed -> Pending
	got.Status = entities.TaskStatusCompleted
	repo.byID[int(got.ID)] = got
	if err := svc.UpdateTaskStatus(ctx, int(got.ID), entities.TaskStatusPending); err == nil {
		t.Fatalf("expected invalid transition error")
	}
}
