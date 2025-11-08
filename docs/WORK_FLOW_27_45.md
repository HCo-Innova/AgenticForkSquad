*** Conversación #27: Main Entry Point (Dependency Wiring) ***

Debes consultar estos doc para la tarea:
- 03-SYSTEM-ARCHITECTURE.md (Dependency Injection, Main Package)
- 10-DEVELOPMENT-WORKFLOW.md (Project Structure, Environment)

Tarea: Crear `main.go` con el wiring completo de dependencias

Ubicación: `backend/cmd/server/main.go`

Alcance/Funcionalidad:
- Cargar configuración (Config Layer)
- Conectar a base de datos
- Ejecutar migraciones si `RUN_MIGRATIONS=true`
- Inicializar Repositories (Task, AgentExecution, Optimization, Benchmark, Consensus)
- Inicializar MCP Client (Tiger Cloud MCP)
- Inicializar LLM Client unificado (Vertex) con modelos: gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash
- Inicializar Agent Implementations y `AgentFactory`
- Inicializar WebSocket Hub (Broadcaster) para tiempo real
- Inicializar Use Cases: `TaskService`, `Router`, `BenchmarkRunner`, `ConsensusEngine`, `Orchestrator`
- Inicializar capa Interfaces: Handlers HTTP y WebSocket
- Configurar router (Fiber v2) y middleware (CORS, logging, error handler)
- Iniciar servidor HTTP/WebSocket

Output:
- `backend/cmd/server/main.go` con wiring completo y comentarios mínimos

Validación:
- La app arranca sin errores y los health checks responden OK
- MCP: Con USE_TIGER_CLOUD=true y variables de MCP configuradas, el Connect() no debe fallar.
- LLM: Requiere VERTEX_PROJECT_ID, VERTEX_LOCATION y variables: GEMINI_CEREBRO_MODEL, GEMINI_OPERATIVO_MODEL, GEMINI_BULK_MODEL
- ✅ Listo si: la API responde 200 en `/health` y el Hub WebSocket inicia sin errores

---

*** Conversación #28: Main Entry Point - Finalización y Servidor ***

Debes consultar estos doc para la tarea:
- 03-SYSTEM-ARCHITECTURE.md (Interfaces Layer, HTTP/WebSocket)
- 08-API-SPECIFICATION.md (Rutas REST y WS)

Tarea: Completar inicialización de Handlers y servidor

Ubicación: `backend/cmd/server/main.go`

Alcance/Funcionalidad:
- Instanciar y registrar handlers HTTP (tasks, agents, proposals, consensus, health)
- Mapear endpoints según 08-API-SPECIFICATION.md
- Integrar Hub WebSocket en el router (endpoint `/ws`)
- Iniciar servidor y manejar señales de apagado

Output:
- `main.go` consolidado con rutas y WS registrados

Validación:
- Endpoints clave responden (GET `/health`, POST `/tasks` en entorno local)
- ✅ Listo si: router y WS quedan operativos y accesibles

---

### 🌐 FASE 5: Interfaces Layer (API & Handlers)

*** Conversación #29: HTTP Handlers - Task Management ***

Debes consultar estos doc para la tarea:
- 08-API-SPECIFICATION.md (Tasks endpoints)
- 03-SYSTEM-ARCHITECTURE.md (Interfaces → Use Cases)

Tarea: Implementar handlers REST de tareas

Ubicación:
- `backend/internal/interfaces/http/handlers/task_handler.go`
- `backend/internal/interfaces/http/router.go`

Alcance/Funcionalidad:
- POST `/tasks` (crear tarea)
- GET `/tasks/{id}` (detalle)
- GET `/tasks` (listar con filtros/paginación)
- Validar DTOs, invocar `TaskService`, mapear a respuestas

Output:
- `task_handler.go`, actualización de `router.go`

Validación:
- Respuestas HTTP correctas (201/200), validación de entrada
- Listo si: alta, consulta y listado de tareas funcionan

---

