# 🚀 Setup Instructions

Guide for setting up the AFS (Agentic Fork Squad) project locally or for deployment.

## Table of Contents

1. [Local Development Setup](#local-development-setup)
2. [Prerequisites](#prerequisites)
3. [Configuration](#configuration)
4. [Running the Application](#running-the-application)
5. [Deployment Setup](#deployment-setup)

---

## Prerequisites

### Required

- **Docker & Docker Compose** (v20.10+)
  - [Install Docker](https://docs.docker.com/get-docker/)
  
- **Go 1.25+** (for backend development)
  - [Install Go](https://golang.org/doc/install)
  
- **Node.js 18+** (for frontend development)
  - [Install Node.js](https://nodejs.org/)

### Optional (for local Tiger Cloud testing)

- **Tiger Cloud CLI**
  - [Install Tiger CLI](https://docs.timescale.com/use-timescale/latest/cloud-cli/)

---

## Local Development Setup

### 1. Clone the Repository

```bash
git clone https://github.com/HCo-Innova/AgenticForkSquad.git
cd AgenticForkSquad
```

### 2. Create Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit with your local values
nano .env  # or use your preferred editor
```

**Minimum required values for local dev:**

```dotenv
# PostgreSQL (if using local)
POSTGRES_PASSWORD=dev_password_here
REDIS_PASSWORD=dev_redis_password_here

# Tiger Cloud (get from Tiger Cloud dashboard)
TIGER_PROJECT_ID=your_project_id
TIGER_MAIN_SERVICE=your_main_service_id
TIGER_DB_USER=your_db_user
TIGER_DB_PASSWORD=your_db_password
TIGER_DB_HOST=your_host.tsdb.cloud.timescale.com
TIGER_PUBLIC_KEY=your_public_key
TIGER_SECRET_KEY=your_secret_key

# GCP (for Gemini models)
VERTEX_PROJECT_ID=your_gcp_project_id
# Place credentials file at ./secrets/gcp_credentials.json

# Frontend URLs (for local dev behind Caddy reverse proxy)
VITE_API_URL=http://localhost/api
VITE_WS_URL=ws://localhost/ws
```

### 3. Set Up Secrets

```bash
# Create secrets directory
mkdir -p ./secrets

# Add GCP service account JSON
# Download from Google Cloud Console and place at:
cp /path/to/gcp_service_account.json ./secrets/gcp_credentials.json

# Verify it's ignored
git check-ignore ./secrets/gcp_credentials.json
# Expected output: ./secrets/gcp_credentials.json
```

### 4. Install Dependencies

#### Backend
```bash
cd backend
go mod download
cd ..
```

#### Frontend
```bash
cd frontend
npm install
cd ..
```

### 5. Build and Start Services

```bash
# Using Docker Compose (recommended)
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

**Services will be available at:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8000
- Reverse Proxy: http://localhost
- MCP Server: http://localhost:9090

### 6. Run Migrations (First Time)

```bash
# Migrations run automatically via docker-compose
# Verify database setup:
docker exec afs-backend goose status

# Manual migration (if needed)
docker exec afs-backend goose up
```

---

## Running the Application

### Start All Services

```bash
docker-compose up -d
```

### Start Individual Services

```bash
# Backend only
docker-compose up -d backend

# Frontend only
docker-compose up -d frontend

# Development with hot-reload
cd backend && air  # Requires 'air' installed
cd frontend && npm run dev
```

### Stop Services

```bash
# Stop all
docker-compose down

# Stop and remove volumes (CAREFUL: deletes data)
docker-compose down -v
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend

# Last 50 lines
docker-compose logs --tail 50 backend
```

### Health Checks

Each service has a health endpoint:

```bash
# Backend
curl http://localhost:8000/health

# Frontend
curl http://localhost:3000

# MCP Server
docker exec afs-mcp tiger service list
```

---

## Configuration

### Backend Configuration Files

Located in `backend/internal/config/`:

- **`config.go`** - Configuration structure and validation
- **`tiger.go`** - Tiger Cloud service configuration

### Frontend Configuration Files

Located in `frontend/`:

- **`vite.config.ts`** - Vite build configuration
- **`tailwind.config.js`** - Tailwind CSS configuration
- **`tsconfig.json`** - TypeScript configuration

### Database Configuration

Migrations are in `backend/migrations/`:

1. `001_create_schema.sql` - Core schema
2. `002_afs_tables.sql` - AFS-specific tables
3. `003_seed_data.sql` - Test data
4. `004_query_logs_hybrid_search.sql` - Hybrid search support

---

## Deployment Setup

### Prerequisites for Deployment

1. **GitHub Repository** - Code version control
2. **Vercel Account** - Frontend hosting
3. **Railway/Render/Fly.io Account** - Backend hosting
4. **Tiger Cloud Account** - Database (production)
5. **GCP Project** - Vertex AI access

### Backend Deployment (Railway Example)

1. **Connect GitHub repository to Railway**
   - Create Railway account
   - Click "New Project" → "Deploy from GitHub"
   - Select `AgenticForkSquad` repository

2. **Add Environment Variables**
   - Use Railway dashboard to set:
     - `DATABASE_URL` - Connection string
     - `TIGER_PROJECT_ID`, `TIGER_DB_PASSWORD` - Tiger Cloud
     - `VERTEX_PROJECT_ID`, `GOOGLE_APPLICATION_CREDENTIALS` - GCP
     - All other `.env` variables

3. **Add GCP Service Account**
   - Upload via Railway's file upload or base64-encode in env var
   - Or use external secret management (e.g., 1Password, HashiCorp Vault)

4. **Configure Domain**
   - Railway provides URL automatically
   - Point to custom domain if desired

### Frontend Deployment (Vercel)

1. **Connect GitHub Repository to Vercel**
   - Go to vercel.com
   - Click "Add New" → "Project"
   - Select `AgenticForkSquad` repository
   - Framework preset: "Vite"

2. **Add Environment Variables**
   - Dashboard → Settings → Environment Variables
   - Set:
     - `VITE_API_URL` - Backend API URL (Railway/Render domain)
     - `VITE_WS_URL` - WebSocket URL
     - `NODE_ENV=production`

3. **Deploy**
   - Vercel auto-deploys on git push to `main`
   - Check deployment at: `https://<project-name>.vercel.app`

### Database Deployment (Tiger Cloud)

1. **Create Tiger Cloud Services**
   - Main database for queries
   - Fork services (A1, A2) for agent isolation

2. **Configure Connection**
   - Update `.env` with Tiger Cloud credentials
   - Test connection:
     ```bash
     psql postgres://user:pass@host:port/dbname?sslmode=require \
       -c "SELECT version();"
     ```

3. **Verify Migrations**
   - Ensure all 4 migrations are applied
   - Check: `\dt` in psql to list tables

---

## Troubleshooting

### Backend Won't Start

```bash
# Check if port 8000 is in use
lsof -i :8000

# View detailed logs
docker-compose logs backend

# Rebuild container
docker-compose up -d --build backend
```

### Frontend Build Fails

```bash
# Clear cache and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Database Connection Issues

```bash
# Verify DATABASE_URL format
echo $DATABASE_URL

# Test connection directly
psql $DATABASE_URL -c "SELECT 1"

# Check credentials
grep TIGER_DB_ .env
```

### Tiger Cloud MCP Not Responding

```bash
# Check MCP health
docker-compose logs mcp

# Restart MCP
docker-compose restart mcp

# Verify credentials in secrets/mcp-config.json
cat secrets/mcp-config.json
```

---

## Development Workflow

### Local Development Checklist

- [ ] Clone repository
- [ ] Copy `.env.example` to `.env`
- [ ] Fill in credentials (Tiger Cloud, GCP, etc.)
- [ ] Run `docker-compose up -d`
- [ ] Verify all services healthy: `docker-compose ps`
- [ ] Access frontend at `http://localhost`
- [ ] Run migrations: `docker exec afs-backend goose up`

### Before Pushing Code

```bash
# 1. No secrets in git
git diff --cached | grep -iE "password|secret|key|token" && exit 1

# 2. Format code
cd backend && go fmt ./...
cd frontend && npm run lint

# 3. Run tests (if available)
cd backend && go test ./...

# 4. Verify .env is ignored
git check-ignore .env

# 5. Safe to push
git push origin main
```

---

## Additional Resources

- [Backend Architecture](docs/03-SYSTEM-ARCHITECTURE.md)
- [Tiger Cloud Documentation](https://docs.timescale.com/)
- [Vertex AI Documentation](https://cloud.google.com/vertex-ai)
- [Vercel Deployment Guide](https://vercel.com/docs)
- [Railway Deployment Guide](https://docs.railway.app/)

---

**Last Updated:** November 2025  
**Maintained By:** AFS Team
