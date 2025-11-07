# 03-SYSTEM-ARCHITECTURE.md

```markdown
# 03-SYSTEM-ARCHITECTURE.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** System Architecture Specification  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[01-BUSINESS-LOGIC.md](01-BUSINESS-LOGIC.md)

---

## 📖 Table of Contents

1. [Architectural Overview](#architectural-overview)
2. [Clean Architecture Layers](#clean-architecture-layers)
3. [Dependency Rules](#dependency-rules)
4. [Design Patterns](#design-patterns)
5. [Data Flow](#data-flow)
6. [Module Organization](#module-organization)
7. [Error Handling Strategy](#error-handling-strategy)
8. [Testing Strategy](#testing-strategy)

---

## 🏛️ Architectural Overview

### Guiding Principles

**Clean Architecture**
- Business logic isolated from technical details
- Framework independence (can swap Fiber for Gin)
- Database independence (can swap Tiger Cloud for vanilla Postgres)
- Testability without external dependencies
- UI independence (React is replaceable)

**SOLID Principles**
- **Single Responsibility:** Each module has one reason to change
- **Open/Closed:** Extend behavior without modifying existing code
- **Liskov Substitution:** Implementations replaceable by interface
- **Interface Segregation:** Small, focused interfaces
- **Dependency Inversion:** Depend on abstractions, not implementations

**Additional Principles**
- **DRY (Don't Repeat Yourself):** Shared logic extracted to utilities
- **YAGNI (You Aren't Gonna Need It):** Build only what's required now
- **KISS (Keep It Simple):** Prefer simple solutions over complex ones
- **Separation of Concerns:** Each layer has distinct responsibility

### Architecture Style

**Layered Architecture with Clean Architecture variant**

```
┌─────────────────────────────────────────────────┐
│  Interfaces Layer (HTTP/WebSocket Handlers)     │
│  - Receive requests                             │
│  - Validate input                               │
│  - Call use cases                               │
│  - Format responses                             │
└─────────────────┬───────────────────────────────┘
                  │ depends on
                  ▼
┌─────────────────────────────────────────────────┐
│  Use Cases Layer (Application Logic)            │
│  - Orchestrate workflows                        │
│  - Coordinate domain entities                   │
│  - Enforce business rules                       │
└─────────────────┬───────────────────────────────┘
                  │ depends on
                  ▼
┌─────────────────────────────────────────────────┐
│  Domain Layer (Business Core)                   │
│  - Entities and value objects                   │
│  - Business logic                               │
│  - Interface definitions                        │
│  - Zero external dependencies                   │
└─────────────────▲───────────────────────────────┘
                  │ implemented by
                  │
┌─────────────────┴───────────────────────────────┐
│  Infrastructure Layer (Technical Implementation)│
│  - Database access                              │
│  - External APIs (MCP, LLMs)                    │
│  - File system                                  │
│  - Third-party services                         │
└─────────────────────────────────────────────────┘
```

**Cross-cutting: Configuration Layer**
- Loaded at startup
- Injected into all layers
- Environment-specific settings

---

## 🎯 Clean Architecture Layers

### Layer 1: Domain (Inner Circle)

**Location:** `backend/internal/domain/`

**Responsibility:**  
Contains pure business logic and rules. Defines the core entities and 
their behaviors. This is the heart of the application.

**What Goes Here:**

**Entities:**
- Task - Represents optimization request
- Agent - Interface defining agent contract
- OptimizationProposal - Proposed solution
- BenchmarkResult - Performance measurement
- ConsensusDecision - Final decision with scoring
- AgentExecution - Agent work tracking

**Value Objects:**
- AgentType (enum: gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
- TaskStatus (enum: pending, in_progress, completed, failed)
- ProposalType (enum: index, partitioning, etc)
- ScoringCriteria - Consensus weights configuration

**Interfaces (Contracts for outer layers):**
- AgentRepository - Persist agent data
- TaskRepository - Persist task data
- OptimizationRepository - Persist proposals
- BenchmarkRepository - Persist benchmark results
- ConsensusRepository - Persist decisions
- Agent - Core agent interface (all agents implement)

**Business Rules:**
- Task validation logic
- Scoring calculations
- Status transition rules
- Proposal validation

**What Does NOT Go Here:**
- Database queries (no SQL)
- HTTP handlers
- External API calls
- Framework-specific code
- Concrete implementations

**Dependencies:**
- ZERO external dependencies
- Only standard Go library
- No imports from outer layers

**Example Entities (Conceptual):**

```
Task Entity:
- Properties: ID, Type, Description, TargetQuery, Status, Metadata
- Methods: Validate(), CanTransitionTo(newStatus), IsComplete()
- Business Rules: Description max 500 chars, Query not empty

