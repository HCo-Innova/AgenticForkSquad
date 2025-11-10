package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

// EventType representa el tipo de evento WebSocket
type EventType = string

// TaskProcessor orquesta el procesamiento completo de una tarea:
// 1. Asignar agentes
// 2. Crear forks
// 3. Ejecutar agentes en paralelo
// 4. Consenso
// 5. Aplicar solución
// 6. Limpiar forks
type TaskProcessor struct {
	taskRepo      domainif.TaskRepository
	agentExecRepo domainif.AgentExecutionRepository
	proposalRepo  domainif.OptimizationRepository
	benchmarkRepo domainif.BenchmarkRepository
	consensusRepo domainif.ConsensusRepository
	orchestrator  *Orchestrator
	consensus     *ConsensusEngine
	hub           *Hub
	agentFactory  *AgentFactory
	mainService   string
}

func NewTaskProcessor(
	taskRepo domainif.TaskRepository,
	agentExecRepo domainif.AgentExecutionRepository,
	proposalRepo domainif.OptimizationRepository,
	benchmarkRepo domainif.BenchmarkRepository,
	consensusRepo domainif.ConsensusRepository,
	orchestrator *Orchestrator,
	consensus *ConsensusEngine,
	hub *Hub,
	agentFactory *AgentFactory,
	mainService string,
) *TaskProcessor {
	return &TaskProcessor{
		taskRepo:      taskRepo,
		agentExecRepo: agentExecRepo,
		proposalRepo:  proposalRepo,
		benchmarkRepo: benchmarkRepo,
		consensusRepo: consensusRepo,
		orchestrator:  orchestrator,
		consensus:     consensus,
		hub:           hub,
		agentFactory:  agentFactory,
		mainService:   mainService,
	}
}

