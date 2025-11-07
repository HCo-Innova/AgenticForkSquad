# 01-BUSINESS-LOGIC.md

```markdown
# 01-BUSINESS-LOGIC.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Business Logic Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md)

---

## 📖 Table of Contents

1. [System Actors](#system-actors)
2. [Primary Use Case: Query Optimization](#primary-use-case-query-optimization)
3. [Business Rules](#business-rules)
4. [Task Lifecycle](#task-lifecycle)
5. [Agent Workflow](#agent-workflow)
6. [Decision Making Process](#decision-making-process)
7. [Error Scenarios](#error-scenarios)

---

## 👥 System Actors

### Human Actors

**End User (Database Administrator / Developer)**
- **Role:** Submits slow queries for optimization
- **Capabilities:**
  - Create optimization tasks
  - View task progress in real-time
  - Review optimization proposals
  - Understand consensus decisions
- **Limitations:**
  - Cannot directly control which agents are assigned
  - Cannot override consensus decisions (by design)
  - Cannot access forks directly

### AI Actors

**gemini-2.5-pro**
- **Specialization:** Planificación/razonamiento crítico y QA
- **Strengths:**
  - Desambiguación de requerimientos y división de tareas
  - Revisión de EXPLAIN/SQL y consistencia antes de ejecutar
  - Generación de planes y checklist de validación
  - Excelente razonamiento paso a paso
- **Typical Proposals/Outputs:**
  - Plan de acciones priorizado por riesgo/impacto
  - Checklist de validaciones y rollback
  - Revisiones de SQL con anotaciones
- **When Assigned:**
  - Tareas críticas/alta complejidad o alto impacto
  - Revisión previa a aplicar cambios en DB
  - Verificación de propuestas de otros agentes

**gemini-2.5-flash**
- **Specialization:** Generación de SQL/código y pruebas con baja latencia
- **Strengths:**
  - Generación de índices, reescrituras de consultas y scripts
  - Aplicación de cambios seguros en forks
  - Ejecución de benchmarks y captura de EXPLAIN
  - Alto throughput/costo eficiente
- **Typical Proposals:**
  - Índices (simples, compuestos, parciales)
  - Reescrituras de consultas y optimización de JOINs
  - Scripts de pruebas y medición
- **When Assigned:**
  - La mayoría de las tareas operativas y concurrentes
  - Preparación de propuestas y pruebas de rendimiento

**gemini-2.0-flash**
- **Specialization:** Tareas masivas/repetitivas de bajo riesgo (boilerplate)
- **Strengths:**
  - Refactors simples y formatters
  - Transformaciones mecánicas
  - Máximo ahorro de costo/latencia
- **Typical Proposals/Outputs:**
  - Plantillas de SQL/código
  - Limpieza/normalización de estilos
- **When Assigned:**
  - Ahorro máximo en tareas de bajo impacto
  - Procesamiento en lote

### System Actors

**Task Router**
- **Role:** Intelligent task assignment dispatcher
- **Responsibilities:**
  - Analyze incoming optimization tasks
  - Determine task complexity
  - Select appropriate agents based on rules
  - Balance agent workload
- **Decision Inputs:**
  - Task type (query_optimization, schema_improvement, etc.)
  - Query complexity score
  - Presence of JOINs, aggregations, subqueries
  - Table sizes involved
  - Task priority metadata
  - Current agent availability
- **Outputs:**
  - List of assigned agents (1-3 agents)
  - Assignment rationale (for logging)

**Orchestrator**
- **Role:** Central workflow coordinator
- **Responsibilities:**
  - Manage complete task lifecycle
  - Coordinate parallel agent execution
  - Collect proposals and benchmark results
  - Trigger consensus engine
  - Apply winning optimization
  - Handle cleanup operations
  - Broadcast real-time progress updates
- **Error Handling:**
  - Retry failed operations (with exponential backoff)
  - Continue if partial agent failures (≥1 agent succeeds)
  - Rollback if main DB application fails
  - Cleanup orphaned forks

**Consensus Engine**
- **Role:** Multi-criteria decision maker
- **Responsibilities:**
  - Score all agent proposals
  - Weight criteria appropriately
  - Select optimal proposal
  - Generate decision rationale
  - Ensure transparency in decision
- **Scoring Criteria:**
  - Performance improvement (50% weight)
  - Storage efficiency (20% weight)
  - Implementation complexity (20% weight)
  - Risk level (10% weight)
- **Outputs:**
  - Winning proposal identification
  - Complete score breakdown for all proposals
  - Human-readable rationale
  - Ranking of all proposals

**Benchmark Runner**
- **Role:** Performance validation system
- **Responsibilities:**
  - Execute standardized test queries
  - Measure execution times accurately
  - Analyze EXPLAIN plans
  - Calculate storage impact
  - Ensure result consistency
- **Test Suite:**
  - Original query (baseline)
  - Query with LIMIT (selectivity test)
  - Query with additional filter (scalability test)
  - Query with ORDER BY (sort performance test)
- **Metrics Collected:**
  - Execution time (average of 3 runs)
  - Rows returned
  - EXPLAIN plan analysis
  - Storage overhead (MB)
  - Planning time vs execution time

**WebSocket Hub**
- **Role:** Real-time event broadcaster
- **Responsibilities:**
  - Maintain active client connections
  - Broadcast task progress events
  - Notify on agent status changes
  - Push proposal updates
  - Deliver consensus decisions
- **Event Types:**
  - `task_created`
  - `agents_assigned`
  - `fork_created`
  - `analysis_completed`
  - `proposal_submitted`
  - `benchmark_completed`
  - `consensus_reached`
  - `optimization_applied`
  - `task_completed`

**Tiger Cloud MCP Server**
- **Role:** Database operations provider
- **Responsibilities:**
  - Create database forks (zero-copy)
  - Execute queries in specific forks
  - Provide schema information
  - Manage fork lifecycle
  - Handle Point-in-Time Recovery
- **Capabilities:**
  - Fork creation in <10 seconds
  - Isolated query execution per fork
  - Shared storage (fluid storage)
  - 24-hour PITR window

---

## 🎯 Primary Use Case: Query Optimization

### Scenario Description

**Goal:** Optimize a slow monthly revenue report query  
**Current Performance:** 2.3 seconds  
**Target:** Sub-second execution  
**User:** Database Administrator

### End-to-End Flow (Step by Step)

#### **Step 1: Task Submission**

**Actor:** End User

**Actions:**
1. User opens AFS web dashboard
2. Navigates to "New Optimization Task" page
3. Selects task type: "Query Optimization"
4. Pastes slow SQL query:
   ```
   SELECT u.email, SUM(o.total) as revenue
   FROM users u 
   JOIN orders o ON u.id = o.user_id
   WHERE o.status = 'completed'
   GROUP BY u.email
   ORDER BY revenue DESC
   ```
5. Adds description: "Monthly revenue report - takes 2.3 seconds"
6. Optionally sets priority: "High"
7. Clicks "Submit for Optimization"

**System Response:**
- Validates query syntax
- Creates Task record (ID: 123, Status: pending)
- Returns task ID to user
- Displays "Task queued" message
- Redirects to task detail page

**Business Rules Applied:**
- Query must be valid SQL
- Description max 500 characters
- Task type must be valid enum value
- User cannot submit empty query

---

#### **Step 2: Task Routing**

**Actor:** Task Router

**Analysis Process:**
1. Retrieve task 123 from database
2. Parse target query to extract features:
   - Contains 1 JOIN → complexity +1
   - Contains 1 GROUP BY aggregation → complexity +1
   - Contains 1 ORDER BY → complexity +1
   - **Total complexity score:** 3 (medium)
3. Check query characteristics:
   - Has JOIN: TRUE
   - Has aggregation: TRUE
   - Target table: orders (10,000 rows)
4. Check task metadata:
   - Priority: "high"
5. Apply routing rules (in order):
   - Rule: "high priority → assign all agents" → MATCH
   - Agents selected: [gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash]
6. Verify agent availability:
   - gemini-2.5-pro: 1 active task → OK
   - gemini-2.5-flash: 0 active tasks → OK
   - gemini-2.0-flash: 2 active tasks → OK (below limit of 3)
7. Final assignment: All 3 agents

**Assignment Rationale Generated:**
"High priority task with JOIN and aggregation. Assigned all agents for comprehensive analysis."

**System Actions:**
- Update task status: pending → in_progress
- Create 3 agent_execution records (one per agent)
- Trigger Orchestrator with assigned agents
- Broadcast WebSocket event: `agents_assigned`

**Business Rules Applied:**
- High priority always gets maximum agent coverage
- No agent can have >3 concurrent tasks
- At least 1 agent must be assigned per task
- Router decision must be logged

---

#### **Step 3: Parallel Agent Execution**

**Actor:** Orchestrator

**Coordination Process:**
1. Receive task 123 with agent list: [gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash]
2. Spawn 3 concurrent goroutines (one per agent)
3. Each goroutine executes identical workflow:
   - Create fork
   - Analyze task
   - Propose optimization
   - Run benchmark
4. Wait for all agents to complete (or timeout after 10 minutes)
5. Collect all results
6. Proceed to consensus

**Parallel Execution (All Happening Simultaneously):**

---

**gemini-2.5-pro Execution (via Vertex AI):**

**Step 3a: Fork Creation**

**Actions:**
1. Generate unique fork name: `afs-fork-gemini-2.5-pro-task123-1699901234`
2. Call Tiger MCP: `service_fork` with parent: `afs-main`
3. Receive fork service ID
4. Update agent_execution record with fork_id
5. Broadcast: `fork_created` (agent: gemini-2.5-pro)

**Timing:** ~5 seconds (zero-copy)

**Step 3b: Analysis**

**Actions:**
1. Connect to fork: `afs-fork-gemini-2.5-pro-task123`
2. Execute EXPLAIN ANALYZE on original query
3. Parse explain plan output:
   - Detects: Sequential Scan on orders table
   - Detects: Hash Join with users table
   - Detects: Sort operation for ORDER BY
   - Planning time: 0.5ms
   - Execution time: 2,300ms
   - Rows processed: 10,000 orders → 7,500 completed → 1,000 users
4. Build prompt for Vertex AI Client (model: gemini-2.5-pro) with context:
   - "You are a PostgreSQL expert DBA"
   - Include: explain plan, query, schema info
   - Request: JSON analysis with insights and actions
5. Call Vertex AI Client (model: gemini-2.5-pro)
6. Receive and parse response:
   ```
   {
     "insights": [
       "Sequential scan on orders table processing 10k rows",
       "75% of rows match status='completed' filter",
       "No index on (user_id, status) causing inefficient JOIN",
       "Sort operation expensive due to lack of index"
     ],
     "proposed_actions": [
       "Create partial index on orders(user_id, status) WHERE status='completed'",
       "Index will cover JOIN lookup and filter simultaneously"
     ],
     "confidence": 0.92
   }
   ```

**Timing:** ~30 seconds (LLM call)

**Step 3c: Proposal Creation**

**Actions:**
1. Based on analysis, determine proposal type: INDEX
2. Generate specific SQL:
   ```
   CREATE INDEX idx_orders_user_completed 
   ON orders(user_id, status) 
   WHERE status = 'completed'
   ```
3. Validate SQL syntax
4. Estimate impact:
   - Query time improvement: 75-80% (based on explain analysis)
   - Storage overhead: ~12MB (calculated from table statistics)
   - Complexity: LOW (simple index, easily reversible)
   - Risk: LOW (partial index, no data modification)
5. Create OptimizationProposal record
6. Broadcast: `proposal_submitted` (agent: gemini-2.5-pro)

**Rationale Generated:**
"Partial index targets only completed orders (75% of data), reducing index size. Covers both JOIN condition (user_id) and filter (status), enabling index-only scans."

**Timing:** ~10 seconds

**Step 3d: Benchmark Execution**

**Actions:**
1. Measure baseline (before applying proposal):
   - Run original query 3 times: [2.31s, 2.28s, 2.32s]
   - Average baseline: 2.30s
2. Apply optimization SQL in fork
3. Verify index created successfully
4. Run benchmark suite:

   **Test 1: Original query**
   - Run 3 times: [0.44s, 0.46s, 0.45s]
   - Average: 0.45s
   - Improvement: 80.4%
   - EXPLAIN shows: Index Scan using idx_orders_user_completed

   **Test 2: Query with LIMIT 10**
   - Run 3 times: [0.11s, 0.13s, 0.12s]
   - Average: 0.12s
   - Baseline was: 0.80s
   - Improvement: 85%

   **Test 3: Query with date filter (last 30 days)**
   - Run 3 times: [0.18s, 0.19s, 0.18s]
   - Average: 0.18s
   - Baseline was: 1.20s
   - Improvement: 85%

   **Test 4: Query with ORDER BY revenue DESC**
   - Run 3 times: [0.46s, 0.47s, 0.45s]
   - Average: 0.46s
   - Improvement: 80%

5. Measure storage impact:
   - Table size before: 850MB
   - Table size after: 862MB
   - Storage overhead: 12MB

6. Create 4 BenchmarkResult records (one per test query)
7. Broadcast: `benchmark_completed` (agent: gemini-2.5-pro)

**Aggregate Metrics Calculated:**
- Average improvement: 82.6%
- Storage overhead: 12MB
- Consistency: All tests show 80-85% improvement (good consistency)

**Timing:** ~2 minutes (query executions)

**Total gemini-2.5-pro Time:** ~3 minutes

---

**gemini-2.5-flash Execution (via Vertex AI):**

**Steps 3a-3b:** Similar to gemini-2.5-pro (fork creation + analysis)

**Step 3c: Proposal Creation**

**gemini-2.5-flash Analysis:**
- Detects 75% of orders are 'completed'
- Proposes partitioning strategy
- Reasoning: Partition pruning can eliminate 25% of data from scans

**SQL Generated:**
```
-- Convert orders to partitioned table
CREATE TABLE orders_partitioned (LIKE orders INCLUDING ALL)
  PARTITION BY LIST (status);

