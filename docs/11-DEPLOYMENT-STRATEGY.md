# 11-DEPLOYMENT-STRATEGY.md

```markdown
# 11-DEPLOYMENT-STRATEGY.md

**Project:** Agentic Fork Squad (AFS)  
**Document Type:** Deployment Strategy  
**Last Updated:** 2024  
**Related Docs:** [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md), 
[10-DEVELOPMENT-WORKFLOW.md](10-DEVELOPMENT-WORKFLOW.md), 
[06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md)

---

## 📖 Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Pre-Deployment Checklist](#pre-deployment-checklist)
3. [Tiger Cloud Setup](#tiger-cloud-setup)
4. [Migration from Local to Tiger](#migration-from-local-to-tiger)
5. [Environment Configuration](#environment-configuration)
6. [Database Migration](#database-migration)
7. [Application Deployment](#application-deployment)
8. [Verification & Testing](#verification--testing)
9. [Challenge Submission](#challenge-submission)
10. [Rollback Procedures](#rollback-procedures)
11. [Monitoring & Maintenance](#monitoring--maintenance)

---

## 🎯 Deployment Overview

### Deployment Phases

**Phase 1: Local Development (Days 1-5)**
- Full stack running in Docker containers
- PostgreSQL + Redis locally
- Hot reload for rapid development
- No Tiger Cloud integration

**Phase 2: Tiger Cloud Integration (Days 6-7)**
- Migrate database to Tiger Cloud
- Enable MCP client
- Test fork operations
- Production configuration

**Phase 3: Challenge Submission (Day 7)**
- Final deployment verification
- Create demo credentials
- Submit to DEV.to
- Provide judge access

---

### Architecture Transition

**Before (Local Development):**
```
┌─────────────────────────────────────┐
│  Docker Compose                     │
│  ┌──────────────────────────────┐   │
│  │ Backend (Go)                 │   │
│  │ ├─ Fiber API                 │   │
│  │ └─ PostgreSQL connection     │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ PostgreSQL 16                │   │
│  │ └─ Direct connection         │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ Frontend (React)             │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

**After (Production - Tiger Cloud):**
```
┌─────────────────────────────────────┐
│  Backend (Deployed)                 │
│  ┌──────────────────────────────┐   │
│  │ MCP Client                   │   │
│  │ └─ Tiger Cloud API           │   │
│  └──────────────────────────────┘   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Tiger Cloud                        │
│  ┌──────────────────────────────┐   │
│  │ afs-main (PostgreSQL)        │   │
│  │ └─ MCP Server                │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ Dynamic Forks                │   │
│  │ └─ Zero-copy isolation       │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│  Frontend (Deployed)                │
│  └─ Static hosting or container     │
└─────────────────────────────────────┘
```

---

### Deployment Options

**Option 1: Hybrid Deployment (Recommended for Challenge)**
- Backend: Cloud hosting (Railway, Render, Fly.io)
- Database: Tiger Cloud (main feature)
- Frontend: Vercel or Netlify (static)

**Option 2: Full Container Deployment**
- Backend + Frontend: Single VPS with Docker
- Database: Tiger Cloud
- Reverse proxy: Caddy or Nginx

**Option 3: Minimal Deployment (Demo Only)**
- Backend: Local machine with ngrok (temporary)
- Database: Tiger Cloud
- Frontend: Local with public URL
- **Use Case:** Quick demo for judges

---

## ✅ Pre-Deployment Checklist

### Code Readiness

**Backend:**
- [ ] All tests passing (go test ./...)
- [ ] Linter clean (golangci-lint run)
- [ ] No hardcoded credentials
- [ ] Environment variables documented
- [ ] Error handling comprehensive
- [ ] Logging configured for production

**Frontend:**
- [ ] All tests passing (npm run test)
- [ ] Build succeeds (npm run build)
- [ ] No console errors in production build
- [ ] API URL configurable via env
- [ ] Assets optimized

**Database:**
- [ ] Migrations tested (up and down)
- [ ] Seed data script working
- [ ] Indexes created
- [ ] No sensitive data in migrations

---

### Documentation

