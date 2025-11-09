package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type MetricsHandler struct {
	db *sqlx.DB
}

func NewMetricsHandler(db *sqlx.DB) *MetricsHandler {
	return &MetricsHandler{db: db}
}

type DashboardMetrics struct {
	TotalTasks       int     `json:"total_tasks"`
	CompletedTasks   int     `json:"completed_tasks"`
	FailedTasks      int     `json:"failed_tasks"`
	InProgressTasks  int     `json:"in_progress_tasks"`
	SuccessRate      float64 `json:"success_rate"`
	AvgDuration      float64 `json:"avg_duration_seconds"`
	TotalOptimizations int   `json:"total_optimizations"`
	AvgImprovement   float64 `json:"avg_improvement_percent"`
}

type AgentMetrics struct {
	AgentType   string  `json:"agent_type"`
	TotalTasks  int     `json:"total_tasks"`
	SuccessRate float64 `json:"success_rate"`
	WinRate     float64 `json:"win_rate"`
	AvgDuration float64 `json:"avg_duration_seconds"`
}

func (h *MetricsHandler) GetOverview(c *fiber.Ctx) error {
	// Query metrics from database
	type TaskStats struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	var stats []TaskStats
	query := `
		SELECT status, COUNT(*) as count
		FROM tasks
		GROUP BY status
	`
	err := h.db.Select(&stats, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch metrics",
		})
	}

	metrics := DashboardMetrics{}
	for _, s := range stats {
		metrics.TotalTasks += s.Count
		switch s.Status {
		case "completed":
			metrics.CompletedTasks = s.Count
		case "failed":
			metrics.FailedTasks = s.Count
		case "in_progress":
			metrics.InProgressTasks = s.Count
		}
	}

	// Calculate success rate
	if metrics.TotalTasks > 0 {
		metrics.SuccessRate = float64(metrics.CompletedTasks) / float64(metrics.TotalTasks) * 100
	}

	// Avg duration
	var avgDuration *float64
	durationQuery := `
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) as avg_duration
		FROM tasks
		WHERE completed_at IS NOT NULL
	`
	h.db.Get(&avgDuration, durationQuery)
	if avgDuration != nil {
		metrics.AvgDuration = *avgDuration
	}

	// Total optimizations
	optQuery := `SELECT COUNT(*) FROM optimizations`
	h.db.Get(&metrics.TotalOptimizations, optQuery)

	// Avg improvement
	var avgImp *float64
	impQuery := `
		SELECT AVG(performance_improvement) as avg_imp
		FROM optimizations
		WHERE performance_improvement IS NOT NULL
	`
	h.db.Get(&avgImp, impQuery)
	if avgImp != nil {
		metrics.AvgImprovement = *avgImp
	}

	return c.JSON(metrics)
}

func (h *MetricsHandler) GetAgentMetrics(c *fiber.Ctx) error {
	query := `
		SELECT 
			ae.agent_type,
			COUNT(*) as total_tasks,
			AVG(CASE WHEN ae.status = 'completed' THEN 1.0 ELSE 0.0 END) * 100 as success_rate,
			AVG(EXTRACT(EPOCH FROM (ae.completed_at - ae.started_at))) as avg_duration
		FROM agent_executions ae
		WHERE ae.started_at IS NOT NULL
		GROUP BY ae.agent_type
		ORDER BY total_tasks DESC
	`

	var metrics []AgentMetrics
	err := h.db.Select(&metrics, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch agent metrics",
		})
	}

	// Calculate win rate (optimizations that were selected)
	for i := range metrics {
		winQuery := `
			SELECT COUNT(*) 
			FROM optimizations o
			INNER JOIN consensus_decisions cd ON cd.winning_proposal_id = o.id
			INNER JOIN agent_executions ae ON ae.id = o.agent_execution_id
			WHERE ae.agent_type = $1
		`
		var wins int
		h.db.Get(&wins, winQuery, metrics[i].AgentType)
		
		if metrics[i].TotalTasks > 0 {
			metrics[i].WinRate = float64(wins) / float64(metrics[i].TotalTasks) * 100
		}
	}

	return c.JSON(metrics)
}

type PerformanceData struct {
	Date            string  `json:"date" db:"date"`
	TasksCompleted  int     `json:"tasks_completed" db:"tasks_completed"`
	AvgImprovement  float64 `json:"avg_improvement" db:"avg_improvement"`
}

func (h *MetricsHandler) GetPerformance(c *fiber.Ctx) error {
	days := c.QueryInt("days", 7)

	query := `
		SELECT 
			DATE(t.created_at) as date,
			COUNT(*) as tasks_completed,
			COALESCE(AVG(o.performance_improvement), 0) as avg_improvement
		FROM tasks t
		LEFT JOIN optimizations o ON o.task_id = t.id
		WHERE t.created_at >= NOW() - INTERVAL '1 day' * $1
		AND t.status = 'completed'
		GROUP BY DATE(t.created_at)
		ORDER BY date ASC
	`

	var data []PerformanceData
	err := h.db.Select(&data, query, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch performance data",
		})
	}

	return c.JSON(data)
}
