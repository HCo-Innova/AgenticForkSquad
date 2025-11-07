# 02-DATA-MODEL.md

```markdown
# 02-DATA-MODEL.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Data Model Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), [01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md)

---

## 📖 Table of Contents

1. [Database Overview](#database-overview)
2. [Existing Tables (E-commerce Data)](#existing-tables-e-commerce-data)
3. [New Tables (AFS System)](#new-tables-afs-system)
4. [Relationships Diagram](#relationships-diagram)
5. [JSONB Structures](#jsonb-structures)
6. [Indexes Strategy](#indexes-strategy)
7. [Migration Strategy](#migration-strategy)
8. [Data Retention Policy](#data-retention-policy)

---

## 🗄️ Database Overview

### Database Roles

**Development Phase:**
- **Engine:** PostgreSQL 16
- **Purpose:** Local development and testing
- **Location:** Docker container
- **Connection:** Direct via pgx driver

**Production Phase:**
- **Engine:** Tiger Cloud Postgres
- **Purpose:** Challenge deployment with fork capabilities
- **Location:** Tiger Cloud infrastructure
- **Connection:** Via MCP protocol

### Schema Organization

**Two Logical Domains:**

1. **E-commerce Domain** (Demo Data)
   - Tables: `users`, `orders`, `payments`
   - Purpose: Provide realistic data for optimization tasks
   - Volume: 1,000 users, 10,000 orders
   - Status: Already seeded and populated

2. **AFS System Domain** (Application Data)
   - Tables: `tasks`, `agent_executions`, `optimization_proposals`, 
     `benchmark_results`, `consensus_decisions`, `query_logs`
   - Purpose: Store optimization workflow data
   - Volume: Grows with usage
   - Status: To be created via migrations

### Naming Conventions

**Tables:**
- Plural nouns (users, tasks, proposals)
- Lowercase with underscores (agent_executions)
- No prefixes (avoid afs_tasks, just tasks)

**Columns:**
- Lowercase with underscores (created_at, fork_id)
- Boolean fields prefixed with is/has (is_valid, has_error)
- Timestamps suffixed with _at (completed_at, started_at)
- Foreign keys suffixed with _id (task_id, agent_execution_id)

**Indexes:**
- Pattern: idx_{table}_{columns}
- Example: idx_tasks_status, idx_orders_user_completed
- Unique indexes: uniq_{table}_{columns}

---

## 📦 Existing Tables (E-commerce Data)

### Table: users

**Purpose:**  
Represents customers in the e-commerce system. Used as demo data for 
optimization tasks involving user-related queries.

**Lifecycle:**
- Created during initial migration
- Populated by seeder script
- Read-only during normal AFS operations
- Can be regenerated for testing

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique user identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email address |
| created_at | TIMESTAMP | DEFAULT NOW() | Account creation timestamp |

**Volume:** 1,000 records

**Sample Data Pattern:**
- Emails generated via gofakeit library
- Realistic email formats
- Created_at distributed over past year

**Usage in AFS:**
- Target of JOIN queries for optimization
- Provides realistic query patterns
- Not modified by AFS system

---

### Table: orders

**Purpose:**  
Represents purchase orders. Primary table for optimization demos,
contains JOINs, aggregations, and filtering opportunities.

**Lifecycle:**
- Created during initial migration
- Populated by seeder script with 10K records
- Read-only during normal operations
- AFS optimizations tested against this table

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique order identifier |
| user_id | INTEGER | FOREIGN KEY (users.id), NOT NULL | Customer reference |
| total | NUMERIC(10,2) | NOT NULL | Order total amount |
| status | VARCHAR(20) | NOT NULL | Order status enum |
| created_at | TIMESTAMP | DEFAULT NOW() | Order creation timestamp |

**Volume:** 10,000 records

**Status Distribution:**
- 75% completed (~7,500 records)
- 12.5% pending (~1,250 records)
- 12.5% processing (~1,250 records)

**Data Characteristics:**
- Total amounts: $10 - $500 (random via gofakeit)
- Created_at: Distributed over time for realistic queries
- User_id: References all 1,000 users with varied order counts

**Optimization Opportunities:**
- Missing index on (user_id, status) - common in demos
- Sequential scans on status filtering
- JOIN inefficiencies with users table
- Aggregation performance on total column

**Usage in AFS:**
- Primary table for "slow query" demonstrations
- Target for index recommendations
- Candidate for partitioning proposals
- Aggregation optimization scenarios

---

### Table: payments

**Purpose:**  
Payment transactions linked to orders. Provides additional JOIN
complexity for multi-table optimization scenarios.

**Lifecycle:**
- Created during initial migration
- One payment per order (1:1 relationship)
- Read-only during operations

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique payment identifier |
| order_id | INTEGER | FOREIGN KEY (orders.id), NOT NULL | Order reference |
| amount | NUMERIC(10,2) | NOT NULL | Payment amount |
| status | VARCHAR(20) | NOT NULL | Payment status |
| created_at | TIMESTAMP | DEFAULT NOW() | Payment timestamp |

**Volume:** 10,000 records (1:1 with orders)

**Data Characteristics:**
- Amount matches order.total
- Status matches order.status
- Created_at same as order.created_at

**Usage in AFS:**
- Multi-table JOIN scenarios
- Complex query optimization (users → orders → payments)
- Demonstrates JOIN ordering optimization

---

## 🆕 New Tables (AFS System)

### Table: tasks

**Purpose:**  
Central entity representing an optimization task submitted by a user.
Tracks the complete lifecycle from submission to completion.

**Lifecycle:**
1. Created when user submits query for optimization
2. Status: pending (waiting for processing)
3. Router assigns agents → status: in_progress
4. Agents work in parallel
5. Consensus reached and applied → status: completed
6. Or error occurs → status: failed
7. Record persists indefinitely for audit trail

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique task identifier |
| type | VARCHAR(50) | NOT NULL | Task type enum |
| description | TEXT | NULL | User-provided description |
| target_query | TEXT | NOT NULL | SQL query to optimize |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | Current task status |
| created_at | TIMESTAMP | DEFAULT NOW() | Task submission time |
| completed_at | TIMESTAMP | NULL | Task completion time |
| metadata | JSONB | NULL | Additional configuration |

**Type Enum Values:**
- `query_optimization` - Optimize slow SELECT query
- `schema_improvement` - Suggest schema changes
- `index_recommendation` - Recommend indexes

**Status Enum Values:**
- `pending` - Queued, not started
- `in_progress` - Agents actively working
- `completed` - Successfully finished
- `failed` - Error occurred

**Metadata JSONB Structure:**
See [JSONB Structures](#jsonb-structures) section below.

**Relationships:**
- Parent of: agent_executions (1:N)
- Parent of: consensus_decisions (1:1)

**Indexes:**
- PRIMARY KEY on id
- INDEX on status (for filtering pending/in_progress tasks)
- INDEX on created_at DESC (for recent tasks queries)
- INDEX on type (for analytics)

**Typical Queries:**
- Get pending tasks: `WHERE status = 'pending' ORDER BY created_at`
- Get user's recent tasks: `WHERE created_at > ? ORDER BY created_at DESC`
- Get failed tasks for debugging: `WHERE status = 'failed'`

**Volume Expectations:**
- Development: ~100 tasks
- Production demo: ~50 tasks
- Growth rate: Low (manual submissions)

---

### Table: agent_executions

**Purpose:**  
Tracks individual agent participation in a task. One record per agent
assigned to a task. Stores fork information and execution status.

**Lifecycle:**
1. Created when router assigns agent to task
2. Status: running (default)
3. Agent creates fork → fork_id populated
4. Agent completes workflow → completed_at set
5. Status updated: completed or failed
6. Record persists for audit and analytics

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique execution identifier |
| task_id | INTEGER | FOREIGN KEY (tasks.id) ON DELETE CASCADE, NOT NULL | Parent task |
| agent_type | VARCHAR(20) | NOT NULL | Agent identifier |
| fork_id | VARCHAR(100) | NULL | Tiger Cloud fork service ID |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'running' | Execution status |
| started_at | TIMESTAMP | DEFAULT NOW() | Agent start time |
| completed_at | TIMESTAMP | NULL | Agent completion time |
| error_message | TEXT | NULL | Error details if failed |

**Agent_type Enum Values:**
- `gemini-2.5-pro` - Planner/QA
- `gemini-2.5-flash` - Generación/ejecución
- `gemini-2.0-flash` - Bajo costo / tareas masivas

**Status Enum Values:**
- `running` - Agent currently executing
- `completed` - Agent finished successfully
- `failed` - Agent encountered error

**Fork_id Format:**
- Pattern: `afs-fork-{agentType}-task{taskID}-{timestamp}`
- Example: `afs-fork-gemini-2.5-pro-task123-1699901234`
- NULL if fork creation failed

**Relationships:**
- Child of: tasks (N:1)
- Parent of: optimization_proposals (1:1 typically)

**Indexes:**
- PRIMARY KEY on id
- INDEX on task_id (for "get all agents for task X")
- INDEX on fork_id (for fork lookup)
- INDEX on agent_type, status (for agent analytics)

**Typical Queries:**
- Get agents for task: `WHERE task_id = ?`
- Get active forks: `WHERE status = 'running' AND fork_id IS NOT NULL`
- Get failed executions: `WHERE status = 'failed'`
- Agent performance: `SELECT agent_type, AVG(completed_at - started_at)`

**Volume Expectations:**
- Average: 2-3 records per task
- Development: ~200-300 records
- Cleanup: Never delete (audit trail)

**Constraint Notes:**
- ON DELETE CASCADE: If task deleted, executions deleted too
- Multiple executions per task allowed (different agents)
- One execution per (task_id, agent_type) combination

---

### Table: optimization_proposals

**Purpose:**  
Stores optimization strategies proposed by agents. Contains SQL
commands, rationale, and estimated impact. Core of the decision-making
process.

**Lifecycle:**
1. Created after agent completes analysis phase
2. Contains generated SQL and estimates
3. Used as input to benchmarking
4. Evaluated by consensus engine
5. Winner gets applied to main database
6. Record persists for comparison and learning

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique proposal identifier |
| agent_execution_id | INTEGER | FOREIGN KEY (agent_executions.id) ON DELETE CASCADE, NOT NULL | Parent execution |
| proposal_type | VARCHAR(50) | NOT NULL | Optimization strategy type |
| sql_commands | TEXT[] | NOT NULL | Array of SQL statements |
| rationale | TEXT | NULL | Agent's explanation |
| estimated_impact | JSONB | NOT NULL | Predicted metrics |
| created_at | TIMESTAMP | DEFAULT NOW() | Proposal creation time |

**Proposal_type Enum Values:**
- `index` - Simple index creation
- `partial_index` - Index with WHERE clause
- `composite_index` - Multi-column index
- `materialized_view` - Pre-computed view
- `partitioning` - Table partitioning
- `denormalization` - Schema restructuring
- `query_rewrite` - SQL reformulation

**SQL_commands Array:**
- PostgreSQL TEXT[] array type
- Each element is one SQL statement
- Order matters (executed sequentially)
- Example: `['CREATE INDEX ...', 'ANALYZE table']`

**Estimated_impact JSONB Structure:**
See [JSONB Structures](#jsonb-structures) section below.

**Relationships:**
- Child of: agent_executions (N:1, but typically 1:1)
- Parent of: benchmark_results (1:N)
- Referenced by: consensus_decisions.winning_proposal_id

**Indexes:**
- PRIMARY KEY on id
- INDEX on agent_execution_id
- INDEX on proposal_type (for analytics)
- GIN INDEX on estimated_impact (for JSONB queries)

**Typical Queries:**
- Get proposals for task: 
  `JOIN agent_executions ON ... WHERE task_id = ?`
- Get winning proposals: 
  `JOIN consensus_decisions ON winning_proposal_id = id`
- Analytics by type: 
  `GROUP BY proposal_type`

**Volume Expectations:**
- 1 proposal per successful agent execution
- ~2-3 proposals per task average
- Development: ~200-300 records

**Validation Rules:**
- sql_commands array cannot be empty
- Each SQL must pass syntax validation
- estimated_impact must contain required keys
- rationale recommended but not required

---

### Table: benchmark_results

**Purpose:**  
Stores actual performance metrics from testing proposals in forks.
Multiple records per proposal (one per test query). Provides objective
data for consensus decision.

**Lifecycle:**
1. Created during agent benchmark phase
2. One record per test query in benchmark suite
3. Execution times averaged from 3 runs
4. EXPLAIN plans stored for analysis
5. Used by consensus engine for scoring
6. Persists for performance history

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique result identifier |
| proposal_id | INTEGER | FOREIGN KEY (optimization_proposals.id) ON DELETE CASCADE, NOT NULL | Parent proposal |
| query_name | VARCHAR(100) | NOT NULL | Test query identifier |
| query_executed | TEXT | NOT NULL | Actual SQL executed |
| execution_time_ms | NUMERIC(10,2) | NOT NULL | Average execution time |
| rows_returned | INTEGER | NULL | Result set size |
| explain_plan | JSONB | NULL | Parsed EXPLAIN output |
| storage_impact_mb | NUMERIC(10,2) | NULL | Storage overhead |
| created_at | TIMESTAMP | DEFAULT NOW() | Benchmark timestamp |

**Query_name Values:**
- `baseline` - Original query before optimization
- `test_limit` - Query with LIMIT 10
- `test_filter` - Query with additional WHERE
- `test_sort` - Query with ORDER BY

**Execution_time_ms:**
- Average of 3 runs (minimum)
- Precision: 2 decimal places (milliseconds)
- Excludes outliers (>2 std deviations)

**Explain_plan JSONB Structure:**
See [JSONB Structures](#jsonb-structures) section below.

**Storage_impact_mb:**
- Calculated as: (size_after - size_before) / 1024 / 1024
- NULL for baseline (no optimization applied yet)
- Can be negative (rare, due to vacuuming)

**Relationships:**
- Child of: optimization_proposals (N:1)

**Indexes:**
- PRIMARY KEY on id
- INDEX on proposal_id
- INDEX on query_name (for filtering specific tests)

**Typical Queries:**
- Get all benchmarks for proposal: `WHERE proposal_id = ?`
- Get baseline results: `WHERE query_name = 'baseline'`
- Calculate improvement: 
  ```
  SELECT 
    (baseline.time - optimized.time) / baseline.time * 100
  FROM benchmark_results baseline
  JOIN benchmark_results optimized ON ...
  ```

**Volume Expectations:**
- 4-5 records per proposal (benchmark suite size)
- ~8-15 records per task (2-3 agents × 4 tests)
- Development: ~1,000 records

**Data Integrity:**
- Every proposal must have at least 1 baseline benchmark
- Execution times must be positive
- Rows_returned should match query result

---

### Table: consensus_decisions

**Purpose:**  
Stores the final decision made by the consensus engine. One record per
task. Contains scoring breakdown, winner selection, and rationale.
Critical for transparency and auditability.

**Lifecycle:**
1. Created after all agents complete benchmarking
2. Consensus engine calculates scores
3. Winner selected based on weighted criteria
4. Rationale generated explaining decision
5. Initially applied_to_main = false
6. After successful application → applied_to_main = true
7. Record persists indefinitely

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique decision identifier |
| task_id | INTEGER | FOREIGN KEY (tasks.id) ON DELETE CASCADE, UNIQUE, NOT NULL | Parent task |
| winning_proposal_id | INTEGER | FOREIGN KEY (optimization_proposals.id), NULL | Selected proposal |
| all_scores | JSONB | NOT NULL | Complete scoring data |
| decision_rationale | TEXT | NULL | Human-readable explanation |
| applied_to_main | BOOLEAN | DEFAULT false | Application status |
| created_at | TIMESTAMP | DEFAULT NOW() | Decision timestamp |

**Task_id Uniqueness:**
- UNIQUE constraint ensures one decision per task
- If consensus fails, winning_proposal_id is NULL

**Winning_proposal_id:**
- NULL if no valid proposals (all failed validation)
- NULL if all improvements below threshold (<10%)
- Foreign key ensures referential integrity

**All_scores JSONB Structure:**
See [JSONB Structures](#jsonb-structures) section below.

**Applied_to_main Boolean:**
- false: Decision made but not yet applied
- true: Optimization successfully applied to main DB
- Separate from task.status for granularity

**Relationships:**
- Child of: tasks (1:1)
- References: optimization_proposals (for winner)

**Indexes:**
- PRIMARY KEY on id
- UNIQUE INDEX on task_id (enforce 1:1)
- INDEX on winning_proposal_id
- INDEX on applied_to_main (for pending applications)
- GIN INDEX on all_scores (for JSONB queries)

**Typical Queries:**
- Get decision for task: `WHERE task_id = ?`
- Get pending applications: `WHERE applied_to_main = false`
- Analytics on winning types:
  ```
  SELECT 
    proposals.proposal_type, 
    COUNT(*)
  FROM consensus_decisions
  JOIN optimization_proposals proposals 
    ON winning_proposal_id = proposals.id
  GROUP BY proposals.proposal_type
  ```

**Volume Expectations:**
- 1 record per task (1:1 relationship)
- Development: ~100 records
- Growth: Matches tasks table

**Constraint Notes:**
- ON DELETE CASCADE on task_id (cleanup with task)
- winning_proposal_id can be NULL (valid state)
- all_scores cannot be NULL (even if empty object)

---

### Table: query_logs (Optional - Advanced Feature)

**Purpose:**  
Logs executed queries for pattern detection and hybrid search. Enables
AI agents to learn from historical query patterns and find similar
optimizations. Demonstrates Tiger Cloud's hybrid search capabilities.

**Lifecycle:**
1. Populated via trigger on orders/users/payments queries
2. Query text captured automatically
3. Embeddings generated asynchronously (background job)
4. Used for semantic similarity search
5. Optional retention policy (keep last 90 days)

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique log identifier |
| query_text | TEXT | NOT NULL | Executed SQL statement |
| execution_time_ms | NUMERIC(10,2) | NULL | Query duration |
| executed_at | TIMESTAMP | DEFAULT NOW() | Execution timestamp |
| query_embedding | vector(1536) | NULL | Semantic vector (Vertex AI) |
| model_name | VARCHAR(100) | NULL | Embedding model identifier |

**Query_embedding:**
- Type: PostgreSQL pgvector extension (vector(1536))
- Generated via Vertex AI Embeddings
- Dimensions depend on chosen model (example uses 1536)
- NULL until background job processes
- Enables semantic similarity search

**Usage Scenarios:**

**Full-Text Search:**
```
Find queries mentioning 'revenue' or 'monthly':
WHERE to_tsvector('english', query_text) @@ 
      to_tsquery('english', 'revenue | monthly')
