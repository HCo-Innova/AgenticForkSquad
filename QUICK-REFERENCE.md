# 🚀 Quick Reference - Deployment Hardening

**Archivo de referencia rápida para deployment a GitHub público (Vercel + Railway)**

---

## ⚡ 30-Segundo Setup

```bash
# 1. Clone
git clone https://github.com/HCo-Innova/AgenticForkSquad.git
cd AgenticForkSquad

# 2. Environment
cp .env.example .env
# Edit .env with your credentials (NEVER commit this)

# 3. Start
docker-compose up -d

# 4. Before push: Verify security
./scripts/pre-push-check.sh  # Should show ✅ All checks passed
```

---

## 🔐 Critical Security Rules

### ❌ NEVER Commit:
```
.env (with real values)
gcp_credentials.json
secrets/ (with real credentials)
Tiger Cloud passwords
Database connection strings with passwords
API keys or tokens
```

### ✅ ALWAYS Commit:
```
.env.example (with placeholders)
SECURITY.md
SETUP.md
CONTRIBUTING.md
.gitignore
docker-compose.example.yml
scripts/pre-push-check.sh
```

---

## 📋 Pre-Push Checklist (90 seconds)

```bash
# 1. Run security script
./scripts/pre-push-check.sh

# 2. Verify .env is ignored
git check-ignore .env && echo "✅ Ignored" || echo "❌ NOT ignored"

# 3. Verify no secrets tracked
git ls-files | grep -E "\.env|secrets/|gcp_" && echo "❌ FOUND" || echo "✅ Clean"

# 4. If all green ✅
git push origin main
```

---

## 📚 Key Documents

| Document | Purpose | Read Time |
|----------|---------|-----------|
| `SETUP.md` | Post-clone setup guide | 10 min |
| `SECURITY.md` | Security policy & guidelines | 15 min |
| `CONTRIBUTING.md` | Developer guidelines | 10 min |
| `HARDENING-CHECKLIST.md` | Complete pre-GitHub checklist | 20 min |
| `DEPLOYMENT-PREP-SUMMARY.md` | What we prepared (this conv) | 5 min |

---

## 🎯 Deployment Steps

### 1. GitHub (First Time)

```bash
git init
git add .
git commit -m "Initial commit: AFS with security hardening"
git remote add origin https://github.com/HCo-Innova/AgenticForkSquad.git
git branch -M main
git push -u origin main
```

### 2. Vercel (Frontend)

```
1. Go to vercel.com
2. "Add New" → "Project"
3. Select AgenticForkSquad repo
4. Environment Variables:
   - VITE_API_URL = https://backend-url/api
   - VITE_WS_URL = wss://backend-url/ws
5. Deploy
```

### 3. Railway/Render (Backend)

```
1. Connect to GitHub repo
2. Environment Variables (from .env):
   - DATABASE_URL = postgres://...
   - TIGER_DB_PASSWORD = ...
   - VERTEX_PROJECT_ID = ...
   - GOOGLE_APPLICATION_CREDENTIALS = (uploaded/base64)
   - All other TIGER_* variables
3. Deploy
```

### 4. Tiger Cloud (Database)

```
1. Create services:
   - Main database
   - Fork A1 for agents
   - Fork A2 for agents
2. Get connection strings
3. Set in Railway/Render env vars
```

---

## 🚨 If You Accidentally Committed Secrets

**STOP! Do NOT push to GitHub**

```bash
# 1. Undo last commit
git reset HEAD~1

# 2. Remove sensitive file
git rm --cached .env

# 3. Verify it's in .gitignore
grep "^.env" .gitignore

# 4. New commit without secrets
git commit -m "Remove .env from tracking"

# 5. Rotate all credentials (Tiger Cloud, GCP, Database)

# 6. THEN push
git push origin main
```

---

## 💡 Pro Tips

### Make Commands
```bash
make dev              # Start all services
make pre-push         # Full pre-push validation
make security-check   # Security verification only
make setup            # Initial setup
make logs             # View service logs
```

### Verify Setup Locally
```bash
# Backend health
curl http://localhost:8000/health

# Frontend
open http://localhost:3000

# Database
psql $DATABASE_URL -c "SELECT version();"

# All services
docker-compose ps
```

### Check Credentials Location
```bash
# Should exist (local only, not in git)
ls -la .env

# Should be ignored
git check-ignore .env

# Should NOT be in git
git ls-files | grep ".env" | wc -l  # Should be 0
```

---

## 🔧 Environment Variables Quick Reference

### Backend Required
```
POSTGRES_PASSWORD
REDIS_PASSWORD
TIGER_DB_PASSWORD
TIGER_PUBLIC_KEY
TIGER_SECRET_KEY
VERTEX_PROJECT_ID
PORT=8000
```

### Frontend Required
```
VITE_API_URL=http://localhost/api
VITE_WS_URL=ws://localhost/ws
NODE_ENV=development
```

### All Variables
See: `.env.example` (complete reference)

---

## ❓ Troubleshooting Quick Fixes

| Problem | Solution |
|---------|----------|
| Docker won't start | `docker-compose down -v && docker-compose up -d` |
| Port 8000 in use | `lsof -i :8000 && kill PID` |
| Frontend not building | `cd frontend && rm -rf node_modules && npm install` |
| Database connection fails | `echo $DATABASE_URL` and verify connection string |
| MCP not responding | `docker-compose restart mcp` |
| .env.example outdated | Check current .env and update .env.example |

---

## 📞 Need Help?

1. **Local Setup Issues** → See `SETUP.md`
2. **Security Questions** → See `SECURITY.md`
3. **Contributing** → See `CONTRIBUTING.md`
4. **Pre-GitHub Checklist** → See `HARDENING-CHECKLIST.md`
5. **Deployment Issues** → See `docs/11-DEPLOYMENT-STRATEGY.md`

---

## ✅ Status

- ✅ All files prepared for GitHub
- ✅ Security hardening complete
- ✅ Documentation ready
- ✅ Verification scripts working
- ✅ Ready for: `git push`

---

**Last Updated:** November 2025  
**Quick Ref Version:** 1.0
