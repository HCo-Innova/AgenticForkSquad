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
	TotalTasks         int     `json:"total_tasks"`
	CompletedTasks     int     `json:"completed_tasks"`
	FailedTasks        int     `json:"failed_tasks"`
	InProgressTasks    int     `json:"in_progress_tasks"`
	SuccessRate        float64 `json:"success_rate"`
	AvgDuration        float64 `json:"avg_duration_seconds"`
	TotalOptimizations int     `json:"total_optimizations"`
	AvgImprovement     float64 `json:"avg_improvement_percent"`
}

type AgentMetrics struct {
	AgentType   string  `json:"agent_type" db:"agent_type"`
	Name        string  `json:"name"`
	TotalTasks  int     `json:"total_tasks" db:"total_tasks"`
	Wins        int     `json:"wins"`
	SuccessRate float64 `json:"success_rate" db:"success_rate"`
	WinRate     float64 `json:"win_rate"`
	AvgDuration float64 `json:"avg_duration" db:"avg_duration"`
}

type PerformanceData struct {
	Date            string  `json:"date" db:"date"`
	TasksCompleted  int     `json:"tasks_completed" db:"tasks_completed"`
	AvgImprovement  float64 `json:"avg_improvement" db:"avg_improvement"`
}

type AgentMetricData struct {
	AgentID          string  `json:"agent_id" db:"agent_id"`
	TotalTasks       int     `json:"total_tasks" db:"total_tasks"`
	AvgExecutionTime float64 `json:"avg_execution_time" db:"avg_execution_time"`
	SuccessRate      float64 `json:"success_rate" db:"success_rate"`
}

type OverviewData struct {
	TotalTasks           int     `json:"total_tasks"`
	OptimizationsApplied int     `json:"optimizations_applied"`
	AvgPerformance       float64 `json:"avg_performance"`
	ActiveAgents         int     `json:"active_agents"`
}

func (h *MetricsHandler) GetOverview(c *fiber.Ctx) error {
	type TaskStats struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	var stats []TaskStats
	query := "SELECT status, COUNT(*) as count FROM tasks GROUP BY status"
	
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

	if metrics.TotalTasks > 0 {
		metrics.SuccessRate = float64(metrics.CompletedTasks) / float64(metrics.TotalTasks) * 100
	}

	var avgDuration *float64
	durationQuery := "SELECT AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) as avg_duration FROM tasks WHERE completed_at IS NOT NULL"
	h.db.Get(&avgDuration, durationQuery)
	if avgDuration != nil {
		metrics.AvgDuration = *avgDuration
	}

	optQuery := "SELECT COUNT(*) FROM optimization_proposals"
	var totalOpts int
	if err := h.db.Get(&totalOpts, optQuery); err == nil {
		metrics.TotalOptimizations = totalOpts
	}

	metrics.AvgImprovement = 0

	return c.JSON(metrics)
}

func (h *MetricsHandler) GetAgentMetrics(c *fiber.Ctx) error {
	query := `
		SELECT 
			ae.agent_type,
			COUNT(*) as total_tasks,
			COALESCE(AVG(CASE WHEN ae.status = 'completed' THEN 1.0 ELSE 0.0 END) * 100, 0) as success_rate,
			COALESCE(AVG(EXTRACT(EPOCH FROM (ae.completed_at - ae.started_at))), 0) as avg_duration
		FROM agent_executions ae
		WHERE ae.started_at IS NOT NULL
		GROUP BY ae.agent_type
		ORDER BY total_tasks DESC
	`

	var metrics []AgentMetrics
	err := h.db.Select(&metrics, query)
	if err != nil {
		return c.JSON([]AgentMetrics{})
	}

	for i := range metrics {
		metrics[i].Name = metrics[i].AgentType
		winQuery := `
			SELECT COUNT(*)
			FROM consensus_decisions cd
			INNER JOIN optimization_proposals o ON cd.winning_proposal_id = o.id
			INNER JOIN agent_executions ae ON ae.id = o.agent_execution_id
			WHERE ae.agent_type = $1
		`
		var wins int
		if err := h.db.Get(&wins, winQuery, metrics[i].AgentType); err == nil {
			metrics[i].Wins = wins
			if metrics[i].TotalTasks > 0 {
				metrics[i].WinRate = float64(wins) / float64(metrics[i].TotalTasks) * 100
			}
		}
	}

	if metrics == nil {
		metrics = []AgentMetrics{}
	}

	return c.JSON(metrics)
}

func (h *MetricsHandler) GetPerformance(c *fiber.Ctx) error {
	days := c.QueryInt("days", 7)

	type perfRow struct {
		Date           string  `db:"date"`
		TasksCompleted int     `db:"tasks_completed"`
		SuccessRate    float64 `db:"success_rate"`
		AvgDuration    float64 `db:"avg_duration"`
	}

	query := `
		SELECT 
			DATE(t.created_at)::TEXT as date,
			COUNT(*)::INT as tasks_completed,
			COALESCE(AVG(CASE WHEN t.status = 'completed' THEN 100.0 ELSE 0 END), 0) as success_rate,
			COALESCE(AVG(EXTRACT(EPOCH FROM (t.completed_at - t.created_at))), 0) as avg_duration
		FROM tasks t
		WHERE t.created_at >= CURRENT_DATE - INTERVAL '1 day' * $1
		GROUP BY DATE(t.created_at)
		ORDER BY date ASC
	`

	var data []perfRow
	err := h.db.Select(&data, query, days)
	
	if err != nil {
		return c.JSON([]fiber.Map{})
	}

	response := make([]fiber.Map, 0, len(data))
	for _, row := range data {
		response = append(response, fiber.Map{
			"date":         row.Date,
			"tasks":        row.TasksCompleted,
			"success_rate": row.SuccessRate,
			"avg_duration": row.AvgDuration,
		})
	}

	if response == nil {
		response = []fiber.Map{}
	}

	return c.JSON(response)
}
