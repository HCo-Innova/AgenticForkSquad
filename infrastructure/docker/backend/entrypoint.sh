#!/bin/sh
set -e

echo "🚀 Starting Go Backend (Development)..."

# Authenticate with Tiger Cloud if credentials provided
if [ -n "$TIGER_PUBLIC_KEY" ] && [ -n "$TIGER_SECRET_KEY" ] && [ -n "$TIGER_PROJECT_ID" ]; then
  echo "🔐 Authenticating with Tiger Cloud..."
  echo "   Public Key: ${TIGER_PUBLIC_KEY:0:10}..."
  echo "   Project ID: $TIGER_PROJECT_ID"
  echo "   Config Dir: ${CONFIG_DIR:-/app/.tiger}"
  
  # Enable verbose tiger CLI logging
  export TIGER_DEBUG=1
  
  # Create config directory if doesn't exist
  mkdir -p "${CONFIG_DIR:-/app/.tiger}"
  chmod 700 "${CONFIG_DIR:-/app/.tiger}"
  
  # Attempt login and capture full output - RETRY logic
  LOGIN_ATTEMPTS=0
  LOGIN_MAX_ATTEMPTS=3
  until [ $LOGIN_ATTEMPTS -ge $LOGIN_MAX_ATTEMPTS ]; do
    LOGIN_ATTEMPTS=$((LOGIN_ATTEMPTS + 1))
    echo "   Attempt $LOGIN_ATTEMPTS/$LOGIN_MAX_ATTEMPTS..."
    
    if tiger --config-dir "${CONFIG_DIR:-/app/.tiger}" auth login \
      --public-key "$TIGER_PUBLIC_KEY" \
      --secret-key "$TIGER_SECRET_KEY" \
      --project-id "$TIGER_PROJECT_ID" 2>&1 | tee /tmp/tiger_login.log; then
      echo "✅ Tiger authentication successful"
      break
    else
      if [ $LOGIN_ATTEMPTS -lt $LOGIN_MAX_ATTEMPTS ]; then
        echo "   ⚠️  Attempt failed, retrying..."
        sleep 2
      fi
    fi
  done
  
  if [ $LOGIN_ATTEMPTS -eq $LOGIN_MAX_ATTEMPTS ]; then
    echo "❌ Tiger authentication failed after $LOGIN_MAX_ATTEMPTS attempts:"
    cat /tmp/tiger_login.log
    exit 1
  fi
  
  # Verify auth status with full output
  echo "🔍 Verifying Tiger auth status..."
  tiger --config-dir "${CONFIG_DIR:-/app/.tiger}" auth status 2>&1 | tee /tmp/tiger_status.log || true
  
  # List services to ensure connectivity
  echo "📋 Listing services to verify connectivity..."
  tiger --config-dir "${CONFIG_DIR:-/app/.tiger}" service list --limit 1 2>&1 | head -5 || true
fi

echo "🎯 Starting Air hot reload..."

# Execute the command passed to this entrypoint (e.g., "air -c .air.toml")
exec "$@"