Agent Interface:
- Methods: GetType(), CreateFork(), AnalyzeTask(), ProposeOptimization()
- Contracts: All implementations must provide these behaviors
```

**Testing:**
- 90%+ coverage target
- Pure unit tests (no mocks needed)
- Fast execution (milliseconds)
- No external dependencies

---

### Layer 2: Use Cases (Application Logic)

**Location:** `backend/internal/usecases/`

**Responsibility:**  
Orchestrates domain entities to fulfill application-specific workflows.
Implements business use cases by coordinating multiple entities.

**What Goes Here:**

**Orchestrator:**
- Coordinates complete task optimization workflow
- Spawns goroutines for parallel agent execution
- Collects results from all agents
- Triggers consensus engine
- Applies winning optimization
- Manages cleanup operations
- Broadcasts WebSocket events

**Router:**
- Analyzes incoming tasks
- Calculates complexity scores
- Applies routing rules
- Selects appropriate agents
- Balances agent workload

**ConsensusEngine:**
- Receives all proposals with benchmarks
- Calculates performance scores
- Calculates storage scores
- Calculates complexity scores
- Calculates risk scores
- Applies weighted formula
- Ranks proposals
- Selects winner
- Generates rationale

**BenchmarkRunner:**
- Defines benchmark suite (4 test queries)
- Executes baseline measurements
- Applies optimization in fork
- Runs test queries (3 iterations each)
- Calculates averages
- Measures storage impact
- Parses EXPLAIN plans
- Aggregates metrics

**TaskService:**
- Create new task
- Retrieve task by ID
- List tasks with filters
- Update task status
- Validate task data

**Dependencies:**
- Domain layer entities and interfaces (allowed)
- Repository interfaces from domain (allowed)
- No infrastructure imports (forbidden)
- No HTTP handler imports (forbidden)

**Dependency Injection:**
- Repositories injected via constructor
- Agent implementations injected
- Configuration injected
- MCP client injected (as interface)

**Example Use Case Structure (Conceptual):**

```
Orchestrator:
- Constructor: Receives all dependencies (repositories, agents, etc)
- ExecuteTask(ctx, taskID):
  1. Load task from repository
  2. Call Router.SelectAgents(task)
  3. Spawn goroutines for each agent
  4. Wait for all agents (WaitGroup)
  5. Call ConsensusEngine.Decide(proposals)
  6. Apply winning proposal to main DB
  7. Cleanup forks
  8. Update task status to completed
  9. Broadcast completion event
