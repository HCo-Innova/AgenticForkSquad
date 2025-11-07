# 10-DEVELOPMENT-WORKFLOW.md

```markdown
# 10-DEVELOPMENT-WORKFLOW.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Development Workflow Guide  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md), 
[11-DEPLOYMENT-STRATEGY.md](11-DEPLOYMENT-STRATEGY.md)

---

## 📖 Table of Contents

1. [Development Environment Setup](#development-environment-setup)
2. [Project Structure](#project-structure)
3. [Local Development](#local-development)
4. [Database Management](#database-management)
5. [Testing Strategy](#testing-strategy)
6. [Code Quality](#code-quality)
7. [Debugging](#debugging)
8. [Git Workflow](#git-workflow)
9. [Common Tasks](#common-tasks)
10. [Troubleshooting](#troubleshooting)

---

## 🛠️ Development Environment Setup

### Prerequisites

**Required Software:**

| Software | Version | Purpose |
|----------|---------|---------|
| Docker | 24+ | Container orchestration |
| Docker Compose | 2.20+ | Multi-container management |
| Git | 2.40+ | Version control |
| Node.js | 20 LTS | Frontend development |
| Go | 1.21+ | Backend development |
| Make | Any | Task automation (optional) |

**Recommended Tools:**

| Tool | Purpose |
|------|---------|
| VS Code | Code editor |
| DBeaver / TablePlus | Database GUI |
| Postman / Insomnia | API testing |
| Docker Desktop | Docker GUI (optional) |

---

### Initial Setup

**Step 1: Clone Repository**

```bash
git clone https://github.com/your-org/afs.git
cd afs
```

**Step 2: Environment Configuration**

**Copy example environment file:**
```bash
cp .env.example .env
```

**Edit .env file:**
```bash
# PostgreSQL (Development)
POSTGRES_DB=afs_dev
POSTGRES_USER=afs_user
POSTGRES_PASSWORD=afs_password

# Redis
REDIS_PASSWORD=afs_redis_pass

# Backend
PORT=8000
ENV=development
LOG_LEVEL=debug
RUN_MIGRATIONS=true

# Vertex AI (Google Cloud - Only Gemini models)
VERTEX_PROJECT_ID=your-gcp-project
VERTEX_LOCATION=us-central1
GEMINI_CEREBRO_MODEL=gemini-2.5-pro          # Planner/QA
GEMINI_OPERATIVO_MODEL=gemini-2.5-flash      # Generación/Ejecución
GEMINI_BULK_MODEL=gemini-2.0-flash           # Bajo costo
GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-service-account.json

# Tiger Cloud (disabled for local dev)
USE_TIGER_CLOUD=false
TIGER_MAIN_SERVICE=
```

**Security Note:**
- Never commit .env file to git
- Use .env.example as template
- Rotate API keys regularly

---

**Step 3: Install Dependencies**

**Backend (Go):**
```bash
cd backend
go mod download
```

**Frontend (Node.js):**
```bash
cd frontend
npm install
```

---

**Step 4: Verify Docker**

**Check Docker running:**
```bash
docker --version
docker-compose --version
```

**Expected output:**
```
Docker version 24.x.x
Docker Compose version 2.20.x
```

---

### IDE Configuration

**VS Code (Recommended)**

**Extensions to Install:**
```
Go (golang.go)
ESLint (dbaeumer.vscode-eslint)
Prettier (esbenp.prettier-vscode)
Tailwind CSS IntelliSense (bradlc.vscode-tailwindcss)
Docker (ms-azuretools.vscode-docker)
Thunder Client (rangav.vscode-thunder-client)
```

**Workspace Settings (.vscode/settings.json):**
```
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "editor.rulers": [80],
  "files.trimTrailingWhitespace": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

**Benefits:**
- Auto-format on save
- Linting errors shown inline
- 80-character ruler (line limit)
- Consistent formatting across team

---

## 📁 Project Structure

### Root Directory