✅ *** Conversación #30: HTTP Handlers - Resultados y Salud ***

Debes consultar estos doc para la tarea:
- 08-API-SPECIFICATION.md (Agents/Proposals/Consensus/Health)

Tarea: Implementar handlers restantes

Ubicación: `backend/internal/interfaces/http/handlers/`

Alcance/Funcionalidad:
- GET `/tasks/{id}/agents`
- GET `/tasks/{id}/proposals`
- GET `/proposals/{id}/benchmarks`
- GET `/tasks/{id}/consensus`
- GET `/health`

Output:
- Handlers y registro en `router.go`

Validación:
- Todos los endpoints devuelven datos esperados
- Listo si: 100% de endpoints de 08 están operativos

 ✅ *** Conversación #31: WebSocket Handlers y Eventos ***

Debes consultar estos doc para la tarea:
- 08-API-SPECIFICATION.md (WebSocket API y eventos)
- 03-SYSTEM-ARCHITECTURE.md (Observer Pattern)

Tarea: Implementar capa WebSocket (server-side)

Ubicación: `backend/internal/interfaces/websocket/`

Alcance/Funcionalidad:
- `hub.go`, `client.go`, `events.go` para registrar/broadcast
- Manejo de eventos: `task_created`, `agents_assigned`, `fork_created`, `analysis_completed`, `proposal_submitted`, `benchmark_completed`, `consensus_reached`, `optimization_applied`, `task_completed`, `task_failed`
- Mensajes opcionales cliente→servidor: `ping`, `subscribe`

Output:
- Archivos WS implementados e integrados con Orchestrator

Validación:
- Conexión WS estable y eventos recibidos en tiempo real
- Listo si: múltiples clientes reciben broadcast del Hub

---

### 🎨 FASE 6: Frontend & UI

✅  *** Conversación #32: Frontend - Estructura, Hooks y Rutas ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md (Arquitectura y hooks)
- 08-API-SPECIFICATION.md (contratos de datos)

Tarea: Base del frontend

Ubicación: `frontend/src/`

Alcance/Funcionalidad:
- Estructura de directorios (components, hooks, services, pages)
- Hooks: `useTasks`, `useAgents`, `useOptimizations`, `useWebSocket`
- Rutas: Home, Tasks, Task Detail, Agents

Output:
- Árbol base, hooks y rutas iniciales

Validación:
- Navegación funcional y datos básicos cargan con React Query
- Listo si: SPA navega entre páginas sin errores

---

✅ *** Conversación #33: Frontend - Task Submission UI ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md (TaskSubmission)
- 08-API-SPECIFICATION.md (POST /tasks)

Tarea: Implementar formulario de creación de tareas

Ubicación: `frontend/src/pages/TaskSubmissionPage.tsx`

Alcance/Funcionalidad:
- Validación de campos, estados de carga/errores
- Envío al endpoint `/tasks` y redirección a detalle

Output:
- Página y componentes asociados

Validación:
- Creación exitosa y feedback visual correcto
- Listo si: se crea y redirige a la vista de detalle

---

✅ *** Conversación #34: Frontend - Task List y Estado ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md (TaskList/TaskCard)

Tarea: Lista de tareas con estado en tiempo real

Ubicación: `frontend/src/pages/TaskListPage.tsx`

Alcance/Funcionalidad:
- Listado con filtros/paginación
- Badges de estado (Pending, In Progress, Completed, Failed)

Output:
- Componentes de lista y tarjeta

Validación:
- Lista reactiva y filtros operativos
- Listo si: estados y filtros se reflejan en UI

---

✅ *** Conversación #35: Frontend - Task Detail y Timeline de Agentes ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md (TaskDetail, AgentStatus)
- 08-API-SPECIFICATION.md (WebSocket events)

Tarea: Detalle de tarea con actualizaciones en tiempo real

Ubicación: `frontend/src/pages/TaskDetailPage.tsx`

Alcance/Funcionalidad:
- Suscripción WS por `task_id`
- Timeline de eventos y estado por agente

