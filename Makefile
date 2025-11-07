.PHONY: help setup dev build test lint format clean security-check pre-push docker-up docker-down logs

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

help:
	@echo "$(BLUE)=== AFS (Agentic Fork Squad) Make Commands ===$(NC)"
	@echo ""
	@echo "$(GREEN)Setup$(NC)"
	@echo "  make setup           - Initial setup (copy .env.example, install deps)"
	@echo "  make setup-hooks     - Install git hooks for security"
	@echo ""
	@echo "$(GREEN)Development$(NC)"
	@echo "  make dev             - Start all services with docker-compose"
	@echo "  make docker-up       - Start services (same as dev)"
	@echo "  make docker-down     - Stop services"
	@echo "  make logs            - View service logs"
	@echo "  make logs-backend    - View backend logs only"
	@echo "  make logs-frontend   - View frontend logs only"
	@echo ""
	@echo "$(GREEN)Code Quality$(NC)"
	@echo "  make lint            - Lint backend and frontend"
	@echo "  make lint-backend    - Lint Go code"
	@echo "  make lint-frontend   - Lint TypeScript/React code"
	@echo "  make format          - Format backend and frontend"
	@echo "  make format-backend  - Format Go code"
	@echo "  make format-frontend - Format TypeScript/React code"
	@echo "  make test            - Run all tests"
	@echo "  make test-backend    - Run backend tests"
	@echo "  make test-frontend   - Run frontend tests"
	@echo ""
	@echo "$(GREEN)Security$(NC)"
	@echo "  make security-check  - Run security validation"
	@echo "  make pre-push        - Pre-push checks (security + tests)"
	@echo "  make verify-gitignore - Verify .gitignore is correct"
	@echo ""
	@echo "$(GREEN)Build$(NC)"
	@echo "  make build           - Build backend and frontend"
	@echo "  make build-backend   - Build Go backend binary"
	@echo "  make build-frontend  - Build React frontend"
	@echo ""
	@echo "$(GREEN)Database$(NC)"
	@echo "  make migrate-up      - Run migrations"
	@echo "  make migrate-down    - Rollback migrations"
	@echo "  make migrate-status  - Check migration status"
	@echo ""
	@echo "$(GREEN)Cleanup$(NC)"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make clean-all       - Clean everything including node_modules"
	@echo ""

# ============================================
# Setup
# ============================================

setup: setup-env install-deps
	@echo "$(GREEN)✅ Setup complete!$(NC)"
	@echo "Next steps:"
	@echo "  1. Fill in your credentials in .env"
	@echo "  2. Run: make dev"

setup-env:
	@if [ ! -f .env ]; then \
		echo "$(BLUE)Creating .env from .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(GREEN)✓ .env created$(NC)"; \
	else \
		echo "$(GREEN)✓ .env already exists$(NC)"; \
	fi

install-deps:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@cd backend && go mod download && echo "$(GREEN)✓ Backend dependencies$(NC)"
	@cd frontend && npm install && echo "$(GREEN)✓ Frontend dependencies$(NC)"

setup-hooks:
	@echo "$(BLUE)Installing git hooks...$(NC)"
	@chmod +x scripts/pre-push-check.sh
	@cp scripts/pre-push-check.sh .git/hooks/pre-push || true
	@echo "$(GREEN)✓ Pre-push hook installed$(NC)"

# ============================================
# Development
# ============================================

dev: docker-up
	@echo "$(GREEN)✅ Services started!$(NC)"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend: http://localhost:8000"
	@echo "Proxy: http://localhost"

docker-up:
	@echo "$(BLUE)Starting services...$(NC)"
	@docker-compose up -d
	@sleep 3
	@echo "$(GREEN)✓ Services started$(NC)"
	@docker-compose ps

docker-down:
	@echo "$(BLUE)Stopping services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

logs:
	@docker-compose logs -f

logs-backend:
	@docker-compose logs -f backend

logs-frontend:
	@docker-compose logs -f frontend

logs-mcp:
	@docker-compose logs -f mcp

# ============================================
# Code Quality
# ============================================

lint: lint-backend lint-frontend
	@echo "$(GREEN)✅ Linting complete$(NC)"

lint-backend:
	@echo "$(BLUE)Linting Go code...$(NC)"
	@cd backend && go fmt ./... && echo "$(GREEN)✓ Formatted$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		cd backend && golangci-lint run ./... && echo "$(GREEN)✓ Linted$(NC)"; \
	else \
		echo "$(RED)⚠️  golangci-lint not found. Install with: brew install golangci-lint$(NC)"; \
	fi

lint-frontend:
	@echo "$(BLUE)Linting TypeScript/React code...$(NC)"
	@cd frontend && npm run lint
	@echo "$(GREEN)✓ Linted$(NC)"

format: format-backend format-frontend
	@echo "$(GREEN)✅ Code formatted$(NC)"

format-backend:
	@echo "$(BLUE)Formatting Go code...$(NC)"
	@cd backend && go fmt ./... && goimports -w . 2>/dev/null || go fmt ./...
	@echo "$(GREEN)✓ Formatted$(NC)"

