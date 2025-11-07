package agents

import (
	"context"
	"fmt"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// Agent is the common contract that all concrete agents implement.
// It matches the Analyze/Propose/Benchmark trio used throughout the system.
type Agent interface {
	AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (AnalysisResult, error)
	ProposeOptimization(ctx context.Context, analysis AnalysisResult, forkID string) (*entities.OptimizationProposal, error)
	RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error)
}

// NewAgent constructs an agent instance for the given type, wiring dependencies.
// Supported types:
// - AgentCerebro (gemini-2.5-pro): Planner/QA
// - AgentOperativo (gemini-2.5-flash): Generación/Ejecución
// - AgentBulk (gemini-2.0-flash): Bajo costo/masivas
func NewAgent(agentType values.AgentType, mcpClient *mcp.MCPClient, llmClient llm.LLMClient, cfg *cfgpkg.Config) (Agent, error) {
	base := &BaseAgent{
		MCP:       mcpClient,
		LLM:       llmClient,
		Cfg:       cfg,
		AgentType: agentType,
	}
	switch agentType {
	case values.AgentCerebro:
		return &CerebroAgent{Base: base, MCPQ: mcpClient, LLM: llmClient}, nil
	case values.AgentOperativo:
		return &OperativoAgent{Base: base, MCPQ: mcpClient, LLM: llmClient}, nil
	case values.AgentBulk:
		return &CerebroAgent{Base: base, MCPQ: mcpClient, LLM: llmClient}, nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s (supported: cerebro, operativo, bulk)", agentType)
	}
}