- [ ] README.md updated with deployment steps
- [ ] API documentation complete
- [ ] Environment variables listed
- [ ] Architecture diagrams current
- [ ] Challenge submission draft ready

---

### Tiger Cloud

- [ ] Account created and verified
- [ ] CLI installed and authenticated
- [ ] Test service created successfully
- [ ] MCP configuration working locally
- [ ] Fork operations tested
- [ ] Connection string obtained

---

### Security

- [ ] API keys in environment variables
- [ ] No secrets in git repository
- [ ] .gitignore configured properly
- [ ] Demo credentials separate from dev
- [ ] Rate limiting configured
- [ ] CORS configured correctly

---

## ☁️ Tiger Cloud Setup

### Account Preparation

**Step 1: Create Account**
```
Visit: https://tiger.cloud (or actual URL)
Sign up with email
Verify email address
Complete profile
```

**Step 2: Install CLI**
```bash
# macOS
brew install tiger-cloud/tap/tiger

# Linux
curl -sSL https://get.tigercloud.io | sh

# Verify
tiger --version
```

**Step 3: Authenticate**
```bash
# Login (opens browser)
tiger auth login

# Verify
tiger auth whoami
```

**Expected Output:**
```
Logged in as: your-email@example.com
Organization: personal
```

---

### Create Main Service

**Step 1: Create Service**
```bash
tiger service create --name afs-production
```

**Output:**
```
✓ Service created successfully

Service ID: afs-production
Region: us-west-2
PostgreSQL: 16.1
Status: active

Created: 2024-11-08 10:00:00 UTC
```

**Step 2: Get Connection String**
```bash
tiger db connection-string afs-production
```

**Output:**
```
postgresql://user:password@db.tiger.cloud:5432/afs_production
```

**Step 3: Save Credentials**
```bash
# Store securely (not in git)
echo "TIGER_DATABASE_URL=postgresql://user:pass@..." >> .env.production

# Or use secret manager
# AWS Secrets Manager
# Vercel Environment Variables
# Railway Variables
```

---

### Test Connection

**Direct Connection:**
```bash
psql $(tiger db connection-string afs-production)
```

**Expected:**
```
psql (16.1)
Type "help" for help.

afs_production=>
```

**Test Query:**
```sql
SELECT version();
```

**If successful:** Tiger Cloud database ready for deployment

---

## 🔄 Migration from Local to Tiger

### Phase 1: Database Migration

**Step 1: Backup Local Database**
```bash
# Create backup of local development database
docker-compose exec -T postgres pg_dump \
  -U afs_user \
  -d afs_dev \
  > backup_pre_migration.sql
```

**Step 2: Run Migrations on Tiger**
```bash
# Get Tiger connection string
export TIGER_DB_URL=$(tiger db connection-string afs-production)

# Run migrations
migrate -path ./backend/migrations \
        -database "$TIGER_DB_URL" \
        up

# Or using custom migration script
DATABASE_URL=$TIGER_DB_URL go run backend/scripts/migrate/main.go up
```

**Step 3: Seed Production Data**
```bash
# Run seeder with production data
DATABASE_URL=$TIGER_DB_URL \
  go run backend/scripts/seed/main.go \
    --users=1000 \
    --orders=10000
```

**Step 4: Verify Schema**
```bash
# Connect to Tiger database
psql $TIGER_DB_URL

# Verify tables
\dt

# Check row counts
SELECT 'users' as table_name, COUNT(*) FROM users
UNION ALL
SELECT 'orders', COUNT(*) FROM orders
UNION ALL
SELECT 'tasks', COUNT(*) FROM tasks;
```

**Expected:**
```
 table_name | count
------------+-------
 users      |  1000
 orders     | 10000
 tasks      |     0
```

---

### Phase 2: Backend Configuration

**Step 1: Update Environment Variables**

