### 🎯 Plan Completo de Desarrollo AFS - Paso a Paso
📊 Visión General

Total: 45 conversaciones organizadas en 10 fases

Cada conversación incluye:
    ✅ Documentos exactos a compartir
    ✅ Códidocker compose exec backend go específico a crear
    ✅ Validación concreta
    ✅ Criterio de "listo para continuar"

# 🏗️ FASE 1: Setup Inicial y Base de Datos (3 conversaciones)

***Conversación #1: Estructura de Proyecto y Configuración***

Objetivo: Crear estructura de directorios, .env, docker-compose

Debes consultar estos doc para la tarea:
- 00-PROJECT-OVERVIEW.md (completo)
- 10-DEVELOPMENT-WORKFLOW.md (secciones: Project Structure, Environment Setup)

Prompt:
Contexto: Proyecto AFS - Sistema multi-agente para optimización de DB

Tarea: Setup inicial del proyecto

Crear:
1. Estructura de directorios completa (backend/ y frontend/)
2. Directorio `secrets/` para credenciales.
3. .env.example con todas las variables necesarias
4. docker-compose.yml con:
   - PostgreSQL 16
   - Redis 7
   - Backend (docker compose exec backend go con Air), configurado para montar el volumen de credenciales (`./secrets/application_default_credentials.json`) y usar la variable de entorno `GOOGLE_APPLICATION_CREDENTIALS`.
   - Frontend (React con Vite)
   - Caddy

Requisitos:
- Seguir estructura exacta de 10-DEVELOPMENT-WORKFLOW.md
- Health checks en todos los servicios
- Volúmenes para persistencia
- Variables de entorno según doc

Output:
- Árbol de directorios completo (comando mkdir -p)
- .env.example
- docker-compose.yml
- .gitignore (backend y frontend)

Validación:
# Crear estructura
[ejecutar comandos mkdir -p del output]
# Verificar docker-compose
docker-compose config
# Debe mostrar configuración sin errores

✅ Listo si: docker-compose config ejecuta sin errores

***Conversación #2: Migraciones Base de Datos (E-commerce)***

Debes consultar estos doc para la tarea:
- 00-PROJECT-OVERVIEW.md (sección Tech Stack)
- 02-DATA-MODEL.md (sección: Existing Tables)
- 10-DEVELOPMENT-WORKFLOW.md (sección: Database Management)

Tarea: Crear migration inicial para tablas e-commerce

Ubicación: backend/migrations/001_initial_schema.sql

Tablas a crear (según DATA-MODEL):
- users (id, email, created_at)
- orders (id, user_id, total, status, created_at)
- payments (id, order_id, amount, status, created_at)

Requisitos:
- Formato: -- +migrate Up / -- +migrate Down
- Foreign keys con CASCADE
- Índices PRIMARY KEY
- Down migration completa

Output:
- 001_initial_schema.sql completo

Validación:
# Iniciar servicios
docker-compose up -d postgres
# Aplicar migration
docker-compose exec postgres psql -U afs_user -d afs_dev \
  < backend/migrations/001_initial_schema.sql
# Verificar tablas
docker-compose exec postgres psql -U afs_user -d afs_dev -c "\dt"
# Debe mostrar: users, orders, payments

✅ Listo si: Las 3 tablas existen en la BD

*** Conversación #3: Migraciones AFS + Seeder ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (sección: New Tables AFS System)
- 10-DEVELOPMENT-WORKFLOW.md (sección: Seeding Data)

Tarea 1: Crear migration 002_afs_tables.sql

Tablas (según DATA-MODEL completo):
- tasks
- agent_executions
- optimization_proposals
- benchmark_results
- consensus_decisions

Incluir:
- Todos los campos de cada tabla
- Foreign keys correctos
- JSONB para metadata/scores
- Índices según doc 02

Tarea 2: Crear seeder

Ubicación: backend/scripts/seed/main.go

