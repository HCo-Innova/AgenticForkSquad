package usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// PrecisionReport validates that fork execution stays within 20% deviation from main DB
type PrecisionReport struct {
	TaskType           string
	TestQueries        int
	DeviationThreshold float64
	Results            []QueryDeviationResult
	OverallPassed      bool
	OverallDeviation   float64
	Timestamp          time.Time
}

// QueryDeviationResult represents deviation for a single query
type QueryDeviationResult struct {
	QueryName        string
	MainTimeMS       float64
	ForkTimeMS       float64
	DeviationPercent float64
	WithinThreshold  bool
	Runs             int
}

// MockMCPForPrecision wraps real MCPClient for precision testing
type MockMCPForPrecision struct {
	client *mcp.MCPClient
	config struct {
		mainServiceID  string
		fork1ServiceID string
		fork2ServiceID string
	}
}

// NewMockMCPForPrecision creates precision test wrapper
func NewMockMCPForPrecision(client *mcp.MCPClient, mainID, fork1ID, fork2ID string) *MockMCPForPrecision {
	m := &MockMCPForPrecision{client: client}
	m.config.mainServiceID = mainID
	m.config.fork1ServiceID = fork1ID
	m.config.fork2ServiceID = fork2ID
	return m
}

// MeasureQuery executes query N times and returns average execution time
func (m *MockMCPForPrecision) MeasureQuery(ctx context.Context, serviceID, sqlQuery string, runs int, timeoutMs int) (float64, error) {
	if runs <= 0 {
		runs = 3
	}

	var totalMS float64
	for i := 0; i < runs; i++ {
		qr, err := m.client.ExecuteQuery(ctx, serviceID, sqlQuery, timeoutMs)
		if err != nil {
			return 0, fmt.Errorf("query execution failed (run %d/%d): %w", i+1, runs, err)
		}
		totalMS += qr.ExecutionTimeMs
	}

	return totalMS / float64(runs), nil
}

// CalculateDeviation calculates percentage deviation between fork and main
func (m *MockMCPForPrecision) CalculateDeviation(mainTimeMS, forkTimeMS float64) float64 {
	if mainTimeMS == 0 {
		return 0
	}
	deviation := ((forkTimeMS - mainTimeMS) / mainTimeMS) * 100
	if deviation < 0 {
		deviation = -deviation
	}
	return deviation
}

