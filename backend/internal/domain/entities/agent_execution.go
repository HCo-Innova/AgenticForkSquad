package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

// ExecutionStatus represents the state of an agent execution lifecycle.
type ExecutionStatus string

const (
	ExecutionRunning   ExecutionStatus = "running"
	ExecutionCompleted ExecutionStatus = "completed"
	ExecutionFailed    ExecutionStatus = "failed"
)

// AgentExecution represents a single execution instance of an AI agent
// operating on a given task.
type AgentExecution struct {
	ID          int64
	TaskID      int64
	AgentType   values.AgentType
	ForkID      string
	Status      ExecutionStatus
	StartedAt   time.Time
	CompletedAt *time.Time
	ErrorMsg    string
}

// Validate checks whether the entity satisfies the domain business rules.
// It should be called before persisting or processing the execution.
func (ae *AgentExecution) Validate() error {
	if ae.TaskID <= 0 {
		return errors.New("task_id must be a positive integer")
	}

	switch ae.AgentType {
	case values.AgentCerebro, values.AgentOperativo, values.AgentBulk:
		// valid type
	default:
		return errors.New("invalid agent type")
	}

	switch ae.Status {
	case ExecutionRunning, ExecutionCompleted, ExecutionFailed:
		// valid status
	default:
		return errors.New("invalid execution status")
	}

	if strings.TrimSpace(ae.ForkID) == "" {
		return errors.New("fork_id cannot be empty")
	}

	if ae.Status == ExecutionCompleted && ae.CompletedAt == nil {
		return errors.New("completed_at must be set when status is completed")
	}

	if ae.Status == ExecutionFailed && ae.ErrorMsg == "" {
		return errors.New("error message must be provided when status is failed")
	}

	return nil
}

// IsTerminal returns true if the execution has finished, either successfully or with failure.
func (ae *AgentExecution) IsTerminal() bool {
	return ae.Status == ExecutionCompleted || ae.Status == ExecutionFailed
}

// Duration returns the elapsed time of the execution if completed,
// or zero if not yet completed.
func (ae *AgentExecution) Duration() time.Duration {
	if ae.CompletedAt == nil {
		return 0
	}
	return ae.CompletedAt.Sub(ae.StartedAt)
}
