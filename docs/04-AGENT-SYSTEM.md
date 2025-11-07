# 04-AGENT-SYSTEM.md

```markdown
# 04-AGENT-SYSTEM.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Agent System Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md), 
[03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md)

---

## 📖 Table of Contents

1. [Agent System Overview](#agent-system-overview)
2. [Agent Interface Contract](#agent-interface-contract)
3. [Agent Specializations](#agent-specializations)
4. [Task Router](#task-router)
5. [Parallel Coordination](#parallel-coordination)
6. [Prompt Engineering](#prompt-engineering)
7. [Response Parsing](#response-parsing)
8. [Error Handling](#error-handling)

---

## 🤖 Agent System Overview

### Core Concept

**Multi-Agent Collaboration:**  
Instead of a single AI making optimization decisions, AFS employs 
multiple specialized AI agents working in parallel. Each agent brings 
unique expertise and proposes different solutions. A consensus mechanism 
selects the optimal approach.

**Why Multiple Agents?**

**Diverse Perspectives:**
- gemini-2.5-pro planifica, desambigua y valida SQL/código
- gemini-2.5-flash genera SQL/código y ejecuta transformaciones/pruebas
- gemini-2.0-flash resuelve tareas masivas/repetitivas de bajo riesgo
- Different approaches = better coverage of solution space

**Risk Mitigation:**
- If one agent fails, others continue
- No single point of failure
- Partial success acceptable (2 of 3 agents)

**Quality Through Competition:**
- Agents "compete" via objective benchmarks
- Best proposal wins based on metrics
- Eliminates bias toward specific approach

**Validation:**
- Multiple proposals = cross-validation
- If all agents suggest similar solution, high confidence
- If proposals differ significantly, indicates complex problem

### Agent Lifecycle

**Per-Task Lifecycle:**

```
1. Router assigns agent to task
2. Agent creates isolated fork
3. Agent analyzes query in fork
4. Agent proposes optimization
5. Agent benchmarks proposal in fork
6. Agent returns results
7. Fork cleaned up after consensus
```

**Stateless Design:**
- Agents don't maintain state between tasks
- Each task execution independent
- Fork provides isolation

---

## 🔌 Agent Interface Contract

### Interface Definition (Conceptual)

**All agents must implement these methods:**

**GetType() AgentType**

**Purpose:** Identify which agent this is

**Returns:**
- AgentType enum: gemini25pro, gemini25flash, or gemini20flash

**Usage:**
- Logging and tracking
- Display in UI
- Routing decisions

---

**GetSpecialization() AgentSpecialization**

**Purpose:** Describe agent's strengths and focus areas

**Returns:**
- AgentSpecialization struct containing:
  - Type: AgentType
  - Strengths: Array of capability strings
  - Description: Human-readable summary
  - PreferredTaskTypes: Which task types agent excels at

**Example Returns:**

```
gemini-2.5-pro:
  Type: gemini25pro
  Strengths: ["Advanced reasoning", "Index design",
              "Query plan analysis", "Trade-off analysis"]
  Description: "Analiza y valida con alta precisión"
  PreferredTaskTypes: ["query_optimization", "schema_improvement"]

gemini-2.5-flash:
  Type: gemini25flash
  Strengths: ["SQL generation", "Partitioning",
              "Schema changes", "Benchmark execution"]
  Description: "Entrega rápida y costo-eficiente"
  PreferredTaskTypes: ["query_optimization", "schema_improvement"]

gemini-2.0-flash:
  Type: gemini20flash
  Strengths: ["Batch tasks", "Boilerplate",
              "Materialized views", "Low-risk changes"]
  Description: "Optimiza tareas masivas a bajo costo"
  PreferredTaskTypes: ["query_optimization"]
```

**Usage:**
- Router uses to match agents to tasks
- UI displays capabilities
- Documentation generation

---

**CreateFork(ctx Context, parentServiceID string) (forkID string, error)**

**Purpose:** Create isolated database fork for experimentation

**Parameters:**
- ctx: Context for cancellation and timeout
- parentServiceID: Main database service ID (e.g., "afs-main")

**Returns:**
- forkID: Unique identifier for created fork
- error: If fork creation failed

**Process:**
1. Generate unique fork name (includes agent type, task ID, timestamp)
2. Call MCP client ServiceFork operation
3. Wait for fork creation (timeout: 30 seconds)
4. Verify fork accessible
5. Return fork service ID

**Naming Convention:**
- Pattern: afs-fork-{agentType}-task{taskID}-{unixTimestamp}
- Example: afs-fork-gemini25pro-task123-1699901234

**Error Cases:**
- MCP server unavailable
- Parent service doesn't exist
- Fork quota exceeded
- Timeout

---

**DestroyFork(ctx Context, forkID string) error**

**Purpose:** Clean up fork after use

**Parameters:**
- ctx: Context for timeout
- forkID: Fork to delete

**Returns:**
- error: If deletion failed (or nil)

**Process:**
1. Call MCP client ServiceDelete operation
2. Wait for confirmation
3. Verify fork no longer exists

**Error Cases:**
- Fork already deleted (treat as success - idempotent)
- MCP server unavailable
- Timeout

**Cleanup Responsibility:**
- Called even if agent fails
- Must not fail entire task
- Log errors but don't propagate

---

**AnalyzeTask(ctx Context, task Task, forkID string) (AnalysisResult, error)**

**Purpose:** Analyze query and identify optimization opportunities

**Parameters:**
- ctx: Context for timeout (2 minutes)
- task: Domain entity with target query and metadata
- forkID: Fork to execute analysis in

**Returns:**
- AnalysisResult: Insights and proposed actions
- error: If analysis failed

**AnalysisResult Structure:**

```
AnalysisResult:
  Insights: Array of strings (what agent observed)
    Example: ["Sequential scan on orders table", 
              "No index on user_id column"]
  
  ProposedActions: Array of strings (high-level recommendations)
    Example: ["Create index on orders(user_id, status)",
              "Consider partial index for completed orders"]
  
  Confidence: Float (0.0 - 1.0) (how certain agent is)
    Example: 0.87
  
  Metadata: Map (additional context)
    Example: {"query_complexity": "medium", 
              "estimated_improvement": "70-80%"}