CREATE TABLE orders_completed PARTITION OF orders_partitioned
  FOR VALUES IN ('completed');

CREATE TABLE orders_other PARTITION OF orders_partitioned
  FOR VALUES IN ('pending', 'processing');

-- Migrate data (conceptual - would need careful execution)
INSERT INTO orders_partitioned SELECT * FROM orders;
```

**Estimated Impact:**
- Query improvement: 50-60%
- Storage overhead: 40MB (partition metadata)
- Complexity: HIGH (table restructuring, requires migration)
- Risk: MEDIUM (hard to revert, potential downtime)

**Rationale:**
"Partitioning by status enables partition pruning, automatically excluding 'pending' and 'processing' partitions from query scans. Scales better for future growth."

**Step 3d: Benchmark Results**

**Aggregate Metrics:**
- Average improvement: 61.5%
- Storage overhead: 40MB
- Consistency: Moderate (improvement varies by query type)

**Total gemini-2.5-flash Time:** ~3 minutos

---

**gemini-2.0-flash Execution (via Vertex AI):**

**Steps 3a-3b:** Similar to others

**Step 3c: Proposal Creation**

**gemini-2.0-flash Analysis:**
- Identifies this as frequently executed query (assumes based on query pattern)
- Proposes materialized view with pre-aggregation
- Trade-off: Extreme read performance vs write overhead

**SQL Generated:**
```
CREATE MATERIALIZED VIEW mv_user_revenue AS
SELECT u.email, SUM(o.total) as revenue
FROM users u 
JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed'
GROUP BY u.email;