```

**Error Handling:**
- Return errors up the stack (don't panic)
- Wrap errors with context
- Log errors before returning
- Partial failures tolerated (1 of 3 agents can fail)

**Testing:**
- 80%+ coverage target
- Unit tests with mocked dependencies
- Repository mocks (mockery or manual)
- Agent mocks for testing orchestration
- Test both success and failure paths

---

### Layer 3: Infrastructure (Technical Implementation)

**Location:** `backend/internal/infrastructure/`

**Responsibility:**  
Implements interfaces defined in Domain layer using concrete technologies.
Handles all external communication and technical details.

**Submodules:**

**MCP (Tiger Cloud Integration)**
- `mcp/client.go` - MCP client wrapper
- `mcp/service.go` - Fork lifecycle management
- `mcp/query.go` - Query execution helpers

**Responsibilities:**
- Connect to Tiger Cloud MCP server
- Authenticate with API token
- Execute `service_fork` operation
- Execute `db_execute_query` operation
- Execute `service_delete` operation
- Parse MCP responses
- Handle MCP-specific errors
- Retry logic with exponential backoff

**Interface Implemented:**
- MCPClient (defined in domain or usecases)

**Dependencies:**
- Tiger Cloud MCP SDK (if available) or HTTP client
- Domain interfaces

---

**Agents (AI Implementations)**
- `agents/base.go` - Shared agent logic
- `agents/gemini25pro.go` - gemini-2.5-pro (planner/QA)
- `agents/gemini25flash.go` - gemini-2.5-flash (código/SQL)
- `agents/gemini20flash.go` - gemini-2.0-flash (tareas masivas)

**Responsibilities:**
- Implement Agent interface from domain
- Build prompts for each model
- Call Vertex AI Client
- Parse JSON responses
- Validate proposal SQL
- Create forks via MCP
- Execute queries in forks
- Run benchmarks

**BaseAgent (Shared Logic):**
- Fork creation
- Fork destruction
- Logging
- Error handling
- Prompt template utilities

**Agent-Specific Logic:**
- Specialization-specific prompts
- Response parsing (per model quirks)
- Model selection via Vertex AI

**Interface Implemented:**
- Agent (from domain)

**Dependencies:**
- LLM client libraries
- MCP client
- Domain entities

---

**LLM (Language Model Clients)**
- `llm/vertex_client.go` - Vertex AI API wrapper

**Responsibilities:**
- Authenticate with Vertex AI
- Send prompts
- Receive responses
- Handle rate limiting
- Retry on transient failures
- Parse errors
- Track token usage (optional)
- Cost tracking (optional)

**Interface Implemented:**
- LLMClient (generic interface for all LLMs)

**Dependencies:**
- Google Cloud SDK / Vertex AI client
- Configuration for project/region/credentials

---

**Database (Data Persistence)**
- `database/postgres.go` - Local PostgreSQL (development)
- `database/tiger.go` - Tiger Cloud specific logic
- `database/repositories/` - Repository implementations

**Repositories:**
- `task_repository.go` - Implements TaskRepository interface
- `agent_execution_repository.go` - Implements AgentExecutionRepo
- `optimization_repository.go` - Implements OptimizationRepo
- `benchmark_repository.go` - Implements BenchmarkRepo
- `consensus_repository.go` - Implements ConsensusRepo

**Responsibilities:**
- Execute SQL queries
- Map rows to domain entities
- Handle transactions
- Connection pooling
- Error mapping (SQL errors → domain errors)

**Interface Implemented:**
- All repository interfaces from domain

**Dependencies:**
- pgx (PostgreSQL driver)
- sqlx (query builder)
- Domain entities

**Pattern:**
- Repository Pattern (isolates data access)
- Each repository focuses on one entity
- No business logic in repositories (pure CRUD)

---

**What Does NOT Go Here:**
- Business logic (belongs in domain/usecases)
- HTTP request handling (belongs in interfaces)
- Direct calls from handlers (must go through usecases)

**Dependencies:**
- Domain interfaces (implements them)
- External SDKs and libraries
- Configuration
- No imports from usecases or interfaces layers

**Testing:**
- 60%+ coverage target
- Integration tests preferred (test with real DB/APIs)
- Use Docker for DB in tests
- Mock external APIs (LLMs, MCP) if rate limits concern
- Focus on critical paths (happy path + key error cases)

---

### Layer 4: Interfaces (API & Presentation)

**Location:** `backend/internal/interfaces/`

**Responsibility:**  
Exposes application functionality to external clients. Handles HTTP 
requests, WebSocket connections, and response formatting.

**Submodules:**

**HTTP (REST API)**
- `http/router.go` - Fiber app setup and route definitions
- `http/middleware/` - CORS, logging, auth, error handling
- `http/handlers/` - Request handlers

**Handlers:**
- `task_handler.go` - Task CRUD endpoints
- `agent_handler.go` - Agent status endpoints
- `optimization_handler.go` - Proposals and results endpoints
- `consensus_handler.go` - Decision endpoints
- `health_handler.go` - Health check endpoint

**Responsibilities:**
- Parse HTTP requests
- Validate request bodies (basic validation)
- Call use case methods
- Format responses as JSON
- Set HTTP status codes
- Handle errors gracefully
- Apply middleware

**Pattern:**
- Thin handlers (no business logic)
- Dependency injection of use cases
- Consistent error response format
- RESTful conventions

**Example Handler Structure (Conceptual):**

```
TaskHandler:
- Dependencies: TaskService (use case)
- CreateTask(c *fiber.Ctx):
  1. Parse request body to DTO
  2. Validate DTO fields
  3. Map DTO to domain entity
  4. Call taskService.CreateTask(entity)
  5. If error, return error response
  6. Map entity to response DTO
  7. Return JSON response
