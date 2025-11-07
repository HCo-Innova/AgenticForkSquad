# 00-PROJECT-OVERVIEW.md

```markdown
# 🎯 Agentic Fork Squad (AFS) - Project Overview

**Version:** 1.0  
**Last Updated:** 2024  
**Project Type:** Tiger Cloud Challenge Submission  
**Deadline:** November 9, 2024, 11:59 PM PST

---

## 📖 Table of Contents

1. [Vision & Problem Statement](#vision--problem-statement)
2. [Solution Architecture](#solution-architecture)
3. [Technology Stack](#technology-stack)
4. [Project Status](#project-status)
5. [Documentation Index](#documentation-index)
6. [Glossary](#glossary)
7. [Quick Start Guide](#quick-start-guide)

---

## 🎯 Vision & Problem Statement

### The Problem

Database performance degrades over time in production environments. Current DBA workflows suffer from:

- **High Risk:** Testing optimizations directly in production databases
- **Slow & Costly:** Creating full database copies for testing (time-consuming, storage-expensive)
- **Single Perspective:** One person making optimization decisions without multi-angle validation
- **No Comparison:** Unable to test multiple optimization strategies simultaneously

### The AFS Solution

**Agentic Fork Squad** is a multi-agent collaborative system where:

1. **Multiple AI agents** via Vertex AI (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash) analyze the same database optimization task
2. **Each agent works in isolation** using Tiger Cloud's zero-copy database forks
3. **Agents propose different strategies** (indexes, partitioning, materialized views, query rewrites)
4. **Each proposal is benchmarked** in its own fork with real performance metrics
5. **Consensus algorithm decides** the optimal solution based on weighted criteria
6. **Winning optimization is applied** to the main database safely

### Value Proposition

- **Safety:** All experimentation happens in isolated forks, never touching production
- **Speed:** Zero-copy forks create in <10 seconds regardless of database size
- **Intelligence:** Multiple AI perspectives identify the truly optimal solution
- **Transparency:** Full visibility into why each decision was made
- **Innovation:** Demonstrates creative use of Tiger Cloud's unique capabilities (fork + MCP + hybrid search)

---

## 🏗️ Solution Architecture

### High-Level Flow

```
User Submits Query
         ↓
   Task Router
    (Analyzes & Assigns Agents)
         ↓
    ┌────┴────┬────────┐
    ↓         ↓        ↓
  gemini-2.5-pro   gemini-2.5-flash   gemini-2.0-flash
    │         │        │
    ↓         ↓        ↓
 Fork A    Fork B   Fork C
 (Isolated Experiments)
    │         │        │
    ↓         ↓        ↓
Proposal  Proposal Proposal
    │         │        │
    ↓         ↓        ↓
Benchmark Benchmark Benchmark
    │         │        │
    └────┬────┴────────┘
         ↓
  Consensus Engine
  (Scores & Decides)
         ↓
   Apply to Main DB
         ↓
    Cleanup Forks
```

### Clean Architecture Layers

**Layer 1: Domain** (Business Core)
- Pure business logic and entities
- No external dependencies
- Defines interfaces for other layers to implement
- Components: Agent, Task, Optimization, Benchmark, Consensus entities

**Layer 2: Use Cases** (Application Logic)
- Orchestrates domain entities
- Implements business workflows
- Components: Orchestrator, Router, ConsensusEngine, BenchmarkRunner

**Layer 3: Infrastructure** (Technical Implementation)
- Implements domain interfaces with concrete technologies
- Components: MCP Client, LLM Clients, Database Access, Agent Implementations

**Layer 4: Interfaces** (Presentation & API)
- HTTP REST API handlers
- WebSocket real-time communication
- Components: Fiber routes, WebSocket hub

**Layer 5: Configuration**
- Environment variables
- Tiger Cloud credentials
- Agent API keys

### Data Flow Pattern

```
HTTP Request → Handler → Use Case → Domain Logic → Infrastructure → External Service
                  ↓
            WebSocket Event Broadcast (real-time updates to UI)