CREATE UNIQUE INDEX idx_mv_user_revenue_email 
ON mv_user_revenue(email);

CREATE INDEX idx_mv_user_revenue_revenue 
ON mv_user_revenue(revenue DESC);

-- Refresh strategy (would need scheduled job)
-- REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_revenue;
```

**Estimated Impact:**
- Query improvement: 90-95% (pre-computed results)
- Storage overhead: 80MB (full materialized data)
- Complexity: MEDIUM (needs refresh strategy)
- Risk: MEDIUM (staleness risk, refresh overhead)

**Rationale:**
"Materialized view pre-computes aggregation, turning complex query into simple table scan. Best for frequently-read, infrequently-updated data. Requires concurrent refresh strategy to minimize locking."

**Step 3d: Benchmark Results**

**Key Finding:**
- All queries return ~0.15s (consistent, no variation)
- Massive improvement but results slightly stale (last refresh)

**Aggregate Metrics:**
- Average improvement: 93.5%
- Storage overhead: 80MB
- Note: Performance consistent regardless of filters (reading pre-computed data)

**Total gemini-2.0-flash Time:** ~3 minutes

---

#### **Step 4: Consensus Decision**

**Actor:** Consensus Engine

**Input Data:**
- 3 Proposals with benchmark results
- Scoring criteria configuration (default weights)

**Scoring Process:**

**gemini-2.5-pro Proposal Scoring:**

1. **Performance Score:**
   - Average improvement: 82.6%
   - Formula: `min(100, improvement_pct × 1.1)`
   - Calculation: `min(100, 82.6 × 1.1) = 90.86`
   - **Score: 90.86**

2. **Storage Score:**
   - Storage overhead: 12MB
   - Rule: overhead ≤ 10MB = 100, ≤ 25MB = 90
   - Interpolation: `90 + (25-12)/(25-10) × 10 = 98.67`
   - **Score: 98.67**

3. **Complexity Score:**
   - Proposal type: INDEX
   - Mapping: index = 90 (simple, reversible)
   - SQL commands count: 1 (no penalty)
   - **Score: 90**

4. **Risk Score:**
   - Estimated risk: LOW
   - Mapping: low = 95
   - No validation warnings
   - **Score: 95**

5. **Weighted Total:**
   ```
   (90.86 × 0.5) + (98.67 × 0.2) + (90 × 0.2) + (95 × 0.1)
   = 45.43 + 19.73 + 18.00 + 9.50
   = 92.66
   ```
   **Final Score: 92.66**

---

**gemini-2.0-flash Proposal Scoring:**

1. Performance: 61.5 → **Score: 67.65**
2. Storage: 40MB → **Score: 75** (moderate overhead)
3. Complexity: PARTITIONING → **Score: 50** (major restructuring)
4. Risk: MEDIUM → **Score: 70**
5. **Weighted Total: 64.83**

---

**gemini-2.5-flash Proposal Scoring:**

1. Performance: 93.5% → **Score: 100** (capped)
2. Storage: 80MB → **Score: 60** (significant overhead)
3. Complexity: MATERIALIZED_VIEW → **Score: 70**
4. Risk: MEDIUM → **Score: 70** (staleness concern)
5. **Weighted Total: 81.00**

---

**Ranking:**
1. 🥇 gemini-2.5-pro: 92.66
2. 🥈 gemini-2.5-flash: 81.00
3. 🥉 gemini-2.0-flash: 64.83

**Winner: gemini-2.5-pro**

**Rationale Generated:**

"gemini-2.5-pro partial index proposal selected as optimal solution.

**Performance:** 82.6% average improvement (2.30s → 0.40s)  
**Efficiency:** Minimal storage overhead (12MB vs 80MB for materialized view)  
**Simplicity:** Low complexity with easy rollback (simple DROP INDEX)  
**Safety:** Low risk implementation, no data modification

**Runner-up Analysis:**
gemini-2.5-flash materialized view achieved highest performance (93.5%) but incurred 80MB storage overhead and introduces data staleness concerns requiring refresh management. gemini-2.0-flash partitioning strategy showed moderate improvement (61.5%) but requires complex table restructuring with migration downtime.

**Decision:** gemini-2.5-pro offers the best balance of performance gain, operational simplicity, and minimal risk."

**System Actions:**
- Create ConsensusDecision record
- Mark winning_proposal_id: gemini-2.5-pro proposal
- Store all_scores JSONB
- Broadcast: `consensus_reached`

**Timing:** ~2 seconds (calculation)

---

#### **Step 5: Apply to Main Database**

**Actor:** Orchestrator

**Pre-Application Checks:**
1. Validate winning proposal SQL syntax (defensive check)
2. Verify main DB connection active
3. Acquire optimization lock (prevent concurrent optimizations)
4. Record current timestamp for PITR rollback capability

**Application Process:**
1. Begin transaction (if DDL allows)
2. Execute SQL command:
   ```
   CREATE INDEX idx_orders_user_completed 
   ON orders(user_id, status) 
   WHERE status = 'completed'
   ```
3. Verify index created:
   - Query pg_indexes table
   - Confirm index exists
4. Commit transaction

**Post-Application Validation:**
1. Run original query on main DB
2. Measure execution time: 0.47s
3. Compare with benchmark prediction: 0.45s (within 5% margin ✓)
4. Execute EXPLAIN ANALYZE:
   - Verify Index Scan is used
   - Confirm expected execution plan

**Validation Result:** SUCCESS

**System Actions:**
- Update consensus_decisions.applied_to_main = true
- Update task status: in_progress → completed
- Record completed_at timestamp
- Broadcast: `optimization_applied`

**Timing:** ~5 seconds (index creation)

---

#### **Step 6: Fork Cleanup**

**Actor:** Orchestrator

**Cleanup Process:**
1. Retrieve all fork IDs for task 123:
   - `afs-fork-gemini-2.5-pro-task123-1699901234`
   - `afs-fork-gemini-2.5-flash-task123-1699901235`
   - `afs-fork-gemini-2.0-flash-task123-1699901236`
2. For each fork:
   - Call Tiger MCP: `service_delete`
   - Verify deletion successful
   - Update agent_execution.status = completed
3. Release optimization lock

**Storage Reclaimed:** ~0MB (zero-copy forks, only metadata deleted)

**Timing:** ~3 seconds per fork (sequential deletion)

**System Actions:**
- Broadcast: `task_completed`
- Log final task statistics
- Cleanup temporary data

---

#### **Step 7: User Notification**

**Actor:** WebSocket Hub + UI

**Real-Time Updates Received by User:**

Throughout process, user saw:
1. "Agents assigned: gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash" (Step 2)
2. "gemini-2.5-pro created fork..." (Step 3a)
3. "gemini-2.5-flash created fork..." (Step 3a)
4. "gemini-2.0-flash created fork..." (Step 3a)
5. "gemini-2.5-pro submitted proposal: Partial Index" (Step 3c)
6. "gemini-2.5-flash submitted proposal: Materialized View" (Step 3c)
7. "gemini-2.0-flash submitted proposal: Partitioning" (Step 3c)
8. "Benchmarking completed" (Step 3d)
9. "Consensus reached: gemini-2.5-pro proposal wins" (Step 4)
10. "Optimization applied successfully" (Step 5)
11. "Task completed" (Step 6)

**Final Dashboard Display:**

```
✅ Optimization Completed