```

---

**WebSocket (Real-time Updates)**
- `websocket/hub.go` - Central message broadcaster
- `websocket/client.go` - Individual connection handler
- `websocket/events.go` - Event type definitions

**Responsibilities:**
- Maintain active WebSocket connections
- Register new clients
- Unregister disconnected clients
- Broadcast events to all clients
- Handle ping/pong for keepalive

**Hub Pattern:**
- Central registry of connected clients
- Goroutine runs continuously
- Channels for register, unregister, broadcast
- Thread-safe client management

**Event Types:**
- task_created
- agents_assigned
- fork_created
- analysis_completed
- proposal_submitted
- benchmark_completed
- consensus_reached
- optimization_applied
- task_completed
- task_failed

**Event Structure (Conceptual):**

```
Event:
- Type: string (event type enum)
- TaskID: int (which task this relates to)
- AgentType: string (which agent, if applicable)
- Payload: interface{} (event-specific data)
- Timestamp: time
```

---

**What Does NOT Go Here:**
- Business logic (call use cases instead)
- Direct database access (use repositories via use cases)
- Direct MCP/LLM calls (use infrastructure via use cases)

**Dependencies:**
- Use cases layer (allowed)
- Domain entities (for types only)
- HTTP framework (Fiber)
- WebSocket library
- No infrastructure imports (forbidden)

**Testing:**
- Integration tests (HTTP endpoints)
- Use httptest package
- Test request/response formats
- Test error handling
- Test middleware behavior

---

### Layer 5: Configuration

**Location:** `backend/internal/config/`

**Responsibility:**  
Load and validate configuration from environment variables and files.
Provide typed configuration structs to all layers.

**Files:**
- `config.go` - Main configuration loader
- `tiger.go` - Tiger Cloud specific configuration
- `validation.go` - Configuration validation logic

**Configuration Sections:**

**Server Config:**
- Port number
- Host address
- Environment (development, production)
- Log level (debug, info, warn, error)

**Database Config (Development):**
- PostgreSQL host
- Port
- Database name
- Username
- Password
- Connection pool settings

**Tiger Cloud Config (Production):**
- Main service ID
- MCP server URL
- Authentication token
- Max concurrent forks
- PITR window

**LLM API Config (Vertex AI):**
- GCP project ID
- Region/location (e.g., us-central1)
- Credentials (ADC/Service Account)
- Default model IDs (gemini-2-5-pro, gemini-2-5-flash, gemini-2-0-flash)
- Timeout settings
- Retry settings

**Redis Config:**
- Host and port
- Password
- Database number

**Responsibilities:**
- Load from environment variables
- Load from config files (optional)
- Validate required fields present
- Validate field formats (URLs, etc)
- Provide default values
- Fail fast if invalid configuration

**Pattern:**
- Singleton instance
- Loaded at startup
- Injected as dependency
- Immutable after load

**Testing:**
- Test validation logic
- Test default values
- Test environment variable parsing
- Mock configuration in tests

---

## 🔒 Dependency Rules

### The Dependency Rule

**Core Principle:**  
Source code dependencies must point INWARD. Inner layers never depend 
on outer layers.

**Allowed Dependencies:**

```
Domain: 
  ✓ Standard library only
  ✗ No imports from any other layer

Use Cases:
  ✓ Domain entities and interfaces
  ✗ Infrastructure implementations
  ✗ Interface handlers
  ✗ Config (inject instead)

Infrastructure:
  ✓ Domain interfaces (to implement them)
  ✓ External libraries (pgx, LLM SDKs)
  ✗ Use cases
  ✗ Interface handlers

Interfaces:
  ✓ Use cases
  ✓ Domain entities (for types)
  ✓ HTTP/WebSocket frameworks
  ✗ Infrastructure implementations

Config:
  ✓ Standard library
  ✓ Config libraries (viper, etc)
  ✗ Any application layers
```

### Dependency Inversion

**Problem:**  
Use cases need database access, but can't depend on infrastructure.

**Solution:**  
Domain defines interface, infrastructure implements it, use case depends 
on interface.

**Example (Conceptual):**

```
Domain layer defines:
  TaskRepository interface {
    Create(task) error
    GetByID(id) Task
    Update(task) error
  }

Infrastructure implements:
  PostgresTaskRepository struct {
    implements TaskRepository
    uses pgx to execute SQL
  }

Use case depends on interface:
  TaskService struct {
    repo TaskRepository (interface, not concrete type)
  }