Funcionalidad:
- Conectar a BD
- Truncar tablas existentes
- Crear 1,000 users (gofakeit)
- Crear 10,000 orders
- Crear 10,000 payments
- Distribución: 75% completed, 12.5% pending, 12.5% processing

Output:
- 002_afs_tables.sql (Up y Down)
- backend/scripts/seed/main.go
- backend/go.mod (dependencias: gofakeit, sqlx, pq)

Validación:
# Aplicar migration
docker-compose exec postgres psql -U afs_user -d afs_dev \
  < backend/migrations/002_afs_tables.sql
# Verificar tablas AFS
docker-compose exec postgres psql -U afs_user -d afs_dev \
  -c "SELECT tablename FROM pg_tables WHERE schemaname='public';"
# Debe mostrar 8 tablas (3 + 5 nuevas)

# Ejecutar seeder
cd backend
docker compose exec backend go mod download
docker compose exec backend go run scripts/seed/main.go

# Verificar datos
docker-compose exec postgres psql -U afs_user -d afs_dev \
  -c "SELECT COUNT(*) FROM users; SELECT COUNT(*) FROM orders;"
# Debe mostrar: users=1000, orders=10000

✅ Listo si: 8 tablas existen y datos seeded correctamente

### 🎯 FASE 2: Domain Layer (6 conversaciones)

*** Conversación #4: Domain Entities - Task ***

Debes consultar estos doc para la tarea:
- 00-PROJECT-OVERVIEW.md (sección: Code Quality Standards)
- 02-DATA-MODEL.md (sección: Table tasks)
- 03-SYSTEM-ARCHITECTURE.md (sección: Layer 1 Domain)

Tarea: Crear entidad Task en Domain Layer

Ubicación: backend/internal/domain/entities/task.go

Estructura:
- Struct Task con todos los campos de tabla tasks
- Tipo TaskType (enum: query_optimization, schema_improvement, etc.)
- Tipo TaskStatus (enum: pending, in_progress, completed, failed)
- Método Validate() error
- Método CanTransitionTo(newStatus) bool
- Método IsComplete() bool

Requisitos:
- Zero dependencias externas (solo stdlib)
- Validación de business rules (query no vacío, type válido, etc.)
- Comentarios para funciones públicas
- Max 300 líneas

Output:
- task.go completo
- task_test.go con tests de validación y transiciones

Validación:
cd backend
# Ejecutar tests
docker compose exec backend go test ./internal/domain/entities -v
# Debe pasar todos los tests
# Coverage debe ser >80%

docker compose exec backend go test ./internal/domain/entities -cover

✅ Listo si: Tests pasan y coverage >80%

*** Conversación #5: Domain Entities - Agent, Proposal, Benchmark ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (secciones: agent_executions, optimization_proposals, benchmark_results)
- 03-SYSTEM-ARCHITECTURE.md (Layer 1 Domain)
- 04-AGENT-SYSTEM.md (sección: Agent Interface Contract)

Prompt:

Tarea: Crear entidades relacionadas

Ubicación: backend/internal/domain/entities/

Crear:

1. agent_execution.go
   - Struct AgentExecution
   - Enum AgentType (cerebro, operativo, bulk)
   - Enum ExecutionStatus (running, completed, failed)
   - Validaciones

2. optimization_proposal.go
   - Struct OptimizationProposal
   - Enum ProposalType (index, partitioning, materialized_view, etc.)
   - Struct EstimatedImpact (JSONB fields)
   - Validaciones (SQL commands no vacíos, etc.)

3. benchmark_result.go
   - Struct BenchmarkResult
   - Struct ExplainPlan (JSONB fields)
   - Validaciones (execution time positivo, etc.)

Requisitos:
- Zero dependencias externas
- Cada archivo max 300 líneas
- Tests en *_test.go

Output:
- agent_execution.go + test
- optimization_proposal.go + test
- benchmark_result.go + test

Validación:
docker compose exec backend go test ./internal/domain/entities/... -v -cover

# Todos los tests deben pasar
# Coverage >80% en cada archivo