```


# 00-PROJECT-OVERVIEW.md (ACTUALIZADO)

Agrego nueva sección después de "Solution Architecture" y antes de "Technology Stack":

```markdown
---

## 🏆 Code Quality Standards

### Architectural Principles

**SOLID Principles** (Mandatory)
- **S**ingle Responsibility: Each module/function has one reason to change
- **O**pen/Closed: Open for extension, closed for modification
- **L**iskov Substitution: Interfaces must be substitutable by implementations
- **I**nterface Segregation: No client forced to depend on unused methods
- **D**ependency Inversion: Depend on abstractions, not concretions

**Clean Architecture** (Enforced)
- Domain layer: Zero external dependencies
- Use Cases: Depend only on Domain
- Infrastructure: Implements Domain interfaces
- Interfaces: Depend on Use Cases, never on Infrastructure directly
- **Dependency Rule:** Inner layers never know about outer layers

**Design Patterns**
- Repository Pattern: Data access abstraction
- Factory Pattern: Agent creation
- Strategy Pattern: Consensus scoring algorithms
- Observer Pattern: WebSocket event broadcasting
- Dependency Injection: All external dependencies injected

### Code Quality Rules

**File Organization**
- **Maximum file length:** 300 lines per source file
- **Maximum line length:** 80 characters per line
- **Single responsibility:** One struct/interface per file when logical
- **Naming convention:** Clear, descriptive names over comments

**Clean Code Practices**
- **Functions:** Small, focused, single purpose (max 20-30 lines)
- **No magic numbers:** Use named constants
- **Error handling:** Explicit, never ignored
- **Comments:** Explain "why", not "what" (code should be self-documenting)
- **DRY principle:** Don't Repeat Yourself
- **YAGNI principle:** You Aren't Gonna Need It (no premature optimization)

**Go-Specific Standards**
- Follow official Go conventions (gofmt, golint)
- Package names: lowercase, single word, no underscores
- Exported vs unexported: Capital for public, lowercase for private
- Error handling: Always check errors, use custom error types
- Context propagation: First parameter in functions
- Interfaces: Small, focused (1-3 methods ideal)

**TypeScript/React Standards**
- Strict TypeScript mode enabled
- No `any` types (use `unknown` if truly dynamic)
- Functional components with hooks (no class components)
- Props interfaces defined explicitly
- Custom hooks for reusable logic

### Testing Requirements

**Coverage Targets**
- Domain layer: 90%+ coverage (pure logic, easy to test)
- Use Cases: 80%+ coverage
- Infrastructure: 60%+ coverage (focus on business-critical paths)
- Interfaces: Integration tests for happy paths

**Test Organization**
- Unit tests: Same package as code, `_test.go` suffix
- Integration tests: Separate `/tests` directory
- Mocks: Generated or in `/mocks` directory
- Test data: Fixtures in `/testdata` directory

### Code Review Checklist

Before committing code, verify:
- [ ] No file exceeds 300 lines
- [ ] No line exceeds 80 characters
- [ ] SOLID principles respected
- [ ] Clean Architecture boundaries not violated
- [ ] All errors handled explicitly
- [ ] No TODO/FIXME comments (use issue tracker)
- [ ] Tests written and passing
- [ ] Documentation updated if public API changed

### Project Structure

**Directory layout already prepared following Clean Architecture:**

```
backend/
├── cmd/                    # Application entry points
├── internal/
│   ├── domain/            # Layer 1: Business entities & interfaces
│   ├── usecases/          # Layer 2: Application logic
│   ├── infrastructure/    # Layer 3: Technical implementations
│   │   ├── mcp/          # Tiger Cloud MCP client
│   │   ├── agents/       # LLM agent implementations
│   │   ├── llm/          # LLM API clients
│   │   └── database/     # Data access implementations
│   ├── interfaces/        # Layer 4: API & WebSocket handlers
│   │   ├── http/
│   │   └── websocket/
│   └── config/           # Layer 5: Configuration
├── pkg/                   # Shared utilities (if needed)
└── tests/                # Integration tests
```

**Dependency flow enforcement:**
```
cmd → interfaces → usecases → domain
              ↓         ↓
       infrastructure
```