Wiring at startup (main.go):
  repo := PostgresTaskRepository{...}
  service := TaskService{repo: repo}
  
Now TaskService can be tested with mock repository
```

### Avoiding Import Cycles

**Common Mistake:**  
Domain imports infrastructure, infrastructure imports domain → cycle

**Prevention:**
- Domain never imports infrastructure
- Infrastructure imports domain (interfaces only)
- Use interfaces for all cross-layer dependencies

**Main Package Responsibility:**
- Wire up dependencies
- Create concrete implementations
- Inject into use cases
- Inject into handlers
- Start server

---

## 🎨 Design Patterns

### Repository Pattern

**Purpose:**  
Abstract data access logic. Domain doesn't know about SQL.

**Structure:**
- Interface defined in domain
- Implementation in infrastructure
- One repository per aggregate root (Task, Proposal, etc)

**Benefits:**
- Swap database without changing business logic
- Easy to mock for testing
- Centralized data access logic

**Usage in AFS:**
- TaskRepository, OptimizationRepository, BenchmarkRepository, etc.
- All use same pattern consistently

---

### Factory Pattern

**Purpose:**  
Create complex objects without exposing creation logic.

**Usage in AFS:**

**Agent Factory:**
- Creates appropriate agent based on AgentType
- Handles dependency injection for each agent
- Centralizes agent instantiation

**Conceptual:**

```
AgentFactory:
  CreateAgent(agentType AgentType) Agent {
    switch agentType {
      case gemini-2.5-pro: return Gemini25ProAgent{...}
      case gemini-2.5-flash: return Gemini25FlashAgent{...}
      case gemini-2.0-flash: return Gemini20FlashAgent{...}
    }
  }
```

**Benefits:**
- Adding new agent type requires changing only factory
- Handlers don't need to know about agent implementations

---

### Strategy Pattern

**Purpose:**  
Define family of algorithms, make them interchangeable.

**Usage in AFS:**

**Consensus Scoring Strategies:**
- Different scoring algorithms
- Swap algorithm without changing consensus engine
- Example: WeightedScoring, SimpleAveraging, MLBasedScoring

**Routing Strategies:**
- RuleBasedRouting (current)
- LoadBalancedRouting (future)
- MLBasedRouting (future)

**Benefits:**
- Open/Closed principle (open for extension)
- Easy to A/B test different strategies

---

### Observer Pattern

**Purpose:**  
One-to-many dependency. When subject changes, notify all observers.

**Usage in AFS:**

**WebSocket Event Broadcasting:**
- Orchestrator is subject
- WebSocket clients are observers
- When task state changes, all clients notified

**Implementation:**
- Hub maintains list of clients
- Orchestrator sends events to hub
- Hub broadcasts to all connected clients

**Benefits:**
- Decouples event producers from consumers
- Easy to add new event consumers
- Real-time updates for UI

---

### Dependency Injection

**Purpose:**  
Provide dependencies from outside rather than creating internally.

**Usage in AFS:**

**Constructor Injection (Preferred):**

```
Orchestrator receives:
  - TaskRepository
  - ConsensusEngine
  - BenchmarkRunner
  - Agent implementations
  - MCP client
  - WebSocket hub
  
All injected at creation time
```

**Benefits:**
- Testability (inject mocks)
- Flexibility (swap implementations)
- Explicit dependencies (visible in constructor)

**Wiring:**
- Done in main.go or dedicated wire.go
- Use dependency injection framework (optional: wire, fx)
- Manual injection acceptable for simplicity

---

### Adapter Pattern

**Purpose:**  
Convert interface of a class into another interface clients expect.

**Usage in AFS:**

**MCP Client Adapter:**
- Tiger Cloud MCP has specific API
- Wrap it with domain-friendly interface
- Hide MCP-specific details from use cases

**LLM Client Adapters:**
- Each LLM provider has different SDK
- Wrap with common LLMClient interface
- Use cases don't know which LLM being used

**Benefits:**
- Isolate external dependencies
- Easy to swap providers
- Simplified testing (mock adapter interface)

---

## 🔄 Data Flow

### Request Flow (Task Creation Example)

**Step-by-step:**

```
1. HTTP Request arrives
   POST /api/v1/tasks
   Body: {type, description, target_query}

2. Middleware processes
   - CORS check
   - Request logging
   - Authentication (if implemented)