✅ Listo si: Tests pasan, coverage >80%

*** Conversación #6: Domain Entities - Consensus ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (sección: consensus_decisions)
- 03-SYSTEM-ARCHITECTURE.md (Layer 1)
- 05-CONSENSUS-BENCHMARKING.md (sección: Score Calculation)

Tarea: Crear entidad ConsensusDecision

Ubicación: backend/internal/domain/entities/consensus_decision.go
Incluir:
- Struct ConsensusDecision
- Struct ProposalScore (performance, storage, complexity, risk, weighted_total)
- Struct ScoringCriteria (weights configurables)
- Método CalculateWeightedTotal(scores) float64
- Validaciones (weights suman 1.0, etc.)

Output:
- consensus_decision.go
- consensus_decision_test.go (tests de cálculo de scores)

Validación:
docker compose exec backend go test ./internal/domain/entities/... -v -cover

# Tests de scoring deben pasar
# Verificar fórmula: (perf*0.5) + (storage*0.2) + (complexity*0.2) + (risk*0.1)

✅ Listo si: Cálculos de scoring correctos

*** Conversación #7: Domain Interfaces - Repositories ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (todas las tablas)
- 03-SYSTEM-ARCHITECTURE.md (Dependency Inversion, Repository Pattern)

Tarea: Definir interfaces de repositorios

Ubicación: backend/internal/domain/interfaces/repositories.go
Interfaces a crear:
type TaskRepository interface {
    Create(ctx context.Context, task *Task) error
    GetByID(ctx context.Context, id int) (*Task, error)
    List(ctx context.Context, filters TaskFilters) ([]*Task, error)
    Update(ctx context.Context, task *Task) error
}

type AgentExecutionRepository interface {
    Create(ctx context.Context, exec *AgentExecution) error
    GetByID(ctx context.Context, id int) (*AgentExecution, error)
    GetByTaskID(ctx context.Context, taskID int) ([]*AgentExecution, error)
    Update(ctx context.Context, exec *AgentExecution) error
}

(Similar para: OptimizationRepository, BenchmarkRepository, ConsensusRepository)

Requisitos:
- Solo interfaces (sin implementación)
- Context como primer parámetro
- Error handling
- Filters structs para List operations

Output:
- repositories.go con todas las interfaces

Validación:
# Verificar compilación
docker compose exec backend go build ./internal/domain/interfaces
# No debe haber errores (solo interfaces, no ejecuta)

✅ Listo si: Compila sin errores

*** Conversación #8: Domain Values - Enums y Constants ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (enums documentados)
- 03-SYSTEM-ARCHITECTURE.md (Layer 1)

Tarea: Centralizar enums y constantes

Ubicación: backend/internal/domain/values/
Crear:
1. task_status.go
   - Const para TaskStatus (Pending, InProgress, Completed, Failed)
   - Función IsValid(status) bool
   - Función String() string

2. agent_type.go
   - Const para AgentType (cerebro, operativo, bulk)
   - Función GetSpecialization(agentType) AgentSpecialization

3. proposal_type.go
   - Const para ProposalType (Index, Partitioning, etc.)

Output:
- task_status.go
- agent_type.go  
- proposal_type.go
- Cada uno con tests de validación

Validación:
docker compose exec backend go test ./internal/domain/values/... -v
# Tests de enums deben pasar
✅ Listo si: Tests pasan

*** Conversación #9: Config Layer ***

Debes consultar estos doc para la tarea:
- 03-SYSTEM-ARCHITECTURE.md (Layer 5: Configuration)
- 10-DEVELOPMENT-WORKFLOW.md (Environment Variables)

Tarea: Implementar Configuration Layer
Ubicación: backend/internal/config/
Crear:
1. config.go
   - Struct Config con todas las secciones:
     * Server (Port, Host, Environment, LogLevel)
     * Database (URL, MaxConnections)
     * Redis (URL, Password)
     * Vertex AI: ProjectID, Location, Model IDs, y la ruta al archivo de credenciales (GOOGLE_APPLICATION_CREDENTIALS) para la autenticación ADC.
     * Tiger Cloud (UseTigerCloud, MainService, MCP URL)
   - Función Load() (*Config, error) que lee de env vars
   - Validación de campos requeridos