ORDER BY ts_rank(...)
```

**Vector Similarity Search:**
```
Find semantically similar queries:
ORDER BY query_embedding <=> '[user_query_embedding]'
LIMIT 10
```

**Hybrid Search:**
```
Combine both approaches with weighted scores:
(fts_score × 0.4) + (vector_similarity_score × 0.6)
```

**Relationships:**
- Standalone table (no foreign keys)
- Referenced conceptually by agents during analysis

**Indexes:**
- PRIMARY KEY on id
- GIN INDEX on to_tsvector('english', query_text) (full-text)
- IVFFLAT INDEX on query_embedding (vector similarity)
- INDEX on executed_at (for time-based queries)

**Typical Queries:**
- Recent slow queries: 
  `WHERE execution_time_ms > 1000 ORDER BY executed_at DESC`
- Similar queries: 
  `ORDER BY query_embedding <=> ? LIMIT 10`
- Pattern analysis: 
  `GROUP BY date_trunc('day', executed_at)`

**Volume Expectations:**
- Growth: 100-1000 records per day (if capturing all queries)
- Retention: Keep 90 days → ~9,000-90,000 records
- Cleanup: Scheduled job deletes old records

**Implementation Considerations:**
- pgvector extension must be installed
- Embedding generation is async (Vertex AI API latency)
- Index creation slow on large datasets (use IVFFLAT)
- Storage: ~6KB per embedding (1536 × 4 bytes)

**Optional Nature:**
- Can skip in MVP to save time
- Include if time permits (impressive demo feature)
- Demonstrates Tiger Cloud's advanced capabilities

---

## 🔗 Relationships Diagram

### Entity Relationship Overview

```
tasks (1) ──────────────┬────────────> (N) agent_executions
                        │                         │
                        │                         │
                        │                         ▼
                        │              (1) optimization_proposals
                        │                         │
                        │                         │
                        │                         ▼
                        │              (N) benchmark_results
                        │
                        │
                        └────────────> (1) consensus_decisions
                                                   │
                                                   │
                                        (references winning proposal)
