package entities

import (
	"testing"
	"time"
)

func TestValidate_ValidTask(t *testing.T) {
	task := &Task{
		Type:        TaskTypeQueryOptimization,
		Description: "Optimize SELECT query",
		TargetQuery: "SELECT * FROM users;",
		Status:      TaskStatusPending,
	}

	if err := task.Validate(); err != nil {
		t.Fatalf("expected valid task, got error: %v", err)
	}
}

func TestValidate_InvalidProposal(t *testing.T) {
	task := &Task{
		Type:        "invalid_type",
		Description: "Invalid type test",
		TargetQuery: "SELECT * FROM test;",
	}

	if err := task.Validate(); err == nil {
		t.Fatal("expected error for invalid task type, got nil")
	}
}

func TestValidate_EmptyQuery(t *testing.T) {
	task := &Task{
		Type:        TaskTypeSchemaImprovement,
		TargetQuery: "",
	}

	if err := task.Validate(); err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

func TestValidate_TooLongDescription(t *testing.T) {
	desc := make([]byte, 501)
	for i := range desc {
		desc[i] = 'a'
	}

	task := &Task{
		Type:        TaskTypeIndexTuning,
		Description: string(desc),
		TargetQuery: "SELECT * FROM orders;",
	}

	if err := task.Validate(); err == nil {
		t.Fatal("expected error for long description, got nil")
	}
}

func TestCanTransitionTo(t *testing.T) {
	task := &Task{Status: TaskStatusPending}

	if !task.CanTransitionTo(TaskStatusInProgress) {
		t.Error("expected transition from pending → in_progress to be valid")
	}
	if task.CanTransitionTo(TaskStatusCompleted) {
		t.Error("expected transition from pending → completed to be invalid")
	}

	task.Status = TaskStatusInProgress
	if !task.CanTransitionTo(TaskStatusCompleted) {
		t.Error("expected transition from in_progress → completed to be valid")
	}
	if !task.CanTransitionTo(TaskStatusFailed) {
		t.Error("expected transition from in_progress → failed to be valid")
	}
}

func TestIsComplete(t *testing.T) {
	task := &Task{Status: TaskStatusCompleted}
	if !task.IsComplete() {
		t.Error("expected completed task to return true")
	}

	task = &Task{Status: TaskStatusPending}
	if task.IsComplete() {
		t.Error("expected pending task to return false")
	}
}

func TestValidate_DefaultStatus(t *testing.T) {
	task := &Task{
		Type:        TaskTypeQueryOptimization,
		Description: "Missing status test",
		TargetQuery: "SELECT * FROM test;",
	}

	err := task.Validate()
	if err != nil {
		t.Fatalf("expected valid task with default status, got: %v", err)
	}

	if task.Status != TaskStatusPending {
		t.Errorf("expected default status 'pending', got %s", task.Status)
	}
}

func TestTask_Timestamps(t *testing.T) {
	now := time.Now()
	task := &Task{
		ID:          1,
		Type:        TaskTypeSchemaImprovement,
		Description: "Check timestamps",
		TargetQuery: "SELECT 1;",
		Status:      TaskStatusCompleted,
		CreatedAt:   now,
		CompletedAt: &now,
	}

	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if task.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}