**Create .env.production:**
```bash
# Environment
ENV=production
LOG_LEVEL=info

# Server
PORT=8000
HOST=0.0.0.0

# Database (Tiger Cloud)
USE_TIGER_CLOUD=true
TIGER_MAIN_SERVICE=afs-production
TIGER_MCP_URL=https://mcp.tigercloud.io
DATABASE_URL=${TIGER_DATABASE_URL}

# Redis (if using managed Redis)
REDIS_URL=redis://...

# LLM APIs - Vertex AI Only
VERTEX_PROJECT_ID=${GCP_PROJECT_ID}
VERTEX_LOCATION=us-central1
GEMINI_CEREBRO_MODEL=gemini-2.5-pro
GEMINI_OPERATIVO_MODEL=gemini-2.5-flash
GEMINI_BULK_MODEL=gemini-2.0-flash
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Migrations
RUN_MIGRATIONS=false  # Already run manually
```

---

**Step 2: Update Application Code**

**MCP Client Initialization:**
```
Conceptual change:

Before (Local):
  dbClient = NewPostgresClient(config.DatabaseURL)

After (Production):
  if config.UseTigerCloud {
    mcpClient = NewMCPClient(config.TigerMCPURL, config.TigerAPIToken)
    dbClient = NewTigerClient(mcpClient)
  } else {
    dbClient = NewPostgresClient(config.DatabaseURL)
  }
```

**Configuration Switch:**
```
Single environment variable controls behavior:
  USE_TIGER_CLOUD=true  → Use MCP client
  USE_TIGER_CLOUD=false → Use direct PostgreSQL

Allows testing locally before production
```

---

**Step 3: Test Locally with Tiger Cloud**

**Temporary local testing:**
```bash
# Stop local PostgreSQL
docker-compose stop postgres

# Update .env
USE_TIGER_CLOUD=true
TIGER_MAIN_SERVICE=afs-production
DATABASE_URL=<tiger-connection-string>

# Run backend locally
cd backend
go run cmd/server/main.go

# Verify connection
curl http://localhost:8000/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "dependencies": {
    "database": "connected",
    "mcp_server": "available"
  }
}
```

**If successful:** Backend can connect to Tiger Cloud

---

### Phase 3: Frontend Configuration

**Step 1: Update Environment Variables**

**Create .env.production:**
```bash
VITE_API_URL=https://your-backend-url.com/api
VITE_WS_URL=wss://your-backend-url.com/ws
```

**Step 2: Build for Production**
```bash
cd frontend

# Install dependencies
npm ci

# Build
npm run build

# Output in frontend/dist/
```

**Step 3: Test Production Build Locally**
```bash
# Serve production build
npm run preview

# Access at http://localhost:4173
# Verify:
# - No console errors
# - API calls work
# - WebSocket connects
```

---

## ⚙️ Environment Configuration

### Production Environment Variables

**Backend (.env.production):**
```bash
# === Application ===
ENV=production
PORT=8000
LOG_LEVEL=info
ALLOWED_ORIGINS=https://your-frontend-url.com

# === Tiger Cloud ===
USE_TIGER_CLOUD=true
TIGER_MAIN_SERVICE=afs-production
TIGER_MCP_URL=https://mcp.tigercloud.io
TIGER_API_TOKEN=<from-tiger-cli-credentials>

# === Vertex AI ===
VERTEX_PROJECT_ID=your-gcp-project
VERTEX_LOCATION=us-central1
VERTEX_MODEL_CEREBRO=gemini-2.5-pro      # gemini-2.5-pro
VERTEX_MODEL_OPERATIVO=gemini-2.5-flash  # gemini-2.5-flash
VERTEX_MODEL_BULK=gemini-2.0-flash       # gemini-2.0-flash
GOOGLE_APPLICATION_CREDENTIALS=/run/secrets/gcp_service_account

# === Redis (Optional) ===
REDIS_URL=<if-using-managed-redis>

# === Database ===
DATABASE_URL=<tiger-connection-string>
```

**Frontend (.env.production):**
```bash
VITE_API_URL=https://api.your-domain.com/api
VITE_WS_URL=wss://api.your-domain.com/ws
```

---

### Secret Management

**Option 1: Platform Environment Variables**

**Railway:**
```
Dashboard → Project → Variables
Add each variable
Redeploy to apply
```

**Render:**
```
Dashboard → Service → Environment
Add Key-Value pairs
Auto-redeploys
```

