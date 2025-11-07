package agents

import (
	"context"
	"testing"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

type mockMCP struct{
	created []struct{parent,name string}
	deleted []string
	forkID  string
	err     error
}

func (m *mockMCP) CreateFork(ctx context.Context, parent, name string) (string, error) {
	m.created = append(m.created, struct{parent,name string}{parent,name})
	if m.err != nil { return "", m.err }
	if m.forkID == "" { m.forkID = "afs-fork-cerebro-task1-1234567890" }
	return m.forkID, nil
}
func (m *mockMCP) DeleteFork(ctx context.Context, serviceID string) error {
	m.deleted = append(m.deleted, serviceID)
	return m.err
}

// Mock repository implementing the real domain interface
type mockAgentExecutionRepo struct{ created int }
func (r *mockAgentExecutionRepo) Create(ctx context.Context, exec *entities.AgentExecution) error {
	r.created++
	return nil
}
func (r *mockAgentExecutionRepo) GetByID(ctx context.Context, id int) (*entities.AgentExecution, error) {
	return nil, nil
}
func (r *mockAgentExecutionRepo) GetByTaskID(ctx context.Context, taskID int) ([]*entities.AgentExecution, error) {
	return nil, nil
}
func (r *mockAgentExecutionRepo) Update(ctx context.Context, exec *entities.AgentExecution) error { return nil }

// Now test BaseAgent
func TestBaseAgent_CreateFork_Naming(t *testing.T) {
	mcp := &mockMCP{forkID: ""}
	cfg := &cfgpkg.Config{}
	cfg.TigerCloud.MainService = "afs-main"

	repo := &mockAgentExecutionRepo{}

	a := &BaseAgent{
		MCP:       mcp,
		LLM:       nil,
		Cfg:       cfg,
		Repo:      repo,
		AgentType: values.AgentOperativo,
	}
	fork, err := a.CreateFork(context.Background(), 123)
	if err != nil { t.Fatalf("CreateFork err: %v", err) }
	if !IsValidForkName(fork) {
		t.Fatalf("fork name does not match convention: %s", fork)
	}
	if repo.created == 0 { t.Fatalf("expected repo Create to be called") }
}

func TestBaseAgent_DestroyFork(t *testing.T) {
	mcp := &mockMCP{forkID: "id"}
	a := &BaseAgent{MCP: mcp}
	if err := a.DestroyFork(context.Background(), "id"); err != nil {
		t.Fatalf("DestroyFork err: %v", err)
	}
}
