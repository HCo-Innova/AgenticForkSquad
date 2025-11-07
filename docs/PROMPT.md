Trabajamos en AFS Challenge (Go backend + Tiger Cloud PostgreSQL).
Bloqueador actual: Tiger Cloud fork API devuelve "Error: unknown error".
Solución: Resolver este error OR usar workaround diferente.

- Arquitectura: CLI proxy pattern (Go exec.Command → tiger CLI v0.15.1 → Tiger Cloud API)
- Código: MCPClient en backend/internal/infrastructure/mcp/client.go
- Problema: docker compose exec backend /app/validate_pitr falla en CreateFork()
- Error exacto: "mcp: fork failed: exit status 1 (output: Error: unknown error)"
- Lo que funciona: auth (✅), service list (✅), service describe (✅)
- Lo que falla: fork (❌)

Diagnosticados:
- Tiger CLI version: v0.15.1 (correct)
- Credentials: Almacenadas en ~/.config/tiger/config.yaml (auth succeeded)
- Service: o120o0yba9 (status READY)
- Project: a1lqw18o6u (valid)
- Network: OK (otros comandos llegan a Tiger Cloud)

Likely causa: Fork capability no habilitada en cuenta test o limitación de plan.

Hacer funcionar `tiger service fork <service-id> --name <name> --now -o json` 
(actualmente: "Error: unknown error").

- /srv/afs-challenge/backend/internal/infrastructure/mcp/client.go (CreateFork method)
- /srv/afs-challenge/docs/IMPLEMENTATION-STATUS.md (status completo + diagnostics)
- /srv/afs-challenge/docker-compose.yml (Tiger credentials injected)

# Historiar Rápida:

Pasamos de HTTP MCP complejo → CLI proxy stateless
Resolvimos permisos (separate config dirs)
Corregimos sintaxis tiger CLI (positional args, -o json flags)
Ahora: Tiger Cloud API rechaza fork request (opaque error)

# Pasa esta instrucción al siguiente agente:

Proyecto: AFS Challenge (Go + Tiger Cloud). Bloqueador: tiger service fork <service-id> --name <name> --now -o json falla con "Error: unknown error".

Código: client.go (CreateFork method, línea ~120).
Diagnostics completos: IMPLEMENTATION-STATUS.md (sección "Known Issues").

Tu tarea: (1) revisar Tiger Cloud account/service para fork capability; (2) si no está habilitado, habilitarlo; (3) si está habilitado, investigar error con Tiger support o probar con service diferente; (4) reportar solución o workaround.