**Note:** Full directory structure available upon request for detailed task planning.

---
```

Y actualizo la sección "For AI Assistants" en Quick Start Guide:

```markdown
### For AI Assistants

**When starting a new development task:**

1. **Context Loading:** Read this overview + relevant specific document for the task area
2. **Verify Understanding:** Confirm which layer of Clean Architecture the task belongs to
3. **Check Dependencies:** Ensure no prohibited cross-layer dependencies
4. **Follow Patterns:** Match existing architectural patterns in the codebase
5. **Code Quality:**
   - Keep files under 300 lines
   - Keep lines under 80 characters
   - Apply SOLID principles
   - Write self-documenting code
6. **Document Changes:** Update relevant docs if introducing new concepts

**Quality Gates (Must Pass):**
- Clean Architecture layer boundaries respected
- SOLID principles applied
- File/line length limits enforced
- Error handling comprehensive
- No magic numbers or unexplained constants
- Tests written for business logic
```

---

## 💻 Technology Stack

### Development Phase (Local Environment)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend** | Go 1.21+ | REST API, business logic, agent orchestration |
| **Web Framework** | Fiber v2 | HTTP routing, middleware, WebSocket support |
| **Database** | PostgreSQL 16 | Local development database |
| **Cache** | Redis 7 | Session management, task queue |
| **Frontend** | React 18 + TypeScript | User interface |
| **Build Tool** | Vite | Frontend bundling, hot reload |
| **Styling** | Tailwind CSS | Utility-first styling |
| **Reverse Proxy** | Caddy 2 | Local SSL, routing |
| **Containerization** | Docker + Docker Compose | Service orchestration |

### Production Phase (Tiger Cloud Challenge)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Database Engine** | Tiger Cloud Postgres | Main DB + dynamic forks with zero-copy |
| **Agent Protocol** | Tiger MCP Server | Standardized agent ↔ database communication |
| **Fork Management** | Tiger CLI + Custom Layer | Create, query, destroy database forks |
| **Hybrid Search** | pg_text + pgvector | Full-text + semantic similarity search |
| **Storage** | Fluid Storage | Shared data between forks (no duplication) |
| **AI Agents** | Vertex AI Client (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash) | Especialización por rol |

### AI Agent Stack (por defecto)

| Model | Source | Specialization |
|-------|--------|----------------|
| **gemini-2.5-pro** | Vertex AI | Planner/QA: desambiguación, planificación, verificación de SQL/código |
| **gemini-2.5-flash** | Vertex AI | Generación/ejecución: SQL/código, transformaciones y pruebas |
| **gemini-2.0-flash** | Vertex AI | Bajo costo: tareas masivas, boilerplate y refactors simples |

---

## 📊 Project Status

### ✅ Completed

**Infrastructure**
- [x] Docker Compose setup with all services
- [x] PostgreSQL local database with health checks
- [x] Redis cache configured
- [x] Caddy reverse proxy configured
- [x] Go backend with hot reload (Air)
- [x] React frontend with Vite
- [x] Basic schema: `users`, `orders`, `payments` tables
- [x] Seeder: 1,000 users + 10,000 orders with realistic data

**Tiger Cloud**
- [x] Account created and authenticated
- [x] Tiger CLI installed and configured locally
- [x] Initial service instance created
- [x] MCP configuration file present

### 🚧 In Progress (Current Phase)

**Documentation**
- [x] Project overview document (this file)
- [ ] All supporting technical documents (11 remaining)

**Core System Development**
- [ ] AFS-specific database tables (tasks, agent_executions, proposals, etc.)
- [ ] Database migrations system
- [ ] Domain entities and interfaces
- [ ] MCP client implementation
- [ ] Agent system architecture
- [ ] Task orchestrator
- [ ] Consensus engine
- [ ] Benchmark runner

**Integrations**
- [ ] Vertex AI Client
- [ ] Tiger Cloud fork management
- [ ] WebSocket real-time system

**Frontend**
- [ ] Task submission UI
- [ ] Real-time progress visualization
- [ ] Proposal comparison dashboard
- [ ] Consensus decision display

### 🎯 Upcoming Milestones

**Phase 1: Core Backend (Days 1-3)**
- Implement all database tables and migrations
- Build domain layer entities
- Create use case orchestrators
- Implement MCP client for Tiger Cloud
- Build agent interface and base implementation

**Phase 2: Agent Intelligence (Days 4-5)**
- Integrate Vertex AI Client (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
- Implement specialized agent logic
- Build task router with assignment rules
- Create consensus engine with scoring algorithm
- Implement benchmark runner

**Phase 3: Tiger Cloud Migration (Day 6)**
- Migrate from local PostgreSQL to Tiger Cloud
- Implement fork lifecycle management
- Test zero-copy fork creation
- Validate MCP operations
- Implement hybrid search (if time permits)

**Phase 4: Frontend & Polish (Day 7)**
- Complete React UI components
- WebSocket real-time updates
- Final testing and bug fixes
- Deployment preparation
- Documentation for judges

---

## 📚 Documentation Index

### Core Documentation

- **00-PROJECT-OVERVIEW.md** ← You are here
- **01-BUSINESS-LOGIC.md** - End-to-end flows, actors, business rules
- **02-DATA-MODEL.md** - Database tables, relationships, JSONB structures

### Architecture Documentation

- **03-SYSTEM-ARCHITECTURE.md** - Clean Architecture layers, dependencies, patterns
- **04-AGENT-SYSTEM.md** - Agent interface, specializations, routing, coordination
- **05-CONSENSUS-BENCHMARKING.md** - Scoring algorithm, benchmark suite, decision logic

### Integration Documentation

- **06-TIGER-CLOUD-MCP.md** - MCP protocol, fork operations, Tiger Cloud features
- **07-LLM-INTEGRATION.md** - Vertex AI Client, models, prompt engineering, response parsing

### Interface Documentation

- **08-API-SPECIFICATION.md** - REST endpoints, WebSocket events, request/response schemas
- **09-FRONTEND-COMPONENTS.md** - React components, hooks, state management

### Operational Documentation

- **10-DEVELOPMENT-WORKFLOW.md** - Local setup, testing, debugging
- **11-DEPLOYMENT-STRATEGY.md** - Tiger Cloud migration, production config, validation

---

## 📖 Glossary

### Key Terms

**AFS (Agentic Fork Squad)**
The name of this project - a multi-agent system using database forks for collaborative optimization.

**Agent**
An AI worker using Vertex AI (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash) that analyzes queries, proposes optimizations y ejecuta benchmarks.

**Fork**
An isolated copy of the database created using Tiger Cloud's zero-copy technology. Allows experimentation without affecting the main database.

**Zero-Copy Fork**
Tiger Cloud's technique for creating database forks instantly by copying only metadata, not actual data. Forks share underlying storage.

**MCP (Model Context Protocol)**
A standardized protocol that allows AI agents to interact with tools (databases, APIs, etc.) uniformly. Tiger Cloud implements an MCP server for database operations.

**Orchestrator**
The central coordinator that manages the complete optimization workflow from task submission to consensus decision.

**Task Router**
Component that analyzes incoming optimization tasks and decides which agents should work on them based on rules and agent specializations.

**Proposal**
An optimization strategy suggested by an agent (e.g., create an index, partition a table, create a materialized view).

**Benchmark Suite**
A standardized set of test queries used to measure the performance impact of each proposal.

**Consensus Engine**
Algorithm that evaluates all agent proposals using weighted scoring criteria and selects the optimal solution.

**Weighted Scoring**
Decision method that combines multiple metrics (performance, storage, complexity, risk) with different importance weights to calculate a final score.

**Hybrid Search**
Technique combining traditional full-text search (pg_text) with semantic vector similarity (pgvector) for intelligent query pattern matching.

**Clean Architecture**
Software design pattern that separates business logic from technical implementation through defined layers with strict dependency rules.

**PITR (Point-in-Time Recovery)**
Tiger Cloud feature that allows forking a database from a specific timestamp in the past.

**Fluid Storage**
Tiger Cloud's shared storage system that enables zero-copy forks by allowing multiple forks to reference the same underlying data.

### Component Abbreviations

- **UI** - User Interface
- **API** - Application Programming Interface
- **REST** - Representational State Transfer
- **JSONB** - JSON Binary (PostgreSQL data type)
- **LLM** - Large Language Model
- **DBA** - Database Administrator
- **CRUD** - Create, Read, Update, Delete

---

## 🚀 Quick Start Guide

### For New Developers

1. **Start Here:** Read this document completely to understand the vision
2. **Understand the Business:** Read `01-BUSINESS-LOGIC.md` for the end-to-end flow
3. **Learn the Data:** Read `02-DATA-MODEL.md` to understand database structure
4. **Dive into Architecture:** Read `03-SYSTEM-ARCHITECTURE.md` for technical layers

### For AI Assistants

**When starting a new development task:**

1. **Context Loading:** Read this overview + relevant specific document for the task area
2. **Verify Understanding:** Confirm which layer of Clean Architecture the task belongs to
3. **Check Dependencies:** Ensure no prohibited cross-layer dependencies
4. **Follow Patterns:** Match existing architectural patterns in the codebase
5. **Document Changes:** Update relevant docs if introducing new concepts

### For Code Reviewers

**Key Quality Checkpoints:**

- Does the code respect Clean Architecture layer boundaries?
- Are domain entities free of infrastructure dependencies?
- Is error handling comprehensive (especially for external API calls)?
- Are agent operations properly isolated (using correct fork IDs)?
- Is WebSocket broadcasting implemented for user-facing operations?
- Are Tiger Cloud operations using proper MCP protocol?

---

## ⚠️ Important Constraints

### Time Constraint
- **Deadline:** November 9, 2024, 11:59 PM PST
- **Estimated Development Time:** 7 days
- **Priority:** Functional demo > Feature completeness

### Technical Constraints
- **Tiger Cloud Free Tier:** 24-hour PITR, shared storage
- **Fork Limit:** Recommend max 5 concurrent forks to avoid connection pool issues
- **Vertex AI Quotas:** Respect quotas and rate limits per project/region
- **Storage Budget:** Minimize overhead, leverage zero-copy forks

### Scope Constraints
- **MVP Focus:** One complete optimization workflow working perfectly
- **Agent Count:** 2-3 agents minimum (can skip Gemini if time-constrained)
- **Benchmark Suite:** 3-4 test queries sufficient for demo
- **UI Complexity:** Functional > Beautiful (judges care about backend innovation)

---

## 🎯 Success Criteria

### For Tiger Cloud Challenge Judges

**Innovation (35 points)**
- Creative use of zero-copy forks for parallel agent experimentation
- Novel multi-agent consensus system
- Hybrid search integration (bonus points)

**Technical Implementation (30 points)**
- Clean Architecture with proper separation of concerns
- MCP protocol correctly implemented
- Robust error handling and logging

**Functionality (25 points)**
- Complete end-to-end workflow functional
- Real performance improvements demonstrated
- Multiple agents collaborating successfully

**Deployment (10 points)**
- Working deployment on Tiger Cloud
- Accessible demo with test credentials
- Clear documentation for judges

### Internal Success Metrics

- **Fork Speed:** <10s to create fork from 1GB database
- **Agent Coordination:** ≥2 agents working in parallel successfully
- **Consensus Accuracy:** Winning proposal consistently performs best in validation
- **System Reliability:** Handles partial agent failures gracefully

---

## 📞 Project Meta

**Team Size:** 1-4 developers (list DEV.to handles in submission)  
**Repository:** Private during development, public after submission  
**Submission Format:** DEV.to post using official template  
**Demo Requirements:** Functional deployment + test credentials for judges

---

## 🔄 Document Maintenance

**This document should be updated when:**
- Project milestones are completed (update status section)
- New major components are added (update architecture diagram)
- Technology decisions change (update stack table)
- New documentation is created (update index)

**Owner:** Project lead  
**Review Frequency:** Daily during active development  
**Version Control:** Track changes in git commit messages

---

**Last Section Update:** Project initialization  
**Next Review Date:** After Phase 1 completion
```