// ProcessTask ejecuta el flujo completo de procesamiento de una tarea
func (p *TaskProcessor) ProcessTask(ctx context.Context, taskID int64) error {
	// 1. Obtener tarea
	task, err := p.taskRepo.GetByID(ctx, int(taskID))
	if err != nil || task == nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 2. Actualizar estado a "in_progress" (routing)
	task.Status = entities.TaskStatusInProgress
	if err := p.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	p.broadcastEvent(EventAgentsAssigned, map[string]interface{}{
		"task_id": taskID,
		"status":  "routing",
	})

	// 3. Asignar agentes (3 agentes especializados)
	agentTypes := []values.AgentType{
		values.AgentCerebro,   // gemini-2.5-pro (Planner/QA)
		values.AgentOperativo, // gemini-2.5-flash (Generator)
		values.AgentBulk,      // gemini-2.0-flash (Bulk ops)
	}

	var agentInstances []Agent
	var forkIDs []string
	var agentExecutionIDs []int64

	// 4. Crear agentes y forks reales via MCP
	for _, agentType := range agentTypes {
		fmt.Printf("      🔨 Creating agent: %s\n", agentType)
		agent, err := p.agentFactory.CreateAgent(agentType)
		if err != nil {
			return fmt.Errorf("failed to create agent %s: %w", agentType, err)
		}
		agentInstances = append(agentInstances, agent)

		// Crear fork real usando MCP Client
		forkName := fmt.Sprintf("fork-%s-task%d", agentType, taskID)
		forkID, err := p.orchestrator.MCPClient.CreateFork(ctx, p.mainService, forkName)
		if err != nil {
			return fmt.Errorf("failed to create fork for %s: %w", agentType, err)
		}
		forkIDs = append(forkIDs, forkID)

		// Crear registro de agent_execution en DB
		agentExec := &entities.AgentExecution{
			TaskID:     int64(taskID),
			AgentType:  agentType,
			ForkID:     forkID,
			Status:     "running",
			StartedAt:  time.Now().UTC(),
		}
		if err := p.agentExecRepo.Create(ctx, agentExec); err != nil {
			return fmt.Errorf("failed to create agent execution record: %w", err)
		}
		fmt.Printf("      ✅ Created agent_execution ID=%d for task=%d agent=%s\n", agentExec.ID, taskID, agentType)
		agentExecutionIDs = append(agentExecutionIDs, agentExec.ID)

		p.broadcastEvent(EventForkCreated, map[string]interface{}{
			"task_id":    taskID,
			"agent_type": agentType,
			"fork_id":    forkID,
			"execution_id": agentExec.ID,
		})
	}

	// 5. Ejecutar agentes en paralelo usando orchestrator
	p.broadcastEvent(EventAnalysisCompleted, map[string]interface{}{
		"task_id": taskID,
		"status":  "executing",
		"agents":  len(agentInstances),
	})

	// 6. Ejecutar agentes en paralelo usando orchestrator con forks reales
	proposals, benchmarks, err := p.orchestrator.ExecuteAgentsInParallel(ctx, task, agentInstances, forkIDs, agentExecutionIDs)
	if err != nil {
		task.Status = entities.TaskStatusFailed
		p.taskRepo.Update(ctx, task)
		p.broadcastEvent(EventTaskFailed, map[string]interface{}{
			"task_id": taskID,
			"error":   err.Error(),
		})
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// 7. Guardar propuestas primero (para obtener IDs autogenerados por DB)
	proposalIDMap := make(map[int64]int64) // old ID -> new ID
	for i, prop := range proposals {
		oldID := prop.ID
		if err := p.proposalRepo.Create(ctx, prop); err != nil {
			return fmt.Errorf("failed to save proposal: %w", err)
		}
		proposalIDMap[oldID] = prop.ID
		
		p.broadcastEvent(EventProposalSubmitted, map[string]interface{}{
			"task_id":     taskID,
			"proposal_id": prop.ID,
			"type":        prop.ProposalType,
			"agent_index": i,
		})
	}

	// 8. Actualizar proposal_id en benchmarks y guardarlos
	for _, bench := range benchmarks {
		// Mapear el ID temporal al ID real de la DB
		if newID, ok := proposalIDMap[bench.ProposalID]; ok {
			bench.ProposalID = newID
		}
		if err := p.benchmarkRepo.Create(ctx, bench); err != nil {
			return fmt.Errorf("failed to save benchmark: %w", err)
		}
	}

	// 9. Ejecutar consenso
	p.broadcastEvent(EventBenchmarkCompleted, map[string]interface{}{
		"task_id":   taskID,
		"proposals": len(proposals),
	})

	// Marcar agent_executions como completados
	for _, execID := range agentExecutionIDs {
		exec, err := p.agentExecRepo.GetByID(ctx, int(execID))
		if err == nil && exec != nil {
			exec.Status = entities.ExecutionCompleted
			now := time.Now().UTC()
			exec.CompletedAt = &now
			p.agentExecRepo.Update(ctx, exec)
		}
	}

	criteria := entities.ScoringCriteria{
		PerformanceWeight: 0.5,
		StorageWeight:     0.2,
		ComplexityWeight:  0.2,
		RiskWeight:        0.1,
	}

	decision, err := p.consensus.Decide(ctx, proposals, benchmarks, criteria)
	if err != nil {
		task.Status = entities.TaskStatusFailed
		p.taskRepo.Update(ctx, task)
		p.broadcastEvent(EventTaskFailed, map[string]interface{}{
			"task_id": taskID,
			"error":   fmt.Sprintf("consensus failed: %v", err),
		})
		return fmt.Errorf("consensus failed: %w", err)
	}

	// Actualizar proposals con score_breakdown calculado por consenso
	scoreAgentTypes := []values.AgentType{values.AgentCerebro, values.AgentOperativo, values.AgentBulk}
	for i, prop := range proposals {
		agentType := scoreAgentTypes[i]
		if score, ok := decision.AllScores[agentType]; ok {
			prop.EstimatedImpact.ScoreBreakdown = map[string]float64{
				"performance":    score.Performance,
				"storage":        score.Storage,
				"complexity":     score.Complexity,
				"risk":           score.Risk,
				"weighted_total": score.WeightedTotal,
			}
			if updateErr := p.proposalRepo.Update(ctx, prop); updateErr != nil {
				fmt.Printf("Warning: failed to update proposal %d with scores: %v\n", prop.ID, updateErr)
			}
		}
	}

	// 9. Guardar decisión de consenso
	decision.TaskID = int64(task.ID)
	if err := p.consensusRepo.Create(ctx, decision); err != nil {
		return fmt.Errorf("failed to save consensus decision: %w", err)
	}

	p.broadcastEvent(EventConsensusReached, map[string]interface{}{
		"task_id":             taskID,
		"winning_proposal_id": decision.WinningProposalID,
	})

	// 10. Aplicar solución ganadora (COMENTADO por ahora para evitar cambios en DB real)
	/*
	var winner *entities.OptimizationProposal
	if decision.WinningProposalID != nil {
		for _, p := range proposals {
			if p.ID == *decision.WinningProposalID {
				winner = p
				break
			}
		}
	}

	if winner != nil {
		if err := p.orchestrator.ApplyToMainDB(ctx, p.mainService, winner); err != nil {
			task.Status = entities.TaskStatusFailed
			p.taskRepo.Update(ctx, task)
			p.broadcastEvent(EventTaskFailed, map[string]interface{}{
				"task_id": taskID,
				"error":   fmt.Sprintf("failed to apply optimization: %v", err),
			})
			return fmt.Errorf("failed to apply optimization: %w", err)
		}
	}
	*/

	// 11. Limpiar forks
	if err := p.orchestrator.CleanupForks(ctx, forkIDs); err != nil {
		// Log error pero no fallar el proceso
		fmt.Printf("Warning: failed to cleanup forks: %v\n", err)
	}

	// 12. Marcar tarea como completada
	now := time.Now().UTC()
	task.Status = entities.TaskStatusCompleted
	task.CompletedAt = &now
	if err := p.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	p.broadcastEvent(EventTaskCompleted, map[string]interface{}{
		"task_id":             taskID,
		"status":              "completed",
		"winning_proposal_id": decision.WinningProposalID,
		"completed_at":        now.Format(time.RFC3339),
	})

	return nil
}

func (p *TaskProcessor) broadcastEvent(eventType EventType, payload map[string]interface{}) {
	if p.hub != nil {
		p.hub.Broadcast(Event{
			Type:    eventType,
			Payload: payload,
		})
	}
}