```

**Process:**
1. Execute EXPLAIN ANALYZE on target query in fork
2. Parse explain plan (identify scan types, costs, etc.)
3. Query database statistics (table sizes, existing indexes)
4. Build context-rich prompt for LLM
5. Call LLM API with prompt
6. Parse LLM JSON response
7. Validate response completeness
8. Return AnalysisResult

**Error Cases:**
- Fork not accessible
- Query execution fails
- LLM API timeout/failure
- Response parsing fails

---

**ProposeOptimization(ctx Context, analysis AnalysisResult, forkID string) 
(OptimizationProposal, error)**

**Purpose:** Generate specific SQL optimization based on analysis

**Parameters:**
- ctx: Context for timeout (1 minute)
- analysis: Previous analysis result
- forkID: Fork for validation

**Returns:**
- OptimizationProposal: Domain entity with SQL and rationale
- error: If proposal generation failed

**OptimizationProposal Structure:**

```
OptimizationProposal:
  ID: Generated after persistence
  AgentExecutionID: Parent execution
  ProposalType: Enum (index, partitioning, etc.)
  SQLCommands: Array of SQL statements
  Rationale: Explanation text
  EstimatedImpact: JSONB with predictions
```

**Process:**
1. Based on analysis insights, determine optimization type
2. Build prompt asking LLM to generate specific SQL
3. Call LLM API
4. Parse SQL from response
5. Validate SQL syntax (basic checks)
6. Estimate impact metrics (query time, storage, etc.)
7. Generate rationale explaining decision
8. Create OptimizationProposal entity
9. Return proposal

**Validation:**
- SQL must be valid syntax
- SQL must be safe (no DROP DATABASE, etc.)
- Commands must be reversible (for rollback)
- Estimated impact must be reasonable

**Error Cases:**
- LLM generates invalid SQL
- LLM returns non-SQL response
- Unsafe SQL detected
- LLM API failure

---

**RunBenchmark(ctx Context, proposal OptimizationProposal, forkID string) 
(BenchmarkResults, error)**

**Purpose:** Execute performance tests on proposal

**Parameters:**
- ctx: Context for timeout (5 minutes)
- proposal: Optimization to test
- forkID: Fork to run tests in

**Returns:**
- BenchmarkResults: Array of result entities (one per test query)
- error: If benchmarking failed

**Process:**
1. Define benchmark suite (4 test queries)
2. Execute baseline measurements (before optimization)
3. Apply proposal SQL commands in fork
4. Verify optimization applied successfully
5. Execute each test query 3 times
6. Calculate average execution times
7. Capture EXPLAIN plans
8. Measure storage impact
9. Create BenchmarkResult entities
10. Return array of results

**Benchmark Suite (Standard):**

**Test 1: Baseline**
- Original query exactly as submitted
- Establishes performance before optimization
- Critical for improvement calculation

**Test 2: Limited Result Set**
- Original query with LIMIT 10 added
- Tests overhead of query processing vs data transfer
- Identifies if bottleneck is computation or I/O

**Test 3: Additional Filter**
- Original query with date range filter added
- Tests scalability with different selectivity
- Validates index works with varied conditions

**Test 4: Sorting**
- Original query with ORDER BY added
- Tests if optimization helps with sorting
- Identifies if index enables sort optimization

**Metrics Collected Per Test:**
- Execution time (milliseconds, 2 decimal precision)
- Rows returned
- EXPLAIN plan (parsed JSON)
- Planning time vs execution time

**Storage Measurement:**
- Table/index size before optimization
- Table/index size after optimization
- Overhead in megabytes

**Error Cases:**
- SQL application fails
- Test query fails
- Fork becomes inaccessible
- Timeout exceeded

---

**ValidateProposal(ctx Context, proposal OptimizationProposal) 
(ValidationResult, error)**

**Purpose:** Safety check before benchmarking (optional but recommended)

**Parameters:**
- ctx: Context
- proposal: Proposal to validate

**Returns:**
- ValidationResult: Pass/fail with warnings
- error: If validation process failed

**ValidationResult Structure:**

```
ValidationResult:
  IsValid: Boolean (pass/fail)
  Warnings: Array of warning messages
  RiskLevel: Enum (low, medium, high)
  ReversibilityScore: Float (0.0-1.0, how easy to rollback)
