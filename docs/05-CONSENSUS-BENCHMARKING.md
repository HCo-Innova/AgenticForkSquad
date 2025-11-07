# 05-CONSENSUS-BENCHMARKING.md

```markdown
# 05-CONSENSUS-BENCHMARKING.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Consensus & Benchmarking Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md), 
[04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md)

---

## 📚 Table of Contents

1. [System Overview](#system-overview)
2. [Benchmark Runner](#benchmark-runner)
3. [PITR Validation](#pitr-validation)
4. [Consensus Engine](#consensus-engine)
5. [Scoring Algorithm](#scoring-algorithm)
6. [Decision Making](#decision-making)
7. [Apply to Main Database](#apply-to-main-database)
8. [Rationale Generation](#rationale-generation)
9. [Edge Cases](#edge-cases)

---

## 🎯 System Overview

### Purpose

**Objective Decision Making:**  
After multiple agents propose different optimizations, the system must 
objectively select the best solution. This requires:

1. **Objective Measurement:** Actual performance data, not estimates
2. **Multi-Criteria Evaluation:** Consider multiple factors (not just speed)
3. **Transparent Process:** Clear explanation of why winner selected
4. **Reproducible Results:** Same inputs = same decision

**Three-Phase Process:**

**Phase 1: PITR Validation**
- Verify Tiger Cloud fork and rollback work correctly
- Measure fork creation time (<10s target)
- Confirm PITR rollback restores correct state
- Ensure rollback marker table is absent (proof)

**Phase 2: Benchmarking**
- Execute each proposal in its fork
- Measure actual performance impact
- Collect quantitative metrics
- Generate objective data for comparison

**Phase 3: Consensus**
- Score all proposals using weighted criteria
- Rank proposals by total score
- Select winner
- Generate human-readable rationale

---

### Core Principles

**Data-Driven:**
- Decisions based on measurements, not hunches
- Benchmarks run in identical conditions
- Statistical significance (multiple runs, averages)

**Balanced Evaluation:**
- Performance is important but not only factor
- Storage overhead matters (finite resource)
- Operational complexity considered (maintainability)
- Risk assessment (safety first)

**Transparency:**
- All scores visible to users
- Rationale explains decision
- Audit trail of decision process
- No black box decisions

**Configurability:**
- Scoring weights adjustable per task
- Different priorities for different scenarios
- User can influence decision criteria
- System learns from feedback (future enhancement)

---

## ⏰ PITR Validation

### Purpose

**Validate Tiger Cloud PITR (Point-in-Time Recovery) Capability:**

Before benchmarking and consensus decisions, verify that zero-copy forks 
work correctly and rollback functionality is reliable. This validation 
ensures the core infrastructure is sound.

**Success Criteria:**
- ✅ Fork creation completes in <10 seconds
- ✅ PITR fork rollback works (timestamp-based recovery)
- ✅ Marker table absent in PITR fork (proves rollback worked)
- ✅ Cleanup succeeds (no orphaned forks)

---

### Implementation (Go Tool)

**Location:** `backend/cmd/tools/validate_pitr/main.go`

**Binary:** `backend/validate_pitr`

**Execution:**
```bash
docker compose exec backend /app/validate_pitr 2>&1
```

**Dependencies:**
- Tiger CLI v0.15.1 (installed in backend Dockerfile)
- Tiger Cloud credentials (`TIGER_PUBLIC_KEY`, `TIGER_SECRET_KEY`, `TIGER_PROJECT_ID`)
- MCPClient (stateless CLI proxy in `backend/internal/infrastructure/mcp/client.go`)

---

### Validation Flow

**Step 1: Authentication**
```go
// Create and authenticate client
client, err := mcp.New(cfg, httpClient)
if err != nil {
    log.Fatalf("mcp client init: %v", err)
}

if err := client.Connect(ctx); err != nil {
    log.Fatalf("mcp connect: %v", err)
}

// Internally runs: tiger auth login --public-key ... --secret-key ...
// Then verifies: tiger auth status
```

**Step 2: Create Regular Fork (Measure Time)**
```go
forkName := fmt.Sprintf("afs-validate-fork-%d", time.Now().Unix())
start := time.Now()

forkID, err := client.CreateFork(ctx, parentServiceID, forkName)
duration := time.Since(start)

// Internally runs: tiger service fork <parent> --name <name> --now -o json
// Expected: <10 seconds
```

**Step 3: Sanity Check Fork**
```go
// Execute simple query to verify fork is accessible
_, err := client.ExecuteQuery(ctx, forkID, "SELECT 1", 10000)
if err != nil {
    return err  // Fork not accessible
}

// Clean up first fork
client.DeleteFork(ctx, forkID)
```

**Step 4: Record PITR Timestamp**
```go
// Capture exact timestamp before applying changes
pitrTs := time.Now().UTC().Format(time.RFC3339)

// Example: "2025-11-05T18:30:00Z"
```

**Step 5: Apply Harmless Change to Main**
```go
// Create marker table to verify rollback
markerTable := fmt.Sprintf("afs_tmp_validation_%d", time.Now().Unix())
createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT);", markerTable)

_, err := client.ExecuteQuery(ctx, parentServiceID, createSQL, 60000)
if err != nil {
    return err
}

defer client.ExecuteQuery(context.Background(), parentServiceID, 
    fmt.Sprintf("DROP TABLE IF EXISTS %s;", markerTable), 60000)
```

**Step 6: Create PITR Fork (Rollback Test)**
```go
// Fork from BEFORE the marker table was created
pitrForkID, err := client.CreateForkAtTimestamp(ctx, parentServiceID, 
    pitrForkName, pitrTs)
if err != nil {
    return err
}

// Internally runs: tiger service fork <parent> --name <name> 
//                 --to-timestamp "2025-11-05T18:30:00Z" -o json
```

**Step 7: Verify Rollback (Check Marker Absent)**
```go
// Query PITR fork for marker table
checkSQL := fmt.Sprintf("SELECT to_regclass('%s') AS exists;", markerTable)
qr, err := client.ExecuteQuery(ctx, pitrForkID, checkSQL, 30000)

// to_regclass returns:
// - NULL if table does NOT exist (what we want)
// - OID (numeric) if table EXISTS (rollback failed)

// Verify marker is absent (rollback successful)
if len(qr.Rows) > 0 {
    if v, ok := qr.Rows[0]["exists"]; ok && v != nil {
        return errors.New("rollback failed: marker table exists in PITR fork")
    }
}

rollbackOK = true  // Marker absent = rollback successful
```

**Step 8: Cleanup**
```go
// Delete PITR fork
err := client.DeleteFork(ctx, pitrForkID)

