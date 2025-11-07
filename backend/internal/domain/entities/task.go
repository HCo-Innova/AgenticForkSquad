package entities

import (
	"errors"
	"strings"
	"time"
)

// TaskType represents the type of task to be executed.
// Enumerated types help ensure only valid task categories are used.
type TaskType string

const (
	TaskTypeQueryOptimization TaskType = "query_optimization"
	TaskTypeSchemaImprovement TaskType = "schema_improvement"
	TaskTypeIndexTuning       TaskType = "index_tuning"
	TaskTypePartitioning      TaskType = "partitioning"
)

// TaskStatus represents the current lifecycle status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Task represents a single optimization or analysis request in the system.
// It is the core entity for orchestration and agent execution tracking.
type Task struct {
	ID          int64
	Type        TaskType
	Description string
	TargetQuery string
	Status      TaskStatus
	CreatedAt   time.Time
	CompletedAt *time.Time
	Metadata    map[string]interface{}
}

// Validate checks whether the Task entity satisfies all business rules.
// It should be called before persisting or processing a task.
func (t *Task) Validate() error {
	if t.Type == "" {
		return errors.New("task type cannot be empty")
	}

	switch t.Type {
	case TaskTypeQueryOptimization, TaskTypeSchemaImprovement, TaskTypeIndexTuning, TaskTypePartitioning:
		// valid type
	default:
		return errors.New("invalid task type")
	}

	if strings.TrimSpace(t.TargetQuery) == "" {
		return errors.New("target query cannot be empty")
	}

	if len(t.Description) > 500 {
		return errors.New("description exceeds 500 characters")
	}

	if t.Status == "" {
		t.Status = TaskStatusPending
	}

	switch t.Status {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusFailed:
		// valid status
	default:
		return errors.New("invalid task status")
	}

	return nil
}

// CanTransitionTo validates whether the task can move to a given new status
// according to the defined business rules.
func (t *Task) CanTransitionTo(newStatus TaskStatus) bool {
	validTransitions := map[TaskStatus][]TaskStatus{
		TaskStatusPending:    {TaskStatusInProgress, TaskStatusFailed},
		TaskStatusInProgress: {TaskStatusCompleted, TaskStatusFailed},
		TaskStatusCompleted:  {},
		TaskStatusFailed:     {},
	}

	next, ok := validTransitions[t.Status]
	if !ok {
		return false
	}

	for _, s := range next {
		if s == newStatus {
			return true
		}
	}
	return false
}

// IsComplete returns true if the task has reached a terminal successful state.
func (t *Task) IsComplete() bool {
	return t.Status == TaskStatusCompleted
}
