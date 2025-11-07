package entities

import (
	"testing"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

func TestValidate_ValidExecution(t *testing.T) {
	now := time.Now()
	exec := &AgentExecution{
		ID:          1,
		TaskID:      10,
		AgentType:   values.AgentOperativo,
		ForkID:      "fork123",
		Status:      ExecutionCompleted,
		StartedAt:   now.Add(-2 * time.Minute),
		CompletedAt: &now,
	}

	if err := exec.Validate(); err != nil {
		t.Fatalf("expected valid execution, got error: %v", err)
	}
}

func TestValidate_InvalidAgentType(t *testing.T) {
	e := &AgentExecution{
		TaskID:    1,
		AgentType: "invalid",
		ForkID:    "f123",
		Status:    ExecutionRunning,
	}
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for invalid agent type, got nil")
	}
}

func TestValidate_MissingForkID(t *testing.T) {
	e := &AgentExecution{
		TaskID:    1,
		AgentType: values.AgentCerebro,
		ForkID:    "",
		Status:    ExecutionRunning,
	}
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for missing fork_id, got nil")
	}
}

func TestValidate_CompletedWithoutTimestamp(t *testing.T) {
	e := &AgentExecution{
		TaskID:    1,
		AgentType: values.AgentOperativo,
		ForkID:    "f123",
		Status:    ExecutionCompleted,
	}
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for missing completed_at, got nil")
	}
}

func TestValidate_FailedWithoutError(t *testing.T) {
	e := &AgentExecution{
		TaskID:    1,
		AgentType: values.AgentOperativo,
		ForkID:    "f123",
		Status:    ExecutionFailed,
	}
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for failed status without error message")
	}
}

func TestIsTerminal(t *testing.T) {
	e := &AgentExecution{Status: ExecutionRunning}
	if e.IsTerminal() {
		t.Error("expected running execution not to be terminal")
	}

	e.Status = ExecutionCompleted
	if !e.IsTerminal() {
		t.Error("expected completed execution to be terminal")
	}
}

func TestDuration(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now()
	e := &AgentExecution{
		StartedAt:   start,
		CompletedAt: &end,
	}

	if d := e.Duration(); d <= 0 {
		t.Error("expected positive duration for completed execution")
	}
}