3. TaskHandler.CreateTask()
   - Parse request body to CreateTaskDTO
   - Validate DTO (basic checks)
   - Map DTO to domain.Task entity

4. Call TaskService.CreateTask()
   - Use case validates business rules
   - Creates Task entity
   - Calls TaskRepository.Create()

5. TaskRepository (Infrastructure)
   - Executes SQL INSERT
   - Returns created Task with ID

6. TaskService triggers async orchestration
   - Spawn goroutine
   - Calls Orchestrator.ExecuteTask()

7. Handler formats response
   - Map Task entity to TaskResponseDTO
   - Return JSON response
   - HTTP 201 Created

8. Orchestrator runs in background
   - Router selects agents
   - Agents work in parallel
   - Consensus decides
   - Apply to main DB
   - Broadcast events via WebSocket

9. WebSocket clients receive updates
   - Real-time progress shown in UI
   - Shows "gemini-2.5-pro completed analysis"
```

**Key Points:**
- Request flows inward (Interfaces → Use Cases → Domain)
- Response flows outward (Domain → Use Cases → Interfaces)
- Infrastructure called by use cases, never directly from handlers
- Async work spawned as goroutine from use case

---

### Event Flow (WebSocket Broadcasting)

**Step-by-step:**

```
1. Event occurs in Orchestrator
   Example: Agent completed analysis

2. Orchestrator creates Event
  {
    Type: "analysis_completed",
    TaskID: 123,
    AgentType: "gemini-2.5-pro",
    Payload: {...}
  }

3. Orchestrator sends to WebSocket Hub
   hub.Broadcast(event)

4. Hub receives on broadcast channel
   Hub goroutine listening continuously

5. Hub iterates connected clients
   for client in hub.clients:
     client.send <- event

6. Each Client goroutine sends to connection
   websocket.WriteJSON(event)

7. Frontend receives event
  React component updates UI
  Shows "gemini-2.5-pro completed analysis"
```

**Concurrency:**
- Hub runs in dedicated goroutine
- Each client has dedicated read/write goroutines
- Channels for thread-safe communication

---

### Agent Workflow Data Flow

**Step-by-step:**

```
1. Orchestrator calls agent.CreateFork()
   Agent: Gemini25ProAgent (infrastructure)

2. Gemini25ProAgent calls MCPClient
   MCPClient.ServiceFork(parentID, forkName)

3. MCPClient sends HTTP to Tiger Cloud
   POST to MCP server
   Returns fork service ID

4. Agent stores fork ID
   Calls AgentExecutionRepository.Update()

5. Agent calls own AnalyzeTask()
   Executes EXPLAIN in fork via MCP
   Builds prompt with results

6. Agent calls LLMClient
   VertexLLMClient.SendMessage(model=gemini-2.5-pro, prompt)

7. LLMClient calls Vertex AI API
   HTTP POST with prompt
   Returns JSON response

8. Agent parses response
   Extracts insights, creates AnalysisResult

9. Agent calls ProposeOptimization()
   Generates SQL via another LLM call
   Creates OptimizationProposal entity

10. Agent calls OptimizationRepository.Create()
    SQL INSERT into proposals table

11. Agent calls RunBenchmark()
    Executes queries in fork
    Measures performance
    Creates BenchmarkResult entities

12. Agent calls BenchmarkRepository.CreateBatch()
    SQL INSERT multiple benchmark records

13. Agent returns to Orchestrator
    All data persisted, ready for consensus
