package interfaces

import (
	"context"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id int) (*entities.Task, error)
	List(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error)
	Update(ctx context.Context, task *entities.Task) error
}

type AgentExecutionRepository interface {
	Create(ctx context.Context, exec *entities.AgentExecution) error
	GetByID(ctx context.Context, id int) (*entities.AgentExecution, error)
	GetByTaskID(ctx context.Context, taskID int) ([]*entities.AgentExecution, error)
	Update(ctx context.Context, exec *entities.AgentExecution) error
}

type OptimizationRepository interface {
	Create(ctx context.Context, proposal *entities.OptimizationProposal) error
	GetByID(ctx context.Context, id int) (*entities.OptimizationProposal, error)
	GetByAgentExecutionID(ctx context.Context, execID int) ([]*entities.OptimizationProposal, error)
	Update(ctx context.Context, proposal *entities.OptimizationProposal) error
}

type BenchmarkRepository interface {
	Create(ctx context.Context, result *entities.BenchmarkResult) error
	GetByProposalID(ctx context.Context, proposalID int) ([]*entities.BenchmarkResult, error)
}

type ConsensusRepository interface {
	Create(ctx context.Context, decision *entities.ConsensusDecision) error
	GetByTaskID(ctx context.Context, taskID int) (*entities.ConsensusDecision, error)
	Update(ctx context.Context, decision *entities.ConsensusDecision) error
}