```

**Checks:**
- SQL syntax validity
- No destructive operations (DROP, TRUNCATE)
- No data modification (INSERT, UPDATE, DELETE)
- Only DDL allowed (CREATE, ALTER for indexes/views)
- Estimate reversibility (can we undo easily?)

**Risk Assessment:**
- Low: Simple index creation
- Medium: Materialized view, partitioning
- High: Schema restructuring, denormalization

---

### BaseAgent (Shared Implementation)

**Purpose:**  
Common functionality shared across all agent implementations to avoid 
code duplication.

**Implemented Methods (Shared):**

**CreateFork and DestroyFork:**
- Logic identical across agents
- Only fork naming differs (agent type in name)
- BaseAgent provides default implementation

**Logging:**
- Structured logging with agent type
- Performance metrics tracking
- Error context enrichment

**Error Handling:**
- Retry logic for transient failures
- Timeout management
- Error wrapping with context

**Prompt Utilities:**
- Template rendering
- Context building
- Token counting (for LLM limits)

**Response Parsing:**
- JSON extraction from LLM responses
- Schema validation
- Error recovery

**Agent-Specific Methods (Must Implement):**

**AnalyzeTask:**
- Each agent has specialized prompts
- Different emphasis (gemini-2.5-pro: análisis/validación, gemini-2.5-flash: índices/particiones, gemini-2.0-flash: tareas masivas)

**ProposeOptimization:**
- Different proposal strategies per agent
- Specialized SQL generation

**RunBenchmark:**
- Can use shared logic but may customize suite

---

## 🎓 Agent Specializations

### gemini-2.5-pro

**Identity:**
- LLM (via Vertex AI): gemini-2.5-pro
- Focus: Planificación/razonamiento crítico y QA
- Strength: Gran ventana de contexto y verificación exhaustiva

**Specialization Details:**

**Primary Expertise:**
- Desambiguación de requerimientos y división de tareas
- Revisión de EXPLAIN/SQL y consistencia antes de ejecutar
- Generación de planes y checklist de validación

**Secondary Expertise:**
- Query plan analysis
- Cost-based optimization
- Statistics-driven decisions

**Typical Proposals:**

**Indexes (Most Common):**
- B-tree indexes on filtered columns
- Composite indexes matching query patterns
- Partial indexes with WHERE conditions
- Covering indexes for index-only scans

**Query Rewrites:**
- Replacing correlated subqueries with JOINs
- Eliminating unnecessary DISTINCT
- Optimizing EXISTS vs IN
- CTE optimization

**When to Assign:**
- Tareas críticas/alta complejidad o alto impacto
- Revisión previa a aplicar cambios en DB
- Verificación de propuestas de otros agentes

**Prompting Strategy:**
- Emphasize step-by-step reasoning
- Request detailed EXPLAIN plan analysis
- Ask for multiple index options with trade-offs

**Expected Behavior:**
- Proposes low-risk, high-reward optimizations
- Favors simplicity over cleverness
- Excellent at explaining decisions
- Conservative storage overhead estimates

**Performance Characteristics:**
- Proposals typically: 60-85% improvement
- Storage overhead: 10-30 MB typical
- Complexity: Low (easy to implement/revert)
- Risk: Low (well-understood techniques)

---

### gemini-2.5-flash

**Identity:**
- LLM (via Vertex AI): gemini-2.5-flash
- Focus: Generación de SQL/código con baja latencia
- Strength: Alto throughput/costo eficiente

**Specialization Details:**

**Primary Expertise:**
- Generación de índices, reescrituras de consultas y scripts
- Aplicación de cambios seguros en forks
- Ejecución de benchmarks y captura de EXPLAIN

**Secondary Expertise:**
- Denormalization strategies
- Archive table patterns
- Vertical/horizontal partitioning

**Typical Proposals:**

**Partitioning (Most Common):**
- Partition by date range (time-series data)
- Partition by status (active vs archived)
- Partition by geographic region
- Hash partitioning for even distribution

**Schema Modifications:**
- Denormalize frequently joined tables
- Split large tables vertically
- Create summary/rollup tables
- Archival strategies

**When to Assign:**
- La mayoría de las tareas operativas y concurrentes
- Preparación de propuestas y pruebas de rendimiento

**Prompting Strategy:**
- Ask for data distribution analysis
- Request long-term scalability consideration
- Emphasize maintenance implications

**Expected Behavior:**
- Proposes structural changes
- Considers growth patterns
- Thinks about future data volume
- More aggressive structural changes when appropriate

**Performance Characteristics:**
- Proposals typically: 40-70% improvement
- Storage overhead: 20-60 MB typical
- Complexity: Medium to High (requires migration)
- Risk: Medium (harder to revert)

**Trade-offs:**
- Higher implementation complexity
- Better long-term scalability
- Moderate short-term gains
- Significant maintenance considerations

---

### gemini-2.0-flash

**Identity:**
- LLM (via Vertex AI): gemini-2.0-flash
- Focus: Tareas masivas/repetitivas de bajo riesgo
- Strength: Máximo ahorro de costo/latencia

**Specialization Details:**

**Primary Expertise:**
- Boilerplate, refactors simples, formatters
- Transformaciones mecánicas y generación de plantillas

**Secondary Expertise:**
- Caching strategies
- Read/write optimization trade-offs
- Staleness tolerance analysis

**Typical Proposals:**

**Materialized Views (Most Common):**
- Pre-compute expensive aggregations
- Refresh strategies (concurrent, incremental)
- Indexed materialized views
- Partial materialized views

**Advanced Techniques:**
- Combination of index + query rewrite
- Trigger-maintained denormalization
- Summary tables with delta updates
- Scheduled pre-computation jobs

**When to Assign:**
- Ahorro máximo en tareas de bajo impacto
- Procesamiento en lote

**Prompting Strategy:**
- Ask for creative solutions
- Request trade-off analysis
- Emphasize maintenance cost consideration

**Expected Behavior:**
- Proposes sophisticated solutions
- Excellent trade-off analysis
- Considers operational complexity
- May over-engineer simple problems

**Performance Characteristics:**
- Proposals typically: 80-95% improvement
- Storage overhead: 40-100 MB typical
- Complexity: Medium to High
- Risk: Medium (staleness, refresh overhead)

**Trade-offs:**
- Highest performance gains
- Most storage overhead
- Requires ongoing maintenance
- Staleness concerns for materialized views

---

### Agent Comparison Matrix

**Quick Reference:**

| Aspect | gemini-2.5-pro | gemini-2.5-flash | gemini-2.0-flash |
|--------|---------------------|-------------------------|-------------------|
| **Best For** | Planificación/QA crítica | Generación/ejecución operativa | Tareas masivas/bajo riesgo |
| **Complexity** | Media | Baja | Muy baja |
| **Risk** | Bajo | Bajo-Medio | Bajo |
| **Improvement** | Alta precisión en decisiones | Alta velocidad de entrega | Ahorro de costo/latencia |
| **Storage** | N/A | Variable según propuesta | N/A |
| **Time to Implement** | Moderado | Rápido | Muy rápido |
| **Maintenance** | Baja | Baja | Muy baja |
| **Reversibility** | Fácil | Fácil | Fácil |

**Assignment Priority (By Task Type):**

**query_optimization with JOIN:**
- Primary: gemini-2.5-flash
- Secondary: gemini-2.5-pro

**schema_improvement:**
- Primary: gemini-2.5-flash
- Secondary: gemini-2.5-pro

**High priority:**
- Assign all three models según disponibilidad

**Low priority:**
- gemini-2.0-flash o gemini-2.5-flash

---

## 🎯 Task Router

### Router Responsibility

**Core Function:**  
Analyze incoming optimization tasks and intelligently assign the most 
suitable agents based on task characteristics and business rules.

**Goals:**
- Maximize success probability
- Optimize resource usage (don't assign unnecessary agents)
- Balance agent workload
- Respect priority levels

---

### Routing Algorithm

**Input:**
- Task entity (with type, query, metadata)
- Available agents (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
- Current agent workloads

**Output:**
- Array of assigned agents (1-3 agents)
- Assignment rationale (for logging/audit)

**Process Steps:**

**Step 1: Initialize**
- candidates = empty array
- complexity_score = 0

**Step 2: Parse Query**
- Extract query features:
  - Number of JOINs
  - Number of subqueries
  - Aggregation functions present
  - Window functions present
  - Set operations (UNION, etc.)
  - ORDER BY clauses

**Step 3: Calculate Complexity Score**
```
complexity_score = 0
complexity_score += count(JOINs) × 1
complexity_score += count(subqueries) × 1
complexity_score += count(aggregations) × 1
complexity_score += count(window_functions) × 1
complexity_score += count(set_operations) × 2

