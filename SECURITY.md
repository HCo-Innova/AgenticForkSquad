# 🔐 Security Guidelines

This document outlines security best practices for the AFS (Agentic Fork Squad) project.

## Critical: Never Commit Secrets

**NEVER** commit the following to version control:

- ❌ `.env` files with real credentials
- ❌ GCP service account JSON files (`gcp_credentials.json`)
- ❌ Tiger Cloud credentials (project ID, API keys, passwords)
- ❌ Database passwords or connection strings
- ❌ Private keys, certificates, or tokens
- ❌ API keys of any kind
- ❌ Files in `secrets/` directory

**All sensitive data must be in `.gitignore`.**

## Pre-Push Security Checklist

Before pushing any code to GitHub:

```bash
# 1. Verify no secrets are staged
git diff --cached | grep -i "password\|secret\|key\|token\|credentials" && \
  echo "⚠️  FOUND SECRETS!" && exit 1 || \
  echo "✅ No secrets detected"

# 2. Check git status for ignored files
git status --ignored

# 3. Verify .env is not tracked
git ls-files | grep -E "\.env|secrets/" && \
  echo "⚠️  FOUND TRACKED SECRETS!" && exit 1 || \
  echo "✅ All secrets properly ignored"

# 4. Run git hooks (if configured)
# make pre-push-check
```

## Environment Setup

### Local Development

1. **Copy template files:**
   ```bash
   cp .env.example .env
   cp docker-compose.example.yml docker-compose.yml  # optional
   ```

2. **Fill in local credentials:**
   ```bash
   # Edit .env with your local values
   POSTGRES_PASSWORD=your_local_password
   TIGER_DB_PASSWORD=your_tiger_password
   # etc...
   ```

3. **Verify .env is ignored:**
   ```bash
   git check-ignore .env  # Should return: .env
   ```

### Production Deployment

Use platform-specific credential management:

**Vercel (Frontend):**
- Use Vercel dashboard: Settings → Environment Variables
- Set `VITE_API_URL` and `VITE_WS_URL` 
- Never hardcode credentials in code

**Railway/Render/Fly.io (Backend):**
- Use platform dashboard for environment variables
- Enable secret rotation
- Use managed databases when available

**Tiger Cloud Database:**
- Store in Railway/Render/Fly.io environment variables
- Use connection string format: `postgres://user:pass@host:port/db?sslmode=require`
- Enable SSL/TLS for all connections

## Credential Handling by Component

### Backend (.env variables)

| Variable | Type | Sensitivity | Storage |
|----------|------|-------------|---------|
| `DATABASE_URL` | Connection String | 🔴 CRITICAL | Platform secrets |
| `TIGER_DB_PASSWORD` | Password | 🔴 CRITICAL | Platform secrets |
| `TIGER_PUBLIC_KEY` | API Key | 🟠 HIGH | Platform secrets |
| `TIGER_SECRET_KEY` | API Key | 🔴 CRITICAL | Platform secrets |
| `VERTEX_PROJECT_ID` | Project ID | 🟡 MEDIUM | Platform secrets |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path | 🟡 MEDIUM | Local file only |
| `PORT` | Integer | 🟢 LOW | .env.example |
| `LOG_LEVEL` | String | 🟢 LOW | .env.example |

### Frontend (.env variables)

| Variable | Type | Sensitivity | Storage |
|----------|------|-------------|---------|
| `VITE_API_URL` | URL | 🟡 MEDIUM | .env.example (local), Platform secrets (prod) |
| `VITE_WS_URL` | URL | 🟡 MEDIUM | .env.example (local), Platform secrets (prod) |
| `NODE_ENV` | String | 🟢 LOW | .env.example |

## GCP Service Account Setup

The project uses GCP for Vertex AI (Gemini models).

### Getting Your Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project
3. Navigate to: **Service Accounts** → **Create Service Account**
4. Grant role: **Vertex AI User** (or **Editor** for full access)
5. Create JSON key and download
6. Store locally: `./secrets/gcp_credentials.json`

### For Vercel Deployment

Option A (Recommended): Use Railway/Render for backend, not Vercel
Option B: Store in Railway/Render backend, frontend calls via API

**Do NOT upload GCP JSON to Vercel.** Vercel is for frontend only.

## Tiger Cloud Credentials

Tiger Cloud provides:
- **Project ID**: Identifies your Tiger Cloud project
- **Service IDs**: Unique identifiers for database instances
- **Credentials**: Username and password for each service
- **MCP URL**: Model Context Protocol server endpoint

### Storage for Local Development

Create `./secrets/mcp-config.json` with Tiger credentials:
```json
{
  "tiger_project_id": "your_project_id",
  "tiger_main_service": "your_main_service_id",
  "tiger_fork_services": [
    {
      "service_id": "fork_1_id",
      "password": "fork_1_password"
    },
    {
      "service_id": "fork_2_id", 
      "password": "fork_2_password"
    }
  ],
  "tiger_public_key": "your_public_key",
  "tiger_secret_key": "your_secret_key"
}
```

This file is in `.gitignore` and never committed.

### Storage for Production

Deploy the backend to Railway/Render/Fly.io and store all Tiger credentials as environment variables through their dashboards.

## .gitignore Verification

Verify these patterns are in `.gitignore`:

```
# Environment files
.env
.env.local
.env.*.local
.env.production

# Secrets directory
secrets/
secrets/**

# Credentials
*.pem
*.key
*.crt
*.p8
gcp_credentials.json
mcp-config.json

# IDE and OS
.vscode/
.idea/
*.swp
.DS_Store
.env.*.bak

# Build outputs
dist/
build/
vendor/
node_modules/
```

## Code Security Practices

### Do NOT:
- ❌ Hardcode credentials in source code
- ❌ Use credentials from .env in frontend code (will be bundled!)
- ❌ Log passwords or tokens
- ❌ Commit `.env` or `.env.*` files
- ❌ Share credentials via chat/email/PR
- ❌ Use placeholder credentials in comments

### DO:
- ✅ Use environment variables for all secrets
- ✅ Load credentials at runtime
- ✅ Use `.env.example` to document required variables
- ✅ Rotate credentials regularly
- ✅ Use principle of least privilege (minimal permissions)
- ✅ Enable SSL/TLS for all connections

## Reviewing Secrets Before Pushing

Use this script to catch secrets before committing:

```bash
#!/bin/bash
# scripts/pre-push-check.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

PATTERNS=(
  "password"
  "secret"
  "api.?key"
  "bearer"
  "token"
  "credential"
  "private.?key"
)

for pattern in "${PATTERNS[@]}"; do
  if git diff --cached | grep -iq "$pattern"; then
    echo -e "${RED}✗ Found potential secret: $pattern${NC}"
    git diff --cached | grep -in "$pattern"
    exit 1
  fi
done

echo -e "${GREEN}✓ No obvious secrets detected${NC}"
exit 0
```

## Incident Response

If you accidentally commit a secret:

1. **Immediately revoke the credential** (rotate passwords/keys)
2. **Do NOT push to public repo**
3. **If already pushed, rotate ALL credentials**
4. **Remove from git history:**
   ```bash
   git filter-repo --path credentials.json --invert-paths
   git push origin --force-with-lease
   ```
5. **Update all services** with new credentials

## Additional Resources

- [GitHub: Removing Sensitive Data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [OWASP: Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Go: Secure Development](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet.html)
- [React: Environment Variables](https://vitejs.dev/guide/env-and-mode.html)

---

**Last Updated:** November 2025  
**Maintained By:** AFS Team