```

### Detailed Relationships

**tasks → agent_executions (One-to-Many)**
- Relationship: Parent-Child
- Cardinality: 1 task has 1-3 agent_executions
- Foreign Key: agent_executions.task_id → tasks.id
- Delete Rule: CASCADE (delete executions if task deleted)
- Business Rule: Minimum 1 agent per task

**agent_executions → optimization_proposals (One-to-One typically)**
- Relationship: Parent-Child
- Cardinality: 1 execution produces 0-1 proposals
- Foreign Key: optimization_proposals.agent_execution_id → 
  agent_executions.id
- Delete Rule: CASCADE
- Business Rule: 0 proposals if agent failed, 1 if succeeded

**optimization_proposals → benchmark_results (One-to-Many)**
- Relationship: Parent-Child
- Cardinality: 1 proposal has 4-5 benchmark results
- Foreign Key: benchmark_results.proposal_id → 
  optimization_proposals.id
- Delete Rule: CASCADE
- Business Rule: Minimum 1 baseline benchmark required

**tasks → consensus_decisions (One-to-One)**
- Relationship: Parent-Child
- Cardinality: 1 task has exactly 1 decision (or 0 if in progress)
- Foreign Key: consensus_decisions.task_id → tasks.id
- Delete Rule: CASCADE
- Uniqueness: UNIQUE constraint on task_id
- Business Rule: Created only after all agents complete

**consensus_decisions → optimization_proposals (Many-to-One reference)**
- Relationship: Reference
- Cardinality: 1 decision references 0-1 winning proposals
- Foreign Key: consensus_decisions.winning_proposal_id → 
  optimization_proposals.id
- Delete Rule: Not specified (proposal shouldn't be deleted if winner)
- Business Rule: NULL if no valid winner

### E-commerce Tables (Independent)

**users → orders (One-to-Many)**
- Foreign Key: orders.user_id → users.id
- No cascade (demo data, not managed by AFS)

**orders → payments (One-to-One)**
- Foreign Key: payments.order_id → orders.id
- Demo data relationship

---

## 📊 JSONB Structures

### tasks.metadata

**Purpose:** Store flexible configuration and context for tasks.

**Schema:**
```
{
  "priority": string,          // "low" | "medium" | "high"
  "complexity_hint": string,   // "low" | "medium" | "high"
  "target_tables": string[],   // ["orders", "users"]
  "user_preferences": {
    "max_storage_mb": number,  // Maximum acceptable overhead
    "risk_tolerance": string   // "low" | "medium" | "high"
  },
  "scoring_weights": {         // Optional custom weights
    "performance": number,     // 0-1, default 0.5
    "storage": number,         // 0-1, default 0.2
    "complexity": number,      // 0-1, default 0.2
    "risk": number             // 0-1, default 0.1
  }
}
```

**Example:**
```
{
  "priority": "high",
  "complexity_hint": "medium",
  "target_tables": ["orders", "users"],
  "user_preferences": {
    "max_storage_mb": 50,
    "risk_tolerance": "low"
  }
}
```

**Usage:**
- Router reads priority for agent assignment
- Agents read target_tables for context
- Consensus reads scoring_weights for custom scoring
- All fields optional (defaults applied)

**Validation:**
- Priority must be valid enum if present
- Scoring weights must sum to 1.0 if present
- Max_storage_mb must be positive if present

---

### optimization_proposals.estimated_impact

**Purpose:** Store agent's predictions about optimization impact.

**Schema:**
```
{
  "query_time_improvement": number,  // Percentage (0-100)
  "storage_overhead_mb": number,     // Megabytes
  "complexity": string,              // "low" | "medium" | "high"
  "risk": string,                    // "low" | "medium" | "high"
  "additional_notes": string         // Optional free text
}
```

**Example:**
```
{
  "query_time_improvement": 80.0,
  "storage_overhead_mb": 12.5,
  "complexity": "low",
  "risk": "low",
  "additional_notes": "Partial index reduces size by covering only 
                       completed orders"
}
```

**Usage:**
- Display in UI before benchmarking
- Compare with actual benchmark results (calibration)
- Used by consensus if benchmarks fail

**Validation:**
- query_time_improvement: 0-100 range
- storage_overhead_mb: non-negative
- complexity/risk: valid enum values

---

### benchmark_results.explain_plan

**Purpose:** Store parsed PostgreSQL EXPLAIN ANALYZE output.

**Schema (Simplified):**
```
{
  "planning_time_ms": number,
  "execution_time_ms": number,
  "total_cost": number,
  "actual_rows": number,
  "plan_type": string,           // "Seq Scan" | "Index Scan" | etc
  "index_name": string,           // If index used
  "filter_removed_rows": number,  // Rows filtered out
  "sort_method": string,          // If sorting
  "buffers": {
    "shared_hit": number,
    "shared_read": number
  },
  "full_plan": object             // Complete EXPLAIN JSON
}
```

**Example:**
```
{
  "planning_time_ms": 0.5,
  "execution_time_ms": 450.23,
  "total_cost": 1234.56,
  "actual_rows": 1000,
  "plan_type": "Index Scan",
  "index_name": "idx_orders_user_completed",
  "filter_removed_rows": 0,
  "buffers": {
    "shared_hit": 128,
    "shared_read": 4
  }
}
```

**Usage:**
- Display in UI for technical users
- Parse plan_type to verify index usage
- Compare planning vs execution time
- Detect inefficiencies (high filter_removed_rows)

**Source:**
- Generated by PostgreSQL EXPLAIN (FORMAT JSON)
- Parsed and simplified for storage
- Full plan preserved in full_plan field

---

### consensus_decisions.all_scores

**Purpose:** Store complete scoring breakdown for all proposals.

**Schema:**
```
{
  "[agent_type]": {
    "proposal_id": number,
    "performance_score": number,    // 0-100
    "storage_score": number,        // 0-100
    "complexity_score": number,     // 0-100
    "risk_score": number,           // 0-100
    "weighted_total": number,       // 0-100
    "rank": number,                 // 1, 2, 3...
    "improvement_pct": number       // Actual benchmark improvement
  }
}
```

**Example:**
```
{
  "gemini-2.5-pro": {
    "proposal_id": 45,
    "performance_score": 90.86,
    "storage_score": 98.67,
    "complexity_score": 90.0,
    "risk_score": 95.0,
    "weighted_total": 92.66,
    "rank": 1,
    "improvement_pct": 82.6
  },
  "gemini25pro": {
    "proposal_id": 46,
    "performance_score": 100.0,
    "storage_score": 60.0,
    "complexity_score": 70.0,
    "risk_score": 70.0,
    "weighted_total": 81.00,
    "rank": 2,
    "improvement_pct": 93.5
  },
  "gemini-2.0-flash": {
    "proposal_id": 47,
    "performance_score": 67.65,
    "storage_score": 75.0,
    "complexity_score": 50.0,
    "risk_score": 70.0,
    "weighted_total": 64.83,
    "rank": 3,
    "improvement_pct": 61.5
  }
}
```

**Usage:**
- Display score comparison in UI
- Understand why winner was selected
- Analytics on scoring patterns
- Tune scoring algorithm

**Validation:**
- All score fields 0-100 range
- Ranks must be sequential (1, 2, 3...)
- weighted_total must match calculation
- proposal_id must exist in proposals table

---

## 🔍 Indexes Strategy

### Performance-Critical Indexes

**tasks Table:**
1. `idx_tasks_status` on (status)
   - Query: Filter pending/in_progress tasks
   - Cardinality: Low (4 values)
   - Type: B-tree
   
2. `idx_tasks_created_at` on (created_at DESC)
   - Query: Recent tasks listing
   - Cardinality: High (timestamp)
   - Type: B-tree

3. `idx_tasks_type` on (type)
   - Query: Analytics by task type
   - Cardinality: Low (3 values)
   - Type: B-tree

**agent_executions Table:**
1. `idx_agent_executions_task_id` on (task_id)
   - Query: Get agents for specific task
   - Cardinality: Medium (N agents per task)
   - Type: B-tree

2. `idx_agent_executions_fork_id` on (fork_id)
   - Query: Lookup by fork for cleanup
   - Cardinality: High (unique per execution)
   - Type: B-tree

3. `idx_agent_executions_status` on (agent_type, status)
   - Query: Agent performance analytics
   - Cardinality: Low×Low composite
   - Type: B-tree

**optimization_proposals Table:**
1. `idx_proposals_agent_execution_id` on (agent_execution_id)
   - Query: Get proposal for execution
   - Cardinality: High (1:1 typically)
   - Type: B-tree

2. `idx_proposals_type` on (proposal_type)
   - Query: Analytics by optimization type
   - Cardinality: Medium (7 types)
   - Type: B-tree

3. `idx_proposals_estimated_impact` on (estimated_impact)
   - Query: JSONB queries on impact data
   - Cardinality: N/A (JSONB)
   - Type: GIN

**benchmark_results Table:**
1. `idx_benchmark_proposal_id` on (proposal_id)
   - Query: Get all benchmarks for proposal
   - Cardinality: Medium (4-5 per proposal)
   - Type: B-tree

2. `idx_benchmark_query_name` on (query_name)
   - Query: Filter specific test types
   - Cardinality: Low (4 values)
   - Type: B-tree

**consensus_decisions Table:**
1. `uniq_consensus_task_id` on (task_id) UNIQUE
   - Constraint: One decision per task
   - Cardinality: High (1:1 with tasks)
   - Type: B-tree

2. `idx_consensus_winning_proposal` on (winning_proposal_id)
   - Query: Join to winning proposal
   - Cardinality: High
   - Type: B-tree

3. `idx_consensus_applied` on (applied_to_main)
   - Query: Find pending applications
   - Cardinality: Very Low (boolean)
   - Type: B-tree

4. `idx_consensus_all_scores` on (all_scores)
   - Query: JSONB analytics
   - Cardinality: N/A
   - Type: GIN

**query_logs Table (Optional):**
1. `idx_query_logs_text` on (to_tsvector('english', query_text))
   - Query: Full-text search
   - Cardinality: High
   - Type: GIN

2. `idx_query_logs_embedding` on (query_embedding) 
   USING ivfflat
   - Query: Vector similarity
   - Cardinality: High
   - Type: IVFFLAT (pgvector)

3. `idx_query_logs_executed_at` on (executed_at DESC)
   - Query: Recent logs
   - Cardinality: High
   - Type: B-tree

### Index Maintenance

**Considerations:**
- All indexes created after initial data load
- ANALYZE run after index creation
- Monitor index bloat in production
- Rebuild if fragmentation >30%

**Statistics:**
- Auto-vacuum enabled
- Stats updated after bulk operations
- Analyze scheduled daily

---

## 🔄 Migration Strategy

### Migration Files Organization

**Naming Convention:**
```
YYYYMMDDHHMMSS_descriptive_name.sql
```

**Examples:**
- `20240101000001_create_ecommerce_tables.sql`
- `20240101000002_create_afs_core_tables.sql`
- `20240101000003_create_indexes.sql`
- `20240101000004_add_query_logs_table.sql` (optional)

### Migration Sequence

**Migration 001: E-commerce Tables**
- Purpose: Create demo data tables
- Tables: users, orders, payments
- Foreign keys: orders.user_id, payments.order_id
- Status: Already executed (existing schema)

**Migration 002: AFS Core Tables**
- Purpose: Create AFS system tables
- Tables: tasks, agent_executions, optimization_proposals,
  benchmark_results, consensus_decisions
- Foreign keys: All relationships defined
- Order: Parent tables before child tables

**Migration 003: Core Indexes**
- Purpose: Performance indexes for common queries
- Indexes: All B-tree indexes listed above
- ANALYZE: Run after index creation

**Migration 004: JSONB Indexes**
- Purpose: GIN indexes for JSONB columns
- Indexes: tasks.metadata, proposals.estimated_impact,
  consensus.all_scores
- Note: GIN indexes slower to create

**Migration 005: Query Logs (Optional)**
- Purpose: Enable hybrid search feature
- Prerequisites: pgvector extension
- Tables: query_logs
- Indexes: GIN for text, IVFFLAT for vectors
- Skip if: Time constrained in challenge

**Migration 006: Extensions (Tiger Cloud)**
- Purpose: Enable required PostgreSQL extensions
- Extensions: pgvector (if query_logs included)
- Note: May already be installed in Tiger Cloud

### Migration Tools

**Development:**
- Tool: golang-migrate or custom migration system
- Direction: Up (apply) and Down (rollback) supported
- Tracking: migrations table stores applied versions

**Production (Tiger Cloud):**
- Manual execution via Tiger CLI or MCP
- Verification queries after each migration
- Backup PITR timestamp before migrations

### Rollback Strategy

**Safe Rollback:**
- Each migration has corresponding down migration
- Rollback tested in development before production
- Data loss acceptable during development

**Production Considerations:**
- Use Tiger Cloud PITR for rollback if needed
- Fork main DB before migrations (safety)
- Test migrations on fork first

---

## 🗑️ Data Retention Policy

### Retention Rules

**tasks Table:**
- Retention: Indefinite (audit trail)
- Cleanup: Manual only (administrative)
- Rationale: Historical optimization patterns valuable

**agent_executions Table:**
- Retention: Indefinite (tied to tasks)
- Cleanup: Cascade delete with parent task
- Rationale: Execution history for debugging

**optimization_proposals Table:**
- Retention: Indefinite
- Cleanup: Cascade delete with execution
- Rationale: Learning from past proposals

**benchmark_results Table:**
- Retention: Indefinite
- Cleanup: Cascade delete with proposal
- Rationale: Performance baselines valuable

**consensus_decisions Table:**
- Retention: Indefinite
- Cleanup: Cascade delete with task
- Rationale: Decision audit trail

**query_logs Table (Optional):**
- Retention: 90 days (configurable)
- Cleanup: Scheduled job deletes old records
- Rationale: Recent patterns most relevant
- Query: `DELETE FROM query_logs 
          WHERE executed_at < NOW() - INTERVAL '90 days'`

### Cleanup Jobs

**Scheduled Tasks:**
1. **Orphaned Forks Cleanup** (Daily)
   - Find agent_executions with fork_id but status != running
   - Fork older than 24 hours
   - Call Tiger MCP to delete fork
   - Update agent_execution (clear fork_id)

2. **Query Logs Cleanup** (Weekly, if applicable)
   - Delete records older than retention period
   - Vacuum table after deletion
   - Rebuild vector index if fragmented

### Archival Strategy

**Not Implemented (Future Enhancement):**
- Move old tasks to archive schema
- Compress proposal/benchmark data
- Export to cold storage (S3, etc.)

**Development Environment:**
- Truncate all tables between test runs
- Reseed demo data as needed
- No retention concerns

---

## 📐 Data Integrity Constraints

### Foreign Key Constraints

**Cascade Deletes:**
- agent_executions → tasks (CASCADE)
- optimization_proposals → agent_executions (CASCADE)
- benchmark_results → optimization_proposals (CASCADE)
- consensus_decisions → tasks (CASCADE)

**Rationale:** Deleting a task should clean up all related data.

**Non-Cascade:**
- consensus_decisions.winning_proposal_id (no ON DELETE)
- Winning proposals should not be deleted

### Uniqueness Constraints

**tasks:**
- No unique constraints beyond PRIMARY KEY
- Duplicate queries allowed (different optimizations over time)

**agent_executions:**
- No explicit unique constraint
- Logical uniqueness: (task_id, agent_type) but not enforced
- Allows retries with same agent

**consensus_decisions:**
- UNIQUE on task_id (one decision per task)
- Critical for data integrity

### Check Constraints

**Recommended (Not Mandatory):**
- tasks.status IN ('pending', 'in_progress', 'completed', 'failed')
- agent_executions.status IN ('running', 'completed', 'failed')
- benchmark_results.execution_time_ms > 0
- consensus_decisions.applied_to_main IN (true, false)

**Implementation:**
- Can use CHECK constraints or application-level validation
- Application-level preferred for flexibility

### NULL Constraints

**Critical NOT NULL Fields:**
- All id fields (PRIMARY KEY implies NOT NULL)
- tasks.target_query (required)
- tasks.type (required)
- agent_executions.task_id (required)
- optimization_proposals.sql_commands (required)
- benchmark_results.execution_time_ms (required)

**Nullable Fields (By Design):**
- tasks.completed_at (NULL until completed)
- agent_executions.fork_id (NULL if fork creation failed)
- consensus_decisions.winning_proposal_id (NULL if no winner)
- All description/rationale text fields (optional)

---

## 🎯 Summary

This data model provides:

**6 Core Tables:**
1. tasks - Central optimization requests
2. agent_executions - Agent participation tracking
3. optimization_proposals - Proposed solutions
4. benchmark_results - Performance measurements
5. consensus_decisions - Final decisions
6. query_logs - Historical patterns (optional)

**3 Demo Tables:**
- users, orders, payments (existing)

**Key Features:**
- Clear parent-child relationships with cascade deletes
- JSONB flexibility for extensibility
- Comprehensive indexing strategy
- Migration path from development to production
- Audit trail preservation

**Design Principles:**
- Normalization: 3NF (Third Normal Form)
- Flexibility: JSONB for evolving schemas
- Performance: Indexes on all foreign keys and query patterns
- Integrity: Foreign keys and constraints
- Auditability: Timestamps and status tracking

---

**Related Documentation:**
- Previous: [01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md) - How data flows
- Next: [03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md) - Code structure
- See also: [06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md) - Fork management

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```
