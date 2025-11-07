package usecases

import (
	"context"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
)

type fakeAgent struct{}
func (f *fakeAgent) AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (agents.AnalysisResult, error) {
	return agents.AnalysisResult{}, nil
}
func (f *fakeAgent) ProposeOptimization(ctx context.Context, analysis agents.AnalysisResult, forkID string) (*entities.OptimizationProposal, error) {
	return &entities.OptimizationProposal{}, nil
}
func (f *fakeAgent) RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error) {
	return nil, nil
}

type mockFactory struct{}
func (m *mockFactory) New(t values.AgentType) (agents.Agent, error) { return &fakeAgent{}, nil }

func TestRouter_SimpleQuery(t *testing.T) {
	factory := &mockFactory{}
	r := NewRouter(factory)
	task := &entities.Task{TargetQuery: "SELECT * FROM orders"}
	agentsList, err := r.SelectAgents(context.Background(), task)
	if err != nil { t.Fatalf("SelectAgents err: %v", err) }
	if len(agentsList) != 1 { t.Fatalf("expected 1 agent, got %d", len(agentsList)) }
}

func TestRouter_JoinQuery(t *testing.T) {
	factory := &mockFactory{}
	r := NewRouter(factory)
	task := &entities.Task{TargetQuery: "SELECT * FROM orders JOIN users ON users.id=orders.user_id"}
	agentsList, err := r.SelectAgents(context.Background(), task)
	if err != nil { t.Fatalf("SelectAgents err: %v", err) }
	if len(agentsList) < 2 { t.Fatalf("expected at least 2 agents for JOIN, got %d", len(agentsList)) }
}

func TestRouter_HighPriority(t *testing.T) {
	factory := &mockFactory{}
	r := NewRouter(factory)
	task := &entities.Task{TargetQuery: "SELECT * FROM orders", Metadata: map[string]interface{}{"priority": "high"}}
	agentsList, err := r.SelectAgents(context.Background(), task)
	if err != nil { t.Fatalf("SelectAgents err: %v", err) }
	if len(agentsList) != 3 { t.Fatalf("expected 3 agents for high priority, got %d", len(agentsList)) }
}