Result:
  0-2: Low complexity
  3-5: Medium complexity
  6+: High complexity
```

**Step 4: Query Schema Metadata**
- Get target table row counts
- Identify large tables (>1M rows)
- Check existing indexes

**Step 5: Apply Routing Rules (Priority Order)**

**Rule 1: Priority Override**
```
IF task.metadata.priority == "high":
  candidates = [Gemini25Pro, Gemini25Flash, Gemini20Flash]
  rationale = "High priority - assign all Gemini agents"
  RETURN
```

**Rule 2: Table Size Check**
```
IF any_table_size > 1M rows:
  candidates.add(Gemini25Flash)
  rationale += "Large table - include gemini-2.5-flash for partitioning"
```

**Rule 3: Query Features**
```
IF query contains JOIN:
  candidates.add(Gemini25Flash)
  rationale += "JOIN detected - include gemini-2.5-flash for optimization"

IF query contains aggregation:
  IF complexity_score >= 5:
    candidates.add(Gemini25Pro)
    rationale += "Complex aggregation - include gemini-2.5-pro for mat views"
  ELSE:
    candidates.add(Gemini25Flash)
    rationale += "Simple aggregation - gemini-2.5-flash sufficient"
```

**Rule 4: Complexity-Based**
```
IF complexity_score <= 2:
  candidates.ensure_at_least([Gemini25Flash])
  rationale += "Low complexity - gemini-2.5-flash only"

ELSE IF complexity_score 3-5:
  candidates.ensure_at_least([Gemini25Flash, Gemini25Pro])
  rationale += "Medium complexity - gemini-2.5-flash + gemini-2.5-pro"

ELSE IF complexity_score >= 6:
  candidates = [Gemini25Pro, Gemini25Flash, Gemini20Flash]
  rationale += "High complexity - all Gemini agents"
```

**Rule 5: Task Type**
```
IF task.type == "schema_improvement":
  candidates.ensure_has(Gemini25Flash)
  rationale += "Schema task - gemini-2.5-flash required"

IF task.type == "index_recommendation":
  candidates.ensure_has(Gemini25Pro)
  rationale += "Index task - gemini-2.5-pro required"
```

**Step 6: Check Agent Availability**
```
FOR each candidate agent:
  IF agent.current_task_count >= 3:
    candidates.remove(agent)
    rationale += "Agent at capacity - skipped"
```

**Step 7: Ensure Minimum**
```
IF candidates.empty():
  candidates = [Gemini25Flash]  // Default fallback
  rationale += "No agents available - default to gemini-2.5-flash"