2. tiger.go
   - Struct TigerConfig
   - Función para leer ~/.config/tiger/mcp-config.json

Output:
- config.go
- tiger.go
- config_test.go (test de validación)

Validación:
# Test con env vars
export VERTEX_PROJECT_ID=test
export POSTGRES_DB=test
docker compose exec backend go test ./internal/config -v
# Debe pasar validación
✅ Listo si: Config carga y valida correctamente

### ⚙️ FASE 3: Infrastructure Layer (10 conversaciones)

*** Conversación #10: MCP Client - Base ***

Debes consultar estos doc para la tarea:
- 06-TIGER-CLOUD-MCP.md (secciones: MCP Protocol, Request/Response)
- 03-SYSTEM-ARCHITECTURE.md (Infrastructure Layer)
Tarea: Implementar MCP Client base

Ubicación: backend/internal/infrastructure/mcp/client.go
Funcionalidad:
- Struct MCPClient con http.Client y config
- Método Connect() error (test conexión)
- Método ExecuteQuery(serviceID, sql) (QueryResult, error)
- Método Close() error
- Retry logic con exponential backoff (3 attempts)
- Timeout handling

Request format según doc 06
Response parsing

Output:
- client.go
- client_test.go (con mocks)

Validación:
# Test unitario (con mock HTTP)
docker compose exec backend go test ./internal/infrastructure/mcp -v -run TestMCPClient
# Debe pasar sin llamar API real
✅ Listo si: Tests con mocks pasan

*** Conversación #11: MCP Client - Service Management***

Debes consultar estos doc para la tarea:
- 06-TIGER-CLOUD-MCP.md (secciones: Fork Operations, Service Management)

Tarea: Agregar operaciones de fork management
Ubicación: backend/internal/infrastructure/mcp/service.go
Métodos:
- CreateFork(parent, name) (forkID, error)
- DeleteFork(serviceID) error
- ListForks(parent) ([]ServiceInfo, error)
- GetServiceInfo(serviceID) (ServiceInfo, error)
Naming convention: afs-fork-{agent}-task{id}-{timestamp}
Output:
- service.go
- service_test.go

Validación:
docker compose exec backend go test ./internal/infrastructure/mcp -v -run TestFork
# Tests con mock deben pasar
✅ Listo si: Fork operations mockeadas funcionan

*** Conversación #12: LLM Client - Vertex AI (Interface y Modelos) ***

Debes consultar estos doc para la tarea:
- 07-LLM-INTEGRATION.md (secciones: Cliente Unificado Vertex, Modelos)
- 03-SYSTEM-ARCHITECTURE.md (Infrastructure)
Tarea: Implementar LLM Client unificado (Vertex) con selección de modelo
Ubicación: backend/internal/infrastructure/llm/
Crear:
1. client.go (interface)
   type LLMClient interface {
       SendMessage(prompt, system string) (string, error)
       SendMessageWithJSON(prompt, system string) (map[string]interface{}, error)
       GetUsage() (inputTokens, outputTokens int)
   }
2. vertex_client.go
   - Implementa LLMClient
   - Selección de modelo: gemini-2.5-pro | gemini-2.5-flash | gemini-2.0-flash
   - JSON parsing con markdown fence removal
   - Error handling según doc 07
Output:
- client.go (interface)
- vertex_client.go (implementation)
- vertex_client_test.go (con mock HTTP)

Validación:
# Test unitario (con mock HTTP)
docker compose exec backend go test ./internal/infrastructure/llm -v -run TestVertexClient
# Mock debe simular una respuesta de Vertex AI
# JSON parsing debe extraer correctamente
✅ Listo si: Tests pasan, JSON parsing funciona

