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

// Orchestrator - Fase 1: ejecución paralela de agentes
// Dependencias incluidas para futuras fases, aquí sólo se usan Router/Factory opcionalmente.
type Orchestrator struct {
	Router       interface{}
	AgentFactory interface{}
	MCPClient    mcpFullPort
	Config       interface{}
}

func NewOrchestrator() *Orchestrator { return &Orchestrator{} }

// ExecuteAgentsInParallel ejecuta N agentes en paralelo y recopila propuestas y benchmarks.
// Errores parciales son tolerados (si al menos uno entrega resultados).
func (o *Orchestrator) ExecuteAgentsInParallel(ctx context.Context, task *entities.Task, ags []agents.Agent) ([]*entities.OptimizationProposal, []*entities.BenchmarkResult, error) {
	if task == nil { return nil, nil, errors.New("task is required") }
	if len(ags) == 0 { return nil, nil, errors.New("no agents provided") }

	var wg sync.WaitGroup
	propCh := make(chan *entities.OptimizationProposal, len(ags))
	benchCh := make(chan []*entities.BenchmarkResult, len(ags))
	errCh := make(chan error, len(ags))

	for _, a := range ags {
		wg.Add(1)
		go func(a agents.Agent) {
			defer wg.Done()
			// Timeout por agente: 10 minutos
			aCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel()

			// CreateFork y demás pasos se simulan a nivel de Agent (BaseAgent lo maneja),
			// en esta fase sólo llamamos a Analyze/Propose/RunBenchmark secuencialmente dentro de la goroutine.
			analysis, err := a.AnalyzeTask(aCtx, task, "fork-mock")
			if err != nil { errCh <- err; return }
			prop, err := a.ProposeOptimization(aCtx, analysis, "fork-mock")
			if err != nil { errCh <- err; return }
			// asignar ID simulado si no viene
			if prop.ID == 0 { prop.ID = time.Now().UnixNano() }
			res, err := a.RunBenchmark(aCtx, prop, "fork-mock")
			if err != nil { errCh <- err; return }
			propCh <- prop
			benchCh <- res
		}(a)
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

// mcpFullPort abstracts MCP operations needed for apply/cleanup.
type mcpFullPort interface {
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
func (o *Orchestrator) ExecuteTask(ctx context.Context, task *entities.Task, ags []agents.Agent, ce *ConsensusEngine, mainService string, forkIDs []string) (*entities.ConsensusDecision, error) {
	if task == nil || ce == nil { return nil, errors.New("invalid inputs") }
	props, benches, err := o.ExecuteAgentsInParallel(ctx, task, ags)
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
