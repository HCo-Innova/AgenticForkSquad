# Contributing to AFS (Agentic Fork Squad)

Thank you for your interest in contributing to the AFS project! This document provides guidelines for contributing.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## Getting Started

### 1. Fork & Clone

```bash
git clone https://github.com/HCo-Innova/AgenticForkSquad.git
cd AgenticForkSquad
```

### 2. Set Up Development Environment

```bash
# Copy environment template
cp .env.example .env

# Fill in your local credentials (see SETUP.md)
nano .env

# Start services
docker-compose up -d

# Install pre-push hook for security
chmod +x scripts/pre-push-check.sh
cp scripts/pre-push-check.sh .git/hooks/pre-push
```

### 3. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

## Development Workflow

### Backend (Go)

```bash
# Format code
cd backend
go fmt ./...

# Lint (if available)
golangci-lint run ./...

# Run tests
go test ./...

# Build
go build -o bin/api ./cmd/api/main.go
```

### Frontend (React + TypeScript)

```bash
cd frontend

# Install dependencies
npm install

# Format and lint
npm run lint
npm run format

# Build
npm run build

# Preview production build
npm run preview
```

### Database

Migrations are in `backend/migrations/`:

```bash
# Apply migrations (auto on docker-compose up)
docker exec afs-backend goose up

# Rollback one migration
docker exec afs-backend goose down

# Check status
docker exec afs-backend goose status
```

## Commit Guidelines

### Message Format

Follow conventional commits:

```
type(scope): subject

body (optional)

footer (optional)
```

### Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `test:` - Tests
- `refactor:` - Code refactoring (no behavior change)
- `perf:` - Performance improvement
- `chore:` - Build, CI, or dependency updates
- `security:` - Security fixes

### Examples

```bash
git commit -m "feat(hybrid-search): add pgvector support"
git commit -m "fix(backend): handle null database responses"
git commit -m "docs: update deployment guide for Vercel"
git commit -m "test(consensus): add benchmark suite"
```

## Pre-Push Checklist

**BEFORE pushing to GitHub:**

```bash
# 1. Run security checks
./scripts/pre-push-check.sh

# 2. Verify no .env files
git diff --cached --name-only | grep "\.env" && \
  echo "❌ .env files found!" && exit 1

# 3. Verify no secrets/ directory
git diff --cached --name-only | grep "^secrets/" && \
  echo "❌ secrets/ files found!" && exit 1

# 4. Run tests
cd backend && go test ./...
cd frontend && npm test

# 5. Format code
cd backend && go fmt ./...
cd frontend && npm run format

# 6. Only then push
git push origin feature/your-feature-name
```

## Pull Request Process

### 1. Create PR with Description

Use the PR template and include:

- **What problem does this solve?**
- **How were you tested?**
- **Any breaking changes?**
- **Screenshots** (if UI change)
- **Checklist:**
  - [ ] Tests passing
  - [ ] No new security issues
  - [ ] Documentation updated
  - [ ] Code formatted and linted
  - [ ] No hardcoded credentials

### 2. Code Review

- Address reviewer feedback promptly
- Discuss disagreements respectfully
- Request re-review after changes

### 3. Merge

- Merge strategy: Squash commits (keeps history clean)
- Delete feature branch after merge

## Security Guidelines

### CRITICAL: Never Commit Secrets

- ❌ `.env` files with real values
- ❌ GCP service account JSON
- ❌ Tiger Cloud credentials
- ❌ Database passwords
- ❌ API keys or tokens

**All sensitive data must be in `.env` (ignored).**

### Code Security

- Use environment variables for all secrets
- Never log passwords or tokens
- Use HTTPS for all external connections
- Validate all user inputs
- Follow [OWASP Guidelines](https://owasp.org/)

See [SECURITY.md](SECURITY.md) for detailed guidelines.

## Testing

### Backend

```bash
cd backend

# Run all tests
go test ./...

# Run specific package
go test ./internal/usecases

# With coverage
go test -cover ./...

# Run benchmarks (if available)
go test -bench=. ./...
```

### Frontend

```bash
cd frontend

# Run tests (if configured)
npm test

# Component testing
npm run test -- --watch
```

## Documentation

Update documentation for:

- New features
- API changes
- Configuration changes
- Deployment procedures

See `docs/` directory for structure.

## Reporting Issues

Use GitHub Issues with:

1. **Clear title** describing the problem
2. **Description** with steps to reproduce
3. **Expected vs actual** behavior
4. **Environment info:**
   - OS and version
   - Go/Node.js versions
   - Docker version

Example:

```
Title: Backend crashes when database connection times out

Description:
When the Tiger Cloud database is unreachable, the backend crashes instead of gracefully degrading.

Steps to reproduce:
1. Start backend with Tiger Cloud
2. Simulate network outage (disconnect internet)
3. Observe backend logs

Expected: Backend logs error and retries
Actual: Backend process exits

Environment:
- Ubuntu 22.04
- Go 1.25
- Docker 25.0
```

## Project Structure

```
backend/
├── cmd/               # Entry points
├── internal/
│   ├── domain/        # Business logic
│   ├── usecases/      # Use cases
│   ├── infrastructure/# External services
│   └── config/        # Configuration
└── migrations/        # Database migrations

frontend/
├── src/
│   ├── components/    # React components
│   ├── pages/         # Page components
│   ├── hooks/         # Custom hooks
│   ├── services/      # API clients
│   └── types/         # TypeScript types
└── public/            # Static assets

docs/                  # Documentation
scripts/               # Helper scripts
```

## Architecture

- **Backend:** Go with Clean Architecture (Domain → Use Cases → Infrastructure)
- **Database:** PostgreSQL 16 via Tiger Cloud
- **Frontend:** React 19 + TypeScript 5 + Vite
- **AI:** Vertex AI (Gemini models)
- **Real-time:** WebSockets

See [System Architecture](docs/03-SYSTEM-ARCHITECTURE.md) for details.

## Deployment

Before deploying to production:

1. ✅ All tests passing
2. ✅ Code reviewed and approved
3. ✅ Security checks passed (no secrets in code)
4. ✅ Documentation updated
5. ✅ Performance tested (if applicable)

See [Deployment Strategy](docs/11-DEPLOYMENT-STRATEGY.md) for details.

## Getting Help

- **Questions:** Open a GitHub Discussion
- **Bugs:** File a GitHub Issue
- **Security Issues:** Email security team (do NOT create public issue)
- **Documentation:** See `docs/` directory

## Recognition

Contributors will be recognized in:
- `README.md` Contributors section
- GitHub Contributors graph
- Release notes

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

**Thank you for contributing! 🙏**

Questions? Open a discussion or reach out to the maintainers.