*** Conversación #13: LLM Client - Modelos Vertex adicionales ***

Debes consultar estos doc para la tarea:
- 07-LLM-INTEGRATION.md (secciones: Modelos soportados en Vertex)

Tarea: Añadir soporte de modelos adicionales en VertexClient
Ubicación: backend/internal/infrastructure/llm/
Crear:
1. Extender vertex_client.go para modelos:
   - gemini-2.5-pro
   - gemini-2.5-flash
   - gemini-2.0-flash

Output:
- vertex_client_test.go (tests por modelo)

Validación:
docker compose exec backend go test ./internal/infrastructure/llm/... -v
# VertexClient debe pasar tests mockeados para los 3 modelos
✅ Listo si: 3 clients funcionan con mocks

*** Conversación #14: Database - Repository Base ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (todas las tablas)
- 03-SYSTEM-ARCHITECTURE.md (Repository Pattern)

Tarea: Implementar TaskRepository
Ubicación: backend/internal/infrastructure/database/repositories/task_repository.go
Implementación:
- Struct PostgresTaskRepository con *sqlx.DB
- Implementar TaskRepository interface
- CRUD operations con SQL queries
- Error handling
- Context propagation
Output:
- task_repository.go
- task_repository_test.go (con test DB en Docker)

Validación:
# Iniciar test DB
docker-compose up -d postgres
# Aplicar migrations
docker-compose exec postgres psql -U afs_user -d afs_dev < backend/migrations/001_initial_schema.sql
docker-compose exec postgres psql -U afs_user -d afs_dev < backend/migrations/002_afs_tables.sql
# Run tests
docker compose exec backend go test ./internal/infrastructure/database/repositories -v -run TestTaskRepository
# Debe crear/leer/actualizar tasks en DB real
✅ Listo si: CRUD operations funcionan en DB real

*** Conversación #15: Database - Resto de Repositories ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (tablas relacionadas)

Tarea: Implementar repositories restantes
Ubicación: backend/internal/infrastructure/database/repositories/
Crear:
- agent_execution_repository.go
- optimization_repository.go
- benchmark_repository.go
- consensus_repository.go
Cada uno implementa su interface del domain
SQL queries según schema en DATA-MODEL
Output:
- 4 archivos _repository.go
- 4 archivos _repository_test.go

Validación:
docker compose exec backend go test ./internal/infrastructure/database/repositories/... -v
# Todos los repositories deben pasar tests con DB real
✅ Listo si: 5 repositories funcionan

*** Conversación #16: Agent Base Implementation ***

Debes consultar estos doc para la tarea:
- 04-AGENT-SYSTEM.md (secciones: Agent Interface, BaseAgent)
- 03-SYSTEM-ARCHITECTURE.md (Agents Infrastructure)

Tarea: Implementar BaseAgent (shared logic)
Ubicación: backend/internal/infrastructure/agents/base.go
Funcionalidad:
- Struct BaseAgent con MCPClient, LLMClient, Config
- Método CreateFork(taskID) (forkID, error)
  * Genera nombre: afs-fork-{agent}-task{id}-{timestamp}
  * Llama MCP CreateFork
  * Registra en AgentExecutionRepository
- Método DestroyFork(forkID) error
- Logging helpers
- Error handling helpers
Output:
- base.go
- base_test.go

Validación:
docker compose exec backend go test ./internal/infrastructure/agents -v -run TestBase
# Mock MCP client
# Verificar fork naming convention
✅ Listo si: BaseAgent funciona con mocks

*** Conversación #17: Agent Implementation (gemini-2.5-pro) ***

Debes consultar estos doc para la tarea:
- 04-AGENT-SYSTEM.md (sección: gemini-2.5-pro, Prompt Templates)
- 01-BUSINESS-LOGIC.md (gemini-2.5-pro Execution)

