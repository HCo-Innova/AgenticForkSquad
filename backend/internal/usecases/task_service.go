package usecases

import (
	"context"
	"errors"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
)

// TaskService coordinates task lifecycle operations applying business rules.
type TaskService struct {
	repo domainif.TaskRepository
}

func NewTaskService(repo domainif.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// CreateTask validates and persists a new task, returning it with ID set.
func (s *TaskService) CreateTask(ctx context.Context, task *entities.Task) (*entities.Task, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("task service not initialized")
	}
	if task == nil {
		return nil, errors.New("task is required")
	}
	if err := task.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// GetTask returns a task by its ID.
func (s *TaskService) GetTask(ctx context.Context, id int) (*entities.Task, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("task service not initialized")
	}
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

// ListTasks returns tasks matching the given filters.
func (s *TaskService) ListTasks(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("task service not initialized")
	}
	return s.repo.List(ctx, filters)
}

// UpdateTaskStatus validates transition and persists the new status.
func (s *TaskService) UpdateTaskStatus(ctx context.Context, id int, status entities.TaskStatus) error {
	if s == nil || s.repo == nil {
		return errors.New("task service not initialized")
	}
	if id <= 0 {
		return errors.New("invalid id")
	}
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}
	if !existing.CanTransitionTo(status) {
		return errors.New("invalid status transition")
	}
	existing.Status = status
	return s.repo.Update(ctx, existing)
}
