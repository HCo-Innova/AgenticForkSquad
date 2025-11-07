package handlers

import (
    "context"
    "encoding/json"
    "errors"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/tuusuario/afs-challenge/internal/domain/entities"
    domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
    "github.com/tuusuario/afs-challenge/internal/domain/values"
)

type fakeExecRepo struct{ domainif.AgentExecutionRepository }

func TestResultsHandlers_RepoErrors(t *testing.T) {
    app := fiber.New()
    // agents
    h1 := &ResultsHandler{ExecRepo: &errExecRepo{}}
    app.Get("/api/v1/tasks/:id/agents", h1.GetTaskAgents)
    resp, _ := app.Test(httptest.NewRequest("GET", "/api/v1/tasks/1/agents", nil))
    if resp.StatusCode != 500 { t.Fatalf("expected 500 agents, got %d", resp.StatusCode) }

    // proposals
    h2 := &ResultsHandler{OptRepo: &errOptRepo{}}
    app.Get("/api/v1/tasks/:id/proposals", h2.GetTaskProposals)
    resp, _ = app.Test(httptest.NewRequest("GET", "/api/v1/tasks/1/proposals", nil))
    if resp.StatusCode != 500 { t.Fatalf("expected 500 proposals, got %d", resp.StatusCode) }

    // benchmarks
    h3 := &ResultsHandler{BenchRepo: &errBenchRepo{}}
    app.Get("/api/v1/proposals/:id/benchmarks", h3.GetProposalBenchmarks)
    resp, _ = app.Test(httptest.NewRequest("GET", "/api/v1/proposals/10/benchmarks", nil))
    if resp.StatusCode != 500 { t.Fatalf("expected 500 benchmarks, got %d", resp.StatusCode) }

    // consensus
    h4 := &ResultsHandler{ConsRepo: &errConsRepo{}}
    app.Get("/api/v1/tasks/:id/consensus", h4.GetTaskConsensus)
    resp, _ = app.Test(httptest.NewRequest("GET", "/api/v1/tasks/1/consensus", nil))
    if resp.StatusCode != 404 { t.Fatalf("expected 404 consensus, got %d", resp.StatusCode) }
}
type fakeOptRepo struct{ domainif.OptimizationRepository }
type fakeBenchRepo struct{ domainif.BenchmarkRepository }
type fakeConsRepo struct{ domainif.ConsensusRepository }
type errExecRepo struct{ domainif.AgentExecutionRepository }
type errOptRepo struct{ domainif.OptimizationRepository }
type errBenchRepo struct{ domainif.BenchmarkRepository }
type errConsRepo struct{ domainif.ConsensusRepository }

// erroring repo methods (use errors import)
func (e *errExecRepo) GetByTaskID(_ context.Context, taskID int) ([]*entities.AgentExecution, error) {
    return nil, errors.New("repo error")
}
func (e *errOptRepo) GetByAgentExecutionID(_ context.Context, execID int) ([]*entities.OptimizationProposal, error) {
    return nil, errors.New("repo error")
}
func (e *errBenchRepo) GetByProposalID(_ context.Context, proposalID int) ([]*entities.BenchmarkResult, error) {
    return nil, errors.New("repo error")
}
func (e *errConsRepo) GetByTaskID(_ context.Context, taskID int) (*entities.ConsensusDecision, error) {
    return nil, errors.New("repo error")
}

func (f *fakeExecRepo) GetByTaskID(_ context.Context, taskID int) ([]*entities.AgentExecution, error) {
    now := time.Now().UTC()
    return []*entities.AgentExecution{{
        ID: 1,
        TaskID: int64(taskID),
        AgentType: values.AgentCerebro,
        ForkID: "fork-1",
        Status: entities.ExecutionCompleted,
        StartedAt: now,
        CompletedAt: &now,
    }}, nil
}

func (f *fakeOptRepo) GetByAgentExecutionID(_ context.Context, execID int) ([]*entities.OptimizationProposal, error) {
    now := time.Now().UTC()
    return []*entities.OptimizationProposal{{
        ID: 10,
        AgentExecutionID: int64(execID),
        ProposalType: values.ProposalIndex,
        SQLCommands: []string{"CREATE INDEX idx ON t(c)"},
        Rationale: "create index",
        EstimatedImpact: entities.EstimatedImpact{
            QueryTimeImprovement: 50,
            StorageOverheadMB: 1,
            Complexity: "low",
            Risk: "low",
        },
        CreatedAt: now,
    }}, nil
}