Tarea: Implementar Agent para gemini-2.5-pro
Ubicación: backend/internal/infrastructure/agents/gemini25pro_agent.go
Implementar Agent interface:
1. AnalyzeTask(task, forkID) (AnalysisResult, error)
   - Ejecuta EXPLAIN ANALYZE en fork (via MCP)
   - Construye prompt con contexto (EXPLAIN + schema)
   - Llama VertexClient con modelo gemini-2.5-pro
   - Parsea JSON response
   - Retorna AnalysisResult
2. ProposeOptimization(analysis, forkID) (OptimizationProposal, error)
   - Prompt para generar SQL (índices típicamente)
   - Valida SQL generado
   - Estima impacto
   - Retorna Proposal
3. RunBenchmark(proposal, forkID) ([]BenchmarkResult, error)
   - Define 4 test queries (baseline, limit, filter, sort)
   - Ejecuta cada query 3 veces
   - Calcula promedios
   - Mide storage impact
   - Retorna results
Prompts según doc 04 (templates específicos de gemini-2.5-pro)
Output:
- gemini25pro_agent.go (max 300 líneas, dividir si necesario)
- gemini25pro_agent_test.go
Validación:
# Test con mocks (LLM y MCP mockeados)
docker compose exec backend go test ./internal/infrastructure/agents -v -run TestAgentGemini25Pro
# Verificar:
# - Prompt construction correcta
# - JSON parsing funciona
# - Benchmark suite ejecuta 4 queries
✅ Listo si: Agent gemini-2.5-pro funciona end-to-end con mocks

*** Conversación #18: Agents (gemini-2.5-flash y gemini-2.0-flash) ***

Debes consultar estos doc para la tarea:
- 04-AGENT-SYSTEM.md (secciones: gemini-2.5-flash y gemini-2.0-flash)
Tarea: Implementar Agents para gemini-2.5-flash y gemini-2.0-flash
Ubicación: backend/internal/infrastructure/agents/
Siguiendo mismo patrón que Cerebro pero con:
- Prompts específicos de cada modelo (doc 04)
- Especialización diferente:
  * gemini-2.5-flash: Partitioning, schema redesign, ejecución rápida
  * gemini-2.0-flash: Materialized views, tareas masivas de bajo riesgo
Output:
- gemini25flash_agent.go + test
- gemini20flash_agent.go + test
Validación:
docker compose exec backend go test ./internal/infrastructure/agents/... -v
# 3 agents (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash) deben pasar tests con mocks
✅ Listo si: 3 agents completos y testeados

*** Conversación #19: Agent Factory ***

Debes consultar estos doc para la tarea:
- 03-SYSTEM-ARCHITECTURE.md (Factory Pattern)
- 04-AGENT-SYSTEM.md (Agent Types)

Tarea: Crear Agent Factory
Ubicación: backend/internal/infrastructure/agents/factory.go
Funcionalidad:
- Función NewAgent(agentType, mcpClient, llmClient, config) (Agent, error)
- Switch por AgentType (gemini25pro, gemini25flash, gemini20flash)
- Inyección de dependencias
- Error si agentType inválido
Output:
- factory.go
- factory_test.go (verificar todos los tipos)

Validación:
docker compose exec backend go test ./internal/infrastructure/agents -v -run TestFactory
# Debe crear instancias de gemini25pro, gemini25flash, gemini20flash
✅ Listo si: Factory crea correctamente los 3 tipos

### 🎮 FASE 4: Use Cases Layer (8 conversaciones)

*** Conversación #20: Task Service ***

Debes consultar estos doc para la tarea:
- 01-BUSINESS-LOGIC.md (Task Lifecycle)
- 03-SYSTEM-ARCHITECTURE.md (Use Cases Layer)

Tarea: Implementar TaskService
Ubicación: backend/internal/usecases/task_service.go
Métodos:
- CreateTask(task) (Task, error)
  * Valida task
  * Persiste en repository
  * Retorna task con ID