**Vercel (Frontend):**
```
Dashboard → Project → Settings → Environment Variables
Add variables for Production environment
Redeploy
```

---

**Option 2: Secret Files (Docker)**

**Create secrets directory:**
```bash
mkdir -p secrets/

# Add to .gitignore
echo "secrets/" >> .gitignore

# Create secret files
# Store your GCP service account JSON (ADC)
cp path/to/service-account.json secrets/gcp_service_account
```

**Docker Compose (Production):**
```
Conceptual:
services:
  backend:
    secrets:
      - gcp_service_account
    environment:
      GOOGLE_APPLICATION_CREDENTIALS: /run/secrets/gcp_service_account
      VERTEX_PROJECT_ID: ${VERTEX_PROJECT_ID}
      VERTEX_LOCATION: ${VERTEX_LOCATION}
      VERTEX_MODEL_CEREBRO: ${VERTEX_MODEL_CEREBRO}
      VERTEX_MODEL_OPERATIVO: ${VERTEX_MODEL_OPERATIVO}
      VERTEX_MODEL_BULK: ${VERTEX_MODEL_BULK}

secrets:
  gcp_service_account:
    file: ./secrets/gcp_service_account
```

---

## 🗄️ Database Migration

### Migration Execution

**Pre-Migration Checklist:**
- [ ] Backup current database
- [ ] Test migrations in staging
- [ ] Verify rollback procedures
- [ ] Schedule maintenance window (if needed)

**Execute Migrations:**

**Method 1: Manual Execution**
```bash
# Get connection string
TIGER_DB=$(tiger db connection-string afs-production)

# Run all pending migrations
migrate -path ./backend/migrations \
        -database "$TIGER_DB" \
        up

# Verify version
migrate -path ./backend/migrations \
        -database "$TIGER_DB" \
        version
```

**Method 2: Application Startup**
```
Set in environment:
  RUN_MIGRATIONS=true

On startup:
  Backend detects flag
  Runs pending migrations
  Continues to serve

Risk: Longer startup time
Benefit: Automatic migration
```

---

**Post-Migration Verification:**
```sql
-- Connect to Tiger database
psql <tiger-connection-string>

-- Verify all tables exist
\dt

-- Expected tables:
-- users, orders, payments (existing)
-- tasks, agent_executions, optimization_proposals,
-- benchmark_results, consensus_decisions (new)

-- Check row counts
SELECT 
  schemaname, 
  tablename, 
  n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY tablename;

-- Verify indexes
\di

-- Check schema version
SELECT version FROM schema_migrations;
```

---

### Data Seeding

**Production Seed Strategy:**

**Option 1: Full Seed**
```bash
# Seed all demo data
DATABASE_URL=$TIGER_DB \
  go run backend/scripts/seed/main.go
```

**Option 2: Minimal Seed**
```bash
# Only essential data for demo
DATABASE_URL=$TIGER_DB \
  go run backend/scripts/seed/main.go \
    --users=100 \
    --orders=1000
```

**Option 3: No Seed**
```
Start with empty tables
Create demo tasks during judge demo
Show real-time optimization
```

**Recommendation for Challenge:**
- Option 2: Minimal seed
- Enough data for demo
- Not overwhelming
- Fast to reset if needed

---

## 🚀 Application Deployment

### Deployment Platforms

**Backend Deployment Options:**

**Option 1: Railway (Recommended)**

**Advantages:**
- Easy Docker deployment
- Automatic HTTPS
- Environment variables UI
- Affordable pricing
- PostgreSQL addon (if needed for Redis alternative)

**Steps:**
```
1. Create Railway account
2. Create new project
3. Deploy from GitHub:
   - Connect repository
   - Select backend/ directory
   - Set environment variables
   - Deploy
4. Assign domain or use railway.app subdomain
```

---

**Option 2: Render**

**Advantages:**
- Free tier available
- Automatic deployments from git
- Managed PostgreSQL (if needed)

**Steps:**
```
1. Create Render account
2. New Web Service
3. Connect repository
4. Configure:
   - Build Command: go build -o main cmd/server/main.go
   - Start Command: ./main
   - Environment: Docker or Native
5. Add environment variables
6. Deploy
```