```
afs/
├── backend/                 # Go application
├── frontend/                # React application
├── infrastructure/          # Docker, Caddy config
├── docs/                    # Documentation (this file)
├── scripts/                 # Utility scripts
├── .env.example             # Environment template
├── .gitignore               # Git ignore rules
├── docker-compose.yml       # Service orchestration
├── Makefile                 # Task automation (optional)
└── README.md                # Quick start guide
```

---

### Backend Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
│
├── internal/
│   ├── domain/              # Business entities
│   │   ├── entities/
│   │   ├── interfaces/
│   │   └── values/
│   │
│   ├── usecases/            # Application logic
│   │   ├── orchestrator.go
│   │   ├── router.go
│   │   └── consensus_engine.go
│   │
│   ├── infrastructure/      # External integrations
│   │   ├── mcp/
│   │   ├── agents/
│   │   ├── llm/
│   │   └── database/
│   │
│   ├── interfaces/          # API handlers
│   │   ├── http/
│   │   └── websocket/
│   │
│   └── config/              # Configuration
│
├── migrations/              # Database migrations
│   ├── 001_initial_schema.sql
│   └── 002_afs_tables.sql
│
├── scripts/                 # Helper scripts
│   ├── seed/
│   │   └── main.go          # Data seeder
│   └── migrate/
│       └── main.go          # Migration runner
│
├── tests/                   # Integration tests
│   ├── integration/
│   └── fixtures/
│
├── go.mod                   # Go dependencies
├── go.sum                   # Dependency checksums
└── .air.toml                # Air config (hot reload)
```

**Key Directories:**

**cmd/server:**
- Application entry point
- Dependency wiring
- Server startup

**internal/domain:**
- Pure business logic
- Zero external dependencies
- Core entities and rules

**internal/usecases:**
- Application workflows
- Orchestration logic
- Business use cases

**internal/infrastructure:**
- External service implementations
- Database repositories
- LLM clients
- MCP integration

**internal/interfaces:**
- HTTP handlers
- WebSocket handlers
- Request/response DTOs

**migrations:**
- SQL migration files
- Numbered sequentially
- Up and down migrations

---

### Frontend Structure

```
frontend/
├── src/
│   ├── components/          # React components
│   │   ├── layout/
│   │   ├── task/
│   │   ├── agent/
│   │   ├── optimization/
│   │   ├── consensus/
│   │   └── common/
│   │
│   ├── hooks/               # Custom hooks
│   ├── services/            # API clients
│   ├── types/               # TypeScript types
│   ├── context/             # React contexts
│   ├── utils/               # Utilities
│   ├── pages/               # Route pages
│   │
│   ├── App.tsx              # Root component
│   ├── main.tsx             # Entry point
│   └── index.css            # Global styles
│
├── public/                  # Static assets
│   ├── favicon.ico
│   └── logo.png
│
├── package.json             # Node dependencies
├── tsconfig.json            # TypeScript config
├── vite.config.ts           # Vite config
├── tailwind.config.js       # Tailwind config
└── .eslintrc.json           # ESLint config
```

---

### Infrastructure Structure

```
infrastructure/
├── docker/
│   ├── backend/
│   │   ├── Dockerfile.dev       # Dev container
│   │   ├── Dockerfile.prod      # Prod container
│   │   └── entrypoint.sh        # Startup script
│   │
│   ├── frontend/
│   │   ├── Dockerfile.dev
│   │   └── Dockerfile.prod
│   │
│   └── caddy/
│       └── Caddyfile             # Reverse proxy config
│
└── scripts/
    ├── init-db.sh                # Database init
    └── cleanup.sh                # Resource cleanup
```

---

## 🚀 Local Development

### Starting the Application

**Full Stack (Recommended):**

```bash
# From project root
docker-compose up

# Or with rebuild
docker-compose up --build

