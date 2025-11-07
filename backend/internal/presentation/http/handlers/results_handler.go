package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/usecases"
)

// ResultsHandler agrupa endpoints de lectura: agentes, propuestas, benchmarks, consenso.
type ResultsHandler struct {
	ExecRepo  domainif.AgentExecutionRepository
	OptRepo   domainif.OptimizationRepository
	BenchRepo domainif.BenchmarkRepository
	ConsRepo  domainif.ConsensusRepository
	Hub       *usecases.Hub
}

func NewResultsHandler(exec domainif.AgentExecutionRepository, opt domainif.OptimizationRepository, bench domainif.BenchmarkRepository, cons domainif.ConsensusRepository, hub *usecases.Hub) *ResultsHandler {
	return &ResultsHandler{ExecRepo: exec, OptRepo: opt, BenchRepo: bench, ConsRepo: cons, Hub: hub}
}

// GET /api/v1/tasks/:id/agents
func (h *ResultsHandler) GetTaskAgents(c *fiber.Ctx) error {
	if h == nil || h.ExecRepo == nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "repositories not available", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"code": "VALIDATION_ERROR", "message": "Invalid id parameter", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	execs, err := h.ExecRepo.GetByTaskID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error(), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	resp := make([]fiber.Map, 0, len(execs))
	for _, e := range execs {
		m := fiber.Map{
			"id":           e.ID,
			"task_id":      e.TaskID,
			"agent_type":   e.AgentType,
			"fork_id":      e.ForkID,
			"status":       e.Status,
			"started_at":   e.StartedAt.Format(time.RFC3339),
			"completed_at": nil,
			"error":        e.ErrorMsg,
			"links": fiber.Map{
				"proposals": "/api/v1/tasks/" + strconv.Itoa(id) + "/proposals",
			},
		}
		if e.CompletedAt != nil {
			m["completed_at"] = e.CompletedAt.Format(time.RFC3339)
		}
		resp = append(resp, m)
	}
	return c.JSON(fiber.Map{"data": resp})
}

// GET /api/v1/tasks/:id/proposals
func (h *ResultsHandler) GetTaskProposals(c *fiber.Ctx) error {
	if h == nil || h.ExecRepo == nil || h.OptRepo == nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "repositories not available", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"code": "VALIDATION_ERROR", "message": "Invalid id parameter", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	execs, err := h.ExecRepo.GetByTaskID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error(), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	out := make([]fiber.Map, 0)
	for _, e := range execs {
		props, err := h.OptRepo.GetByAgentExecutionID(c.Context(), int(e.ID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error(), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
		}
		for _, p := range props {
			out = append(out, fiber.Map{
				"id":                  p.ID,
				"agent_execution_id":  p.AgentExecutionID,
				"proposal_type":       p.ProposalType,
				"sql_commands":        p.SQLCommands,
				"rationale":           p.Rationale,
				"estimated_impact":    p.EstimatedImpact,
				"created_at":          p.CreatedAt.Format(time.RFC3339),
				"links": fiber.Map{
					"benchmarks": "/api/v1/proposals/" + strconv.FormatInt(p.ID, 10) + "/benchmarks",
				},
			})
		}
	}
	return c.JSON(fiber.Map{"data": out})
}

// GET /api/v1/proposals/:id/benchmarks
func (h *ResultsHandler) GetProposalBenchmarks(c *fiber.Ctx) error {
	if h == nil || h.BenchRepo == nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "repositories not available", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"code": "VALIDATION_ERROR", "message": "Invalid id parameter", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	bms, err := h.BenchRepo.GetByProposalID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error(), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	resp := make([]fiber.Map, 0, len(bms))
	for _, b := range bms {
		m := fiber.Map{
			"id":                b.ID,
			"proposal_id":       b.ProposalID,
			"query_name":        b.QueryName,
			"query_executed":    b.QueryExecuted,
			"execution_time_ms": b.ExecutionTimeMS,
			"rows_returned":     b.RowsReturned,
			"explain_plan":      b.ExplainPlan,
			"storage_impact_mb": b.StorageImpactMB,
			"created_at":        b.CreatedAt.Format(time.RFC3339),
		}
		resp = append(resp, m)
	}
	return c.JSON(fiber.Map{"data": resp})
}

// GET /api/v1/tasks/:id/consensus
func (h *ResultsHandler) GetTaskConsensus(c *fiber.Ctx) error {
	if h == nil || h.ConsRepo == nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "repositories not available", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"code": "VALIDATION_ERROR", "message": "Invalid id parameter", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	d, err := h.ConsRepo.GetByTaskID(c.Context(), id)
	if err != nil {
		// Si no existe, devolver 404 conforme a la spec
		return c.Status(404).JSON(fiber.Map{"error": fiber.Map{"code": "CONSENSUS_NOT_FOUND", "message": "Consensus not found", "timestamp": time.Now().UTC().Format(time.RFC3339)}})
	}
	var winning *int64
	if d.WinningProposalID != nil { winning = d.WinningProposalID }
	return c.JSON(fiber.Map{
		"id":                  d.ID,
		"task_id":             d.TaskID,
		"winning_proposal_id": winning,
		"all_scores":          d.AllScores,
		"decision_rationale":  d.DecisionRationale,
		"applied_to_main":     d.AppliedToMain,
		"created_at":          d.CreatedAt.Format(time.RFC3339),
	})
}