---

**Option 3: Fly.io**

**Advantages:**
- Global edge deployment
- Docker-native
- Good performance

**Steps:**
```
1. Install flyctl CLI
2. fly launch (in backend directory)
3. Configure fly.toml
4. Set secrets: fly secrets set KEY=value
5. Deploy: fly deploy
```

---

**Frontend Deployment Options:**

**Option 1: Vercel (Recommended)**

**Advantages:**
- Optimized for React/Vite
- Automatic deployments
- Global CDN
- Free tier generous

**Steps:**
```
1. Create Vercel account
2. Import repository
3. Configure:
   - Framework: Vite
   - Root Directory: frontend/
   - Build Command: npm run build
   - Output Directory: dist
4. Add environment variables
5. Deploy
```

---

**Option 2: Netlify**

**Advantages:**
- Similar to Vercel
- Easy setup
- Good free tier

**Steps:**
```
1. Create Netlify account
2. New site from Git
3. Configure build:
   - Base directory: frontend
   - Build command: npm run build
   - Publish directory: frontend/dist
4. Deploy
```

---

**Option 3: Static Hosting (Cloudflare Pages, GitHub Pages)**

**Advantages:**
- Very fast (CDN)
- Free
- Simple

**Limitations:**
- Static only (no server-side rendering)
- Manual deployment (or CI/CD)

---

### Docker Deployment (VPS)

**If using single VPS:**

**Dockerfile.prod (Backend):**
```
Conceptual multi-stage build:

Stage 1: Build
  FROM golang:1.21-alpine
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go build -o main cmd/server/main.go

Stage 2: Runtime
  FROM alpine:latest
  COPY --from=builder /app/main /main
  EXPOSE 8000
  CMD ["/main"]
```

**Dockerfile.prod (Frontend):**
```
Conceptual multi-stage build:

Stage 1: Build
  FROM node:20-alpine
  COPY package*.json ./
  RUN npm ci
  COPY . .
  RUN npm run build

Stage 2: Serve
  FROM nginx:alpine
  COPY --from=builder /app/dist /usr/share/nginx/html
  EXPOSE 80
  CMD ["nginx", "-g", "daemon off;"]
```

**Deploy to VPS:**
```bash
# Build images
docker build -f Dockerfile.prod -t afs-backend:latest ./backend
docker build -f Dockerfile.prod -t afs-frontend:latest ./frontend

# Push to registry (Docker Hub, GitHub Container Registry)
docker push your-registry/afs-backend:latest
docker push your-registry/afs-frontend:latest

# On VPS: Pull and run
docker pull your-registry/afs-backend:latest
docker pull your-registry/afs-frontend:latest

docker run -d \
  --name afs-backend \
  --env-file .env.production \
  -p 8000:8000 \
  afs-backend:latest

docker run -d \
  --name afs-frontend \
  -p 80:80 \
  afs-frontend:latest
```

---

## ✓ Verification & Testing

### Post-Deployment Checks

**Backend Health:**

```bash
# Health endpoint
curl https://your-backend.com/health

# Expected response
{
  "status": "healthy",
  "version": "1.0.0",
  "dependencies": {
    "database": "connected",
    "mcp_server": "available"
  }
}
```

**Database Connectivity:**
```bash
# Test query via API
curl https://your-backend.com/api/v1/agents

# Should return agent list
```

**WebSocket:**
```
Open browser console:
const ws = new WebSocket('wss://your-backend.com/ws')
ws.onopen = () => console.log('Connected')
ws.onmessage = (msg) => console.log(msg.data)

Expected: "Connected" logged
```

---

### End-to-End Test

**Complete User Flow:**

1. **Access Frontend:**
   ```
   Navigate to: https://your-frontend.com
   Verify: Page loads, no console errors
   ```

2. **Create Task:**
   ```
   Click "New Task"
   Fill form with test query:
     SELECT u.email, SUM(o.total) as revenue
     FROM users u 
     JOIN orders o ON u.id = o.user_id
     WHERE o.status = 'completed'
     GROUP BY u.email
   Submit
   Verify: Redirects to task detail
   ```