```

**Step 8: Return**
```
RETURN {
  agents: candidates,
  rationale: rationale
}
```

---

### Routing Configuration

**Rule Definitions (Externalized):**

Rules stored in configuration file (YAML/JSON) for easy modification 
without code changes.

**Example Configuration Structure:**

```
routing_rules:
  - name: high_priority_all_agents
    priority: 1
    condition:
      field: metadata.priority
      operator: equals
      value: high
    action:
      assign: [gemini25pro, gemini25flash, gemini20flash]
      rationale: "High priority requires all Gemini agents"

  - name: large_table_partitioning
    priority: 2
    condition:
      field: stats.max_table_rows
      operator: greater_than
      value: 1000000
    action:
      include: [gemini25flash]
      rationale: "Large table - include gemini-2.5-flash"

  - name: join_optimization
    priority: 3
    condition:
      field: features.join_count
      operator: greater_than
      value: 0
    action:
      include: [gemini25flash]
      rationale: "JOIN detected - include gemini-2.5-flash"
```

**Benefits:**
- Non-developers can tune routing
- A/B test different strategies
- Rule precedence explicit
- Audit trail of routing logic

---

### Load Balancing

**Agent Capacity Tracking:**

**Per-Agent Metrics:**
- Current active tasks (count)
- Average task duration (historical)
- Success rate (percentage)
- Current load percentage

**Capacity Limits:**
- Max concurrent tasks: 3 per agent
- Soft limit: 2 (prefer assigning if under 2)
- Hard limit: 3 (never assign if at 3)

**Load Balancing Strategy:**

**Scenario 1: All agents available**
- Assign based on rules alone
- Optimal specialization match

**Scenario 2: One agent at capacity**
- Skip that agent
- Assign alternative if available
- Example: gemini-2.5-flash at capacity → try gemini-2.5-pro for index task

**Scenario 3: Multiple agents at capacity**
- Queue task (wait for agent availability)
- Or assign only available agents
- Log degraded assignment

**Scenario 4: All agents at capacity**
- Queue task in pending state
- Retry assignment when agent frees
- User sees "waiting for agents" status

---

### Router Testing Scenarios

**Test Cases:**

**Simple Query:**
- Query: SELECT * FROM orders WHERE status = 'completed'
- Features: Simple filter, no JOIN
- Expected: gemini-2.5-flash only
- Complexity: 0

**Medium Query:**
- Query: Users JOIN Orders with aggregation
- Features: 1 JOIN, 1 aggregation, GROUP BY
- Expected: gemini-2.5-flash + gemini-2.5-pro
- Complexity: 3

**Complex Query:**
- Query: Multi-table JOINs, subqueries, window functions
- Features: 3 JOINs, 2 subqueries, 1 window function
- Expected: All three agents
- Complexity: 7

**Large Table:**
- Query: Simple query on 5M row table
- Features: Basic query but huge table
- Expected: gemini-2.5-flash + gemini-2.5-pro
- Rationale: gemini-2.5-flash para particionamiento; gemini-2.5-pro para validación

**High Priority:**
- Any query with priority = high
- Expected: All three agents
- Overrides complexity rules

---

## ⚡ Parallel Coordination

### Concurrency Model

**Goroutine-Based Parallelism:**

Each assigned agent executes in its own goroutine, allowing truly 
parallel execution across all phases (fork creation, analysis, proposal, 
benchmarking).

**Orchestrator Coordination:**

**Main Goroutine:**
- Spawns worker goroutines (one per agent)
- Waits for all to complete (or timeout)
- Collects results
- Handles partial failures

**Worker Goroutines:**
- Execute complete agent workflow
- Independent of each other
- No inter-agent communication
- Send results via channel

---

### Synchronization Pattern

**WaitGroup Pattern:**

**Setup:**
```
Create WaitGroup
Set counter to number of agents (e.g., 3)
Create result channel (buffered, size = agent count)
Create error channel (buffered, size = agent count)
```

**Spawn Workers:**
```
FOR each assigned agent:
  Increment WaitGroup counter
  Spawn goroutine:
    Execute agent workflow
    Send result to result channel
    Send error to error channel (if any)
    Decrement WaitGroup counter
```

**Wait for Completion:**
```
In main goroutine:
  Wait on WaitGroup (blocks until all done)
  Close result channel
  Close error channel
```

**Collect Results:**
```
results = []
FOR result in result_channel:
  results.append(result)

errors = []
FOR error in error_channel:
  errors.append(error)
  Log error