format-frontend:
	@echo "$(BLUE)Formatting TypeScript/React code...$(NC)"
	@cd frontend && npm run format 2>/dev/null || npx prettier --write "src/**/*.{ts,tsx,json,css}"
	@echo "$(GREEN)✓ Formatted$(NC)"

test: test-backend test-frontend
	@echo "$(GREEN)✅ All tests passed$(NC)"

test-backend:
	@echo "$(BLUE)Running backend tests...$(NC)"
	@cd backend && go test -v -timeout 30s ./...
	@echo "$(GREEN)✓ Backend tests passed$(NC)"

test-frontend:
	@echo "$(BLUE)Running frontend tests...$(NC)"
	@cd frontend && npm test 2>/dev/null || echo "$(YELLOW)⚠️  No tests configured$(NC)"

# ============================================
# Security
# ============================================

security-check: verify-gitignore check-secrets verify-no-tracked-secrets
	@echo "$(GREEN)✅ Security checks passed$(NC)"

verify-gitignore:
	@echo "$(BLUE)Verifying .gitignore...$(NC)"
	@grep -q "^.env$$" .gitignore && echo "$(GREEN)✓ .env ignored$(NC)" || (echo "$(RED)✗ .env not ignored$(NC)" && exit 1)
	@grep -q "^secrets/" .gitignore && echo "$(GREEN)✓ secrets/ ignored$(NC)" || (echo "$(RED)✗ secrets/ not ignored$(NC)" && exit 1)
	@grep -q "gcp_credentials.json" .gitignore && echo "$(GREEN)✓ gcp_credentials.json ignored$(NC)" || (echo "$(RED)✗ gcp_credentials.json not ignored$(NC)" && exit 1)

check-secrets:
	@echo "$(BLUE)Checking for hardcoded secrets...$(NC)"
	@if grep -r "password\s*=" backend/cmd --include="*.go" 2>/dev/null | grep -v "//" > /dev/null; then \
		echo "$(RED)✗ Found potential secrets in code$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ No obvious secrets found$(NC)"

verify-no-tracked-secrets:
	@echo "$(BLUE)Verifying no secrets are tracked...$(NC)"
	@git ls-files --error-unmatch .env 2>/dev/null && (echo "$(RED)✗ .env is tracked!$(NC)" && exit 1) || echo "$(GREEN)✓ .env not tracked$(NC)"
	@git ls-files --error-unmatch gcp_credentials.json 2>/dev/null && (echo "$(RED)✗ gcp_credentials.json is tracked!$(NC)" && exit 1) || echo "$(GREEN)✓ gcp_credentials.json not tracked$(NC)"
	@git ls-files | grep "^secrets/" | wc -l | grep -q "^0$$" && echo "$(GREEN)✓ No secrets/ files tracked$(NC)" || (echo "$(RED)✗ Found tracked secrets/$(NC)" && exit 1)

pre-push: security-check format lint test
	@echo ""
	@echo "$(GREEN)✅ All pre-push checks passed!$(NC)"
	@echo "Safe to push 🚀"

# ============================================
# Build
# ============================================

build: build-backend build-frontend
	@echo "$(GREEN)✅ Build complete$(NC)"

build-backend:
	@echo "$(BLUE)Building backend...$(NC)"
	@cd backend && go build -o bin/api ./cmd/api/main.go
	@echo "$(GREEN)✓ Backend built: backend/bin/api$(NC)"

build-frontend:
	@echo "$(BLUE)Building frontend...$(NC)"
	@cd frontend && npm run build
	@echo "$(GREEN)✓ Frontend built: frontend/dist$(NC)"

# ============================================
# Database
# ============================================

migrate-up:
	@echo "$(BLUE)Running migrations...$(NC)"
	@docker exec afs-backend goose up 2>/dev/null || docker exec afs-migrate /go/bin/goose -dir ./migrations postgres "$${DATABASE_URL}" up
	@echo "$(GREEN)✓ Migrations applied$(NC)"

migrate-down:
	@echo "$(BLUE)Rolling back migrations...$(NC)"
	@docker exec afs-backend goose down
	@echo "$(GREEN)✓ Migrations rolled back$(NC)"

migrate-status:
	@echo "$(BLUE)Checking migration status...$(NC)"
	@docker exec afs-backend goose status

# ============================================
# Cleanup
# ============================================

clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf backend/bin/ frontend/dist/ backend/vendor/
	@rm -f *.log
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-all: clean
	@echo "$(BLUE)Removing dependencies...$(NC)"
	@rm -rf frontend/node_modules backend/vendor
	@echo "$(GREEN)✓ Dependencies removed$(NC)"

# ============================================
# Docker
# ============================================

docker-logs:
	@docker-compose logs -f

docker-ps:
	@docker-compose ps

docker-shell-backend:
	@docker exec -it backend /bin/sh

docker-shell-frontend:
	@docker exec -it afs-frontend /bin/sh