- GetTask(id) (Task, error)
- ListTasks(filters) ([]Task, error)
- UpdateTaskStatus(id, status) error
Requisitos:
- Validación de business rules
- Context propagation
- Error handling
Output:
- task_service.go
- task_service_test.go (con mock repository)

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestTaskService
# Tests con mock repository deben pasar
✅ Listo si: CRUD básico de tasks funciona

*** Conversación #21: Task Router ***

Debes consultar estos doc para la tarea:
- 01-BUSINESS-LOGIC.md (Task Routing)
- 04-AGENT-SYSTEM.md (Router, Routing Algorithm)

Tarea: Implementar Task Router
Ubicación: backend/internal/usecases/router.go
Funcionalidad:
- Struct Router con AgentFactory
- Método SelectAgents(task) ([]Agent, error)
  * Calcula complexity score (según doc 04)
  * Aplica routing rules (prioridad, features, tamaño tabla)
  * Retorna lista de agents
  * Genera rationale
Reglas según doc 04 (solo Gemini):
- High priority → asignar gemini-2.5-pro, gemini-2.5-flash y gemini-2.0-flash
- JOINs → incluir gemini-2.5-flash
- Table >1M rows → incluir gemini-2.5-flash
- Aggregations complejas → incluir gemini-2.5-pro
Output:
- router.go
- router_test.go (test cada regla)

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestRouter
# Test scenarios:
# - Simple query → 1 agent (gemini-2.5-flash)
# - JOIN query → 2 agents (gemini-2.5-flash + gemini-2.5-pro)
# - High priority → 3 agents (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
✅ Listo si: Routing rules funcionan correctamente

*** Conversación #22: Benchmark Runner ***

Debes consultar estos doc para la tarea:
- 05-CONSENSUS-BENCHMARKING.md (Benchmark Runner section completa)
- 01-BUSINESS-LOGIC.md (Agent Workflow - Benchmarking)

Tarea: Implementar BenchmarkRunner (orquestador)
Ubicación: backend/internal/usecases/benchmark_runner.go
Funcionalidad:
- Método EvaluateProposal(proposal, forkID) ([]BenchmarkResult, error)
  * Define benchmark suite (4 queries según doc 05)
  * Ejecuta baseline (antes de aplicar proposal)
  * Aplica proposal SQL en fork
  * Ejecuta queries optimizadas (3 veces cada una)
  * Calcula promedios
  * Mide storage impact
  * Parsea EXPLAIN plans
  * Retorna array de results
Suite según doc 05:
- Test 1: Original query
- Test 2: Con LIMIT 10
- Test 3: Con filtro adicional
- Test 4: Con ORDER BY
Output:
- benchmark_runner.go
- benchmark_runner_test.go

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestBenchmarkRunner
# Mock MCP client
# Verificar 4 queries ejecutadas
# Verificar cálculo de improvement %
✅ Listo si: Benchmark suite ejecuta correctamente

*** Conversación #23: Consensus Engine ***

Debes consultar estos doc para la tarea:
- 05-CONSENSUS-BENCHMARKING.md (Consensus Engine, Scoring Algorithm completo)
- 01-BUSINESS-LOGIC.md (Consensus Decision)

Tarea: Implementar Consensus Engine
Ubicación: backend/internal/usecases/consensus_engine.go
Funcionalidad:
- Método Decide(proposals, benchmarks, criteria) (ConsensusDecision, error)
Algoritmo según doc 05:
1. Por cada proposal:
   - CalculatePerformanceScore (0-100 según improvement %)
   - CalculateStorageScore (0-100 según overhead MB)
   - CalculateComplexityScore (0-100 según proposal type)
   - CalculateRiskScore (0-100 según risk level) 
2. Calcular weighted_total:
   (perf × 0.5) + (storage × 0.2) + (complexity × 0.2) + (risk × 0.1)
3. Ordenar por weighted_total DESC
4. Seleccionar winner (rank 1)
5. Generar rationale (template doc 05)
Tie-breaking según doc 05
Output:
- consensus_engine.go
- consensus_engine_test.go (con data de ejemplo doc 01)

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestConsensus
# Test case ejemplo del doc 01:
# cerebro: 93.0 pts (winner)
# gemini25pro: 78.5 pts
# bulk: 66.5 pts
# Verificar fórmulas correctas
✅ Listo si: Scoring y ranking funcionan según ejemplo