# Or in background (detached)
docker-compose up -d
```

**Process:**
1. PostgreSQL starts (port 5432)
2. Redis starts (port 6379)
3. Backend waits for DB healthy
4. Backend runs migrations
5. Backend starts (port 8000)
6. Frontend starts (port 3000)
7. Caddy starts (port 80, 443)

**Verify Services:**
```bash
docker-compose ps
```

**Expected output:**
```
NAME                STATUS    PORTS
afs-postgres        Up        5432
afs-redis           Up        6379
afs-backend         Up        8000
afs-frontend        Up        3000
afs-caddy           Up        80, 443
```

**Access Application:**
```
Frontend: http://localhost
Backend API: http://localhost/api
Health Check: http://localhost/health
```

---

### Hot Reload

**Backend (Air):**

**How it works:**
- Air watches Go files for changes
- On change: Recompiles and restarts server
- No manual restart needed

**Configuration (.air.toml):**
```
[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor", "tests"]
  delay = 1000

[log]
  time = true
```

**Verify working:**
1. Edit any Go file
2. Save file
3. Watch terminal: "Building..." → "Running..."
4. ~2-3 seconds for rebuild

---

**Frontend (Vite):**

**How it works:**
- Vite watches React/TS files
- On change: Hot Module Replacement (HMR)
- Instant updates without full reload

**Verify working:**
1. Edit any React component
2. Save file
3. Browser updates immediately (no refresh)
4. <500ms typically

---

### Stopping the Application

**Graceful shutdown:**
```bash
docker-compose down
```

**Remove volumes (clears database):**
```bash
docker-compose down -v
```

**Remove everything (including images):**
```bash
docker-compose down -v --rmi all
```

---

### Individual Service Development

**Backend Only:**

```bash
# Terminal 1: Start dependencies
docker-compose up postgres redis

# Terminal 2: Run backend locally
cd backend
go run cmd/server/main.go
```

**Benefits:**
- Faster startup
- Direct debugging
- No container overhead

**Requirements:**
- Go installed locally
- Adjust DATABASE_URL to localhost

---

**Frontend Only:**

```bash
# Terminal 1: Start backend
docker-compose up postgres redis backend

# Terminal 2: Run frontend locally
cd frontend
npm run dev
```

**Benefits:**
- Faster HMR
- Direct debugging
- Full Vite features

---

## 🗄️ Database Management

### Migrations

**Migration Files:**

**Location:** `backend/migrations/`

**Naming:** `{number}_{description}.sql`
```
001_initial_schema.sql
002_afs_tables.sql
003_add_indexes.sql
```

**File Structure:**
```
-- Migration: 002_afs_tables
-- Description: Create AFS system tables

-- +migrate Up
CREATE TABLE tasks (
  id SERIAL PRIMARY KEY,
  type VARCHAR(50) NOT NULL,
  ...
);

-- +migrate Down
DROP TABLE IF EXISTS tasks CASCADE;
```

---

**Running Migrations:**

**Automatic (on startup):**
```
Set in .env:
RUN_MIGRATIONS=true

Migrations run when backend container starts
```

**Manual (recommended for production):**
```bash
# Using goose
docker-compose exec backend \\
  goose -dir /app/migrations postgres "${DATABASE_URL}" up

# Or custom script
docker-compose exec backend \\
  go run scripts/migrate/main.go up
```

**Rollback:**
```bash
# Down one migration
docker-compose exec backend \\
  goose -dir ./migrations postgres "${DATABASE_URL}" down 1

# Down to specific version
docker-compose exec backend \\
  goose -dir ./migrations postgres "${DATABASE_URL}" up-to 2
```

---

**Creating New Migrations:**

**Process:**
1. Determine next number (check existing)
2. Create file: `003_add_feature.sql`
3. Write Up section (apply changes)
4. Write Down section (revert changes)
5. Test locally before committing

**Example:**
```
-- +migrate Up
ALTER TABLE tasks ADD COLUMN priority VARCHAR(20);

-- +migrate Down
ALTER TABLE tasks DROP COLUMN priority;
```

---

### Seeding Data

**Automatic Seeding:**

**On first run:**
```
Backend detects empty database
Runs seeder automatically
Creates:
  - 1,000 users
  - 10,000 orders
  - 10,000 payments
```

**Manual Seeding:**

```bash
# Run seeder
docker-compose exec backend \
  go run scripts/seed/main.go

# Or with specific counts
docker-compose exec backend \
  go run scripts/seed/main.go \
    --users=500 \
    --orders=5000
```

**Seeder Logic:**
- Truncates existing data
- Resets sequences
- Creates realistic test data
- Uses gofakeit for generation

---

### Database Access

**Connection via CLI:**

```bash
# Using docker-compose
docker-compose exec postgres psql \
  -U afs_user \
  -d afs_dev

# Direct connection
psql postgresql://afs_user:afs_password@localhost:5432/afs_dev
```

**Useful SQL Commands:**
```sql
-- List tables
\dt

-- Describe table
\d tasks

-- Count records
SELECT COUNT(*) FROM tasks;

-- Check migrations
SELECT * FROM schema_migrations;

-- Active connections
SELECT * FROM pg_stat_activity;
```

---

**Connection via GUI:**

**DBeaver:**
```
Host: localhost
Port: 5432
Database: afs_dev
Username: afs_user
Password: afs_password
```

**TablePlus:**
```
Connection type: PostgreSQL
Host/Socket: localhost
Port: 5432
User: afs_user
Password: afs_password
Database: afs_dev
```

---

### Database Backup & Restore

**Backup:**
```bash
# Backup to file
docker-compose exec -T postgres pg_dump \
  -U afs_user \
  -d afs_dev \
  > backup_$(date +%Y%m%d).sql

# Or with compression
docker-compose exec -T postgres pg_dump \
  -U afs_user \
  -d afs_dev \
  | gzip > backup_$(date +%Y%m%d).sql.gz
```

**Restore:**
```bash
# From SQL file
docker-compose exec -T postgres psql \
  -U afs_user \
  -d afs_dev \
  < backup_20240115.sql

# From compressed
gunzip -c backup_20240115.sql.gz | \
  docker-compose exec -T postgres psql \
    -U afs_user \
    -d afs_dev
```

---

## 🧪 Testing Strategy

### Testing Pyramid

```
        ┌──────────┐
        │   E2E    │  Small (5%)
        └──────────┘
       ┌────────────┐
       │Integration │  Medium (25%)
       └────────────┘
      ┌──────────────┐
      │  Unit Tests  │  Large (70%)
      └──────────────┘
```

---

### Unit Testing (Backend)

**Location:** Same package as code (`_test.go` suffix)

**Running Tests:**

```bash
# All tests
cd backend
go test ./...

# Specific package
go test ./internal/domain/entities

# With coverage
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Verbose
go test -v ./...
```

**Test File Example (Conceptual):**
```
File: internal/domain/entities/task_test.go

Tests:
  - TestTask_Validate_ValidData
  - TestTask_Validate_EmptyQuery
  - TestTask_Validate_InvalidType
  - TestTask_CanTransitionTo_PendingToInProgress
  - TestTask_CanTransitionTo_InvalidTransition

Focus:
  - Business logic validation
  - State transitions
  - Edge cases
  - Error conditions
```

**Coverage Targets:**
- Domain layer: 90%+
- Use cases: 80%+
- Infrastructure: 60%+

---

### Integration Testing (Backend)

**Location:** `backend/tests/integration/`

**Setup:**
- Uses test database (Docker container)
- Runs migrations before tests
- Seeds test data
- Cleans up after tests

**Running:**
```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
cd backend
go test -tags=integration ./tests/integration/...

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

**Test Categories:**
- API endpoint tests (HTTP)
- Repository tests (Database)
- External service tests (LLM mocks)

---

### Unit Testing (Frontend)

**Framework:** Vitest

**Running Tests:**
```bash
cd frontend

# Run tests
npm run test

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage

# UI mode
npm run test:ui
```

**Test File Example:**
```
File: src/components/task/TaskCard.test.tsx

Tests:
  - Renders task information correctly
  - Shows correct status badge color
  - Truncates long queries
  - Calls onClick when clicked
  - Formats timestamps correctly

Tools:
  - Vitest (test runner)
  - React Testing Library (component testing)
  - MSW (API mocking)
```

---

### E2E Testing (Optional)

**Framework:** Playwright or Cypress

**Scenarios:**
- Complete task submission flow
- Real-time updates via WebSocket
- Multi-step user journeys

**Not in MVP:**
- Time-intensive to write
- Focus on unit + integration first
- Add later for critical paths

---

## ✅ Code Quality

### Linting

**Backend (Go):**

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Auto-fix where possible
golangci-lint run --fix

# Specific directories
golangci-lint run ./internal/...
```

**Configuration (.golangci.yml):**
```
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused

linters-settings:
  govet:
    check-shadowing: true
  gofmt:
    simplify: true

issues:
  max-same-issues: 0
```

---

**Frontend (ESLint + Prettier):**

```bash
cd frontend

# Lint
npm run lint

# Auto-fix
npm run lint:fix

# Format
npm run format

# Check formatting
npm run format:check
```

**Configuration (.eslintrc.json):**
```
{
  "extends": [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:@typescript-eslint/recommended",
    "prettier"
  ],
  "rules": {
    "max-len": ["error", { "code": 80 }],
    "max-lines": ["error", 300],
    "react/prop-types": "off"
  }
}
```

---

### Formatting

**Backend:**
```bash
# Format all Go files
gofmt -w .

# Or using goimports (preferred)
goimports -w .
```

**Frontend:**
```bash
# Format all files
npm run format

# Prettier config (.prettierrc)
{
  "semi": true,
  "trailingComma": "all",
  "singleQuote": false,
  "printWidth": 80,
  "tabWidth": 2
}
```

---

### Pre-commit Hooks

**Using Husky (Frontend):**

```bash
# Install
npm install --save-dev husky lint-staged

# Setup
npx husky install
```

**Configuration (package.json):**
```
{
  "lint-staged": {
    "*.{ts,tsx}": [
      "eslint --fix",
      "prettier --write"
    ],
    "*.go": [
      "gofmt -w",
      "golangci-lint run"
    ]
  }
}
```

**Pre-commit hook (.husky/pre-commit):**
```
#!/bin/sh
npm run lint-staged
```

**Benefits:**
- Automatic formatting before commit
- Prevents bad code from entering repo
- Consistent code style

---

### Code Review Checklist

**Architecture:**
- [ ] Correct layer for new code?
- [ ] Dependencies point inward?
- [ ] Interfaces used for cross-layer communication?

**SOLID Principles:**
- [ ] Single responsibility per function/file?
- [ ] No god objects?
- [ ] Dependencies injected?

**Code Quality:**
- [ ] Files under 300 lines?
- [ ] Lines under 80 characters?
- [ ] No magic numbers?
- [ ] Errors handled explicitly?
- [ ] No TODO/FIXME (use issues)?

**Testing:**
- [ ] Unit tests written?
- [ ] Coverage maintained/improved?
- [ ] Edge cases covered?

**Documentation:**
- [ ] Public functions documented?
- [ ] Complex logic explained?
- [ ] README updated if needed?

---

## 🐛 Debugging

### Backend Debugging

**Logging:**

**Structured logging:**
```
Using zerolog or zap:

log.Info().
  Str("agent", "gemini-2.5-pro").
  Int("task_id", 123).
  Msg("Agent started analysis")

Output:
{"level":"info","agent":"gemini-2.5-pro","task_id":123,
 "msg":"Agent started analysis","time":"2024-01-15T10:30:00Z"}
```

**Log Levels:**
- DEBUG: Detailed trace (development only)
- INFO: Normal operations
- WARN: Degraded but functioning
- ERROR: Failures requiring attention

**Viewing Logs:**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend

# Last 100 lines
docker-compose logs --tail=100 backend

# Since timestamp
docker-compose logs --since 2024-01-15T10:00:00 backend
```

---

**Delve Debugger:**

**Setup:**
```
Install delve:
go install github.com/go-delve/delve/cmd/dlv@latest
```

**Debug with breakpoints:**
```bash
# Run with debugger
dlv debug ./cmd/server/main.go

# Commands in dlv:
(dlv) break main.main
(dlv) break orchestrator.go:45
(dlv) continue
(dlv) next
(dlv) print taskID
(dlv) quit
```

**VS Code Debugging:**

**Configuration (.vscode/launch.json):**
```
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/server",
      "env": {
        "DATABASE_URL": "postgres://...",
        "ENV": "development"
      }
    }
  ]
}
```

**Usage:**
1. Set breakpoints in VS Code
2. Press F5 (Start Debugging)
3. Step through code (F10, F11)
4. Inspect variables in sidebar

---

### Frontend Debugging

**Browser DevTools:**

**React DevTools:**
- Install Chrome/Firefox extension
- Inspect component hierarchy
- View props and state
- Trigger re-renders

**Network Tab:**
- Monitor API requests
- Inspect request/response
- Check WebSocket messages
- Identify slow requests

**Console Logging:**
```
Development mode:
  console.log("Task data:", task)
  console.table(proposals)
  
Production:
  Remove or use debug flag
  console.debug (stripped in prod build)
```

---

**VS Code Debugging:**

**Configuration (.vscode/launch.json):**
```
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Frontend",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:3000",
      "webRoot": "${workspaceFolder}/frontend/src",
      "sourceMapPathOverrides": {
        "webpack:///src/*": "${webRoot}/*"
      }
    }
  ]
}
```

**Usage:**
1. Start frontend: `npm run dev`
2. Set breakpoints in VS Code
3. Press F5
4. Debug in VS Code (not browser)

---

### Database Debugging

**Query Performance:**

```sql
-- Enable timing
\timing on

-- Explain query
EXPLAIN ANALYZE
SELECT ...;

-- Slow query log
SELECT * FROM pg_stat_statements
WHERE mean_exec_time > 1000
ORDER BY mean_exec_time DESC;

-- Active queries
SELECT pid, query, state, query_start
FROM pg_stat_activity
WHERE state = 'active';
```

**Connection Issues:**

```bash
# Check if PostgreSQL accepting connections
docker-compose exec postgres pg_isready

# Check connection count
docker-compose exec postgres psql -U afs_user -d afs_dev \
  -c "SELECT count(*) FROM pg_stat_activity;"

# Kill idle connections
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND query_start < NOW() - INTERVAL '1 hour';
```

---

## 🔀 Git Workflow

### Branch Strategy

**Main Branches:**
- `main` - Production-ready code
- `develop` - Integration branch

**Feature Branches:**
- Pattern: `feature/{ticket-id}-{description}`
- Example: `feature/AFS-123-task-submission`

**Bugfix Branches:**
- Pattern: `bugfix/{ticket-id}-{description}`
- Example: `bugfix/AFS-456-websocket-disconnect`

**Hotfix Branches:**
- Pattern: `hotfix/{description}`
- Example: `hotfix/rate-limit-overflow`

---

### Commit Messages

**Format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- feat: New feature
- fix: Bug fix
- docs: Documentation
- style: Formatting
- refactor: Code restructuring
- test: Adding tests
- chore: Maintenance

**Examples:**
```
feat(task): add task submission form

Implement React form with validation for creating optimization
tasks. Includes real-time validation and error handling.

Closes #123

---

fix(websocket): reconnect on disconnect

Add exponential backoff retry logic for WebSocket disconnections.
Prevents infinite reconnection loops.

Fixes #456
```

---

### Pull Request Process

**Creating PR:**

1. **Create feature branch:**
   ```bash
   git checkout -b feature/AFS-123-task-submission
   ```

2. **Make changes and commit:**
   ```bash
   git add .
   git commit -m "feat(task): add submission form"
   ```

3. **Push to remote:**
   ```bash
   git push origin feature/AFS-123-task-submission
   ```

4. **Open PR on GitHub/GitLab:**
   - Title: `[AFS-123] Add task submission form`
   - Description: Link to ticket, screenshots, testing notes
   - Assign reviewers

---

**PR Checklist (Template):**
```
## Changes
- [ ] Describe main changes

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Code Quality
- [ ] Linter passes
- [ ] No console errors
- [ ] Files <300 lines
- [ ] Lines <80 chars

## Documentation
- [ ] README updated (if needed)
- [ ] Comments added for complex logic
- [ ] API docs updated (if API changes)

## Screenshots (if UI changes)
[Add screenshots here]
```

---

**Review Process:**

1. **Reviewer checks:**
   - Code quality (SOLID, Clean Architecture)
   - Tests coverage
   - Documentation
   - Security concerns

2. **Feedback:**
   - Request changes if needed
   - Suggest improvements
   - Approve when ready

3. **Merge:**
   - Squash commits (clean history)
   - Delete branch after merge
   - Verify CI/CD passes

---

## 🔧 Common Tasks

### Adding a New Endpoint

**Steps:**

1. **Define route (interfaces/http/router.go):**
   ```
   Conceptual:
   app.Post("/api/v1/tasks", taskHandler.Create)
   ```

2. **Create handler (interfaces/http/handlers/):**
   ```
   TaskHandler.Create method:
     - Parse request body
     - Validate
     - Call use case
     - Return response
   ```

3. **Implement use case (usecases/):**
   ```
   TaskService.CreateTask method:
     - Business logic
     - Call repository
     - Return result
   ```

4. **Add repository method (infrastructure/database/):**
   ```
   TaskRepository.Create method:
     - Execute SQL INSERT
     - Return entity
   ```

5. **Write tests:**
   - Unit test for handler
   - Unit test for use case
   - Integration test for endpoint

6. **Update API docs:**
   - Add to 08-API-SPECIFICATION.md
   - Update Postman collection

---

### Adding a New Component

**Steps:**

1. **Create component file:**
   ```
   frontend/src/components/task/NewComponent.tsx
   ```

2. **Define TypeScript types:**
   ```
   interface NewComponentProps {
     data: Task
     onClick: (id: number) => void
   }
   ```

3. **Implement component:**
   ```
   export function NewComponent({ data, onClick }: NewComponentProps)
   Component logic and JSX
   ```

4. **Add styles (Tailwind):**
   ```
   className="bg-white rounded-lg shadow p-4"
   ```

5. **Write tests:**
   ```
   NewComponent.test.tsx:
     - Render test
     - Props test
     - Event handler test
   ```

6. **Export from index:**
   ```
   components/task/index.ts:
   export { NewComponent } from './NewComponent'
   ```

---

### Adding a Database Migration

**Steps:**

1. **Create migration file:**
   ```bash
   cd backend/migrations
   touch 003_add_feature.sql
   ```

2. **Write Up section:**
   ```sql
   -- +migrate Up
   ALTER TABLE tasks ADD COLUMN new_field VARCHAR(255);
   CREATE INDEX idx_tasks_new_field ON tasks(new_field);
   ```

3. **Write Down section:**
   ```sql
   -- +migrate Down
   DROP INDEX idx_tasks_new_field;
   ALTER TABLE tasks DROP COLUMN new_field;
   ```

4. **Test migration:**
   ```bash
   # Apply
   migrate -path ./migrations -database "${DATABASE_URL}" up
   
   # Verify
   psql ${DATABASE_URL} -c "\d tasks"
   
   # Rollback test
   migrate -path ./migrations -database "${DATABASE_URL}" down 1
   
   # Re-apply
   migrate -path ./migrations -database "${DATABASE_URL}" up
   ```

5. **Update domain entities:**
   ```
   Add new field to Task struct
   Update validation logic
   ```

6. **Commit migration file:**
   ```bash
   git add migrations/003_add_feature.sql
   git commit -m "feat(db): add new_field to tasks table"
   ```

---

## 🔍 Troubleshooting

### Common Issues

**Issue: Port already in use**

**Symptoms:**
```
Error: bind: address already in use
```

**Solution:**
```bash
# Find process using port
lsof -i :8000
# or
netstat -vanp tcp | grep 8000

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "8001:8000"  # host:container
```

---

**Issue: Database connection refused**

**Symptoms:**
```
Error: connection refused
pq: could not connect to server
```

**Solution:**
```bash
# Check if PostgreSQL running
docker-compose ps postgres

# Restart PostgreSQL
docker-compose restart postgres

# Check logs
docker-compose logs postgres

# Verify connection
docker-compose exec postgres pg_isready
```

---

**Issue: Hot reload not working**

**Backend (Air):**
```bash
# Check Air is running
docker-compose logs backend | grep "Building"

# Verify .air.toml present
ls backend/.air.toml

# Restart backend
docker-compose restart backend
```

**Frontend (Vite):**
```bash
# Check Vite HMR connected
Browser console: "[vite] connected"

# Clear cache
rm -rf frontend/node_modules/.vite

# Restart frontend
docker-compose restart frontend
```

---

**Issue: Migration fails**

**Symptoms:**
```
Error: migration X failed
```

**Solution:**
```bash
# Check current version
docker-compose exec backend \\
  goose -dir ./migrations postgres "${DATABASE_URL}" status

# Force version (if stuck) - use with caution
docker-compose exec backend \\
  goose -dir ./migrations postgres set <version>

# Re-run migration
docker-compose exec backend \\
  goose -dir ./migrations postgres "${DATABASE_URL}" up

# Manual fix (last resort)
# 1. Manually apply SQL
# 2. Update the migration version table (goose stores applied versions in its table)
```

---

**Issue: Docker out of disk space**

**Symptoms:**
```
Error: no space left on device
```

**Solution:**
```bash
# Check disk usage
docker system df

# Clean up
docker system prune -a --volumes

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune
```

---

**Issue: Frontend blank page**

**Symptoms:**
- Browser shows blank page
- No errors in console
- Network requests succeed

**Solution:**
```bash
# Check browser console for errors
# Look for: Failed to fetch, CORS errors

# Verify API URL in .env
VITE_API_URL=http://localhost/api

# Clear browser cache
Hard refresh: Ctrl+Shift+R (Windows) / Cmd+Shift+R (Mac)

# Rebuild frontend
docker-compose build frontend
docker-compose up -d frontend
```

---

## 🎯 Summary

This development workflow provides:

**Environment Setup:**
- Prerequisites clearly defined
- Step-by-step initial setup
- IDE configuration for consistency

**Project Structure:**
- Clear directory organization
- Clean Architecture layers
- Feature-based component structure

**Development Process:**
- Docker Compose for orchestration
- Hot reload for rapid iteration
- Individual service development

**Database Management:**
- Migration workflow (up/down)
- Seeding for test data
- Backup and restore procedures

**Testing Strategy:**
- Testing pyramid (70% unit, 25% integration, 5% e2e)
- Coverage targets per layer
- Test commands and examples

**Code Quality:**
- Linting and formatting tools
- Pre-commit hooks
- Code review checklist

**Debugging:**
- Logging best practices
- Debugger setup (Delve, VS Code)
- Database query debugging

**Git Workflow:**
- Branch strategy
- Commit message format
- Pull request process

**Common Tasks:**
- Add endpoint (step-by-step)
- Add component (step-by-step)
- Create migration (step-by-step)

**Troubleshooting:**
- 6 common issues with solutions
- Quick diagnostics
- Recovery procedures

---

**Related Documentation:**
- Next: [11-DEPLOYMENT-STRATEGY.md](11-DEPLOYMENT-STRATEGY.md) 
  - Production deployment
- See also: [03-SYSTEM-ARCHITECTURE.md](03-SYSTEM-ARCHITECTURE.md) 
  - Architecture principles enforced
- See also: [02-DATA-MODEL.md](02-DATA-MODEL.md) 
  - Database schema for migrations

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead
```
