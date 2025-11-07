package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/database"
)

// RouterContext contains context-enriched information for agent assignment.
type RouterContext struct {
	Task                *entities.Task
	SimilarPastQueries  []*SimilarQuery
	RecommendedAgents   []string
	SearchStats         *SearchStats
	RoutingTimeMs       float64
}

// QueryRouter uses hybrid search to enrich task context before agent assignment.
type QueryRouter struct {
	hybridSearch *HybridSearchService
	queryLogger  *database.QueryLogger
}

// NewQueryRouter creates a new query router with hybrid search context.
func NewQueryRouter(hybridSearch *HybridSearchService, queryLogger *database.QueryLogger) *QueryRouter {
	return &QueryRouter{
		hybridSearch: hybridSearch,
		queryLogger:  queryLogger,
	}
}

// RouteTask enriches task information with similar queries from historical logs
// and provides recommendations for agent selection.
//
// This serves as an input layer for Orchestrator/Router:
// 1. Find similar queries from past optimizations
// 2. Extract patterns and successful strategies
// 3. Use patterns to guide agent recommendations
// 4. Pass enriched context to agent assignment logic
func (qr *QueryRouter) RouteTask(ctx context.Context, task *entities.Task) (*RouterContext, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if task.TargetQuery == "" {
		return nil, fmt.Errorf("task.TargetQuery cannot be empty")
	}

	startTime := time.Now()

	// Perform hybrid search for similar queries
	searchResult, err := qr.hybridSearch.HybridSearch(ctx, task.TargetQuery, 10)
	if err != nil {
		// If search fails, continue with enriched context but note the failure
		fmt.Printf("warning: hybrid search failed: %v\n", err)
	}

	// Get search statistics
	stats, err := qr.hybridSearch.GetSearchStats(ctx)
	if err != nil {
		fmt.Printf("warning: failed to get search stats: %v\n", err)
	}

	// Analyze similar queries to recommend agents
	recommendedAgents := qr.recommendAgents(searchResult, task)

	routingTime := time.Since(startTime).Milliseconds()

	return &RouterContext{
		Task:               task,
		SimilarPastQueries: searchResult.SimilarQueries,
		RecommendedAgents: recommendedAgents,
		SearchStats:        stats,
		RoutingTimeMs:      float64(routingTime),
	}, nil
}

// recommendAgents analyzes similar queries and recommends agent types.
// Strategy:
// - If similar queries had fast execution: recommend speed-optimized agents
// - If similar queries involved complex joins: recommend deep analysis agents
// - If similar queries had high storage overhead: recommend storage-optimized agents
// - Default: all agents (let orchestrator decide via consensus)
func (qr *QueryRouter) recommendAgents(searchResult *HybridSearchResult, task *entities.Task) []string {
	if searchResult == nil || len(searchResult.SimilarQueries) == 0 {
		// No historical data - use all agents
		return []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"}
	}

	// Analyze similar queries for patterns
	var totalExecTime float64
	var maxExecTime float64
	fastCount := 0
	slowCount := 0

	for _, sq := range searchResult.SimilarQueries {
		if sq.ExecutionTimeMs != nil && *sq.ExecutionTimeMs > 0 {
			totalExecTime += *sq.ExecutionTimeMs
			if *sq.ExecutionTimeMs > maxExecTime {
				maxExecTime = *sq.ExecutionTimeMs
			}

			if *sq.ExecutionTimeMs < 100 {
				fastCount++
			} else if *sq.ExecutionTimeMs > 1000 {
				slowCount++
			}
		}
	}

	// Recommendation logic based on patterns
	recommendations := make(map[string]int)

	// Always include at least one capable agent
	recommendations["gemini-2.5-pro"] = 10 // Main/planning agent

	// If similar queries are fast, trust flash models (cheaper)
	if fastCount > 0 && fastCount >= len(searchResult.SimilarQueries)/2 {
		recommendations["gemini-2.5-flash"] = 8
	}

	// If similar queries are slow or complex, add all agents
	if slowCount > 0 || totalExecTime > 5000 {
		recommendations["gemini-2.5-flash"] = 7
		recommendations["gemini-2.0-flash"] = 6
	} else if len(searchResult.SimilarQueries) < 3 {
		// Few similar queries - use all agents for robustness
		recommendations["gemini-2.5-flash"] = 7
		recommendations["gemini-2.0-flash"] = 6
	}

	// Convert to sorted list
	result := make([]string, 0, len(recommendations))
	if score, ok := recommendations["gemini-2.5-pro"]; ok && score > 0 {
		result = append(result, "gemini-2.5-pro")
	}
	if score, ok := recommendations["gemini-2.5-flash"]; ok && score > 0 {
		result = append(result, "gemini-2.5-flash")
	}
	if score, ok := recommendations["gemini-2.0-flash"]; ok && score > 0 {
		result = append(result, "gemini-2.0-flash")
	}

	if len(result) == 0 {
		// Fallback: all agents
		result = []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"}
	}

	return result
}

// LogQueryExecution records a query execution for future hybrid search.
// Called after task completion to enrich the knowledge base.
func (qr *QueryRouter) LogQueryExecution(ctx context.Context, task *entities.Task, executionTimeMs float64, rowsReturned int) error {
	if task == nil || task.TargetQuery == "" {
		return fmt.Errorf("invalid task")
	}

	if qr.queryLogger == nil {
		return fmt.Errorf("query logger not configured")
	}

	entry := &database.QueryLogEntry{
		QueryText:       task.TargetQuery,
		ExecutionTimeMs: &executionTimeMs,
		RowsReturned:    &rowsReturned,
		ExecutedAt:      time.Now(),
		TaskID:          nil, // Could be populated from task.ID if available
		Notes:           &task.Description,
	}

	_, err := qr.queryLogger.LogQuery(ctx, entry, false) // false = don't generate embedding (would need LLM)
	if err != nil {
		return fmt.Errorf("failed to log query: %w", err)
	}

	return nil
}

// GetRouterStats retrieves statistics about the routing system.
func (qr *QueryRouter) GetRouterStats(ctx context.Context) (map[string]interface{}, error) {
	if qr.hybridSearch == nil {
		return nil, fmt.Errorf("hybrid search service not configured")
	}

	stats, err := qr.hybridSearch.GetSearchStats(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_queries":              stats.TotalQueries,
		"queries_with_embedding":     stats.QueriesWithEmbedding,
		"slow_query_count":           stats.SlowQueryCount,
		"average_execution_ms":       stats.AverageExecutionMs,
		"last_log_time":              stats.LastLogTime,
		"embedding_generation_rate":  float64(stats.QueriesWithEmbedding) / float64(stats.TotalQueries),
	}, nil
}
