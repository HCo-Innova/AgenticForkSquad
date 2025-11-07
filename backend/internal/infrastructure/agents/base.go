package agents

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
)

// mcpClientPort abstracts the MCP client used by agents.
type mcpClientPort interface {
	CreateFork(ctx context.Context, parent, name string) (string, error)
	DeleteFork(ctx context.Context, serviceID string) error
}

// BaseAgent provides shared agent functionality.
type BaseAgent struct {
	MCP       mcpClientPort
	LLM       llm.LLMClient
	Cfg       *cfgpkg.Config
	Repo      domainif.AgentExecutionRepository
	AgentType values.AgentType
}

// CreateFork creates a fork with standard naming and registers an AgentExecution.
func (a *BaseAgent) CreateFork(ctx context.Context, taskID int64) (string, error) {
	if a == nil || a.MCP == nil || a.Cfg == nil || a.Repo == nil {
		return "", errors.New("base agent not properly initialized")
	}
	parent := a.Cfg.TigerCloud.MainService
	name := a.forkName(taskID)
	forkID, err := a.MCP.CreateFork(ctx, parent, name)
	if err != nil {
		return "", err
	}
	// Register execution
	exec := &entities.AgentExecution{
		TaskID:    taskID,
		AgentType: a.AgentType,
		ForkID:    forkID,
		Status:    entities.ExecutionRunning,
		StartedAt: time.Now().UTC(),
	}
	if err := a.Repo.Create(ctx, exec); err != nil {
		return "", err
	}
	return forkID, nil
}

// DestroyFork deletes a fork by ID.
func (a *BaseAgent) DestroyFork(ctx context.Context, forkID string) error {
	if a == nil || a.MCP == nil {
		return errors.New("base agent not properly initialized")
	}
	if forkID == "" {
		return errors.New("forkID required")
	}
	return a.MCP.DeleteFork(ctx, forkID)
}

// forkName generates: afs-fork-{agent}-task{id}-{timestamp}
func (a *BaseAgent) forkName(taskID int64) string {
	ts := time.Now().UTC().Unix()
	return fmt.Sprintf("afs-fork-%s-task%d-%d", string(a.AgentType), taskID, ts)
}

// IsValidForkName validates the naming convention.
func IsValidForkName(name string) bool {
	re := regexp.MustCompile(`^afs-fork-[a-z0-9]+-task[0-9]+-[0-9]{10}$`)
	return re.MatchString(name)
}
