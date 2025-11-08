#!/bin/bash

# ========================================
# AFS - Deployment Setup Script
# Purpose: Prepare credentials and URLs for Vercel/Railway deployment
# ========================================

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SECRETS_DIR="${PROJECT_DIR}/secrets"
TMP_DIR="${PROJECT_DIR}/.deployment-temp"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Main Script
main() {
    log_info "🚀 AFS Deployment Setup"
    echo ""
    
    # Check prerequisites
    log_info "Checking prerequisites..."
    
    if [ ! -f "${SECRETS_DIR}/gcp_credentials.json" ]; then
        log_error "GCP credentials not found at ${SECRETS_DIR}/gcp_credentials.json"
        exit 1
    fi
    log_success "GCP credentials found"
    
    if [ ! -f "${PROJECT_DIR}/.env" ]; then
        log_error ".env file not found"
        exit 1
    fi
    log_success ".env file found"
    
    # Create temp directory
    mkdir -p "${TMP_DIR}"
    log_success "Temp directory created"
    
    echo ""
    log_info "=== Step 1: GCP Credentials Base64 Encoding ==="
    echo ""
    
    # Generate base64
    log_info "Encoding GCP credentials to base64..."
    GCP_B64=$(cat "${SECRETS_DIR}/gcp_credentials.json" | base64 -w 0)
    GCP_B64_FILE="${TMP_DIR}/gcp_credentials.b64"
    echo "${GCP_B64}" > "${GCP_B64_FILE}"
    log_success "Base64 encoded (${#GCP_B64} chars)"
    
    echo ""
    log_info "📋 GCP_CREDENTIALS_BASE64 (for Railway):"
    echo "---"
    echo "${GCP_B64}"
    echo "---"
    echo ""
    
    log_warn "⚠️  Copy the above value to Railway → Project Settings → Variables"
    log_warn "    Variable Name: GCP_CREDENTIALS_BASE64"
    
    echo ""
    log_info "=== Step 2: Extract Tiger Cloud Credentials ==="
    echo ""
    
    # Source .env
    set +e
    source "${PROJECT_DIR}/.env" 2>/dev/null
    set -e
    
    # Create credentials summary
    cat > "${TMP_DIR}/tiger-cloud-credentials.txt" << 'EOF'
=== TIGER CLOUD CREDENTIALS (for Railway) ===

Tiger Cloud - Main Database:
USE_TIGER_CLOUD=true
TIGER_PROJECT_ID={{TIGER_PROJECT_ID}}
TIGER_MAIN_SERVICE={{TIGER_MAIN_SERVICE}}
TIGER_SERVICE_ID={{TIGER_SERVICE_ID}}
TIGER_DB_USER={{TIGER_DB_USER}}
TIGER_DB_PASSWORD={{TIGER_DB_PASSWORD}}
TIGER_DB_HOST={{TIGER_DB_HOST}}
TIGER_DB_PORT={{TIGER_DB_PORT}}
TIGER_DB_NAME={{TIGER_DB_NAME}}
TIGER_DB_SSLMODE={{TIGER_DB_SSLMODE}}
DATABASE_URL={{DATABASE_URL}}
TIGER_PUBLIC_KEY={{TIGER_PUBLIC_KEY}}
TIGER_SECRET_KEY={{TIGER_SECRET_KEY}}

Tiger Cloud - Fork A1:
TIGER_FORK_A1_SERVICE_ID={{TIGER_FORK_A1_SERVICE_ID}}
TIGER_FORK_A1_PASSWORD={{TIGER_FORK_A1_PASSWORD}}
TIGER_FORK_A1_HOST={{TIGER_FORK_A1_HOST}}
TIGER_FORK_A1_PORT={{TIGER_FORK_A1_PORT}}
TIGER_FORK_A1_SERVICE_URL={{TIGER_FORK_A1_SERVICE_URL}}

Tiger Cloud - Fork A2:
TIGER_FORK_A2_SERVICE_ID={{TIGER_FORK_A2_SERVICE_ID}}
TIGER_FORK_A2_PASSWORD={{TIGER_FORK_A2_PASSWORD}}
TIGER_FORK_A2_HOST={{TIGER_FORK_A2_HOST}}
TIGER_FORK_A2_PORT={{TIGER_FORK_A2_PORT}}
TIGER_FORK_A2_SERVICE_URL={{TIGER_FORK_A2_SERVICE_URL}}

GCP & Vertex AI:
VERTEX_PROJECT_ID={{VERTEX_PROJECT_ID}}
VERTEX_LOCATION={{VERTEX_LOCATION}}
GEMINI_CEREBRO_MODEL={{GEMINI_CEREBRO_MODEL}}
GEMINI_OPERATIVO_MODEL={{GEMINI_OPERATIVO_MODEL}}
GEMINI_BULK_MODEL={{GEMINI_BULK_MODEL}}

Backend Config:
PORT={{PORT}}
ENV={{ENV}}
LOG_LEVEL={{LOG_LEVEL}}
RUN_MIGRATIONS={{RUN_MIGRATIONS}}
EOF
    
    log_success "Tiger Cloud credentials extracted"
    
    echo ""
    log_info "=== Step 3: Generate Deployment Checklist ==="
    echo ""
    
    cat > "${TMP_DIR}/deployment-checklist.md" << 'EOF'
# 🚀 Deployment Checklist

## Railway Backend

### 1. Create Railway Project
- [ ] Go to https://railway.app
- [ ] Click "New Project"
- [ ] Select "Deploy from GitHub"
- [ ] Select repo: HCo-Innova/AgenticForkSquad
- [ ] Wait for build to complete

### 2. Configure Railway Variables
- [ ] Go to Project Settings → Variables
- [ ] Add all Tiger Cloud credentials (see tiger-cloud-credentials.txt)
- [ ] Add GCP_CREDENTIALS_BASE64 (see above)
- [ ] Add VERTEX_PROJECT_ID, VERTEX_LOCATION
- [ ] Set PORT=8000
- [ ] Set ENV=production
- [ ] Set LOG_LEVEL=info
- [ ] Save and redeploy

### 3. Verify Railway Deployment
- [ ] Check build status (should be green)
- [ ] Copy Railway public URL (e.g., https://afs-backend-prod.railway.app)
- [ ] Test health endpoint:
  curl https://afs-backend-prod.railway.app/health

### 4. Get Railway URL
- [ ] Note the public URL for Vercel configuration

## Vercel Frontend

### 1. Create Vercel Project
- [ ] Go to https://vercel.com
- [ ] Click "Add New" → "Project"
- [ ] Import repo: HCo-Innova/AgenticForkSquad
- [ ] Select root: ./frontend (if prompted)
- [ ] Select Build: Vite

### 2. Configure Vercel Variables
- [ ] Go to Settings → Environment Variables
- [ ] Add VITE_API_URL=https://[RAILWAY_URL]/api/v1
  (Replace [RAILWAY_URL] with actual Railway URL from step 4 above)
- [ ] Add VITE_WS_URL=wss://[RAILWAY_URL]/ws
- [ ] Add NODE_ENV=production
- [ ] Save and redeploy

### 3. Verify Vercel Deployment
- [ ] Check build status (should be green)
- [ ] Open Vercel URL in browser
- [ ] Check browser console for errors
- [ ] Verify Network tab shows API calls to Railway

## Post-Deployment Verification

### Health Checks
- [ ] Backend health: curl https://[RAILWAY_URL]/health
- [ ] Frontend loads: curl https://[VERCEL_URL]
- [ ] WebSocket connects (check browser console)
- [ ] Tiger Cloud accessible (check backend logs)
- [ ] GCP credentials working (check backend logs for Vertex AI)

### Full E2E Test
- [ ] Login to frontend
- [ ] Create optimization task
- [ ] Verify data flows through Gemini
- [ ] Check fork operations work
- [ ] Verify results displayed in UI

EOF
    
    log_success "Checklist generated"
    
    echo ""
    log_info "=== Step 4: Save Configuration ==="
    echo ""
    
    log_success "All files saved to ${TMP_DIR}/"
    echo ""
    log_info "Files generated:"
    echo "  1. gcp_credentials.b64 - Base64 encoded GCP credentials"
    echo "  2. tiger-cloud-credentials.txt - All Tiger Cloud values"
    echo "  3. deployment-checklist.md - Step-by-step deployment guide"
    
    echo ""
    log_info "=== Next Steps ==="
    echo ""
    echo "1. Copy GCP_CREDENTIALS_BASE64:"
    echo "   cat ${TMP_DIR}/gcp_credentials.b64"
    echo ""
    echo "2. Add to Railway project:"
    echo "   Railway → Project Settings → Variables"
    echo "   Name: GCP_CREDENTIALS_BASE64"
    echo "   Value: [paste from above]"
    echo ""
    echo "3. Add Tiger Cloud credentials to Railway:"
    echo "   cat ${TMP_DIR}/tiger-cloud-credentials.txt"
    echo ""
    echo "4. Follow deployment-checklist.md:"
    echo "   cat ${TMP_DIR}/deployment-checklist.md"
    echo ""
    
    log_success "Deployment setup complete!"
}

main "$@"