```

---

### Timeout Management

**Per-Agent Timeout:**
- Total timeout: 10 minutes per agent
- If agent exceeds, goroutine cancelled via context
- Other agents continue unaffected

**Global Timeout:**
- Total workflow timeout: 15 minutes (all agents)
- If exceeded, cancel all remaining agents
- Collect results from completed agents

**Context Propagation:**

**Parent Context:**
- Created by orchestrator with 15-minute deadline
- Passed to all agents

**Child Contexts:**
- Each agent operation has sub-context
- Example: Fork creation 30s, LLM call 2m, benchmark 5m
- Child timeout shorter than parent

**Cancellation:**
- If parent cancelled, all children cancelled
- Agents check context.Done() periodically
- Clean shutdown (destroy forks before exit)

---

### Error Handling in Parallel Execution

**Partial Failure Tolerance:**

**1 of 3 Agents Fails:**
- Task continues with 2 proposals
- Consensus runs normally
- Successful outcome possible
- Failure logged for debugging

**2 of 3 Agents Fail:**
- Task fails (insufficient proposals)
- Consensus cannot make reliable decision
- User notified
- Cleanup forks from failed agents

**All 3 Agents Fail:**
- Task marked as failed
- Error message aggregates all failures
- User sees detailed error report
- Retry allowed

**Failure Isolation:**
- One agent's failure doesn't affect others
- Each agent in separate goroutine
- Panic recovery per goroutine
- Errors sent via channel, not propagated

---

### Resource Management

**Fork Lifecycle:**

**Creation Phase (Parallel):**
- All agents create forks simultaneously
- 3 forks created in ~5-10 seconds total (not 15-30 sequential)
- Zero-copy makes this feasible

**Usage Phase (Parallel):**
- Each agent works in its own fork
- No cross-fork contamination
- Independent query execution

**Cleanup Phase (Parallel or Sequential):**
- Option 1: Parallel cleanup (faster, 3-5s total)
- Option 2: Sequential cleanup (safer, 10-15s total)
- Cleanup even if agent failed

**Connection Pooling:**

**Challenge:**
- Multiple forks = multiple DB connections
- Connection pool limits (typically 10-20)

**Solution:**
- Each agent gets dedicated connection from pool
- 3 agents = 3 connections max
- Well within typical limits
- MCP handles connection multiplexing

---

### Progress Broadcasting

**Real-Time Updates:**

**Events Emitted During Parallel Execution:**

**agents_assigned:**
- Timestamp: Start of parallel execution
- Data: List of assigned agents

**fork_created (Per Agent):**
- Timestamp: When each agent creates fork
- Data: Agent type, fork ID
- Frequency: 3 events (one per agent)

**analysis_completed (Per Agent):**
- Timestamp: When each agent completes analysis
- Data: Agent type, insights preview
- Frequency: 3 events

**proposal_submitted (Per Agent):**
- Timestamp: When each agent submits proposal
- Data: Agent type, proposal type
- Frequency: 3 events

**benchmark_completed (Per Agent):**
- Timestamp: When each agent finishes benchmarks
- Data: Agent type, preliminary results
- Frequency: 3 events

**User Experience:**
- See agents working in real-time
- Progress bar updates per agent
- Identifies slow/stuck agents
- Engaging visualization

---

## 🎨 Prompt Engineering

### Prompting Principles

**Consistency:**
- All agents use similar prompt structure
- Variations only in specialization content
- Predictable response format

**Specificity:**
- Clear instructions
- Explicit output format (JSON)
- Concrete examples
- Constrained responses

**Context:**
- Provide complete information
- Explain plan output
- Include schema details
- Specify constraints

**Role Assignment:**
- Define agent persona
- Set expertise context
- Emphasize specialization

---

### gemini-2.5-pro Prompt Templates

**System Prompt (Role Assignment):**

```
You are an expert PostgreSQL Database Administrator with 15 years of 
experience specializing in query optimization and index design. 

Your strengths:
- Deep understanding of PostgreSQL execution plans
- Expert knowledge of B-tree index structures
- Proven track record with JOIN optimization
- Conservative, production-safe recommendations

Your approach:
- Analyze step-by-step
- Explain reasoning clearly
- Prefer simple, proven solutions
- Consider operational simplicity
```

**Analysis Prompt Template:**

```
Analyze this PostgreSQL query execution plan and identify optimization 
opportunities.

QUERY:
{target_query}

EXPLAIN ANALYZE OUTPUT:
{explain_plan}

SCHEMA CONTEXT:
Tables: {table_list}
Existing Indexes: {index_list}
Table Sizes: {size_info}

TASK:
Identify performance bottlenecks and recommend optimizations.

OUTPUT FORMAT (JSON):
{
  "insights": [
    "List specific observations from explain plan",
    "Each insight should be actionable"
  ],
  "proposed_actions": [
    "High-level recommendations",
    "Focus on indexes and query structure"
  ],
  "confidence": 0.85,
  "metadata": {
    "bottleneck_type": "sequential_scan | nested_loop | sort",
    "estimated_improvement": "percentage range"
  }
}

Respond ONLY with valid JSON.
```

**Proposal Prompt Template:**

```
Based on your analysis, generate specific SQL commands to optimize 
this query.

ANALYSIS INSIGHTS:
{insights_from_analysis}

REQUIREMENTS:
- Generate production-ready SQL (PostgreSQL 16 syntax)
- Each command must be independently executable
- Prefer indexes over schema changes
- Consider partial indexes when appropriate
- Include index covering frequently filtered columns

OUTPUT FORMAT (JSON):
{
  "proposal_type": "index | composite_index | partial_index",
  "sql_commands": [
    "CREATE INDEX ...",
    "ANALYZE table_name"
  ],
  "rationale": "Explain why this approach is optimal",
  "estimated_impact": {
    "query_time_improvement": 75.0,
    "storage_overhead_mb": 12.0,
    "complexity": "low",
    "risk": "low"
  }
}

Respond ONLY with valid JSON.
```

---

### gemini-2.5-flash Prompt Templates

**System Prompt:**

```
You are a Senior Data Architect specializing in large-scale database 
design and optimization for high-volume systems.

Your expertise:
- Table partitioning strategies (range, list, hash)
- Schema redesign for scalability
- Data distribution analysis
- Long-term growth planning

Your perspective:
- Holistic view of database architecture
- Consider future data growth
- Balance current performance with maintainability
- Structural solutions over tactical fixes
```

**Analysis Prompt Template:**

```
Analyze this query from a data architecture perspective, considering 
structural optimization opportunities.

QUERY:
{target_query}

DATA STATISTICS:
Table Sizes: {table_sizes}
Row Counts: {row_counts}
Data Distribution: {distribution_stats}
Growth Rate: {growth_info}

CURRENT SCHEMA:
{schema_definition}

TASK:
Identify architectural improvements (partitioning, denormalization, etc.)

