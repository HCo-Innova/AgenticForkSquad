# ✅ Hardening Checklist - Before Public GitHub Push

This checklist ensures your repository is secure before pushing to GitHub public.

**Use this before creating the initial commit to: https://github.com/HCo-Innova/AgenticForkSquad.git**

---

## Phase 1: Local Environment Setup ✓

- [ ] **Environment Files**
  - [ ] `.env` created from `.env.example`
  - [ ] `.env` is in `.gitignore` (verify: `git check-ignore .env`)
  - [ ] All `.env.local` variants in `.gitignore`
  - [ ] Run: `git ls-files --cached | grep "\.env" | wc -l` → should be 0

- [ ] **Secrets Directory**
  - [ ] `secrets/` folder exists
  - [ ] `secrets/` is in `.gitignore` (verify: `git check-ignore secrets/`)
  - [ ] Only `.gitkeep` and non-credential files in `secrets/`
  - [ ] No real credentials in `secrets/` on this machine (safe to verify)

- [ ] **Credentials**
  - [ ] GCP credentials downloaded locally (if using local dev)
  - [ ] GCP JSON file location: `./secrets/gcp_credentials.json`
  - [ ] Tiger Cloud `.env` filled with real values (LOCAL ONLY)
  - [ ] NO real credentials will be pushed (stay in .env)

---

## Phase 2: .gitignore Verification ✓

```bash
# Run these commands:

# 1. Verify all .env patterns
grep "^.env\|^.env\." .gitignore && echo "✅ .env patterns present"

# 2. Verify secrets/ is ignored
grep "^secrets/" .gitignore && echo "✅ secrets/ ignored"

# 3. Verify credential files
grep "gcp_credentials.json\|mcp-config.json" .gitignore && echo "✅ Credential files ignored"

# 4. Verify nothing tracked that should be ignored
git ls-files | grep -E "\.env|^secrets/" && echo "❌ FOUND TRACKED FILES!" || echo "✅ No sensitive files tracked"
```

- [ ] `.env` in `.gitignore`
- [ ] `.env.local`, `.env.production` in `.gitignore`
- [ ] `secrets/` directory in `.gitignore`
- [ ] `gcp_credentials.json` in `.gitignore`
- [ ] `mcp-config.json` in `.gitignore`
- [ ] Certificate files (`.pem`, `.key`, `.crt`) in `.gitignore`

---

## Phase 3: Code Review for Secrets ✓

```bash
# Scan for hardcoded credentials
grep -r "password\s*=" backend/cmd --include="*.go" | grep -v "//" && echo "❌ Found secrets!" || echo "✅ No secrets"
grep -r "secret\s*=" backend/internal --include="*.go" && echo "❌ Found secrets!" || echo "✅ No secrets"
grep -r "api.?key\|token\|Bearer" backend/cmd --include="*.go" && echo "❌ Found secrets!" || echo "✅ No secrets"
grep -r "password\|secret\|token" frontend/src --include="*.ts*" && echo "❌ Found secrets!" || echo "✅ No secrets"
```

- [ ] No hardcoded Tiger Cloud credentials in Go code
- [ ] No GCP credentials hardcoded anywhere
- [ ] No database passwords in source code
- [ ] No API keys in comments or code
- [ ] No real URLs with credentials in code

---

## Phase 4: Configuration Files ✓

- [ ] **`.env.example`** exists and is updated
  - [ ] Contains all required variables
  - [ ] Uses placeholder values: `your_value_here`, `change_me`, etc.
  - [ ] Includes comments explaining each variable
  - [ ] NO real values (verify all are placeholders)

- [ ] **`docker-compose.example.yml`** exists
  - [ ] References `.env` variables (${VAR_NAME})
  - [ ] Does NOT include real credentials
  - [ ] Can be used as template for local setup

- [ ] **`SECURITY.md`** created
  - [ ] Includes credential handling guidelines
  - [ ] Explains what files to .gitignore
  - [ ] Pre-push checklist included

- [ ] **`SETUP.md`** created
  - [ ] Instructions for post-clone setup
  - [ ] How to create `.env` from `.env.example`
  - [ ] How to add credentials locally

