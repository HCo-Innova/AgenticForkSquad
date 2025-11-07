package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
)

type mockAgentOK struct{ id int64 }
func (m *mockAgentOK) AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (agents.AnalysisResult, error) {
	return agents.AnalysisResult{Insights: []string{"ok"}}, nil
}
func (m *mockAgentOK) ProposeOptimization(ctx context.Context, analysis agents.AnalysisResult, forkID string) (*entities.OptimizationProposal, error) {
	return &entities.OptimizationProposal{ID: m.id, SQLCommands: []string{"SELECT 1"}}, nil
}
func (m *mockAgentOK) RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error) {
	return []*entities.BenchmarkResult{{ProposalID: proposal.ID, QueryName: entities.QueryNameBaseline, ExecutionTimeMS: 1}}, nil
}

type mockAgentFail struct{}
func (m *mockAgentFail) AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (agents.AnalysisResult, error) { return agents.AnalysisResult{}, errors.New("fail") }
func (m *mockAgentFail) ProposeOptimization(ctx context.Context, analysis agents.AnalysisResult, forkID string) (*entities.OptimizationProposal, error) { return nil, errors.New("fail") }
func (m *mockAgentFail) RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error) { return nil, errors.New("fail") }

func TestOrchestratorParallel(t *testing.T) {
	orch := NewOrchestrator()
	ag1 := &mockAgentOK{id: 1}
	ag2 := &mockAgentOK{id: 2}
	ag3 := &mockAgentFail{}

	task := &entities.Task{Type: entities.TaskTypeQueryOptimization, TargetQuery: "SELECT * FROM orders"}
	props, benches, err := orch.ExecuteAgentsInParallel(context.Background(), task, []agents.Agent{ag1, ag2, ag3})
	if err != nil { t.Fatalf("unexpected err (partial failures allowed): %v", err) }
	if len(props) != 2 { t.Fatalf("expected 2 proposals, got %d", len(props)) }
	if len(benches) != 2 { t.Fatalf("expected 2 benchmark result sets flattened, got %d", len(benches)) }
}