// TestPrecisionDeviation validates 20% deviation threshold across multiple queries
func TestPrecisionDeviation(t *testing.T) {
	t.Skip("Integration test - requires live Tiger Cloud forks. Run with: docker compose exec backend go test -v ./internal/usecases -run TestPrecisionDeviation -timeout 10m -skip=false")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Configuration from environment - loaded by MCP client
	mainServiceID := "wuj5xa6zpz"      // TIGER_MAIN_SERVICE from .env
	fork1ServiceID := "gwb579t287"     // TIGER_FORK_A1_SERVICE_ID from .env
	deviationThreshold := 20.0          // 20% tolerance
	queryTimeoutMS := 60000             // 60s per query

	// Initialize MCP client (reads from .env directly)
	httpClient := &HTTPClientForTest{}
	
	client, err := mcp.New(nil, httpClient)
	if err != nil {
		t.Fatalf("failed to create MCP client: %v", err)
	}

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect MCP client: %v", err)
	}

	// Create precision tester
	precision := NewMockMCPForPrecision(client, mainServiceID, fork1ServiceID, fork1ServiceID)

	// Test suite: variety of queries
	testQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "simple_select",
			query: "SELECT COUNT(*) as cnt FROM orders",
		},
		{
			name:  "with_filter",
			query: "SELECT COUNT(*) as cnt FROM orders WHERE status = 'completed'",
		},
		{
			name:  "with_join",
			query: "SELECT COUNT(DISTINCT u.id) as unique_users FROM users u JOIN orders o ON u.id = o.user_id",
		},
		{
			name:  "aggregation",
			query: "SELECT status, COUNT(*) as cnt FROM orders GROUP BY status",
		},
	}

	report := &PrecisionReport{
		TaskType:           "query_execution_precision",
		TestQueries:        len(testQueries),
		DeviationThreshold: deviationThreshold,
		Results:            make([]QueryDeviationResult, 0),
		Timestamp:          time.Now(),
	}

	t.Logf("\n========== PRECISION DEVIATION TEST ==========")
	t.Logf("Task: %s", report.TaskType)
	t.Logf("Threshold: %.1f%%", deviationThreshold)
	t.Logf("Queries: %d", len(testQueries))
	t.Logf("Runs per query: 3")
	t.Logf("Timeout per query: %dms", queryTimeoutMS)
	t.Logf("============================================\n")

	totalDeviation := 0.0
	passedTests := 0

	// Execute tests
	for _, q := range testQueries {
		t.Logf("Testing: %s", q.name)

		// Measure on main DB
		mainTimeMS, err := precision.MeasureQuery(ctx, mainServiceID, q.query, 3, queryTimeoutMS)
		if err != nil {
			t.Logf("  ❌ Main DB measurement failed: %v", err)
			continue
		}
		t.Logf("  Main DB avg: %.2f ms", mainTimeMS)

		// Measure on fork
		forkTimeMS, err := precision.MeasureQuery(ctx, fork1ServiceID, q.query, 3, queryTimeoutMS)
		if err != nil {
			t.Logf("  ❌ Fork DB measurement failed: %v", err)
			continue
		}
		t.Logf("  Fork DB avg: %.2f ms", forkTimeMS)

		// Calculate deviation
		deviation := precision.CalculateDeviation(mainTimeMS, forkTimeMS)
		withinThreshold := deviation <= deviationThreshold
		t.Logf("  Deviation: %.2f%%", deviation)
		if withinThreshold {
			t.Logf("  ✅ PASSED (within %.1f%% threshold)", deviationThreshold)
			passedTests++
		} else {
			t.Logf("  ❌ FAILED (exceeds %.1f%% threshold)", deviationThreshold)
		}
		t.Logf("")

		// Store result
		result := QueryDeviationResult{
			QueryName:        q.name,
			MainTimeMS:       mainTimeMS,
			ForkTimeMS:       forkTimeMS,
			DeviationPercent: deviation,
			WithinThreshold:  withinThreshold,
			Runs:             3,
		}
		report.Results = append(report.Results, result)
		totalDeviation += deviation
	}

	// Calculate overall statistics
	report.OverallPassed = passedTests == len(testQueries)
	if len(testQueries) > 0 {
		report.OverallDeviation = totalDeviation / float64(len(testQueries))
	}

	// Print summary
	t.Logf("========== PRECISION TEST SUMMARY ==========")
	t.Logf("Passed: %d/%d", passedTests, len(testQueries))
	t.Logf("Avg Deviation: %.2f%%", report.OverallDeviation)
	t.Logf("Overall Status: %v", report.OverallPassed)
	t.Logf("==========================================\n")

	// Final validation
	if !report.OverallPassed {
		t.Errorf("Precision test failed: %d queries exceeded deviation threshold", len(testQueries)-passedTests)
	}

	if report.OverallDeviation > deviationThreshold {
		t.Logf("Warning: Average deviation (%.2f%%) close to threshold (%.1f%%)",
			report.OverallDeviation, deviationThreshold)
	}
}

// Helper functions
func NewHTTPClient() interface{} {
	// Return a basic HTTP client for MCP initialization
	// Actual implementation depends on MCP client expectations
	return nil
}

func LoadTestConfig() interface{} {
	// In production, this would load from environment
	// For testing, we rely on .env being loaded by docker compose
	return nil
}

// HTTPClientForTest is a minimal HTTP client wrapper for testing
type HTTPClientForTest struct{}

func (c *HTTPClientForTest) Do(req interface{}) (interface{}, error) {
	return nil, nil
}
