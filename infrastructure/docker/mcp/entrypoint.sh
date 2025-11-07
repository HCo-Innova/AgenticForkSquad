#!/bin/sh
set -e

# Tiger MCP Server Entrypoint
# Handles authentication and starts MCP server

CONFIG_DIR="${CONFIG_DIR:-/root/.config/tiger}"
mkdir -p "$CONFIG_DIR"

# Log helper
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

log "🔐 Tiger MCP Entrypoint: Starting authentication flow"

# Check if credentials already exist (from previous login or mounted secret)
if [ -f "$CONFIG_DIR/credentials" ] || [ -f "$CONFIG_DIR/config.toml" ]; then
  log "✅ Credentials file found. Attempting status check..."
  if tiger auth status 2>/dev/null; then
    log "✅ Authentication successful via stored credentials"
  else
    log "⚠️  Stored credentials may be invalid. Proceeding with environment variables..."
  fi
else
  log "ℹ️  No stored credentials found. Checking environment variables..."
  
  # Try to authenticate using environment variables
  if [ -n "$TIGER_PUBLIC_KEY" ] && [ -n "$TIGER_SECRET_KEY" ] && [ -n "$TIGER_PROJECT_ID" ]; then
    log "🔑 Found TIGER_PUBLIC_KEY, TIGER_SECRET_KEY, and TIGER_PROJECT_ID in environment"
    
    if tiger auth login \
      --public-key "$TIGER_PUBLIC_KEY" \
      --secret-key "$TIGER_SECRET_KEY" \
      --project-id "$TIGER_PROJECT_ID" 2>&1 | tee /tmp/tiger_auth.log; then
      log "✅ Tiger authentication successful via environment variables"
    else
      log "❌ Tiger authentication failed. Check logs at /tmp/tiger_auth.log"
      cat /tmp/tiger_auth.log
      exit 1
    fi
  else
    log "⚠️  TIGER_PUBLIC_KEY, TIGER_SECRET_KEY, or TIGER_PROJECT_ID not set"
    log "ℹ️  MCP will start but requests will fail without valid credentials"
    log "📝 To authenticate, set: TIGER_PUBLIC_KEY, TIGER_SECRET_KEY, TIGER_PROJECT_ID"
  fi
fi

# Verify authentication status
log "🔍 Verifying Tiger CLI status..."
tiger auth status || log "⚠️  Could not verify Tiger status (may fail during runtime)"

# Start MCP server with verbose logging
log "🚀 Starting Tiger MCP server on http://0.0.0.0:9090"
log "📋 Environment: TIGER_DEBUG=$(echo $TIGER_DEBUG), CONFIG_DIR=$CONFIG_DIR"

# Export debug flag for tiger CLI
export TIGER_DEBUG=1

exec tiger mcp start http --host 0.0.0.0 --port 9090