3. **Monitor Progress:**
   ```
   Watch real-time updates
   Verify: WebSocket events appear
   Verify: Agent status updates
   Expected timeline:
     - Agents assigned (~2s)
     - Forks created (~5-10s)
     - Analysis completed (~30s per agent)
     - Proposals submitted (~1m)
     - Benchmarks completed (~2-3m)
     - Consensus reached (~5s)
     - Total: ~4-5 minutes
   ```

4. **View Results:**
   ```
   Verify: Consensus decision shown
   Verify: Winning proposal highlighted
   Verify: Score breakdown visible
   Verify: Rationale displayed
   ```

5. **Check Database:**
   ```bash
   psql <tiger-connection-string>
   
   -- Verify task created
   SELECT id, type, status FROM tasks ORDER BY id DESC LIMIT 1;
   
   -- Verify proposals exist
   SELECT COUNT(*) FROM optimization_proposals;
   
   -- Verify consensus
   SELECT winning_proposal_id FROM consensus_decisions;
   ```

**If all steps succeed:** Deployment successful

---

### Performance Testing

**Load Test (Optional):**

```bash
# Using Apache Bench
ab -n 100 -c 10 https://your-backend.com/api/v1/tasks

# Using k6
k6 run load-test.js

# Metrics to watch:
# - Response time <500ms (API)
# - No 500 errors
# - Database connections stable
```

**Tiger Cloud Fork Performance:**
```
Create task with 3 agents
Measure:
  - Fork creation time (<10s each)
  - Total task completion (<5 minutes)
  - No fork creation failures

Success criteria:
  ✓ All 3 forks created successfully
  ✓ Zero-copy performance evident
  ✓ Cleanup completes (forks deleted)
```

---

## 📤 Challenge Submission

### Submission Requirements

**DEV.to Post:**

**Required Sections:**
1. Project Title and Description
2. Link to Live Demo
3. GitHub Repository
4. Video Demo (optional but recommended)
5. Technical Implementation Details
6. Tiger Cloud Integration Highlights
7. Challenges Faced
8. Future Enhancements

**Template:**
```markdown
# Agentic Fork Squad - Multi-Agent Database Optimization

## 🎯 What It Does
AFS uses multiple AI agents (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash via Vertex AI) working in 
parallel to optimize slow database queries. Each agent proposes 
different optimizations, benchmarks them in isolated Tiger Cloud 
forks, and a consensus system picks the best solution.

## 🔗 Links
- Live Demo: https://afs-demo.vercel.app
- GitHub: https://github.com/your-username/afs
- Demo Video: https://youtu.be/xxxxx

## 🏆 Tiger Cloud Innovation
- Zero-copy forks enable parallel agent experimentation
- MCP protocol for standardized DB operations
- 3 agents × 3 forks created in <10 seconds
- Storage efficiency: Shared data via Fluid Storage

## 🛠️ Tech Stack
- Backend: Go + Fiber
- Frontend: React + TypeScript + Vite
- Database: Tiger Cloud PostgreSQL
- AI: Vertex AI Client (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)

## 📊 Demo Credentials
Email: demo@afs.com
Password: demo123
(Read-only access for judges)

## 🎥 How to Test
1. Visit live demo
2. Click "New Task"
3. Use example query or paste your own
4. Watch agents work in real-time
5. See consensus decision

## 💡 Implementation Highlights
[Describe architecture, challenges, learnings]

## 🚀 Future Work
- Agent learning from past decisions
- Support for more database types
- Advanced optimization strategies
```

---

### Demo Credentials

**For Judges to Test:**

**Option 1: Public Access (Recommended)**
```
No authentication required
Anyone can create tasks
Rate limited to prevent abuse
Demo data pre-seeded
```

**Option 2: Demo Account**
```
Username: judge@tigercloud.com
Password: TigerChallenge2024!
Full access to system
Pre-created example tasks
```

**Option 3: API Keys (For Technical Judges)**
```
API Key: demo_key_for_judges
Postman Collection: Provided in README
curl examples: Documented
```

---

### Video Demo (Optional but Recommended)

**Content to Cover (3-5 minutes):**

