#!/bin/bash

echo "🏥 AFS Challenge - Health Monitor"
echo "=================================="
echo ""

# Load env file (default .env) so we don't need to export vars manually
ENV_FILE=${ENV_FILE:-.env}
if [ -f "$ENV_FILE" ]; then
  echo "Using env file: $ENV_FILE"
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
else
  echo "No env file found at $ENV_FILE (you can set ENV_FILE to point to another path)"
fi

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAIL=0
CRIT_SUMMARY=()

# Check service status via docker compose (independent del nombre completo del contenedor)
check_container() {
    local service=$1
    if docker compose ps --services --status running | grep -qx "$service"; then
        return 0
    else
        return 1
    fi
}

# Get container id for a given service (useful for exec checks)
get_cid() {
    local service=$1
    docker compose ps -q "$service"
}

# Check endpoint
check_endpoint() {
    local name=$1
    local url=$2
    
    if curl -sf --max-time 5 "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ $name is DOWN or unreachable${NC}"
        return 1
    fi
}

# Check each service
echo "📦 Container Status:"
echo "-------------------"

# Backend
if check_container "backend"; then
    echo -e "${GREEN}✓${NC} afs-backend is running"
    if check_endpoint "Backend API" "http://localhost:8000/health"; then :; else
        FAIL=$((FAIL+1)); CRIT_SUMMARY+=("backend_api")
    fi
    # Smoke test: POST /api/v1/tasks should return 201 (requires fields now)
    echo -n "🔧 Testing POST /api/v1/tasks ... "
    HTTP_CODE=$(curl -s -o /tmp/afs_task_resp.json -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -X POST \
        -d '{"type":"query_optimization","target_query":"SELECT 1"}' \
        http://localhost:8000/api/v1/tasks || true)
    if [ "$HTTP_CODE" = "201" ]; then
        echo -e "${GREEN}OK (201)${NC}"
    else
        echo -e "${RED}FAIL (code=$HTTP_CODE)${NC}"
        if [ -f /tmp/afs_task_resp.json ]; then
            echo "Response:"; cat /tmp/afs_task_resp.json; echo ""
        fi
    fi
else
    echo -e "${RED}✗${NC} afs-backend is NOT running"
    FAIL=$((FAIL+1)); CRIT_SUMMARY+=("backend_container")
fi
echo ""

# Frontend
if check_container "frontend"; then
    echo -e "${GREEN}✓${NC} afs-frontend is running"
    check_endpoint "Frontend" "http://localhost:3000/"
else
    echo -e "${RED}✗${NC} afs-frontend is NOT running"
fi
echo ""

# Caddy
if check_container "caddy"; then
    echo -e "${GREEN}✓${NC} afs-caddy is running"
    check_endpoint "Caddy Proxy" "http://localhost/health"
else
    echo -e "${YELLOW}⚠${NC}  afs-caddy is NOT running"
    echo "   Run: docker compose up -d caddy"
fi
echo ""

# PostgreSQL
if check_container "postgres"; then
    echo -e "${GREEN}✓${NC} afs-postgres is running"
    if docker compose exec -T postgres pg_isready -U afs_user > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL is healthy${NC}"
    else
        echo -e "${RED}❌ PostgreSQL is DOWN${NC}"
    fi
else
    echo -e "${RED}✗${NC} afs-postgres is NOT running"
fi
echo ""

# Redis
if check_container "redis"; then
    echo -e "${GREEN}✓${NC} afs-redis is running"
    if docker compose exec -T redis redis-cli -a "${REDIS_PASSWORD:-afs_redis_pass}" ping > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Redis is healthy${NC}"
    else
        echo -e "${RED}❌ Redis is DOWN${NC}"
    fi
else
    echo -e "${RED}✗${NC} afs-redis is NOT running"
fi
echo ""

# Tiger Cloud MCP
if [ "${USE_TIGER_CLOUD}" = "true" ] || [ -n "${TIGER_MCP_URL}" ]; then
    URL="${TIGER_MCP_URL:-https://mcp.tigercloud.io}"
    echo "🌐 MCP Server: $URL"
    MCP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "User-Agent: afs-health" --max-time 5 "$URL" 2>/dev/null || echo "000")
    # Accept only valid 3-digit codes; treat others as failure
    if [[ "$MCP_CODE" =~ ^[0-9]{3}$ ]] && [ "$MCP_CODE" -ge 200 ] && [ "$MCP_CODE" -lt 500 ]; then
        echo -e "${GREEN}✅ MCP reachable (HTTP $MCP_CODE)${NC}"
    else
        echo -e "${RED}❌ MCP unreachable (code=$MCP_CODE)${NC}"
        FAIL=$((FAIL+1)); CRIT_SUMMARY+=("mcp")
    fi
    echo ""