func (f *fakeBenchRepo) GetByProposalID(_ context.Context, proposalID int) ([]*entities.BenchmarkResult, error) {
    now := time.Now().UTC()
    return []*entities.BenchmarkResult{{
        ID: 100,
        ProposalID: int64(proposalID),
        QueryName: entities.QueryNameBaseline,
        QueryExecuted: "SELECT 1",
        ExecutionTimeMS: 12.3,
        RowsReturned: 1,
        ExplainPlan: entities.ExplainPlan{PlanType: "Index Scan"},
        StorageImpactMB: 0.0,
        CreatedAt: now,
    }}, nil
}

func (f *fakeConsRepo) GetByTaskID(_ context.Context, taskID int) (*entities.ConsensusDecision, error) {
    now := time.Now().UTC()
    return &entities.ConsensusDecision{
        ID: 200,
        TaskID: int64(taskID),
        WinningProposalID: nil,
        AllScores: map[values.AgentType]entities.ProposalScore{
            values.AgentCerebro: {ProposalID: 10, Performance: 90, Storage: 80, Complexity: 90, Risk: 90},
        },
        DecisionRationale: "ok",
        AppliedToMain: false,
        CreatedAt: now,
    }, nil
}

func TestResultsHandlers_ValidationErrors(t *testing.T) {
    app := fiber.New()
    h := &ResultsHandler{}
    app.Get("/api/v1/tasks/:id/agents", h.GetTaskAgents)
    app.Get("/api/v1/tasks/:id/proposals", h.GetTaskProposals)
    app.Get("/api/v1/proposals/:id/benchmarks", h.GetProposalBenchmarks)
    app.Get("/api/v1/tasks/:id/consensus", h.GetTaskConsensus)

    req := httptest.NewRequest("GET", "/api/v1/tasks/abc/agents", nil)
    resp, _ := app.Test(req)
    if resp.StatusCode != 500 && resp.StatusCode != 400 { t.Fatalf("expected 400/500, got %d", resp.StatusCode) }

    req = httptest.NewRequest("GET", "/api/v1/tasks/abc/proposals", nil)
    resp, _ = app.Test(req)
    if resp.StatusCode != 500 && resp.StatusCode != 400 { t.Fatalf("expected 400/500, got %d", resp.StatusCode) }

    req = httptest.NewRequest("GET", "/api/v1/proposals/abc/benchmarks", nil)
    resp, _ = app.Test(req)
    if resp.StatusCode != 500 && resp.StatusCode != 400 { t.Fatalf("expected 400/500, got %d", resp.StatusCode) }

    req = httptest.NewRequest("GET", "/api/v1/tasks/abc/consensus", nil)
    resp, _ = app.Test(req)
    if resp.StatusCode != 500 && resp.StatusCode != 400 { t.Fatalf("expected 400/500, got %d", resp.StatusCode) }
}

func TestResultsHandlers_HappyPaths(t *testing.T) {
    app := fiber.New()
    h := &ResultsHandler{ExecRepo: &fakeExecRepo{}, OptRepo: &fakeOptRepo{}, BenchRepo: &fakeBenchRepo{}, ConsRepo: &fakeConsRepo{}}
    app.Get("/api/v1/tasks/:id/agents", h.GetTaskAgents)
    app.Get("/api/v1/tasks/:id/proposals", h.GetTaskProposals)
    app.Get("/api/v1/proposals/:id/benchmarks", h.GetProposalBenchmarks)
    app.Get("/api/v1/tasks/:id/consensus", h.GetTaskConsensus)

    // agents
    req := httptest.NewRequest("GET", "/api/v1/tasks/1/agents", nil)
    resp, err := app.Test(req)
    if err != nil || resp.StatusCode != 200 { t.Fatalf("agents failed: %v code=%d", err, resp.StatusCode) }

    // proposals
    req = httptest.NewRequest("GET", "/api/v1/tasks/1/proposals", nil)
    resp, err = app.Test(req)
    if err != nil || resp.StatusCode != 200 { t.Fatalf("proposals failed: %v code=%d", err, resp.StatusCode) }

    // benchmarks
    req = httptest.NewRequest("GET", "/api/v1/proposals/10/benchmarks", nil)
    resp, err = app.Test(req)
    if err != nil || resp.StatusCode != 200 { t.Fatalf("benchmarks failed: %v code=%d", err, resp.StatusCode) }

    // consensus
    req = httptest.NewRequest("GET", "/api/v1/tasks/1/consensus", nil)
    resp, err = app.Test(req)
    if err != nil || resp.StatusCode != 200 { t.Fatalf("consensus failed: %v code=%d", err, resp.StatusCode) }

    // verify JSON parses
    var m map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&m); err != nil { t.Fatalf("invalid json: %v", err) }
}