```

**Data Transformations:**
- HTTP request → DTO → Domain Entity
- Domain Entity → SQL params → Database
- Database rows → Domain Entity → Response DTO
- External API response → Parsed data → Domain Entity

---

## 📁 Module Organization

### Directory Structure (Conceptual)

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, wiring
│
├── internal/
│   ├── domain/                  # Layer 1: Pure business logic
│   │   ├── entities/
│   │   │   ├── task.go
│   │   │   ├── agent.go
│   │   │   ├── proposal.go
│   │   │   ├── benchmark.go
│   │   │   └── consensus.go
│   │   ├── interfaces/          # Contracts
│   │   │   ├── repositories.go
│   │   │   └── services.go
│   │   └── values/              # Value objects, enums
│   │       ├── agent_type.go
│   │       ├── task_status.go
│   │       └── proposal_type.go
│   │
│   ├── usecases/                # Layer 2: Application logic
│   │   ├── orchestrator.go
│   │   ├── router.go
│   │   ├── consensus_engine.go
│   │   ├── benchmark_runner.go
│   │   └── task_service.go
│   │
│   ├── infrastructure/          # Layer 3: Technical impl
│   │   ├── mcp/
│   │   │   ├── client.go
│   │   │   ├── service.go
│   │   │   └── query.go
│   │   ├── agents/
│   │   │   ├── base.go
│   │   │   ├── gemini25flash.go
│   │   │   ├── gemini20flash.go
│   │   │   └── gemini25pro.go
│   │   ├── llm/
│   │   │   └── vertex_client.go
│   │   └── database/
│   │       ├── postgres.go
│   │       ├── tiger.go
│   │       └── repositories/
│   │           ├── task_repo.go
│   │           ├── agent_execution_repo.go
│   │           ├── optimization_repo.go
│   │           ├── benchmark_repo.go
│   │           └── consensus_repo.go
│   │
│   ├── interfaces/              # Layer 4: API & UI
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   │   ├── cors.go
│   │   │   │   ├── logger.go
│   │   │   │   └── error_handler.go
│   │   │   └── handlers/
│   │   │       ├── task_handler.go
│   │   │       ├── agent_handler.go
│   │   │       ├── optimization_handler.go
│   │   │       ├── consensus_handler.go
│   │   │       └── health_handler.go
│   │   └── websocket/
│   │       ├── hub.go
│   │       ├── client.go
│   │       └── events.go
│   │
│   └── config/                  # Layer 5: Configuration
│       ├── config.go
│       ├── tiger.go
│       └── validation.go
│
├── pkg/                         # Shared utilities (optional)
│   ├── errors/
│   │   └── errors.go
│   ├── logger/
│   │   └── logger.go
│   └── validator/
│       └── validator.go
│
└── tests/                       # Integration tests
    ├── integration/
    └── fixtures/
```

### File Naming Conventions

**Entities:**
- Singular noun: `task.go`, `agent.go`
- One entity per file
- Keep under 300 lines

**Repositories:**
- Pattern: `{entity}_repository.go`
- Example: `task_repository.go`

**Handlers:**
- Pattern: `{resource}_handler.go`
- Example: `task_handler.go`

**Tests:**
- Pattern: `{file}_test.go`
- Same directory as source file
- Integration tests in `/tests`

### Package Naming

**Rules:**
- Lowercase, single word preferred
- No underscores: `websocket` not `web_socket`
- Descriptive: `orchestrator` not `orch`
- Avoid generic names: `handlers` not `http`

---

## ⚠️ Error Handling Strategy

### Error Types

**Domain Errors:**
- Validation errors (invalid task data)
- Business rule violations
- State transition errors

**Infrastructure Errors:**
- Database connection failures
- SQL execution errors
- External API failures (MCP, LLM)
- Network timeouts