Original Performance: 2.30 seconds
New Performance: 0.47 seconds
Improvement: 79.6% faster

Winning Solution: Partial Index (by gemini-2.5-pro)
SQL Applied:
  CREATE INDEX idx_orders_user_completed 
  ON orders(user_id, status) 
  WHERE status = 'completed'

Why This Won:
- Best balance of performance and simplicity
- Minimal storage overhead (12MB)
- Low risk, easily reversible

All Proposals Compared:
  🥇 gemini-2.5-pro (92.66 pts): Partial Index - 82.6% improvement
  🥈 gemini-2.5-flash (81.00 pts): Materialized View - 93.5% improvement
  🥉 gemini-2.0-flash (64.83 pts): Partitioning - 61.5% improvement

[View Detailed Benchmark Results]
[View EXPLAIN Plans]
```

---

### Flow Summary

**Total Time:** ~4-5 minutes end-to-end

**Key Success Factors:**
- Parallel agent execution (3 mins vs 9 mins sequential)
- Zero-copy forks (instant creation vs minutes for full copy)
- Automated benchmarking (objective metrics vs subjective assessment)
- Transparent consensus (clear rationale vs black box decision)

---

## 📏 Business Rules

### Task Management Rules

**Rule 1: Task Validation**
- Query must be non-empty valid SQL
- Description maximum 500 characters
- Task type must be valid enum: `query_optimization`, `schema_improvement`, `index_recommendation`
- Target query must not contain destructive operations (DROP, TRUNCATE, DELETE without WHERE)

**Rule 2: Task Prioritization**
- Priority levels: `low`, `medium`, `high`
- High priority tasks get all available agents
- Low priority tasks get minimum 1 agent
- Default priority is `medium` if not specified

**Rule 3: Task Queuing**
- Maximum 10 pending tasks in queue
- Tasks processed FIFO (First In, First Out)
- User cannot submit duplicate queries within 1 hour
- Task cannot be canceled once agents start executing

### Agent Assignment Rules

**Rule 4: Routing by Complexity**
- Complexity score calculation:
  - +1 per JOIN
  - +1 per subquery
  - +1 per aggregate function (SUM, COUNT, AVG, etc.)
  - +1 per window function
  - +2 per set operation (UNION, INTERSECT, EXCEPT)
- Score 0-2: Assign 1 agent (gemini-2.5-pro preferred)
- Score 3-5: Assign 2 agents (gemini-2.5-pro + gemini-2.5-flash)
- Score 6+: Assign all 3 agents

**Rule 5: Routing by Query Features**
- If query contains JOIN → Always include gemini-2.5-pro
- If query targets table >1M rows → Always include gemini-2.5-flash
- If task metadata indicates "critical" → Always include gemini-2.5-pro
- If query has aggregation + JOIN → Assign gemini-2.5-pro + gemini-2.5-flash minimum

**Rule 6: Agent Availability**
- Agent can handle maximum 3 concurrent tasks
- If agent at capacity, skip and assign alternative
- If no agents available, queue task until capacity frees

**Rule 7: Minimum Assignment**
- Every task must have at least 1 agent assigned
- If routing rules produce empty assignment, default to gemini-2.5-pro

### Fork Management Rules

**Rule 8: Fork Naming**
- Pattern: `afs-fork-{agentType}-task{taskID}-{unixTimestamp}`
- Must be lowercase
- Maximum 100 characters
- Must be globally unique

**Rule 9: Fork Lifecycle**
- Fork must be created before any query execution
- Fork must exist for entire agent workflow
- Fork must be destroyed after benchmark completion
- Maximum fork lifetime: 2 hours (safety timeout)

**Rule 10: Fork Isolation**
- Each agent gets dedicated fork
- No cross-fork data access
- Main database never accessed during experimentation
- Fork queries cannot affect other forks

### Benchmarking Rules

**Rule 11: Benchmark Suite**
- Must execute exactly 4 test queries:
  1. Original query (baseline)
  2. Original with LIMIT 10
  3. Original with additional date filter
  4. Original with ORDER BY
- Each query executed 3 times minimum
- Average of 3 runs used as final metric
- Outliers (>2 std deviations) discarded and re-run

**Rule 12: Baseline Measurement**
- Baseline must be measured BEFORE applying optimization
- Baseline executed in same fork as optimized query
- Minimum 3 runs for baseline stability
- If baseline variance >10%, increase to 5 runs

**Rule 13: Storage Measurement**
- Measure table size before and after optimization
- Include all indexes in size calculation
- Report overhead in MB (2 decimal precision)
- If overhead >100MB, flag as warning

### Consensus Rules

**Rule 14: Scoring Weights**
- Performance: 50% (most important)
- Storage: 20%
- Complexity: 20%
- Risk: 10%
- Weights must sum to 100%
- Weights configurable per task via metadata

**Rule 15: Minimum Improvement**
- Proposal must show minimum 10% improvement to be valid
- If all proposals <10% improvement, reject task (no optimization needed)
- Improvement calculated as: `(baseline - optimized) / baseline × 100`

**Rule 16: Tie Breaking**
- If two proposals have same weighted score (within 0.1 points):
  1. Prefer lower complexity
  2. If still tied, prefer lower storage overhead
  3. If still tied, prefer agent precedence: gemini-2.5-pro > gemini-2.5-flash > gemini-2.0-flash

**Rule 17: Safety Threshold**
- Proposals with risk score <40 are automatically rejected
- Proposals requiring downtime must be flagged in rationale
- Proposals modifying data (not just schema) require explicit user approval

### Application Rules

**Rule 18: Pre-Application Validation**
- Winning proposal SQL must be re-validated
- Main database must be accessible
- No other optimization in progress (mutex lock)
- Transaction support checked (DDL transactionality)

**Rule 19: Post-Application Validation**
- Actual performance must be within 20% of benchmark prediction
- If degradation >10%, automatic rollback triggered
- If validation fails, preserve rollback capability (PITR timestamp)

**Rule 20: Rollback Capability**
- PITR timestamp recorded before any main DB change
- Rollback SQL generated for all DDL operations
- Maximum rollback window: 24 hours (Tiger Cloud limit)
- Manual rollback option available to user

### Error Handling Rules

**Rule 21: Partial Failure Tolerance**
- If 1 of 3 agents fails, continue with 2 agents
- If 2 of 3 agents fail, abort task
- Minimum 1 successful proposal required for consensus
- Failed agent forks must still be cleaned up

**Rule 22: Timeout Handling**
- Agent fork creation timeout: 30 seconds
- Agent analysis timeout: 2 minutes
- Agent proposal timeout: 1 minute
- Agent benchmark timeout: 5 minutes
- Total per-agent timeout: 10 minutes
- If timeout exceeded, mark agent execution as failed

**Rule 23: LLM API Failures**
- Retry LLM calls maximum 3 times
- Exponential backoff: 1s, 2s, 4s
- If all retries fail, mark agent as failed
- Log failure details for debugging
- Do not expose API keys in logs

---

## 🔄 Task Lifecycle

### State Diagram

```
[pending] → [in_progress] → [completed]
              ↓
           [failed]
