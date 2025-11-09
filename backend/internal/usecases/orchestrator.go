package usecases

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/agents"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// Agent is the common interface that all agents must implement
// Re-exported from agents package for convenience
type Agent = agents.Agent

// AnalysisResult is the result of an agent's analysis phase
type AnalysisResult = agents.AnalysisResult

// Orchestrator - Fase 1: ejecución paralela de agentes
// Dependencias incluidas para futuras fases, aquí sólo se usan Router/Factory opcionalmente.
type Orchestrator struct {
	Router       interface{}
	AgentFactory interface{}
	MCPClient    mcpFullPort
	Config       interface{}
}

func NewOrchestrator() *Orchestrator { return &Orchestrator{} }

// ExecuteAgentsInParallel ejecuta N agentes en paralelo con fork IDs reales y recopila propuestas y benchmarks.
// Errores parciales son tolerados (si al menos uno entrega resultados).
func (o *Orchestrator) ExecuteAgentsInParallel(ctx context.Context, task *entities.Task, ags []Agent, forkIDs []string) ([]*entities.OptimizationProposal, []*entities.BenchmarkResult, error) {
	if task == nil { return nil, nil, errors.New("task is required") }
	if len(ags) == 0 { return nil, nil, errors.New("no agents provided") }
	if len(ags) != len(forkIDs) { return nil, nil, errors.New("agents and forkIDs count mismatch") }

	var wg sync.WaitGroup
	propCh := make(chan *entities.OptimizationProposal, len(ags))
	benchCh := make(chan []*entities.BenchmarkResult, len(ags))
	errCh := make(chan error, len(ags))

	for i, a := range ags {
		wg.Add(1)
		go func(idx int, ag Agent, forkID string) {
			defer wg.Done()
			// Timeout por agente: 10 minutos
			aCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel()

			// Usar fork ID real en todas las operaciones del agente
			analysis, err := ag.AnalyzeTask(aCtx, task, forkID)
			if err != nil { errCh <- err; return }
			prop, err := ag.ProposeOptimization(aCtx, analysis, forkID)
			if err != nil { errCh <- err; return }
			// asignar ID simulado si no viene
			if prop.ID == 0 { prop.ID = time.Now().UnixNano() }
			res, err := ag.RunBenchmark(aCtx, prop, forkID)
			if err != nil { errCh <- err; return }
			propCh <- prop
			benchCh <- res
		}(i, a, forkIDs[i])
	}

	wg.Wait()
	close(propCh)
	close(benchCh)
	close(errCh)

	var proposals []*entities.OptimizationProposal
	var benchmarks []*entities.BenchmarkResult
	for p := range propCh { proposals = append(proposals, p) }
	for bs := range benchCh { benchmarks = append(benchmarks, bs...) }
	// si todos fallaron, devolver error
	if len(proposals) == 0 {
		var firstErr error
		for e := range errCh { firstErr = e; break }
		if firstErr == nil { firstErr = errors.New("all agents failed") }
		return nil, nil, firstErr
	}
	return proposals, benchmarks, nil
}

// mcpFullPort abstracts MCP operations needed for fork management, query execution and cleanup.
type mcpFullPort interface {
	CreateFork(ctx context.Context, parentServiceID, forkName string) (string, error)
	ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error)
	DeleteFork(ctx context.Context, serviceID string) error
}

// ApplyToMainDB aplica la propuesta ganadora en el servicio principal (main DB) vía MCP.
func (o *Orchestrator) ApplyToMainDB(ctx context.Context, mainService string, winning *entities.OptimizationProposal) error {
	if o == nil || o.MCPClient == nil { return errors.New("orchestrator not initialized") }
	if winning == nil || len(winning.SQLCommands) == 0 { return errors.New("invalid winning proposal") }
	if mainService == "" { return errors.New("main service required") }
	for _, stmt := range winning.SQLCommands {
		if stmt == "" { continue }
		if _, err := o.MCPClient.ExecuteQuery(ctx, mainService, stmt, 600000); err != nil {
			return err
		}
	}
	return nil
}

// CleanupForks elimina forks usados durante la ejecución.
func (o *Orchestrator) CleanupForks(ctx context.Context, forkIDs []string) error {
	if o == nil || o.MCPClient == nil { return errors.New("orchestrator not initialized") }
	for _, id := range forkIDs {
		if id == "" { continue }
		if err := o.MCPClient.DeleteFork(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTask ejecuta el flujo completo con dependencias provistas (simplificado para tests):
// agents ya seleccionados, consensus engine provisto, mainService y forkIDs a limpiar.
func (o *Orchestrator) ExecuteTask(ctx context.Context, task *entities.Task, ags []Agent, ce *ConsensusEngine, mainService string, forkIDs []string) (*entities.ConsensusDecision, error) {
	if task == nil || ce == nil { return nil, errors.New("invalid inputs") }
	props, benches, err := o.ExecuteAgentsInParallel(ctx, task, ags, forkIDs)
	if err != nil { return nil, err }
	// Scoring criteria por defecto
	criteria := entities.ScoringCriteria{PerformanceWeight: 0.5, StorageWeight: 0.2, ComplexityWeight: 0.2, RiskWeight: 0.1}
	dec, err := ce.Decide(ctx, props, benches, criteria)
	if err != nil { return nil, err }
	// Aplicar a main DB
	var winner *entities.OptimizationProposal
	if dec != nil && dec.WinningProposalID != nil {
		for _, p := range props {
			if p.ID == *dec.WinningProposalID { winner = p; break }
		}
	}
	if winner == nil { return nil, errors.New("winner not found") }
	if err := o.ApplyToMainDB(ctx, mainService, winner); err != nil { return nil, err }
	// Cleanup forks
	if err := o.CleanupForks(ctx, forkIDs); err != nil { return nil, err }
	return dec, nil
}