**Application Errors:**
- Not found errors (task doesn't exist)
- Conflict errors (duplicate submission)
- Unauthorized errors

### Error Handling Patterns

**Domain Layer:**
- Return errors with context
- Use custom error types for specific cases
- Never panic (except truly unrecoverable)

**Use Cases Layer:**
- Wrap errors with additional context
- Log errors before returning
- Translate infrastructure errors to domain errors

**Infrastructure Layer:**
- Catch and wrap third-party errors
- Add context (which operation failed)
- Retry transient failures

**Interfaces Layer:**
- Catch all errors from use cases
- Map to HTTP status codes
- Return consistent error JSON format

### Error Response Format

**Consistent Structure:**

```
Error Response:
{
  "error": {
    "code": string,      // "VALIDATION_ERROR", "NOT_FOUND", etc.
    "message": string,   // Human-readable message
    "details": object,   // Additional context (optional)
    "timestamp": string
  }
}
```

### Logging

**Levels:**
- ERROR: Unexpected failures
- WARN: Degraded state but continuing
- INFO: Normal operations (task created, etc)
- DEBUG: Detailed trace for development

**What to Log:**
- All errors with full context
- External API calls (request/response)
- State transitions
- Performance metrics (slow queries)

**What NOT to Log:**
- API keys or secrets
- Full SQL queries with sensitive data
- Personally identifiable information

---

## 🧪 Testing Strategy

### Testing Pyramid

```
        ┌────────────┐
        │   E2E      │  Small number, critical paths
        │   Tests    │  Full system, slow
        └────────────┘
       ┌──────────────┐
       │ Integration  │  Medium number
       │   Tests      │  Multiple components, moderate speed
       └──────────────┘
      ┌────────────────┐
      │  Unit Tests    │  Large number, fast
      │  90% coverage  │  Single components, mocked deps
      └────────────────┘
```

### Unit Testing

**Domain Layer:**
- Test all business logic
- No mocks needed (pure functions)
- Fast execution
- 90%+ coverage target

**Example Tests:**
- Task validation rules
- Status transition logic
- Scoring calculations
- Proposal validation

**Use Cases Layer:**
- Mock all dependencies (repositories, agents)
- Test orchestration logic
- Test error handling
- 80%+ coverage target

**Example Tests:**
- Orchestrator handles agent failures
- Router selects correct agents
- Consensus engine calculates scores correctly

**Infrastructure Layer:**
- Integration tests preferred
- Use test database (Docker container)
- Mock external APIs if rate limits concern
- 60%+ coverage target

**Example Tests:**
- Repository CRUD operations
- MCP client fork creation
- LLM client prompt/response

### Integration Testing

**API Endpoints:**
- Test HTTP handlers end-to-end
- Use httptest package
- Test request validation
- Test error responses

**Database:**
- Test with real PostgreSQL
- Use transactions for isolation
- Rollback after each test
- Seed test data in fixtures

**Example Tests:**
- Create task via API returns 201
- Invalid task returns 400
- Get task by ID returns correct data

### End-to-End Testing

**Full Workflow:**
- Submit task → agents work → consensus → applied
- Use test database
- Mock LLM APIs (expensive, rate limited)
- Real MCP if possible (or mocked)

**Critical Path:**
- One complete optimization workflow
- Verify all steps execute
- Verify data persisted correctly

### Test Organization

**Location:**
- Unit tests: Same package as code (`_test.go`)
- Integration tests: `/tests/integration/`
- Fixtures: `/tests/fixtures/`
- Mocks: Generated or `/mocks/` directory

**Naming:**
- Test functions: `TestFunctionName_Scenario_ExpectedResult`
- Example: `TestCreateTask_ValidData_ReturnsCreatedTask`

---

## 📐 Code Quality Enforcement

### Static Analysis Tools

**Recommended:**
- `gofmt` - Code formatting
- `golint` - Style mistakes
- `go vet` - Suspicious constructs
- `staticcheck` - Advanced static analysis

**Run in CI/CD:**
- All tools run on every commit
- Build fails if any issues found

### Code Review Checklist

**Architecture:**
- [ ] Correct layer for new code?
- [ ] Dependencies point inward?
- [ ] Interfaces used for cross-layer?

**SOLID:**
- [ ] Single responsibility per file/function?
- [ ] Interfaces not over-specified?
- [ ] Dependencies injected?

**Code Quality:**
- [ ] File under 300 lines?
- [ ] Lines under 80 characters?
- [ ] No magic numbers?
- [ ] Errors handled explicitly?

**Testing:**
- [ ] Unit tests written?
- [ ] Coverage maintained/improved?
- [ ] Tests pass locally?

---

## 🎯 Summary

This architecture provides:

**5 Clear Layers:**
1. Domain - Pure business logic (0 dependencies)
2. Use Cases - Application workflows (depends on domain)
3. Infrastructure - Technical implementations (implements domain)
4. Interfaces - API/WebSocket (calls use cases)
5. Config - Settings (injected everywhere)

**Key Benefits:**
- Testability (mock any dependency)
- Maintainability (change isolation)
- Flexibility (swap implementations)
- Clarity (explicit dependencies)

**Design Patterns:**
- Repository (data access)
- Factory (agent creation)
- Strategy (scoring algorithms)
- Observer (WebSocket events)
- Dependency Injection (all layers)
- Adapter (external services)

**Quality Standards:**
- SOLID principles enforced
- Clean Architecture rules enforced
- File/line limits enforced
- Testing strategy defined

---

**Related Documentation:**
- Previous: [02-DATA-MODEL.md](02-DATA-MODEL.md) - Database structure
- Next: [04-AGENT-SYSTEM.md](04-AGENT-SYSTEM.md) - Agent deep dive
- See also: [10-DEVELOPMENT-WORKFLOW.md](10-DEVELOPMENT-WORKFLOW.md) - 
  How to develop

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```