*** Conversación #24: Orchestrator - Parte 1 (Agent Execution) ***

Debes consultar estos doc para la tarea:
- 01-BUSINESS-LOGIC.md (Complete User Flow, steps 3-4)
- 03-SYSTEM-ARCHITECTURE.md (Orchestrator)
- 04-AGENT-SYSTEM.md (Parallel Coordination)

Tarea: Implementar Orchestrator - Fase 1 (Parallel Agent Execution)
Ubicación: backend/internal/usecases/orchestrator.go
Struct Orchestrator con dependencias:
- Router
- AgentFactory
- Repositories (todos)
- MCPClient
- Config
Método ExecuteAgentsInParallel(task, agents) ([]Proposal, []Benchmark, error):
1. Crear WaitGroup
2. Crear channels para results y errors
3. Por cada agent:
   - Spawn goroutine
   - Dentro goroutine:
     * CreateFork
     * AnalyzeTask
     * ProposeOptimization
     * RunBenchmark
     * Enviar results a channel
4. Wait for all
5. Recopilar results
6. Manejar errores parciales (1 de 3 falla = ok)
Timeout: 10 min por agent
Output:
- orchestrator.go (primera versión, solo ejecución paralela)
- orchestrator_test.go (con mocks)

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestOrchestratorParallel
# Mock agents
# Verificar goroutines se ejecutan
# Verificar WaitGroup funciona
# Test error parcial (1 agent falla, 2 continúan)
✅ Listo si: Ejecución paralela funciona

*** Conversación #25: Orchestrator - Parte 2 (Consensus y Apply) ***

Debes consultar estos doc para la tarea:
- 01-BUSINESS-LOGIC.md (steps 5-6: Consensus y Apply)
- 05-CONSENSUS-BENCHMARKING.md (Apply to Main Database)

Tarea: Completar Orchestrator con Consensus y Apply
En orchestrator.go agregar:
1. Método ApplyToMainDB(winningProposal) error:
   - Pre-validation checks
   - Execute SQL en main DB (via MCP)
   - Post-application validation
   - Record PITR timestamp
   - Update consensus.applied_to_main
2. Método CleanupForks(forkIDs) error:
   - Por cada fork, llamar MCP DeleteFork
   - Log cleanup
3. Método ExecuteTask(taskID) error (método principal):
   - Load task
   - Call Router
   - ExecuteAgentsInParallel
   - Call ConsensusEngine
   - ApplyToMainDB
   - CleanupForks
   - Update task status
Error handling en cada paso
Output:
- orchestrator.go (versión completa)
- Agregar tests de flujo completo

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestOrchestratorComplete
# Test end-to-end con todos los mocks
# Verificar flujo completo: router → agents → consensus → apply → cleanup

✅ Listo si: Flujo completo funciona end-to-end

*** Conversación #26: WebSocket Event Broadcaster ***

Debes consultar estos doc para la tarea:
- 08-API-SPECIFICATION.md (WebSocket API)
- 03-SYSTEM-ARCHITECTURE.md (Observer Pattern)

Tarea: Implementar WebSocket Hub
Ubicación: backend/internal/usecases/websocket_hub.go
Funcionalidad:
- Struct Hub con:
  * clients map
  * broadcast channel
  * register/unregister channels
- Método Run() (goroutine principal)
- Método Broadcast(event) (emitir a todos los clients)
- Event types según doc 08
Integración con Orchestrator:
- Orchestrator llama hub.Broadcast en cada paso
Output:
- websocket_hub.go
- websocket_hub_test.go

Validación:
docker compose exec backend go test ./internal/usecases -v -run TestWebSocketHub
# Test registro de clients
# Test broadcast a múltiples clients
✅ Listo si: Hub broadcast funciona