```

### State Descriptions

**pending**
- Initial state when task created
- Waiting for router assignment
- User can view task but no activity yet
- Transitions to: `in_progress` (when orchestrator starts)

**in_progress**
- Agents are actively working
- Forks created and experiments running
- User sees real-time updates
- Cannot be canceled in this state
- Transitions to: `completed` (success) or `failed` (error)

**completed**
- Optimization successfully applied to main DB
- All forks cleaned up
- Final results available
- Immutable state (cannot change)

**failed**
- Error occurred during processing
- Error message stored in task record
- Partial cleanup performed (forks destroyed)
- User notified with failure reason
- Can be re-submitted as new task

### State Transition Triggers

**pending → in_progress:**
- Trigger: Orchestrator begins execution
- Precondition: At least 1 agent assigned
- Action: Update status, broadcast event

**in_progress → completed:**
- Trigger: Optimization applied and validated
- Precondition: Consensus reached, validation passed
- Action: Update status, set completed_at, cleanup forks

**in_progress → failed:**
- Trigger: Unrecoverable error
- Conditions:
  - All agents failed
  - Consensus impossible (<1 valid proposal)
  - Main DB application failed after retries
  - Timeout exceeded
- Action: Update status, set error message, cleanup forks

---

## 🤖 Agent Workflow

### Individual Agent Steps

**Phase 1: Initialization**
1. Receive task assignment from orchestrator
2. Validate task data completeness
3. Initialize LLM client connection
4. Initialize MCP client connection

**Phase 2: Fork Creation**
1. Generate unique fork name
2. Call MCP `service_fork` operation
3. Wait for fork creation (with timeout)
4. Verify fork accessible
5. Update agent_execution record with fork_id
6. Broadcast fork_created event

**Phase 3: Analysis**
1. Execute EXPLAIN ANALYZE on target query in fork
2. Parse explain plan output (JSON format)
3. Extract key metrics:
   - Scan types (Seq Scan, Index Scan, etc.)
   - Join strategies
   - Filter selectivity
   - Sort operations
   - Timing breakdown
4. Query database statistics:
   - Table sizes
   - Row counts
   - Existing indexes
   - Column data types
5. Build context-rich prompt for LLM
6. Call LLM API with prompt
7. Parse JSON response
8. Validate analysis completeness
9. Store analysis result

**Phase 4: Proposal Generation**
1. Based on analysis insights, determine optimization type
2. Generate specific SQL commands via LLM
3. Validate SQL syntax
4. Estimate impact metrics:
   - Performance improvement (percentage)
   - Storage overhead (MB)
   - Complexity level (low/medium/high)
   - Risk level (low/medium/high)
5. Create optimization_proposals record
6. Broadcast proposal_submitted event

**Phase 5: Benchmarking**
1. Measure baseline performance:
   - Execute original query 3 times
   - Calculate average execution time
   - Store baseline metrics
2. Apply optimization SQL in fork
3. Verify optimization applied successfully
4. Execute benchmark suite:
   - For each test query:
     - Run 3 times
     - Calculate average
     - Capture EXPLAIN plan
     - Count rows returned
5. Measure storage impact:
   - Query table/index sizes
   - Calculate overhead
6. Create benchmark_results records (1 per test query)
7. Calculate aggregate metrics
8. Broadcast benchmark_completed event

**Phase 6: Completion**
1. Mark agent_execution status as completed
2. Set completed_at timestamp
3. Return results to orchestrator
4. Fork remains alive for consensus phase
5. Await cleanup signal

### Agent Coordination

**Parallelization:**
- All agents execute Phases 1-6 concurrently
- No inter-agent communication during execution
- Complete isolation via separate forks
- Orchestrator waits for all agents (or timeout)

**Synchronization Points:**
- After Phase 6: Orchestrator collects all results
- Consensus phase: Sequential (all proposals evaluated together)
- Cleanup phase: Parallel (forks deleted concurrently)

---

## ⚖️ Decision Making Process

### Consensus Algorithm

**Input:**
- N proposals (where N = number of successful agents, 1-3)
- Each proposal has benchmark results
- Scoring criteria configuration

**Process:**

1. **For each proposal:**
   - Calculate performance_score (0-100)
   - Calculate storage_score (0-100)
   - Calculate complexity_score (0-100)
   - Calculate risk_score (0-100)

2. **Apply weights:**
   - weighted_total = (perf × 0.5) + (storage × 0.2) + (complexity × 0.2) + (risk × 0.1)

3. **Rank proposals:**
   - Sort by weighted_total descending
   - Assign rank (1, 2, 3, ...)

4. **Select winner:**
   - Winner = rank 1 proposal
   - If tie, apply tie-breaking rules

5. **Generate rationale:**
   - Explain why winner selected
   - Compare with runners-up
   - Highlight key trade-offs

**Output:**
- Winning proposal ID
- Complete score breakdown for all proposals
- Human-readable rationale
- Ranking of all proposals

### Score Calculation Details

**Performance Score Formula:**
```
improvement_pct = ((baseline_time - optimized_time) / baseline_time) × 100