OUTPUT FORMAT (JSON):
{
  "insights": [
    "Data distribution observations",
    "Scalability concerns",
    "Structural bottlenecks"
  ],
  "proposed_actions": [
    "Partitioning strategies",
    "Schema modifications",
    "Archive patterns"
  ],
  "confidence": 0.80,
  "metadata": {
    "data_hotspots": "description",
    "partition_candidates": ["table_name"]
  }
}

Focus on structural solutions. Consider long-term implications.
Respond ONLY with valid JSON.
```

---

### gemini-2.5-pro Prompt Templates

**System Prompt:**

```
You are an elite Performance Engineer specializing in advanced database 
optimization techniques and complex trade-off analysis.

Your toolkit:
- Materialized views with refresh strategies
- Pre-aggregation patterns
- Denormalization with trigger maintenance
- Hybrid optimization approaches

Your methodology:
- Creative problem-solving
- Rigorous trade-off analysis
- Consider read/write balance
- Maintenance cost awareness
```

**Proposal Prompt Template:**

```
Design an advanced optimization strategy for this high-value query, 
considering sophisticated techniques.

QUERY:
{target_query}

EXECUTION FREQUENCY:
{execution_frequency}

PERFORMANCE REQUIREMENT:
Current: {current_time}ms
Target: <{target_time}ms

CONSTRAINTS:
Max Storage: {max_storage}MB
Staleness Tolerance: {staleness}

AVAILABLE TECHNIQUES:
- Materialized views (with refresh strategy)
- Pre-aggregation tables
- Denormalization with triggers
- Combination approaches

OUTPUT FORMAT (JSON):
{
  "proposal_type": "materialized_view | hybrid | denormalization",
  "sql_commands": [
    "CREATE MATERIALIZED VIEW ...",
    "CREATE INDEX ON ...",
    "-- Refresh strategy comment"
  ],
  "rationale": "Explain trade-offs and why optimal",
  "estimated_impact": {
    "query_time_improvement": 90.0,
    "storage_overhead_mb": 80.0,
    "complexity": "medium",
    "risk": "medium"
  },
  "maintenance_strategy": {
    "refresh_method": "concurrent | incremental",
    "refresh_frequency": "hourly | daily",
    "maintenance_cost": "description"
  }
}

Be creative but practical. Justify complexity with performance gains.
Respond ONLY with valid JSON.
```

---

### Prompt Best Practices

**JSON Enforcement:**
- Explicitly state "Respond ONLY with valid JSON"
- Provide JSON schema in prompt
- Include example responses
- Use JSON mode if LLM supports it

**Error Handling:**
- LLMs sometimes add markdown formatting (```json)
- Strip markdown code fences in parsing
- Validate JSON schema after parsing
- Re-prompt if invalid response

**Token Management:**
- Monitor prompt token count
- Truncate explain plans if too long
- Summarize schema if many tables
- Stay within model limits (per Vertex AI model)

**Consistency:**
- Use same output format across all agents
- Consistent field names
- Same confidence scale (0.0-1.0)
- Standardized metadata keys

---

## 🔍 Response Parsing

### Parsing Strategy

**Challenge:**
LLM responses may include:
- Markdown code fences
- Explanatory text before/after JSON
- Incorrect JSON formatting
- Extra whitespace

**Robust Parsing Steps:**

**Step 1: Extract JSON**
```
1. Trim whitespace
2. Look for markdown code fence (```json or ```)
3. If found, extract content between fences
4. If not found, assume entire response is JSON
5. Trim again
```

**Step 2: Parse JSON**
```
1. Attempt JSON parsing (unmarshal)
2. If fails, try to fix common issues:
   - Remove trailing commas
   - Fix single quotes to double quotes
   - Remove comments (// or /* */)
3. Retry parsing
4. If still fails, return error
```

**Step 3: Validate Schema**
```
1. Check required fields present
2. Validate field types (string, number, array)
3. Validate enum values (proposal_type, etc.)
4. Check array lengths (non-empty where required)
5. Validate ranges (confidence 0-1, percentages 0-100)
```

**Step 4: Sanitize Data**
```
1. Trim string fields
2. Normalize enum values (lowercase)
3. Round numbers to expected precision
4. Remove null/undefined fields
```

---

### Schema Validation

**AnalysisResult Schema:**

```
Required fields:
- insights: array of strings, min length 1
- proposed_actions: array of strings, min length 1
- confidence: number, range 0.0-1.0

Optional fields:
- metadata: object

Validation rules:
- Insights should be specific (min 20 chars each)
- Proposed actions actionable (not vague)
- Confidence realistic (not always 1.0)
```

**OptimizationProposal Schema:**

```
Required fields:
- proposal_type: enum (index, partitioning, etc.)
- sql_commands: array of strings, min length 1
- rationale: string, min 50 chars
- estimated_impact: object with specific structure

estimated_impact validation:
- query_time_improvement: number, range 0-100
- storage_overhead_mb: number, >= 0
- complexity: enum (low, medium, high)
- risk: enum (low, medium, high)

sql_commands validation:
- Each command is valid SQL (basic syntax check)
- No forbidden keywords (DROP DATABASE, etc.)
```

---

### Error Recovery

**Invalid JSON Response:**

**Strategy 1: Re-prompt**
- Send new prompt with error context
- "Your previous response was not valid JSON. Try again."
- Include original prompt + error message
- Limit: 2 retry attempts

**Strategy 2: Partial Extraction**
- Attempt to extract partial data
- If insights present but proposed_actions missing, use insights
- Set confidence to 0.5 (uncertain)
- Log warning

**Strategy 3: Fail Gracefully**
- After retries exhausted, mark agent as failed
- Log complete response for debugging
- Continue with other agents