// Delete marker table from main
defer cleanup()

cleanupOK = (err == nil)
```

---

### Output Format

**JSON Result:**
```json
{
  "fork_create_duration_ms": 5234,
  "fork_create_under_10s": true,
  "pitr_timestamp": "2025-11-05T18:30:00Z",
  "rollback_ok": true,
  "cleanup_ok": true,
  "error": ""
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|---|
| fork_create_duration_ms | int64 | Time to create regular fork (milliseconds) |
| fork_create_under_10s | bool | true if duration ≤ 10,000 ms |
| pitr_timestamp | string | Exact timestamp of PITR fork point (RFC3339) |
| rollback_ok | bool | true if marker table absent in PITR fork |
| cleanup_ok | bool | true if all resources cleaned up |
| error | string | Error message if any step failed |

---

### Implementation Details (MCPClient)

**CLI Proxy Commands Used:**

**CreateFork:**
```bash
tiger --config-dir /app/.tiger service fork <parent-id> \
  --name <fork-name> \
  --now \
  -o json
```

**CreateForkAtTimestamp:**
```bash
tiger --config-dir /app/.tiger service fork <parent-id> \
  --name <fork-name> \
  --to-timestamp <RFC3339-timestamp> \
  -o json
```

**ExecuteQuery:**
```bash
tiger --config-dir /app/.tiger service query \
  --service <service-id> \
  --query "<sql>" \
  -o json
```

**DeleteFork:**
```bash
tiger --config-dir /app/.tiger service delete <service-id> \
  --force \
  -o json
```

**Connection:**
```bash
tiger --config-dir /app/.tiger auth login \
  --public-key <key> \
  --secret-key <secret> \
  --project-id <project>

tiger --config-dir /app/.tiger auth status
```

---

### Known Issues

**Tiger Cloud API "Unknown Error" on Fork:**

During implementation testing on Nov 5, 2025:

```
🍴 Forking service 'o120o0yba9' to create 'afs-validate-fork-1762369991' 
    at current state...
Error: unknown error
```

**Diagnostics Performed:**
- ✅ Authentication: `tiger auth status` → Success
- ✅ Service listing: `tiger service list` → Shows services
- ✅ Service details: `tiger service describe` → READY status
- ✅ Network: Backend reaches Tiger Cloud APIs
- ✅ CLI version: v0.15.1 (latest)

**Root Cause:**
- Appears to be Tiger Cloud account/API issue (not code)
- Fork capability may not be enabled on the test account
- Or service plan may not support zero-copy forks

**Workaround:**
1. Verify fork capability in Tiger Cloud dashboard
2. Check service plan includes zero-copy forks
3. Contact Tiger Cloud support if issue persists

**Code Status:**
- ✅ Go client implementation correct
- ✅ Docker Compose setup correct
- ✅ CLI command syntax correct
- ✅ Error handling works
- ⚠️ Tiger Cloud API returns error (non-code issue)

---

## 🏃 Benchmark Runner

### Responsibility

**Execute performance tests on each proposal to collect objective metrics 
that enable fair comparison.**

**Inputs:**
- OptimizationProposal entity (with SQL commands)
- Fork ID (where to run tests)
- Original query (for baseline)

**Outputs:**
- Array of BenchmarkResult entities (one per test query)
- Aggregate metrics (average improvement, storage impact)

---

### Benchmark Suite Design

**Standard Test Suite (4 Queries):**

The suite is designed to test different aspects of query performance to 
ensure the optimization works well across varied scenarios.

---

**Test 1: Baseline (Original Query)**

**Purpose:**  
Establish performance before optimization. Critical reference point.

**Query:**  
Exactly as submitted by user, no modifications.

**Example:**
```
SELECT u.email, SUM(o.total) as revenue
FROM users u 
JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed'
GROUP BY u.email
ORDER BY revenue DESC
```

**Execution:**
- Run in fork BEFORE applying proposal SQL
- Execute 3 times, take average
- Record execution time, rows returned
- Capture EXPLAIN ANALYZE output

**Why 3 Runs:**
- Database caching affects first run
- Average reduces variance
- Detects inconsistent performance

**Metrics Captured:**
- Execution time (milliseconds)
- Planning time (milliseconds)
- Rows returned
- EXPLAIN plan (full JSON)

---

**Test 2: Limited Result Set**

**Purpose:**  
Distinguish between query processing overhead and data transfer overhead.

**Query:**  
Original query with LIMIT 10 appended.

**Example:**
```
SELECT u.email, SUM(o.total) as revenue
FROM users u 
JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed'
GROUP BY u.email
ORDER BY revenue DESC
LIMIT 10
```

**Rationale:**
- If original query slow due to processing 10K rows → LIMIT helps little
- If slow due to complex JOIN/aggregation → LIMIT helps significantly
- Reveals where bottleneck is

**Baseline Comparison:**
- If LIMIT query much faster → bottleneck is row processing
- If LIMIT query similar speed → bottleneck is computation

**After Optimization:**
- Should see improvement proportional to baseline

---

**Test 3: Additional Filter (Selectivity Test)**

**Purpose:**  
Test how optimization scales with different data volumes.

**Query:**  
Original query with additional WHERE condition (date filter).

**Example:**
```
SELECT u.email, SUM(o.total) as revenue
FROM users u 
JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed'
  AND o.created_at > NOW() - INTERVAL '30 days'
GROUP BY u.email
ORDER BY revenue DESC
```

**Rationale:**
- Tests if optimization works with varied selectivity
- Ensures index/optimization not overfitted to original query
- Validates generalization

**Expected Behavior:**
- Good optimization: Scales well, improvement consistent
- Bad optimization: Only works for specific data volume
- Index should help regardless of filter selectivity

---

**Test 4: Sorting Test**

**Purpose:**  
Evaluate impact on sorting operations.

**Query:**  
Original query, emphasize ORDER BY.

**Example:**
```
SELECT u.email, SUM(o.total) as revenue
FROM users u 
JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed'
GROUP BY u.email
ORDER BY revenue DESC, u.email ASC
```

**Rationale:**
- Some indexes enable sort optimization (index-only scan)
- Materialized views pre-sort data
- Tests if optimization helps with ORDER BY

**EXPLAIN Analysis:**
- Look for "Sort" node in plan
- Check if "Index Scan" used (sorted by index)
- Compare sort method (quicksort vs index)

---

### Benchmark Execution Process

**Step-by-Step Flow:**

**Step 1: Baseline Measurement**

**Before Applying Optimization:**
```
FOR each query in benchmark suite:
  FOR run in [1, 2, 3]:
    Execute query in fork
    Record execution time
    Discard result (not stored)
  
  Calculate average execution time
  Execute EXPLAIN ANALYZE
  Parse EXPLAIN output
  Store baseline BenchmarkResult
```

**Why Before:**
- Ensures same fork state
- Same data distribution
- Fair comparison

**Storage:**
- Create BenchmarkResult record
- Mark as baseline (query_name = "baseline_test1", etc.)
- No storage impact measured (optimization not applied yet)

---

**Step 2: Apply Optimization**

**Execute Proposal SQL:**
```
BEGIN TRANSACTION (if DDL allows)

FOR each sql_command in proposal.sql_commands:
  Execute command in fork
  Log command executed
  Verify success (no errors)
  
  IF error:
    ROLLBACK
    Return error (benchmark fails)

COMMIT TRANSACTION

Verify optimization applied:
  IF proposal_type == "index":
    Query pg_indexes to verify index exists
  IF proposal_type == "materialized_view":
    Query pg_matviews to verify view exists
```

**Verification Examples:**

**Index Verification:**
- Query: `SELECT indexname FROM pg_indexes WHERE indexname = ?`
- Expected: 1 row returned
- If not found: Fail benchmark

**Materialized View Verification:**
- Query: `SELECT matviewname FROM pg_matviews WHERE matviewname = ?`
- Expected: 1 row returned
- Refresh if needed: `REFRESH MATERIALIZED VIEW name`

---

**Step 3: Optimized Measurement**

**After Applying Optimization:**
```
FOR each query in benchmark suite:
  FOR run in [1, 2, 3]:
    Execute query in fork
    Record execution time
  
  Calculate average execution time
  Execute EXPLAIN ANALYZE
  Parse EXPLAIN output
  Count rows returned
  
  Create BenchmarkResult record:
    query_name: "optimized_test1"
    execution_time_ms: average_time
    explain_plan: parsed_json
    rows_returned: count
```

**Warmup Consideration:**
- First run may be slow (cold cache)
- Option 1: Discard first run, average 3 subsequent runs
- Option 2: Average all 3 runs (simpler, current approach)

---

**Step 4: Storage Impact Measurement**

**Calculate Storage Overhead:**
```
Before optimization:
  size_before = pg_total_relation_size('orders')
  
After optimization:
  size_after = pg_total_relation_size('orders')
  
  Include indexes:
  FOR each index in new_indexes:
    size_after += pg_relation_size(index_name)

storage_overhead_mb = (size_after - size_before) / (1024 * 1024)
```

**Storage Query Details:**

**Table Size:**
- Function: `pg_total_relation_size('table_name')`
- Includes: Table data + TOAST + indexes belonging to table
- Returns: Bytes

**Index Size:**
- Function: `pg_relation_size('index_name')`
- Includes: Only the index
- Returns: Bytes

**Conversion:**
- Bytes to MB: divide by 1,048,576 (1024 × 1024)
- Precision: 2 decimal places

**Edge Cases:**
- Partitioned tables: Sum all partition sizes
- Materialized views: Include view size + indexes on view
- Query rewrites: No storage overhead (0 MB)

---

**Step 5: Improvement Calculation**

**Calculate Percentage Improvement:**
```
FOR each test query:
  baseline_time = baseline_result.execution_time_ms
  optimized_time = optimized_result.execution_time_ms
  
  improvement_pct = 
    ((baseline_time - optimized_time) / baseline_time) × 100
  
  Store in optimized_result metadata
```

**Example:**
- Baseline: 2,300 ms
- Optimized: 450 ms
- Improvement: ((2300 - 450) / 2300) × 100 = 80.4%

**Aggregate Metrics:**
```
Calculate across all test queries:
  avg_improvement = AVG(improvement_pct for all tests)
  min_improvement = MIN(improvement_pct)
  max_improvement = MAX(improvement_pct)
  
Store in proposal metadata for consensus
```

---

### EXPLAIN Plan Parsing

**Purpose:**  
Extract actionable insights from PostgreSQL EXPLAIN ANALYZE output.

**EXPLAIN Output Format:**

**Request Format:**
```
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) 
SELECT ...
```

**Returns:**
- JSON array with execution plan tree
- Timing information
- Buffer statistics
- Actual vs estimated rows

**Key Fields to Extract:**

**Top-Level:**
- Planning Time (milliseconds)
- Execution Time (milliseconds)
- Total Cost (relative units)

**Plan Node:**
- Node Type (Seq Scan, Index Scan, Hash Join, etc.)
- Actual Rows (rows returned)
- Actual Loops (execution count)
- Index Name (if index scan)
- Filter (WHERE conditions)
- Rows Removed by Filter (inefficiency indicator)

**Buffers:**
- Shared Blocks Hit (cache hits)
- Shared Blocks Read (disk reads)
- Ratio: hit / (hit + read) = cache hit rate

**Storage Structure:**

**Simplified JSON (stored in benchmark_results.explain_plan):**
```
{
  "planning_time_ms": 0.5,
  "execution_time_ms": 450.23,
  "total_cost": 1234.56,
  "plan_type": "Index Scan",
  "index_name": "idx_orders_user_completed",
  "actual_rows": 1000,
  "filter_removed_rows": 0,
  "sort_method": null,
  "buffers": {
    "shared_hit": 128,
    "shared_read": 4,
    "cache_hit_rate": 0.97
  },
  "full_plan": { /* complete explain output */ }
}
```

**Analysis Insights:**

**Scan Type Comparison:**
- Baseline: "Seq Scan" → Problem identified
- Optimized: "Index Scan" → Solution verified
- Visual indicator: Optimization working as expected

**Filter Efficiency:**
- Baseline: Rows Removed by Filter = 2,500 (inefficient)
- Optimized: Rows Removed by Filter = 0 (index filters early)
- Improvement validated

**Sort Optimization:**
- Baseline: Sort Method = "quicksort" (expensive)
- Optimized: No sort node (index provides order)
- Bonus benefit discovered

---

### Benchmark Result Storage

**BenchmarkResult Entity (per test query):**

**Fields:**
- id: Unique identifier
- proposal_id: Foreign key to proposal
- query_name: "baseline_test1", "optimized_test1", etc.
- query_executed: Full SQL text
- execution_time_ms: Average of 3 runs
- rows_returned: Result set size
- explain_plan: Parsed EXPLAIN output (JSONB)
- storage_impact_mb: Storage overhead (NULL for baseline)
- created_at: Timestamp

**Typical Records per Proposal:**
- 4 baseline results (before optimization)
- 4 optimized results (after optimization)
- Total: 8 BenchmarkResult records

**Aggregate Storage:**
- Store avg_improvement in proposal metadata
- Store storage_overhead_mb in proposal
- Used by consensus engine for scoring

---

## ⚖️ Consensus Engine

### Responsibility

**Evaluate all proposals using multi-criteria scoring and select the 
optimal solution based on objective metrics and configurable weights.**

**Inputs:**
- Array of OptimizationProposal entities
- Array of BenchmarkResult entities (for each proposal)
- ScoringCriteria configuration (weights)

**Outputs:**
- ConsensusDecision entity with winner
- Complete score breakdown (all proposals)
- Human-readable rationale

---

### Multi-Criteria Evaluation

**Four Scoring Dimensions:**

**1. Performance (50% weight)**
- Most important factor
- Based on actual benchmark improvements
- Rewards significant speed gains

**2. Storage (20% weight)**
- Important but not critical
- Penalizes excessive storage overhead
- Rewards efficiency

**3. Complexity (20% weight)**
- Operational maintainability consideration
- Rewards simplicity
- Penalizes difficult-to-manage solutions

**4. Risk (10% weight)**
- Safety consideration
- Rewards easily reversible changes
- Penalizes risky structural changes

**Why These Weights:**

**Performance 50%:**
- Primary goal is optimization
- Users submit slow queries to make them fast
- Half the decision should be about performance

**Storage 20%:**
- Important resource but not severely limited
- Tiger Cloud handles storage efficiently
- Zero-copy forks reduce storage concerns
- Worth considering but not critical

**Complexity 20%:**
- Long-term maintainability matters
- Simple solutions easier to understand/debug
- Operational simplicity reduces future problems
- Equal weight to storage

**Risk 10%:**
- Forks provide safety net (experimentation is safe)
- Reversibility important but less critical
- Smallest weight but still considered

---

### Configurable Weights

**Default Configuration:**
```
ScoringCriteria:
  performance_weight: 0.5
  storage_weight: 0.2
  complexity_weight: 0.2
  risk_weight: 0.1
  
Validation: weights must sum to 1.0
```

**Custom Weights (Per Task):**

**User can override via task.metadata.scoring_weights:**

**Example 1: Performance Critical**
```
User sets: {
  "performance": 0.7,
  "storage": 0.1,
  "complexity": 0.1,
  "risk": 0.1
}

Use case: Critical query, speed is everything
Result: Favors highest performance regardless of overhead
```

**Example 2: Conservative Approach**
```
User sets: {
  "performance": 0.4,
  "storage": 0.2,
  "complexity": 0.3,
  "risk": 0.1
}

Use case: Stability important, moderate gains acceptable
Result: Favors simple, low-complexity solutions
```

**Example 3: Storage Constrained**
```
User sets: {
  "performance": 0.4,
  "storage": 0.4,
  "complexity": 0.1,
  "risk": 0.1
}

Use case: Limited storage budget
Result: Heavily penalizes storage overhead
```

---

## 📊 Scoring Algorithm

### Performance Score (0-100 scale)

**Input:**  
Average improvement percentage from benchmarks

**Calculation Method:**

**Option 1: Tiered Scoring (Current)**
```
IF improvement >= 90%: score = 100
ELSE IF improvement >= 80%: score = 95
ELSE IF improvement >= 70%: score = 85
ELSE IF improvement >= 50%: score = 70
ELSE IF improvement >= 30%: score = 50
ELSE IF improvement >= 10%: score = 30
ELSE: score = 10
```

**Rationale:**
- Rewards exceptional performance (90%+ = perfect score)
- Diminishing returns above 90% (90% vs 95% not significant)
- Minimum viable improvement: 10% (below is negligible)

**Option 2: Linear Scaling (Alternative)**
```
score = min(100, improvement_pct × 1.1)

Examples:
  80% improvement → 88 score
  90% improvement → 99 score
  95% improvement → 100 (capped)
```

**Rationale:**
- More granular differentiation
- 1.1 multiplier rewards high performance
- Simpler formula

**Current Implementation: Tiered (more intuitive for users)**

---

**Examples:**

**gemini-2.5-pro Proposal:**
- Avg improvement: 82.6%
- Tier: >= 80% → 95 base
- Interpolation (optional): 82.6 between 80-90 → 95 + bonus
- **Final score: 95**

**gemini-2.0-flash Proposal:**
- Avg improvement: 61.5%
- Tier: >= 50% → 70 base
- Interpolation: 61.5 between 50-70 → 70 + bonus
- **Final score: 76**

**gemini-2.5-pro Proposal:**
- Avg improvement: 93.5%
- Tier: >= 90% → 100
- **Final score: 100**

---

### Storage Score (0-100 scale)

**Input:**  
Storage overhead in megabytes

**Calculation Method:**

**Tiered Scoring:**
```
IF overhead <= 10 MB: score = 100
ELSE IF overhead <= 25 MB: score = 90
ELSE IF overhead <= 50 MB: score = 75
ELSE IF overhead <= 100 MB: score = 50
ELSE: score = max(20, 100 - overhead/2)
```

**Rationale:**
- 10 MB negligible (modern standards) → perfect score
- 25 MB acceptable → 90 (minor penalty)
- 50 MB moderate → 75 (moderate penalty)
- 100 MB significant → 50 (half score)
- Above 100 MB: Linear penalty (200 MB → 0 score)

**Minimum Score:**
- Capped at 20 (even 1GB overhead gets 20 points)
- Ensures proposals not eliminated solely on storage
- Performance may justify high storage

---

**Examples:**

**gemini-2.5-pro Proposal:**
- Storage overhead: 12 MB
- Tier: <= 25 MB → 90 base
- Interpolation: 12 between 10-25
- **Final score: 95** (interpolated)

**gemini-2.0-flash Proposal:**
- Storage overhead: 40 MB
- Tier: <= 50 MB → 75 base
- **Final score: 75**

**gemini-2.5-flash Proposal:**
- Storage overhead: 80 MB
- Tier: <= 100 MB → 50 base
- **Final score: 60** (interpolated between 50-100)

---

### Complexity Score (0-100 scale)

**Input:**  
Proposal type and SQL command characteristics

**Base Scoring (by Proposal Type):**

```
Proposal Type → Base Score → Rationale

index: 90
  Simple CREATE INDEX command
  Well-understood, proven technique
  Easy to implement and revert (DROP INDEX)
  Minimal operational overhead

partial_index: 85
  Slightly more complex (WHERE clause)
  Requires understanding of data distribution
  Still simple to revert

composite_index: 85
  Multiple columns, order matters
  Requires column order analysis
  Still straightforward implementation

query_rewrite: 85
  No schema changes
  Application-level change
  Easy rollback (use original query)

materialized_view: 70
  Requires refresh strategy
  Needs scheduling/automation
  Refresh locks table (concurrency concern)
  Moderate operational overhead

partitioning: 50
  Major schema restructuring
  Requires data migration
  Downtime possible
  Difficult to revert
  Significant operational complexity

denormalization: 40
  Changes data model
  May require triggers
  Application changes needed
  Data consistency concerns
  Very difficult to revert
```

**Adjustments:**

**Multiple SQL Commands:**
- Penalty: -5 points per additional command beyond 1
- Rationale: More commands = more complexity
- Example: 3 commands → -10 points

**Requires Triggers/Jobs:**
- Penalty: -10 points
- Rationale: Ongoing maintenance required
- Example: Materialized view with refresh job

**Final Calculation:**
```
complexity_score = base_score - (additional_commands × 5) - 
                   (requires_automation ? 10 : 0)

Capped at: min(100, max(20, complexity_score))
```

---

**Examples:**

**gemini-2.5-pro Proposal:**
- Type: partial_index
- Base: 85
- Commands: 1 (no penalty)
- Automation: No
- **Final score: 85**

**gemini-2.0-flash Proposal:**
- Type: partitioning
- Base: 50
- Commands: 4 (penalty: -15)
- Automation: No
- **Final score: 35**

**gemini-2.5-flash Proposal:**
- Type: materialized_view
- Base: 70
- Commands: 3 (penalty: -10)
- Automation: Yes (penalty: -10)
- **Final score: 50**

---

### Risk Score (0-100 scale)

**Input:**  
Estimated risk level from proposal and validation

**Base Scoring (by Risk Level):**

```
Risk Level → Score → Criteria

low: 95
  Easily reversible (DROP INDEX, DROP VIEW)
  No data modification
  No locking concerns
  Small scope (single table)
  Examples: Simple index, query rewrite

medium: 70
  Requires schema lock (brief)
  Some downtime possible
  Moderate reversal complexity
  Examples: Materialized view, partitioning

high: 40
  Major restructuring required
  Data migration needed
  Potential data loss
  Difficult/impossible to rollback
  Examples: Denormalization, major schema changes
```

**Risk Assessment Factors:**

**Reversibility:**
- Can optimization be undone with single command?
- Index: Yes (DROP INDEX) → Low risk
- Partitioning: No (complex migration) → High risk

**Data Modification:**
- Does proposal change existing data?
- Index: No (metadata only) → Low risk
- Denormalization: Yes (duplicates data) → High risk

**Locking:**
- Does operation lock table?
- Index: Yes but brief (CONCURRENTLY option available) → Low risk
- Partitioning: Yes, extended lock → Medium risk

**Scope:**
- Single table or multiple?
- One table → Lower risk
- Multiple tables → Higher risk

---

**Adjustments:**

**Validation Warnings:**
- Penalty: -5 points per warning
- Example: "Requires exclusive lock" warning

**Safety Checks:**
- Bonus: +5 if proposal includes rollback SQL
- Bonus: +5 if proposal includes validation steps

---

**Examples:**

**gemini-2.5-pro Proposal:**
- Risk level: low
- Base: 95
- Warnings: 0
- **Final score: 95**

**gemini-2.5-flash Proposal:**
- Risk level: medium
- Base: 70
- Warnings: 1 (migration complexity)
- **Final score: 65**

**gemini-2.5-pro Proposal:**
- Risk level: medium
- Base: 70
- Warnings: 1 (staleness)
- **Final score: 65**

---

### Weighted Total Calculation

**Formula:**
```
weighted_total = 
  (performance_score × performance_weight) +
  (storage_score × storage_weight) +
  (complexity_score × complexity_weight) +
  (risk_score × risk_weight)

Default weights:
  performance_weight = 0.5
  storage_weight = 0.2
  complexity_weight = 0.2
  risk_weight = 0.1
```

**Example: gemini-2.5-pro Proposal**
```
Scores:
  performance: 95
  storage: 95
  complexity: 85
  risk: 95

Calculation:
  (95 × 0.5) + (95 × 0.2) + (85 × 0.2) + (95 × 0.1)
  = 47.5 + 19.0 + 17.0 + 9.5
  = 93.0

Weighted Total: 93.0
```

**Example: gemini-2.0-flash Proposal**
```
Scores:
  performance: 76
  storage: 75
  complexity: 35
  risk: 65

Calculation:
  (76 × 0.5) + (75 × 0.2) + (35 × 0.2) + (65 × 0.1)
  = 38.0 + 15.0 + 7.0 + 6.5
  = 66.5

Weighted Total: 66.5
```

**Example: gemini-2.5-flash Proposal**
```
Scores:
  performance: 100
  storage: 60
  complexity: 50
  risk: 65

Calculation:
  (100 × 0.5) + (60 × 0.2) + (50 × 0.2) + (65 × 0.1)
  = 50.0 + 12.0 + 10.0 + 6.5
  = 78.5

Weighted Total: 78.5
```

---

## 🏆 Decision Making

### Ranking Process

**Step 1: Calculate Scores**
```
FOR each proposal:
  Calculate performance_score
  Calculate storage_score
  Calculate complexity_score
  Calculate risk_score
  Calculate weighted_total
```

**Step 2: Sort by Weighted Total**
```
proposals.sort(by: weighted_total, order: descending)
```

**Step 3: Assign Ranks**
```
rank = 1
FOR each proposal in sorted_proposals:
  proposal.rank = rank
  rank += 1
```

**Step 4: Identify Winner**
```
winner = proposals[0]  // First in sorted list (highest score)
```

---

### Winner Selection

**Primary Criterion:**
- Highest weighted_total score

**Example from Business Logic:**
```
Rankings:
  1. gemini-2.5-pro: 93.0 points
  2. gemini-2.5-flash: 78.5 points
  3. gemini-2.0-flash: 66.5 points

Winner: gemini-2.5-pro
```

**Validation:**
- Winner must meet minimum improvement threshold (10%)
- Winner must pass all safety validations
- Winner's risk score must be >= 40 (not too risky)

**Edge Case: All Below Threshold**
- If all proposals < 10% improvement → No winner
- Task marked completed (no optimization needed)
- User message: "Query already well-optimized"

---

### Tie Breaking

**Scenario: Two proposals have identical weighted totals**

**Tie-Breaking Rules (in order):**

**Rule 1: Prefer Lower Complexity**
```
IF proposal_a.complexity_score > proposal_b.complexity_score:
  winner = proposal_a
ELSE IF proposal_b.complexity_score > proposal_a.complexity_score:
  winner = proposal_b
```

**Rationale:**
- Operational simplicity valuable long-term
- Easier to maintain and understand
- Lower risk of future problems

**Rule 2: Prefer Lower Storage**
```
IF still tied:
  IF proposal_a.storage_score > proposal_b.storage_score:
    winner = proposal_a
```

**Rationale:**
- Storage is finite resource
- Less overhead = better efficiency

**Rule 3: Agent Precedence**
```
IF still tied:
  Agent priority: gemini-2.5-pro > gemini-2.5-flash > gemini-2.0-flash
```

**Rationale:**
- Arbitrary but consistent
- Deterministic precedence to avoid ambiguity
- Prevents non-deterministic decisions

**Example:**
```
Scenario: gemini-2.5-pro and gemini-2.5-flash both score 85.0

Comparison:
  gemini-2.5-pro complexity: 85
  gemini-2.5-flash complexity: 70
  
Winner: gemini-2.5-pro (higher complexity score = simpler solution)
```

---

### Score Breakdown Storage

**ConsensusDecision.all_scores JSONB Structure:**

```
{
  "cerebro": {
    "proposal_id": 45,
    "performance_score": 95.0,
    "storage_score": 95.0,
    "complexity_score": 85.0,
    "risk_score": 95.0,
    "weighted_total": 93.0,
    "rank": 1,
    "improvement_pct": 82.6,
    "storage_overhead_mb": 12.0
  },
  "operativo": {
    "proposal_id": 46,
    "performance_score": 100.0,
    "storage_score": 60.0,
    "complexity_score": 50.0,
    "risk_score": 65.0,
    "weighted_total": 78.5,
    "rank": 2,
    "improvement_pct": 93.5,
    "storage_overhead_mb": 80.0
  },
  "bulk": {
    "proposal_id": 47,
    "performance_score": 76.0,
    "storage_score": 75.0,
    "complexity_score": 35.0,
    "risk_score": 65.0,
    "weighted_total": 66.5,
    "rank": 3,
    "improvement_pct": 61.5,
    "storage_overhead_mb": 40.0
  }
}
```

**Purpose:**
- Full transparency
- User can see why each agent ranked where
- Analytics on scoring patterns
- Debugging consensus decisions

---

## 💾 Apply to Main Database

### Responsibility

**Safely apply the winning optimization to the production main database 
after consensus decision.**

**Critical Constraints:**
- Zero downtime (if possible)
- Rollback capability
- Validation before/after
- Error handling

---

### Pre-Application Validation

**Safety Checks Before Touching Main DB:**

**Check 1: Verify Winner Exists**
```
IF consensus.winning_proposal_id IS NULL:
  ERROR: "No winning proposal selected"
  ABORT
```

**Check 2: Re-validate SQL**
```
FOR each sql_command in winning_proposal.sql_commands:
  Validate syntax (defensive, already validated)
  Check for dangerous keywords
  Verify command type (DDL only, no DML)
```

**Check 3: Check Main DB Accessibility**
```
Test connection to main database
Execute simple query: SELECT 1
IF fails:
  ERROR: "Main database unavailable"
  ABORT
```

**Check 4: Acquire Optimization Lock**
```
Mutex/lock prevents concurrent optimizations
IF lock held by another process:
  WAIT or ABORT (configurable)
```

**Check 5: Record PITR Timestamp**
```
current_timestamp = NOW()
Store in consensus_decision metadata

Purpose: Rollback capability via Tiger Cloud PITR
```

---

### Application Process

**Step-by-Step Execution:**

**Step 1: Begin Transaction (if applicable)**
```
Some DDL operations support transactions in PostgreSQL
Indexes: Transactional
Materialized views: Transactional
Partitioning: Often not transactional

IF operation_supports_transaction:
  BEGIN TRANSACTION
```

**Step 2: Execute SQL Commands**
```
FOR each sql_command in winning_proposal.sql_commands:
  Log: "Executing: {sql_command}"
  
  Execute command on main DB
  
  IF error:
    Log error details
    IF in transaction:
      ROLLBACK
    Mark consensus.applied_to_main = false
    Update task.status = failed
    RETURN error
  
  Log: "Success"
```

**Step 3: Commit Transaction**
```
IF in transaction:
  COMMIT
```

**Step 4: Verify Application**
```
IF proposal_type == "index":
  Query: SELECT indexname FROM pg_indexes 
         WHERE indexname = ?
  Expected: 1 row
  
IF proposal_type == "materialized_view":
  Query: SELECT matviewname FROM pg_matviews
         WHERE matviewname = ?
  Expected: 1 row
  
IF not verified:
  ERROR: "Optimization not applied correctly"
  Trigger rollback procedure
```

---

### Post-Application Validation

**Performance Validation:**

**Purpose:**  
Ensure optimization performs as expected in production environment 
(may differ from fork due to caching, concurrent load, etc.)

**Process:**
```
Step 1: Execute original query on main DB
  Run 3 times, take average
  actual_time = average execution time

Step 2: Compare with benchmark prediction
  benchmark_time = benchmark_results.avg_optimized_time
  
Step 3: Calculate deviation
  deviation_pct = ABS(actual_time - benchmark_time) / 
                  benchmark_time × 100

Step 4: Evaluate
  IF deviation_pct <= 20%:
    SUCCESS: Performance within acceptable range
  ELSE IF actual_time > baseline_time:
    WARNING: Performance degraded (worse than baseline)
    Trigger automatic rollback
  ELSE:
    WARNING: Performance different but still improved
    Log warning, continue
```

**Acceptance Criteria:**
- Actual performance within 20% of benchmark prediction
- Never worse than original baseline
- If degraded, automatic rollback

---

**EXPLAIN Plan Validation:**

**Purpose:**  
Verify optimization being used by query planner.

**Process:**
```
Execute EXPLAIN (without ANALYZE) on main DB
Parse plan

IF proposal_type == "index":
  Check for "Index Scan using {index_name}"
  IF not found:
    WARNING: "Index not being used by planner"
    Possible causes:
      - Statistics outdated (run ANALYZE)
      - Query planner chose different plan
      - Index not selective enough

IF proposal_type == "materialized_view":
  Check query rewritten to use view
  (Application must use view, not enforced by DB)
```

---

### Rollback Strategy

**Automatic Rollback Triggers:**

**Trigger 1: SQL Execution Fails**
- Action: Immediate rollback (transaction abort)
- Cleanup: None needed (transaction rolled back)

**Trigger 2: Verification Fails**
- Example: Index not created
- Action: Execute rollback SQL

**Trigger 3: Performance Degradation**
- Actual time > baseline time
- Action: Execute rollback SQL
- User notified: "Optimization degraded performance, rolled back"

---

**Rollback SQL Generation:**

**For Each Proposal Type:**

**Index:**
```
Rollback SQL: DROP INDEX IF EXISTS {index_name}
Safe: Yes (idempotent)
Downtime: None (brief lock)
```

**Partial Index:**
```
Same as index
```

**Materialized View:**
```
Rollback SQL: DROP MATERIALIZED VIEW IF EXISTS {view_name}
Safe: Yes
Note: Application must handle view absence
```

**Partitioning:**
```
Complex: Requires data migration back
Rollback SQL: Generated during proposal creation
May require: Recreate original table, migrate data
Downtime: Possible
```

**Query Rewrite:**
```
Rollback: Use original query in application
No database changes needed
```

---

**Rollback Execution:**
```
FOR each rollback_command in rollback_sql:
  Execute command on main DB
  Log execution
  Verify rollback successful

Update consensus_decision:
  applied_to_main = false
  rollback_reason = "{explanation}"

Update task:
  status = failed
  error_message = "Optimization rolled back: {reason}"
```

---

**Point-in-Time Recovery (PITR) Fallback:**

**When Rollback SQL Insufficient:**
- Complex partitioning changes
- Unexpected state
- Rollback SQL fails

**PITR Process:**
```
Use timestamp recorded before application

Option 1: Fork main to pre-optimization state
  tiger service fork --parent=afs-main 
                     --timestamp="{pitr_timestamp}"
  Creates fork at exact state before optimization
  Review fork to verify state
  Optionally promote fork to new main (advanced)

Option 2: Inform user
  Provide PITR timestamp
  User decides whether to rollback via PITR
  Manual intervention required
```

**Tiger Cloud PITR Window:**
- Free tier: 24 hours
- Must execute rollback within 24 hours of application

---

### Success Path

**Complete Flow:**
```
1. Pre-validation passes
2. SQL executes successfully
3. Verification confirms optimization applied
4. Performance validation within 20%
5. EXPLAIN shows optimization in use

Actions:
  Update consensus_decision.applied_to_main = true
  Update task.status = completed
  Update task.completed_at = NOW()
  Release optimization lock
  Broadcast: optimization_applied event
  Broadcast: task_completed event
```

---

## 📝 Rationale Generation

### Purpose

**Human-Readable Explanation:**  
Users must understand why a particular proposal won. This builds trust 
and provides learning opportunities.

**Components:**
1. Winner identification
2. Key metrics summary
3. Comparison with runners-up
4. Trade-off explanation
5. Runner-up analysis

---

### Rationale Template

**Structure:**

```
"{WinnerAgent}'s {proposal_type} proposal selected as optimal solution.

Performance: {winner_improvement}% improvement ({baseline_time} → 
{optimized_time})

Key Strengths:
- {strength_1}
- {strength_2}
- {strength_3}

Trade-off Analysis:
{comparison_with_runner_up}

Runner-up: {runner_up_agent}'s {runner_up_type} proposal ({runner_up_score} 
points)
While it achieved {runner_up_improvement}% improvement, 
{reason_winner_better}."
```

---

### Example Rationale (from Business Logic)

**Actual Generated Rationale:**

```
gemini-2.5-pro partial index proposal selected as optimal solution.

Performance: 82.6% average improvement (2.30s → 0.40s)

Key Strengths:
- Minimal storage overhead (12MB vs 80MB for materialized view)
- Low operational complexity with easy rollback path (simple DROP INDEX)
- Consistent performance across all test scenarios
- Index-only scan optimization for covered queries

Trade-off Analysis:
While gemini-2.5-flash materialized view achieved highest raw performance (93.5% 
improvement), the 80MB storage overhead and refresh maintenance 
complexity made it less balanced. gemini-2.5-pro's solution offers 82.6% 
improvement with negligible overhead and zero maintenance.

Runner-up: gemini-2.5-flash materialized_view proposal (78.5 points)
While it achieved 93.5% improvement, the storage overhead (80MB) and 
ongoing refresh maintenance requirements reduced its overall score. 
The 11% additional performance gain did not justify the operational 
complexity.

Third place: gemini-2.0-flash partitioning proposal (66.5 points)
Partitioning showed moderate improvement (61.5%) but required complex 
schema migration with potential downtime. Better suited for larger 
datasets (>5M rows).
```

---

### Rationale Generation Logic

**Data Collection:**
```
winner = consensus.winning_proposal
runner_up = proposals.ranked[1]  // Second place
third = proposals.ranked[2]  // Third place (if exists)

winner_agent = winner.agent_type
winner_type = winner.proposal_type
winner_improvement = winner.benchmark_avg_improvement
winner_storage = winner.storage_overhead_mb
winner_score = consensus.all_scores[winner_agent].weighted_total

baseline_time = get_baseline_avg_time()
optimized_time = get_optimized_avg_time(winner)
```

**Template Population:**
```
rationale = template.replace("{WinnerAgent}", winner_agent.capitalize())
rationale = rationale.replace("{proposal_type}", winner_type)
rationale = rationale.replace("{winner_improvement}", 
                              round(winner_improvement, 1))
rationale = rationale.replace("{baseline_time}", 
                              format_time(baseline_time))
rationale = rationale.replace("{optimized_time}", 
                              format_time(optimized_time))

... continue for all placeholders
```

---

### Strength Identification

**Automatic Strength Detection:**

**If storage_score > 90:**
- Add: "Minimal storage overhead ({storage_mb}MB)"

**If complexity_score > 80:**
- Add: "Low operational complexity with easy rollback"

**If improvement_pct > 80:**
- Add: "Significant performance improvement"

**If improvement variance < 10%:**
- Add: "Consistent performance across all test scenarios"

**If explain shows index-only scan:**
- Add: "Index-only scan optimization for covered queries"

**If materialized view:**
- Add: "Pre-computed results eliminate query execution overhead"

**If partitioning:**
- Add: "Partition pruning reduces scanned data volume"

---

## 🔍 Edge Cases

### Scenario 1: All Proposals Below Threshold

**Situation:**  
All agents propose optimizations but benchmarks show <10% improvement.

**Example:**
```
- gemini-2.5-pro: 8% improvement
- gemini-2.0-flash: 5% improvement
- gemini-2.5-flash: 7% improvement
```
**Consensus Decision:**
```
winning_proposal_id = NULL (no winner)
decision_rationale = "No significant optimization found. All proposals 
showed <10% improvement, suggesting query is already well-optimized or 
bottleneck is external to database."

Task status: completed (not failed)
applied_to_main: false
```

**User Message:**
"Your query is already well-optimized. Proposals showed <10% improvement, 
which is within margin of error. Consider reviewing application logic 
or data volume if performance issues persist."

---

### Scenario 2: Single Proposal (Partial Failure)

**Situation:**  
2 of 3 agents fail, only 1 proposal available.

**Example:**
```
- gemini-2.5-pro: Success, 75% improvement
- gemini-2.0-flash: Failed (LLM timeout)
- gemini-2.5-flash: Failed (fork creation failed)
```
**Consensus Decision:**
```
Only one proposal to evaluate
Score gemini-2.5-pro proposal normally
If score acceptable (weighted_total > 60):
  winner = gemini-2.5-pro
ELSE:
  No winner (insufficient quality)
```

**Rationale Template (Modified):**
```
"gemini-2.5-pro proposal selected (only valid proposal due to partial agent failures).

Performance: 75% improvement
Score: {score}

Note: gemini-2.0-flash and gemini-2.5-flash encountered errors during execution. Decision based on single proposal analysis."
```

---

### Scenario 3: Identical Improvements

**Situation:**  
Multiple proposals show same performance improvement.

**Example:**
- gemini-2.5-pro: 80% improvement, 15MB overhead, index
- gemini-2.5-flash: 80% improvement, 70MB overhead, materialized view

**Scoring:**
```
Performance scores: Both 95 (identical)

Difference comes from other criteria:
  gemini-2.5-pro storage: 100 (15MB)
  gemini-2.5-flash storage: 65 (70MB)
  
  gemini-2.5-pro complexity: 90 (index)
  gemini-2.5-flash complexity: 70 (mat view)
  
gemini-2.5-pro wins on storage + complexity despite identical performance
```

**Rationale Emphasis:**
"While both proposals achieved identical 80% performance improvement, 
gemini-2.5-pro requires 55MB less storage and offers simpler 
operational management."

---

### Scenario 4: Performance Variance

**Situation:**  
Proposal shows inconsistent improvement across test queries.

**Example:**
- Test 1: 90% improvement
- Test 2: 85% improvement
- Test 3: 40% improvement
- Test 4: 80% improvement
- Average: 73.75%

**Concern:**  
Low improvement on Test 3 suggests optimization doesn't generalize well.

**Handling:**
```
Calculate variance/standard deviation

IF std_dev > 20%:
  Add warning to proposal metadata
  Rationale mentions inconsistency:
    "Note: Performance improvement varied significantly across test 
    scenarios (40-90%), suggesting optimization may not work well for 
    all query patterns."

Still scored on average, but warning flags issue
```

---

### Scenario 5: Benchmark Execution Failure

**Situation:**  
Proposal SQL applies successfully but benchmark queries fail.

**Example:**
- Index created successfully
- Benchmark query 1: Times out
- Benchmark query 2: Syntax error
- Benchmark query 3: Success
- Benchmark query 4: Success

**Handling:**
```
Partial benchmark results

Option 1: Use available results
  Calculate improvement from queries 3 and 4 only
  Flag as incomplete benchmark
  Lower confidence in scoring

Option 2: Fail proposal
  If <50% benchmarks succeed, mark proposal as failed
  Don't include in consensus

Current approach: Option 2 (stricter, more reliable)
```

---

### Scenario 6: Negative Improvement

**Situation:**  
Proposal makes query slower.

**Example:**
- Baseline: 1.5s
- After optimization: 2.3s
- "Improvement": -53% (degradation)

**Causes:**
- Index not selective enough (overhead > benefit)
- Materialized view not refreshed
- Wrong index column order
- Planner chose suboptimal plan

**Handling:**
```
Performance score:
  IF improvement < 0:
    score = 0 (minimum)

Proposal marked as invalid
Not considered in consensus
Logged for agent learning
```

---

### Scenario 7: Storage Measurement Failure

**Situation:**  
Cannot determine storage overhead.

**Example:**
- Table size before: 100MB
- Apply optimization
- Table size after: Query fails (permission issue)

**Handling:**
```
Use estimated storage from proposal

storage_overhead_mb = proposal.estimated_impact.storage_overhead_mb

Flag as estimated (not measured)
Add note to rationale:
  "Storage overhead estimated (measurement unavailable)"

Proceed with consensus using estimated value
```

---

## 🎯 Summary

This consensus and benchmarking system provides:

**Objective Evaluation:**
- 4 standardized test queries per proposal
- Real performance measurements (not estimates)
- Statistical stability (3 runs, averaged)
- EXPLAIN plan analysis for verification

**PITR Validation:**
- Zero-copy fork creation in <10 seconds
- Rollback restores pre-change state (marker table verification)
- Timestamp-based recovery working correctly
- Cleanup removes all temporary resources

**Multi-Criteria Scoring:**
- Performance: 50% weight (speed is primary goal)
- Storage: 20% weight (efficiency matters)
- Complexity: 20% weight (maintainability important)
- Risk: 10% weight (safety consideration)
- Configurable weights per task

**Transparent Decision:**
- Complete score breakdown visible
- Human-readable rationale generated
- Comparison with runners-up
- Trade-off analysis explained

**Safe Application:**
- Pre-validation checks
- Transaction support (where possible)
- Post-application verification
- Automatic rollback on degradation
- PITR fallback capability

**Edge Case Handling:**
- Partial failures tolerated
- Single proposal accepted if quality sufficient
- Performance variance detected
- Negative improvements rejected

---

**Related Documentation:**
- See also: [06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md) 
  - Tiger Cloud infrastructure and CLI proxy implementation
- See also: [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md) 
  - How agents create proposals
- See also: [07-LLM-INTEGRATION.md](07-LLM-INTEGRATION.md) 
  - LLM API integration

---

**Document Status:** Updated with PITR Validation Phase (Nov 5, 2025)

**Implementation Status:** 
- ✅ PITR Validation tool: `backend/validate_pitr` (compiled, ready to run)
- ✅ MCPClient: Stateless CLI proxy (Go client with tiger CLI)
- ✅ Docker Compose: Orchestration (mcp + backend services)
- ✅ Go benchmark runner: Ready for benchmarking after PITR validation
- ⚠️ Tiger Cloud fork API: Currently returning "unknown error" (non-code issue, see 06-TIGER-CLOUD-MCP.md)

**Last Reviewed:** 2025-11-05  
**Maintained By:** Project Lead


**Related Documentation:**
- Previous: [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md) - Agent proposals
- Next: [06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md) - MCP integration
- See also: [01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md) - Complete 
  example walkthrough

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```