if improvement_pct >= 90:  score = 100
elif improvement_pct >= 80: score = 95
elif improvement_pct >= 70: score = 85
elif improvement_pct >= 50: score = 70
elif improvement_pct >= 30: score = 50
elif improvement_pct >= 10: score = 30
else: score = 10
```

**Storage Score Formula:**
```
overhead_mb = storage_after - storage_before

if overhead_mb <= 10:   score = 100
elif overhead_mb <= 25:  score = 90
elif overhead_mb <= 50:  score = 75
elif overhead_mb <= 100: score = 50
else: score = max(20, 100 - overhead_mb/2)
```

**Complexity Score Mapping:**
```
index: 90
partial_index: 85
query_rewrite: 85
materialized_view: 70
partitioning: 50
denormalization: 40

Adjustments:
- Each additional SQL command beyond 1: -5 points
- Requires triggers/jobs: -10 points
```

**Risk Score Mapping:**
```
low risk: 95
  - Easily reversible (DROP INDEX, etc.)
  - No data modification
  - Small scope

medium risk: 70
  - Schema lock required
  - Some downtime possible
  - Moderate reversal complexity

high risk: 40
  - Major restructuring
  - Data loss potential
  - Difficult rollback
```

---

## ❌ Error Scenarios

### Scenario 1: Single Agent Fails

**Situation:** An agent's LLM API times out during analysis phase

**System Response:**
1. The agent's agent_execution marked as failed
2. Error logged with details
3. That agent's fork still created (cleanup needed)
4. Other agents continue normally
5. Consensus proceeds with 2 proposals
6. Task completes successfully

**User Impact:** Minimal - still get optimization, just fewer options

**Business Rule Applied:** Rule 21 (Partial Failure Tolerance)

---

### Scenario 2: All Agents Fail

**Situation:** MCP server is down, all fork creations fail

**System Response:**
1. All 3 agent_executions marked as failed
2. No proposals generated
3. Task status → failed
4. Error message: "Unable to create database forks - MCP server unavailable"
5. User notified via WebSocket
6. No cleanup needed (no forks created)

**User Impact:** Task failed, must retry later

**Business Rule Applied:** Rule 21 (minimum 1 proposal required)

---

### Scenario 3: Consensus Cannot Decide

**Situation:** All proposals show <10% improvement

**System Response:**
1. Consensus engine calculates scores
2. Detects all performance scores <30 (below threshold)
3. Rejects all proposals
4. Task status → completed (not failed)
5. User message: "No significant optimization found - query already efficient"
6. Cleanup forks normally

**User Impact:** No optimization applied, but not an error

**Business Rule Applied:** Rule 15 (Minimum Improvement)

---

### Scenario 4: Main DB Application Fails

**Situation:** Winning SQL causes constraint violation on main DB

**System Response:**
1. Consensus completes successfully
2. Application phase begins
3. SQL execution fails with error
4. Transaction automatically rolled back
5. Error logged with full details
6. Retry attempted (Rule 18)
7. If retry fails:
   - Task status → failed
   - Error message stored
   - PITR timestamp preserved for manual rollback
8. Cleanup forks

**User Impact:** Optimization not applied, manual intervention may be needed

**Business Rule Applied:** Rule 19 (Post-Application Validation)

---

### Scenario 5: Performance Degradation After Application

**Situation:** Applied optimization performs worse than benchmark predicted

**System Response:**
1. Optimization applied successfully
2. Validation query executed
3. Actual time: 1.8s (baseline was 2.3s, benchmark predicted 0.5s)
4. Degradation detected: actual is 260% slower than predicted
5. Automatic rollback triggered:
   - Execute rollback SQL (DROP INDEX)
   - Verify rollback successful
   - Re-run validation
6. Task status → failed
7. Error: "Optimization degraded performance in production - rolled back"
8. Cleanup forks

**User Impact:** Database returned to original state safely

**Business Rule Applied:** Rule 19 (20% tolerance threshold)

---

### Scenario 6: Fork Cleanup Fails

**Situation:** Tiger Cloud error prevents fork deletion

**System Response:**
1. Optimization applied successfully
2. Task marked completed
3. Cleanup phase begins
4. Fork deletion fails for 1 of 3 forks
5. Error logged
6. Retry deletion 3 times with backoff
7. If still fails:
   - Log warning (not error)
   - Schedule background cleanup job
   - Continue (don't fail the task)
8. Task completes

**User Impact:** None - fork will be cleaned up later

**Rationale:** Fork cleanup is housekeeping, not critical to task success

---

## 📋 Summary

This document defines:
- **7 Human/AI/System Actors** with clear roles
- **1 Complete End-to-End Flow** (7 steps, ~5 minutes)
- **23 Business Rules** governing all operations
- **4 Task States** with transition logic
- **6 Agent Workflow Phases** executed in parallel
- **1 Consensus Algorithm** with detailed scoring
- **6 Error Scenarios** with recovery strategies

**Key Principle:** *Safety through isolation, intelligence through collaboration, transparency through metrics.*

---

**Related Documentation:**
- Next: [02-DATA-MODEL.md](02-DATA-MODEL.md) - Database schema for all entities
- See also: [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md) - Deep dive into agent specializations
- See also: [05-CONSENSUS-BENCHMARKING.md](05-CONSENSUS-BENCHMARKING.md) - Scoring algorithms

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```