1. **Introduction (30s)**
   - Problem statement
   - Solution overview

2. **Architecture (1m)**
   - Show diagram
   - Explain agent system
   - Highlight Tiger Cloud integration

3. **Live Demo (2-3m)**
   - Create optimization task
   - Show real-time progress
   - Explain agent proposals
   - Show consensus decision
   - Highlight fork efficiency

4. **Technical Deep Dive (1m)**
   - Code walkthrough (brief)
   - MCP integration
   - Zero-copy forks benefit

5. **Conclusion (30s)**
   - Innovation summary
   - Future potential

**Tools:**
- Loom (screen recording)
- OBS Studio (advanced)
- Zoom (record presentation)

**Upload to:**
- YouTube (unlisted or public)
- Include link in DEV.to post

---

## ↩️ Rollback Procedures

### Database Rollback

**Scenario: Migration Failed**

**Using PITR (Point-in-Time Recovery):**
```bash
# Get timestamp before migration
BEFORE_MIGRATION="2024-11-08T10:00:00Z"

# Create fork from that timestamp
tiger service fork \
  --parent afs-production \
  --name afs-rollback \
  --timestamp "$BEFORE_MIGRATION"

# Verify fork has correct state
psql $(tiger db connection-string afs-rollback)

# If correct:
# Option A: Use rollback fork temporarily
#   Update backend to use afs-rollback

# Option B: Promote fork to main (advanced)
#   Contact Tiger Cloud support
```

**Using Migration Down:**
```bash
# Rollback last migration
migrate -path ./backend/migrations \
        -database "$TIGER_DB" \
        down 1

# Verify
migrate -path ./backend/migrations \
        -database "$TIGER_DB" \
        version
```

---

### Application Rollback

**Platform-Specific:**

**Railway:**
```
Dashboard → Deployments
Select previous successful deployment
Click "Redeploy"
```

**Render:**
```
Dashboard → Service → Events
Find previous deploy
Click "Rollback to this version"
```

**Vercel:**
```
Dashboard → Deployments
Select previous deployment
Click "Promote to Production"
```

**Docker (VPS):**
```bash
# Tag images with versions
docker tag afs-backend:latest afs-backend:v1.2.3

# Rollback
docker stop afs-backend
docker rm afs-backend
docker run -d --name afs-backend afs-backend:v1.2.2
```

---

### Emergency Procedures

**Complete System Down:**

**Step 1: Identify Issue**
```
Check:
- Platform status page
- Tiger Cloud status
- Application logs
- Database connectivity
```

**Step 2: Quick Fix**
```
If backend issue:
  - Rollback application
  - Check environment variables
  - Verify API keys valid

If database issue:
  - Check Tiger Cloud status
  - Verify connection string
  - Test direct connection

If frontend issue:
  - Redeploy static assets
  - Clear CDN cache
  - Check API URL configuration
```

**Step 3: Communication**
```
If during judging:
  - Post status in submission comments
  - Provide ETA for fix
  - Offer alternative demo (video, local)
```

---

## 📊 Monitoring & Maintenance

### Application Monitoring

**Logging:**

**Backend Logs:**
```bash
# Railway
railway logs

# Render
render logs

# Docker
docker logs -f afs-backend
```

**Frontend Logs:**
```bash
# Vercel
vercel logs

# Check browser console for client errors
```

---

**Metrics to Track:**

**Application Health:**
- Uptime percentage
- Response time (p50, p95, p99)
- Error rate (4xx, 5xx)
- WebSocket connections (active)

**Database:**
- Connection count
- Query performance
- Fork creation time
- Storage usage

**Vertex AI (LLM Models):**
- Request count per model (gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash)
- Token usage (input/output)
- Cost per day (by model)
- Error rate (quota/auth/timeouts)

---

### Health Checks

**Automated Monitoring (Optional):**

**UptimeRobot:**
```
Setup:
1. Create account
2. Add monitor:
   - Type: HTTP(s)
   - URL: https://your-backend.com/health
   - Interval: 5 minutes
3. Alert on downtime
```