---

## Phase 5: Documentation Files ✓

- [ ] **`CONTRIBUTING.md`** created and includes:
  - [ ] Security guidelines for contributors
  - [ ] Pre-push checklist
  - [ ] Commit message format
  - [ ] How to avoid committing secrets

- [ ] **`.github/workflows/security-checks.yml`** exists
  - [ ] Scans PRs for secrets
  - [ ] Verifies no .env files committed
  - [ ] Checks .gitignore patterns

- [ ] **`README.md`** updated with:
  - [ ] Security section
  - [ ] Link to SECURITY.md
  - [ ] Pre-push check instructions

---

## Phase 6: Scripts ✓

- [ ] **`scripts/pre-push-check.sh`** exists and executable
  ```bash
  chmod +x scripts/pre-push-check.sh
  ./scripts/pre-push-check.sh  # Test it
  ```

- [ ] **`Makefile`** exists with useful targets
  - [ ] `make security-check` - runs security validation
  - [ ] `make pre-push` - full pre-push checks
  - [ ] `make setup` - initial setup

---

## Phase 7: Final Verification ✓

Run these commands BEFORE first git push:

```bash
# 1. No secrets in any file
git diff --cached --name-only | xargs grep -l "password\|secret\|key" 2>/dev/null && \
  echo "❌ SECRETS FOUND" || echo "✅ No secrets in staged changes"

# 2. .env is properly ignored
git check-ignore .env .env.local && echo "✅ .env ignored" || echo "❌ .env NOT ignored"

# 3. secrets/ is properly ignored  
git check-ignore secrets/ && echo "✅ secrets/ ignored" || echo "❌ secrets/ NOT ignored"

# 4. No sensitive files tracked
git ls-files | grep -E "\.env|^secrets/|gcp_credentials|mcp-config" && \
  echo "❌ SENSITIVE FILES TRACKED" || echo "✅ No sensitive files tracked"

# 5. Run full security check
./scripts/pre-push-check.sh

# 6. Verify Makefile works
make help | head -20
```

- [ ] No secrets in staged changes
- [ ] `.env` and `secrets/` properly ignored
- [ ] No sensitive files tracked in git
- [ ] `scripts/pre-push-check.sh` runs successfully
- [ ] `Makefile` has all essential targets

---

## Phase 8: GitHub Repository Setup ✓

Before pushing initial code:

```bash
# 1. Verify local .env not in repo
cat .gitignore | grep "^.env$"

# 2. Create local .env from template (will not be committed)
cp .env.example .env

# 3. Fill in credentials locally (test setup works)
# nano .env

# 4. TEST that services start
docker-compose up -d --build
docker-compose ps  # Should show healthy services

# 5. Stop before pushing
docker-compose down
```

- [ ] **Local .env created and filled** (will NOT be committed)
- [ ] **Services tested locally** (docker-compose works)
- [ ] **Services stopped** (docker-compose down)
- [ ] **Ready to push to GitHub**

---

## Phase 9: First GitHub Push ✓

```bash
# 1. Initialize git (if not already)
git init

# 2. Add remote (replace with your repo URL)
git remote add origin https://github.com/HCo-Innova/AgenticForkSquad.git

# 3. Verify .gitignore is correct before adding files
git status  # Should show everything except .env, secrets/, node_modules, etc.

# 4. Add all files
git add .

# 5. Verify nothing sensitive is staged
git diff --cached --name-only | head -20
# Review this list - should NOT contain: .env, secrets/, credentials, etc.

# 6. Run final security check
./scripts/pre-push-check.sh

# 7. If check passes, commit
git commit -m "Initial commit: AFS project setup with hardening"

# 8. Push to GitHub
git push -u origin main
```

- [ ] Remote URL set correctly
- [ ] Final review of staged files (no secrets)
- [ ] Pre-push security check passed
- [ ] First commit made
- [ ] Code pushed to GitHub public repository

---

## Phase 10: Post-Push Verification ✓

After pushing to GitHub:

```bash
# 1. Verify on GitHub that:
#    - No .env files visible in repo
#    - No secrets/ directory visible
#    - All docs (SECURITY.md, SETUP.md, etc.) visible

# 2. Verify GitHub Actions ran
#    - Go to: https://github.com/HCo-Innova/AgenticForkSquad/actions
#    - Verify "Security Checks" workflow passed

# 3. Create .env locally (if not done)
cp .env.example .env
# Fill with your real credentials

# 4. Install git hook
chmod +x scripts/pre-push-check.sh
cp scripts/pre-push-check.sh .git/hooks/pre-push
```

- [ ] GitHub shows NO .env, secrets/, or credentials
- [ ] GitHub Actions security checks passed
- [ ] `.env` created locally with real values
- [ ] Pre-push git hook installed locally

---

## Phase 11: Deployment Setup ✓

For Vercel + Railway/Render deployment:

### Frontend (Vercel)
- [ ] Connect GitHub repo to Vercel
- [ ] Set environment variables in Vercel dashboard:
  - [ ] `VITE_API_URL` - Backend URL
  - [ ] `VITE_WS_URL` - WebSocket URL
  - [ ] `NODE_ENV=production`
- [ ] Do NOT upload .env file to Vercel
- [ ] Test build succeeds

### Backend (Railway/Render)
- [ ] Connect GitHub repo to Railway/Render
- [ ] Set environment variables in platform dashboard:
  - [ ] `DATABASE_URL` - Tiger Cloud connection string
  - [ ] `TIGER_PROJECT_ID` - From Tiger Cloud
  - [ ] `GOOGLE_APPLICATION_CREDENTIALS` - Path or base64-encoded
  - [ ] All other .env variables
- [ ] Do NOT commit .env to repo
- [ ] Test deployment succeeds

### Database (Tiger Cloud)
- [ ] Create production services in Tiger Cloud
- [ ] Note connection strings for backend setup
- [ ] Test connection works
- [ ] Verify migrations run on deployment

---

## ⚠️ Critical Reminders

### DO NOT COMMIT:
- ❌ `.env` with real credentials
- ❌ GCP service account JSON
- ❌ Tiger Cloud passwords or keys
- ❌ Database connection strings with passwords
- ❌ Any hardcoded API keys or tokens

### DO COMMIT:
- ✅ `.env.example` with placeholder values
- ✅ `SECURITY.md` with guidelines
- ✅ `SETUP.md` with instructions
- ✅ `CONTRIBUTING.md` with contributor guidelines
- ✅ `.gitignore` with all secret patterns
- ✅ `docker-compose.example.yml` for reference

### If You Accidentally Committed Secrets:

**DO NOT PUSH TO PUBLIC REPO**

1. **Immediately:**
   ```bash
   git reset HEAD~1  # Undo last commit (keeps changes)
   git rm --cached .env  # Remove from git
   git commit -m "Remove .env from tracking"
   ```

2. **Rotate all credentials** (Tiger Cloud, GCP, database)

3. **Then push clean code**

---

## 🎯 Success Criteria

Your repository is ready for public GitHub when:

- ✅ No `.env`, `secrets/`, or credentials in git history
- ✅ `.env.example` with ALL required variables (no real values)
- ✅ SECURITY.md explains credential handling
- ✅ SETUP.md provides post-clone instructions
- ✅ CONTRIBUTING.md includes security guidelines
- ✅ `.github/workflows/security-checks.yml` runs on PRs
- ✅ Pre-push hook prevents accidental credential commits
- ✅ Services start locally with `docker-compose up -d`
- ✅ GitHub repository is public and accessible
- ✅ Vercel deployment configured for frontend
- ✅ Railway/Render deployment configured for backend

---

## 📞 Support

If you have questions about security setup:

1. Review [SECURITY.md](SECURITY.md)
2. Check [SETUP.md](SETUP.md)
3. See [CONTRIBUTING.md](CONTRIBUTING.md)
4. Run: `./scripts/pre-push-check.sh` for validation

---

**Last Updated:** November 2025  
**Status:** Ready for GitHub public push ✅
