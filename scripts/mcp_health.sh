#!/usr/bin/env bash
set -euo pipefail

# MCP/Tiger Cloud health checker
# - Verifica MCP HTTP listener
# - Verifica backend health
# - Opcional: usa TIGER_MCP_URL si está definida

MCP_URL="${TIGER_MCP_URL:-http://0.0.0.0:9090}"
BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://localhost:8000/health}"

ok() { echo "[OK] $1"; }
fail() { echo "[FAIL] $1" >&2; exit 1; }

# 1) MCP puerto abierto
if command -v nc >/dev/null 2>&1; then
  nc -zv "$(echo "$MCP_URL" | sed -E 's#https?://([^:/]+).*#\1#')" "$(echo "$MCP_URL" | sed -E 's#.*:([0-9]+).*#\1#')" >/dev/null 2>&1 || fail "MCP puerto no accesible en $MCP_URL"
  ok "MCP puerto accesible en $MCP_URL"
else
  echo "nc no instalado; usando curl para probar conectividad" >&2
fi

# 2) MCP HTTP respuesta
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$MCP_URL" || true)
if [[ "$HTTP_CODE" == "200" || "$HTTP_CODE" == "405" ]]; then
  ok "MCP HTTP responde (codigo $HTTP_CODE)"
else
  fail "MCP no responde correctamente (codigo $HTTP_CODE)"
fi

# 3) Backend health
curl -fsS "$BACKEND_HEALTH_URL" >/dev/null && ok "Backend healthy: $BACKEND_HEALTH_URL" || fail "Backend health fallo: $BACKEND_HEALTH_URL"

# 4) Resumen
echo "MCP y Backend OK"