**Healthchecks.io:**
```
Setup:
1. Create check
2. Ping from cron job:
   curl https://hc-ping.com/your-uuid
3. Alert if no ping received
```

---

### Tiger Cloud Monitoring

**Fork Inventory:**
```bash
# Daily check for orphaned forks
tiger service list --parent afs-production

# If orphaned forks found (>24h old)
tiger service delete <fork-id>
```

**Storage Usage:**
```bash
# Check database size
psql $TIGER_DB -c "
  SELECT 
    pg_size_pretty(pg_database_size(current_database())) 
    as db_size;
"

# Track growth over time
# Alert if exceeds quota
```

**Connection Health:**
```bash
# Test MCP connectivity
curl https://mcp.tigercloud.io/health

# Test database connection
psql $TIGER_DB -c "SELECT 1;"
```

---

### Cost Monitoring

**LLM API Costs:**

```
Track daily spend per model:
  gemini-2.5-pro: X tokens × $rate = $Y
  gemini-2.5-flash: X tokens × $rate = $Y
  gemini-2.0-flash: X tokens × $rate = $Y
  
Total daily: $Z

Monthly projection: $Z × 30

Set alerts:
  - Daily > $10: Warning
  - Daily > $50: Critical
```

**Platform Costs:**
```
Railway: Free tier or $5/month
Vercel: Free tier
Tiger Cloud: Free tier (challenge period)

Total: ~$0-5/month during challenge
```

---

### Maintenance Tasks

**Weekly:**
- [ ] Review application logs for errors
- [ ] Check orphaned forks and cleanup
- [ ] Monitor LLM API usage and costs
- [ ] Verify backup strategy working

**Monthly:**
- [ ] Review and rotate API keys
- [ ] Update dependencies (security patches)
- [ ] Performance optimization review
- [ ] Documentation updates

**As Needed:**
- [ ] Scale resources if traffic increases
- [ ] Add monitoring/alerts
- [ ] Optimize database queries
- [ ] Refactor based on usage patterns

---

## 🎯 Summary

This deployment strategy provides:

**Phased Approach:**
- Phase 1: Local development (Docker)
- Phase 2: Tiger Cloud integration
- Phase 3: Production deployment

**Tiger Cloud Setup:**
- Account creation
- CLI authentication
- Service provisioning
- Connection verification

**Migration Process:**
- Database migration (local → Tiger)
- Application configuration
- Environment variables
- Testing procedures

**Deployment Options:**
- Backend: Railway/Render/Fly.io
- Frontend: Vercel/Netlify
- Database: Tiger Cloud (mandatory)

**Verification:**
- Health checks
- End-to-end testing
- Performance validation
- Fork operations testing

**Challenge Submission:**
- DEV.to post template
- Demo credentials setup
- Video demo guidelines
- Links and documentation

**Rollback Procedures:**
- Database PITR rollback
- Application version rollback
- Emergency procedures

**Monitoring:**
- Application health
- Database metrics
- Cost tracking
- Maintenance schedule

**Key Success Metrics:**
- Zero-copy forks working (<10s creation)
- Complete optimization flow (end-to-end)
- Multi-agent collaboration (3 agents)
- Consensus decision transparent
- Live demo accessible to judges

---

**Related Documentation:**
- Previous: [10-DEVELOPMENT-WORKFLOW.md](10-DEVELOPMENT-WORKFLOW.md) 
  - Local development setup
- See also: [06-TIGER-CLOUD-MCP.md](06-TIGER-CLOUD-MCP.md) 
  - Tiger Cloud technical details
- See also: [00-PROJECT-OVERVIEW.md](00-PROJECT-OVERVIEW.md) 
  - Project vision and goals

---

**Document Status:** Complete  
**Last Reviewed:** 2024  
**Maintained By:** Project Lead

---

**Challenge Deadline:** November 9, 2024, 11:59 PM PST

**Submission Checklist:**
- [ ] Application deployed and accessible
- [ ] DEV.to post published
- [ ] Demo credentials working
- [ ] GitHub repository public
- [ ] Video demo uploaded (optional)
- [ ] All links tested
- [ ] Tiger Cloud integration verified
- [ ] Ready for judge evaluation
```