else
    echo -e "${YELLOW}⏭ Skipping MCP check (set USE_TIGER_CLOUD=true or TIGER_MCP_URL)${NC}"
    echo ""
fi

# Tiger Cloud Database (via tiger CLI + psql; fallback to DATABASE_URL)
if [ "${USE_TIGER_CLOUD}" = "true" ]; then
    # Determine DSN source
    DSN=""
    DSN_SOURCE=""
    HAVE_TIGER=0; HAVE_PSQL=0
    if command -v tiger >/dev/null 2>&1; then HAVE_TIGER=1; fi
    if command -v psql  >/dev/null 2>&1; then HAVE_PSQL=1; fi

    if [ $HAVE_TIGER -eq 1 ] && [ -n "${TIGER_MAIN_SERVICE}" ]; then
        DSN=$(tiger db connection-string "${TIGER_MAIN_SERVICE}" 2>/dev/null || true)
        DSN_SOURCE="tiger"
    elif [ -n "${DATABASE_URL}" ]; then
        DSN="${DATABASE_URL}"
        DSN_SOURCE="env:DATABASE_URL"
    fi

    if [ -z "$DSN" ]; then
        echo -e "${YELLOW}⏭ Skipping Tiger DB check (no DSN; set TIGER_MAIN_SERVICE or DATABASE_URL)${NC}"
    else
        if [ $HAVE_PSQL -ne 1 ]; then
            echo -e "${YELLOW}⚠ psql not found; install PostgreSQL client (e.g., apt-get install -y postgresql-client) to enable DB checks${NC}"
        else
            # Use TIGER_DB_PASSWORD if provided to avoid interactive prompts
            if [ -n "${TIGER_DB_PASSWORD}" ]; then
                export PGPASSWORD="${TIGER_DB_PASSWORD}"
            fi
            echo "🐯 Tiger DB (${DSN_SOURCE})"
            if psql "$DSN" -c "SELECT 1" >/dev/null 2>&1; then
                echo -e "${GREEN}✅ Tiger DB connection OK${NC}"
                SCHEMA_COUNT=$(psql -At "$DSN" -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('users','orders','payments','tasks','agent_executions','optimization_proposals','benchmark_results','consensus_decisions');" 2>/dev/null || echo 0)
                echo "   Public schema required tables present: $SCHEMA_COUNT/8"
            else
                echo -e "${RED}❌ Tiger DB connection FAILED${NC}"
                FAIL=$((FAIL+1)); CRIT_SUMMARY+=("tiger_db")
            fi
        fi
    fi
    echo ""
else
    echo -e "${YELLOW}⏭ Skipping Tiger DB check (USE_TIGER_CLOUD!=true)${NC}"
    echo ""
fi

# WebSocket (optional, if websocat available)
if command -v websocat >/dev/null 2>&1; then
    echo -n "🧪 WS /ws handshake ... "
    if timeout 3 websocat -t -1 ws://localhost:8000/ws >/dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${YELLOW}SKIP/NO-WS${NC}"
    fi
    echo ""
fi

echo "=================================="
echo "🔍 Quick access URLs:"
echo "  Backend:  http://localhost:8000/health"
echo "  Frontend: http://localhost:3000"
echo "  Caddy:    http://localhost"
if [ -n "${TIGER_MCP_URL}" ]; then
  echo "  MCP:      ${TIGER_MCP_URL}"
fi

if [ $FAIL -gt 0 ]; then
  echo ""
  echo -e "${RED}Overall: ${FAIL} critical check(s) failed -> ${CRIT_SUMMARY[*]}${NC}"
  exit 1
else
  echo ""
  echo -e "${GREEN}Overall: all critical checks passed${NC}"
  exit 0
fi