**Missing Fields:**

**Required Field Missing:**
- If critical field missing (e.g., sql_commands), fail agent
- Cannot proceed without essential data

**Optional Field Missing:**
- Provide default value
- Log warning
- Continue processing

**Invalid Field Values:**

**Out of Range:**
- Clamp to valid range (e.g., confidence > 1.0 → 1.0)
- Log warning

**Invalid Enum:**
- If proposal_type invalid, attempt mapping
- Example: "create_index" → "index"
- If unmappable, fail validation

---

## ❌ Error Handling

### Error Categories

**Transient Errors (Retry):**
- Network timeouts
- LLM API rate limits (429)
- LLM API server errors (5xx)
- MCP temporary unavailability

**Permanent Errors (Fail Fast):**
- Invalid API key (401)
- Invalid SQL syntax (after generation)
- Fork quota exceeded
- Invalid task data

**Degraded Errors (Continue with Warnings):**
- Single benchmark query fails (others succeed)
- Explain plan parsing fails (use raw text)
- Confidence score unrealistic (use default)

---

### Retry Strategy

**Exponential Backoff:**

```
Attempt 1: Immediate
Attempt 2: Wait 1 second
Attempt 3: Wait 2 seconds
Attempt 4: Wait 4 seconds
Max attempts: 3-4

Backoff formula:
wait_time = base_delay × (2 ^ attempt_number)
base_delay = 1 second
```

**Which Operations to Retry:**
- LLM API calls (transient failures common)
- MCP fork creation (occasional server issues)
- MCP query execution (network blips)

**Which NOT to Retry:**
- Invalid SQL syntax (won't fix itself)
- Authentication failures (need new credentials)
- Validation errors (logic error, not transient)

---

### Agent Failure Handling

**Failure in CreateFork:**
- Impact: Cannot proceed with this agent
- Action: Mark agent execution as failed
- Cleanup: N/A (fork never created)
- Task: Continue with other agents

**Failure in AnalyzeTask:**
- Impact: No insights for proposal
- Action: Mark agent execution as failed
- Cleanup: Destroy fork (already created)
- Task: Continue with other agents

**Failure in ProposeOptimization:**
- Impact: No proposal from this agent
- Action: Mark agent execution as failed
- Cleanup: Destroy fork
- Task: Continue with other agents

**Failure in RunBenchmark:**
- Impact: No benchmark data for proposal
- Action: Mark agent execution as failed
- Cleanup: Destroy fork
- Task: Continue with other agents
- Note: Proposal exists but can't be scored

**Failure in DestroyFork:**
- Impact: Orphaned fork (resource leak)
- Action: Log error, schedule retry
- Cleanup: Background job retries later
- Task: Don't fail task (cleanup is housekeeping)

---

### Logging and Monitoring

**Log Levels:**

**ERROR:**
- Agent execution failed completely
- LLM API authentication failed
- MCP service unavailable
- Invalid configuration

**WARN:**
- Single retry attempt failed
- Parsing issues (recovered)
- Degraded performance
- Approaching timeout

**INFO:**
- Agent assigned to task
- Fork created successfully
- Analysis completed
- Proposal submitted
- Benchmark completed

**DEBUG:**
- LLM prompt content
- LLM response content
- SQL execution details
- Timing metrics

**Structured Logging:**

```
Log entry structure:
{
  "timestamp": "ISO 8601",
  "level": "ERROR | WARN | INFO | DEBUG",
  "agent_type": "gemini25pro | gemini25flash | gemini20flash",
  "task_id": 123,
  "fork_id": "afs-fork-...",
  "operation": "create_fork | analyze | propose | benchmark",
  "message": "Human-readable message",
  "error": "Error details if applicable",
  "duration_ms": 1234,
  "metadata": {
    "additional": "context"
  }
}
```

**Metrics to Track:**

**Per Agent:**
- Success rate (percentage)
- Average task duration (ms)
- Average proposal quality (based on consensus rank)
- LLM API cost (approximate)

**Per Operation:**
- Fork creation time
- Analysis duration
- Proposal generation duration
- Benchmark execution time

**System-Wide:**
- Total tasks processed
- Agent assignment distribution
- Failure rate by agent
- Consensus winner distribution

---

## 🎯 Summary

This agent system provides:

**3 Specialized Agents:**
- gemini-2.5-pro: SQL/QA strategist (indexes, validation, trade-offs)
- gemini-2.5-flash: partitioning, schema changes, execution
- gemini-2.0-flash: materialized views, boilerplate

**Intelligent Routing:**
- Rule-based assignment (9 routing rules)
- Complexity scoring (0-10 scale)
- Load balancing (max 3 concurrent per agent)
- Configuration-driven (externalized rules)

**Parallel Execution:**
- Goroutine-based concurrency
- WaitGroup synchronization
- Timeout management (per-agent and global)
- Partial failure tolerance (1 of 3 can fail)

**Robust Prompting:**
- Specialized prompts per agent
- Consistent JSON output format
- Context-rich inputs (explain plans, schema, stats)
- Role-based system prompts

**Error Resilience:**
- Exponential backoff retry (3-4 attempts)
- Multiple parsing strategies
- Graceful degradation
- Comprehensive logging

---

**Related Documentation:**
- Previous: [03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md) 
  - Agent layer details
- Next: [05-CONSENSUS-BENCHMARKING.md](05-CONSENSUS-BENCHMARKING.md) 
  - How proposals are evaluated
- See also: [07-LLM-INTEGRATION.md](07-LLM-INTEGRATION.md) 
  - LLM client implementation

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```