Output:
- Página de detalle con timeline

Validación:
- Eventos WS actualizan UI en caliente
- Listo si: se visualiza el avance por agente en tiempo real

---

✅ *** Conversación #36: Frontend - Proposal Comparison Dashboard ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md (Proposals/Consensus/Charts)

Tarea: Comparación y visualización de resultados

Ubicación: `frontend/src/components/optimization/`

Alcance/Funcionalidad:
- Tabla comparativa (mejora %, overhead, scores)
- Gráficos de benchmarks y breakdown de scoring

Output:
- Componentes de comparación y gráficos

Validación:
- Datos consistentes con API y lectura clara
- Listo si: se muestran comparativas y puntajes correctamente

---

### ☁️ FASE 7: Tiger Cloud Migration y PITR

✅ *** Conversación #37: Tiger Cloud Migration - Configuración ***

Debes consultar estos doc para la tarea:
- 11-DEPLOYMENT-STRATEGY.md (Tiger Cloud Setup)
- 06-TIGER-CLOUD-MCP.md (MCP config)

Tarea: Migrar configuración a Tiger Cloud

Ubicación: Configuración de runtime y arranque

Alcance/Funcionalidad:
- Variables: `USE_TIGER_CLOUD=true`, `TIGER_MAIN_SERVICE`, `TIGER_MCP_URL`
- Conexión a Tiger DB y autenticación MCP

Output:
- Variables y config validadas en entorno

Validación:
- Backend inicia contra Tiger y obtiene esquema
- Listo si: health OK y conexión MCP estable

---

✅ *** Conversación #38: Tiger Cloud - Fork Lifecycle y Rollback PITR ***

Debes consultar estos doc para la tarea:
- 06-TIGER-CLOUD-MCP.md (Forks y PITR)
- 05-CONSENSUS-BENCHMARKING.md (Apply & PITR timestamp)

Tarea: Validar forks zero-copy y rollback PITR

Ubicación: Orchestrator y capa MCP

Alcance/Funcionalidad:
- Crear/usar/eliminar forks (<10s en 1GB)
- Registrar timestamp PITR antes de aplicar
- Crear fork desde timestamp previo (rollback test)

Output:
- Evidencia de tiempos y rollback exitoso

Validación:
- Medidas dentro de límites y PITR efectivo
- Listo si: creación de forks <10s y rollback funcional

---

✅ ### 🧪 FASE 8: Validación Final y Pulido

*** Conversación #39: System Validation - End-to-End Test ***

Debes consultar estos doc para la tarea:
- 01-BUSINESS-LOGIC.md (flujo E2E)
- 05-CONSENSUS-BENCHMARKING.md (criterios)

Tarea: Ejecutar validación E2E completa

Ubicación: Suite de pruebas de integración

Alcance/Funcionalidad:
- POST `/tasks` → orquestación completa → resultados visibles

Output:
- Reporte con resultados y tiempos

Validación:
- Ganador del consenso rinde mejor en validación final
- Listo si: estado final `completed` y métricas coherentes

---

✅ *** Conversación #40: Performance Tuning & Benchmarking Accuracy ***

Debes consultar estos doc para la tarea:
- 07-LLM-INTEGRATION.md (límites/cuotas/retentos)
- 05-CONSENSUS-BENCHMARKING.md (tolerancias de precisión)

Tarea: Afinar tiempos, límites y precisión

Ubicación: Config y puntos críticos de infraestructura/use cases

Alcance/Funcionalidad:
- Ajustes de concurrencia y timeouts (LLM/MCP/DB)
- Verificar desviación ≤ 20% entre fork y main
- Monitoreo de costos de LLM

Output:
- Config ajustada y evidencias de precisión

Validación:
- Sin 429/timeout anómalos y precisión dentro de umbral
- Listo si: estabilidad y precisión confirmadas

---

### 💎 FASE 9: Innovación (Bonus)

✅ *** Conversación #41: Búsqueda Híbrida (pg_text + pgvector) ***

Debes consultar estos doc para la tarea:
- 02-DATA-MODEL.md (query_logs)
- 06-TIGER-CLOUD-MCP.md (operaciones DB)

Tarea: Integrar búsqueda híbrida y consulta de similares

Ubicación: Módulos de infraestructura/DB y use cases

Alcance/Funcionalidad:
- Generar embeddings y crear índices FTS/vector
- Consulta híbrida con ponderación (texto 0.4, vector 0.6)
- Integración como insumo para Router/Orchestrator

Output:
- Tablas/índices y lógica de consulta híbrida

Validación:
- Resultados relevantes y desempeño aceptable
- Listo si: consultas híbridas devuelven similares consistentes

---

### 📤 FASE 10: Despliegue y Sumisión Final

✅ *** Conversación #42: Documentación Final y Pulido ***

Debes consultar estos doc para la tarea:
- 09-FRONTEND-COMPONENTS.md, 10-DEVELOPMENT-WORKFLOW.md, 11-DEPLOYMENT-STRATEGY.md

Tarea: Cerrar documentación y actualizar estado

Ubicación: `docs/` y `README.md`

Alcance/Funcionalidad:
- ✅ Actualizar estado del proyecto, diagramas y guías
- ✅ README.md: Estado actual, roadmap (Conv 27-45), feature matrix
- ✅ IMPLEMENTATION-STATUS.md: Conv #42 update con status actual
- ✅ Todos los 11 docs técnicos verificados y completos
- ✅ Tabla de conversiones progress actualizada

Output:
- ✅ Docs al día y README con enlaces de demo/credenciales (próximo Conv #45)
- ✅ Roadmap visible (45 conversaciones total)
- ✅ Feature completion 100% documentado

Validación:
- ✅ Documentación consistente con la implementación
- ✅ Listo: checklist de docs completo y actualizado

---

✅ *** Conversación #43: Preparación de Despliegue ***

Debes consultar estos doc para la tarea:
- 11-DEPLOYMENT-STRATEGY.md (plataformas y envs)

Tarea: Configurar entorno productivo

Ubicación: Infraestructura de despliegue

Alcance/Funcionalidad:
- Dockerfile.prod backend/frontend
- Variables de entorno por plataforma Vercel
- Checklist de pre-migración

Output:
- Artefactos de despliegue finalizados

Validación:
- Builds reproducibles y configs verificadas

MCP: Con USE_TIGER_CLOUD=true y variables de MCP configuradas, el Connect() no debe fallar.
LLM: Requiere VERTEX_PROJECT_ID, VERTEX_LOCATION y variables: GEMINI_CEREBRO_MODEL, GEMINI_OPERATIVO_MODEL, GEMINI_BULK_MODEL.

- Listo si: pipelines listos para ejecutar


---

✅ *** Conversación #44: Ejecución del Despliegue ***

Debes consultar estos doc para la tarea:

Tarea: Desplegar backend y frontend

Ubicación: Plataformas seleccionadas Vercel

Alcance/Funcionalidad:
- Ejecutar despliegue y pruebas post-deploy (API/WS/health)
- E2E en entorno productivo

Output:
- Servicios en producción operativos

Validación:
- Conectividad, WS y health checks en verde
- ✅ Listo si: E2E productivo pasa sin errores

---

*** Conversación #45: Sumisión Final ***

Debes consultar estos doc para la tarea:
- 11-DEPLOYMENT-STRATEGY.md (Challenge Submission)

Tarea: Completar entrega del desafío

Ubicación: Repositorio y plataforma de publicación

Alcance/Funcionalidad:
- Post de DEV.to (título, demo, repo, video, highlights Tiger Cloud)
- Video demo (arquitectura, live demo, integración Tiger)
- Accesos para jueces (demo/API)

Output:
- Post publicado, video accesible, repo público

Validación:
- Enlaces funcionales y checklist de sumisión completo
- ✅ Listo si: entrega final validada y accesible